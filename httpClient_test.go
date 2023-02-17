package busybox

import (
	"fmt"
	"testing"
)

func TestHttpClient_Get(t *testing.T) {
	client := HttpClient{EnableProxy: true}
	res, err := client.Get("https://httpbin.org/ip", nil)
	if err != nil {
		t.Fail()
	}
	if *res.StatusCode != 200 {
		t.Fail()
	} else {
		fmt.Println(res.Result.String())
	}

}
