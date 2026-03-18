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

## 4) feat: add structured input, markdown report and resilient sync

- 新增多格式输入解析（`txt/md/json`），并支持从 JSON 读取 `meeting_date/date`。
- CLI 支持 `-sync-timeout`、`-max-retries`、`-report` 参数。
- 新增 Markdown 报告生成器，输出决策/任务/检查项/同步结果汇总。
- Jira/Notion 同步新增重试机制（429/5xx/网络异常，指数退避）。
- 编排器支持同步超时控制，且 `sync-target=both` 时并行同步 Jira/Notion。

## 5) test: cover input/report/retry and add sync timeout tests

- 新增输入解析测试：`internal/input/loader_test.go`。
- 新增报告生成测试：`internal/report/markdown_test.go`。
- 新增重试逻辑测试：`internal/syncer/retry_test.go`。
- 补充编排器同步超时测试：`internal/pipeline/orchestrator_test.go`。
- 新增 GitHub Actions CI：`.github/workflows/ci.yml`，在 push/PR 自动跑 `go test ./...`。

## 6) docs: update phase2 checklist and usage docs

- `AGENT.MD` 新增并完成二期子任务清单（P1-P7）及测试验证记录。
- `README.md` 更新二期能力说明（输入格式、并行同步、重试、超时、报告导出、CI）。

## 7) feat: add HTTP API, LLM hybrid mode and mapping config

- 新增 HTTP API 服务程序：`cmd/meeting-to-action-api`，支持 `POST /run` 与 `GET /healthz`。
- 新增 API 处理层：支持前端/Webhook 请求参数（`sync_target/dry_run/max_retries/sync_timeout/llm_mode` 等）。
- 引入 LLM 混合抽取模式（规则 + 模型）：新增 OpenAI 客户端与抽取融合逻辑，失败自动回退规则模式并记录告警。
- 编排器新增 `LLMMode`，支持规则抽取与模型抽取去重合并。
- 新增 Jira/Notion 字段映射配置加载能力，可通过 JSON 模板自定义字段。
- 新增运行时工厂统一构建编排器，CLI 与 API 复用同一初始化逻辑。
- 新增映射样例文件：`examples/mapping.sample.json`。

## 8) test: cover API endpoint, hybrid LLM merge and mapping config

- 新增 `POST /run` API 测试与 Bad Request 测试。
- 新增字段映射配置加载测试（默认配置与文件配置）。
- 新增 LLM 客户端解析测试（mock chat completion）。
- 新增运行时工厂测试（hybrid 模式缺少 API Key 的错误路径）。
- 补充编排器 LLM 混合模式测试、同步客户端映射字段断言测试。

## 9) docs: update phase3 checklist and API/LLM docs

- `AGENT.MD` 新增并勾选三期任务（R1-R5）与验证记录。
- `README.md` 增加 HTTP API 用法、LLM 混合模式配置、字段映射配置说明。
- `log.md` 增补第 7-9 次提交改动摘要。

## 10) feat: add web console for preview, accept sync and feedback revise

- 新增内嵌静态前端页面（`index.html/styles.css/app.js`），默认由 API 服务根路径 `/` 提供。
- 页面支持会议记录输入、概览展示、状态反馈、移动端适配与简洁视觉风格。
- 打通“先预览后确认”流程：预览调用 `/run`（`sync_target=none`），接受后按选择同步 Jira/Notion。
- 打通“不接受->意见反馈->自动改写重生成”流程。
- API 服务新增静态资源路由，支持根路径返回前端页面。

## 11) test: add web page availability coverage

- 新增 API 层测试：`GET /` 返回前端页面并包含标题文本。
- 回归验证 API 与编排链路测试均通过。

## 12) docs: update phase4 checklist and frontend usage

- `AGENT.MD` 新增并勾选四期任务（F1-F6）与测试验证结果。
- `README.md` 增加前端页面启动/访问与交互流程说明。
- `log.md` 增补第 10-12 次提交改动摘要。
