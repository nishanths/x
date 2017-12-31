// Commands tlsserve terminates TLS and proxies requests to the
// backend server that is usually running in a separate process.
package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/nishanths/x/revproxy"
)

var (
	httpsAddr = flag.String("https", ":443", "https service address")
	httpAddr  = flag.String("http", ":80", "http service address")
	certFile  = flag.String("cert", "/usr/local/etc/letsencrypt/live/avalanche.space/fullchain.pem", "path to chained cert file")
	keyFile   = flag.String("key", "/usr/local/etc/letsencrypt/live/avalanche.space/privkey.pem", "path to private key file")
)

// NOTE: the supplied certificate must support all the domains listed here.
var knownHTTPSDomains = map[string]struct{}{
	"avalanche.space":     struct{}{},
	"www.avalanche.space": struct{}{},
	"x.avalanche.space":   struct{}{},
}

var targets = revproxy.ByHost(
	map[string]string{
		"avalanche.space":     "localhost:8090",
		"www.avalanche.space": "localhost:8090",
		"x.avalanche.space":   "localhost:8091",
	},
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.Parse()

	// Redirect HTTP to HTTPS.
	go func() {
		handler := &whitelistedHandler{http.HandlerFunc(redirectHTTPS), knownHTTPSDomains}
		log.Fatal(http.ListenAndServe(*httpAddr, handler))
	}()

	// HTTPS handler.
	p := httputil.ReverseProxy{Director: revproxy.MultiHostDirector(targets)}
	handler := &whitelistedHandler{&p, knownHTTPSDomains}
	log.Fatal(http.ListenAndServeTLS(*httpsAddr, *certFile, *keyFile, handler))
}

// whitelistedHandler is a http.Handler that responds with 404s for
// requests with hosts not in the map.
//
// Mainly useful for guarding against outdated DNS settings through which
// requests for a domain arrive here even though the domain is no longer
// handled by this server.
type whitelistedHandler struct {
	h http.Handler
	d map[string]struct{}
}

func (h *whitelistedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.d[r.Host]; !ok {
		http.NotFound(w, r)
		return
	}
	h.h.ServeHTTP(w, r)
}

func redirectHTTPS(w http.ResponseWriter, r *http.Request) {
	u := *r.URL
	u.Scheme = "https"
	u.Host = r.Host
	http.Redirect(w, r, u.String(), http.StatusFound)
}
