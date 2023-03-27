package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	logbean "github.com/zxzixuanwang/gokit-logbean"
)

type returnCode struct {
	Ret int    `json:"ret,omitempty"`
	Msg string `json:"msg,omitempty"`
}

var (
	l           log.Logger
	username    *string
	password    *string
	logPosition *string

	urlS = []string{
		"https://suyunti55.com",
		"https://suyunti66.com",
		"https://suyunti77.com",
		"https://suyunti88.com",
		"https://suyunti99.com",
		"https://suyunti2.com",
	}
	loginUri *string
	checkUri *string
	logLevel *string
	urlInput *string
	protocol *string
)

func main() {
	username = flag.String("username", "username", "输入用户名")
	password = flag.String("password", "password", "输入密码")
	logPosition = flag.String("logposition", "./app.log", "日志位置")
	loginUri = flag.String("loginUri", "/auth/login", "登陆的路径资源")
	checkUri = flag.String("checkUri", "/user/checkin", "签到的路径资源")
	logLevel = flag.String("level", "info", "日志等级")
	urlInput = flag.String("urls", "", "输入访问的urls,','隔开")
	protocol = flag.String("protocol", "https", "http or https,默认: https")
	flag.Parse()

	if username == nil || password == nil {
		panic("密码或者用户名为空")
	}
	if strings.HasPrefix(*loginUri, "/") {
		*loginUri = "/" + *loginUri
	}
	if strings.HasPrefix(*checkUri, "/") {
		*checkUri = "/" + *checkUri
	}
	if urlInput != nil && *urlInput != "" {
		urlS = strings.Split(*urlInput, ",")
		if len(urlS) < 1 {
			panic("no url")
		}
		for i := 0; i < len(urlS); i++ {
			if !(strings.HasPrefix(urlS[i], "http") || strings.HasPrefix(urlS[i], "https")) {
				urlS[i] = fmt.Sprintf("%s://%s", *protocol, urlS[i])
			}
		}
	}

	l = logbean.GetLog(logbean.NewLogInfo(*logPosition, *logLevel))
	okFlag := true
	level.Debug(l).Log("username", *username, "password", *password)
	checkUrlOne := ""
	cookies := make(map[string]string)
	for _, urlOne := range urlS {

		loginUrl := urlOne + *loginUri
		form := url.Values{}
		form.Add("email", *username)
		form.Add("passwd", *password)
		bf := bytes.NewBufferString(form.Encode())
		level.Debug(l).Log("request", bf.String())

		resp, err := request(POST, loginUrl, nil, map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, bf)
		if err != nil {
			level.Error(l).Log("请求失败Do", err)
			continue
		}
		defer resp.Body.Close()

		answer := readIo(resp.Body)
		if answer != nil &&
			resp.StatusCode == http.StatusOK &&
			answer.Ret != 1 {

			level.Error(l).Log("out", answer.Msg)
			return
		}
		if resp.StatusCode == http.StatusOK {
			checkUrlOne = urlOne
			for _, cookie := range resp.Cookies() {
				level.Info(l).Log("cookie", *cookie)
				cookies[cookie.Name] = cookie.Value
			}
			okFlag = false
			break
		}
	}
	if okFlag {
		level.Error(l).Log("all url", "down")
		return
	}

	checkUrlOne = checkUrlOne + *checkUri
	resp, err := request(POST, checkUrlOne, cookies, map[string]string{"Content-Type": "application/json"}, nil)
	if err != nil {
		if err != nil {
			level.Error(l).Log("check requet error", err)
			return
		}
	}

	defer resp.Body.Close()
	readIo(resp.Body)
}

type HttpMethod string

const (
	POST HttpMethod = "POST"
	GET  HttpMethod = "GET"
	PUT  HttpMethod = "PUT"
)

func request(method HttpMethod, url string,
	cookies map[string]string,
	header map[string]string,
	body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(string(method), url, body)
	if err != nil {
		level.Error(l).Log("请求失败new request", err)
		return nil, err
	}
	if len(cookies) > 0 {
		for k, v := range cookies {
			cookie := http.Cookie{Name: k, Value: v}
			req.AddCookie(&cookie)
		}
	}
	if len(header) > 0 {
		for k, v := range header {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{}
	client.Timeout = time.Minute
	return client.Do(req)
}

func readIo(r io.Reader) *returnCode {
	tempBody, err := io.ReadAll(r)
	if err != nil {
		level.Error(l).Log("read resp err", err)
	}
	answer := new(returnCode)
	err = json.Unmarshal(tempBody, answer)
	if err != nil {
		level.Error(l).Log("unmarshal json err", err)
		level.Debug(l).Log("out", string(tempBody))
	} else {
		level.Info(l).Log("out", string(answer.Msg), "ret code", answer.Ret)
	}
	return answer
}
