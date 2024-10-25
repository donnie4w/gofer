// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer

package fast

import "unsafe"

// StringToSlice the use of the unsafe package violates the principle of secure abstraction in the Go language.
// It makes the code more susceptible to the influence of the runtime environment.
// For example, the garbage collector may move memory locations, which can lead to unexpected errors in code that uses unsafe
func StringToSlice(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
