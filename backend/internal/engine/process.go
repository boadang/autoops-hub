package engine

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID      int
	Name    string
	Command string
	Args    []string
}

type TaskResult struct {
	TaskID  string
	Name    string
	Output  string
	Error   string
	Elapsed time.Duration
	Success bool
}

// IsProcessRunning kiểm tra xem một ứng dụng/tiến trình có đang chạy hay không.
func IsProcessRunning(processName string) (bool, error) {
	cmd := exec.Command("tasklist")
	if runtime.GOOS != "windows" {
		cmd = exec.Command("ps", "-eo", "comm=")
	}

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	output := strings.ToLower(string(outputBytes))
	return strings.Contains(output, strings.ToLower(processName)), nil
}

// KillProcess ép đóng một tiến trình theo tên.
func KillProcess(processName string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("taskkill", "/F", "/IM", processName)
	} else {
		cmd = exec.Command("pkill", "-f", processName)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("không thể tắt tiến trình %s: %v", processName, err)
	}
	return nil
}

func RunWorkerPool(ctx context.Context, tasks []Task, workerCount int) []TaskResult {
	if workerCount <= 0 {
		workerCount = 1
	}
	if len(tasks) == 0 {
		return nil
	}

	jobsCh := make(chan Task)
	resultsCh := make(chan TaskResult, len(tasks))
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobsCh {
				resultsCh <- executeTask(ctx, task)
			}
		}()
	}

	for _, task := range tasks {
		jobsCh <- task
	}
	close(jobsCh)
	wg.Wait()
	close(resultsCh)

	results := make([]TaskResult, 0, len(tasks))
	for result := range resultsCh {
		results = append(results, result)
	}
	return results
}

func executeTask(ctx context.Context, task Task) TaskResult {
	start := time.Now()
	cmd := exec.CommandContext(ctx, task.Command, task.Args...)

	output, err := cmd.CombinedOutput()
	elapsed := time.Since(start)

	result := TaskResult{
		TaskID:  strconv.Itoa(task.ID),
		Name:    task.Name,
		Output:  string(output),
		Elapsed: elapsed,
		Success: err == nil,
	}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}

