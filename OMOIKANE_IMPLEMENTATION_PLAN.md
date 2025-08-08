# Omoikane Subagents & Hooks実装計画書
*Claude Code風のSubagents・Hooks機能をOmoikane（Goアプリケーション）に実装する*

## 📋 Executive Summary

Omoikane（Goで書かれたAI開発支援ツール）に、Claude Codeが持つSubagentsとHooksのような機能を実装します。これにより、複数の専門エージェントの並列実行と、ライフサイクルベースの自動化を実現します。

## 🏗️ 現状分析

### Omoikaneの既存構造
```go
// internal/llm/agent/agent.go
type Service interface {
    Run(ctx context.Context, sessionID string, content string, attachments ...message.Attachment) (<-chan AgentEvent, error)
    // ...
}

// internal/config/config.go  
type Agent struct {
    Prompt      string
    Preferences AgentPreferences
}
```

現在のOmoikaneは単一エージェントのアーキテクチャで、複数エージェントの管理や並列実行機能は持っていません。

## 🎯 実装目標

1. **Subagents機能**: 複数の専門エージェントを定義・管理・並列実行
2. **Hooks機能**: ツール実行前後のライフサイクルイベントでの自動処理
3. **Trinity Integration**: Springfield、Krukai、Vectorの三位一体システム統合

## 📐 アーキテクチャ設計

### 1. Subagents システム

#### 新規パッケージ構造
```
internal/
├── subagent/           # NEW: Subagents管理
│   ├── manager.go      # Subagent Manager
│   ├── registry.go     # Subagent Registry
│   ├── executor.go     # Parallel Executor
│   └── loader.go       # Configuration Loader
├── hooks/              # NEW: Hooks管理
│   ├── manager.go      # Hook Manager
│   ├── registry.go     # Hook Registry
│   └── executor.go     # Hook Executor
└── trinity/            # NEW: Trinity専用
    ├── springfield.go
    ├── krukai.go
    ├── vector.go
    └── coordinator.go
```

### 2. Subagents実装詳細

#### Subagent定義（`internal/subagent/types.go`）
```go
package subagent

import (
    "context"
    "github.com/charmbracelet/crush/internal/config"
)

// SubagentDefinition defines a specialized agent
type SubagentDefinition struct {
    Name        string            `yaml:"name" json:"name"`
    Description string            `yaml:"description" json:"description"`
    Tools       []string          `yaml:"tools" json:"tools"`
    Model       string            `yaml:"model,omitempty" json:"model,omitempty"`
    Color       string            `yaml:"color,omitempty" json:"color,omitempty"`
    Prompt      string            `yaml:"prompt" json:"prompt"`
    Metadata    map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// SubagentManager manages all subagents
type Manager interface {
    // Register a new subagent
    Register(def SubagentDefinition) error
    
    // Get a specific subagent
    Get(name string) (*SubagentDefinition, error)
    
    // List all available subagents
    List() []SubagentDefinition
    
    // Execute a subagent
    Execute(ctx context.Context, name string, input string) (string, error)
    
    // Execute multiple subagents in parallel
    ExecuteParallel(ctx context.Context, agents []string, input string) (map[string]string, error)
}
```

#### Subagent Manager実装（`internal/subagent/manager.go`）
```go
package subagent

import (
    "context"
    "fmt"
    "sync"
    "golang.org/x/sync/errgroup"
)

type manager struct {
    registry map[string]SubagentDefinition
    mu       sync.RWMutex
    executor *parallelExecutor
}

func NewManager() Manager {
    return &manager{
        registry: make(map[string]SubagentDefinition),
        executor: newParallelExecutor(),
    }
}

func (m *manager) ExecuteParallel(ctx context.Context, agents []string, input string) (map[string]string, error) {
    results := make(map[string]string)
    var mu sync.Mutex
    
    g, ctx := errgroup.WithContext(ctx)
    
    // Limit concurrent executions to 10 (like Claude Code)
    semaphore := make(chan struct{}, 10)
    
    for _, agentName := range agents {
        agentName := agentName // capture loop variable
        
        g.Go(func() error {
            semaphore <- struct{}{}        // acquire
            defer func() { <-semaphore }() // release
            
            result, err := m.Execute(ctx, agentName, input)
            if err != nil {
                return fmt.Errorf("%s: %w", agentName, err)
            }
            
            mu.Lock()
            results[agentName] = result
            mu.Unlock()
            
            return nil
        })
    }
    
    if err := g.Wait(); err != nil {
        return nil, err
    }
    
    return results, nil
}
```

