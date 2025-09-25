package realip

import (
	"context"
	"net/netip"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type realipKey struct{}

const (
	XForwardedFor = "X-Forwarded-For"
)

var noIP = netip.Addr{}

// UnaryServerInterceptor returns a gRPC unary server interceptor that extracts the real client IP address
// from the X-Forwarded-For header and injects it into the context for downstream handlers.
// This package is intended for use exclusively in Google Cloud Run environments, where Cloud Run guarantees
// that the first IP address in the X-Forwarded-For header is the actual client IP.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		clientIP := getClientIP(ctx)
		ctx = context.WithValue(ctx, realipKey{}, clientIP)
		return handler(ctx, req)
	}
}

func getClientIP(ctx context.Context) netip.Addr {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return getIPFromPeer(ctx)
	}

	forwardedFor := md.Get(XForwardedFor)
	if len(forwardedFor) == 0 {
		return getIPFromPeer(ctx)
	}

	ips := strings.Split(forwardedFor[0], ",")
	firstIP := strings.TrimSpace(ips[0])

	if ip, err := netip.ParseAddr(firstIP); err == nil {
		return ip
	}

	return getIPFromPeer(ctx)
}

func getIPFromPeer(ctx context.Context) netip.Addr {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return noIP
	}

	if addrPort, err := netip.ParseAddrPort(p.Addr.String()); err == nil {
		return addrPort.Addr()
	}

	return noIP
}

// FromContext retrieves the real client IP address from the context, if available.
// The IP is set by the UnaryServerInterceptor in this package.
// Returns the IP address and a boolean indicating whether the value was present in the context.
func FromContext(ctx context.Context) (netip.Addr, bool) {
	ip, ok := ctx.Value(realipKey{}).(netip.Addr)
	return ip, ok
}
