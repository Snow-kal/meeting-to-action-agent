package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/Snow-kal/meeting-to-action-agent/internal/api"
	"github.com/Snow-kal/meeting-to-action-agent/internal/pipeline"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP 服务监听地址")
	dryRun := flag.Bool("dry-run", true, "默认是否 dry-run")
	maxRetries := flag.Int("max-retries", 3, "默认重试次数")
	syncTarget := flag.String("sync-target", "none", "默认同步目标：none/jira/notion/both")
	syncTimeout := flag.Duration("sync-timeout", 30*time.Second, "默认同步超时")
	llmMode := flag.String("llm-mode", "off", "默认 LLM 模式：off/hybrid")
	mappingConfig := flag.String("mapping-config", "", "字段映射配置 JSON 路径")
	flag.Parse()

	s := api.NewServer(api.ServerOptions{
		DryRun:            *dryRun,
		MaxRetries:        *maxRetries,
		SyncTimeout:       *syncTimeout,
		SyncTarget:        pipeline.SyncTarget(*syncTarget),
		LLMMode:           pipeline.LLMMode(*llmMode),
		MappingConfigPath: *mappingConfig,
	})

	fmt.Printf("HTTP API listening on %s\n", *addr)
	if err := http.ListenAndServe(*addr, s.Handler()); err != nil {
		panic(err)
	}
}
