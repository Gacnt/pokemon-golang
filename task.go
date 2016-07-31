package pgo

import (
	"log"
	"time"
)

// Job .
type Job interface {
	Run()
}

// Task .
type Task struct {
	tasks []Job
	add   chan Job
	stop  chan interface{}
}

// FuncJob .
type FuncJob func()

// Run the function task
func (f FuncJob) Run() { f() }

// AddFunc Add the requested task
// These tasks will be processed in the order they are received
func (t *Task) AddFunc(cmd func()) {
	t.add <- FuncJob(cmd)
}

// Start running tasks
func (t *Task) Start() {
	go t.run()
}

// Clear removes all tasks from the queue
func (t *Task) Clear() {
	t.tasks = []Job{}
}

func (t *Task) run() {
	log.Println("Running")
	for stop := false; !stop; {
		select {
		case <-time.Tick(2 * time.Second):
			for len(t.tasks) > 0 {
				t.tasks[0].Run()
				t.tasks = t.tasks[1:]
			}
			log.Println("Ticking")
		case f := <-t.add:
			log.Println("Added Task")
			t.tasks = append(t.tasks, f)
		case <-t.stop:
			log.Println("Stop Task")
			stop = true
		}
	}
}
