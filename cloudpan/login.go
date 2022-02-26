// Copyright (c) 2020 tickstep.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// WEB网页端API
package cloudpan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/phpc0de/ctapi/cloudpan/apierror"
	"github.com/phpc0de/ctapi/cloudpan/apiutil"
	"github.com/phpc0de/ctlibgo/crypto"
	"github.com/phpc0de/ctlibgo/logger"
	"github.com/phpc0de/ctlibgo/requester"
	"image/png"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// CaptchaName 验证码文件名称
	captchaName = "captcha.png"
)

type (
	loginParams struct {
		CaptchaToken string
		Lt string
		ReturnUrl string
		ParamId string
	}

	loginResult struct {
		Result int `json:"result"`
		Msg string `json:"msg"`
		ToUrl string `json:"toUrl"`
	}

	WebLoginToken struct {
		CookieLoginUser string `json:"cookieLoginUser"`
	}
)

var (
	latestLoginParams loginParams
	client            = requester.NewHTTPClient()
)

func Login(username, password string) (webToken *WebLoginToken, error *apierror.ApiError) {
	client.ResetCookiejar()
	params, err := getLoginParams()
	if err != nil {
		logger.Verboseln("get login params error")
		return nil, err
	}

	err = checkNeedCaptchaCodeOrNot(username, latestLoginParams.Lt)
	if err != nil {
		return nil, err
	}

	// save latest params
	latestLoginParams = params
	token, err := LoginWithCaptcha(username, password, "")
	if err != nil {
		return nil, err
	}
	return token, nil
}

func LoginWithCaptcha(username, password, captchaCode string) (webToken *WebLoginToken, error *apierror.ApiError) {
	//client.ResetCookiejar()
	//latestLoginParams, _ = getLoginParams()

	webToken = &WebLoginToken{}
	if latestLoginParams.CaptchaToken == "" {
		latestLoginParams, _ = getLoginParams()
	}

	r, err := doLoginAct(username, password, captchaCode, latestLoginParams.CaptchaToken,
		latestLoginParams.ReturnUrl, latestLoginParams.ParamId, latestLoginParams.Lt)
	if err != nil || r.Msg != "登录成功" {
		logger.Verboseln("login failed ", err)
		return webToken, apierror.NewFailedApiError(err.Error())
	}
	// request toUrl to get COOKIE_LOGIN_USER cookie
	header := map[string]string {
		"lt":           latestLoginParams.Lt,
		"Content-Type": "application/x-www-form-urlencoded",
		"Referer":      "https://open.e.189.cn/",
	}
	client.Fetch("GET", r.ToUrl, nil, header)

	cloudpanUrl := &url.URL{
		Scheme: "http",
		Host:   "cloud.189.cn",
		Path: "/",
	}
	cks := client.Jar.Cookies(cloudpanUrl)
	for _, cookie := range cks {
		if cookie.Name == "COOKIE_LOGIN_USER" {
			webToken.CookieLoginUser = cookie.Value
			break
		}
	}

	return
}

func GetCaptchaImage() (savePath string, error *apierror.ApiError) {
	if latestLoginParams.CaptchaToken == "" {
		latestLoginParams, _ = getLoginParams()
	}

	removeCaptchaPath()
	picUrl := AUTH_URL + "/picCaptcha.do?token=" + latestLoginParams.CaptchaToken
	// save img to file
	return saveCaptchaImg(picUrl)
}

func getLoginParams() (params loginParams, error *apierror.ApiError) {
	header := map[string]string {
		"Content-Type": "application/x-www-form-urlencoded",
	}
	data, err := client.Fetch("GET", WEB_URL+ "/udb/udb_login.jsp?pageId=1&redirectURL=/main.action",
		nil, header)
	if err != nil {
		logger.Verboseln("login redirectURL occurs error: ", err.Error())
		return params, apierror.NewApiErrorWithError(err)
	}
	content := string(data)

	re, _ := regexp.Compile("captchaToken' value='(.+?)'")
	params.CaptchaToken = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("lt = \"(.+?)\"")
	params.Lt = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("returnUrl = '(.+?)'")
	params.ReturnUrl = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("paramId = \"(.+?)\"")
	params.ParamId = re.FindStringSubmatch(content)[1]
	return
}

