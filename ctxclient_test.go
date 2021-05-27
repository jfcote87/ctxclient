// Copyright 2019 James Cote All rights reserved.
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
	"github.com/jfcote87/testutils"
)

// TestNullFunc ensures that a null Func will return defaults
func TestNullFunc(t *testing.T) {
	var f ctxclient.Func
	ctx := context.Background()
	cl := f.Client(ctx)
	if cl != http.DefaultClient {
		t.Errorf("nil Func.Get expected http.DefaultClient; got %#v", cl)
	}
	f = func(c context.Context) (*http.Client, error) {
		return nil, nil
	}
	cl = f.Client(ctx)
	err := ctxclient.Error(cl)
	if err == nil || err.Error() != "nil client" {
		t.Errorf("Func.Client nil client returned expected nil client err; got %v", err)
	}
}

func Test_do(t *testing.T) {
	testTransport := &testutils.Transport{}
	clx := &http.Client{Transport: testTransport}
	var f ctxclient.Func = func(ctx context.Context) (*http.Client, error) {
		return clx, nil
	}
	hdr := make(http.Header)
	hdr.Set("X-Test1", "Test Value")
	testTransport.Add(
		&testutils.RequestTester{ // test 0
			Response: testutils.MakeResponse(200, nil, nil),
		},
		&testutils.RequestTester{ // test 0
			Response: testutils.MakeResponse(400, []byte("Bad Request"), hdr),
		},
	)
	ctx := context.Background()
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	res, err := f.Do(ctx, r)
	if err != nil {
		t.Errorf("wanted success; got %#v", err)
	}
	if res != nil && res.Body != nil {
		res.Body.Close()
	}
	r, _ = http.NewRequest("GET", "http://example.com", nil)
	res, err = f.Do(ctx, r)
	nsErr, ok := err.(*ctxclient.NotSuccess)
	if !ok || nsErr.StatusCode != 400 || string(nsErr.Body) != "Bad Request" || nsErr.Header.Get("X-Test1") != "Test Value" {
		t.Errorf("wanted *ctxclient.NotSuccess with StatusCode: 400, Body: \"BadRequest\", X-Test1 header == \"Test Value\"; got %#v", err)
	}
	if res != nil && res.Body != nil {
		res.Body.Close()
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

type testCtxKey struct{}

func TestRegister(t *testing.T) {
	tempClient := &http.Client{
		Transport: &ctxclient.ErrorTransport{Err: errors.New("Test Error Transport")},
	}
	var tempErr error = errors.New("CTX Error")

	var ctxkey testCtxKey
	ctxclient.RegisterFunc(func(ctx context.Context) (*http.Client, error) {
		k, _ := ctx.Value(ctxkey).(string)
		if k == "A" {
			return tempClient, nil
		}
		return nil, ctxclient.ErrUseDefault
	})
	ctxclient.RegisterFunc(func(ctx context.Context) (*http.Client, error) {
		k, _ := ctx.Value(ctxkey).(string)
		if k == "B" {
			return nil, tempErr
		}
		return nil, ctxclient.ErrUseDefault
	})
	ctx := context.Background()
	var clFunc ctxclient.Func
	cl := clFunc.Client(ctx)
	if cl != http.DefaultClient {
		t.Fatal("expected http.DefaultClient")
	}

	ctx = context.WithValue(ctx, ctxkey, "B")
	cl = clFunc.Client(ctx)

	if err := ctxclient.Error(cl); err == nil {
		t.Fatalf("expected error")
	} else if err != tempErr {
		t.Fatalf("expected tempErr; got %#v", err)
	}

	ctx = context.WithValue(ctx, ctxkey, "A")
	cl = clFunc.Client(ctx)
	if err := ctxclient.Error(cl); err == nil {
		t.Fatalf("expected tempClient; got %v", err)
	} else if cl != tempClient {
		t.Fatalf("expected tempClient; got %#v", cl)
	}

	ctx = context.WithValue(ctx, ctxkey, "B")
	err := ctxclient.Error(clFunc.Client(ctx))
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != tempErr {
		t.Fatalf("expected tempErr; got %#v", err)
	}
	ctx = context.WithValue(ctx, ctxkey, "A")
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
