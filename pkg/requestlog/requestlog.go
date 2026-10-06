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
