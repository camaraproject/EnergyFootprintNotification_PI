/*
Copyright (C) 2022-2025 Contributors | TIM S.p.A. to CAMARA a Series of LF Projects, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/logger"
)

// This sinkreceiver service listens for CloudEvent callbacks from the Notification service.
// It implements callbacks defined in the following OpenAPI specification:
// https://github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/blob/main/api/energy-footprint-notification.yaml#L286

var (
	mu                 sync.Mutex
	totalRequests      int64
	successCount       int64
	failedCount        int64
	expectedMatchCount int64

	// For change detection in periodic logging
	lastLoggedTotal   int64
	lastLoggedSuccess int64
	lastLoggedFailed  int64
	lastLoggedMatched int64

	// Timing tracking
	timingExpected  int64     // Number of requests to wait for (0 = disabled)
	timingReceived  int64     // Requests received since timing started
	timingStartTime time.Time // When timing started
)

func main() {
	log := logger.Get()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	// Expected result value for validation (default: 0.0044)
	expectedValue := "0.0044"
	if val := os.Getenv("EXPECTED_RESULT_VALUE"); val != "" {
		expectedValue = val
	}
	log.Info("Expected result value configured", zap.String("expectedValue", expectedValue))

	// Start periodic stats logger (only logs if stats changed)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			total, success, failed, matched := totalRequests, successCount, failedCount, expectedMatchCount
			changed := total != lastLoggedTotal || success != lastLoggedSuccess ||
				failed != lastLoggedFailed || matched != lastLoggedMatched
			if changed {
				lastLoggedTotal = total
				lastLoggedSuccess = success
				lastLoggedFailed = failed
				lastLoggedMatched = matched
			}
			mu.Unlock()

			if changed {
				log.Info("Request statistics",
					zap.Int64("total", total),
					zap.Int64("success", success),
					zap.Int64("failed", failed),
					zap.Int64("expectedMatch("+expectedValue+")", matched))
			}
		}
	}()

	// Timing endpoint: POST /timing?count=N to start timing for N requests (0 to disable)
	http.HandleFunc("/timing", func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Timing endpoint called", zap.String("method", r.Method), zap.String("query", r.URL.RawQuery))

		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.Method == http.MethodGet {
			mu.Lock()
			expected := timingExpected
			received := timingReceived
			started := timingStartTime
			mu.Unlock()

			status := map[string]any{
				"expected": expected,
				"received": received,
				"enabled":  expected > 0,
			}
			if expected > 0 && !started.IsZero() {
				status["elapsed"] = time.Since(started).String()
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(status)
			return
		}

		countStr := r.URL.Query().Get("count")
		count, err := strconv.ParseInt(countStr, 10, 64)
		if err != nil || count < 0 {
			http.Error(w, "invalid count parameter", http.StatusBadRequest)
			return
		}

		mu.Lock()
		timingExpected = count
		timingReceived = 0
		timingStartTime = time.Now() // Start timer immediately when timing is configured
		// Reset all counters when timing is configured
		totalRequests = 0
		successCount = 0
		failedCount = 0
		expectedMatchCount = 0
		mu.Unlock()

		if count == 0 {
			log.Info("Timing disabled")
		} else {
			log.Info("Timing configured - timer started, counters reset", zap.Int64("expectedCount", count))
			fmt.Printf("\n>>> TIMING STARTED: Waiting for %d requests (counters reset) <<<\n\n", count)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message":  "timing configured",
			"expected": count,
		})
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		totalRequests++
		currentTotal := totalRequests
		mu.Unlock()

		log.Debug("Received request",
			zap.String("Authorization", r.Header.Get("Authorization")),
			zap.Int64("requestNumber", currentTotal))

		if r.Method != http.MethodPost {
			mu.Lock()
			failedCount++
			currentFailed := failedCount
			mu.Unlock()
			log.Warn("Method not allowed",
				zap.String("method", r.Method),
				zap.Int64("failed", currentFailed))
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		log.With(zap.String("body", string(body))).Debug("Received callback")
		var m map[string]any
		if err := json.Unmarshal(body, &m); err != nil {
			log.With(zap.Error(err)).Error("Failed to unmarshal callback body")
			mu.Lock()
			failedCount++
			currentFailed := failedCount
			mu.Unlock()
			log.Warn("JSON unmarshal failed", zap.Int64("failed", currentFailed))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		result, err := validateCloudEvent(m)
		if err != nil {
			log.With(zap.Error(err)).Error("Invalid CloudEvent")
			mu.Lock()
			failedCount++
			currentFailed := failedCount
			mu.Unlock()
			log.Info("CloudEvent validation failed",
				zap.Int64("total", currentTotal),
				zap.Int64("failed", currentFailed))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Check if this is an error notification by result value (-1 indicates error)
		resultStr := fmt.Sprintf("%v", result)
		resultFloat, _ := strconv.ParseFloat(resultStr, 64)
		isErrorNotification := resultFloat == -1.0

		if isErrorNotification {
			// Increment failed count for error notifications
			mu.Lock()
			failedCount++
			currentFailed := failedCount
			mu.Unlock()
			eventType, _ := m["type"].(string)

			log.Warn("Received error notification (result=-1)",
				zap.String("result", resultStr),
				zap.String("eventType", eventType),
				zap.Int64("total", currentTotal),
				zap.Int64("failed", currentFailed))
		}

		// Check if result matches expected value (only for success notifications)
		matched := !isErrorNotification && checkExpectedValue(resultStr, expectedValue)

		mu.Lock()
		if !isErrorNotification {
			successCount++
			if matched {
				expectedMatchCount++
			}
		}
		currentSuccess := successCount
		currentMatched := expectedMatchCount
		mu.Unlock()

		if !isErrorNotification && !matched {
			log.Warn("Result does not match expected value",
				zap.String("result", resultStr),
				zap.String("expected", expectedValue))
		}

		log.Debug("Parsed callback",
			zap.Any("id", m["id"]),
			zap.Any("type", m["type"]),
			zap.Any("result", result),
			zap.Bool("isError", isErrorNotification),
			zap.Bool("matchesExpected", matched),
			zap.Int64("total", currentTotal),
			zap.Int64("success", currentSuccess),
			zap.Int64("expectedMatch", currentMatched))

		// Timing: check if we reached expected count
		mu.Lock()
		if timingExpected > 0 {
			timingReceived++
			currentReceived := timingReceived
			expected := timingExpected
			if currentReceived == expected {
				elapsed := time.Since(timingStartTime)
				fmt.Println("")
				fmt.Println("==================================================")
				fmt.Printf("  TIMING COMPLETE: %d requests in %s\n", currentReceived, elapsed)
				fmt.Printf("  Requests/second: %.2f\n", float64(currentReceived)/elapsed.Seconds())
				fmt.Println("==================================================")
				fmt.Println("")
				log.Info("TIMING COMPLETE",
					zap.Int64("requests", currentReceived),
					zap.Duration("elapsed", elapsed),
					zap.String("elapsedStr", elapsed.String()),
					zap.Float64("requestsPerSecond", float64(currentReceived)/elapsed.Seconds()))
				// Log final stats immediately
				log.Info("Request statistics",
					zap.Int64("total", totalRequests),
					zap.Int64("success", successCount),
					zap.Int64("failed", failedCount),
					zap.Int64("expectedMatch", expectedMatchCount))
				// Reset timing for next batch
				timingExpected = 0
				timingReceived = 0
				timingStartTime = time.Time{}
			} else {
				log.Debug("Timing progress",
					zap.Int64("received", currentReceived),
					zap.Int64("expected", expected))
			}
		}
		mu.Unlock()

		w.WriteHeader(http.StatusAccepted)
	})

	log.With(zap.String("port", port)).Info("Sinkreceiver listening (HTTP only; TLS via sidecar)")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Failed to start HTTP server", zap.Error(err))
	}
}

var (
	allowedTypes = map[string]struct{}{
		"org.camaraproject.energy-footprint-notification.v1.energy":           {},
		"org.camaraproject.energy-footprint-notification.v1.carbon-footprint": {},
	}
)

// validateCloudEvent performs lightweight validation of the incoming CloudEvent JSON
// according to the CAMARA subset we care about. Returns error if invalid.
// validateCloudEvent validates the CloudEvent and extracts the result value for logging.
func validateCloudEvent(m map[string]any) (any, error) {
	id, _ := m["id"].(string)
	if id == "" {
		return nil, errors.New("missing or empty id")
	}
	source, _ := m["source"].(string)
	if source == "" {
		return nil, errors.New("missing or empty source")
	}
	typ, _ := m["type"].(string)
	if typ == "" {
		return nil, errors.New("missing type")
	}
	if _, ok := allowedTypes[typ]; !ok {
		return nil, errors.New("unsupported type " + typ)
	}
	spec, _ := m["specversion"].(string)
	if spec != "1.0" {
		return nil, errors.New("specversion must be 1.0")
	}
	// Accept both application/json and application/cloudevents+json for datacontenttype
	if dc, ok := m["datacontenttype"].(string); ok && dc != "application/json" && dc != "application/cloudevents+json" {
		return nil, errors.New("datacontenttype must be application/json or application/cloudevents+json")
	}
	timeStr, _ := m["time"].(string)
	if timeStr == "" {
		return nil, errors.New("missing time")
	}
	if _, err := time.Parse(time.RFC3339, timeStr); err != nil {
		return nil, errors.New("invalid time format (expect RFC3339)")
	}
	// data is required and must be object
	dataVal, ok := m["data"]
	if !ok {
		return nil, errors.New("missing data field")
	}
	dataMap, isMap := dataVal.(map[string]interface{})
	if !isMap {
		return nil, errors.New("data must be an object")
	}

	// Extract result value (errors are indicated by -1 value)
	extractResult := func(key string) (string, error) {
		val, ok := dataMap[key]
		if !ok {
			return "", fmt.Errorf("missing %s in data", key)
		}
		switch v := val.(type) {
		case string:
			return v, nil
		case float64:
			return fmt.Sprintf("%v", v), nil
		default:
			return "", fmt.Errorf("%s has unsupported type", key)
		}
	}

	// Extract the correct result field based on event type
	var result string
	switch typ {
	case "org.camaraproject.energy-footprint-notification.v1.energy":
		res, err := extractResult("energyConsumption")
		if err != nil {
			return nil, err
		}
		result = res
	case "org.camaraproject.energy-footprint-notification.v1.carbon-footprint":
		res, err := extractResult("carbonFootprint")
		if err != nil {
			return nil, err
		}
		result = res
	default:
		return nil, errors.New("unsupported type " + typ)
	}
	return result, nil
}

// checkExpectedValue compares the result with the expected value.
// Handles both string comparison and numeric comparison for floating point values.
func checkExpectedValue(result, expected string) bool {
	// Direct string match
	if result == expected {
		return true
	}

	// Try numeric comparison for floating point tolerance
	resultFloat, err1 := strconv.ParseFloat(result, 64)
	expectedFloat, err2 := strconv.ParseFloat(expected, 64)
	if err1 == nil && err2 == nil {
		// Use small epsilon for floating point comparison
		const epsilon = 1e-9
		diff := resultFloat - expectedFloat
		if diff < 0 {
			diff = -diff
		}
		return diff < epsilon
	}

	return false
}
