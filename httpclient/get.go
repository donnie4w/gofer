// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/httpclient

package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Get ：基础用法
func Get(
	ctx context.Context,
	rawURL string,
	header map[string]string,
	cookies []*http.Cookie,
) ([]byte, error) {
	return GetWithQuery(ctx, rawURL, nil, header, cookies)
}

// GetWithQuery ：支持 query 参数
func GetWithQuery(
	ctx context.Context,
	rawURL string,
	query map[string]string,
	header map[string]string,
	cookies []*http.Cookie,
) ([]byte, error) {

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	if len(query) > 0 {
		q := u.Query()
		for k, v := range query {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		u.String(),
		nil,
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
