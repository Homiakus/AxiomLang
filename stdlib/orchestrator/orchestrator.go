package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"AxiomLang/stdlib/actor"
)

// TaskState represents execution state of an orchestrated unit
type TaskState string

const (
	TaskPending   TaskState = "PENDING"
	TaskRunning   TaskState = "RUNNING"
	TaskCompleted TaskState = "COMPLETED"
	TaskFailed    TaskState = "FAILED"
	TaskCanceled  TaskState = "CANCELED"
)

// Task represents a discrete computation or workflow step
type Task struct {
	ID        string
	Name      string
	ActorType string
	Payload   interface{}
	State     TaskState
	Result    interface{}
	Error     error
	StartedAt time.Time
	EndedAt   time.Time
}

// Orchestrator coordinates distributed actor workflows, state consensus, and task graphs
type Orchestrator struct {
	system    *actor.System
	tasks     map[string]*Task
	mu        sync.RWMutex
	completed int64
	failed    int64
	ctx       context.Context
	cancel    context.CancelFunc
}

// New creates an active Axiom Orchestrator instance
func New() *Orchestrator {
	ctx, cancel := context.WithCancel(context.Background())
	return &Orchestrator{
		system: actor.NewSystem(),
		tasks:  make(map[string]*Task),
		ctx:    ctx,
		cancel: cancel,
	}
}

// RegisterTask adds a new task to orchestrator pool
func (o *Orchestrator) RegisterTask(id, name, actorType string, payload interface{}) *Task {
	o.mu.Lock()
	defer o.mu.Unlock()

	task := &Task{
		ID:        id,
		Name:      name,
		ActorType: actorType,
		Payload:   payload,
		State:     TaskPending,
	}
	o.tasks[id] = task
	return task
}

// ExecuteRun runs a task synchronously through an isolated managed actor
func (o *Orchestrator) ExecuteRun(taskID string, fn func(payload interface{}) (interface{}, error)) (*Task, error) {
	o.mu.Lock()
	task, exists := o.tasks[taskID]
	if !exists {
		o.mu.Unlock()
		return nil, fmt.Errorf("E_ORCHESTRATOR_TASK_NOT_FOUND: %s", taskID)
	}
	task.State = TaskRunning
	task.StartedAt = time.Now()
	o.mu.Unlock()

	resultChan := make(chan struct{})
	workerID := fmt.Sprintf("worker_%s_%d", taskID, time.Now().UnixNano())

	_, err := o.system.Spawn(workerID, 10, func(ctx context.Context, msg actor.Message) error {
		defer close(resultChan)
		res, execErr := fn(task.Payload)
		o.mu.Lock()
		task.EndedAt = time.Now()
		if execErr != nil {
			task.State = TaskFailed
			task.Error = execErr
			atomic.AddInt64(&o.failed, 1)
		} else {
			task.State = TaskCompleted
			task.Result = res
			atomic.AddInt64(&o.completed, 1)
		}
		o.mu.Unlock()
		return execErr
	})

	if err != nil {
		o.mu.Lock()
		task.State = TaskFailed
		task.Error = err
		o.mu.Unlock()
		return task, err
	}

	// Send execution trigger
	if err := o.system.Send(workerID, actor.Message{Type: "EXECUTE"}); err != nil {
		return task, err
	}

	// Wait for execution completion
	select {
	case <-resultChan:
		return task, task.Error
	case <-time.After(10 * time.Second):
		o.mu.Lock()
		task.State = TaskCanceled
		task.Error = errors.New("E_ORCHESTRATOR_TIMEOUT: execution exceeded deadline")
		o.mu.Unlock()
		return task, task.Error
	}
}

// GetStatus returns snapshot of orchestrator metrics
func (o *Orchestrator) GetStatus() (total, completed, failed int) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return len(o.tasks), int(atomic.LoadInt64(&o.completed)), int(atomic.LoadInt64(&o.failed))
}

// Shutdown terminates orchestrator
func (o *Orchestrator) Shutdown() {
	o.cancel()
	o.system.Stop()
}
