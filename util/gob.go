// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of t source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// github.com/donnie4w/gofer/util

package util

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"github.com/donnie4w/gothrift/thrift"
)

func JsonEncode(v any) (bs []byte) {
	bs, _ = json.Marshal(v)
	return
}

func JsonDecode[T any](bs []byte) (_r T, err error) {
	err = json.Unmarshal(bs, &_r)
	return
}

func Encode(e any) (by []byte, err error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err = enc.Encode(e)
	by = buf.Bytes()
	return
}

func Decode[T any](buf []byte) (_r *T, err error) {
	decoder := gob.NewDecoder(bytes.NewReader(buf))
	_r = new(T)
	err = decoder.Decode(_r)
	return
}

var tconf = &thrift.TConfiguration{}

func TEncode(ts thrift.TStruct) (_r []byte) {
	buf := thrift.NewTMemoryBuffer()
	protocol := thrift.NewTCompactProtocolConf(buf, tconf)
	ts.Write(context.TODO(), protocol)
	protocol.Flush(context.TODO())
	_r = buf.Bytes()
	return
}

func TDecode[T thrift.TStruct](bs []byte, ts T) (_r T, err error) {
	buf := &thrift.TMemoryBuffer{Buffer: bytes.NewBuffer(bs)}
	protocol := thrift.NewTCompactProtocolConf(buf, tconf)
	err = ts.Read(context.TODO(), protocol)
	return ts, err
}
