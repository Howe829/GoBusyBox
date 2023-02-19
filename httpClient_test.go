package busybox

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
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
	res, err := client.Get("https://www.ti.com.cn/secure-link-forward/?gotoUrl=https://www.ti.com.cn", headers)
	if err != nil {
		fmt.Println(err.Error())

	}
	reg := regexp.MustCompile(`form.setAttribute\('action', '(.*?)'\)`)
	text, err := res.Text()
	action := reg.FindAllStringSubmatch(text, -1)
	if len(action) < 1 {
		log.Fatal("Login Failed")
	}

	actionUrl := fmt.Sprintf("https://login.ti.com%s", action[0][1])
	headers.Set(HeaderContentType, ContentTypeUrlEncoded)
	postData := url.Values{}
	postData.Add("pf.adapterId", "IDPAdapterHTMLFormCIDStandard")
	resp, err := client.Post(actionUrl, headers, postData.Encode())
	if err != nil {
		log.Fatal(err)
		return
	}
	text, err = resp.Text()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(text)
	//if *res.StatusCode != 200 {
	//	fmt.Println(res)
	//} else {
	//	fmt.Println(res.Result.String())
	//}

}
