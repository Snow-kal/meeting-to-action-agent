# meeting-to-action-agent

Go 实现的会议到任务自动化流水线：
`会议记录 -> 决策提取 -> 任务拆解 -> 负责人识别 -> 截止时间补全 -> 同步 Jira/Notion`

## 角色型 Agent

- `Recorder Agent`：清洗会议文本并结构化
- `Decision Agent`：识别关键决策
- `Task Planner Agent`：拆解可执行任务
- `Reviewer Agent`：补齐负责人/截止时间/依赖关系

## 运行方式

```bash
go run ./cmd/meeting-to-action -input ./meeting.txt -meeting-date 2026-03-18 -sync-target both -dry-run=true -sync-timeout 30s -max-retries 3 -llm-mode off -mapping-config ./examples/mapping.sample.json -output result.json -report report.md
```

参数：

- `-input` 会议记录文件（必填）
- `-meeting-date` 会议日期（`YYYY-MM-DD`，默认今天）
- `-sync-target` `none/jira/notion/both`
- `-sync-timeout` 同步超时时间（默认 `30s`）
- `-max-retries` 同步失败重试次数（默认 `3`）
- `-llm-mode` `off/hybrid`（默认 `off`）
- `-mapping-config` Jira/Notion 字段映射配置 JSON 路径
- `-dry-run` 是否模拟同步（默认 `true`）
- `-output` 输出 JSON 文件路径（默认 `result.json`）
- `-report` 可选，输出 Markdown 报告路径

## HTTP API 服务

启动服务：

```bash
go run ./cmd/meeting-to-action-api -addr :8080 -dry-run=true -sync-target none -llm-mode off
```

健康检查：

```bash
curl http://localhost:8080/healthz
```

调用 `POST /run`：

```bash
curl -X POST http://localhost:8080/run \
  -H "Content-Type: application/json" \
  -d '{
    "content": "行动项：@张三 明天提交上线计划",
    "meeting_date": "2026-03-18",
    "sync_target": "none",
    "include_report": true
  }'
```

请求字段支持：`content`、`meeting_date`、`sync_target`、`dry_run`、`max_retries`、`sync_timeout`、`llm_mode`、`mapping_config_path`、`include_report`。

## 输入格式

支持 `txt/md/json`：

- `txt/md`：直接写会议记录文本
- `json`：可用以下字段

```json
{
  "content": "会议记录正文",
  "meeting_date": "2026-03-18"
}
```

兼容字段：`raw_text`、`text`、`date`。

## 同步稳定性

- Jira/Notion 同步支持对 `429`、`5xx`、网络异常自动重试
- `sync-target=both` 时 Jira 与 Notion 并行同步
- 支持 `-sync-timeout` 超时控制，避免卡死

## LLM 混合模式（规则 + 模型）

- 开启方式：`-llm-mode hybrid`
- 行为：先做规则抽取，再由 LLM 增强并去重融合（失败自动回退规则模式）

环境变量：

- `OPENAI_API_KEY`（必填，启用 hybrid 时）
- `OPENAI_BASE_URL`（可选，默认 `https://api.openai.com/v1/chat/completions`）
- `LLM_MODEL`（可选，默认 `gpt-4.1-mini`）

## 环境变量（真实同步）

Jira:

- `JIRA_API_BASE`
- `JIRA_PROJECT_KEY`
- `JIRA_EMAIL`
- `JIRA_TOKEN`

Notion:

- `NOTION_API_BASE`（可选，默认 `https://api.notion.com/v1/pages`）
- `NOTION_TOKEN`
- `NOTION_DATABASE_ID`

## 字段映射配置

通过 `-mapping-config` 指定 JSON 文件，实现 Jira/Notion 字段模板自定义。示例见：

- `examples/mapping.sample.json`

## 测试

```bash
go test ./...
```

仓库已配置 GitHub Actions，在 `push/PR` 时自动执行测试。
