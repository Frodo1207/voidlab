<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import { useRoute, useRouter } from "vue-router";
import MediaPicker from "../../components/MediaPicker.vue";
import { listLeads } from "../../services/leadService";
import { createEvent, getEventById, updateEvent } from "../../services/eventService";
import type { ContentStatus, LeadRecord, LeadStatus } from "../../types";

const route = useRoute();
const router = useRouter();

const eventId = computed(() => {
  const raw = route.params.id;
  return raw ? Number(raw) : null;
});

const pageTitle = computed(() => (eventId.value ? "编辑活动" : "新建活动"));
const loading = ref(false);
const saving = ref(false);
const loadingLeads = ref(false);
const currentStatus = ref<ContentStatus>("draft");
const eventLeads = ref<LeadRecord[]>([]);
const contentStatusOptions: Record<ContentStatus, ContentStatus[]> = {
  draft: ["draft", "published"],
  published: ["published", "draft", "archived"],
  archived: ["archived", "draft"]
};
const leadStatusMeta: Record<LeadStatus, { label: string; type: "" | "warning" | "success" | "info" | "danger" }> = {
  new: { label: "新线索", type: "warning" },
  applied: { label: "已报名", type: "warning" },
  approved: { label: "已通过", type: "success" },
  waitlisted: { label: "候补中", type: "info" },
  rejected: { label: "已拒绝", type: "danger" },
  checked_in: { label: "已签到", type: "success" },
  contacted: { label: "已联系", type: "info" },
  following: { label: "跟进中", type: "" },
  converted: { label: "已转化", type: "success" },
  invalid: { label: "无效", type: "danger" }
};

const form = reactive({
  title: "",
  slug: "",
  summary: "",
  city: "",
  location: "",
  eventType: "",
  eventTime: "",
  coverUrl: "",
  content: "",
  status: "draft" as ContentStatus,
  signupMode: "internal" as "internal" | "external" | "closed",
  signupEnabled: true,
  signupStartsAt: "",
  signupDeadline: "",
  capacity: 0,
  allowSignupDuringLive: false,
  externalSignupUrl: "",
  signupButtonLabel: "",
  signupSuccessMessage: "",
  signupClosedReason: ""
});

const availableStatusOptions = computed(() => {
  if (!eventId.value) {
    return ["draft", "published"] satisfies ContentStatus[];
  }

  return contentStatusOptions[currentStatus.value];
});

const signupOverview = computed(() => {
  const total = eventLeads.value.length;
  const approved = eventLeads.value.filter((lead) => lead.status === "approved" || lead.status === "checked_in").length;
  const pending = eventLeads.value.filter((lead) => lead.status === "applied" || lead.status === "waitlisted").length;
  const checkedIn = eventLeads.value.filter((lead) => lead.status === "checked_in").length;
  const remaining = form.capacity > 0 ? Math.max(form.capacity - approved, 0) : null;

  return {
    total,
    approved,
    pending,
    checkedIn,
    remaining
  };
});

async function loadDetail() {
  if (!eventId.value) {
    return;
  }

  loading.value = true;
  try {
    const detail = await getEventById(eventId.value);
    if (!detail) {
      ElMessage.warning("未找到活动，已保留为空白表单");
      return;
    }

    form.title = detail.title;
    form.slug = detail.slug;
    form.summary = detail.summary;
    form.city = detail.city;
    form.location = detail.location;
    form.eventType = detail.eventType;
    form.eventTime = detail.eventTime;
    form.coverUrl = detail.coverUrl;
    form.content = detail.content;
    form.status = detail.status;
    form.signupMode = detail.signupMode;
    form.signupEnabled = detail.signupEnabled;
    form.signupStartsAt = detail.signupStartsAt;
    form.signupDeadline = detail.signupDeadline;
    form.capacity = detail.capacity;
    form.allowSignupDuringLive = detail.allowSignupDuringLive;
    form.externalSignupUrl = detail.externalSignupUrl;
    form.signupButtonLabel = detail.signupButtonLabel;
    form.signupSuccessMessage = detail.signupSuccessMessage;
    form.signupClosedReason = detail.signupClosedReason;
    currentStatus.value = detail.status;
    await loadEventLeads();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载活动详情失败");
  } finally {
    loading.value = false;
  }
}

