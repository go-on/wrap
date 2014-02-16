// Copyright (c) 2014 Marc René Arns. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

/*
Package wrap creates fast and flexible middleware for http.Handlers.

It has probably the most simple implementation one could think of.
Each middleware is a wrapper for another middleware and fullfills the
Wrapper interface.

This package is considered stable, feature complete and ready for production.

A lot of wrappers can be found at github.com/go-on/wrap-contrib/wraps.

Benchmarks:

	// The overhead of n writes to http.ResponseWriter via n wrappers
	// vs n writes in a loop within a single http.Handler

	BenchmarkServing2Simple     5000000   718 ns/op 1.00x
	BenchmarkServing2Wrappers   2000000   824 ns/op 1.14x

	BenchmarkServing50Simple     100000 17466 ns/op 1.00x
	BenchmarkServing50Wrappers   100000 23984 ns/op 1.37x

	BenchmarkServing100Simple     50000 33686 ns/op 1.00x
	BenchmarkServing100Wrappers   50000 46676 ns/op 1.39x



Credits

Initial inspiration came from Christian Neukirchen's rack for ruby some years ago.

*/
package wrap
