package httplib

import (
	"testing"
)

func TestFluidGet(t *testing.T) {
        s,err := Get("http://google.com").AsString()
        if err != nil {
                t.Fatalf(err.String())
        }
        if len(s) == 0 {
                t.Fatalf("No data available\n")
        }
}

func TestInvalid(t *testing.T) {
        _,err := Post("http://invalidurlsdfsdfasdfsdgf:9999/post").AsString()
        if err == nil {
            t.Fatalf("Expected an error!")
        }
}