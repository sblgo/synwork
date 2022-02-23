package runtime

import (
	"encoding/xml"
	"io"
)

type Version struct {
	Id string `xml:"id,attr"`
}
type Versions struct {
	Version []Version `xml:"version"`
}

func ParseVersionFile(i io.Reader) (*Versions, error) {
	dec := xml.NewDecoder(i)
	versions := &Versions{}
	err := dec.Decode(versions)
	if err != nil {
		return nil, err
	} else {

		return versions, nil
	}
}
