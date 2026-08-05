<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { getRoleLabel } from "../../permissions";
import type { AuditLogRecord, DashboardStatsRecord, LeadRecord, LeadSourceType, LeadStatusStatsRecord } from "../../types";
import { getDashboardStats } from "../../services/dashboardService";

const router = useRouter();
const loading = ref(false);
const stats = ref<DashboardStatsRecord | null>(null);

const overviewCards = computed(() => {
  if (!stats.value) {
    return [];
  }

  return [
    {
      label: "资讯总数",
      value: String(stats.value.articleCount),
      description: `已发布 ${stats.value.publishedArticleCount} 篇`
    },
    {
      label: "活动总数",
      value: String(stats.value.eventCount),
      description: `已发布 ${stats.value.publishedEventCount} 场`
    },
    {
      label: "Builder 总数",
      value: String(stats.value.builderCount),
      description: `已发布 ${stats.value.publishedBuilderCount} 位`
    },
    {
      label: "Leads 总数",
      value: String(stats.value.leadCount),
      description: "包含 Contact、活动报名、Builder 合作三类入口"
    },
    {
      label: "待处理 Leads",
      value: String(stats.value.actionableLeadCount),
      description: "新线索、已联系和跟进中的总量"
    }
  ];
});

const leadStatusItems = computed(() => {
  const distribution: LeadStatusStatsRecord = stats.value?.leadStatusDistribution ?? {
    new: 0,
    contacted: 0,
    following: 0,
    converted: 0,
    invalid: 0
  };

  return [
    { label: "新线索", value: distribution.new, type: "warning" as const },
    { label: "已联系", value: distribution.contacted, type: "info" as const },
    { label: "跟进中", value: distribution.following, type: "" as const },
    { label: "已转化", value: distribution.converted, type: "success" as const },
    { label: "无效", value: distribution.invalid, type: "danger" as const }
  ];
});

const leadTotal = computed(() => stats.value?.leadCount ?? 0);

const healthSummary = computed(() => {
  if (!stats.value) {
    return "正在加载当前运营概况。";
  }

  return `当前共有 ${stats.value.leadCount} 条 Leads，其中 ${stats.value.actionableLeadCount} 条仍需跟进；已发布内容 ${stats.value.publishedArticleCount + stats.value.publishedEventCount + stats.value.publishedBuilderCount} 条。`;
});

const recentActivities = computed(() => stats.value?.recentActivities ?? []);
const recentActionableLeads = computed(() => stats.value?.recentActionableLeads ?? []);

function formatActionLabel(action: AuditLogRecord["action"]) {
  switch (action) {
    case "create":
      return "创建";
    case "update":
      return "更新";
    case "update_status":
      return "状态变更";
    case "delete":
      return "删除";
    case "add_log":
      return "添加跟进";
    case "upload":
      return "上传";
    case "update_role":
      return "角色调整";
    case "reset_password":
      return "重置密码";
    default:
      return action;
  }
}

function formatEntityLabel(entityType: AuditLogRecord["entityType"]) {
  switch (entityType) {
    case "article":
      return "资讯";
    case "event":
      return "活动";
    case "builder":
      return "Builder";
    case "lead":
      return "Lead";
    case "site_config":
      return "站点配置";
    case "media_asset":
      return "媒体资源";
    case "user":
      return "用户";
    default:
      return entityType;
  }
}

function sourceLabel(sourceType: LeadSourceType) {
  switch (sourceType) {
    case "contact":
      return "Contact";
    case "event":
      return "活动报名";
    case "builder":
      return "Builder 合作";
    default:
      return sourceType;
  }
}

function leadTagType(status: LeadRecord["status"]) {
  switch (status) {
    case "new":
      return "warning";
    case "contacted":
      return "info";
    case "following":
      return "";
    case "converted":
      return "success";
    case "invalid":
      return "danger";
    default:
      return "info";
  }
}

function leadStatusLabel(status: LeadRecord["status"]) {
  switch (status) {
    case "new":
      return "新线索";
    case "contacted":
      return "已联系";
    case "following":
      return "跟进中";
    case "converted":
      return "已转化";
    case "invalid":
      return "无效";
    default:
      return status;
  }
}

async function loadDashboard() {
  loading.value = true;
  try {
    stats.value = await getDashboardStats();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载 Dashboard 统计失败");
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  void loadDashboard();
});
</script>

