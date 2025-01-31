package busybox

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/rafaeljesus/retry-go"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

type Proxy struct {
	Host     string
	Port     string
	Username string
	Password string
}

func (proxy *Proxy) String() string {
	if proxy.Username == "" && proxy.Password == "" {
		return fmt.Sprintf("http://%s:%s", proxy.Host, proxy.Port)
	}
	return fmt.Sprintf("http://%s:%s@%s:%s", proxy.Username, proxy.Password, proxy.Host, proxy.Port)
}

func (proxy *Proxy) RawString() string {
	if proxy.Username == "" && proxy.Password == "" {
		return fmt.Sprintf("%s:%s", proxy.Host, proxy.Port)
	}
	return fmt.Sprintf("%s:%s:%s:%s", proxy.Host, proxy.Port, proxy.Username, proxy.Password)
}

type ProxyManager struct {
	proxies []Proxy
	Ava     bool
}

var ProxyM = &ProxyManager{}
var HostName string

func (proxyManager *ProxyManager) Init() {
	if proxyManager.Ava {
		return
	}
	fileContent, err := Read("proxies.txt")
	if err != nil {
		log.Fatal("PROXY LOAD FAILED!", err)
	}
	proxyLines := strings.Split(fileContent, "\n")
	for line := range proxyLines {
		proxySegs := strings.Split(proxyLines[line], ":")
		if len(proxySegs) == 4 {
			proxy := Proxy{
				Host:     strings.TrimRight(proxySegs[0], "\r\n"),
				Port:     strings.TrimRight(proxySegs[1], "\r\n"),
				Username: strings.TrimRight(proxySegs[2], "\r\n"),
				Password: strings.TrimRight(proxySegs[3], "\r\n"),
			}
			proxyManager.proxies = append(proxyManager.proxies, proxy)
		} else if len(proxySegs) == 2 {

			proxy := Proxy{
				Host: strings.TrimRight(proxySegs[0], "\r\n"),
				Port: strings.TrimRight(proxySegs[1], "\r\n"),
			}
			proxyManager.proxies = append(proxyManager.proxies, proxy)
		}
	}
	proxyManager.Ava = true

}

func getRandomDigit(n int) int {
	randomDigit := rand.Intn(n)
	return randomDigit
}

func (proxyManager *ProxyManager) GetRandomOne() Proxy {
	if !proxyManager.Ava {
		proxyManager.Init()
	}
	length := len(proxyManager.proxies)
	if length == 0 {
		log.Fatal("ERR, CHECK YOUR proxies.txt")
	}
	index := getRandomDigit(length)
	return proxyManager.proxies[index]
}

func (proxyManager *ProxyManager) GetOne(index int) Proxy {
	if !proxyManager.Ava {
		proxyManager.Init()
	}
	return proxyManager.proxies[index]
}

type HttpResponse struct {
	Method      string
	Url         string
	StartTime   time.Time
	StatusCode  int
	Elapsed     time.Duration
	Resp        *http.Response
	AttemptsNum int
}

type RetryConfig struct {
	Attempts  int           //最大重试次数
	SleepTime time.Duration //重试延迟时间
	ErrorCode int
}

type HttpClient struct {
	client        *http.Client
	Proxy         *Proxy
	AllowRedirect bool
	Timeout       int
	Retry         RetryConfig
	Debug         bool
}

type ClientConfig struct {
	Proxy          *Proxy
	FollowRedirect bool
	Retry          RetryConfig
	Timeout        int
	Debug          bool
}

func NewHttpClient(config ClientConfig) *HttpClient {

	httpClient := HttpClient{
		Proxy:         config.Proxy,
		AllowRedirect: config.FollowRedirect,
		Retry:         config.Retry,
		Timeout:       config.Timeout,
		Debug:         config.Debug || os.Getenv("HTTP_DEBUG") == "true",
	}
	gCurCookiejar, _ := cookiejar.New(nil)
	httpClient.client = &http.Client{Timeout: time.Duration(httpClient.Timeout) * time.Second, Jar: gCurCookiejar}
	if httpClient.Proxy != nil {
		httpClient.SetProxy(httpClient.Proxy)
	}

	if !httpClient.AllowRedirect {
		httpClient.client.CheckRedirect = myCheckRedirect
	}
	return &httpClient
}

func (httpClient *HttpClient) SetProxy(proxy *Proxy) {
	httpClient.Proxy = proxy
	proxyUrl, err := url.Parse(httpClient.Proxy.String())
	if err != nil {
		log.Println("Warning: proxy is not valid!", err)
	}
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxyUrl),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient.client.Transport = tr
}

func (httpClient *HttpClient) GetCookieString(host string) string {
	clientCookies := httpClient.client.Jar.Cookies(&url.URL{Scheme: "https", Host: host, Path: "/"})
	cookieStrings := make([]string, 0)
	for _, v := range clientCookies {
		cookieStrings = append(cookieStrings, v.String())
	}
	return strings.Join(cookieStrings, ";")
}

func (httpClient *HttpClient) GetCookies(host string) []map[string]string {
	clientCookies := httpClient.client.Jar.Cookies(&url.URL{Scheme: "https", Host: host, Path: "/"})
	cookieStrings := make([]map[string]string, 0)
	for _, v := range clientCookies {
		cookieStrings = append(cookieStrings, map[string]string{"name": v.Name, "value": v.Value})
	}
	return cookieStrings
}

