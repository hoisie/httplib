package httplib

import (
	"testing"
)

func TestFluidGet(t *testing.T) {
    query, err := Get("www.google.com/search").AsString()
    if err != nil {
            println(err.String())
    }
    println(query);
    
}

/*
func TestInvalid(t *testing.T) {
        _,err := Post("http://invalidurlsdfsdfasdfsdgf:9999/post").AsString()
        if err == nil {
            t.Fatalf("Expected an error!")
        }
}
*/