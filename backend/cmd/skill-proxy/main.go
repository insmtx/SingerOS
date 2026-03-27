// main 包是 SingerOS 技能代理服务的主入口
//
// 技能代理服务负责技能的隔离执行和管理，提供安全的技能运行环境。
package main

import (
	"github.com/spf13/cobra"
	"github.com/ygpkg/yg-go/logs"
)

var rootCmd = &cobra.Command{
	Use:   "skill-proxy",
	Short: "Backend service for the Skill Proxy",
	Long:  `This is the backend service for the Skill Proxy, responsible for handling API requests and business logic.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Start the Skill Proxy service
		logs.Info("Starting Skill Proxy service...")
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logs.Error(err)
	}
}
