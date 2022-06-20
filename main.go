package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

var clockData = map[string]string{
	"ClientDetails.QuickClockInPassword": "True",
	"Subdomain":                          "tiersec",
	"ClientDetails.QuickClockIn":         "True",
	"UserName":                           "Elena.Kozlova@septier.com",
	"Password":                           "qwaszx1+",
	"command":                            "",
}

type DataForReq struct {
	client    *http.Client
	url       *url.URL
	serverUrl string
	cookies   []*http.Cookie
}

func NewDataForReq() *DataForReq {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}

	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		Jar:       jar,
	}
	return &DataForReq{
		client:    client,
		url:       &url.URL{Scheme: "https", Host: "apps.timeclockwizard.com"},
		serverUrl: "https://apps.timeclockwizard.com",
		cookies:   nil,
	}
}

func main() {
	data := NewDataForReq()
	if err := data.getCookies(); err != nil {
		log.Fatal("get cookies: %w", err)
	}

	if err := data.clockIn(); err != nil {
		log.Fatal("clock in: %w", err)
	}

	time.Sleep(3 * time.Hour)

	if err := data.logIn(); err != nil {
		log.Fatal("log in: %w", err)
	}
	if err := data.breakIn(); err != nil {
		log.Fatal("break in: %w", err)
	}

	time.Sleep(30 * time.Minute)

	if err := data.logIn(); err != nil {
		log.Fatal("log in: %w", err)
	}
	if err := data.breakOut(); err != nil {
		log.Fatal("break out: %w", err)
	}

	time.Sleep(5 * time.Hour)

	if err := data.logIn(); err != nil {
		log.Fatal("log in: %w", err)
	}
	if err := data.clockOut(); err != nil {
		log.Fatal("clock out: %w", err)
	}
}

func (s *DataForReq) getCookies() error {
	req, err := http.NewRequest("GET", s.serverUrl+"/Login?Subdomain=tiersec", nil)
	if err != nil {
		return err
	}
	res, err := s.client.Do(req)
	if err != nil {
		return err
	}

	s.cookies = s.client.Jar.Cookies(s.url)
	if err := res.Body.Close(); err != nil {
		return err
	}
	return nil
}

func (s *DataForReq) logIn() error {
	clockData["command"] = "LogIn"
	return s.doClockRequest(clockData, "/Login")
}

func (s *DataForReq) clockIn() error {
	clockData["command"] = "ClockIn"
	return s.doClockRequest(clockData, "/Login")
}

func (s *DataForReq) clockOut() error {
	clockData["command"] = "ClockOut"
	return s.doClockRequest(clockData, "/Login")
}

func (s *DataForReq) breakIn() error {
	return s.doBreakRequest("isBreakIN=true&employeeID=0", "/ClockIN/BreakInOut")
}

func (s *DataForReq) breakOut() error {
	return s.doBreakRequest("isBreakIN=false&employeeID=0", "/ClockIN/BreakInOut")
}

func (s *DataForReq) doClockRequest(data map[string]string, uri string) error {
	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)

	for k, v := range data {
		if err := w.WriteField(k, v); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", s.serverUrl+uri, b)
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", "ASP.NET_SessionId="+s.cookies[0].Value+"; Subdomain=tiersec; _culture=en-US; __RequestVerificationToken="+s.cookies[3].Value)
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}

	fmt.Println("response code: ", res.StatusCode)

	if err := res.Body.Close(); err != nil {
		return err
	}
	return nil
}

func (s *DataForReq) doBreakRequest(data string, uri string) error {
	req, err := http.NewRequest("POST", s.serverUrl+uri, strings.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", "ASP.NET_SessionId="+s.cookies[0].Value+"; Subdomain=tiersec; _culture=en-US; __RequestVerificationToken="+s.cookies[3].Value)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}

	fmt.Println("response code: ", res.StatusCode)

	if err := res.Body.Close(); err != nil {
		return err
	}
	return nil
}
