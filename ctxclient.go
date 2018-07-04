// Copyright 2017 James Cote and Liberty Fund, Inc.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctxclient

import (
	"net/http"

	"golang.org/x/net/context"
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
		defaultFuncs = append([]Func{f}, defaultFuncs...)
	}
}

// Func returns an http.Client pointer.
type Func func(ctx context.Context) (*http.Client, error)

// Get retrieves the default client for the passed context
func Get(ctx context.Context) (*http.Client, error) {
	return defaultFunc(ctx)
}

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

// Get safely executes the by executing DefaultFunc if nil
func (f Func) Get(ctx context.Context) (*http.Client, error) {
	if f == nil {
		return defaultFunc(ctx)
	}
	return f(ctx)
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

// ErrorTransport returns the specified error on RoundTrip.
// This RoundTripper should be used in cases where
// error handling can be postponed to response handling time.
type ErrorTransport struct{ Err error }

// RoundTrip always return the embedded err.  The error will be wrapped
// in an url.Error by http.Client
func (t *ErrorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, t.Err
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

// Transport returns the transport of the client
func (f Func) Transport(ctx context.Context) http.RoundTripper {
	if f == nil {
		return Transport(ctx)
	}
	cl, err := f(ctx)
	if err != nil {
		return &ErrorTransport{Err: err}
	}
	return cl.Transport
}
