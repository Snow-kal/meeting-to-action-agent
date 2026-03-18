# Commit Log

## 1) feat: implement go multi-agent pipeline core

- 初始化 Go 工程与 CLI 入口（`cmd/meeting-to-action/main.go`）。
- 建立多 Agent 核心链路：`Recorder`、`Decision`、`Task Planner`、`Reviewer`。
- 实现截止时间解析（绝对日期、相对日期、周几、月底规则）。
- 实现 Jira/Notion 同步适配器，并支持 `dry-run`。
- 增加编排器将“会议记录 -> 决策 -> 任务 -> 复核 -> 同步”串联。
- 补充 `README.md` 使用说明，新增 `.gitignore`。

## 2) test: add unit and integration coverage for pipeline

- 新增时间解析测试：`internal/timeutil/due_parser_test.go`。
- 新增任务拆解测试：`internal/agents/task_planner_test.go`。
- 新增 Reviewer 补全与校验测试：`internal/agents/reviewer_test.go`。
- 新增端到端编排测试：`internal/pipeline/orchestrator_test.go`。
- 新增 Jira/Notion 客户端测试（含 dry-run 与 mock server）：`internal/syncer/clients_test.go`。

## 3) docs: track task completion and add commit log

- 在 `AGENT.MD` 追加并勾选子任务清单，记录测试验证结果。
- 新增 `log.md`，汇总每个 commit 的变更内容，便于追踪迭代过程。
