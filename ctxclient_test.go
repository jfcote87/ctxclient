package ctxclient_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/jfcote87/ctxclient"
	"golang.org/x/net/context"
)

func TestFunc(t *testing.T) {
	var f ctxclient.Func
	cl, err := f.Get(context.Background())
	if err != nil {
		t.Errorf("nil Func.Exec expected nil err; got %v", err)
	}
	if cl != http.DefaultClient {
		t.Errorf("nil Func.Exec expected http.DefaultClient; go %#v", cl)
	}

	var testErr = errors.New("TestError")
	// check for err condition
	f = func(ctx context.Context) (*http.Client, error) {
		return http.DefaultClient, testErr
	}
	cl, err = f.Get(context.Background())
	if err != testErr {
		t.Errorf("error Func.Exec expected testErr; got %v", err)
	}
	if cl != http.DefaultClient {
		t.Errorf("nil Func.Exec expected http.DefaultClient; go %#v", cl)
	}

}
