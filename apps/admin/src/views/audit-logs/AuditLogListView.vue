<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { ElMessage } from "element-plus";
import { listAuditLogs } from "../../services/auditLogService";
import { getRoleLabel } from "../../permissions";
import type { AuditLogRecord } from "../../types";

const records = ref<AuditLogRecord[]>([]);
const loading = ref(false);

const summary = computed(() => ({
  total: records.value.length,
  contentOps: records.value.filter((item) => ["article", "event", "builder"].includes(item.entityType)).length,
  leadsOps: records.value.filter((item) => item.entityType === "lead").length,
  configOps: records.value.filter((item) => ["site_config", "media_asset"].includes(item.entityType)).length
}));

const actionLabelMap: Record<string, string> = {
  create: "创建",
  update: "更新",
  update_status: "状态变更",
  delete: "删除",
  add_log: "添加跟进",
  upload: "上传"
};

const entityLabelMap: Record<string, string> = {
  article: "资讯",
  event: "活动",
  builder: "Builder",
  lead: "Lead",
  site_config: "站点配置",
  media_asset: "媒体资源"
};

async function loadRecords() {
  loading.value = true;
  try {
    records.value = await listAuditLogs();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载审计日志失败");
  } finally {
    loading.value = false;
  }
}

function actionLabel(action: string) {
  return actionLabelMap[action] ?? action;
}

function entityTypeLabel(entityType: string) {
  return entityLabelMap[entityType] ?? entityType;
}

function detailPreview(detail: unknown) {
  if (!detail || typeof detail !== "object") {
    return "-";
  }

  return JSON.stringify(detail, null, 0);
}

onMounted(() => {
  void loadRecords();
});
</script>

<template>
  <div class="view">
    <div class="toolbar">
      <div>
        <h2>审计日志</h2>
        <p>记录后台关键写操作，支持按操作者、对象和时间回溯最近的运营动作。</p>
      </div>
      <el-button @click="loadRecords" :loading="loading">刷新日志</el-button>
    </div>

    <div class="stats-grid">
      <el-card shadow="never">
        <div class="stat-label">最近动作</div>
        <div class="stat-value">{{ summary.total }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">内容操作</div>
        <div class="stat-value">{{ summary.contentOps }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">Leads 操作</div>
        <div class="stat-value">{{ summary.leadsOps }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">配置与素材</div>
        <div class="stat-value">{{ summary.configOps }}</div>
      </el-card>
    </div>

    <el-card shadow="never">
      <el-table :data="records" v-loading="loading">
        <el-table-column prop="createdAt" label="时间" min-width="160" />
        <el-table-column label="操作者" min-width="150">
          <template #default="{ row }">
            <div class="user-cell">
              <strong>{{ row.actorUsername }}</strong>
              <span>{{ getRoleLabel(row.actorRole) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="动作" min-width="120">
          <template #default="{ row }">
            <el-tag effect="plain">{{ actionLabel(row.action) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="对象类型" min-width="120">
          <template #default="{ row }">
            {{ entityTypeLabel(row.entityType) }}
          </template>
        </el-table-column>
        <el-table-column prop="entityLabel" label="对象标识" min-width="180" />
        <el-table-column label="详情" min-width="260">
          <template #default="{ row }">
            <code class="detail-code">{{ detailPreview(row.detail) }}</code>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<style scoped>
.view {
  display: grid;
  gap: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
}

.stats-grid {
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.stat-label {
  font-size: 13px;
  color: rgba(17, 24, 39, 0.55);
}

.stat-value {
  margin-top: 8px;
  font-size: 32px;
  font-weight: 700;
  color: #111827;
}

.user-cell {
  display: grid;
  gap: 4px;
}

.user-cell span {
  font-size: 12px;
  color: rgba(17, 24, 39, 0.55);
}

.detail-code {
  display: -webkit-box;
  overflow: hidden;
  color: rgba(17, 24, 39, 0.72);
  font-size: 12px;
  line-height: 1.6;
  white-space: normal;
  line-clamp: 2;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

h2 {
  margin: 0;
  font-size: 28px;
}

p {
  margin: 8px 0 0;
  color: rgba(17, 24, 39, 0.6);
}
</style>
