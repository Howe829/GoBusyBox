package busybox

import (
	"fmt"
	"log"
	"net/http"
	"testing"
)

func TestHttpClient_Get(t *testing.T) {
	headers := http.Header{}
	headers.Set(HeaderUserAgent, UserAgentDefault)
	headers.Set(HeaderAccept, AcceptDefault)
	headers.Set(HeaderAcceptLanguage, AcceptLanguage)
	headers.Set(HeaderReferer, TiCn)
	headers.Set(HeaderUpgradeInsecureRequests, UpgradeInsecureRequests)
	headers.Set(HeaderSecFetchDest, SecFetchDestDocument)
	headers.Set(HeaderSecFetchMode, SecFetchModeNavigate)
	headers.Set(HeaderSecFetchSite, SecFetchSiteSameOrigin)
	headers.Set(HeaderSecFetchUser, SecFetchUserDefault)
	headers.Set(HeaderTe, TeTrailers)
	reTry := RetryConfig{ErrorCode: 429, Attempts: 3}

	client := NewHttpClient(false, false, reTry, 0)
	cookies := map[string]string{
		"JSESSIONID":     "50C97B005A48F95865E8ABF2BE8CBB90.web01",
		"lang":           "en",
		"cl-font-size":   "16px",
		"cl-theme-color": "default",
		"_ga":            "GA1.2.1334036334.1676380575",
		"_ga_0M1K5NPYZE": "GS1.1.1677677203.6.0.1677677330.0.0.0",
		"_ga_BD9VNGC0M6": "GS1.1.1677677203.6.0.1677677330.0.0.0",
		"AWSALB":         "/kX0tLxKrdnJGgaIuD/fxjBhg7PGcHclMM7AAiE7nHJxaugW0p4/b84lpsi/+l79n3qiGpytl3cgoFhTK00yCqPP9CEtw/iCLxXAF6qJAHLz8j5HpdpS0tc1prp3",
		"AWSALBCORS":     "/kX0tLxKrdnJGgaIuD/fxjBhg7PGcHclMM7AAiE7nHJxaugW0p4/b84lpsi/+l79n3qiGpytl3cgoFhTK00yCqPP9CEtw/iCLxXAF6qJAHLz8j5HpdpS0tc1prp3",
	}
	client.SetCookie("https://httpbin.org/cookies", cookies)
	res, err := client.Get("https://httpbin.org/cookies", headers)
	if err != nil {
		fmt.Println(err.Error())
	}
	log.Println(res.Text(), res.StatusCode)

	//if *res.StatusCode != 200 {
	//	fmt.Println(res)
	//} else {
	//	fmt.Println(res.Result.String())
	//}

}
