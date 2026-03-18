package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/agents"
	"github.com/Snow-kal/meeting-to-action-agent/internal/input"
	"github.com/Snow-kal/meeting-to-action-agent/internal/pipeline"
	"github.com/Snow-kal/meeting-to-action-agent/internal/report"
	"github.com/Snow-kal/meeting-to-action-agent/internal/syncer"
)

func main() {
	inputPath := flag.String("input", "", "会议记录文件路径（必填）")
	meetingDateRaw := flag.String("meeting-date", "", "会议日期（格式：YYYY-MM-DD，默认今天）")
	syncTarget := flag.String("sync-target", "none", "同步目标：none/jira/notion/both")
	syncTimeout := flag.Duration("sync-timeout", 30*time.Second, "同步超时时间，例如 30s/1m")
	maxRetries := flag.Int("max-retries", 3, "同步失败重试次数（含首次请求）")
	dryRun := flag.Bool("dry-run", true, "是否 dry-run（true 时只模拟同步）")
	outputPath := flag.String("output", "result.json", "结果输出 JSON 路径")
	reportPath := flag.String("report", "", "可选：输出 Markdown 报告路径")
	flag.Parse()

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "缺少 -input 参数")
		os.Exit(1)
	}

	meetingInput, err := input.LoadMeetingInput(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取输入失败: %v\n", err)
		os.Exit(1)
	}

	meetingDate := time.Now()
	if meetingInput.MeetingDate != nil {
		meetingDate = *meetingInput.MeetingDate
	}
	if *meetingDateRaw != "" {
		parsed, parseErr := time.Parse("2006-01-02", *meetingDateRaw)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "meeting-date 解析失败: %v\n", parseErr)
			os.Exit(1)
		}
		meetingDate = parsed
	}

	jiraClient := syncer.NewJiraClientFromEnv(*dryRun)
	notionClient := syncer.NewNotionClientFromEnv(*dryRun)
	jiraClient.Retry.MaxAttempts = *maxRetries
	notionClient.Retry.MaxAttempts = *maxRetries

	orch := pipeline.NewOrchestrator(
		agents.NewRecorderAgent(),
		agents.NewDecisionAgent(),
		agents.NewTaskPlannerAgent(),
		agents.NewReviewerAgent(),
		jiraClient,
		notionClient,
	)

	result, err := orch.Run(context.Background(), meetingInput.Content, pipeline.Options{
		MeetingDate: meetingDate,
		SyncTarget:  pipeline.SyncTarget(*syncTarget),
		SyncTimeout: *syncTimeout,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "执行失败: %v\n", err)
		os.Exit(1)
	}

	if *outputPath != "" {
		body, marshalErr := json.MarshalIndent(result, "", "  ")
		if marshalErr != nil {
			fmt.Fprintf(os.Stderr, "结果序列化失败: %v\n", marshalErr)
			os.Exit(1)
		}
		if writeErr := os.WriteFile(*outputPath, body, 0o644); writeErr != nil {
			fmt.Fprintf(os.Stderr, "输出文件写入失败: %v\n", writeErr)
			os.Exit(1)
		}
	}

	if *reportPath != "" {
		md := report.BuildMarkdown(result)
		if writeErr := os.WriteFile(*reportPath, []byte(md), 0o644); writeErr != nil {
			fmt.Fprintf(os.Stderr, "报告写入失败: %v\n", writeErr)
			os.Exit(1)
		}
	}

	fmt.Printf("完成：决策 %d 条，任务 %d 条，检查项 %d 条，同步记录 %d 条\n",
		len(result.Decisions), len(result.Tasks), len(result.Issues), len(result.Synced))
}
