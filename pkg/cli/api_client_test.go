package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScheduleClient(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:       "localhost",
			Port:       8002,
			AuthAPIKey: "test-api-key",
		},
	}

	client := NewScheduleClient(cfg)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8002", client.baseURL)
	assert.Equal(t, "test-api-key", client.apiKey)
}

func TestScheduleClient_SetBaseURL(t *testing.T) {
	client := &ScheduleClient{}
	client.SetBaseURL("http://custom:9000")
	assert.Equal(t, "http://custom:9000", client.baseURL)
}

func TestScheduleClient_IsServerAvailable(t *testing.T) {
	t.Run("server available", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/osm/server-info", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := &ScheduleClient{
			baseURL: server.URL,
			client:  http.DefaultClient,
		}
		assert.True(t, client.IsServerAvailable())
	})

	t.Run("server unavailable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := &ScheduleClient{
			baseURL: server.URL,
			client:  http.DefaultClient,
		}
		assert.False(t, client.IsServerAvailable())
	})

	t.Run("empty baseURL", func(t *testing.T) {
		client := &ScheduleClient{
			baseURL: "",
			client:  http.DefaultClient,
		}
		assert.False(t, client.IsServerAvailable())
	})
}

func TestScheduleClient_RegisterCronTrigger(t *testing.T) {
	workflow := &core.Workflow{
		Name: "test-workflow",
		Kind: core.KindModule,
	}
	trigger := &core.Trigger{
		Name:     "daily-scan",
		Schedule: "0 0 * * *",
		Enabled:  true,
	}

	t.Run("success - 201 created", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/osm/api/schedules", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "test-key", r.Header.Get("x-osm-api-key"))

			var req scheduleRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)

			assert.Equal(t, "test-workflow-daily-scan", req.Name)
			assert.Equal(t, "test-workflow", req.WorkflowName)
			assert.Equal(t, "module", req.WorkflowKind)
			assert.Equal(t, "example.com", req.Target)
			assert.Equal(t, "0 0 * * *", req.Schedule)
			assert.True(t, req.Enabled)

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"data": {"id": "sch_123"}}`))
		}))
		defer server.Close()

		client := &ScheduleClient{
			baseURL: server.URL,
			apiKey:  "test-key",
			client:  http.DefaultClient,
		}

		err := client.RegisterCronTrigger(context.Background(), workflow, trigger, "example.com", map[string]string{"tactic": "default"})
		assert.NoError(t, err)
	})

	t.Run("success - 409 conflict (schedule exists)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusConflict)
			_, _ = w.Write([]byte(`{"error": true, "message": "schedule already exists"}`))
		}))
		defer server.Close()

		client := &ScheduleClient{
			baseURL: server.URL,
			client:  http.DefaultClient,
		}

		err := client.RegisterCronTrigger(context.Background(), workflow, trigger, "example.com", nil)
		assert.NoError(t, err) // 409 is treated as success
	})

	t.Run("failure - 400 bad request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": true, "message": "invalid request"}`))
		}))
		defer server.Close()

		client := &ScheduleClient{
			baseURL: server.URL,
			client:  http.DefaultClient,
		}

		err := client.RegisterCronTrigger(context.Background(), workflow, trigger, "example.com", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "400")
	})

	t.Run("failure - empty baseURL", func(t *testing.T) {
		client := &ScheduleClient{
			baseURL: "",
			client:  http.DefaultClient,
		}

		err := client.RegisterCronTrigger(context.Background(), workflow, trigger, "example.com", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server URL not configured")
	})
}

func TestNewRunClient(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host:       "localhost",
			Port:       8002,
			AuthAPIKey: "test-api-key",
		},
	}

	client := NewRunClient(cfg)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8002", client.baseURL)
	assert.Equal(t, "test-api-key", client.apiKey)
}

func TestRunClient_SetBaseURL(t *testing.T) {
	client := &RunClient{}
	client.SetBaseURL("http://custom:9000")
	assert.Equal(t, "http://custom:9000", client.baseURL)
}

