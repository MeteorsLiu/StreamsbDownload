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

	for _, v := range r.Items() {
		t.Log(v)
	}
}
