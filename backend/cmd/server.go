package cmd

import (
	"autoops-hub/internal/web"
	"fmt"

	"github.com/spf13/cobra"
)

var serverConfigPath string

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Khởi chạy Local Web Dashboard để điều khiển hệ thống",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🌐 Đang khởi động Web Dashboard tại http://localhost:8080 (Config: %s)...\n", serverConfigPath)

		// Khởi chạy Web Server và truyền đường dẫn config vào
		web.StartServer(":9980", serverConfigPath)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&serverConfigPath, "config", "c", "workflows.yaml", "Đường dẫn file cấu hình workflows cho Dashboard")
}
