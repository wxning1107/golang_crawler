package main

import (
	"crawler/engine"
	"crawler/scheduler"
	itemsaver "crawler_distribute/persist/client"
	"crawler_distribute/config"
	"crawler/zhenai/parser"
	worker "crawler_distribute/worker/client"
	"net/rpc"
	"crawler_distribute/rpcsupport"
	"log"
	"flag"
	"strings"
)

var (
	itemSaverHost = flag.String("itemsaver_host", "", "itemsaver host")
	workerHosts = flag.String("worker_host", "", "worker host (comma separated)")
)

func main() {
	flag.Parse()
	itemChan, err := itemsaver.ItemSaver(*itemSaverHost)
	if err != nil {
		panic(err)
	}

	pool := createClientPool(strings.Split(*workerHosts, ","))
	processor := worker.CreateProcessor(pool)

	e := engine.ConcurrentEngine{&scheduler.QueuedScheduler{}, 100, itemChan, processor}
	e.Run(engine.Request{Url: "http://www.zhenai.com/zhenghun", Parser: engine.NewFuncParser(parser.ParseCityList, config.ParseCity)})
	//e.Run(engine.Request{Url: "http://www.zhenai.com/zhenghun/shanghai", ParserFunc: parser.ParseCity})
}

func createClientPool(hosts []string) chan *rpc.Client {
	var clients []*rpc.Client
	for _, h := range hosts {
		client, err:= rpcsupport.NewClient(h)
		if err == nil {
			clients = append(clients, client)
			log.Printf("connected to %s", h)
		}else {
			log.Printf("error connecting to %s: %v", h, err)
		}
	}

	out := make(chan *rpc.Client)
	go func() {
		for {
			for _, client := range clients {
				out <- client //消息传递
			}
		}
	}()
	return out
}
