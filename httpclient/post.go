// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/httpclient

package httpclient

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
	"strings"
)

func Post(bs []byte, url string) (_r []byte, err error) {
	return Post2(bs, false, url)
}

func Post2(bs []byte, close bool, url string) (_r []byte, err error) {
	return Post3(bs, close, url, nil, nil)
}

func Post3(bs []byte, close bool, url string, header map[string]string, c *http.Cookie) (_r []byte, err error) {
	transport := &http.Transport{DisableKeepAlives: true}
	if strings.HasPrefix(url, "https:") {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := http.Client{Transport: transport}
	var req *http.Request
	var reader io.Reader
	if len(bs) > 0 {
		reader = bytes.NewReader(bs)
	}
	if req, err = http.NewRequest(http.MethodPost, url, reader); err == nil {
		if close {
			req.Close = true
		}
		if len(header) > 0 {
			for k, v := range header {
				req.Header.Set(k, v)
			}
		}
		if c != nil {
			req.AddCookie(c)
		}
		var resp *http.Response
		if resp, err = client.Do(req); err == nil {
			defer resp.Body.Close()
			var body []byte
			if body, err = io.ReadAll(resp.Body); err == nil {
				_r = body
			}
		}
	}
	return
}
