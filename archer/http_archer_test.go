package archer

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPArcher(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	cfg := Config{
		Target:     ts.URL,
		Interval:   "1ms",
		ConnNum:    10,
		Data:       []byte{1, 2, 3},
		PrintLog:   true,
		PrintError: true,
		Num:        10000,
	}

	t.Logf("Archer launching at: %s\n", ts.URL)

	if err := StartHTTPArcher(cfg); err != nil {
		t.Errorf("%s", err)
	}
}
