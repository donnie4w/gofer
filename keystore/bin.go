// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/tim/keystore

package keystore

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/donnie4w/gofer/util"
)

var KeyStore *_keyStore

type _keyStore struct {
	mux          *sync.Mutex
	fname        string
	_fileHandler *os.File
}

func NewKeyStore(dir string, name string) (ks *_keyStore, err error) {
	if err = os.MkdirAll(dir, 0777); err != nil {
		return
	}
	fname := fmt.Sprint(dir, "/", name)
	var _fileHandler *os.File
	if _fileHandler, err = os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0666); err == nil {
		ks = &_keyStore{&sync.Mutex{}, fname, _fileHandler}
	}
	return
}

func (this *_keyStore) Write(bs []byte) (err error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this._fileHandler.Seek(0, io.SeekStart)
	this._fileHandler.Truncate(0)
	var obs []byte
	if obs, err = util.Zlib(bs); err != nil {
		obs = bs
	}
	_, err = this._fileHandler.Write(obs)
	return
}

func (this *_keyStore) Read() (bs []byte, err error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	fi, err := this._fileHandler.Stat()
	if fi.Size() > 0 {
		if bs, err = util.ReadFile(this.fname); err == nil && bs != nil {
			if obs, er := util.UnZlib(bs); er == nil {
				return obs, er
			}
		}
	} else {
		err = errors.New("empty file")
	}
	return
}
