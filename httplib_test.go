package httplib

import (
	"io/ioutil"
	"testing"
)

func TestSecureCookie(t *testing.T) {
	client := Client{}
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
