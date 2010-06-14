// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// httplib is a simple, low-level http client for Go

package httplib

import (
    "bufio"
    "bytes"
    "container/vector"
    "fmt"
    "http"
    "io"
    "net"
    "os"
    "strconv"
    "strings"
)

var debugprint = false

type Client struct {
    conn    net.Conn
    lastURL *http.URL
}

type Request struct {
    URL     *http.URL
    Method  string
    Headers map[string]string
    Body    string
}

type Response struct {
    Status  int
    Headers map[string][]string
    Files   map[string][]byte
    Body    io.Reader
}

func (r *Response) getHeader(key string) (value string) {
    if lst, ok := r.Headers[key]; ok {
        value = lst[0]
    }

    return
}

// Used in Send to implement io.ReadCloser by bundling together the
// io.BufReader through which we read the response, and the underlying
// network connection.
type readClose struct {
    io.Reader
    io.Closer
}

func readLineBytes(b *bufio.Reader) (p []byte, err os.Error) {
    if p, err = b.ReadSlice('\n'); err != nil {
        // We always know when EOF is coming.
        // If the caller asked for a line, there should be a line.
        if err == os.EOF {
            err = io.ErrUnexpectedEOF
        }
        return nil, err
    }
    if len(p) >= 1024 {
        return nil, os.NewError("Header Line too long")
    }

    // Chop off trailing white space.
    var i int
    for i = len(p); i > 0; i-- {
        if c := p[i-1]; c != ' ' && c != '\r' && c != '\t' && c != '\n' {
            break
        }
    }
    return p[0:i], nil
}

func readLine(b *bufio.Reader) (s string, err os.Error) {
    p, e := readLineBytes(b)
    if e != nil {
        return "", e
    }
    return string(p), nil
}

// Read a key/value pair from b.
// A key/value has the form Key: Value\r\n
// and the Value can continue on multiple lines if each continuation line
// starts with a space.
func readKeyValue(b *bufio.Reader) (key, value string, err os.Error) {
    line, e := readLineBytes(b)
    if e != nil {
        return "", "", e
    }
    if len(line) == 0 {
        return "", "", nil
    }

    // Scan first line for colon.
    i := bytes.Index(line, []byte{':'})
    if i < 0 {
        goto Malformed
    }

    key = string(line[0:i])
    if strings.Index(key, " ") >= 0 {
        // Key field has space - no good.
        goto Malformed
    }

    // Skip initial space before value.
    for i++; i < len(line); i++ {
        if line[i] != ' ' {
            break
        }
    }
    value = string(line[i:])

    // Look for extension lines, which must begin with space.
    for {
        c, e := b.ReadByte()
        if c != ' ' {
            if e != os.EOF {
                b.UnreadByte()
            }
            break
        }

        // Eat leading space.
        for c == ' ' {
            if c, e = b.ReadByte(); e != nil {
                if e == os.EOF {
                    e = io.ErrUnexpectedEOF
                }
                return "", "", e
            }
        }
        b.UnreadByte()

        // Read the rest of the line and add to value.
        if line, e = readLineBytes(b); e != nil {
            return "", "", e
        }
        value += " " + string(line)

        if len(value) >= 1024 {
            return "", "", os.NewError("value too long for key " + key)
        }
    }
    return key, value, nil

Malformed:
    return "", "", os.NewError("malformed header line " + string(line))
}

type chunkedReader struct {
    r   *bufio.Reader
    n   uint64 // unread bytes in chunk
    err os.Error
}

func newChunkedReader(r *bufio.Reader) *chunkedReader {
    return &chunkedReader{r: r}
}

func (cr *chunkedReader) beginChunk() {
    // chunk-size CRLF
    var line string
    line, cr.err = readLine(cr.r)
    if cr.err != nil {
        return
    }
    cr.n, cr.err = strconv.Btoui64(line, 16)
    if cr.err != nil {
        return
    }
    if cr.n == 0 {
        // trailer CRLF
        for {
            line, cr.err = readLine(cr.r)
            if cr.err != nil {
                return
            }
            if line == "" {
                break
            }
        }
        cr.err = os.EOF
    }
}

