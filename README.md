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
go run ./cmd/meeting-to-action -input ./meeting.txt -meeting-date 2026-03-18 -sync-target both -dry-run=true -sync-timeout 30s -max-retries 3 -output result.json -report report.md
```

参数：

- `-input` 会议记录文件（必填）
- `-meeting-date` 会议日期（`YYYY-MM-DD`，默认今天）
- `-sync-target` `none/jira/notion/both`
- `-sync-timeout` 同步超时时间（默认 `30s`）
- `-max-retries` 同步失败重试次数（默认 `3`）
- `-dry-run` 是否模拟同步（默认 `true`）
- `-output` 输出 JSON 文件路径（默认 `result.json`）
- `-report` 可选，输出 Markdown 报告路径

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

## 测试

```bash
go test ./...
```

仓库已配置 GitHub Actions，在 `push/PR` 时自动执行测试。
