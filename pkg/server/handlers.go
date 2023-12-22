package server

import (
	"net/http"
	"strings"
)

type HttpAndGrpcHandlerFunc func(
	httpHandler http.Handler,
	grpcHandler http.Handler,
	options *HttpAndGrpcHandlerOptions,
) http.Handler

func DefaultHttpAndGrpcHandlerFunc(
	httpHandler http.Handler,
	grpcHandler http.Handler,
	options *HttpAndGrpcHandlerOptions,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 &&
			strings.HasPrefix(r.Header.Get("content-type"), "application/grpc") {
			grpcHandler.ServeHTTP(w, r)
		} else {
			httpHandler.ServeHTTP(w, r)
		}
	})
}

func CorsHttpAndGrpcHandlerFunc(
	httpHandler http.Handler,
	grpcHandler http.Handler,
	options *HttpAndGrpcHandlerOptions,
) http.Handler {
	allowedHeaders := []string{
		"Accept",
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
		"X-User-Agent",
		"X-Grpc-Web",
		"Gprc-Status",
		"Gprc-Message",
		"Api-Token",
		"X-Auth-Token",
		"Traceparent",
	}
	if options != nil && options.AdditionalAllowedHeaders != nil {
		allowedHeaders = append(allowedHeaders, options.AdditionalAllowedHeaders...)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if options == nil || options.GetOrigin == nil {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", options.GetOrigin(r))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
		w.Header().
			Set("Access-Control-Expose-Headers", "Content-Length, Content-Range, grpc-status, grpc-message")
		w.Header().Set("Access-Control-Max-Age", "1728000")

		if r.Method == "OPTIONS" {
			return // respond to preflight requests with the above cors settings
		}

		DefaultHttpAndGrpcHandlerFunc(httpHandler, grpcHandler, options).ServeHTTP(w, r)
	})
}
