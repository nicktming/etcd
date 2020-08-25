package clientv3

import "context"

type Client struct {

	KV

	ctx 		context.Context
	cancel 		context.CancelFunc
}

func NewCtxClient(ctx context.Context) *Client {
	cctx, cancel := context.WithCancel(ctx)
	return &Client{ctx: cctx, cancel: cancel}
}