<template>
  <div class="view">
    <div class="hero" v-loading="loading">
      <div>
        <div class="eyebrow">PHASE 3 / OPERATIONS</div>
        <h2>运营总览看板</h2>
        <p>{{ healthSummary }}</p>
      </div>
      <div class="hero-actions">
        <el-tag type="success" size="large">Dashboard Live</el-tag>
        <el-button @click="loadDashboard" :loading="loading">刷新统计</el-button>
      </div>
    </div>

    <div class="card-grid">
      <el-card v-for="card in overviewCards" :key="card.label" shadow="hover">
        <div class="stat-label">{{ card.label }}</div>
        <div class="stat-value">{{ card.value }}</div>
        <p>{{ card.description }}</p>
      </el-card>
    </div>

    <div class="dashboard-grid">
      <el-card shadow="never">
        <template #header>Leads 状态分布</template>
        <div class="lead-grid">
          <div v-for="item in leadStatusItems" :key="item.label" class="lead-item">
            <div class="lead-item-top">
              <span>{{ item.label }}</span>
              <el-tag :type="item.type" effect="plain">{{ item.value }}</el-tag>
            </div>
            <div class="lead-progress">
              <div
                class="lead-progress-bar"
                :style="{ width: `${leadTotal ? (item.value / leadTotal) * 100 : 0}%` }"
              />
            </div>
          </div>
        </div>
      </el-card>

      <el-card shadow="never">
        <template #header>最近操作流</template>
        <div v-if="recentActivities.length" class="activity-list">
          <div v-for="item in recentActivities" :key="item.id" class="activity-item">
            <div class="activity-main">
              <div class="activity-title">
                <strong>{{ item.actorUsername }}</strong>
                <span>{{ formatActionLabel(item.action) }}</span>
                <span>{{ formatEntityLabel(item.entityType) }}</span>
                <strong>{{ item.entityLabel || "-" }}</strong>
              </div>
              <div class="activity-meta">
                <span>{{ getRoleLabel(item.actorRole) }}</span>
                <span>{{ item.createdAt }}</span>
              </div>
            </div>
          </div>
        </div>
        <el-empty v-else description="暂时还没有最近操作" />
      </el-card>
    </div>

    <el-card shadow="never">
      <template #header>待处理 Leads</template>
      <div v-if="recentActionableLeads.length" class="lead-alert-list">
        <div v-for="item in recentActionableLeads" :key="item.id" class="lead-alert-item">
          <div>
            <div class="lead-alert-title">
              <strong>{{ item.name }}</strong>
              <el-tag :type="leadTagType(item.status)" effect="plain">{{ leadStatusLabel(item.status) }}</el-tag>
            </div>
            <div class="lead-alert-meta">
              <span>{{ sourceLabel(item.sourceType) }}</span>
              <span>{{ item.contact }}</span>
              <span>更新于 {{ item.updatedAt }}</span>
            </div>
            <p class="lead-alert-message">{{ item.message || "暂无补充信息" }}</p>
          </div>
          <el-button link type="primary" @click="router.push({ name: 'lead-detail', params: { id: item.id } })">
            去跟进
          </el-button>
        </div>
      </div>
      <el-empty v-else description="当前没有需要优先处理的 Leads" />
    </el-card>
  </div>
</template>

<style scoped>
.view {
  display: grid;
  gap: 24px;
}

.hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  padding: 28px;
  border-radius: 24px;
  background: linear-gradient(135deg, #101828, #182230);
  color: #fff;
}

.hero-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.eyebrow {
  font-size: 12px;
  letter-spacing: 0.18em;
  color: rgba(255, 255, 255, 0.56);
}

h2 {
  margin: 14px 0 12px;
  font-size: 30px;
}

p {
  margin: 0;
  line-height: 1.8;
  color: rgba(255, 255, 255, 0.72);
}

.card-grid {
  display: grid;
  gap: 18px;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
}

.dashboard-grid {
  display: grid;
  gap: 24px;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1.05fr);
}

.lead-grid {
  display: grid;
  gap: 16px;
}

.lead-item {
  display: grid;
  gap: 10px;
}

.lead-item-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  color: #111827;
  font-size: 14px;
}

.lead-progress {
  height: 10px;
  border-radius: 999px;
  background: rgba(15, 23, 35, 0.08);
  overflow: hidden;
}

.lead-progress-bar {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #42ffd1, #0ea5e9);
}

.activity-list,
.lead-alert-list {
  display: grid;
  gap: 14px;
}

.activity-item,
.lead-alert-item {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 18px;
  border-radius: 16px;
  background: rgba(15, 23, 35, 0.04);
}

.activity-main,
.lead-alert-item > div:first-child {
  display: grid;
  gap: 8px;
}

.activity-title,
.activity-meta,
.lead-alert-meta,
.lead-alert-title {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  align-items: center;
}

.activity-meta,
.lead-alert-meta {
  font-size: 12px;
  color: rgba(17, 24, 39, 0.55);
}

.lead-alert-message {
  margin: 0;
  color: rgba(17, 24, 39, 0.68);
  line-height: 1.6;
}

.stat-label {
  font-size: 13px;
  color: rgba(17, 24, 39, 0.55);
}

.stat-value {
  margin: 10px 0 8px;
  font-size: 24px;
  font-weight: 700;
  color: #111827;
}

.card-grid p {
  color: rgba(17, 24, 39, 0.62);
}

@media (max-width: 1080px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
  }
}
</style>
