# meeting-to-action-agent

把会议记录转成可执行结果的多智能体系统：
`会议记录 -> 决策提取 -> 任务拆解 -> 负责人识别 -> 截止时间补全 -> Jira / Notion 同步`

## 项目更改

- 增加多智能体流水线：`Recorder`、`Decision`、`Task Planner`、`Owner`、`Deadline`、`Reviewer`
- 支持规则 + LLM 混合抽取，保留来源依据、验收标准、风险标记、追问问题
- 提供 Web 页面，可先预览，再接受同步，或提交反馈后重新生成
- 支持同步到 Jira / Notion，包含字段映射、失败重试、超时控制
- 输出更完整的结构化结果：`summary`、`decisions`、`tasks`、`conflicts`、`follow_up_questions`

## 用法

### 1. 启动 Web 服务

```bash
go run ./cmd/meeting-to-action-api -addr :8080 -dry-run=true -sync-target none -llm-mode off
```

打开：

```text
http://localhost:8080/
```

适合直接粘贴会议记录，在页面里查看概览、接受同步或反馈修改。

### 2. 命令行运行

```bash
go run ./cmd/meeting-to-action -input ./examples/meeting.validation.abcd.md -meeting-date 2026-03-18 -sync-target none -dry-run=true -output result.json
```

常用参数：

- `-input`：会议记录文件
- `-meeting-date`：会议日期
- `-sync-target`：`none / jira / notion / both`
- `-dry-run`：是否只演练不真实同步
- `-llm-mode`：`off / hybrid`
- `-output`：结果 JSON 输出路径

### 3. 真实同步前配置

启用 LLM：

- `OPENAI_API_KEY`

同步 Jira：

- `JIRA_API_BASE`
- `JIRA_PROJECT_KEY`
- `JIRA_EMAIL`
- `JIRA_TOKEN`

同步 Notion：

- `NOTION_TOKEN`
- `NOTION_DATABASE_ID`

### 4. 测试

```bash
go test ./...
```

## 架构图

![项目架构图](./docs/architecture-overview.svg)
