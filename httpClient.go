package busybox

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ProxyManager struct {
	proxies []string
	Ava     bool
}

var ProxyM = &ProxyManager{}

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
		if len(proxySegs) < 4 {
			fmt.Println(fmt.Sprintf("PROXY %s LINE: %d FORMAT ERROR WILL BE IGNORED", proxyLines[line], line))
			continue
		}
		proxy := fmt.Sprintf("http://%s:%s@%s:%s", proxySegs[2], proxySegs[3], proxySegs[0], proxySegs[1])
		proxyManager.proxies = append(proxyManager.proxies, proxy)
	}
	proxyManager.Ava = true

}

func (proxyManager *ProxyManager) GetRandomOne() string {
	if !proxyManager.Ava {
		proxyManager.Init()
	}
	rand.Seed(time.Now().Unix())

	index := rand.Intn(len(proxyManager.proxies))
	return proxyManager.proxies[index]
}

func (proxyManager *ProxyManager) GetOne(index int) string {
	if !proxyManager.Ava {
		proxyManager.Init()
	}
	return proxyManager.proxies[index]
}

type HttpResponse struct {
	Method     string
	Url        string
	StartTime  time.Time
	StatusCode *int
	Result     gjson.Result
	Elapsed    time.Duration
	Resp       *http.Response
}

type HttpClient struct {
	EnableProxy   bool
	ava           bool
	client        *http.Client
	ProxyStr      string
	AllowRedirect bool
	Timeout       int
}

func (httpClient *HttpClient) Init() {
	if httpClient.Timeout == 0 {
		httpClient.Timeout = 60
	}
	httpClient.client = &http.Client{Timeout: time.Duration(httpClient.Timeout) * time.Second, CheckRedirect: myCheckRedirect}
	if httpClient.EnableProxy {
		httpClient.ProxyStr = ProxyM.GetRandomOne()
		proxy, _ := url.Parse(httpClient.ProxyStr)
		tr := &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient.client.Transport = tr
	}

	if !httpClient.AllowRedirect {
		httpClient.client.CheckRedirect = myCheckRedirect
	}
	httpClient.ava = true
}

func (httpClient *HttpClient) Post(destination string, header http.Header, data interface{}) (*HttpResponse, error) {
	if !httpClient.ava {
		httpClient.Init()
	}
	statusCode := -1
	httpResponse := HttpResponse{
		Method:     "POST",
		StatusCode: &statusCode,
		Url:        destination,
		StartTime:  time.Now(),
		Result:     gjson.Result{},
		Elapsed:    9999,
	}
	defer RequestTrack(httpResponse)
	defer func() {
		r := recover()
		switch r.(type) {
		case http.Response:
			resp := r.(http.Response)
			httpResponse.StatusCode = &resp.StatusCode
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

	request, err := http.NewRequest("POST", destination, body)
	request.Header = header
	response, err := httpClient.client.Do(request)
	httpResponse.Resp = response
	if err != nil {
		return &httpResponse, err
	}
	statusCode = response.StatusCode
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return &httpResponse, err
	}
	defer response.Body.Close()
	httpResponse.Result = gjson.Parse(string(content))
	return &httpResponse, nil
}
func RequestTrack(response HttpResponse) {
	elapsed := time.Since(response.StartTime)
	response.Elapsed = elapsed

	log.Println(fmt.Sprintf("%s STATUS CODE:%v COST:%s", response.Url, *response.StatusCode, elapsed))
}
func (httpClient *HttpClient) Get(destination string, header http.Header) (*HttpResponse, error) {
	if !httpClient.ava {
		httpClient.Init()
	}
	statusCode := -1
	httpResponse := HttpResponse{
		Method:     "GET",
		StatusCode: &statusCode,
		Url:        destination,
		StartTime:  time.Now(),
		Result:     gjson.Result{},
		Elapsed:    9999,
	}
	defer RequestTrack(httpResponse)
	var body io.Reader
	request, err := http.NewRequest("GET", destination, body)
	request.Header = header
	response, err := httpClient.client.Do(request)
	statusCode = response.StatusCode
	httpResponse.Resp = response
	if err != nil {
		return &httpResponse, err
	}
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return &httpResponse, err
	}
	defer response.Body.Close()
	httpResponse.Result = gjson.Parse(string(content))
	return &httpResponse, nil
}

func myCheckRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 1 {
		return errors.New("stop redirects")
	}
	return nil
}
