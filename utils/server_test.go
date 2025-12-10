package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type testHandler struct {
	called bool
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called = true
	w.WriteHeader(http.StatusTeapot)
	_, _ = w.Write([]byte("ok"))
}

func TestLoggingMiddleware_CallsNextHandler(t *testing.T) {
	th := &testHandler{}
	mw := LoggingMiddleware(th)

	req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
	w := httptest.NewRecorder()

	mw.ServeHTTP(w, req)

	if !th.called {
		t.Fatalf("expected inner handler to be called")
	}

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusTeapot {
		t.Fatalf("expected status %d, got %d", http.StatusTeapot, res.StatusCode)
	}
}
