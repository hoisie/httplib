package httplib

import (
	"io/ioutil"
	"testing"
)

func TestBasic(t *testing.T) {
	client := new(Client)
	resp,err := client.Request("http://google.com/", "GET", nil, "")
	if err != nil {
		t.Fatalf(err.String())
	}
	data,err := ioutil.ReadAll( resp.Body )
	if err != nil {
		t.Fatalf(err.String())
	}
	
	if len(data) == 0 {
		t.Fatalf("No data available\n")
	}
}

func TestFluidGet(t *testing.T) {
        s,err := Get("http://localhost:9999/get").Param("a", "1").Param("b", "2").AsString()
        if err != nil {
                t.Fatalf(err.String())
        }
        if len(s) == 0 {
                t.Fatalf("No data available\n")
        }
}


func TestFluidPost(t *testing.T) {
        s,err := Post("http://localhost:9999/post").Param("a", "1").Param("b", "2").AsString()
        if err != nil {
                t.Fatalf(err.String())
        }
        if len(s) == 0 {
                t.Fatalf("No data available\n")
        }
}

func TestInvalid(t *testing.T) {
        s,err := Post("http://invalidurlsdfsdfasdfsdgf:9999/post").AsString()
        if err != nil {
                t.Fatalf(err.String())
        }
        if len(s) == 0 {
                t.Fatalf("No data available\n")
        }
}