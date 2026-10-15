<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { useAuthStore } from "../../stores/auth";
import {
  approveDailyNews, collectDailyNews, getBeijingToday, listDailyNews, publishDailyNews, sendDailyNews,
  type DailyNewsTask
} from "../../services/dailyNewsService";

const router = useRouter();
const authStore = useAuthStore();
const collectDate = ref(getBeijingToday());
const isAdmin = computed(() => authStore.role === "admin");
const records = ref<DailyNewsTask[]>([]);
const loading = ref(false);
const busy = ref("");
const loadError = ref("");
const actionError = ref("");
const selectedDate = ref("");
const detailVisible = ref(false);
const selectedTask = computed(() => records.value.find((task) => task.date === selectedDate.value));
const locked = computed(() => loading.value || !!busy.value);
const statusLabels: Record<string, string> = {
  pending: "待审核", approved: "已审核", rejected: "已拒绝", published: "已发布"
};
const sendStatusLabels: Record<string, string> = {
  pending: "待发送", sending: "发送中", failed: "发送失败", sent: "已发送"
};

function statusLabel(status: string) {
  return statusLabels[status] ?? status;
}

function statusTagType(status: string) {
  if (["failed", "rejected"].includes(status)) return "danger";
  if (["approved", "published", "sent"].includes(status)) return "success";
  return "warning";
}

function errorText(error: unknown, fallback: string) {
  if (error instanceof SyntaxError) return "接口返回格式异常，请确认每日资讯服务已启用后重试";
  if (error instanceof TypeError) return "网络请求失败，请检查连接后重试";
  return error instanceof Error ? error.message : fallback;
}

async function loadRecords() {
  collectDate.value = getBeijingToday();
  loading.value = true;
  loadError.value = "";
  try {
    records.value = (await listDailyNews()).sort((a, b) => b.date.localeCompare(a.date));
  } catch (error) {
    loadError.value = errorText(error, "加载每日资讯失败");
  } finally {
    loading.value = false;
  }
}

function showDetails(task: DailyNewsTask) {
  selectedDate.value = task.date;
  detailVisible.value = true;
}

function canApprove(task: DailyNewsTask) {
  return task.status === "pending";
}

function canPublish(task: DailyNewsTask) {
  return task.status === "approved";
}

function canSend(task: DailyNewsTask) {
  return isAdmin.value && task.status === "published" && ["pending", "failed"].includes(task.send_status);
}

function safeLink(url?: string) {
  if (!url) return undefined;
  try {
    const parsed = new URL(url);
    return ["https:", "http:"].includes(parsed.protocol) ? parsed.href : undefined;
  } catch {
    return undefined;
  }
}

async function handleCollect() {
  if (locked.value) return;
  collectDate.value = getBeijingToday();
  const existed = records.value.some((task) => task.date === collectDate.value);
  busy.value = "collect";
  actionError.value = "";
  try {
    const task = await collectDailyNews();
    records.value = [task, ...records.value.filter((record) => record.date !== task.date)]
      .sort((a, b) => b.date.localeCompare(a.date));
    showDetails(task);
    ElMessage.success(existed ? "已返回已有任务，未更新资讯或草稿" : "已获取今日任务；已有任务不会重复更新");
    await loadRecords();
  } catch (error) {
    actionError.value = errorText(error, "采集失败");
    ElMessage.error(actionError.value);
    await loadRecords();
  } finally {
    busy.value = "";
  }
}

type TaskAction = "approve" | "publish" | "send";
const actions = {
  approve: { label: "审核通过", description: "请确认资讯内容、来源和文章草稿已检查完毕。无资讯的任务也可审核通过。", run: approveDailyNews },
  publish: { label: "发布资讯", description: "将发布该任务并冻结日报文本。有关联草稿时请确认编辑完成；0 篇文章的任务也可发布。", run: publishDailyNews },
  send: { label: "发送日报", description: "将向配置的消息渠道发送日报。请确认文本和发布结果；失败重试前请检查渠道是否已收到，避免重复发送。", run: sendDailyNews }
};

async function handleAction(task: DailyNewsTask, action: TaskAction) {
  const allowed = { approve: canApprove, publish: canPublish, send: canSend };
  if (locked.value || loadError.value || !allowed[action](task)) return;
  const config = actions[action];
  try {
    await ElMessageBox.confirm(`${task.date}：${config.description}`, config.label, {
      type: "warning", confirmButtonText: "确认", cancelButtonText: "取消"
    });
  } catch {
    return;
  }
  const currentTask = records.value.find((record) => record.date === task.date);
  if (locked.value || loadError.value || !currentTask || !allowed[action](currentTask)) return;
  busy.value = `${task.date}:${action}`;
  actionError.value = "";
  try {
    await config.run(task.date);
    ElMessage.success(`${config.label}成功`);
    await loadRecords();
  } catch (error) {
    actionError.value = errorText(error, `${config.label}失败`);
    ElMessage.error(actionError.value);
    // A failed request may still have updated the task's status or last_error.
    await loadRecords();
  } finally {
    busy.value = "";
  }
}

async function copyMessage(task: DailyNewsTask) {
  if (!task.message.trim()) {
    ElMessage.warning("暂无可复制文本");
    return;
  }
  try {
    if (!navigator.clipboard?.writeText) throw new Error("当前环境不支持剪贴板，请在详情中手动选择并复制文本");
    await navigator.clipboard.writeText(task.message);
    ElMessage.success("日报文本已复制");
  } catch (error) {
    ElMessage.error(errorText(error, "复制失败，请手动选择文本复制"));
    showDetails(task);
  }
}

onMounted(() => void loadRecords());
</script>

