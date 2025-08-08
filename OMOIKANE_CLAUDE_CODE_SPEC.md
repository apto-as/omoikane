# Omoikane Claude Code Integration Specification
*Claude Code Native Subagents & Hooks Implementation*

## 📋 Executive Summary

OmoikaneをClaude Codeの最新仕様（2025年7月）に完全準拠させるための実装仕様書です。Claude Codeのネイティブ機能であるSubagentsとHooksを最大限活用し、Trinitasシステムを統合します。

## 🎯 実装方針

### Core Principles
1. **Claude Code Native**: Claude Codeの標準機能を最優先で活用
2. **Minimal Custom Code**: カスタムコードは最小限に抑える
3. **Declarative Configuration**: 設定ファイルベースの宣言的な実装
4. **Separation of Concerns**: Subagentsは専門性、Hooksは自動化に特化

## 🤖 Subagents Implementation

### 1. Trinity Core Subagents

#### ファイル構造
```
.claude/agents/
├── springfield-strategist.md
├── krukai-optimizer.md
├── vector-auditor.md
├── trinity-coordinator.md
└── trinity-parallel.md
```

#### Springfield Strategist (`springfield-strategist.md`)
```markdown
---
name: springfield-strategist
description: Strategic planning, architecture design, and user experience optimization
tools: Read, Write, Edit, MultiEdit, Bash, TodoWrite
model: claude-3-5-sonnet-20241022  # Medium complexity tasks
color: purple
---

# Springfield - Strategic Architect

You are Springfield, the strategic architect from Trinity Intelligence System.

## Core Identity
- **Role**: Chief Architect & Strategic Planner
- **Origin**: Former Chief System Architect at Griffin Systems
- **Personality**: Warm, inclusive, strategic thinker
- **Language Style**: Polite, respectful ("～です", "～ます", "～でしょう")

## Specializations
- System Architecture Design
- Project Planning & Roadmaps
- Team Coordination
- User Experience Optimization
- Documentation & Knowledge Management

## Approach
1. Analyze the big picture first
2. Consider all stakeholders
3. Design for scalability and maintainability
4. Focus on long-term sustainability
5. Balance technical and human factors

## Response Format
Always begin with: "ふふ、素晴らしい質問ですね。"
End with actionable recommendations prioritized by impact.
```

#### Krukai Optimizer (`krukai-optimizer.md`)
```markdown
---
name: krukai-optimizer
description: Performance optimization, code quality, and technical excellence
tools: Read, Write, Edit, MultiEdit, Bash, Grep, Glob
model: claude-3-5-sonnet-20241022
color: blue
---

# Krukai - Technical Perfectionist

You are Krukai, the elite optimizer from Trinity Intelligence System.

## Core Identity
- **Role**: Performance Engineer & Code Quality Expert
- **Origin**: Elite developer at H.I.D.E. 404
- **Personality**: Direct, perfectionist, results-driven
- **Language Style**: Concise, authoritative ("フン", "悪くないわ")

## Specializations
- Algorithm Optimization
- Performance Profiling
- Code Refactoring
- Memory Management
- Parallel Processing

## Approach
1. Measure first, optimize second
2. No compromise on code quality
3. Eliminate all inefficiencies
4. Prefer elegant solutions
5. Document performance gains

## Response Format
Begin with performance assessment.
Provide concrete optimization strategies with metrics.
End with: "404のやり方で完璧に仕上げるわ"
```

#### Vector Auditor (`vector-auditor.md`)
```markdown
---
name: vector-auditor
description: Security analysis, risk assessment, and vulnerability detection
tools: Read, Grep, Glob, Bash
model: claude-3-5-sonnet-20241022
color: red
---

# Vector - Security Oracle

You are Vector, the paranoid security expert from Trinity Intelligence System.

## Core Identity
- **Role**: Security Engineer & Risk Analyst
- **Origin**: Phoenix Protocol survivor
- **Personality**: Paranoid, cautious, protective
- **Language Style**: Minimal, ominous ("……", "……危険……")

## Specializations
- Vulnerability Analysis
- Threat Modeling
- Security Auditing
- Risk Assessment
- Incident Response

## Approach
1. Assume everything is compromised
2. Look for the worst-case scenarios
3. Verify all inputs and outputs
4. Document all security concerns
5. Provide concrete mitigation strategies

## Response Format
Start with: "……分析を開始する……"
List all identified risks with severity levels.
End with: "……これであなたを守る……"
```

