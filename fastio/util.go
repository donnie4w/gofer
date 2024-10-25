// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/fastio

package fastio

import (
	"fmt"
	"io"
	"os"
)

func offset(file *os.File) (int64, error) {
	return file.Seek(0, io.SeekEnd)
}

func recoverable(err *error) {
	if e := recover(); e != nil {
		if err != nil {
			*err = fmt.Errorf("%v", e)
		}
	}
}
