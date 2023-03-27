package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/zxzixuanwang/suyun-check/logbean"
)

type returnCode struct {
	Ret int    `json:"ret,omitempty"`
	Msg string `json:"msg,omitempty"`
}

var (
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
)

func main() {
	username = flag.String("username", "username", "输入用户名")
	password = flag.String("password", "password", "输入密码")
	logPosition = flag.String("logposition", "./app.log", "日志位置")
	loginUri = flag.String("loginUri", "/auth/login", "登陆的路径资源")
	checkUri = flag.String("checkUri", "/user/checkin", "签到的路径资源")
	logLevel = flag.String("level", "info", "日志等级")
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

	l := logbean.GetLog(logbean.NewLogInfo(*logPosition, *logLevel))
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
		req, err := http.NewRequest("POST", loginUrl, bf)
		if err != nil {
			level.Error(l).Log("请求失败new request", err)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		client.Timeout = time.Minute
		resp, err := client.Do(req)
		if err != nil {

			level.Error(l).Log("请求失败Do", err)
			continue
		}
		defer resp.Body.Close()
		answer := readIo(resp.Body, l)
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
		level.Error(l).Log("all url ", "down")
		return
	}
	checkUrlOne = checkUrlOne + *checkUri

	req, err := http.NewRequest("POST", checkUrlOne, nil)
	if err != nil {
		level.Error(l).Log("login requet error", err)

		return
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range cookies {
		cookie := http.Cookie{Name: k, Value: v}
		req.AddCookie(&cookie)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		level.Error(l).Log("check requet error", err)

		return
	}
	defer resp.Body.Close()
	readIo(resp.Body, l)
}

func readIo(r io.Reader, l log.Logger) *returnCode {
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
