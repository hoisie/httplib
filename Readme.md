
## About
httplib.go is a simple extension of Go's http client that provides keep-alive connections and generic requests.

## Usage

This is a small usage example:

    //get the google home page
    c := new(httplib.Client)
    resp, err := c.Request ("http://google.com", "GET", nil, "")
    data := ioutil.ReadAll( resp.Body )
    println(string(data))

