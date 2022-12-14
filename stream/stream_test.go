package stream

import (
	"testing"
)

func TestParse(t *testing.T) {
	r, err := Parse("https://sblongvu.com/e/j1487gulo4df.html")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	err = r.GetQualityM3U8()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	t.Log(r.Download(""))
}
