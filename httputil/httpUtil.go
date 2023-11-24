// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/gofer/httputil

package httputil

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
	"strings"
)

func HttpPost(bs []byte, close bool, httpurl string) (_r []byte, err error) {
	tr := &http.Transport{DisableKeepAlives: true}
	if strings.HasPrefix(httpurl, "https:") {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := http.Client{Transport: tr}
	var req *http.Request
	if req, err = http.NewRequest(http.MethodPost, httpurl, bytes.NewReader(bs)); err == nil {
		if close {
			req.Close = true
		}
		var resp *http.Response
		if resp, err = client.Do(req); err == nil {
			if close {
				defer resp.Body.Close()
			}
			var body []byte
			if body, err = io.ReadAll(resp.Body); err == nil {
				_r = body
			}
		}
	}
	return
}
