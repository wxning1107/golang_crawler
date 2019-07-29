package worker

import (
	"crawler/engine"
	"crawler_distribute/config"
	"crawler/zhenai/parser"
	"github.com/pkg/errors"
	"fmt"
	"log"
)

type SerializedParser struct {
	Name string
	Args interface{}
}

//{"parseCityList", nil}, {"ProfileParser", userName}

type Request struct {
	Url    string
	Parser SerializedParser
}

type ParseResult struct {
	Items    []engine.Item
	Requests []Request
}

func SerializeRequest(r engine.Request) Request {
	name, args := r.Parser.Serialize()
	return Request{r.Url, SerializedParser{name, args}}
}

func SerializeResult(r engine.ParseResult) ParseResult {
	Requests := []Request{}
	result := ParseResult{ r.Items, Requests}

	for _, req := range r.Requests {
		result.Requests = append(result.Requests, SerializeRequest(req))
	}
	return result
}

func DeserializeRequest(r Request) (engine.Request, error) {
	parser, err := DeserializeParser(r.Parser)
	if err != nil {
		return engine.Request{}, err
	}
	return engine.Request{r.Url, parser}, nil
}

func DeserializeResult(r ParseResult) engine.ParseResult {
	requests := []engine.Request{}
	result := engine.ParseResult{requests, r.Items}
	for _, req := range r.Requests {
		engineReq, err := DeserializeRequest(req)
		if err != nil {
			log.Printf("error deserializing " + "request: %v", err)
			continue
		}
		result.Requests = append(result.Requests, engineReq)
	}
	return result
}

func DeserializeParser(p SerializedParser) (engine.Parser, error) {
	switch p.Name {
	case config.ParseCityList:
		return engine.NewFuncParser(parser.ParseCityList, config.ParseCityList), nil
	case config.ParseCity:
		return engine.NewFuncParser(parser.ParseCity, config.ParseCity), nil
	case config.NilParser:
		return engine.NilParser{}, nil
	case config.ParseProfile:
		if userName, ok := p.Args.(string); ok {
			return parser.NewProfileParser(userName), nil
		}else {
			return nil, fmt.Errorf("invalid " + "arg: %v", p.Args)
		}

	default:
		return nil, errors.New("unknown parser name")

	}
}


