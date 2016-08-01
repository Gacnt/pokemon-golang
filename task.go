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
	Running bool
	tasks   []JobToRun
	add     chan JobToRun
	stop    chan interface{}
}

type JobToRun struct {
	ID  string
	Job Job
}

// FuncJob .
type FuncJob func()

// Run the function task
func (f FuncJob) Run() { f() }

// AddFunc Add the requested task
// These tasks will be processed in the order they are received
func (t *Task) AddFunc(id string, cmd func()) {
	t.add <- JobToRun{id, FuncJob(cmd)}
}

// Start running tasks
func (t *Task) Start() {
	go t.run()
}

// Clear removes all tasks from the queue
func (t *Task) Clear() {
	t.tasks = []JobToRun{}
}

func (t *Task) monitor() {
	for t.Running {
		for len(t.tasks) > 0 {
			t.tasks[0].Job.Run()
			t.tasks = t.tasks[1:]
			time.Sleep(3 * time.Second)
		}
		time.Sleep(2 * time.Second)
		log.Println("Ticking")
	}
}

func (t *Task) run() {
	t.Running = true
	log.Println("Running")
	go t.monitor()
	for stop := false; !stop; {
		select {
		case f := <-t.add:
			found := false
			for _, job := range t.tasks {
				log.Println(job.ID, f.ID)
				if job.ID == f.ID {
					found = true
					break
				}
			}
			if !found {
				log.Println("Added Task")
				t.tasks = append(t.tasks, f)
			}
		case <-t.stop:
			t.Running = false
			stop = true
		}
	}
}
