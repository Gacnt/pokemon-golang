package pgo

import (
	"sort"
	"sync"
	"time"
)

const (
	Low    int = 1
	Med    int = 2
	High   int = 3
	Urgent int = 4
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
	ID       string
	Job      Job
	Priority int
}

// FuncJob .
type FuncJob func()

type Priority []JobToRun

func (p Priority) Len() int           { return len(p) }
func (p Priority) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Priority) Less(i, j int) bool { return p[i].Priority > p[j].Priority }

// Run the function task
func (f FuncJob) Run() { f() }

// AddFunc Add the requested task
// These tasks will be processed in the order they are received
// taskType is optional, leaving the taskType blank will treat it as low priority
// taskType can be "URGENT" "HIGH" "MED" "LOW"
func (t *Task) AddFunc(id string, cmd func(), taskType ...int) {
	if len(taskType) == 0 {
		taskType = []int{1} // Task is Low priority
	}
	t.Mu.Lock()
	defer t.Mu.Unlock()
	highestTaskPriority := 0
	for _, pendingTasks := range t.pending {
		if pendingTasks.Priority > highestTaskPriority {
			highestTaskPriority = pendingTasks.Priority
		}
	}

	if len(t.pending) <= t.MaxTasks {
		// Pending isn't full, add this task
		t.pending = append(t.pending, JobToRun{id, FuncJob(cmd), taskType[0]})
	} else if len(t.pending) >= t.MaxTasks && taskType[0] > highestTaskPriority {
		// This task is more important than one of the tasks in the pending,
		// remove the task and add this one instead
		lowestPriorityTask := 4
		lowestPriorityTaskI := 0
		for i, pendingTasks := range t.pending {
			if pendingTasks.Priority < lowestPriorityTask {
				lowestPriorityTask = pendingTasks.Priority
				lowestPriorityTaskI = i
			}
		}
		// Remove lower priority task
		t.pending = append(t.pending[:lowestPriorityTaskI], t.pending[lowestPriorityTaskI+1:]...)
		// Add new task
		t.pending = append(t.pending, JobToRun{id, FuncJob(cmd), taskType[0]})

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
			sort.Sort(Priority(t.tasks))
			for len(t.tasks) > 0 {
				t.tasks[0].Job.Run()
				t.tasks = t.tasks[1:]
				time.Sleep(3 * time.Second)
			}
			t.Mu.Lock()
			sort.Sort(Priority(t.pending))
			for _, pendingJob := range t.pending {
				found := false
				for _, taskedJobs := range t.tasks {
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
