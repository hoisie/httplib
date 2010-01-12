
## About
httplib.go is a simple http client library for Go. It will probably look like python's [httplib2](http://code.google.com/p/httplib2/wiki/Examples)

## Usage

This is a small usage example:

    c := new(httplib.Client)
    resp, err := c.Request ("http://google.com", "GET", nil, "")
    data := ioutil.ReadAll( resp.Body )
    println(string(data))


