# ctxclient package

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/jfcote87/ctxclient)
[![BSD 3-Clause License](https://img.shields.io/badge/license-bsd%203--clause-blue.svg)](https://github.com/jfcote87/ctxclient/blob/master/LICENSE)

Package ctxclient offers utilities for handling the
selection and creation of http.Clients based on
the context.  This borrows from ideas found in
golang.org/x/oauth2.

I created this package to simplify client selection
when implementing client api packages for vendor's
web services.  The obvious usage exists in the app
engine environment using the urlfetch package.  By
allowing client decision to wait until the actual Do()
call, boiler plate selection code may be replaced with
func.Client(ctx) or func.Get(ctx).