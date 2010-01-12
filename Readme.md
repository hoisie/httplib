
## About
httplib.go is a simple http client library for Go. 

It supports all the features of go's http client package, as well as generic requests and keep-alive connections. 

## Usage

This is a small usage example:

    c := new(httplib.Client)
    resp, err := c.Request ("http://google.com", "GET", nil, "")
    data := ioutil.ReadAll( resp.Body )
    println(string(data))


