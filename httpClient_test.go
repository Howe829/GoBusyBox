package busybox

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"
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

	client := NewHttpClient(ClientConfig{Retry: reTry, Debug: true})
	cookies := []*http.Cookie{
		{
			Name:     "session_id",
			Value:    "abc123",
			Domain:   ".httpbin.org", // Domain for which the cookie is valid
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour), // Expiry set to 24 hours
			HttpOnly: true,
			Secure:   true,
		},
		{
			Name:     "user_pref",
			Value:    "dark_mode",
			Domain:   ".httpbin.org", // Same domain for the cookie
			Path:     "/",
			Expires:  time.Now().Add(48 * time.Hour), // Expiry set to 48 hours
			HttpOnly: false,
			Secure:   false, // Could be false for non-HTTPS traffic
		},
		{
			Name:     "tracking_id",
			Value:    "xyz456",
			Domain:   ".httpbin.org",
			Path:     "/",
			Expires:  time.Now().Add(72 * time.Hour), // Expiry set to 72 hours
			HttpOnly: false,
			Secure:   true,
		},
	}
	client.SetCookie("https://httpbin.org/", cookies)
	resp, err := client.Get("https://httpbin.org/cookies", headers)
	if err != nil {
		fmt.Println(err.Error())
	}
	cookieString := client.GetCookieString("httpbin.org")

	log.Println(cookieString, client.ProxyStr)
	log.Println(resp.Text())
	//if *res.StatusCode != 200 {
	//	fmt.Println(res)
	//} else {
	//	fmt.Println(res.Result.String())
	//}

}
