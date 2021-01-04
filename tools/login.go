package tools

import (
	"bilibili-recording-go/infos"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/asmcos/requests"
	"github.com/kataras/golog"
	"github.com/tidwall/gjson"
)


var (
	appKey = "bca7e84c2d947ac6"
	salt = "60698ba2f68e01ce44738920a0ffe768"
)

func getKey() (string, *rsa.PublicKey, error) {
	url := "https://passport.bilibili.com/api/oauth2/getKey"
	data := requests.Params{
		"appkey": appKey,
		"sign": md5V(fmt.Sprintf("appkey=%s%s", appKey, salt)),
	}
	for {
		resp, _ := requests.Post(url, data)
		code := gjson.Get(resp.Text(), "code").Int()
		if code == 0 {
			keyHash := gjson.Get(resp.Text(), "data").Get("hash").String()
			pubKey, err := getPublicKeyFromString(gjson.Get(resp.Text(), "data").Get("key").String())
			if err!= nil {
				golog.Error(err)
				time.Sleep(1 * time.Second)
				continue
			}
			return keyHash, pubKey, nil
		}
		time.Sleep(1 * time.Second)
	}
}

func getPublicKeyFromString (pub string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pub))
	if block == nil {
		return nil, errors.New("public key error!")
	}
	pubKeyInterface ,err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pubKeyInterface.(*rsa.PublicKey), nil
}

// encrypt and base64 encode
func encrypt(plainText string, publicKey *rsa.PublicKey) (string) {
	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(plainText))
	if err != nil {
		golog.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(cipherText)
}

// LoginByPassword login bilibili
func LoginByPassword(username string, password string) (map[string]string, error) {
	keyHash, pubKey, _ := getKey()
	urls := "https://passport.bilibili.com/api/v3/oauth2/login"
	u := url.QueryEscape(username)
	p := encrypt(fmt.Sprintf("%s%s",keyHash, password), pubKey)
	// p := "i9Rr3vnd5MaAjPFnGQQBpwSttiDscfobc27kT7a7GzqI3f7f4rtA8m6fPP2E49DbvJhXV21jLhKf7g5PVSH930Ld3JXAAqiHaD/JKaEIXXmIAg2cHuluQOj2b+icQcNCNyaza0GCz3ccRr0N95bRCMU91qudSnLqWLsRU1YV+6M="
	ts := time.Now().Unix()
	// ts := 1609312260
	paras := requests.Params{
		"access_key": "",
		"actionKey": "appkey",
		"appkey": appKey,
		"build": "6040500",
		"captcha": "",
		"challenge": "",
		"channel": "bili",
		"cookies": "",
		"device": "phone",
		"mobi_app": "android",
		"password": p,
		"permission": "ALL",
		"platform": "android",
		"seccode": "",
		"subid": "1",
		"ts": fmt.Sprintf("%d", ts),
		"username": u,
		"validate": "",
		"sign": md5V(fmt.Sprintf("access_key=&actionKey=appkey&appkey=%s&build=6040500&captcha=&challenge=&channel=bili&cookies=&device=phone&mobi_app=android&password=%s&permission=ALL&platform=android&seccode=&subid=1&ts=%d&username=%s&validate=%s", appKey, url.QueryEscape(p), ts, u, salt)),
	}

	header := requests.Header{
		"Content-type": "application/x-www-form-urlencoded",
	}
	req := requests.Requests()
	resp, _ := req.Post(urls, paras, header)

	cookies := make(map[string]string)

	if gjson.Get(resp.Text(), "code").Int() == 0 && gjson.Get(resp.Text(), "data").Get("status").Int() == 0 {
		data := gjson.Get(resp.Text(), "data").Get("cookie_info").Get("cookies")
		data.ForEach(func (key, value gjson.Result) bool {
			cookies[value.Get("name").String()] = value.Get("value").String()
			return true
		})
		return cookies, nil
	}
	infs := infos.New()
	for k, v := range cookies {
		infs.BiliInfo.Cookies = append(infs.BiliInfo.Cookies, &http.Cookie{Name:k, Value:v})
	}

	return cookies, errors.New("Login Error")
}