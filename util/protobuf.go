// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/util

package util

import (
	"google.golang.org/protobuf/proto"
)

func Marshal(m proto.Message) []byte {
	bs, _ := proto.Marshal(m)
	return bs
}

func Unmarshal(bs []byte, m proto.Message) error {
	return proto.Unmarshal(bs, m)
}
