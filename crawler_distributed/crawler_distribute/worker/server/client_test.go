package main

import (
	"testing"
	"crawler_distribute/rpcsupport"
	"crawler_distribute/worker"
	"time"
	"crawler_distribute/config"
	"fmt"
)

func TestCrawlService(t *testing.T) {
	const host = ":9000"
	go rpcsupport.ServeRpc(host, worker.CrawlService{})
	time.Sleep(time.Second)

	client, err := rpcsupport.NewClient(host)
	if err != nil {
		panic(err)
	}

	req := worker.Request{"http://album.zhenai.com/u/1552811555", worker.SerializedParser{config.ParseProfile, "芜湖小阿妹"}}
	var result worker.ParseResult
	err = client.Call(config.CrawlServiceRpc, req, &result)

	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(result)
	}
}