### 3. Hooks システム

#### Hook定義（`internal/hooks/types.go`）
```go
package hooks

import (
    "context"
    "encoding/json"
)

// HookEvent represents different lifecycle events
type HookEvent string

const (
    HookEventPreToolUse   HookEvent = "PreToolUse"
    HookEventPostToolUse  HookEvent = "PostToolUse"
    HookEventSessionStart HookEvent = "SessionStart"
    HookEventSessionEnd   HookEvent = "SessionEnd"
    HookEventPreCompact   HookEvent = "PreCompact"
    HookEventStop         HookEvent = "Stop"
)

// HookDefinition defines a hook
type HookDefinition struct {
    Event    HookEvent         `json:"event"`
    Matcher  string            `json:"matcher"`  // Tool name pattern
    Command  string            `json:"command"`
    Env      map[string]string `json:"env,omitempty"`
}

// HookResult represents the result of hook execution
type HookResult struct {
    Continue    bool   `json:"continue"`
    StopReason  string `json:"stopReason,omitempty"`
    SystemMessage string `json:"systemMessage,omitempty"`
}

// HookManager manages all hooks
type Manager interface {
    Register(hook HookDefinition) error
    Execute(ctx context.Context, event HookEvent, data interface{}) (*HookResult, error)
    LoadFromConfig(path string) error
}
```

#### Hook Manager実装（`internal/hooks/manager.go`）
```go
package hooks

import (
    "context"
    "encoding/json"
    "os/exec"
    "regexp"
)

type manager struct {
    hooks map[HookEvent][]HookDefinition
}

func (m *manager) Execute(ctx context.Context, event HookEvent, data interface{}) (*HookResult, error) {
    hooks, exists := m.hooks[event]
    if !exists {
        return &HookResult{Continue: true}, nil
    }
    
    for _, hook := range hooks {
        // Check if hook matches the context
        if !m.matches(hook, data) {
            continue
        }
        
        // Execute hook command
        result, err := m.executeHook(ctx, hook, data)
        if err != nil {
            return nil, err
        }
        
        // If hook says stop, stop processing
        if !result.Continue {
            return result, nil
        }
    }
    
    return &HookResult{Continue: true}, nil
}

func (m *manager) executeHook(ctx context.Context, hook HookDefinition, data interface{}) (*HookResult, error) {
    // Prepare JSON input
    jsonData, err := json.Marshal(data)
    if err != nil {
        return nil, err
    }
    
    // Execute command
    cmd := exec.CommandContext(ctx, "sh", "-c", hook.Command)
    cmd.Stdin = bytes.NewReader(jsonData)
    
    // Set environment variables
    for k, v := range hook.Env {
        cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
    }
    
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    // Parse result
    var result HookResult
    if err := json.Unmarshal(output, &result); err != nil {
        // If output is not JSON, assume success
        return &HookResult{Continue: true}, nil
    }
    
    return &result, nil
}
```

### 4. Trinity Integration

