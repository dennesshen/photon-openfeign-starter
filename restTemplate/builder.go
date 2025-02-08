package restTemplate

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
)

type Executor struct {
	method  string
	ctx     context.Context
	Url     string
	Headers map[string][]string
	Body    *bytes.Buffer
}

func NewExecutor() *Executor {
	return &Executor{
		ctx:  context.Background(),
		Body: bytes.NewBuffer([]byte{}),
	}
}

func (r *Executor) SetContext(ctx context.Context) *Executor {
	r.ctx = ctx
	return r
}

func (r *Executor) SetMethod(method string) *Executor {
	r.method = strings.ToUpper(method)
	return r
}

func (r *Executor) GetMethod() string {
	return r.method
}

func (r *Executor) SetUrl(url string) *Executor {
	r.Url = url
	return r
}

func (r *Executor) AddHeader(key, value string) *Executor {
	if r.Headers == nil {
		r.Headers = make(map[string][]string)
	}
	r.Headers[key] = []string{value}
	return r
}

func (r *Executor) SetHeaders(headers map[string][]string) *Executor {
	r.Headers = headers
	return r
}

func (r *Executor) SetBody(body *bytes.Buffer) {
	r.Body = body
}

func (r *Executor) Execute() (resEntity RawResponse, err error) {
	var req *http.Request
	if req, err = http.NewRequestWithContext(r.ctx, r.method, r.Url, r.Body); err != nil {
		return
	}
	
	req.Header = r.Headers
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	
	var resBody []byte
	if resBody, err = io.ReadAll(resp.Body); err != nil {
		return
	}
	
	return RawResponse{
		Status:     resp.StatusCode,
		Headers:    resp.Header,
		RawResBody: resBody,
	}, nil
}