<template>
  <div class="view">
    <div class="toolbar">
      <div>
        <h2>每日资讯</h2>
        <p>仅采集北京时间今天的滚动 24 小时资讯；重复采集返回已有任务，不更新资讯或草稿。</p>
      </div>
      <div class="toolbar-actions">
        <span>北京时间：{{ collectDate }}</span>
        <el-button type="primary" :loading="busy === 'collect'" :disabled="locked" @click="handleCollect">采集今日资讯</el-button>
        <el-button :loading="loading" :disabled="locked" @click="loadRecords">刷新</el-button>
      </div>
    </div>

    <el-alert v-if="loadError" :title="loadError" description="列表可能不是最新状态，请刷新重试。" type="error" show-icon :closable="false" />
    <el-alert v-if="actionError" :title="actionError" type="error" show-icon @close="actionError = ''" />

    <el-card shadow="never">
      <el-table :data="records" v-loading="loading" row-key="date" empty-text="暂无每日资讯任务，请采集今日资讯">
        <el-table-column prop="date" label="日期" width="130" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="发送状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.send_status)">{{ sendStatusLabels[row.send_status] || row.send_status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="资讯 / 文章" width="120">
          <template #default="{ row }">{{ row.items.length }} / {{ row.article_ids.length }}</template>
        </el-table-column>
        <el-table-column prop="message_id" label="消息 ID" min-width="160" show-overflow-tooltip />
        <el-table-column label="最近错误" min-width="220" show-overflow-tooltip>
          <template #default="{ row }"><span class="error-text">{{ row.last_error || "-" }}</span></template>
        </el-table-column>
        <el-table-column label="操作" width="360" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="showDetails(row)">详情 / 草稿</el-button>
            <el-button link :disabled="!row.message.trim()" @click="copyMessage(row)">复制文本</el-button>
            <el-button link type="primary" :disabled="locked || !!loadError || !canApprove(row)" :loading="busy === `${row.date}:approve`" @click="handleAction(row, 'approve')">审核</el-button>
            <el-button link type="success" :disabled="locked || !!loadError || !canPublish(row)" :loading="busy === `${row.date}:publish`" @click="handleAction(row, 'publish')">发布</el-button>
            <el-button v-if="isAdmin" link type="primary" :disabled="locked || !!loadError || !canSend(row)" :loading="busy === `${row.date}:send`" @click="handleAction(row, 'send')">{{ row.send_status === "failed" ? "重试发送" : "发送" }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-drawer v-model="detailVisible" title="每日资讯详情" size="min(780px, 100vw)">
      <div v-if="selectedTask" class="detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="日期">{{ selectedTask.date }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ statusLabel(selectedTask.status) }}</el-descriptions-item>
          <el-descriptions-item label="发送状态" :span="2">{{ sendStatusLabels[selectedTask.send_status] || selectedTask.send_status }}</el-descriptions-item>
          <el-descriptions-item label="消息 ID" :span="2">{{ selectedTask.message_id || "尚未发送" }}</el-descriptions-item>
        </el-descriptions>
        <el-alert v-if="selectedTask.last_error" :title="selectedTask.last_error" type="error" :closable="false" show-icon />
        <section>
          <h3>资讯内容</h3>
          <el-empty v-if="!selectedTask.items.length" description="暂无资讯" />
          <article v-for="(item, index) in selectedTask.items" :key="item.id || index" class="news-item">
            <h4>{{ index + 1 }}. {{ item.title || "未命名资讯" }}</h4>
            <p class="preserve-text">{{ item.summary || "暂无摘要" }}</p>
            <p>来源：{{ item.source?.name || "未知来源" }}</p>
            <el-link v-if="safeLink(item.links?.aihot)" :href="safeLink(item.links?.aihot)" target="_blank" rel="noopener noreferrer" type="primary">查看 AI HOT 原文</el-link>
            <span v-else>暂无 AI HOT 链接</span>
          </article>
        </section>
        <section>
          <h3>文章草稿</h3>
          <p>打开文章编辑页检查内容，保存后返回刷新任务，再确认审核与发布。</p>
          <div class="draft-links">
            <el-button v-for="id in selectedTask.article_ids" :key="id" plain type="primary" @click="router.push({ name: 'article-edit', params: { id } })">编辑文章 #{{ id }}</el-button>
            <span v-if="!selectedTask.article_ids.length">暂无关联草稿，不影响任务审核与发布。</span>
          </div>
        </section>
        <section>
          <div class="section-header">
            <h3>日报文本</h3>
            <el-button :disabled="!selectedTask.message.trim()" @click="copyMessage(selectedTask)">复制文本</el-button>
          </div>
          <el-input :model-value="selectedTask.message" type="textarea" :rows="12" readonly placeholder="暂无日报文本" aria-label="日报文本" />
        </section>
      </div>
      <el-empty v-else description="任务已不存在，请刷新列表" />
    </el-drawer>
  </div>
</template>

<style scoped>
.view, .detail { display: grid; gap: 20px; }
.toolbar, .toolbar-actions, .section-header { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.toolbar-actions { flex-wrap: wrap; justify-content: flex-end; }
h2 { margin: 0; font-size: 28px; }
h3, h4 { margin: 0 0 12px; }
p { margin: 8px 0 12px; color: rgba(17, 24, 39, 0.65); }
.error-text { color: var(--el-color-danger); }
.news-item { padding: 16px; margin-top: 12px; background: #f3f5f9; border-radius: 8px; }
.preserve-text { white-space: pre-wrap; overflow-wrap: anywhere; }
.draft-links { display: flex; flex-wrap: wrap; gap: 8px; }
.draft-links .el-button { margin-left: 0; }
.detail { overflow-wrap: anywhere; }
@media (max-width: 900px) {
  .toolbar { align-items: flex-start; flex-direction: column; }
  .toolbar-actions { justify-content: flex-start; }
}
</style>
