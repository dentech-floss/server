# server

Since we only have one port on Cloud Run, but want to serve both the "native" gRPC server and the [gRPC-Gateway](https://github.com/grpc-ecosystem/grpc-gateway) we have applied the ideas from here: [Serving gRPC+HTTP/2 from the same Cloud Run container](https://ahmet.im/blog/grpc-http-mux-go/) to get this to work.

Opentelemetry instrumentation is configured for the gRPC server via the [otelgrpc](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc) lib to intercept both unary and stream communication. 

// TODO: continue...
