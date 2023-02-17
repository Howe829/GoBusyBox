package busybox

import (
	"fmt"
	"net/http"
	"testing"
)

func TestHttpClient_Get(t *testing.T) {
	client := HttpClient{EnableProxy: true}
	header := http.Header{}
	header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:106.0) Gecko/20100101 Firefox/106.0")
	res, err := client.Get("https://httpbin.org/headers", header)
	if err != nil {
		t.Fail()
	}
	if *res.StatusCode != 200 {
		t.Fail()
	} else {
		fmt.Println(res.Result.String())
	}

}