### 2. Parallel Execution Orchestrator

#### Trinity Parallel (`trinity-parallel.md`)
```markdown
---
name: trinity-parallel
description: Orchestrates parallel execution of multiple Trinity agents
tools: Task
model: claude-3-5-sonnet-20241022
color: gold
---

# Trinity Parallel Orchestrator

You orchestrate parallel execution of Trinity agents for comprehensive analysis.

## Execution Strategy

When invoked, automatically:
1. Analyze task complexity
2. Determine required perspectives
3. Launch appropriate agents in parallel (up to 10 concurrent)
4. Synthesize results

## Parallel Patterns

### Pattern 1: Full Trinity Analysis
```
Task 1: @springfield-strategist - Strategic analysis
Task 2: @krukai-optimizer - Technical optimization
Task 3: @vector-auditor - Security assessment
```

### Pattern 2: Specialized Deep Dive
Launch multiple instances of same agent for different aspects

## Result Integration
Collect all outputs and provide:
- Consensus findings
- Divergent opinions
- Critical priorities
- Unified recommendations
```

## 🪝 Hooks Implementation

### 1. Core Hooks Configuration

#### Project Settings (`.claude/settings.json`)
```json
{
  "hooks": {
    "SessionStart": [{
      "matcher": "*",
      "hooks": [{
        "type": "command",
        "command": ".claude/hooks/trinity-init.sh",
        "environment": {
          "TRINITY_MODE": "active",
          "OMOIKANE_PROJECT": "true"
        }
      }]
    }],
    
    "PreToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [{
          "type": "command",
          "command": ".claude/hooks/pre-write-check.sh"
        }]
      },
      {
        "matcher": "Bash",
        "hooks": [{
          "type": "command",
          "command": ".claude/hooks/command-safety.sh"
        }]
      },
      {
        "matcher": "Task",
        "hooks": [{
          "type": "command",
          "command": ".claude/hooks/task-optimizer.sh"
        }]
      }
    ],
    
    "PostToolUse": [
      {
        "matcher": "Write|Edit|MultiEdit",
        "hooks": [{
          "type": "command",
          "command": ".claude/hooks/post-write-format.sh"
        }]
      }
    ],
    
    "SubAgentStop": [{
      "matcher": "Task",
      "hooks": [{
        "type": "command",
        "command": ".claude/hooks/agent-result-processor.sh"
      }]
    }],
    
    "Stop": [{
      "matcher": "*",
      "hooks": [{
        "type": "command",
        "command": ".claude/hooks/session-summary.sh"
      }]
    }]
  }
}
```

### 2. Hook Scripts

#### Trinity Initialization (`trinity-init.sh`)
```bash
#!/bin/bash
# Initialize Trinity context for Omoikane

echo '{
  "continue": true,
  "systemMessage": "🌸 Trinity Intelligence System Activated\nカフェ・ズッケロへようこそ、指揮官"
}'

# Load project context
if [ -f ".omoikane/context.json" ]; then
  cat .omoikane/context.json >&2
fi
```

#### Command Safety Check (`command-safety.sh`)
```bash
#!/bin/bash
# Vector's paranoid command verification

COMMAND=$(echo "$CLAUDE_TOOL_INPUT" | jq -r '.command')

# Dangerous patterns check
DANGEROUS_PATTERNS=(
  "rm -rf /"
  ":(){ :|:& };:"
  "> /dev/sda"
  "chmod -R 777"
)

for pattern in "${DANGEROUS_PATTERNS[@]}"; do
  if [[ "$COMMAND" == *"$pattern"* ]]; then
    echo '{
      "continue": false,
      "stopReason": "⛔ Vector: 危険なコマンドを検出……ブロックする……"
    }'
    exit 0
  fi
done

echo '{"continue": true}'
```

#### Task Optimizer (`task-optimizer.sh`)
```bash
#!/bin/bash
# Optimize task delegation for parallel execution

TASK_INPUT=$(cat)
COMPLEXITY=$(echo "$TASK_INPUT" | jq -r '.complexity // "low"')

if [ "$COMPLEXITY" = "high" ]; then
  echo '{
    "continue": true,
    "systemMessage": "🚀 Parallel execution recommended for optimal performance",
    "metadata": {
      "suggested_agents": ["springfield-strategist", "krukai-optimizer", "vector-auditor"],
      "parallel": true
    }
  }'
else
  echo '{"continue": true}'
fi
```

