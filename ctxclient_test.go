// Copyright 2017 James Cote All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ctxclient_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/jfcote87/ctxclient"
)

// TestNullFunc ensures that a null Func will return defaults
func TestNullFunc(t *testing.T) {
	var f ctxclient.Func
	ctx := context.Background()
	cl := f.Client(ctx)
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
	var testErr error

	var testCl = &http.Client{}
	// check for err condition
	var f ctxclient.Func = func(ctx context.Context) (*http.Client, error) {
		return testCl, testErr
	}
	cl := f.Client(ctx)
	if cl != testCl {
		t.Errorf("error Func.Get expected testCl; go %#v", cl)
	}

	testErr = errors.New("TestError")
	testCl = nil
	if cl = f.Client(ctx); ctxclient.Error(cl) != testErr {
		t.Errorf("error Func.Client expected testErr Transport; got %#v", ctxclient.Error(cl))
	}

	// check that the error transport returns testErr wrapped in url.Error
	_, err := cl.Get("http://test.com")
	switch e := err.(type) {
	case *url.Error:
		if e.Err != testErr {
			t.Errorf("error Func.Client expected to return testErr on Get call; got %#v", e.Err)
		}
	default:
		t.Errorf("error Func.Client expected to return *url.Error on Get call; got %#v", err)
	}
}

func TestRegister(t *testing.T) {
	tempClient := &http.Client{
		Transport: &ctxclient.ErrorTransport{Err: errors.New("Test Error Transport")},
	}
	var tempErr error = errors.New("CTX Error")

	ctxclient.RegisterFunc(func(ctx context.Context) (*http.Client, error) {
		k, _ := ctx.Value("ctxkey").(string)
		if k == "A" {
			return tempClient, nil
		}
		return nil, nil
	})
	ctxclient.RegisterFunc(func(ctx context.Context) (*http.Client, error) {
		k, _ := ctx.Value("ctxkey").(string)
		if k == "B" {
			return nil, tempErr
		}
		return nil, nil
	})
	ctx := context.Background()
	var clFunc ctxclient.Func
	cl := clFunc.Client(ctx)
	if cl != http.DefaultClient {
		t.Fatalf("expected http.DefaultClient")
	}

	ctx = context.WithValue(ctx, "ctxkey", "B")
	cl = clFunc.Client(ctx)

	if err := ctxclient.Error(cl); err == nil {
		t.Fatalf("expected error")
	} else if err != tempErr {
		t.Fatalf("expected tempErr; got %#v", err)
	}

	ctx = context.WithValue(ctx, "ctxkey", "A")
	cl = clFunc.Client(ctx)
	if err := ctxclient.Error(cl); err == nil {
		t.Fatalf("expected tempClient; got %v", err)
	} else if cl != tempClient {
		t.Fatalf("expected tempClient; got %#v", cl)
	}

	ctx = context.WithValue(ctx, "ctxkey", "B")
	err := ctxclient.Error(clFunc.Client(ctx))
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != tempErr {
		t.Fatalf("expected tempErr; got %#v", err)
	}
	ctx = context.WithValue(ctx, "ctxkey", "A")
	if cl = clFunc.Client(ctx); cl != tempClient {
		t.Fatalf("expected tempClient; got %#v", cl)
	}
	tempClient2 := &http.Client{}
	clFunc = func(ctx context.Context) (*http.Client, error) {
		return tempClient2, nil
	}
	if cl = clFunc.Client(ctx); cl != tempClient2 {
		t.Fatalf("expected tempClient2; got %#v", cl)
	}
	if err = ctxclient.Error(cl); err != nil {
		t.Fatalf("expected nil error on client; got %v", err)
	}

}

type testCloser struct {
	*bytes.Reader
	IsClosed bool
}

func (t *testCloser) Close() error {
	t.IsClosed = true
	return nil
}

func TestRequestError(t *testing.T) {
	tRdr := &testCloser{}
	req, _ := http.NewRequest("POST", "http://www.example.com", tRdr)
	ctxclient.RequestError(req, errors.New("Error"))
	if !tRdr.IsClosed {
		t.Error("expected reader to be closed by RequestError")
	}
}
