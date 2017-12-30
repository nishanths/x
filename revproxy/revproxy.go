// Package revproxy has utility functions for reverse proxies.
package revproxy

import (
	"errors"
	"net/http"
	"net/url"
)

// TargetFunc returns the target URL to use for the incoming request.
type TargetFunc func(r *http.Request) (*url.URL, error)

// ByHost determines the target URL's host based on the incoming request's
// host. m maps the host of an incoming request to its target host.
// The scheme for the target URL is always "http".
func ByHost(m map[string]string) TargetFunc {
	return func(r *http.Request) (*url.URL, error) {
		v, ok := m[r.Host]
		if !ok {
			// Unknown host. Either the map should be fixed
			// to include an entry for the host, or DNS settings
			// for the domain should be updated to not point
			// to the IP.
			return nil, errors.New("unknown host: " + r.Host)
		}
		return &url.URL{
			Scheme:   "http",
			Host:     v,
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
		}, nil
	}
}

// MultiHostDirector returns a function that can be used as the Director
// in httputil.ReverseProxy. The supplied TargetFunc determines which target
// URL the incoming request should directed to. If TargetFunc returns an
// error, the director panics with that error.
func MultiHostDirector(t TargetFunc) func(r *http.Request) {
	return func(r *http.Request) {
		target, err := t(r)
		if err != nil {
			panic(err)
		}
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host
		r.URL.Path = target.Path
		r.URL.RawQuery = target.RawQuery
	}
}
