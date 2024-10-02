// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
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

func Post(bs []byte, httpurl string) (_r []byte, err error) {
	return Post2(bs, false, httpurl)
}

func Post2(bs []byte, close bool, httpurl string) (_r []byte, err error) {
	return Post3(bs, close, httpurl, nil, nil)
}

func Post3(bs []byte, close bool, httpurl string, header map[string]string, c *http.Cookie) (_r []byte, err error) {
	transport := &http.Transport{DisableKeepAlives: true}
	if strings.HasPrefix(httpurl, "https:") {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := http.Client{Transport: transport}
	var req *http.Request
	var reader io.Reader
	if len(bs) > 0 {
		reader = bytes.NewReader(bs)
	}
	if req, err = http.NewRequest(http.MethodPost, httpurl, reader); err == nil {
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
