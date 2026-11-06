package engine

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

// TestWorkerTask_Execution kiểm tra logic chạy lệnh của worker đơn lẻ
func TestWorkerTask_Execution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	jobsCh := make(chan ServerTarget, 1)
	resultsCh := make(chan TaskResult, 1)

	// Khởi tạo logger dùng riêng cho test
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// QUAN TRỌNG: Phải có từ khóa 'go' để chạy workerTask ngầm,
	// giúp nó liên tục lắng nghe dữ liệu từ jobsCh
	go workerTask(ctx, 1, jobsCh, resultsCh, logger)

	// Đẩy một task giả lập vào kênh job
	jobsCh <- ServerTarget{Name: "Test-Server", Host: "127.0.0.1"}

	// Đóng channel để worker biết là hết việc
	close(jobsCh)

	// Đọc kết quả từ channel với cơ chế select chống treo test
	select {
	case res := <-resultsCh:
		if res.Name != "Test-Server" {
			t.Errorf("Mong đợi tên server là Test-Server, nhận được: %s", res.Name)
		}
		if !res.Success {
			t.Errorf("Task chạy thất bại: %v", res.Error)
		}
	case <-ctx.Done():
		t.Fatal("Test bị timeout, worker không phản hồi kết quả về resultsCh")
	}
}
