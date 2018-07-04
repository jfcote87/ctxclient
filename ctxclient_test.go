// Copyright 2017 James Cote and Liberty Fund, Inc.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctxclient_test

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/jfcote87/ctxclient"
	"golang.org/x/net/context"
)

// TestNullFunc ensures that a null Func will return defaults
func TestNullFunc(t *testing.T) {
	var f ctxclient.Func
	ctx := context.Background()
	cl, err := f.Get(ctx)
	if err != nil {
		t.Errorf("nil Func.Get expected nil err; got %v", err)
	}
	if cl != http.DefaultClient {
		t.Errorf("nil Func.Get expected http.DefaultClient; got %#v", cl)
	}
	cl = f.Client(ctx)
	if cl != http.DefaultClient {
		t.Errorf("nil Func.Client expected http.DefaultClient; got %#v", cl)
	}
}

func TestFuncError(t *testing.T) {
	ctx := context.Background()
	var testErr = errors.New("TestError")

	var testCl = &http.Client{}
	// check for err condition
	var f ctxclient.Func = func(ctx context.Context) (*http.Client, error) {
		return testCl, testErr
	}
	cl, err := f.Get(context.Background())
	if err != testErr {
		t.Errorf("error Func.Get expected testErr; got %v", err)
	}
	if cl != testCl {
		t.Errorf("error Func.Get expected testCl; go %#v", cl)
	}

	if cl = f.Client(ctx); ctxclient.Error(cl) != testErr {
		t.Errorf("error Func.Client expected testErr Transport; got %#v", ctxclient.Error(cl))
	}

	// check that the error transport returns testErr wrapped in url.Error
	_, err = cl.Get("http://test.com")
	switch e := err.(type) {
	case *url.Error:
		if e.Err != testErr {
			t.Errorf("error Func.Client expected to return testErr on Get call; got %#v", e.Err)
		}
	default:
		t.Errorf("error Func.Client expected to return *url.Error on Get call; got %#v", err)
	}
}
