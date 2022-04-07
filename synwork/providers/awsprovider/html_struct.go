package awsprovider

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"golang.org/x/net/html"
)

type (
	HtmlTag struct {
		Body *BodyTag `xml:"body"`
	}

	BodyTag struct {
		Ptags []*PTag `xml:"p"`
	}

	PTag struct {
		Class   string `xml:"class,attr"`
		Content string `xml:",cdata"`
	}
)

func ReadHtmlEmbbededFile(ctx context.Context, r io.Reader) ([]byte, error) {
	htmlTag, err := parseHtmlTag(r)
	if err != nil {
		return nil, err
	}
	readers := make([]io.Reader, len(htmlTag.Body.Ptags))
	for i, p := range htmlTag.Body.Ptags {
		if p.Class == "snw" {
			readers[i] = strings.NewReader(p.Content)
		}
	}
	bDec := base64.NewDecoder(base64.StdEncoding, io.MultiReader(readers...))
	bBuf := new(bytes.Buffer)
	bBuf.ReadFrom(bDec)
	b := bBuf.Bytes()
	zipReader, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}
	files := zipReader.File
	if len(files) != 1 {
		return nil, fmt.Errorf("unexpected content")
	}
	f, err := zipReader.Open(files[0].Name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func isSNW(as []html.Attribute) bool {
	for _, a := range as {
		if a.Key == "class" && a.Val == "snw" {
			return true
		}
	}
	return false
}

func parseHtmlTag(r io.Reader) (*HtmlTag, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	hTag := &HtmlTag{
		Body: &BodyTag{
			Ptags: []*PTag{},
		},
	}
	body := doc.FirstChild.FirstChild
	for body != nil && body.Data != "body" {
		body = body.NextSibling
	}
	if body == nil {
		return nil, fmt.Errorf("no body tag")
	}

	for pTag := body.FirstChild; pTag != nil; pTag = pTag.NextSibling {
		if pTag.Data == "p" && isSNW(pTag.Attr) {
			hTag.Body.Ptags = append(hTag.Body.Ptags, &PTag{
				Class:   "snw",
				Content: pTag.FirstChild.Data,
			})
		}
	}
	return hTag, nil
}
