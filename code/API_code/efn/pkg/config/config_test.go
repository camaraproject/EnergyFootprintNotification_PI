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
package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Config(t *testing.T) {
	setRequireVariables(t)

	t.Run("correctly parse API environment variable", func(t *testing.T) {
		t.Setenv("API_ADDRESS", "127.0.0.1:6969")
		res := GetConf().API
		assert.Equal(t, "127.0.0.1:6969", res.Address)
	})
	t.Run("correctly parse API max time period days", func(t *testing.T) {
		t.Setenv("API_MAX_TIME_PERIOD_DAYS", "180")
		res := GetConf().API
		assert.Equal(t, 180, res.MaxTimePeriodDays)
	})
	t.Run("use default max time period days when not set", func(t *testing.T) {
		res := GetConf().API
		assert.Equal(t, 730, res.MaxTimePeriodDays)
	})
	t.Run("correctly parse database environment variables", func(t *testing.T) {
		t.Setenv("DB_URI", "http://127.0.0.1:6969")
		t.Setenv("DB_NAME", "thisDB")
		res := GetConf().Database
		assert.Equal(t, "http://127.0.0.1:6969", res.Uri)
		assert.Equal(t, "thisDB", res.Name)
	})
	t.Run("correctly parse log environment variables", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("LOG_FORMAT", "development")
		res := GetConf().Log
		assert.Equal(t, "debug", res.Level)
		assert.Equal(t, "development", res.Format)
	})
	t.Run("correctly parse policy decision point environment variables", func(t *testing.T) {
		t.Setenv("PDP_ADDRESS", "http://127.0.0.1:6969")
		t.Setenv("PDP_AUTHORIZED_APP_IDS", "netflix,prime")
		t.Setenv("PDP_SKIP_POLICY_CHECK", "true")
		res := GetConf().PDP
		assert.Equal(t, "http://127.0.0.1:6969", res.Address)
		assert.True(t, res.SkipPolicyCheck)
	})
}

// setRequireVariables sets default environment variables for tests
func setRequireVariables(t *testing.T) {
	t.Helper()
	// Add required variables for the tests
	// t.Setenv("", "")
}

func TestGetLogConfig(t *testing.T) {
	t.Run("correctly parse log environment variables", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("LOG_FORMAT", "development")
		res := GetLogConfig()
		assert.Equal(t, "debug", res.Level)
		assert.Equal(t, "development", res.Format)
	})
}
