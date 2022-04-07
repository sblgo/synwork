package awsprovider

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"
)

func TestHtmlStructRead01(t *testing.T) {
	b, err := ioutil.ReadFile("synwork-processor-randlist_0.0.1_linux_amd64.html")
	if err != nil {
		t.Error(err)
	}
	_, err = ReadHtmlEmbbededFile(context.Background(), bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}
}
