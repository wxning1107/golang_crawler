package main

import (
	"爬虫项目part2/engine"
	"爬虫项目part2/scheduler"
	"爬虫项目part2/zhenai/parser"
)

func main() {
	e := engine.ConcurrentEngine{&scheduler.QueuedScheduler{}, 100}
	//e.Run(engine.Request{Url: "http://www.zhenai.com/zhenghun", ParserFunc: parser.ParseCityList})
	e.Run(engine.Request{Url: "http://www.zhenai.com/zhenghun/shanghai", ParserFunc: parser.ParseCity})

}
