const state = {
  baseContent: "",
  latestContent: "",
  latestResult: null,
};

const meetingContentEl = document.getElementById("meeting-content");
const meetingDateEl = document.getElementById("meeting-date");
const llmModeEl = document.getElementById("llm-mode");
const syncTargetEl = document.getElementById("sync-target");
const dryRunEl = document.getElementById("dry-run");
const toolLLMKeyEl = document.getElementById("tool-llm-key");
const toolNotionKeyEl = document.getElementById("tool-notion-key");
const toolJiraKeyEl = document.getElementById("tool-jira-key");
const toolRememberEl = document.getElementById("tool-remember");
const toolClearBtn = document.getElementById("tool-clear-btn");

const previewBtn = document.getElementById("preview-btn");
const acceptBtn = document.getElementById("accept-btn");
const rejectBtn = document.getElementById("reject-btn");
const applyFeedbackBtn = document.getElementById("apply-feedback-btn");

const resultPanel = document.getElementById("result-panel");
const feedbackPanel = document.getElementById("feedback-panel");
const statusText = document.getElementById("status-text");
const stats = document.getElementById("stats");

const decisionsList = document.getElementById("decisions-list");
const tasksList = document.getElementById("tasks-list");
const issuesList = document.getElementById("issues-list");
const feedbackText = document.getElementById("feedback-text");

const today = new Date().toISOString().slice(0, 10);
meetingDateEl.value = today;
loadToolConfig();

[toolLLMKeyEl, toolNotionKeyEl, toolJiraKeyEl, toolRememberEl].forEach((el) => {
  el.addEventListener("change", persistToolConfig);
  el.addEventListener("input", persistToolConfig);
});

toolClearBtn.addEventListener("click", () => {
  toolLLMKeyEl.value = "";
  toolNotionKeyEl.value = "";
  toolJiraKeyEl.value = "";
  localStorage.removeItem("mta_tool_config");
  setStatus("工具配置已清空。");
});

previewBtn.addEventListener("click", async () => {
  const content = meetingContentEl.value.trim();
  if (!content) {
    setStatus("请先输入会议记录。");
    meetingContentEl.focus();
    return;
  }
  state.baseContent = content;
  state.latestContent = content;
  await previewWithContent(content);
});

acceptBtn.addEventListener("click", async () => {
  if (!state.latestContent) {
    setStatus("请先生成概览。");
    return;
  }
  setLoading(acceptBtn, true);
  setStatus("正在同步到目标系统...");
  try {
    const data = await callRun({
      content: state.latestContent,
      meeting_date: meetingDateEl.value,
      sync_target: syncTargetEl.value,
      dry_run: dryRunEl.checked,
      llm_mode: llmModeEl.value,
      include_report: false,
    });
    renderResult(data.result);
    const syncCount = Array.isArray(data.result?.synced) ? data.result.synced.length : 0;
    setStatus(`已完成同步流程。同步记录 ${syncCount} 条。`);
  } catch (err) {
    setStatus(`同步失败：${err.message}`);
  } finally {
    setLoading(acceptBtn, false);
  }
});

rejectBtn.addEventListener("click", () => {
  if (!state.latestResult) {
    setStatus("请先生成概览。");
    return;
  }
  feedbackPanel.classList.remove("hidden");
  feedbackText.focus();
  setStatus("请填写你的意见，我们将按反馈重新生成。");
});

applyFeedbackBtn.addEventListener("click", async () => {
  const opinion = feedbackText.value.trim();
  if (!opinion) {
    setStatus("请先填写反馈意见。");
    return;
  }
  const revised = [
    state.baseContent,
    "",
    "【用户反馈】",
    opinion,
    "请根据以上反馈重做决策提取和任务拆解，重点处理负责人、截止时间和依赖关系。",
  ].join("\n");

  state.latestContent = revised;
  await previewWithContent(revised);
  setStatus("已根据反馈重生成概览，请再次确认。");
});

