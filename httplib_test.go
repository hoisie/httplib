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

func TestFluid(t *testing.T) {
        s,err := Get("http://www.google.com").AsString()
        if err != nil {
                t.Fatalf(err.String())
        }
        if len(s) == 0 {
                t.Fatalf("No data available\n")
        }
}

