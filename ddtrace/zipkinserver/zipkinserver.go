package zipkinserver

import (
	"github.com/mailru/easyjson"
	traceformat "github.com/signalfx/golib/trace/format"
	"io"
	"net/http"
	"net/http/httptest"
)

// ZipkinServer is an embedded Zipkin server
type ZipkinServer struct {
	server *httptest.Server
	Spans  traceformat.Trace
}

// URL of the Zipkin server
func (z *ZipkinServer) URL() string {
	return z.server.URL+"/v1/trace"
}

// Stop the embedded Zipkin server
func (z *ZipkinServer) Stop() {
	z.server.Close()
}

// Reset received spans
func (z *ZipkinServer) Reset() {
	z.Spans = nil
}

// Start embedded Zipkin server
func Start() *ZipkinServer {
	zipkin := &ZipkinServer{}
	zipkin.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/trace" {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			if r.Header.Get("content-type") != "application/json" {
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}

			var trace traceformat.Trace

			if err := easyjson.UnmarshalFromReader(r.Body, &trace); err != nil {
				_, err = io.WriteString(w, err.Error())
				if err != nil {
					// Probably can't successfully write the err to the response so just
					// panic since this is used for testing.
					panic(err)
				}
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			zipkin.Spans = append(zipkin.Spans, trace...)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	return zipkin
}
