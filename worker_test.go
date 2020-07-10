package worker

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerManagerOnCoWork(t *testing.T) {
	must := assert.New(t)
	manager := NewWorkerManager()
	manager.Register(
		NewWorker("p1"),
		NewWorker("p2"),
	)
	go manager.RunOnCoWork()

	for i := 0; i < 100; i++ {
		err := manager.Consume(i)
		must.Nil(err)
	}
	manager.Exit()
}

func TestWorkerManagerOnDistribute(t *testing.T) {
	must := assert.New(t)
	manager := NewWorkerManager()
	manager.Register(
		NewWorker("p1"),
		NewWorker("p2"),
	)
	go manager.RunOnDistribute()

	for i := 0; i < 100; i++ {
		err := manager.Consume(i)
		must.Nil(err)
	}
	manager.Exit()
}

func NewWorker(name string) Worker {
	return &MultiOutput{
		Name: name,
		ch:   make(chan interface{}, 100),
		done: make(chan struct{}),
	}
}

type MultiOutput struct {
	Name string
	ch   chan interface{}
	done chan struct{}
}

func (mo *MultiOutput) Run() {
	for msg := range mo.ch {
		fmt.Printf("%s: %v\n", mo.Name, msg)
	}
	mo.done <- struct{}{}
}

func (mo *MultiOutput) Close() {
	close(mo.ch)
}

func (mo *MultiOutput) Done() {
	<-mo.done
}

func (mo *MultiOutput) Consume(message interface{}) error {
	select {
	case mo.ch <- message:
		return nil
	case <-time.After(time.Second * 1):
		return errors.New("consumming channel is overloaded")
	}
}
