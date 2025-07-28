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
	GetOrigin                func(r *http.Request) string
	AdditionalAllowedHeaders []string
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
	Port int
	// Actual port used by the server, useful for when the port is 0 (random port)
	ActualPort int
	GrpcMux    *runtime.ServeMux
	GrpcServer *grpc.Server
	HttpServer *http.Server
}

func NewServer(config *ServerConfig) *Server {
	config.setDefaults()

	grpcMux := runtime.NewServeMux() // grpc-gateway

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(
			otelgrpc.NewServerHandler(),
		),
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

// Start starts the server asynchronously (non-blocking) and sets ActualPort.
// Useful in tests where dynamic ports can be beneficial.
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(s.Port))
	if err != nil {
		return err
	}

	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		s.ActualPort = tcpAddr.Port
	}

	go func() {
		err := s.HttpServer.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	return nil
}

// Serve starts the server synchronously (blocking), typically used in production.
// Keeps compatibility with existing code.
func (s *Server) Serve() {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(s.Port))
	if err != nil {
		panic(err)
	}

	if tcpAddr, ok := listener.Addr().(*net.TCPAddr); ok {
		s.ActualPort = tcpAddr.Port
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
