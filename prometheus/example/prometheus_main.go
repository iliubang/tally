// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/liubang/tally/v3"
	promreporter "github.com/liubang/tally/v3/prometheus"
)

func main() {
	reporter := promreporter.NewReporter(promreporter.Options{})
	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Prefix:         "service_router_test",
		Tags:           map[string]string{"name": "liubang_test"},
		CachedReporter: reporter,
		Separator:      promreporter.DefaultSeparator,
	}, time.Second)

	defer closer.Close()
	counter := scope.Tagged(map[string]string{"counter": "bar"}).Counter("test_counter")
	gauge := scope.Tagged(map[string]string{"gauge": "aaa"}).Gauge("test_gauge")
	timer := scope.Tagged(map[string]string{"timer": "bbb"}).Timer("test_timer")
	bucket := tally.DefaultBuckets
	histogram := scope.Tagged(map[string]string{"histogram": "ccc"}).Histogram("test_histogram", bucket)
	meter := scope.Meter("test_meter")
	meter1 := scope.Meter("test_meter")
	_ = meter1
	meter2 := scope.Tagged(map[string]string{"aaa": "bbb"}).Meter("test_meter2")
	meter3 := scope.Tagged(map[string]string{"aaa": "ccc"}).Meter("test_meter2")

	go func() {
		for {
			counter.Inc(1)
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			gauge.Update(rand.Float64() * 1000)
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			tsw := timer.Start()
			hsw := histogram.Start()
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
			meter.Mark(1)
			meter2.Mark(1)
			meter3.Mark(1)
			tsw.Stop()
			hsw.Stop()
		}
	}()

	http.Handle("/metrics", reporter.HTTPHandler())
	http.Handle("/json", reporter.JsonHTTPHandler())
	http.ListenAndServe(":8080", nil)
	select {}
}
