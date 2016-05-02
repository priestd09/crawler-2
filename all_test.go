package crawler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testController struct {
	OnceController
	text  chan string
	value chan []string
}

func (t testController) Handle(r *Response, _ chan<- *Link) {
	t.text <- r.FindText("div.foo")
	t.value <- r.FindAttr("div#hello", "key")
}

func TestAll(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, `
<html>
<head></head>
<body>
<div class="foo">bar</div>
<div id="hello" key="value">Hello, world!</div>
</body>
</html>`)
	}))
	defer ts.Close()

	ctrl := &testController{
		text:  make(chan string, 2),
		value: make(chan []string, 2),
	}

	cw := NewCrawler(&Config{Controller: ctrl})
	assert.Nil(cw.Crawl(ts.URL))
	cw.Wait()
	assert.Equal("bar", <-ctrl.text)
	vs := <-ctrl.value
	assert.Equal(1, len(vs))
	assert.Equal("value", vs[0])

	u, err := url.Parse(ts.URL)
	assert.Nil(err)
	uu, _ := cw.store.Get(u)
	assert.Equal(1, uu.VisitCount)
	assert.True(uu.Last.After(time.Time{}))
}
