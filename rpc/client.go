package rpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn *grpc.ClientConn
}

func NewClient(addr string) (*Client, error) {
	opts := make([]grpc.DialOption, 0, 5)
	// opts = append(opts, grpc.WithUnaryInterceptor(nil))
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}
