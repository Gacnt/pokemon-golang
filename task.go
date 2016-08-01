package pgo

import (
	"log"
	"sync"
	"time"
)

// Job .
type Job interface {
	Run()
}

// Task .
type Task struct {
	MaxTasks int
	pending  []JobToRun
	tasks    []JobToRun
	add      chan JobToRun
	stop     chan interface{}

	Mu sync.Mutex
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
// taskType is optional, leaving the taskType blank will treat it as low priority
// taskType can be "URGENT" "HIGH" "MED" "LOW"
func (t *Task) AddFunc(id string, cmd func(), taskType ...string) {
	if len(taskType) == 0 {
		taskType = []string{""}
	}
	t.Mu.Lock()
	defer t.Mu.Unlock()
	if len(t.pending) <= t.MaxTasks {
		t.pending = append(t.pending, JobToRun{id, FuncJob(cmd)})
	} else if len(t.pending) >= t.MaxTasks && taskType[0] == "URGENT" {
		t.pending = append(t.pending, JobToRun{id, FuncJob(cmd)})
	}

	// Else throw away the task
}

// Start running tasks
func (t *Task) Start() {
	go t.run()
}

// Clear removes all tasks from the queue
func (t *Task) Clear() {
	t.tasks = []JobToRun{}
}

func (t *Task) run() {
	for stop := false; !stop; {
		select {
		case <-time.Tick(2 * time.Second):
			log.Println(t.tasks)
			for len(t.tasks) > 0 {
				t.tasks[0].Job.Run()
				t.tasks = t.tasks[1:]
				time.Sleep(3 * time.Second)
			}
			t.Mu.Lock()
			for _, pendingJob := range t.pending {
				found := false
				for _, taskedJobs := range t.tasks {
					log.Println(pendingJob.ID, taskedJobs.ID)
					if pendingJob.ID == taskedJobs.ID {
						found = true
						break
					}
				}
				if !found {
					t.tasks = append(t.tasks, pendingJob)
				}
			}
			t.pending = []JobToRun{}
			t.Mu.Unlock()
			time.Sleep(2 * time.Second)
		case <-t.stop:
			stop = true
		}
	}
}
