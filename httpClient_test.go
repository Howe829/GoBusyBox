package busybox

import (
	"fmt"
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
	client := HttpClient{EnableProxy: true, AllowRedirect: true}
	res, err := client.Get("https://httpbin.org/ip", headers)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(res.Text())
	client.ChangeProxy()
	res, err = client.Get("https://httpbin.org/ip", headers)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(res.Text())

	//if *res.StatusCode != 200 {
	//	fmt.Println(res)
	//} else {
	//	fmt.Println(res.Result.String())
	//}

}
