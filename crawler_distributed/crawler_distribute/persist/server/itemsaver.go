package main

import (
	"gopkg.in/olivere/elastic.v5"
	"crawler_distribute/rpcsupport"

	persist2 "crawler_distribute/persist"
	"github.com/gpmgo/gopm/modules/log"
	"fmt"
	"crawler_distribute/config"
	"flag"
)

var port = flag.Int("port", 0, "the port for me to listen on")

func main() {
	flag.Parse()
	if *port == 0 {
		fmt.Println("must specify a port")
		return
	}
	log.Fatal("%s", serveRpc(fmt.Sprintf(":%d", *port), config.ElasticIndex))
}

func serveRpc(host, index string) error {
	client, err := elastic.NewClient(elastic.SetSniff(false))
	if err != nil {
		return err
	}

	return  rpcsupport.ServeRpc(host, &persist2.ItemSaverService{Client:client, Index:index})

}
