# Worker

worker package is a manager, mixing all message distribute to workers which registeied in.  You have to implememts `Worker` interface then register, worker package will take over all workers.

```go
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
```

Simple implements:

```go
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
```



sample:

As above `MultiOutput` implements `Worker`,  we new two `MultiOutput` instance, register into worker manager.

Right now, there are two working modes.

- Co-Working

  multi-worker consuming same channel

```go
must := assert.New(t)
worker.Register(
  NewWorker("p1"),
  NewWorker("p2"),
)
go worker.RunOnCoWork()

for i := 0; i < 100; i++ {
  err := worker.Consume(i)
  must.Nil(err)
}
worker.Exit()
```



- Distributing

  read message from channel parallelly send to all workers which registeied

```go
must := assert.New(t)
worker.Register(
  NewWorker("p1"),
  NewWorker("p2"),
)
go worker.RunOnDistribute()

for i := 0; i < 100; i++ {
  err := worker.Consume(i)
  must.Nil(err)
}
worker.Exit()
```

