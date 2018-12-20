/*
Copyright (C)  2018 Yahoo Japan Corporation Athenz team.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package router

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/kpango/glg"
	"github.com/yahoojapan/garm/config"
	"github.com/yahoojapan/garm/handler"
)

// New returns ServeMux with routes using given handler.
func New(cfg config.Server, h handler.Handler) *http.ServeMux {

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 32

	mux := http.NewServeMux()
	dur := parseTimeout(cfg.Timeout)

	// register (route, handler) tuple to server multiplexer
	for _, route := range NewRoutes(h) {
		mux.Handle(route.Pattern, routing(route.Methods, dur, route.HandlerFunc))
	}

	return mux
}

// parseTimeout parses string to time.Duration.
// If there is any errors, return 3s as Duration.
func parseTimeout(timeout string) time.Duration {
	dur, err := time.ParseDuration(timeout)
	if err != nil {
		dur = time.Second * 3
	}
	return dur
}

// routing wraps the handler.Func and returns a new http.Handler.
// routing helps to handle unsupported HTTP method, timeout, and the error returned from the handler.Func.
func routing(m []string, t time.Duration, h handler.Func) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, method := range m {
			if strings.EqualFold(r.Method, method) {
				// execute only if the request method is inside the method list

				// context for timeout
				ctx, cancel := context.WithTimeout(r.Context(), t)
				defer cancel()
				start := time.Now()

				// run the custom handler logic in go routine, report error to error channel
				ech := make(chan error)
				go func() {
					// it is the responsibility for handler to close the request
					ech <- h(w, r.WithContext(ctx))
				}()

				for {
					select {
					case err := <-ech:
						// handler finished first, may have error returned
						if err != nil {
							http.Error(w,
								fmt.Sprintf("Error: %s\t%s",
									err.Error(),
									http.StatusText(http.StatusInternalServerError)),
								http.StatusInternalServerError)
							glg.Error(err)
						}
						return
					case <-ctx.Done():
						// timeout passed or parent context canceled first, it is the responsibility for handler to response to the user
						err := glg.Errorf("Handler Time Out: %v", time.Since(start))
						if err != nil {
							glg.Fatal(err)
						}
						return
					}
				}
			}
		}

		// flush and close the request body; for GET method, r.Body may be nil
		err := flushAndClose(r.Body)
		if err != nil {
			// exit the program here
			glg.Fatalln(err)
		}

		http.Error(w,
			fmt.Sprintf("Method: %s\t%s",
				r.Method,
				http.StatusText(http.StatusMethodNotAllowed)),
			http.StatusMethodNotAllowed)
	})
}

// flushAndClose helps to flush and close a ReadCloser. Used for request body internal.
// Returns if there is any errors.
func flushAndClose(rc io.ReadCloser) error {
	if rc != nil {
		// flush
		_, err := io.Copy(ioutil.Discard, rc)
		if err != nil {
			return err
		}
		// close
		return rc.Close()
	}
	return nil
}
