// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/httpclient

package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// Post ：基础用法
func Post(url string, body []byte) ([]byte, error) {
	return PostWithOptions(context.Background(), url, body, nil, nil)
}

// PostWithHeader ：支持 header
func PostWithHeader(
	url string,
	body []byte,
	header map[string]string,
) ([]byte, error) {
	return PostWithOptions(context.Background(), url, body, header, nil)
}

// PostWithOptions ：完整参数
func PostWithOptions(
	ctx context.Context,
	url string,
	body []byte,
	header map[string]string,
	cookies []*http.Cookie,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	for k, v := range header {
		req.Header.Set(k, v)
	}

	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err := Client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, data)
	}

	return data, nil
}
