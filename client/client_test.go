package grpc_client

import (
	"context"
	grpc_server "github.com/apssouza22/grpc-server-go/server"
	"github.com/apssouza22/grpc-server-go/testdata"
	gtest "github.com/apssouza22/grpc-server-go/testing"
	"github.com/apssouza22/grpc-server-go/tlscert"
	"github.com/apssouza22/grpc-server-go/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"testing"
)

var server gtest.GrpcInProcessingServer

func startServer() {
	builder := gtest.GrpcInProcessingServerBuilder{}
	builder.SetUnaryInterceptors(util.GetDefaultUnaryServerInterceptors())
	server = builder.Build()
	server.RegisterService(func(server *grpc.Server) {
		helloworld.RegisterGreeterServer(server, &testdata.MockedService{})
	})
	server.Start()
}
func startServerWithTLS() grpc_server.GrpcServer {
	builder := grpc_server.GrpcServerBuilder{}
	builder.SetUnaryInterceptors(util.GetDefaultUnaryServerInterceptors())
	builder.SetTlsCert(&tlscert.Cert)
	svr := builder.Build()
	svr.RegisterService(func(server *grpc.Server) {
		helloworld.RegisterGreeterServer(server, &testdata.MockedService{})
	})
	svr.Start("localhost", 8989)
	return svr
}

func TestSayHelloPassingContext(t *testing.T) {
	startServer()
	ctx := context.Background()
	clientBuilder := GrpcClientBuilder{}
	clientBuilder.WithInsecure()
	clientBuilder.WithContext(ctx)
	clientBuilder.WithOptions(grpc.WithContextDialer(gtest.GetBufDialer(server.GetListener())))
	clientConn, err := clientBuilder.GetConn("localhost:50051")

	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer clientConn.Close()
	client := helloworld.NewGreeterClient(clientConn)
	request := &helloworld.HelloRequest{Name: "test"}
	resp, err := client.SayHello(ctx, request)
	if err != nil {
		t.Fatalf("SayHello failed: %v", err)
	}
	server.Cleanup()
	clientConn.Close()
	assert.Equal(t, resp.Message, "This is a mocked service test")
}

func TestSayHelloNotPassingContext(t *testing.T) {
	startServer()
	ctx := context.Background()
	clientBuilder := GrpcClientBuilder{}
	clientBuilder.WithInsecure()
	clientBuilder.WithOptions(grpc.WithContextDialer(gtest.GetBufDialer(server.GetListener())))
	clientConn, err := clientBuilder.GetConn("localhost:50051")

	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer clientConn.Close()
	client := helloworld.NewGreeterClient(clientConn)
	request := &helloworld.HelloRequest{Name: "test"}
	resp, err := client.SayHello(ctx, request)
	if err != nil {
		t.Fatalf("SayHello failed: %v", err)
	}
	server.Cleanup()
	clientConn.Close()
	assert.Equal(t, resp.Message, "This is a mocked service test")
}

func TestTLSConnWithCert(t *testing.T) {
	serverWithTLS := startServerWithTLS()
	defer serverWithTLS.GetListener().Close()

	ctx := context.Background()
	clientBuilder := GrpcClientBuilder{}
	clientBuilder.WithContext(ctx)
	clientBuilder.WithBlock()
	clientBuilder.WithClientTransportCredentials(false, tlscert.CertPool)
	clientConn, _ := clientBuilder.GetTlsConn("localhost:8989")
	defer clientConn.Close()
	client := helloworld.NewGreeterClient(clientConn)
	request := &helloworld.HelloRequest{Name: "test"}
	resp, err := client.SayHello(ctx, request)
	assert.NoError(t, err)
	assert.Equal(t, resp.Message, "This is a mocked service test")
}

func TestTLSConnWithInsecure(t *testing.T) {
	serverWithTLS := startServerWithTLS()
	defer serverWithTLS.GetListener().Close()

	ctx := context.Background()
	clientBuilder := GrpcClientBuilder{}
	clientBuilder.WithContext(ctx)
	clientBuilder.WithBlock()
	clientBuilder.WithClientTransportCredentials(true, nil)
	clientConn, _ := clientBuilder.GetTlsConn("localhost:8989")
	defer clientConn.Close()
	client := helloworld.NewGreeterClient(clientConn)
	request := &helloworld.HelloRequest{Name: "test"}
	resp, err := client.SayHello(ctx, request)
	assert.NoError(t, err)
	assert.Equal(t, resp.Message, "This is a mocked service test")
}
