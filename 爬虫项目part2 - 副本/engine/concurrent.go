package engine

import (
	"log"
)

type ConcurrentEngine struct {
	Scheduler   Scheduler
	WorkerCount int
}
type Scheduler interface {
	ReadyNotifier
	Submit(Request)
	WorkerChan() chan Request
	Run()
}

type ReadyNotifier interface {
	WorkerReady(chan Request)
}

func (e *ConcurrentEngine) Run(seeds ...Request) {
	out := make(chan ParseResult)
	 e.Scheduler.Run()


	for i := 0; i < e.WorkerCount; i++ {
		// simple是所有worker共用channel，queued是每个worker一个channel，到底怎么用问Scheduler.WorkerChan()
		createWorker(e.Scheduler.WorkerChan(), out, e.Scheduler) //疑问：in这个channel的输入是什么
	}
	// 往Scheduler里发任务
	for _, r := range seeds {
		// 1. 提交r到QueuedScheduler中的RequestChan中
		e.Scheduler.Submit(r)
	}
	profileCount := 0
	for {
		// 4. 从out中接收result
		result := <-out
		for _, item := range result.Items {
			log.Printf("Got profile #%d: %v", profileCount, item)
			profileCount++
		}
		for _, request := range result.Requests {
			e.Scheduler.Submit(request)
		}
	}

}
// 整个Scheduler作为参数太重了，把要用到的WorkerReady()分离出来
func createWorker(in chan Request, out chan ParseResult, ready ReadyNotifier) {

	go func() {
		for {
			// 1. tell scheduler i'm ready, 送in到QueuedScheduler里的WorkerChan
			ready.WorkerReady(in)
			// 2. 从in中收到request任务
			request := <-in
			// 收到任务后，worker开始工作
			result, err := Worker(request)
			if err != nil {
				continue
			}
			// 3. result送到out中
			out <- result
		}
	}()
}
