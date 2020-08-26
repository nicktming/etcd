package main

import (
	"go.etcd.io/etcd/myembed"
	"log"
	"go.etcd.io/etcd/myetcdserver/api/v3client"
	"context"
	"time"
)

func main() {
	cfg := myembed.NewConfig()
	cfg.Dir = "default.etcd"

	e, err := myembed.StartEtcd(cfg)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("++++++++e: %v\n", e)

	time.Sleep(3 * time.Second)

	cli := v3client.New(e.Server)

	resp, err := cli.Put(context.TODO(), "some-key", "it workds!")

	if err != nil {
		// handle error
	}

	log.Printf("resp： %v\n", resp)


	stopch := make(chan struct{})
	<- stopch
}


