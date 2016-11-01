package rundeck

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMain(m *testing.M) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()
}
