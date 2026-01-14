package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// 基本 POST 成功
func TestPostSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	resp, err := PostWithOptions(
		context.Background(),
		ts.URL,
		[]byte("hello"),
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

// Header / Cookie 是否正确传递
func TestPostHeaderAndCookie(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "123" {
			t.Fatalf("header not received")
		}

		c, err := r.Cookie("test")
		if err != nil || c.Value != "abc" {
			t.Fatalf("cookie not received")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	_, err := PostWithOptions(
		context.Background(),
		ts.URL,
		nil,
		map[string]string{
			"X-Test": "123",
		},
		[]*http.Cookie{
			{Name: "test", Value: "abc"},
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// 非 2xx 状态码返回 error
func TestPostNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad request"))
	}))
	defer ts.Close()

	_, err := PostWithOptions(
		context.Background(),
		ts.URL,
		nil,
		nil,
		nil,
	)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

// 测试：UpdateConfig 是否生效（Timeout）
func TestUpdateConfigTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// 设置一个很短的超时
	UpdateConfig(func(c *Config) {
		c.Timeout = 50 * time.Millisecond
	})

	start := time.Now()
	_, err := PostWithOptions(
		context.Background(),
		ts.URL,
		nil,
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

func TestSetConfig(t *testing.T) {
	SetConfig(&Config{
		Timeout: 1 * time.Second,
	})

	if Client().Timeout != 1*time.Second {
		t.Fatalf("SetConfig did not apply timeout")
	}
}
