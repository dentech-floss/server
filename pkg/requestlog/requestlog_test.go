package requestlog

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/dentech-floss/logging/pkg/logging"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const fullMethod = "/test.v1.TestService/DoSomething"

// runInterceptor calls the interceptor with a handler that logs one entry, and
// returns that entry.
func runInterceptor(t *testing.T, req interface{}) map[string]any {
	var buf bytes.Buffer
	logger := logging.NewLogger(&logging.LoggerConfig{
		ServiceName: "test-service",
		MinLevel:    logging.DebugLevel,
		Output:      &buf,
	})

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		logger.InfoContext(ctx, "handled")
		return "ok", nil
	}

	resp, err := UnaryServerInterceptor()(context.Background(), req, &grpc.UnaryServerInfo{FullMethod: fullMethod}, handler)
	if err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if resp != "ok" {
		t.Errorf("expected the handler's response, got %v", resp)
	}

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log entry: %v", err)
	}
	return entry
}

func TestUnaryServerInterceptor_ProtoRequest(t *testing.T) {
	entry := runInterceptor(t, wrapperspb.String("hello"))

	if got := entry["grpc.method"]; got != fullMethod {
		t.Errorf("expected grpc.method %v, got %v", fullMethod, got)
	}
	if got := entry["request"]; got != "hello" {
		t.Errorf("expected request %v, got %v", "hello", got)
	}
}

func TestUnaryServerInterceptor_NonProtoRequest(t *testing.T) {
	entry := runInterceptor(t, "not a proto message")

	if got := entry["grpc.method"]; got != fullMethod {
		t.Errorf("expected grpc.method %v, got %v", fullMethod, got)
	}
	if got, ok := entry["request"]; ok {
		t.Errorf("expected no request field, got %v", got)
	}
}
