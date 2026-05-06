package client

import (
	"net/url"

	"github.com/laixhe/gonet/network/header"
)

type Request struct {
	c           *Client
	err         error
	Method      string
	URL         *url.URL
	QueryParams url.Values
	Result      any
}

func (r *Request) SetQueryParam(param, value string) *Request {
	r.QueryParams.Set(param, value)
	return r
}

func (r *Request) SetQueryParams(params map[string]string) *Request {
	for k, v := range params {
		r.SetQueryParam(k, v)
	}
	return r
}

func (r *Request) Text() (string, error) {
	if r.err != nil {
		return "", r.err
	}
	r.URL.RawQuery = r.QueryParams.Encode()
	req, err := HttpRequest(r.Method, r.URL.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set(header.Accept, "*/*")
	req.Header.Set(header.Accept, "*/*")
	return "", nil
}
