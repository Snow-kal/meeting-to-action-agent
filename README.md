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
go run ./cmd/meeting-to-action -input ./meeting.txt -meeting-date 2026-03-18 -sync-target both -dry-run=true -output result.json
```

参数：

- `-input` 会议记录文件（必填）
- `-meeting-date` 会议日期（`YYYY-MM-DD`，默认今天）
- `-sync-target` `none/jira/notion/both`
- `-dry-run` 是否模拟同步（默认 `true`）
- `-output` 输出 JSON 文件路径（默认 `result.json`）

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
