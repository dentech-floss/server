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

import (
	"context"

	"github.com/dentech-floss/logging/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor that adds the
// "grpc.method" and "request" log fields to the context for downstream handlers.
// The "request" field is only added for protobuf requests.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = logging.ContextWithFields(ctx, logging.String("grpc.method", info.FullMethod))
		if msg, ok := req.(proto.Message); ok {
			ctx = logging.ContextWithFields(ctx, logging.Proto("request", msg))
		}
		return handler(ctx, req)
	}
}
