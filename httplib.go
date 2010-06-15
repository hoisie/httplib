// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// httplib is a simple, low-level http client for Go

package httplib

import (
    "bytes"
    "http"
    "io"
    "net"
    "os"
    "strings"
)

var debugprint = true

type Client struct {
    conn    *http.ClientConn
    lastURL *http.URL
}

type nopCloser struct {
    io.Reader
}

func (nopCloser) Close() os.Error { return nil }

func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

func (client *Client) Request(rawurl string, method string, headers map[string]string, body string) (*http.Response, os.Error) {
    var url *http.URL
    var err os.Error
    if url, err = http.ParseURL(rawurl); err != nil {
        return nil, err
    }

    if headers == nil {
        headers = map[string]string{}
    }

    if client.conn == nil || client.lastURL.Host != url.Host {
        addr := url.Host
        if !hasPort(addr) {
            addr += ":http"
        }

        var tcpConn net.Conn
        if tcpConn, err = net.Dial("tcp", "", addr); err != nil {
            return nil, err
        }
        client.conn = http.NewClientConn(tcpConn, nil)
    }

    client.lastURL = url
    var req http.Request
    req.URL = url
    req.Method = method
    req.Header = headers
    req.UserAgent = headers["User-Agent"]
    if req.UserAgent == "" {
        req.UserAgent = "httplib.go"
    }
    req.Body = nopCloser{bytes.NewBufferString(body)}

    if debugprint {
        dump, _ := http.DumpRequest(&req, true)
        print(string(dump))
    }

    err = client.conn.Write(&req)
    if err != nil {
        return nil, err
    }

    resp, err := client.conn.Read()
    if err != nil {
        return nil, err
    }

    return resp, nil
}
