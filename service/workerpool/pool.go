package workerpool

import (
	"fmt"
	"sync"
)

type Pool struct {
	name      string
	tasksChan chan Worker
	wg        sync.WaitGroup
}

func NewPool(maxGoroutines int, buffered uint, name string) *Pool {
	workerChain := make(chan Worker)
	if buffered != 0 {
		workerChain = make(chan Worker, buffered)
	}

	p := Pool{
		name:      name,
		tasksChan: workerChain,
	}

	p.wg.Add(maxGoroutines)
	//fmt.Println(fmt.Sprintf("Start [%v] with %d goroutine", name, maxGoroutines))
	for i := 0; i < maxGoroutines; i++ {
		//fmt.Println(fmt.Sprintf("GoRoutine [%v]: Start processing goroutine... %d", name, i))
		go p.work()
	}

	return &p
}

func (p *Pool) work() {
	for w := range p.tasksChan {
		fmt.Println(fmt.Sprintf("Task [%v]: Start execute task", p.name))
		w.Task()
		fmt.Println(fmt.Sprintf("Task [%v]: End execute task", p.name))
	}

	p.wg.Done()
}

// Run submits work to the pool.
func (p *Pool) Run(w Worker) {
	p.tasksChan <- w
}

// Shutdown waits for all the goroutines to shutdown.
func (p *Pool) Shutdown() {
	//fmt.Println(fmt.Sprintf("GoRoutine [%v]: Close channel", p.name))
	close(p.tasksChan)

	//wait taskchain work done
	p.wg.Wait()
}