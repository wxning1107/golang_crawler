package scheduler

import "爬虫项目part2/engine"

type SimpleScheduler struct {//实现了scheduler接口
	workerChan chan engine.Request
}

func (s *SimpleScheduler) WorkerChan() chan engine.Request {
	return s.workerChan
}

func (s *SimpleScheduler) WorkerReady(chan engine.Request) {

}

func (s *SimpleScheduler) Run() {
	s.workerChan = make(chan engine.Request)
}

func (s *SimpleScheduler)Submit(r engine.Request) {
	// 这里有个循环卡死的问题，改用goroutine来做
	// 每个Request会建一个goroutine，这个goroutine往统一的channel(in chan)里分发Request
	go func() {s.workerChan <- r}()
}