package worker

import (
	"errors"
	"log"
	"sync"
	"time"
)

// Worker is which you have to implement then register.
type Worker interface {
	// Run runs program on background.
	Run()
	// Consume receives message then sends to channel, requires non-blocking.
	Consume(message interface{}) error
	// Close send eixt signal.
	Close()
	// Done is a blocking func, wait background programs exit.
	Done()
}

const defaultConsumingLength = 10000

// Manager is a manager for all worker which registeied.
type Manager struct {
	Workers []Worker
	ch      chan interface{}
	running bool
}

// Register take over worker in manager.
func (wm *Manager) Register(w ...Worker) {
	wm.Workers = append(wm.Workers, w...)
}

// RunOnDistribute runs all worker on distributing mode.
func (wm *Manager) RunOnDistribute() {
	if wm.running {
		panic(errors.New("worker manager already runs"))
	}
	wm.running = true
	for _, w := range wm.Workers {
		go w.Run()
		defer w.Close()
	}

	for req := range wm.ch {
		for _, w := range wm.Workers {
			if err := w.Consume(req); err != nil {
				log.Printf("worker consume err: %s\n", err)
			}
		}
	}
}

// RunOnCoWork runs all worker on co-working mode.
func (wm *Manager) RunOnCoWork() {
	if wm.running {
		panic(errors.New("worker manager already runs"))
	}
	wm.running = true
	for _, w := range wm.Workers {
		go w.Run()
		defer w.Close()
	}

	var wg sync.WaitGroup
	for _, w := range wm.Workers {
		wg.Add(1)
		go func(w Worker) {
			for req := range wm.ch {
				if err := w.Consume(req); err != nil {
					log.Printf("worker consume err: %s\n", err)
				}
			}
			wg.Done()
		}(w)
	}
	wg.Wait()
}

// Exit exits all workers wait then all done.
func (wm *Manager) Exit() {
	close(wm.ch)
	for _, w := range wm.Workers {
		w.Done()
	}
}

// SetConsumingLength set manager max consuming channel length.
func (wm *Manager) SetConsumingLength(length int) error {
	if wm.running {
		return errors.New("worker already runing. cannot change consumming length")
	}
	wm.ch = make(chan interface{}, length)
	return nil
}

// Consume receive message dispatch to all workers.
func (wm *Manager) Consume(message interface{}) (err error) {
	select {
	case wm.ch <- message:
		return nil
	case <-time.After(time.Second * 1):
		return errors.New("worker queue is overloaded")
	}
}

// NewWorkerManager returns Manager instance.
func NewWorkerManager() *Manager {
	return &Manager{
		ch: make(chan interface{}, defaultConsumingLength),
	}
}
