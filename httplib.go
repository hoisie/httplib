package httplib

import (
    "bytes"
    "crypto/tls"
    "http"
    "io"
    "io/ioutil"
    "net"
    "os"
    "strings"
    "url"
)

var defaultUserAgent = "httplib.go"

var debugprint = false

type Client struct {
    conn    *http.ClientConn
    lastURL *url.URL
}

type nopCloser struct {
    io.Reader
}

func (nopCloser) Close() os.Error { return nil }

func getNopCloser(buf *bytes.Buffer) nopCloser {
    return nopCloser{buf}
}

func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

func newConn(url *url.URL) (*http.ClientConn, os.Error) {
    addr := url.Host
    //just set the default scheme to http
    if url.Scheme == "" {
        url.Scheme = "http"
    }
    if !hasPort(addr) {
        addr += ":" + url.Scheme
    }
    var conn net.Conn
    var err os.Error
    if url.Scheme == "http" {
        conn, err = net.Dial("tcp", addr)
        if err != nil {
            return nil, err
        }
    } else { // https
        conn, err = tls.Dial("tcp", addr, nil)
        if err != nil {
            return nil, err
        }
        h := url.Host
        if hasPort(h) {
            h = h[0:strings.LastIndex(h, ":")]
        }
        if err := conn.(*tls.Conn).VerifyHostname(h); err != nil {
            return nil, err
        }
    }

    return http.NewClientConn(conn, nil), nil
}

func getResponse(rawUrl string, req *http.Request) (*http.ClientConn, *http.Response, os.Error) {
    url, err := url.Parse(rawUrl)
    if url.Scheme == "" {
        rawUrl = "http://" + rawUrl
        url, err = url.Parse(rawUrl)
    }

    if err != nil {
        return nil, nil, err
    }
    req.URL = url
    if debugprint {
        dump, err := http.DumpRequest(req, true)
        if err != nil {
            println(err.String())
        }
        print(string(dump))
    }

    conn, err := newConn(url)
    if err != nil {
        println(err.String())
        return nil, nil, err
    }

    resp, err := conn.Do(req)
    if err != nil {
        if err != http.ErrPersistEOF {
            return nil, nil, err
        }
    }
    return conn, resp, nil
}

func Get(url string) *HttpRequestBuilder {
    var req http.Request
    req.Method = "GET"
    req.Header = http.Header{}
    req.Header.Set("User-Agent", defaultUserAgent)
    return &HttpRequestBuilder{url, &req, nil, map[string]string{}}
}

func Post(url string) *HttpRequestBuilder {
    var req http.Request
    req.Method = "POST"
    req.Header = http.Header{}
    req.Header.Set("User-Agent", defaultUserAgent)
    return &HttpRequestBuilder{url, &req, nil, map[string]string{}}
}

func Put(url string) *HttpRequestBuilder {
    var req http.Request
    req.Method = "PUT"
    req.Header = http.Header{}
    req.Header.Set("User-Agent", defaultUserAgent)
    return &HttpRequestBuilder{url, &req, nil, map[string]string{}}
}

func Delete(url string) *HttpRequestBuilder {
    var req http.Request
    req.Method = "DELETE"
    req.Header = http.Header{}
    req.Header.Set("User-Agent", defaultUserAgent)
    return &HttpRequestBuilder{url, &req, nil, map[string]string{}}
}

type HttpRequestBuilder struct {
    url        string
    req        *http.Request
    clientConn *http.ClientConn
    params     map[string]string
}

func (b *HttpRequestBuilder) getResponse() (*http.Response, os.Error) {
    var paramBody string
    if b.params != nil && len(b.params) > 0 {
        var buf bytes.Buffer
        for k, v := range b.params {
            buf.WriteString(url.QueryEscape(k))
            buf.WriteByte('=')
            buf.WriteString(url.QueryEscape(v))
            buf.WriteByte('&')
        }
        paramBody = buf.String()
        paramBody = paramBody[0 : len(paramBody)-1]
    }
    if b.req.Method == "GET" && len(paramBody) > 0 {
        if strings.Index(b.url, "?") != -1 {
            b.url += "&" + paramBody
        } else {
            b.url = b.url + "?" + paramBody
        }
    } else if b.req.Method == "POST" && b.req.Body == nil && len(paramBody) > 0 {
        b.Header("Content-Type", "application/x-www-form-urlencoded")
        b.req.Body = nopCloser{bytes.NewBufferString(paramBody)}
        b.req.ContentLength = int64(len(paramBody))
    }

    conn, resp, err := getResponse(b.url, b.req)
    b.clientConn = conn
    return resp, err
}

func (b *HttpRequestBuilder) Header(key, value string) *HttpRequestBuilder {
    b.req.Header.Set(key, value)
    return b
}

func (b *HttpRequestBuilder) Param(key, value string) *HttpRequestBuilder {
    b.params[key] = value
    return b
}

func (b *HttpRequestBuilder) Body(data interface{}) *HttpRequestBuilder {
    switch t := data.(type) {
    case string:
        b.req.Body = getNopCloser(bytes.NewBufferString(t))
        b.req.ContentLength = int64(len(t))
    case []byte:
        b.req.Body = getNopCloser(bytes.NewBuffer(t))
        b.req.ContentLength = int64(len(t))
    }
    return b
}

func (b *HttpRequestBuilder) AsString() (string, os.Error) {
    resp, err := b.getResponse()
    if err != nil {
        return "", err
    }
    if resp.Body == nil {
        return "", nil
    }
    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(data), nil
}

func (b *HttpRequestBuilder) AsBytes() ([]byte, os.Error) {
    resp, err := b.getResponse()
    if err != nil {
        return nil, err
    }
    if resp.Body == nil {
        return nil, nil
    }
    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return data, nil
}

func (b *HttpRequestBuilder) AsFile(filename string) os.Error {
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()

    resp, err := b.getResponse()
    if err != nil {
        return err
    }
    if resp.Body == nil {
        return nil
    }
    _, err = io.Copy(f, resp.Body)
    if err != nil {
        return err
    }
    return nil
}

func (b *HttpRequestBuilder) AsResponse() (*http.Response, os.Error) {
    return b.getResponse()
}

func (b *HttpRequestBuilder) Close() {
    if b.clientConn != nil {
        b.clientConn.Close()
    }
}
