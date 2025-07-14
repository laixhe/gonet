package grpc

import (
	goGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn *goGrpc.ClientConn
}

func NewClient(addr string) (*Client, error) {
	opts := make([]goGrpc.DialOption, 0, 5)
	// opts = append(opts, goGrpc.WithUnaryInterceptor(nil))
	opts = append(opts, goGrpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := goGrpc.NewClient(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}
