package engine

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"autoops-hub/internal/storage"
)

type ServerTarget struct {
	Name string
	Host string
}

func RunScanServers(timeout time.Duration) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	start := time.Now()
	dbManager, err := storage.NewDBManager("autoops.db")
	if err != nil {
		fmt.Printf("❌ Không thể kết nối SQLite: %v\n", err)
		return
	}
	defer dbManager.Conn.Close()

	servers := []ServerTarget{
		{Name: "Web-Prod-01", Host: "127.0.0.1"},
		{Name: "DB-Primary", Host: "127.0.0.1"},
		{Name: "Cache-Redis", Host: "127.0.0.1"},
		{Name: "Worker-Node-01", Host: "127.0.0.1"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	jobsCh := make(chan ServerTarget, len(servers))
	resultsCh := make(chan TaskResult, len(servers))
	var wg sync.WaitGroup

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			workerTask(ctx, workerID, jobsCh, resultsCh)
		}(i + 1)
	}

	for _, server := range servers {
		jobsCh <- server
	}
	close(jobsCh)

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	for result := range resultsCh {
		fmt.Printf("[Kết quả] Server %s -> Thành công: %v (Thời gian: %v)\n", result.Name, result.Success, result.Elapsed)
		if err := dbManager.SaveExecutionResult(result.TaskID, result.Name, result.Output, result.Error, result.Elapsed); err != nil {
			fmt.Printf("⚠️ Lỗi khi lưu log vào DB cho server %s: %v\n", result.Name, err)
		}
	}

	fmt.Println("✅ Đã hoàn tất toàn bộ tiến trình và lưu lịch sử vào SQLite thành công!")
	fmt.Println("Total time running: ", time.Since(start))
}

func workerTask(ctx context.Context, id int, jobsCh <-chan ServerTarget, resultsCh chan<- TaskResult) {
	for {
		select {
		case <-ctx.Done():
			return
		case server, ok := <-jobsCh:
			if !ok {
				return
			}

			taskStart := time.Now()
			cmdCtx, cmdCancel := context.WithTimeout(ctx, 2*time.Second)
			var execCmd *exec.Cmd

			msg := fmt.Sprintf("Checking connection to %s (%s)", server.Name, server.Host)
			if runtime.GOOS == "windows" {
				execCmd = exec.CommandContext(cmdCtx, "cmd", "/C", "echo", msg)
			} else {
				execCmd = exec.CommandContext(cmdCtx, "sh", "-c", fmt.Sprintf("echo \"%s\"", msg))
			}

			outputBytes, err := execCmd.CombinedOutput()
			cmdCancel()

			elapsed := time.Since(taskStart)
			result := TaskResult{
				TaskID:  server.Name,
				Name:    server.Name,
				Output:  string(outputBytes),
				Elapsed: elapsed,
				Success: err == nil,
			}
			if err != nil {
				result.Error = err.Error()
			}
			resultsCh <- result
		}
	}
}
