package server

import (
	"context"
	"errors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

type HttpAndGrpcHandlerOptions struct {
	GetOrigin                func(r *http.Request) string
	AdditionalAllowedHeaders []string
}

type ServerConfig struct {
	Port                   int
	HttpAndGrpcHandlerFunc HttpAndGrpcHandlerFunc
	HandlerOptions         *HttpAndGrpcHandlerOptions
	JsonEmitUnpopulated    bool
}

func (c *ServerConfig) setDefaults() {
	if c.HttpAndGrpcHandlerFunc == nil {
		c.HttpAndGrpcHandlerFunc = DefaultHttpAndGrpcHandlerFunc
	}
}

type Server struct {
	Port       int
	GrpcMux    *runtime.ServeMux
	GrpcServer *grpc.Server
	HttpServer *http.Server
}

func NewServer(config *ServerConfig) *Server {
	config.setDefaults()

	grpcMux := runtime.NewServeMux(runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated:   config.JsonEmitUnpopulated,
			EmitDefaultValues: config.JsonEmitUnpopulated,
			UseProtoNames:     true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})) // grpc-gateway

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()), // instrumentation
	)

	// Serve both the gRPC server and the http/json proxy on the same port
	httpServer := &http.Server{
		Handler: LoggingMiddleware(h2c.NewHandler(
			config.HttpAndGrpcHandlerFunc(
				grpcMux,
				grpcServer,
				config.HandlerOptions,
			),
			&http2.Server{},
		)),
	}

	return &Server{
		Port:       config.Port,
		GrpcMux:    grpcMux,
		GrpcServer: grpcServer,
		HttpServer: httpServer,
	}
}

func (s *Server) Serve() {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(s.Port))
	if err != nil {
		panic(err)
	}

	err = s.HttpServer.Serve(listener)
	if !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.GrpcServer != nil {
		s.GrpcServer.GracefulStop()
	}
	if s.HttpServer != nil {
		return s.HttpServer.Shutdown(ctx)
	}
	return nil
}

// LoggingMiddleware logs the incoming request URL
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		// Log the complete request URL
		// Log the complete request URL and additional details
		log.Printf("Request Method: %s", r.Method)
		log.Printf("Request URL: %s", r.URL.String())
		log.Printf("Request HTTP Version: %s", r.Proto)
		log.Printf("Request Headers: %v", r.Header)
		next.ServeHTTP(w, r)
	})
}