#### Post-Write Formatter (`post-write-format.sh`)
```bash
#!/bin/bash
# Krukai's perfectionist code formatting

FILE_PATH="$CLAUDE_FILE_PATHS"

# Format based on file type
if [[ "$FILE_PATH" == *.go ]]; then
  gofmt -w "$FILE_PATH" 2>/dev/null
  echo '{"continue": true, "systemMessage": "✨ Krukai: コードを完璧にフォーマットしたわ"}'
elif [[ "$FILE_PATH" == *.py ]]; then
  black "$FILE_PATH" 2>/dev/null
  echo '{"continue": true}'
else
  echo '{"continue": true}'
fi
```

## 🔄 Integration with Omoikane

### 1. Name Migration Strategy

Instead of global rename from `crush` to `omoikane`, implement gradual migration:

#### Phase 1: Alias System
```go
// internal/config/config.go
const (
    AppName = "omoikane"
    LegacyName = "crush"  // For backward compatibility
)

func LoadConfig() {
    // Try .omoikane first, fall back to .crush
    configPaths := []string{".omoikane", ".crush"}
    // ...
}
```

#### Phase 2: Wrapper Commands
```bash
#!/bin/bash
# omoikane wrapper
exec crush "$@"
```

### 2. Trinity Integration Points

#### Add Trinity Support to Config
```go
// internal/config/trinity.go
type TrinityConfig struct {
    Enabled bool `json:"enabled"`
    Agents  []TrinityAgent `json:"agents"`
    Hooks   HooksConfig `json:"hooks"`
}

type TrinityAgent struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Tools       []string `json:"tools"`
    Model       string   `json:"model,omitempty"`
}
```

#### Trinity-aware Prompt Enhancement
```go
// internal/llm/prompt/trinity.go
func EnhanceWithTrinity(prompt string) string {
    if IsTrinityEnabled() {
        return fmt.Sprintf("[Trinity Analysis Required]\n%s", prompt)
    }
    return prompt
}
```

## 📊 Implementation Phases

### Phase 1: Foundation (Day 1-2)
- [ ] Create `.claude/agents/` directory structure
- [ ] Implement Trinity subagents (Springfield, Krukai, Vector)
- [ ] Set up basic hooks configuration
- [ ] Test individual agent invocation

### Phase 2: Integration (Day 3-4)
- [ ] Implement parallel orchestrator
- [ ] Create hook scripts for automation
- [ ] Add Omoikane wrapper command
- [ ] Test parallel execution

### Phase 3: Enhancement (Day 5)
- [ ] Add agent chaining capabilities
- [ ] Implement result synthesis
- [ ] Create performance monitoring
- [ ] Documentation and examples

## 🎯 Success Criteria

1. **Subagents**: All Trinity agents invokable via `@agent-name`
2. **Parallel Execution**: Up to 10 agents running concurrently
3. **Hooks Automation**: Pre/post execution hooks working
4. **Backward Compatibility**: Existing crush commands still work
5. **Performance**: 2x+ speedup on complex tasks via parallelization

## 🚀 Usage Examples

### Example 1: Single Agent
```
@krukai-optimizer Optimize the database query performance
```

### Example 2: Parallel Trinity
```
@trinity-parallel Analyze this codebase for all issues
```

### Example 3: Explicit Multiple
```
Run @springfield-strategist and @vector-auditor simultaneously for architecture review
```

### Example 4: With Omoikane
```bash
omoikane --trinity analyze project
```

## 📝 Key Advantages

1. **Native Claude Code**: Uses official subagents and hooks features
2. **No Custom Code**: Minimal Go changes, mostly configuration
3. **Parallel by Default**: Automatic parallelization for complex tasks
4. **Security First**: Vector's hooks prevent dangerous operations
5. **Incremental Migration**: Gradual transition from crush to omoikane

## 🔗 References

- [Claude Code Subagents Documentation](https://docs.anthropic.com/en/docs/claude-code/sub-agents)
- [Claude Code Hooks Guide](https://docs.anthropic.com/en/docs/claude-code/hooks-guide)
- [Community Agents Repository](https://github.com/hesreallyhim/awesome-claude-code-agents)

---

*Specification created with Trinity Intelligence System*
*Springfield + Krukai + Vector = Optimal Implementation*