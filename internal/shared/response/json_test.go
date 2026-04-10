package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONWritesEnvelope(t *testing.T) {
	res := httptest.NewRecorder()

	JSON(res, http.StatusCreated, Envelope{
		Success: true,
		Message: "created",
		Data: map[string]string{
			"id": "1",
		},
	})

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", res.Code)
	}
	if res.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected application/json content type, got %q", res.Header().Get("Content-Type"))
	}

	var payload Envelope
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if !payload.Success || payload.Message != "created" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}
