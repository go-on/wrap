// Copyright (c) 2014 Marc René Arns. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

/*
Package wrap creates a fast and flexible middleware stack for http.Handlers.

Each middleware is a wrapper for another middleware and fullfills the
Wrapper interface.

A nice introduction into this library can be found at http://metakeule.github.io/article/wrap-go-middlware-framework.html.

Features

- small; core is only 13 LOC
- based on http.Handler interface; integrates fine with net/http
- middleware stacks are http.Handlers too and may be embedded
- low memory footprint
- fast

Wrappers can be found at github.com/go-on/wrap-contrib.

Status

100% test coverage.
This package is considered complete and the API is stable.

Benchmarks (Go 1.3):

	// The overhead of n writes to http.ResponseWriter via n wrappers
	// vs n writes in a loop within a single http.Handler

	BenchmarkServing2Simple     1000000 1067    ns/op   1,00x
  BenchmarkServing2Wrappers   1000000 1121    ns/op   1,05x

  BenchmarkServing50Simple    100000  26041   ns/op   1,00x
  BenchmarkServing50Wrappers  100000  27053   ns/op   1,04x

  BenchmarkServing100Simple   50000   52074   ns/op   1,00x
  BenchmarkServing100Wrappers 50000   53450   ns/op   1,03x

Credits

Initial inspiration came from Christian Neukirchen's rack for ruby some years ago.

*/
package wrap
