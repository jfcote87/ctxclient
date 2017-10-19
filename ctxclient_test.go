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
	cl, err := f.Exec(context.Background())
	if err != nil {
		t.Errorf("nil Func.Exec expected nil err; got %v", err)
	}
	if cl != http.DefaultClient {
		t.Errorf("nil Func.Exec expected http.DefaultClient; go %#v", cl)
	}

	// check for err condition
	f = func(ctx context.Context) (*http.Client, error) {
		return http.DefaultClient, errors.New("Test Error")
	}
	cl, err = f.Exec(context.Background())
	if err != nil {
		t.Errorf("nil Func.Exec expected nil err; got %v", err)
	}
	if cl != http.DefaultClient {
		t.Errorf("nil Func.Exec expected http.DefaultClient; go %#v", cl)
	}

}