async function loadEventLeads() {
  if (!eventId.value) {
    eventLeads.value = [];
    return;
  }

  loadingLeads.value = true;
  try {
    eventLeads.value = await listLeads({
      sourceType: "event",
      sourceId: eventId.value
    });
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载报名名单失败");
  } finally {
    loadingLeads.value = false;
  }
}

async function handleSave(status?: ContentStatus) {
  if (status) {
    form.status = status;
  }

  saving.value = true;

  try {
    const payload = {
      title: form.title,
      slug: form.slug,
      summary: form.summary,
      city: form.city,
      location: form.location,
      event_type: form.eventType,
      event_time: form.eventTime,
      cover_url: form.coverUrl,
      content: form.content,
      status: form.status,
      signup_mode: form.signupMode,
      signup_enabled: form.signupEnabled,
      signup_starts_at: form.signupStartsAt,
      signup_deadline: form.signupDeadline,
      capacity: form.capacity,
      allow_signup_during_live: form.allowSignupDuringLive,
      external_signup_url: form.externalSignupUrl,
      signup_button_label: form.signupButtonLabel,
      signup_success_message: form.signupSuccessMessage,
      signup_closed_reason: form.signupClosedReason
    };

    if (eventId.value) {
      await updateEvent(eventId.value, payload);
      ElMessage.success(`${pageTitle.value}已保存为 ${form.status}`);
      await loadDetail();
    } else {
      const created = await createEvent(payload);
      ElMessage.success(`活动已创建并保存为 ${form.status}`);
      await router.push({ name: "event-edit", params: { id: created.id } });
      return;
    }
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "保存活动失败");
  } finally {
    saving.value = false;
  }
}

function leadStatusLabel(status: LeadStatus) {
  return leadStatusMeta[status].label;
}

function leadStatusType(status: LeadStatus) {
  return leadStatusMeta[status].type;
}

watch(eventId, () => {
  void loadDetail();
}, { immediate: true });

watch(() => route.fullPath, () => {
  if (!eventId.value) {
    eventLeads.value = [];
  }
});
</script>

