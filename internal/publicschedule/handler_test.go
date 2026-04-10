package publicschedule

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerListDoesNotExposeSensitiveFields(t *testing.T) {
	location := mustLoadLocation(t, "Asia/Jakarta")
	handler := Handler{
		Service: &Service{
			Store: fakeStore{
				items: []OccupiedRange{
					{Start: "09:30", End: "11:30", Status: "occupied"},
				},
			},
			Timezone: location,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/public/schedules?date=2026-05-20", nil)
	res := httptest.NewRecorder()

	handler.List(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", res.Code, res.Body.String())
	}

	body := res.Body.String()
	for _, sensitiveField := range []string{"client_name", "address", "notes"} {
		if strings.Contains(body, sensitiveField) {
			t.Fatalf("response must not contain sensitive field %q: %s", sensitiveField, body)
		}
	}
}

func TestHandlerListReturnsEmptyItems(t *testing.T) {
	location := mustLoadLocation(t, "Asia/Jakarta")
	handler := Handler{
		Service: &Service{
			Store:    fakeStore{},
			Timezone: location,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/public/schedules?date=2026-06-01", nil)
	res := httptest.NewRecorder()

	handler.List(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", res.Code, res.Body.String())
	}
	if !strings.Contains(res.Body.String(), `"items":[]`) {
		t.Fatalf("expected empty items response, got %s", res.Body.String())
	}
}
