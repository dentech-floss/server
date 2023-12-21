package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

type HttpAndGrpcHandlerOptions struct {
	GetOrigin func(r *http.Request) string
}

type ServerConfig struct {
	Port                   int
	HttpAndGrpcHandlerFunc HttpAndGrpcHandlerFunc
	HandlerOptions         *HttpAndGrpcHandlerOptions
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

	grpcMux := runtime.NewServeMux() // grpc-gateway

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),   // instrumentation
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()), // instrumentation
	)

	// Serve both the gRPC server and the http/json proxy on the same port
	httpServer := &http.Server{
		Handler: h2c.NewHandler(
			config.HttpAndGrpcHandlerFunc(
				grpcMux,
				grpcServer,
				config.HandlerOptions,
			),
			&http2.Server{},
		),
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