func checkNeedCaptchaCodeOrNot(username, lt string) (error *apierror.ApiError) {
	url := AUTH_URL + "/needcaptcha.do"
	rsa, err := crypto.RsaEncrypt([]byte(apiutil.RsaPublicKey), []byte(username))
	if err != nil {
		return apierror.NewApiErrorWithError(err)
	}
	postData := map[string]string {
		"accountType": "01",
		"userName": "{RSA}" + apiutil.B64toHex(string(crypto.Base64Encode(rsa))),
		"appKey": "cloud",
	}
	header := map[string]string {
		"lt": lt,
		"Content-Type": "application/x-www-form-urlencoded",
		"Referer": "https://open.e.189.cn/",
	}
	body, err := client.Fetch("POST", url, postData, header)
	if err != nil {
		logger.Verboseln("get captcha code error: ", err.Error())
		return apierror.NewApiErrorWithError(err)
	}
	text := string(body)
	if text != "0" {
		// need captcha
		return apierror.NewApiError(apierror.ApiCodeNeedCaptchaCode, "需要验证码")
	}
	return
}

func saveCaptchaImg(imgURL string) (savePath string, error *apierror.ApiError) {
	logger.Verboseln("try to download captcha image: ", imgURL)
	imgContents, err := client.Fetch("GET", imgURL, nil, nil)
	if err != nil {
		return "", apierror.NewApiErrorWithError(fmt.Errorf("获取验证码失败, 错误: %s", err))
	}

	_, err = png.Decode(bytes.NewReader(imgContents))
	if err != nil {
		return "", apierror.NewApiErrorWithError(fmt.Errorf("验证码解析错误: %s", err))
	}

	savePath = captchaPath()
	return savePath, apierror.NewApiErrorWithError(ioutil.WriteFile(savePath, imgContents, 0777))
}

func captchaPath() string {
	return filepath.Join(os.TempDir(), captchaName)
}

func removeCaptchaPath() error {
	return os.Remove(captchaPath())
}

func doLoginAct(username, password, validateCode, captchaToken, returnUrl, paramId, lt string) (result *loginResult, error *apierror.ApiError) {
	url := AUTH_URL + "/loginSubmit.do"
	rsaUserName, _ := crypto.RsaEncrypt([]byte(apiutil.RsaPublicKey), []byte(username))
	rsaPassword, _ := crypto.RsaEncrypt([]byte(apiutil.RsaPublicKey), []byte(password))
	data := map[string]string {
		"appKey": "cloud",
		"accountType": "01",
		"userName": "{RSA}" + apiutil.B64toHex(string(crypto.Base64Encode(rsaUserName))),
		"password": "{RSA}" + apiutil.B64toHex(string(crypto.Base64Encode(rsaPassword))),
		"validateCode": validateCode,
		"captchaToken": captchaToken,
		"returnUrl": returnUrl,
		"mailSuffix": "@189.cn",
		"paramId": paramId,
	}
	header := map[string]string {
		"lt": lt,
		"Content-Type": "application/x-www-form-urlencoded",
		"Referer": "https://open.e.189.cn/",
	}

	body, err := client.Fetch("POST", url, data, header)
	if err != nil {
		logger.Verboseln("login with captch error ", err)
		return nil, apierror.NewFailedApiError(err.Error())
	}

	r := &loginResult{}
	if err := json.Unmarshal(body, r); err != nil {
		logger.Verboseln("parse login resutl json error ", err)
		return nil, apierror.NewFailedApiError(err.Error())
	}
	return r, nil
}

func buildCookie(cookieMap map[string]string) []*http.Cookie {
	if cookieMap == nil {
		return nil
	}

	c := make([]*http.Cookie, 0, 0)
	for k,v := range cookieMap {
		c = append(c,
			&http.Cookie{
				Name: k,
				Value: v,
				Path: "/",
			})
	}
	return c
}

func RefreshCookieToken(sessionKey string) string {
	client := requester.NewHTTPClient()

	header := map[string]string {
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8,ja;q=0.7",
	}

	//cloudpanUrl := &url.URL{
	//	Scheme: "http",
	//	Host:   "cloud.189.cn",
	//	Path: "/",
	//}

	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/ssoLogin.action?sessionKey=%s&redirectUrl=main.action%%23recycle",
		WEB_URL, sessionKey)
	logger.Verboseln("do request url: " + fullUrl.String())
	resp, err := client.Req("GET", fullUrl.String(), nil, header)
	if err != nil {
		logger.Verboseln("refresh web token cookie error ", err)
		return ""
	}
	cks := resp.Request.Cookies()
	for _, cookie := range cks {
		if cookie.Name == "COOKIE_LOGIN_USER" {
			return cookie.Value
		}
	}
	return ""
}