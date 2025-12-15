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
	"net/http"

	"github.com/cloudevents/sdk-go/v2/binding"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"go.uber.org/zap"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/internal/database"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/internal/reciever/worker"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/config"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/logger"
	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/orchestrator"
)

func main() {
	conf := config.GetConf()
	log := logger.Get()

	// Single service; manual path-based CloudEvent parsing for normal and DLQ.

	db, err := database.NewMongoDB(conf.Database)
	if err != nil {
		log.With(zap.Error(err), zap.String("Mongo URI", conf.Database.Uri), zap.String("DB Name", conf.Database.Name)).
			Fatal("Failed to connect to mongo Database")
	}

	orch, err := orchestrator.NewDummyClient()
	if err != nil {
		log.With(zap.Error(err)).Fatal("Failed to create orchestrator client")
	}

	handler, err := worker.NewHandler(db, orch)
	if err != nil {
		log.With(zap.Error(err)).
			Fatal("Failed to create worker handler")
	}

	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Normal event ingress (broker deliveries)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Ignore explicit healthz here (handled above)
		if r.URL.Path != "/" { // any other non-root path -> 404
			http.NotFound(w, r)
			return
		}
		msg := cehttp.NewMessageFromHttpRequest(r)
		defer msg.Finish(nil)
		evt, err := binding.ToEvent(r.Context(), msg)
		if err != nil {
			http.Error(w, "invalid CloudEvent", http.StatusBadRequest)
			return
		}
		if _, herr := handler.Handle(r.Context(), *evt); herr != nil {
			http.Error(w, "processing error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})

	// Dead letter sink ingress (/dlq). Always 202 to avoid further broker retries.
	mux.HandleFunc("/dlq", func(w http.ResponseWriter, r *http.Request) {
		msg := cehttp.NewMessageFromHttpRequest(r)
		defer msg.Finish(nil)
		evt, err := binding.ToEvent(r.Context(), msg)
		if err != nil {
			http.Error(w, "invalid CloudEvent", http.StatusBadRequest)
			return
		}
		// We intentionally ignore processing error to avoid infinite DLQ loops; log is inside handler.
		_, _ = handler.HandleDLQEvent(r.Context(), *evt)
		w.WriteHeader(http.StatusAccepted)
	})

	addr := conf.API.Address
	if addr == "" {
		addr = ":8080"
	}
	log.With(zap.String("address", addr)).Info("Starting worker server (single service with / and /dlq paths)")
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.With(zap.Error(err)).Fatal("Failed to start HTTP server")
	}
}
