// Copyright 2019 James Cote All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package ctxclient_test

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/jfcote87/ctxclient"
)

func Example_func_nil() {
	ctx := context.Background()
	var clfunc ctxclient.Func

	req, _ := http.NewRequest("GET", "http://example.com", nil)

	res, err := clfunc.Do(ctx, req)
	switch ex := err.(type) {
	case *ctxclient.NotSuccess:
		log.Printf("server returned %s status: %s", ex.StatusMessage, ex.Body)
	case nil:
		log.Printf("successful server response %s", res.Status)
	default:
		log.Printf("transport error: %v", err)
	}
}

func Example_func() {
	var clfunc ctxclient.Func = func(ctx context.Context) (*http.Client, error) {
		k, _ := ctx.Value(UserKey).(string)
		if k == "" {

			return nil, errors.New("no user key provided in context")
			// or to use the default client instead:
			// return nil, ctxclient.ErrUseDefault
		}
		return &http.Client{Transport: &UserKeyTransport{UserKey: k}}, nil
	}
	ctx := context.WithValue(context.Background(), UserKey, "USER_GUID")
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	res, err := clfunc.Do(ctx, req)
	switch ex := err.(type) {
	case *ctxclient.NotSuccess:
		log.Printf("server returned %s status: %s", ex.StatusMessage, ex.Body)
	case nil:
		log.Printf("successful server response %s", res.Status)
	default:
		log.Printf("Transport error: %v", err)
	}
}

func Example_register() {
	// should be done during a package init()
	var clfunc ctxclient.Func = func(ctx context.Context) (*http.Client, error) {
		k, _ := ctx.Value(UserKey).(string)
		if k == "" {
			return nil, ctxclient.ErrUseDefault // use default instead
		}
		return &http.Client{Transport: &UserKeyTransport{UserKey: k}}, nil
	}
	ctxclient.RegisterFunc(clfunc)
}

var UserKey userKey

type userKey struct{}

type UserKeyTransport struct {
	UserKey string
}

func (t *UserKeyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	h := make(http.Header)
	for k, v := range r.Header {
		h[k] = v
	}
	newReq := *r
	h.Set("X-USERKEY", t.UserKey)
	newReq.Header = h
	return http.DefaultTransport.RoundTrip(&newReq)
}
