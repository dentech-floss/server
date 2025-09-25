package realip

import (
	"context"
	"net/netip"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// mockHandler is a simple gRPC handler for testing.
func mockHandler(t *testing.T, wantIP netip.Addr) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		ip, ok := FromContext(ctx)
		if !ok {
			t.Errorf("expected IP in context, got none")
		}
		if ip != wantIP {
			t.Errorf("expected IP %v, got %v", wantIP, ip)
		}
		return "ok", nil
	}
}

func TestUnaryServerInterceptor_XForwardedFor(t *testing.T) {
	wantIP := netip.MustParseAddr("203.0.113.42")
	md := metadata.New(map[string]string{
		XForwardedFor: "203.0.113.42, 70.41.3.18, 150.172.238.178",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	interceptor := UnaryServerInterceptor()
	handler := mockHandler(t, wantIP)

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
}

func TestUnaryServerInterceptor_NoXForwardedFor(t *testing.T) {
	// No X-Forwarded-For, so expect noIP
	ctx := context.Background()
	interceptor := UnaryServerInterceptor()
	handler := mockHandler(t, noIP)

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
}

func TestFromContext_NotSet(t *testing.T) {
	ctx := context.Background()
	ip, ok := FromContext(ctx)
	if ok {
		t.Errorf("expected no IP in context, got %v", ip)
	}
}
