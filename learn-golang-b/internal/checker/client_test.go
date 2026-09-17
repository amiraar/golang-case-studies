package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestChecker_CheckOne_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	defer srv.Close()

	c := NewChecker()
	r := c.CheckOne(context.Background(), URLEntry{URL: srv.URL}, time.Second)

	if !r.OK() {
		t.Fatalf("CheckOne() error = %v, want nil", r.Err)
	}
	if r.StatusCode != http.StatusTeapot {
		t.Errorf("StatusCode = %d, want %d", r.StatusCode, http.StatusTeapot)
	}
	if r.Duration <= 0 {
		t.Errorf("Duration = %v, want > 0", r.Duration)
	}
}

func TestChecker_CheckOne_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewChecker()
	r := c.CheckOne(context.Background(), URLEntry{URL: srv.URL}, 20*time.Millisecond)

	if r.OK() {
		t.Fatal("CheckOne() error = nil, want timeout error")
	}
}

func TestChecker_CheckOne_ParsesServerDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewChecker()
	r := c.CheckOne(context.Background(), URLEntry{URL: srv.URL}, time.Second)

	if r.ServerDate.IsZero() {
		t.Error("ServerDate is zero, want parsed value from response Date header")
	}
}
