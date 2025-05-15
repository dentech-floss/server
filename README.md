# server

Since we only have one port on Cloud Run, but want to serve both the "native" gRPC server and the [gRPC-Gateway](https://github.com/grpc-ecosystem/grpc-gateway) we have applied the ideas from here: [Serving gRPC+HTTP/2 from the same Cloud Run container](https://ahmet.im/blog/grpc-http-mux-go/) to get this to work.

Opentelemetry instrumentation is configured for the gRPC server via the [otelgrpc](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc) lib to intercept both unary and stream communication. 

Which handler to use is configurable, provide your implementation of choice by setting "HttpAndGrpcHandlerFunc" in the config. If this is not provided then the "DefaultHttpAndGrpcHandlerFunc" is used, and we also provide a "CorsHttpAndGrpcHandlerFunc" that handles preflight requests so have a closer look at: [handlers.go](https://github.com/dentech-floss/server/blob/master/pkg/server/handlers.go).

## Install

```
go get github.com/dentech-floss/server@v0.2.8
```

## Usage

```go
package example

import (
    "github.com/dentech-floss/server/pkg/server"
)

func main() {

    ctx := context.Background()

    ...

    appointmentServiceV1 := service.NewAppointmentServiceV1(repo, publisher, logger)
    patientGatewayServiceV1 := service.NewPatientGatewayServiceV1(repo, publisher, logger)

    ...

    server := server.NewServer(
        &server.ServerConfig{
            Port: config.Port,
            // The "server.DefaultHttpAndGrpcHandlerFunc" will be used if you don't set this
            // HttpAndGrpcHandlerFunc: server.CorsHttpAndGrpcHandlerFunc // ...for cors...
        },
    )

    appointmentServiceV1.Register(server.GrpcServer)
    appointmentServiceV1.RegisterMux(ctx, server.GrpcMux) // if you use the grpc gateway

    patientGatewayServiceV1.Register(server.GrpcServer)
    patientGatewayServiceV1.RegisterMux(ctx, server.GrpcMux)

    go server.Serve()
    handleShutdown(server, logger)
}

func handleShutdown(
    server *server.Server,
    logger *logging.Logger,
) {
    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
    sig := <-done // Block until we receive our signal
    logger.Info("Got signal: " + sig.String())

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := server.Shutdown(ctx); err != nil {
        logger.Warn("Server shutdown failed: " + err.Error())
    }
}
```
