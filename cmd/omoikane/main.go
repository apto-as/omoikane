package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/apto-as/omoikane/internal/subagent"
	"github.com/apto-as/omoikane/internal/security"
)

var (
	version = "0.1.0-trinity"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "omoikane",
	Short: "Omoikane - AI-powered development assistant with Trinity intelligence",
	Long: `Omoikane (思兼) - An AI-powered development assistant enhanced with
Trinity intelligence system (Springfield, Krukai, Vector).

This is a fork of Charmbracelet's Crush, enhanced with:
- High-performance Subagent management system
- Comprehensive Security Auditor
- Trinity coordination for multi-perspective analysis`,
	Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔮 Omoikane v" + version + " - Trinity Enhanced")
		fmt.Println("=====================================")
		fmt.Println()
		fmt.Println("Welcome to Omoikane, powered by Trinity intelligence system.")
		fmt.Println()
		fmt.Println("Available commands:")
		fmt.Println("  chat       - Start interactive chat session")
		fmt.Println("  test       - Run Trinity system tests")
		fmt.Println("  benchmark  - Run performance benchmarks")
		fmt.Println()
		fmt.Println("Use 'omoikane <command> --help' for more information.")
	},
}

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start interactive chat session",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔮 Starting Omoikane Chat with Trinity System...")
		fmt.Println()
		// TODO: Integrate with upstream crush chat functionality
		fmt.Println("Chat functionality will integrate with upstream Crush.")
		fmt.Println("For now, use the original crush command for chat.")
	},
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run Trinity system tests",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🧪 Running Trinity System Tests...")
		fmt.Println()
		
		// Test Subagent Manager
		fmt.Println("📦 Testing Subagent Manager...")
		config := subagent.DefaultManagerConfig()
		manager := subagent.NewManager(config)
		
		testAgent := subagent.SubagentDefinition{
			Name:          "test-agent",
			Description:   "Test agent",
			Version:       "1.0.0",
			Tools:         []string{"test"},
			Prompt:        "Test",
			TrinityRole:   subagent.TrinityRoleStrategist,
			SecurityLevel: subagent.SecurityLevelPublic,
		}
		
		if err := manager.Register(testAgent); err != nil {
			fmt.Printf("  ❌ Failed to register agent: %v\n", err)
		} else {
			fmt.Println("  ✅ Subagent registration successful")
		}
		
		// Test Security Auditor
		fmt.Println()
		fmt.Println("🛡️ Testing Security Auditor...")
		secConfig := security.DefaultSecurityConfig()
		auditor := security.NewSecurityAuditor(secConfig)
		
		testInputs := []string{
			"ls -la",           // Safe
			"rm -rf /",         // Dangerous
			"echo $(whoami)",   // Command injection
		}
		
		for _, input := range testInputs {
			assessment, _ := auditor.AnalyzeSecurity(nil, input, nil)
			if assessment.IsBlocked() {
				fmt.Printf("  🚫 BLOCKED: %s (Risk: %.2f)\n", input, assessment.RiskScore)
			} else {
				fmt.Printf("  ✅ ALLOWED: %s (Risk: %.2f)\n", input, assessment.RiskScore)
			}
		}
		
		fmt.Println()
		fmt.Println("✨ Trinity System Tests Complete!")
	},
}

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Run performance benchmarks",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📊 Running Performance Benchmarks...")
		fmt.Println()
		fmt.Println("Use 'go test ./internal/benchmark -bench=.' for detailed benchmarks")
		fmt.Println()
		
		// Quick performance test
		config := subagent.DefaultManagerConfig()
		manager := subagent.NewManager(config)
		
		// Register test agents
		for i := 0; i < 3; i++ {
			agent := subagent.SubagentDefinition{
				Name:          fmt.Sprintf("bench-agent-%d", i),
				Description:   "Benchmark agent",
				Version:       "1.0.0",
				Tools:         []string{"bench"},
				Prompt:        "Benchmark",
				TrinityRole:   subagent.TrinityRole(i % 3),
				SecurityLevel: subagent.SecurityLevelPublic,
			}
			manager.Register(agent)
		}
		
		metrics := manager.GetMetrics()
		fmt.Printf("📈 Manager Metrics:\n")
		fmt.Printf("  Total Executions: %d\n", metrics.TotalExecutions)
		fmt.Printf("  Active Executions: %d\n", metrics.ActiveExecutions)
		fmt.Printf("  Failed Executions: %d\n", metrics.FailedExecutions)
		
		fmt.Println()
		fmt.Println("✨ Quick benchmark complete!")
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(benchmarkCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}