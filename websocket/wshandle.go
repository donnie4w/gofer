// Copyright (c) , donnie <donnie4w@gmail.com>
// All rights reserved.
//
// github.com/donnie4w/gofer/websocket

package websocket

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	wss "golang.org/x/net/websocket"
)

type Config struct {
	Url       string
	Origin    string
	CertFiles []string
	CertBytes [][]byte
	TimeOut   time.Duration
	OnOpen    func(c *Handler)
	OnError   func(c *Handler, err error)
	OnClose   func(c *Handler)
	OnMessage func(c *Handler, msg []byte)
}

type Handler struct {
	Cfg  *Config
	conn *wss.Conn
	mux  *sync.Mutex
	err  error
}

func NewHandler(cfg *Config) (wh *Handler, err error) {
	var conn *wss.Conn
	config := &wss.Config{Version: wss.ProtocolVersionHybi13}
	if cfg.TimeOut > 0 {
		config.Dialer = &net.Dialer{Timeout: cfg.TimeOut}
	}
	if strings.HasPrefix(cfg.Url, "wss:") {
		if cfg.CertFiles != nil {
			if rootcas, e := loadCACertificatesFromFiles(cfg.CertFiles); e == nil {
				config.TlsConfig = &tls.Config{
					RootCAs:            rootcas,
					InsecureSkipVerify: false,
				}
			} else {
				return nil, e
			}
		} else if cfg.CertBytes != nil {
			if rootcas, e := loadCACertificatesFromBytes(cfg.CertBytes); e == nil {
				config.TlsConfig = &tls.Config{
					RootCAs:            rootcas,
					InsecureSkipVerify: false,
				}
			} else {
				return nil, e
			}
		} else {
			config.TlsConfig = &tls.Config{InsecureSkipVerify: true}
		}
	} else if !strings.HasPrefix(cfg.Url, "ws:") {
		return nil, errors.New("network transfer protocol error:" + cfg.Url)
	}
	if config.Location, err = url.ParseRequestURI(cfg.Url); err == nil {
		if config.Origin, err = url.ParseRequestURI(cfg.Origin); err == nil {
			conn, err = wss.DialConfig(config)
		}
	}
	if err == nil && conn != nil {
		wh = &Handler{cfg, conn, &sync.Mutex{}, nil}
		if cfg.OnOpen != nil {
			cfg.OnOpen(wh)
		}
		go wh.read()
	}
	return
}

func (wh *Handler) sendws(bs []byte) (err error) {
	wh.mux.Lock()
	defer wh.mux.Unlock()
	return wss.Message.Send(wh.conn, bs)
}

func (wh *Handler) Send(bs []byte) error {
	if wh.err != nil {
		return wh.err
	}
	return wh.sendws(bs)
}

func (wh *Handler) Close() (err error) {
	wh.err = errors.New("connection is closed")
	if wh.conn != nil {
		err = wh.conn.Close()
	}
	return
}

func (wh *Handler) Error() error {
	return wh.err
}

func (wh *Handler) read() {
	var err error
	for wh.err == nil {
		var byt []byte
		if err = wss.Message.Receive(wh.conn, &byt); err != nil {
			wh.err = err
			break
		}
		if byt != nil && wh.Cfg.OnMessage != nil {
			go wh.Cfg.OnMessage(wh, byt)
		}
	}
	if wh.Cfg.OnError != nil {
		go wh.Cfg.OnError(wh, err)
	}
	wh.Close()
	if wh.Cfg.OnClose != nil {
		wh.Cfg.OnClose(wh)
	}
}

func recoverable(err *error) {
	if e := recover(); e != nil {
		if err != nil {
			*err = fmt.Errorf("panic: %v", e)
		}
	}
}

func loadCACertificatesFromFiles(certFiles []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, certFile := range certFiles {
		pemData, err := os.ReadFile(certFile)
		if err != nil {
			return nil, err
		}

		if ok := pool.AppendCertsFromPEM(pemData); !ok {
			return nil, errors.New("failed to append certificates from PEM data")
		}
	}
	return pool, nil
}

func loadCACertificatesFromBytes(certBytes [][]byte) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, bs := range certBytes {
		if ok := pool.AppendCertsFromPEM(bs); !ok {
			return nil, errors.New("failed to append certificates from PEM data")
		}
	}
	return pool, nil
}
