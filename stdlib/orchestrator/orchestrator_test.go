package orchestrator_test

import (
	"errors"
	"testing"

	"AxiomLang/stdlib/orchestrator"
)

func TestOrchestrator_ExecuteRun(t *testing.T) {
	orch := orchestrator.New()
	defer orch.Shutdown()

	task := orch.RegisterTask("task_math_1", "DoubleValue", "math_worker", 21)

	resTask, err := orch.ExecuteRun(task.ID, func(payload interface{}) (interface{}, error) {
		val := payload.(int)
		return val * 2, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resTask.State != orchestrator.TaskCompleted {
		t.Fatalf("expected state COMPLETED, got %v", resTask.State)
	}
	if resTask.Result.(int) != 42 {
		t.Fatalf("expected result 42, got %v", resTask.Result)
	}

	total, completed, failed := orch.GetStatus()
	if total != 1 || completed != 1 || failed != 0 {
		t.Fatalf("unexpected status: total=%d, comp=%d, fail=%d", total, completed, failed)
	}
}

func TestOrchestrator_ExecuteFailure(t *testing.T) {
	orch := orchestrator.New()
	defer orch.Shutdown()

	task := orch.RegisterTask("task_fail_1", "FailingTask", "worker", nil)
	_, err := orch.ExecuteRun(task.ID, func(payload interface{}) (interface{}, error) {
		return nil, errors.New("simulated error")
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if task.State != orchestrator.TaskFailed {
		t.Fatalf("expected state FAILED, got %v", task.State)
	}
}
