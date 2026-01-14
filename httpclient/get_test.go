package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// 基础 GET 成功
func TestGetSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	resp, err := Get(
		context.Background(),
		ts.URL,
		nil,
		nil,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(resp) != "ok" {
		t.Fatalf("unexpected response: %s", resp)
	}
}

// 测试：Query 参数是否正确
func TestGetWithQuery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("a") != "1" {
			t.Fatalf("missing query param a")
		}
		if r.URL.Query().Get("b") != "2" {
			t.Fatalf("missing query param b")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	_, err := GetWithQuery(
		context.Background(),
		ts.URL,
		map[string]string{
			"a": "1",
			"b": "2",
		},
		nil,
		nil,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// 测试：Header / Cookie 是否生效
func TestGetHeaderAndCookie(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "123" {
			t.Fatalf("header not received")
		}

		c, err := r.Cookie("token")
		if err != nil || c.Value != "abc" {
			t.Fatalf("cookie not received")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	_, err := Get(
		context.Background(),
		ts.URL,
		map[string]string{
			"X-Test": "123",
		},
		[]*http.Cookie{
			{Name: "token", Value: "abc"},
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// 非 2xx 状态码返回 error
func TestGetNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error"))
	}))
	defer ts.Close()

	_, err := Get(
		context.Background(),
		ts.URL,
		nil,
		nil,
	)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// Timeout 是否生效
func TestGetTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	UpdateConfig(func(c *Config) {
		c.Timeout = 50 * time.Millisecond
	})

	start := time.Now()
	_, err := Get(
		context.Background(),
		ts.URL,
		nil,
		nil,
	)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("expected timeout error")
	}

	if elapsed > 150*time.Millisecond {
		t.Fatalf("timeout not applied, elapsed=%v", elapsed)
	}
}