<template>
  <div class="view">
    <div class="toolbar">
      <div>
        <h2>{{ pageTitle }}</h2>
        <p>活动模块先把主数据录入和封面、状态、正文结构统一下来。</p>
      </div>
      <div class="actions">
        <el-button @click="router.push({ name: 'events' })">返回列表</el-button>
        <el-button :loading="saving" @click="handleSave('draft')">保存草稿</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave('published')">发布活动</el-button>
        <el-button
          v-if="eventId && availableStatusOptions.includes('archived')"
          type="warning"
          :loading="saving"
          @click="handleSave('archived')"
        >
          归档活动
        </el-button>
      </div>
    </div>

    <div v-if="eventId" class="stats-grid">
      <el-card shadow="never">
        <div class="stat-label">报名总数</div>
        <div class="stat-value">{{ signupOverview.total }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">已通过</div>
        <div class="stat-value">{{ signupOverview.approved }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">待处理 / 候补</div>
        <div class="stat-value">{{ signupOverview.pending }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">{{ form.capacity > 0 ? "剩余名额" : "已签到" }}</div>
        <div class="stat-value">{{ form.capacity > 0 ? signupOverview.remaining : signupOverview.checkedIn }}</div>
      </el-card>
    </div>

    <el-card shadow="never" v-loading="loading">
      <el-form label-position="top" class="form-grid">
        <el-form-item label="标题">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="Slug">
          <el-input v-model="form.slug" />
        </el-form-item>
        <el-form-item label="城市">
          <el-input v-model="form.city" />
        </el-form-item>
        <el-form-item label="地点">
          <el-input v-model="form.location" />
        </el-form-item>
        <el-form-item label="活动类型">
          <el-input v-model="form.eventType" />
        </el-form-item>
        <el-form-item label="活动时间">
          <el-input v-model="form.eventTime" placeholder="YYYY-MM-DD HH:mm" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option v-for="item in availableStatusOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="报名模式">
          <el-select v-model="form.signupMode">
            <el-option label="站内表单" value="internal" />
            <el-option label="外部链接" value="external" />
            <el-option label="手动关闭" value="closed" />
          </el-select>
        </el-form-item>
        <el-form-item label="开放报名">
          <el-switch v-model="form.signupEnabled" />
        </el-form-item>
        <el-form-item label="报名开始时间">
          <el-input v-model="form.signupStartsAt" placeholder="YYYY-MM-DD HH:mm" />
        </el-form-item>
        <el-form-item label="报名截止时间">
          <el-input v-model="form.signupDeadline" placeholder="YYYY-MM-DD HH:mm" />
        </el-form-item>
        <el-form-item label="名额上限">
          <el-input-number v-model="form.capacity" :min="0" :step="1" class="w-full" />
        </el-form-item>
        <el-form-item label="活动进行中允许报名">
          <el-switch v-model="form.allowSignupDuringLive" />
        </el-form-item>
        <el-form-item v-if="form.signupMode === 'external'" label="外部报名链接" class="full">
          <el-input v-model="form.externalSignupUrl" placeholder="https://..." />
        </el-form-item>
        <el-form-item label="报名按钮文案" class="full">
          <el-input v-model="form.signupButtonLabel" placeholder="例如：立即报名 / 预约席位" />
        </el-form-item>
        <el-form-item label="报名成功提示" class="full">
          <el-input v-model="form.signupSuccessMessage" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="关闭提示文案" class="full">
          <el-input v-model="form.signupClosedReason" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="封面资源" class="full">
          <MediaPicker v-model="form.coverUrl" title="活动封面" />
        </el-form-item>
        <el-form-item label="摘要" class="full">
          <el-input v-model="form.summary" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="活动详情" class="full">
          <el-input v-model="form.content" type="textarea" :rows="10" />
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-if="eventId" shadow="never" v-loading="loadingLeads">
      <template #header>
        <div class="section-header">
          <div>
            <div class="section-title">报名名单</div>
            <div class="section-subtitle">
              当前共有 {{ signupOverview.total }} 人报名
              <span v-if="form.capacity > 0">，名额上限 {{ form.capacity }}，剩余 {{ signupOverview.remaining }}</span>
            </div>
          </div>
          <el-button @click="loadEventLeads">刷新名单</el-button>
        </div>
      </template>

      <el-table :data="eventLeads">
        <el-table-column prop="name" label="姓名" min-width="120" />
        <el-table-column prop="contact" label="联系方式" min-width="180" />
        <el-table-column label="状态" min-width="120">
          <template #default="{ row }">
            <el-tag :type="leadStatusType(row.status)">
              {{ leadStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="报名时间" min-width="160" />
        <el-table-column label="说明" min-width="280">
          <template #default="{ row }">
            <div class="lead-message">{{ row.message || "无补充说明" }}</div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="router.push({ name: 'lead-detail', params: { id: row.id } })">
              查看详情
            </el-button>
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
  align-items: flex-start;
  gap: 20px;
}

.actions {
  display: flex;
  gap: 12px;
}

.stats-grid {
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

h2 {
  margin: 0;
  font-size: 28px;
}

p {
  margin: 8px 0 0;
  color: rgba(17, 24, 39, 0.6);
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

.form-grid {
  display: grid;
  gap: 12px 16px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.full {
  grid-column: 1 / -1;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.section-title {
  font-size: 18px;
  font-weight: 700;
  color: #111827;
}

.section-subtitle {
  margin-top: 6px;
  font-size: 13px;
  color: rgba(17, 24, 39, 0.6);
}

.lead-message {
  white-space: pre-line;
  color: rgba(17, 24, 39, 0.72);
}
</style>