func TestRunClient_IsServerAvailable(t *testing.T) {
	t.Run("server available", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/server-info", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"version": "5.0.0"}`))
		}))
		defer server.Close()

		client := &RunClient{
			baseURL: server.URL,
			client:  http.DefaultClient,
		}
		assert.True(t, client.IsServerAvailable())
	})

	t.Run("server unavailable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := &RunClient{
			baseURL: server.URL,
			client:  http.DefaultClient,
		}
		assert.False(t, client.IsServerAvailable())
	})

	t.Run("empty baseURL", func(t *testing.T) {
		client := &RunClient{
			baseURL: "",
			client:  http.DefaultClient,
		}
		assert.False(t, client.IsServerAvailable())
	})
}

func TestRunClient_CreateRun(t *testing.T) {
	t.Run("success - 202 accepted", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/osm/api/runs", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "test-key", r.Header.Get("x-osm-api-key"))

			var req CreateRunRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)

			assert.Equal(t, "test-module", req.Module)
			assert.Equal(t, "example.com", req.Target)
			assert.Equal(t, "high", req.Priority)

			w.WriteHeader(http.StatusAccepted)
			resp := CreateRunResponse{
				Message:     "Run started",
				Workflow:    "test-module",
				Kind:        "module",
				TargetCount: 1,
				Priority:    "high",
				JobID:       "abc123",
				Status:      "queued",
				PollURL:     "/osm/api/jobs/abc123",
				RunUUID:     "run-uuid-123",
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := &RunClient{
			baseURL: server.URL,
			apiKey:  "test-key",
			client:  http.DefaultClient,
		}

		req := &CreateRunRequest{
			Module:   "test-module",
			Target:   "example.com",
			Priority: "high",
		}

		resp, err := client.CreateRun(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "Run started", resp.Message)
		assert.Equal(t, "test-module", resp.Workflow)
		assert.Equal(t, "high", resp.Priority)
		assert.Equal(t, "abc123", resp.JobID)
		assert.Equal(t, "run-uuid-123", resp.RunUUID)
	})

	t.Run("failure - 400 bad request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": true, "message": "Invalid priority"}`))
		}))
		defer server.Close()

		client := &RunClient{
			baseURL: server.URL,
			client:  http.DefaultClient,
		}

		req := &CreateRunRequest{
			Module:   "test-module",
			Target:   "example.com",
			Priority: "invalid",
		}

		resp, err := client.CreateRun(context.Background(), req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid priority")
	})

	t.Run("failure - 404 workflow not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error": true, "message": "Workflow not found"}`))
		}))
		defer server.Close()

		client := &RunClient{
			baseURL: server.URL,
			client:  http.DefaultClient,
		}

		req := &CreateRunRequest{
			Module: "nonexistent-module",
			Target: "example.com",
		}

		resp, err := client.CreateRun(context.Background(), req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Workflow not found")
	})

	t.Run("failure - empty baseURL", func(t *testing.T) {
		client := &RunClient{
			baseURL: "",
			client:  http.DefaultClient,
		}

		req := &CreateRunRequest{
			Module: "test-module",
			Target: "example.com",
		}

		resp, err := client.CreateRun(context.Background(), req)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server URL not configured")
	})

	t.Run("success - multiple targets", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req CreateRunRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)

			assert.Equal(t, 3, len(req.Targets))
			assert.Equal(t, 2, req.Concurrency)

			w.WriteHeader(http.StatusAccepted)
			resp := CreateRunResponse{
				Message:     "Run started",
				Workflow:    "test-module",
				Kind:        "module",
				TargetCount: 3,
				Priority:    "normal",
				JobID:       "multi123",
				Status:      "queued",
				PollURL:     "/osm/api/jobs/multi123",
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := &RunClient{
			baseURL: server.URL,
			client:  http.DefaultClient,
		}

		req := &CreateRunRequest{
			Module:      "test-module",
			Targets:     []string{"t1.com", "t2.com", "t3.com"},
			Concurrency: 2,
			Priority:    "normal",
		}

		resp, err := client.CreateRun(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, 3, resp.TargetCount)
	})
}