func (cr *chunkedReader) Read(b []uint8) (n int, err os.Error) {
    if cr.err != nil {
        return 0, cr.err
    }
    if cr.n == 0 {
        cr.beginChunk()
        if cr.err != nil {
            return 0, cr.err
        }
    }
    if uint64(len(b)) > cr.n {
        b = b[0:cr.n]
    }
    n, cr.err = cr.r.Read(b)
    cr.n -= uint64(n)
    if cr.n == 0 && cr.err == nil {
        // end of chunk (CRLF)
        b := make([]byte, 2)
        if _, cr.err = io.ReadFull(cr.r, b); cr.err == nil {
            if b[0] != '\r' || b[1] != '\n' {
                cr.err = os.NewError("malformed chunked encoding")
            }
        }
    }
    return n, cr.err
}

func readResponse(r *bufio.Reader) (*Response, os.Error) {
    resp := new(Response)

    // Parse the first line of the response.
    resp.Headers = make(map[string][]string)

    line, err := readLine(r)
    if err != nil {
        return nil, err
    }
    f := strings.Split(line, " ", 3)
    if len(f) < 3 {
        return nil, os.NewError("malformed HTTP response:" + line)
    }
    resp.Status, err = strconv.Atoi(f[1])
    if err != nil {
        return nil, os.NewError("malformed HTTP status code")
    }

    // Parse the response headers.
    for {
        key, value, err := readKeyValue(r)
        if err != nil {
            return nil, err
        }
        if key == "" {
            break // end of response header
        }
        if _, ok := resp.Headers[key]; !ok {
            resp.Headers[key] = []string{}
        }
        vec := vector.StringVector(resp.Headers[key])
        vec.Push(value)
        resp.Headers[key] = vec
    }

    return resp, nil
}


func (req *Request) Write(buf io.Writer) (err os.Error) {

    uri := req.URL.Path

    if req.URL.RawQuery != "" {
        uri += "?" + req.URL.RawQuery
    }

    if _, err = fmt.Fprintf(buf, "%s %s HTTP/1.1\r\n", req.Method, uri); err != nil {
        return
    }

    for name, val := range (req.Headers) {
        if _, err = fmt.Fprintf(buf, "%s: %s\r\n", name, val); err != nil {
            return
        }
    }

    if _, err = fmt.Fprintf(buf, "\r\n"); err != nil {
        return
    }

    if cl, ok := req.Headers["Content-Length"]; ok {
        //check if body-length
        size, _ := strconv.Atoi(cl)
        if _, err = buf.Write([]byte(req.Body[0:size])); err != nil {
            return
        }
    }

    return nil

}

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

func (client *Client) Request(rawurl string, method string, headers map[string]string, body string) (resp *Response, err os.Error) {

    var url *http.URL

    if url, err = http.ParseURL(rawurl); err != nil {
        return nil, err
    }

    if headers == nil {
        headers = map[string]string{}
    }

    if _,ok := headers["User-Agent"]; !ok {
	headers["User-Agent"]= "httplib.go"
    }

    if _,ok := headers["Host"]; !ok {
	headers["Host"]= url.Host
    }


    if client.conn == nil || client.lastURL.Host != url.Host {
        addr := url.Host
        if !hasPort(addr) {
            addr += ":http"
        }

        var conn net.Conn
        if conn, err = net.Dial("tcp", "", addr); err != nil {
            return nil, err
        }
        client.conn = conn
    }

    client.lastURL = url
    req := Request{url, method, headers, body}

    if debugprint {
        var buf bytes.Buffer
        req.Write(&buf)
        fmt.Printf("%#v\n", buf.String())
    }

    err = req.Write(client.conn)
    if err != nil {
        return nil, err
    }

    reader := bufio.NewReader(client.conn)

    resp, err = readResponse(reader)

    if err != nil {
        client.conn.Close()
        return nil, err
    }

    r := io.Reader(reader)
    if v := resp.getHeader("Transfer-Encoding"); v == "chunked" {
        r = newChunkedReader(reader)
    } else if v := resp.getHeader("Content-Length"); v != "" {
        n, err := strconv.Atoi64(v)
        if err != nil {
            return nil, os.NewError("invalid Content-Length : " + v)
        }
        r = io.LimitReader(r, n)
    }
    resp.Body = readClose{r, client.conn}

    return

    return nil, nil
}