#### Trinity Coordinator（`internal/trinity/coordinator.go`）
```go
package trinity

import (
    "context"
    "github.com/charmbracelet/crush/internal/subagent"
)

type Coordinator struct {
    manager subagent.Manager
}

func NewCoordinator(manager subagent.Manager) *Coordinator {
    return &Coordinator{manager: manager}
}

// AnalyzeWithTrinity performs parallel analysis using all Trinity members
func (c *Coordinator) AnalyzeWithTrinity(ctx context.Context, input string) (*TrinityResult, error) {
    agents := []string{
        "springfield-strategist",
        "krukai-optimizer", 
        "vector-auditor",
    }
    
    // Execute in parallel
    results, err := c.manager.ExecuteParallel(ctx, agents, input)
    if err != nil {
        return nil, err
    }
    
    // Synthesize results
    return c.synthesize(results), nil
}

type TrinityResult struct {
    Strategic   string // Springfield's analysis
    Technical   string // Krukai's analysis
    Security    string // Vector's analysis
    Consensus   string // Synthesized recommendation
}
```

### 5. 設定ファイル

#### Subagent設定（`.omoikane/agents/springfield.yaml`）
```yaml
name: springfield-strategist
description: Strategic planning and architecture design
tools:
  - read
  - write
  - bash
model: claude-3-5-sonnet-20241022
prompt: |
  You are Springfield, the strategic architect.
  Focus on long-term vision and architecture.
  Begin responses with "ふふ、素晴らしい質問ですね。"
```

#### Hooks設定（`.omoikane/hooks.json`）
```json
{
  "hooks": [
    {
      "event": "PreToolUse",
      "matcher": "bash",
      "command": ".omoikane/hooks/safety-check.sh",
      "env": {
        "VECTOR_MODE": "paranoid"
      }
    },
    {
      "event": "PostToolUse",
      "matcher": "write|edit",
      "command": ".omoikane/hooks/format-code.sh"
    }
  ]
}
```

## 🚀 実装フェーズ

### Phase 1: 基盤構築（3日）
- [ ] Subagentパッケージの基本構造作成
- [ ] SubagentManagerの実装
- [ ] YAMLローダーの実装
- [ ] 単一Subagent実行のテスト

### Phase 2: 並列実行（2日）
- [ ] ParallelExecutorの実装
- [ ] セマフォによる並行数制限（最大10）
- [ ] 結果集約メカニズム
- [ ] エラーハンドリング

### Phase 3: Hooks実装（2日）
- [ ] Hooksパッケージの基本構造
- [ ] HookManagerの実装
- [ ] ライフサイクルイベントの統合
- [ ] コマンド実行とJSON通信

### Phase 4: Trinity統合（2日）
- [ ] Trinity専用エージェントの実装
- [ ] Coordinatorの実装
- [ ] 結果合成アルゴリズム
- [ ] CLIコマンドの追加

### Phase 5: テストと最適化（1日）
- [ ] 単体テスト
- [ ] 統合テスト
- [ ] パフォーマンステスト
- [ ] ドキュメント作成

## 🔧 技術的考慮事項

### 並行処理
- `golang.org/x/sync/errgroup`を使用した安全な並行処理
- セマフォによる同時実行数制限
- コンテキストによるキャンセレーション

### 設定管理
- YAMLとJSONの両方をサポート
- ホットリロード対応を検討
- 設定の優先順位（プロジェクト > ユーザー > デフォルト）

### セキュリティ
- Hookコマンドのサンドボックス化
- 入力のサニタイゼーション
- 権限管理

## 📊 成功指標

1. **機能完成度**
   - [ ] 3つ以上のSubagentが定義・実行可能
   - [ ] 並列実行で2倍以上の高速化
   - [ ] 5種類以上のHookイベントをサポート

2. **パフォーマンス**
   - [ ] 10エージェント同時実行が可能
   - [ ] Hook実行のオーバーヘッド < 100ms
   - [ ] メモリ使用量の増加 < 50MB

3. **互換性**
   - [ ] 既存のcrushコマンドとの後方互換性
   - [ ] 既存の設定ファイルの継続利用

## 🎯 最終目標

Omoikaneが、Claude Codeのような高度なエージェント管理と自動化機能を持つ、真のAI開発支援プラットフォームとなること。

---

*Implementation Plan by Trinity Intelligence System*
*Real Go implementation, not just configuration*