func (httpClient *HttpClient) Request(Method, destination string, header http.Header, data interface{}) (*HttpResponse, error) {

	httpResponse := HttpResponse{
		Method:    Method,
		Url:       destination,
		StartTime: time.Now(),
	}
	defer httpClient.RequestTrack(&httpResponse)
	defer func() {
		r := recover()
		switch r.(type) {
		case http.Response:
			resp := r.(http.Response)
			httpResponse.StatusCode = resp.StatusCode
		}
	}()
	var body io.Reader
	switch data.(type) {
	case string:
		if val, ok := data.(string); ok {
			body = strings.NewReader(val)
		} else {
			return &httpResponse, errors.New("value error")
		}

	case []byte:
		if val, ok := data.([]byte); ok {
			body = bytes.NewReader(val)
		} else {
			return &httpResponse, errors.New("value error")
		}

	default:
		return &httpResponse, errors.New("value error")
	}

	request, err := http.NewRequest(Method, destination, body)
	if err != nil {
		fmt.Println(err)
		return &httpResponse, err
	}
	request.Header = header
	response, err := retry.DoHTTP(func() (*http.Response, error) {
		return makeRequest(httpClient, request, &httpResponse)
	}, httpClient.Retry.Attempts, httpClient.Retry.SleepTime)

	if response != nil {
		httpResponse.Elapsed += time.Since(httpResponse.StartTime)
		httpResponse.StatusCode = response.StatusCode
		httpResponse.Resp = response
	}

	if err != nil {
		return &httpResponse, err
	}
	return &httpResponse, nil
}

func (httpClient *HttpClient) Post(destination string, header http.Header, data interface{}) (*HttpResponse, error) {
	return httpClient.Request("POST", destination, header, data)
}

func (httpClient *HttpClient) Patch(destination string, header http.Header, data interface{}) (*HttpResponse, error) {
	return httpClient.Request("PATCH", destination, header, data)
}

func (httpClient *HttpClient) Put(destination string, header http.Header, data interface{}) (*HttpResponse, error) {
	return httpClient.Request("PUT", destination, header, data)
}
func (httpClient *HttpClient) RequestTrack(response *HttpResponse) {
	if httpClient.Debug == false {
		return
	}

	log.Println(fmt.Sprintf("%s %s STATUS CODE:%v COST:%s ATTEMPTS:%v", HostName, response.Url, response.StatusCode, response.Elapsed.Truncate(time.Millisecond), response.AttemptsNum))
}

func (httpResponse *HttpResponse) Json() gjson.Result {
	if httpResponse.Resp == nil || httpResponse.Resp.Body == nil {
		return gjson.Result{}
	}
	content, err := io.ReadAll(httpResponse.Resp.Body)
	if err != nil {
		log.Println(err)
		return gjson.Result{}
	}
	defer httpResponse.Resp.Body.Close()
	return gjson.Parse(string(content))
}

func (httpResponse *HttpResponse) Text() string {
	if httpResponse.Resp == nil || httpResponse.Resp.Body == nil {
		return ""
	}
	content, err := io.ReadAll(httpResponse.Resp.Body)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer httpResponse.Resp.Body.Close()
	return string(content)
}

func makeRequest(httpClient *HttpClient, request *http.Request, httpResponse *HttpResponse) (*http.Response, error) {
	httpResponse.AttemptsNum += 1
	response, err := httpClient.client.Do(request)

	if err != nil {
		return response, err
	}

	if response.StatusCode == httpClient.Retry.ErrorCode {
		return response, errors.New(fmt.Sprintf("%v Occurred", httpClient.Retry.ErrorCode))
	}
	return response, err
}

func (httpClient *HttpClient) SetCookies(destination string, cookies map[string]string) {
	httpCookies := make([]*http.Cookie, 0)
	for k, v := range cookies {
		cookie := &http.Cookie{Name: k, Value: v}
		httpCookies = append(httpCookies, cookie)
	}
	urlObj, _ := url.Parse(destination)

	httpClient.client.Jar.SetCookies(urlObj, httpCookies)

}

func (httpClient *HttpClient) SetCookie(destination string, cookies []*http.Cookie) {
	urlObj, _ := url.Parse(destination)

	httpClient.client.Jar.SetCookies(urlObj, cookies)

}

func (httpClient *HttpClient) Get(destination string, header http.Header) (*HttpResponse, error) {

	httpResponse := HttpResponse{
		Method:    "GET",
		Url:       destination,
		StartTime: time.Now(),
	}
	defer httpClient.RequestTrack(&httpResponse)
	var body io.Reader
	request, err := http.NewRequest("GET", destination, body)

	request.Header = header
	response, err := retry.DoHTTP(func() (*http.Response, error) {
		return makeRequest(httpClient, request, &httpResponse)
	}, httpClient.Retry.Attempts, httpClient.Retry.SleepTime)
	if response != nil {
		httpResponse.Elapsed += time.Since(httpResponse.StartTime)
		httpResponse.StatusCode = response.StatusCode
		httpResponse.Resp = response
	}

	if err != nil {
		return &httpResponse, err
	}

	if err != nil {
		return &httpResponse, err
	}

	return &httpResponse, nil
}

func myCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 1 {
		return errors.New("stop redirects")
	}
	return nil
}

func init() {
	HostName, _ = os.Hostname()
}
