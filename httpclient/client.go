// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/httpclient

package httpclient

import (
	"net/http"
	"sync"
)

var (
	mu        sync.RWMutex
	cfg       = DefaultConfig()
	transport *http.Transport
	client    *http.Client
)

func init() {
	rebuild()
}

// rebuild 会在配置变更后重建 Transport / Client
func rebuild() {
	transport = &http.Transport{
		TLSClientConfig:     cfg.TLSConfig,
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
		DisableCompression:  cfg.DisableCompression,
		Proxy:               cfg.Proxy,
	}

	client = &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}
}

// SetConfig ：整体替换（推荐用于启动阶段）
func SetConfig(c *Config) {
	mu.Lock()
	defer mu.Unlock()
	cfg = c
	rebuild()
}

// UpdateConfig ：按需修改（线程安全）
func UpdateConfig(fn func(*Config)) {
	mu.Lock()
	defer mu.Unlock()
	fn(cfg)
	rebuild()
}

// Client ：获取当前 client
func Client() *http.Client {
	mu.RLock()
	defer mu.RUnlock()
	return client
}
