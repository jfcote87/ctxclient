// Copyright 2019 James Cote All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ctxclient offers utilities for handling the
// selection and creation of http.Clients based on
// the context.  This borrows from ideas found in
// golang.org/x/oauth2.
//
// Usage example:
//
//   import (
//       "github.com/jfcote87/ctxclient"
//   )
//   ...
//   var clf *ctxclient.Func
//   req, _ := http.NewRequest("GET","http://example.com",nil)
//   res, err := clf.Do(req)
//   ...
//
package ctxclient // import "github.com/jfcote87/ctxclient"

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var defaultFuncs []Func

func defaultFunc(ctx context.Context) (*http.Client, error) {
	for _, f := range defaultFuncs {
		if cl, err := f(ctx); err != nil || cl != nil {
			return cl, err
		}
	}
	return http.DefaultClient, nil
}

// RegisterFunc adds f to the list of Funcs
// checked by the Default Func.  This should only be called
// during init as it is not thread safe.
func RegisterFunc(f Func) {
	if f != nil {
		// Place newly registered func at the top of list allowing
		// appengine default to always be last.
		defaultFuncs = append([]Func{f}, defaultFuncs...)
	}
}

// Func returns an http.Client pointer.
type Func func(ctx context.Context) (*http.Client, error)

// Client retrieves the default client.  If an error
// occurs, the error will be stored as an ErrorTransport
// in the client.  The error will be returned on all
// calls the client makes.
func Client(ctx context.Context) *http.Client {
	cl, err := defaultFunc(ctx)
	if err != nil {
		return &http.Client{
			Transport: &ErrorTransport{Err: err},
		}
	}
	return cl
}

// Client retrieves the Func's client.  If an error
// occurs, the error will be stored as an ErrorTransport
// in the client.  The error will be returned on all
// calls the client makes.
func (f Func) Client(ctx context.Context) *http.Client {
	if f == nil {
		return Client(ctx)
	}
	cl, err := f(ctx)
	if err != nil {
		return &http.Client{
			Transport: &ErrorTransport{Err: err},
		}
	}
	return cl
}

// Error checks the passed client for an ErrorTransport
// and returns the embedded error.
func Error(cl *http.Client) error {
	if t, ok := cl.Transport.(*ErrorTransport); ok {
		return t.Err
	}
	return nil
}

func do(ctx context.Context, cl *http.Client, req *http.Request) (*http.Response, error) {
	res, err := cl.Do(req.WithContext(ctx))
	// If we got an error, and the context has been canceled,
	// the context's error is probably more useful.
	if err != nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		default:
		}
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		buff, err := ioutil.ReadAll(res.Body)
		if err != nil {
			buff = []byte(fmt.Sprintf("%v", err))
		}
		res.Body.Close()
		return nil, &NotSuccess{
			StatusCode:    res.StatusCode,
			StatusMessage: res.Status,
			Header:        res.Header,
			Body:          buff,
		}
	}
	return res, err
}

// Do sends the request using the default client and checks for timeout/cancellation.
// Returns ResponseErr if response status is not 2xx. ctx must be non-nil
func Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	return do(ctx, Client(ctx), req)
}

// Do sends the request using the calculated client and checks for timeout/cancellation.
// Returns ResponseErr if response status is not 2xx. ctx must be non-nil
func (f Func) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if f == nil {
		return do(ctx, Client(ctx), req)
	}
	return do(ctx, f.Client(ctx), req)
}

// PostForm issues a POST request through the default http.Client
func PostForm(ctx context.Context, url string, payload url.Values) (*http.Response, error) {
	req, err := newPostFormRequest(url, payload)
	if err != nil {
		return nil, err
	}
	return do(ctx, Client(ctx), req)
}

// PostForm issues a POST request through the http.Client determined by f
func (f Func) PostForm(ctx context.Context, url string, payload url.Values) (*http.Response, error) {
	if f == nil {
		return PostForm(ctx, url, payload)
	}
	req, err := newPostFormRequest(url, payload)
	if err != nil {
		return nil, err
	}
	return do(ctx, f.Client(ctx), req)
}

func newPostFormRequest(url string, data url.Values) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

// NotSuccess contains body of a non 2xx http response
type NotSuccess struct {
	StatusCode    int
	StatusMessage string
	Body          []byte
	Header        http.Header
}

// Error fulfills error interface
func (re NotSuccess) Error() string {
	return fmt.Sprintf("response returned %d %s: %s", re.StatusCode, re.StatusMessage, string(re.Body))
}

// ErrorTransport returns the pass error on RoundTrip call.
// This RoundTripper should be used in cases where error
// handling can be postponed due to short response handling time.
type ErrorTransport struct{ Err error }

// RoundTrip always return the embedded err.  The error will be wrapped
// in an url.Error by http.Client
func (t *ErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return RequestError(req, t.Err)
}

// RequestError is a helper func to use in RoundTripper interfaces.
// Closes request body, checking for nils to you don't have to.
func RequestError(req *http.Request, err error) (*http.Response, error) {
	if req != nil && req.Body != nil {
		req.Body.Close()
	}
	return nil, err
}

// Transport returns the transport from the context's
// default client
func Transport(ctx context.Context) http.RoundTripper {
	cl, err := defaultFunc(ctx)
	if err != nil {
		return &ErrorTransport{Err: err}
	}
	if cl.Transport == nil {
		return http.DefaultTransport
	}
	return cl.Transport
}

// RoundTrip executes the clients RoundTripper
func RoundTrip(cl *http.Client, req *http.Request) (*http.Response, error) {
	if cl == nil || cl.Transport == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return cl.Transport.RoundTrip(req)
}
