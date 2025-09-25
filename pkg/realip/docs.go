// Package realip provides utilities for extracting the real client IP address from gRPC requests
// in Google Cloud Run environments.
//
// This package offers a gRPC unary server interceptor that reads the X-Forwarded-For header
// and injects the real client IP (the first IP in the header) into the request context.
// Cloud Run guarantees that the first IP in X-Forwarded-For is the actual client IP.
//
// Usage:
//
//	import "github.com/dentech-floss/server/pkg/realip"
//
//	grpc.NewServer(grpc.UnaryInterceptor(realip.UnaryServerInterceptor()))
//
// The real client IP can be retrieved in handlers using realip.FromContext(ctx).
package realip
