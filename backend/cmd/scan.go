package cmd

import (
	"fmt"
	"time"

	"autoops-hub/internal/engine"

	"github.com/spf13/cobra"
)

var scanTimeout time.Duration

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Quét trạng thái server và lưu lịch sử vào SQLite",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Đang khởi động tiến trình quét hệ thống...")
		engine.RunScanServers(scanTimeout)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().DurationVarP(&scanTimeout, "timeout", "t", 5*time.Second, "Thời gian tối đa cho mỗi quét")
}