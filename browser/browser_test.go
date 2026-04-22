package browser

import (
	"fmt"
	"testing"
	"time"
)

func TestGetHTML(t *testing.T) {
	b, err := New(false)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if err := b.Navigate("https://duckduckgo.com"); err != nil {
		t.Log(err)
		t.FailNow()
	}

	html, err := b.GetHTML()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if len(html) < 1000 {
		t.FailNow()
	}

	currentURL, err := b.GetCurrentURL()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if currentURL != "https://duckduckgo.com/" {
		t.FailNow()
	}

	s, err := b.Snapshot()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	fmt.Println(s)

	if err := b.ClickByRef("e24"); err != nil {
		t.Log(err)
		t.FailNow()
	}

	time.Sleep(time.Second * 10)

	currentURL, err = b.GetCurrentURL()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	fmt.Println(currentURL)
}
