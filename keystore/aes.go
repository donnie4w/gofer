package keystore

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
)

type Aes struct {
	// commonKey128 [16]byte
	// commonKey192 [24]byte
	commonKey256 [32]byte
	commonIV     [16]byte
}

func NewAes(iv string, key string) (_r *Aes) {
	_r = &Aes{}
	copy(_r.commonIV[:], _md5(iv))
	copy(_r.commonKey256[:], _sha256(key))
	return
}

func (this *Aes) EncrypterCbc(src []byte) (_r []byte, err error) {
	if c, e := aes.NewCipher(this.commonKey256[:]); e == nil {
		blockSize := c.BlockSize()
		bs := bytes.Repeat([]byte{byte(blockSize - len(src)%blockSize)}, blockSize-len(src)%blockSize)
		_r = append(src, bs...)
		encrypter := cipher.NewCBCEncrypter(c, this.commonIV[:])
		encrypter.CryptBlocks(_r, _r)
	} else {
		err = e
	}
	return
}

func (this *Aes) DecrypterCbc(bs []byte) (_r []byte, err error) {
	if c, e := aes.NewCipher(this.commonKey256[:]); e == nil {
		decrypter := cipher.NewCBCDecrypter(c, this.commonIV[:])
		_r = make([]byte, len(bs))
		decrypter.CryptBlocks(_r, bs)
		_r = _r[:len(_r)-(int(_r[len(_r)-1]))]
	} else {
		err = e
	}
	return
}

func AesEncrypter(src []byte, key string) (_r []byte, err error) {
	return NewAes(key, key).EncrypterCbc(src)
}

func AesDecrypter(bs []byte, key string) (_r []byte, err error) {
	return NewAes(key, key).DecrypterCbc(bs)
}

func _md5(s string) []byte {
	m := md5.New()
	m.Write([]byte(s))
	return m.Sum(nil)
}

func _sha256(s string) []byte {
	m := sha256.New()
	m.Write([]byte(s))
	return m.Sum(nil)
}
