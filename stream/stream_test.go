package stream

import (
	"testing"
)

func TestParse(t *testing.T) {
	r, err := Parse("https://sblongvu.com/e/pynjd03yambo.html")
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	err = r.GetQualityM3U8()
	if err != nil {
		t.Errorf("%v", err)
		return
	}
	r.Download("/home/nfs/py/ttt.mp4")
}
