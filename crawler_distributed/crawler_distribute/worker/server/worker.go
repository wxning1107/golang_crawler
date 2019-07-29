package main

import (
	"crawler_distribute/rpcsupport"
	"fmt"
	"crawler_distribute/worker"
	"github.com/gpmgo/gopm/modules/log"
	"flag"
)

var port = flag.Int("port", 0, "the port for me to listen on")

func main() {
	flag.Parse()
	if *port == 0 {
		fmt.Println("must specify a port")
		return
	}
	log.Fatal("%v", rpcsupport.ServeRpc(fmt.Sprintf(":%d", *port), worker.CrawlService{}))
	
}
