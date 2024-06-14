package preprocessing

import (
	"testing"
)

func TestHTML2Text(t *testing.T) {
	html := `<html><head><title>Test</title></head><body><h1>Test</h1><p>This is a test.</p><a href="http://www.google.com">google</a></body></html>`
	instance := NewHtml2Text(nil)
	text, url := instance.Parse(html)
	t.Log(text)
	t.Log(url)
}
