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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if options == nil {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", getOrigin(r, options.AllowedOrigins))
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, "+
			"Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, x-user-agent, "+
			"x-grpc-web, grpc-status, grpc-message, api-token, X-Auth-Token")
		w.Header().
			Set("Access-Control-Expose-Headers", "Content-Length, Content-Range, grpc-status, grpc-message")
		w.Header().Set("Access-Control-Max-Age", "1728000")

		if r.Method == "OPTIONS" {
			return // respond to preflight requests with the above cors settings
		}

		DefaultHttpAndGrpcHandlerFunc(httpHandler, grpcHandler, options).ServeHTTP(w, r)
	})
}

func getOrigin(r *http.Request, allowedOrigins []string) string {
	if len(allowedOrigins) == 0 {
		return "*"
	}
	origin := r.Header.Get("Origin")
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return origin
		}
	}
	return ""
}