async function previewWithContent(content) {
  setLoading(previewBtn, true);
  setStatus("正在生成概览...");
  try {
    const data = await callRun({
      content,
      meeting_date: meetingDateEl.value,
      sync_target: "none",
      dry_run: true,
      llm_mode: llmModeEl.value,
      include_report: true,
    });
    renderResult(data.result);
    resultPanel.classList.remove("hidden");
    feedbackPanel.classList.add("hidden");
    feedbackText.value = "";
    setStatus("概览已生成，请选择接受或反馈修改。");
  } catch (err) {
    setStatus(`生成失败：${err.message}`);
  } finally {
    setLoading(previewBtn, false);
  }
}

async function callRun(payload) {
  const toolConfig = getToolConfigPayload();
  const reqPayload = { ...payload, ...toolConfig };
  const resp = await fetch("/run", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(reqPayload),
  });
  if (!resp.ok) {
    const text = await resp.text();
    throw new Error(text || `HTTP ${resp.status}`);
  }
  return resp.json();
}

function renderResult(result) {
  state.latestResult = result;

  const dCount = safeLen(result?.decisions);
  const tCount = safeLen(result?.tasks);
  const iCount = safeLen(result?.issues);
  const wCount = safeLen(result?.warnings);

  stats.innerHTML = `
    <div class="stat"><div class="v">${dCount}</div><div class="k">决策</div></div>
    <div class="stat"><div class="v">${tCount}</div><div class="k">任务</div></div>
    <div class="stat"><div class="v">${iCount}</div><div class="k">检查项</div></div>
    <div class="stat"><div class="v">${wCount}</div><div class="k">告警</div></div>
  `;

  fillList(decisionsList, result?.decisions, (d) => `${d.id || ""} ${d.text || ""}`.trim());
  fillList(tasksList, result?.tasks, (t) => {
    const due = t?.due_date ? formatDate(t.due_date) : "未设置";
    return `${t.id || ""} ${t.title || ""} / owner: ${t.owner || "待指派"} / due: ${due}`;
  });
  fillList(issuesList, result?.issues, (i) => `${i.task_id || ""} [${i.type || ""}] ${i.message || ""}`);
}

function fillList(el, arr, format) {
  el.innerHTML = "";
  if (!Array.isArray(arr) || arr.length === 0) {
    const li = document.createElement("li");
    li.textContent = "无";
    el.appendChild(li);
    return;
  }
  arr.forEach((item) => {
    const li = document.createElement("li");
    li.className = "fade-in";
    li.textContent = format(item);
    el.appendChild(li);
  });
}

function safeLen(v) {
  return Array.isArray(v) ? v.length : 0;
}

function formatDate(dateString) {
  const d = new Date(dateString);
  if (Number.isNaN(d.getTime())) return dateString;
  return d.toISOString().slice(0, 10);
}

function setStatus(text) {
  statusText.textContent = text;
}

function setLoading(button, loading) {
  button.disabled = loading;
  if (loading) {
    button.dataset.origin = button.textContent;
    button.textContent = "处理中...";
    return;
  }
  if (button.dataset.origin) {
    button.textContent = button.dataset.origin;
  }
}

function getToolConfigPayload() {
  const payload = {};
  const llmKey = toolLLMKeyEl.value.trim();
  const notionKey = toolNotionKeyEl.value.trim();
  const jiraKey = toolJiraKeyEl.value.trim();
  if (llmKey) payload.llm_api_key = llmKey;
  if (notionKey) payload.notion_database_id = notionKey;
  if (jiraKey) payload.jira_project_key = jiraKey;
  return payload;
}

function persistToolConfig() {
  if (!toolRememberEl.checked) {
    localStorage.removeItem("mta_tool_config");
    return;
  }
  const data = {
    llm_api_key: toolLLMKeyEl.value,
    notion_database_id: toolNotionKeyEl.value,
    jira_project_key: toolJiraKeyEl.value,
    remember: toolRememberEl.checked,
  };
  localStorage.setItem("mta_tool_config", JSON.stringify(data));
}

function loadToolConfig() {
  try {
    const raw = localStorage.getItem("mta_tool_config");
    if (!raw) return;
    const data = JSON.parse(raw);
    toolLLMKeyEl.value = data.llm_api_key || "";
    toolNotionKeyEl.value = data.notion_database_id || "";
    toolJiraKeyEl.value = data.jira_project_key || "";
    toolRememberEl.checked = data.remember !== false;
  } catch (err) {
    localStorage.removeItem("mta_tool_config");
  }
}
