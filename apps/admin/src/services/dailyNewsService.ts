import { apiRequest } from "./apiClient";

export interface DailyNewsItem {
  id: string;
  title: string;
  summary: string;
  category: string;
  publishedAt: string;
  discoveredAt: string;
  source: { name: string };
  links: { aihot: string; original: string };
}

export interface DailyNewsTask {
  date: string;
  status: "pending" | "approved" | "rejected" | "published";
  send_status: "pending" | "sending" | "failed" | "sent";
  items: DailyNewsItem[];
  article_ids: number[];
  message: string;
  copy_text?: string;
  message_id: string;
  last_error: string;
  runner_report?: DailyNewsRunnerReport | null;
}

export interface DailyNewsRunnerReport {
  date: string;
  collect: string;
  top_up: string;
  approve: string;
  publish: string;
  generate_covers: string;
  upload_assets: string;
  bind_covers: string;
  build_digest: string;
  send_feishu: string;
  failures?: string[];
  warnings?: string[];
  complete: boolean;
  recorded_at?: string;
}

const basePath = "/api/v1/daily-news";

export function getBeijingToday(): string {
  const parts = new Intl.DateTimeFormat("en-US", {
    timeZone: "Asia/Shanghai", year: "numeric", month: "2-digit", day: "2-digit"
  }).formatToParts(new Date());
  const value = (type: string) => parts.find((part) => part.type === type)?.value;
  return `${value("year")}-${value("month")}-${value("day")}`;
}

function normalizeTask(task: DailyNewsTask): DailyNewsTask {
  return {
    ...task,
    items: task.items ?? [],
    article_ids: task.article_ids ?? [],
    message: task.message ?? "",
    copy_text: task.copy_text ?? "",
    message_id: task.message_id ?? "",
    last_error: task.last_error ?? "",
    runner_report: task.runner_report
      ? {
          ...task.runner_report,
          failures: task.runner_report.failures ?? [],
          warnings: task.runner_report.warnings ?? []
        }
      : null
  };
}

export async function listDailyNews(): Promise<DailyNewsTask[]> {
  const tasks = await apiRequest<DailyNewsTask[] | null>(basePath);
  return (tasks ?? []).map(normalizeTask);
}

export async function collectDailyNews(): Promise<DailyNewsTask> {
  const task = await apiRequest<DailyNewsTask>(`${basePath}/collect`, {
    method: "POST",
    body: JSON.stringify({ date: getBeijingToday() })
  });
  return normalizeTask(task);
}

async function runTaskAction(date: string, action: "approve" | "publish" | "send"): Promise<DailyNewsTask> {
  const task = await apiRequest<DailyNewsTask>(`${basePath}/${encodeURIComponent(date)}/${action}`, { method: "POST" });
  return normalizeTask(task);
}

export function approveDailyNews(date: string): Promise<DailyNewsTask> {
  return runTaskAction(date, "approve");
}

export function publishDailyNews(date: string): Promise<DailyNewsTask> {
  return runTaskAction(date, "publish");
}

export function sendDailyNews(date: string): Promise<DailyNewsTask> {
  return runTaskAction(date, "send");
}
