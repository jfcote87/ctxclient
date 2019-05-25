# ctxclient package

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/jfcote87/ctxclient)
[![BSD 3-Clause License](https://img.shields.io/badge/license-bsd%203--clause-blue.svg)](https://github.com/jfcote87/ctxclient/blob/master/LICENSE)

Package ctxclient offers utilities for handling the
selection and creation of http.Clients based on
the context.  To remove boiler plate code checking
for 2xx statuses and handling of timeouts, the Func.Do
method provides these functions for http requests.

See examples.

This package borrows from ideas found in
golang.org/x/oauth2.
