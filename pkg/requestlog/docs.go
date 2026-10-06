// Package requestlog adds log fields describing the current gRPC call to the
// request context.
//
// This package offers a gRPC unary server interceptor that adds the
// "grpc.method" and "request" fields with logging.ContextWithFields. Every
// entry logged with the context, through a *Context method of a
// dentech-floss/logging logger, includes them. The request is logged with
// logging.Proto, so fields marked with [debug_redact = true] in the .proto
// schema are left out.
//
// Usage:
//
//	import "github.com/dentech-floss/server/pkg/requestlog"
//
//	grpc.NewServer(grpc.ChainUnaryInterceptor(requestlog.UnaryServerInterceptor()))
//
// With server.NewServer, set WithRequestLogFields in ServerConfig instead.
package requestlog
