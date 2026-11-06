package cmd

import (
	"context"
	"fmt"
	"time"

	"autoops-hub/internal/config"
	"autoops-hub/internal/engine"

	"github.com/spf13/cobra"
)

var configPath string
var timeoutFlag time.Duration

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Thực thi các kịch bản tự động hóa từ file cấu hình YAML",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("📂 Đang đọc file cấu hình kịch bản từ: %s\n", configPath)

		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("❌ Lỗi đọc file cấu hình: %v\n", err)
			return
		}

		var tasks []engine.Task
		for _, task := range cfg.Workflows {
			tasks = append(tasks, engine.Task{
				ID:      task.ID,
				Name:    task.Name,
				Command: task.Command,
				Args:    task.Args,
			})
		}

		if len(tasks) == 0 {
			fmt.Println("⚠️ Không tìm thấy tác vụ nào trong file cấu hình!")
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeoutFlag)
		defer cancel()

		fmt.Printf("🚀 Khởi chạy Core Engine (Worker Pool) với %d tác vụ...\n", len(tasks))

		// Đã sửa 'task' thành 'tasks' (đúng tên biến chứa danh sách)
		results := engine.RunWorkerPool(ctx, tasks, 2)

		fmt.Println("\n📊 --- KẾT QUẢ THỰC THI WORKFLOWS ---")
		for _, res := range results {
			if res.Success {
				fmt.Printf("✅ [SUCCESS] Task ID: %s | Name: %s\nOutput:\n%s\n", res.TaskID, res.Name, res.Output)
			} else {
				fmt.Printf("❌ [FAILED] Task ID: %s | Name: %s | Error: %v\n", res.TaskID, res.Name, res.Error)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&configPath, "config", "c", "workflows.yaml", "Đường dẫn đến file cấu hình YAML")
	runCmd.Flags().DurationVarP(&timeoutFlag, "timeout", "t", 30*time.Second, "Thời gian tối đa (timeout) cho chuỗi tác vụ")
}
