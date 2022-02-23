package runtime

import (
	"strings"
	"testing"
)

func TestVersionsXml01(t *testing.T) {
	data := `<versions><version id="0.0.1"/><version id="0.0.2"/><version id="0.0.3"/></versions>`
	versions, err := ParseVersionFile(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if len(versions.Version) != 3 {
		t.Fatal("miss expected versions amount 3")
	}
	for _, v := range versions.Version {
		if v.Id == "" {
			t.Fatal("version id is empty")
		}
	}
}
