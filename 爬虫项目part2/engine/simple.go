package engine

import (

	"log"
	"爬虫项目part2/fetcher"
)
type SimpleEngine struct {}

// input: Request
func (e SimpleEngine)Run(seeds ...Request) {
	var requests []Request
	for _, r := range seeds {
		requests = append(requests, r)
	}
	for len(requests) > 0 {
		r := requests[0]
		requests = requests[1:]
		parseResult, err := Worker(r)
		if err != nil {
			continue
		}
		requests = append(requests, parseResult.Requests...) // ...就是把slice的内容展开加进去
		for _, item := range parseResult.Items {
			log.Printf("Got item %v", item)
		}

	}
}

func Worker(r Request) (ParseResult,error){
	log.Printf("Fetching %s", r.Url)
	body, err := fetcher.Fetch(r.Url)
	if err != nil {
		log.Printf("Fetcher: error"+"fetching url %s: %v", r.Url, err)
		return ParseResult{}, err
	}
	return r.ParserFunc(body), nil //ParserFunc(body) = ParseCityList(body)
}
