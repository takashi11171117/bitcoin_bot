package main

import (
	"context"
	"errors"
	"io/ioutil"

	"google.golang.org/appengine/urlfetch"
)

const (
	GET_METHOD  = "GET"
	POST_METHOD = "POST"
	END_POINT   = "https://api.bitflyer.jp"
)

type Lightning struct {
	AccessKey    string
	AccessSecret string
}

func (l *Lightning) RequestPublic(method string, path string, params map[string]string, ctx context.Context) (err error, s string) {
	if method != GET_METHOD {
		return errors.New("Request Error"), ""
	}

	client := urlfetch.Client(ctx)

	query := "?"
	for key, param := range params {
		query += key + "=" + param
	}

	resp, err := client.Get(END_POINT + path + query)
	if err != nil {
		return err, ""
	}

	defer resp.Body.Close()
	byteArray, _ := ioutil.ReadAll(resp.Body)

	return nil, string(byteArray)
}
