// Copyright 2017 James Cote and Liberty Fund, Inc.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ctxclient offers utilities for handling the
// selection and creation of http.Clients based on
// the context.  This borrows from ideas found in
// golang.org/x/oauth2.
//
// The obvious usage exists app engine environment using
// the urlfetch package.  By allowing client decision to wait
// until the actual Do() call, boiler plate selection code
// can be removed.
package ctxclient

import (
	"net/http"

	"golang.org/x/net/context"
)

var defaultFuncs []Func

func defaultFunc(ctx context.Context) (*http.Client, error) {
	for _, f := range defaultFuncs {
		cl, err := f(ctx)
		if err != nil {
			return nil, err
		}
		if cl != nil {
			return cl, nil
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

// DefaultFunc provides the default system client.  In app engine environments
// this will be overwritten by the urlfetch.Client(ctx) function.
func DefaultFunc(ctx context.Context) (*http.Client, error) {
	return defaultFunc(ctx)
}

// Func returns an http.Client pointer.
type Func func(ctx context.Context) (*http.Client, error)

// Exec safely executes the by executing DefaultFunc if nil
func (r Func) Exec(ctx context.Context) (*http.Client, error) {
	if r == nil {
		return defaultFunc(ctx)
	}
	return r(ctx)
}

// Client handles selection errors by wrapping them in an
// ErrorTransport so that errors are not found until the
// actual Client.Do(...)
func (r Func) Client(ctx context.Context) *http.Client {
	if r == nil {
		r = DefaultFunc
	}
	cl, err := r(ctx)
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
// This RoundTripper should be used in rare error cases where
// error handling can be postponed to response handling time.
type ErrorTransport struct{ Err error }

// RoundTrip always return the embedded err.  The error will be wrapped
// in an url.Error by http.Client
func (t *ErrorTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, t.Err
}
