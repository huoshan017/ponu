package http

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

func Get(url string) (data []byte, err error) {
	var resp *http.Response
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

func Post(url string, body io.Reader) (data []byte, err error) {
	var resp *http.Response
	resp, err = http.Post(url, "application/x-www-form-urlencoded", body)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

func PostForm(url_str string, url_values url.Values) (data []byte, err error) {
	var resp *http.Response
	resp, err = http.PostForm(url_str, url_values)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
}

type Request struct {
	req *http.Request
}

func (this *Request) Init(method, url string, body io.Reader) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	this.req = req
	return nil
}

func (this *Request) SetHeaderKV(key, value string) {
	if this.req != nil {
		this.req.Header.Set(key, value)
	}
}

func (this *Request) SetContentType(value string) {
	if this.req != nil {
		this.req.Header.Set("Content-Type", value)
	}
}

func (this *Request) SetCookie(value string) {
	if this.req != nil {
		this.req.Header.Set("Cookie", value)
	}
}

func (this *Request) SetDefaultContentType() {
	this.SetContentType("application/x-www-form-urlencoded")
}

func (this *Request) Do() (data []byte, err error) {
	client := &http.Client{}
	resp, err := client.Do(this.req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}
	data, err = ioutil.ReadAll(resp.Body)
	return
}
