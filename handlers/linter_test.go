package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"codecollab/config"
	"codecollab/models"
)

func TestInvokeLinterCustomEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/2015-03-31/functions/codecollab-python-linter/invocations" {
			t.Errorf("unexpected Lambda request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Amz-Security-Token") != "session-token" {
			t.Error("temporary credential session token missing")
		}
		var request models.LambdaRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		if request.Language != "python" || request.Code != "x = unknown" {
			t.Errorf("unexpected payload: %+v", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"statusCode":200,"body":"{\"errors\":[{\"line\":1,\"column\":5,\"message\":\"undefined name\",\"severity\":\"warning\",\"length\":7}]}"}`))
	}))
	defer server.Close()
	cfg := &config.Config{
		AWSRegion: "ap-south-1", AWSAccessKeyID: "test", AWSSecretAccessKey: "test",
		AWSSessionToken: "session-token", LambdaEndpointURL: server.URL,
		LambdaARNPython: "codecollab-python-linter",
	}
	errors, err := InvokeLinter("python", "x = unknown", cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(errors) != 1 || errors[0].Message != "undefined name" || errors[0].Column != 5 {
		t.Fatalf("unexpected lint diagnostics: %+v", errors)
	}
}
