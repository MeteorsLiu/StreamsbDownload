package stream

import (
	"testing"
)

func TestParse(t *testing.T) {
	r, err := Parse("https://sblongvu.com/e/j1487gulo4df.html")
	if err != nil {
		t.Errorf(err)
		return
	}

	t.Log(r.String())
}