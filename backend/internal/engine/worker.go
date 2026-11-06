package engine

import (
	"context"
	"fmt"
	"log/slog"
	"os"
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

// RunScanServers thực thi tiến trình quét hệ thống bằng Worker Pool
func RunScanServers(timeout time.Duration) {
	// Khởi tạo slog logger (có thể cấu hình JSONHandler cho production, ở đây dùng TextHandler cho terminal)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	start := time.Now()
	dbManager, err := storage.NewDBManager("autoops.db")
	if err != nil {
		logger.Error("Không thể kết nối SQLite", slog.Any("error", err))
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

	numWorkers := 2
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			workerTask(ctx, workerID, jobsCh, resultsCh, logger)
		}(i + 1)
	}

	for _, server := range servers {
		jobsCh <- server
	}
	close(jobsCh)

	// Lắng nghe và đợi các worker hoàn tất để đóng kênh kết quả
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Thu thập kết quả và lưu vào SQLite
	for result := range resultsCh {
		logger.Info("Nhận kết quả thực thi task",
			slog.String("server", result.Name),
			slog.Bool("success", result.Success),
			slog.Duration("elapsed", result.Elapsed),
		)

		if dbErr := dbManager.SaveExecutionResult(result.TaskID, result.Name, result.Output, result.Error, result.Elapsed); dbErr != nil {
			logger.Error("Lỗi khi lưu log vào DB",
				slog.String("server", result.Name),
				slog.Any("error", dbErr),
			)
		}
	}

	logger.Info("Đã hoàn tất toàn bộ tiến trình quét và lưu lịch sử vào SQLite thành công",
		slog.Duration("total_duration", time.Since(start)),
	)
}

func workerTask(ctx context.Context, id int, jobsCh <-chan ServerTarget, resultsCh chan<- TaskResult, logger *slog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		case server, ok := <-jobsCh:
			if !ok {
				return
			}

			logger.Debug("Worker bắt đầu xử lý task",
				slog.Int("worker_id", id),
				slog.String("server", server.Name),
			)

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
