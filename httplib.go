package httplib

import (
    "bytes"
    "fmt"
    "http"
    "io"
    "io/ioutil"
    "net"
    "os"
    "strings"
)

var defaultUserAgent = "httplib.go"

var debugprint = false

type Client struct {
    conn    *http.ClientConn
    lastURL *http.URL
}

type nopCloser struct {
    io.Reader
}

func (nopCloser) Close() os.Error { return nil }

func getNopCloser(s string) nopCloser {
    return nopCloser{bytes.NewBufferString(s)}
}

func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

func newConn(url *http.URL) (*http.ClientConn, os.Error) {
    addr := url.Host
    if !hasPort(addr) {
        addr += ":http"
    }
    tcpConn, err := net.Dial("tcp", "", addr)
    if err != nil {
        return nil, err
    }

    return http.NewClientConn(tcpConn, nil), nil
}

func getResponse(req *http.Request) (*http.Response, os.Error) {
    println("url", req.RawURL)
    url, err := http.ParseURL(req.RawURL)
    if err != nil {
        return nil, err
    }
    req.URL = url

    conn, err := newConn(url)
    if err != nil {
        return nil, err
    }

    err = conn.Write(req)
    if err != nil {
        return nil, err
    }

    resp, err := conn.Read()
    if err != nil {
        return nil, err
    }
    return resp, nil
}


func (client *Client) Request(rawurl string, method string, headers map[string]string, body string) (*http.Response, os.Error) {
    var url *http.URL
    var err os.Error
    if url, err = http.ParseURL(rawurl); err != nil {
        return nil, err
    }

    if client.conn == nil || client.lastURL.Host != url.Host {
        client.conn, err = newConn(url)
    }

    if headers == nil {
        headers = map[string]string{}
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

type RequestBuilder interface {
    Header(key, value string) RequestBuilder
    Param(key, value string) RequestBuilder
    Body(data interface{}) RequestBuilder
    AsString() (string, os.Error)
    AsFile(filename string) os.Error
    AsResponse() (*http.Response, os.Error)
}

func Get(url string) RequestBuilder {
    var req http.Request
    req.RawURL = url
    req.Method = "GET"
    req.Header = map[string]string{}
    req.UserAgent = defaultUserAgent
    return &HttpGetRequestBuilder{url, &req}
}

type HttpGetRequestBuilder struct {
    url string
    req *http.Request
}

func (b *HttpGetRequestBuilder) Header(key, value string) RequestBuilder {
    b.req.Header[key] = value
    return b
}

func (b *HttpGetRequestBuilder) Param(key, value string) RequestBuilder {
    if strings.Index(b.req.RawURL, "?") != -1 {
        b.req.RawURL = b.req.RawURL + fmt.Sprintf("&%s=%s", http.URLEscape(key), http.URLEscape(value))
    } else {
        b.req.RawURL = b.req.RawURL + fmt.Sprintf("?%s=%s", http.URLEscape(key), http.URLEscape(value))
    }
    return b
}

func (b *HttpGetRequestBuilder) Body(data interface{}) RequestBuilder {
    switch t := data.(type) {
    case string:
        b.req.Body = getNopCloser(t)
    }
    return b
}

func (b *HttpGetRequestBuilder) AsString() (string, os.Error) {
    resp, err := getResponse(b.req)

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(data), nil
}

func (b *HttpGetRequestBuilder) AsFile(filename string) os.Error {
    f, err := os.Open(filename, os.O_RDWR|os.O_CREATE, 0644)
    if err != nil {
        return err
    }
    resp, err := getResponse(b.req)
    _, err = io.Copy(f, resp.Body)
    if err != nil {
        return err
    }
    return nil
}

func (b *HttpGetRequestBuilder) AsResponse() (*http.Response, os.Error) {
    return getResponse(b.req)
}
