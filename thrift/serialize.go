// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
// github.com/donnie4w/gofer/thrift
package thrift

import (
	"bytes"
	"context"

	gothrift "github.com/apache/thrift/lib/go/thrift"
)

var tconf = &gothrift.TConfiguration{}

func TEncode(ts gothrift.TStruct) (_r []byte) {
	buf := &gothrift.TMemoryBuffer{Buffer: bytes.NewBuffer([]byte{})}
	protocol := gothrift.NewTCompactProtocolConf(buf, tconf)
	ts.Write(context.Background(), protocol)
	protocol.Flush(context.Background())
	_r = buf.Bytes()
	return
}

func TDecode[T gothrift.TStruct](bs []byte, ts T) (_r T, err error) {
	buf := &gothrift.TMemoryBuffer{Buffer: bytes.NewBuffer(bs)}
	protocol := gothrift.NewTCompactProtocolConf(buf, tconf)
	err = ts.Read(context.Background(), protocol)
	return ts, err
}
