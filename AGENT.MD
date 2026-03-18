目标 会议记录 -> 决策提取 -> 任务拆解 -> 责任人识别 -> 截止时间补全 -> 同步到 Jira/Notion

多 Agent 协作架构
可以设计成几个角色型 Agent：

Recorder Agent：整理会议内容
Decision Agent：识别关键决策
Task Planner Agent：拆解成可执行任务
Reviewer Agent：检查任务是否缺负责人、截止时间、依赖关系
这比单一 prompt 更像真正的 agent system，简历里会更有技术含量。

---

## 执行子任务清单（Go 实现）

- [x] T1 初始化 Go 项目结构（`go.mod`、目录约定、CLI 入口）
- [x] T2 实现 `Recorder Agent`：会议文本清洗与结构化整理
- [x] T3 实现 `Decision Agent`：关键决策识别与抽取
- [x] T4 实现 `Task Planner Agent`：把决策拆成可执行任务
- [x] T5 实现负责人识别（Owner 提取规则）
- [x] T6 实现截止时间补全（绝对/相对时间解析 + 默认策略）
- [x] T7 实现 `Reviewer Agent`：补齐负责人、截止时间、依赖关系并校验
- [x] T8 实现 Jira/Notion 同步适配层（支持 `dry-run`）
- [x] T9 实现端到端编排器（多 Agent 串联）
- [x] T10 补充测试（单元测试 + 集成测试）
- [x] T11 测试通过后提交并推送到 GitHub
- [x] T12 编写 `log.md`，汇总每个 commit 的改动

## 测试验证

- [x] `go test ./...` 通过（使用仓库内 `GOCACHE/.gocache` 与 `GOTMPDIR/.gotmp` 执行）
