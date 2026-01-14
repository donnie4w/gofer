// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/httpclient

package httpclient

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

// Config 定义 httpclient 的全局配置项。
// 一般在程序启动阶段设置，运行中修改会触发 Transport 重建。
type Config struct {

	// Timeout 指整个 HTTP 请求的超时时间。
	// 包含：连接、重定向、读取响应体的总耗时。
	// 建议：对第三方 API 设置较小值（如 3~5 秒）， 对内部服务可适当放宽。
	Timeout time.Duration

	// MaxIdleConns 控制所有主机的最大空闲连接数。
	// 影响全局连接池大小。 默认值过小会导致频繁建连，过大会占用过多资源。
	MaxIdleConns int

	// MaxIdleConnsPerHost 控制每个 Host 的最大空闲连接数。
	// 在高并发访问同一服务时非常重要。
	// 如果你访问的是单一 API 服务，这个值应该适当调大。
	MaxIdleConnsPerHost int

	// IdleConnTimeout 指空闲连接在连接池中保持的最长时间。
	// 超过该时间未使用的连接会被关闭。
	// 合理设置可以避免长期占用无用连接。
	IdleConnTimeout time.Duration

	// DisableCompression 控制是否禁用自动解压响应体。
	// 默认 false，表示支持 gzip/deflate。
	// 在调试抓包或对接不规范服务时可能需要禁用。
	DisableCompression bool

	// TLSConfig 用于自定义 TLS 行为（HTTPS）。
	// 常见用途：
	//   - 设置最低 TLS 版本
	//   - 指定根证书（自签名证书）
	//   - 测试环境下跳过证书校验（不推荐用于生产）
	TLSConfig *tls.Config

	// Proxy 指定 HTTP 代理函数。
	// 可用于：
	//   - 公司内网代理
	//   - 调试流量（Charles / Fiddler）
	//   - 特定请求走代理
	// 若为 nil，则不使用代理。
	Proxy func(*http.Request) (*url.URL, error)
}

func DefaultConfig() *Config {
	return &Config{
		Timeout:             15 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
}
