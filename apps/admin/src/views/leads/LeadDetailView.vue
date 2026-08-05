<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { ElMessage } from "element-plus";
import { useRoute, useRouter } from "vue-router";
import type { LeadRecord, LeadSourceType, LeadStatus } from "../../types";
import { addLeadLog, getLeadById, updateLeadStatus } from "../../services/leadService";

const route = useRoute();
const router = useRouter();

const leadId = computed(() => Number(route.params.id));
const loading = ref(false);
const savingStatus = ref(false);
const savingLog = ref(false);
const record = ref<LeadRecord | null>(null);
const selectedStatus = ref<LeadStatus>("new");
const nextLog = ref("");

const statusOptions: Array<{ label: string; value: LeadStatus }> = [
  { label: "新线索", value: "new" },
  { label: "已联系", value: "contacted" },
  { label: "跟进中", value: "following" },
  { label: "已转化", value: "converted" },
  { label: "无效", value: "invalid" }
];

const allowedLeadTransitions: Record<LeadStatus, LeadStatus[]> = {
  new: ["new", "contacted", "invalid"],
  contacted: ["contacted", "following", "invalid"],
  following: ["following", "contacted", "converted", "invalid"],
  converted: ["converted"],
  invalid: ["invalid"]
};

const sourceMeta: Record<LeadSourceType, string> = {
  contact: "Contact 页面",
  event: "活动报名",
  builder: "Builder 合作"
};

const sourceAction = computed(() => {
  if (!record.value?.sourceId) {
    return null;
  }

  if (record.value.sourceType === "event") {
    return {
      label: "查看活动对象",
      route: {
        name: "event-edit" as const,
        params: { id: record.value.sourceId }
      }
    };
  }

  if (record.value.sourceType === "builder") {
    return {
      label: "查看 Builder 对象",
      route: {
        name: "builder-edit" as const,
        params: { id: record.value.sourceId }
      }
    };
  }

  return null;
});

const availableStatusOptions = computed(() => {
  const currentStatus = record.value?.status ?? "new";
  const allowed = new Set(allowedLeadTransitions[currentStatus]);
  return statusOptions.filter((item) => allowed.has(item.value));
});

async function loadDetail() {
  loading.value = true;
  try {
    const detail = await getLeadById(leadId.value);
    if (!detail) {
      ElMessage.warning("未找到 Lead");
      await router.push({ name: "leads" });
      return;
    }

    record.value = detail;
    selectedStatus.value = detail.status;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载 Lead 详情失败");
  } finally {
    loading.value = false;
  }
}

async function handleStatusSave() {
  if (!record.value) {
    return;
  }

  savingStatus.value = true;
  try {
    record.value = await updateLeadStatus(record.value.id, selectedStatus.value);
    ElMessage.success("Lead 状态已更新");
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "更新状态失败");
  } finally {
    savingStatus.value = false;
  }
}

async function handleAddLog() {
  if (!record.value) {
    return;
  }

  savingLog.value = true;
  try {
    record.value = await addLeadLog(record.value.id, nextLog.value);
    nextLog.value = "";
    ElMessage.success("跟进备注已记录");
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "记录备注失败");
  } finally {
    savingLog.value = false;
  }
}

function sourceLabel(sourceType: LeadSourceType) {
  return sourceMeta[sourceType];
}

function openSourceObject() {
  if (!sourceAction.value) {
    return;
  }

  void router.push(sourceAction.value.route);
}

onMounted(() => {
  void loadDetail();
});
</script>

<template>
  <div class="view" v-loading="loading">
    <div class="toolbar">
      <div>
        <h2>Lead 详情</h2>
        <p>查看来源、留言与当前跟进状态，并持续记录跟进备注。</p>
      </div>
      <el-button @click="router.push({ name: 'leads' })">返回列表</el-button>
    </div>

    <template v-if="record">
      <div class="content-grid">
        <el-card shadow="never">
          <template #header>
            <div class="card-title">基础信息</div>
          </template>

          <div class="meta-list">
            <div class="meta-item">
              <span>姓名</span>
              <strong>{{ record.name }}</strong>
            </div>
            <div class="meta-item">
              <span>联系方式</span>
              <strong>{{ record.contact }}</strong>
            </div>
            <div class="meta-item">
              <span>来源</span>
              <strong>{{ sourceLabel(record.sourceType) }}</strong>
            </div>
            <div class="meta-item">
              <span>来源对象 ID</span>
              <div class="source-object">
                <strong>{{ record.sourceId ?? "-" }}</strong>
                <el-button
                  v-if="sourceAction"
                  link
                  type="primary"
                  @click="openSourceObject"
                >
                  {{ sourceAction.label }}
                </el-button>
              </div>
            </div>
            <div class="meta-item">
              <span>创建时间</span>
              <strong>{{ record.createdAt }}</strong>
            </div>
            <div class="meta-item">
              <span>更新时间</span>
              <strong>{{ record.updatedAt }}</strong>
            </div>
          </div>
        </el-card>

        <el-card shadow="never">
          <template #header>
            <div class="card-title">跟进状态</div>
          </template>

          <div class="status-block">
            <el-select v-model="selectedStatus">
              <el-option v-for="option in availableStatusOptions" :key="option.value" :label="option.label" :value="option.value" />
            </el-select>
            <el-button type="primary" :loading="savingStatus" @click="handleStatusSave">
              保存状态
            </el-button>
          </div>

          <div class="notes-box">
            <div class="notes-title">运营备注</div>
            <p>{{ record.notes || "当前还没有结构化备注，可先通过下方跟进日志记录。" }}</p>
          </div>
        </el-card>
      </div>

      <el-card shadow="never">
        <template #header>
          <div class="card-title">线索留言</div>
        </template>
        <p class="message-box">{{ record.message || "用户没有填写额外说明。" }}</p>
      </el-card>

      <el-card shadow="never">
        <template #header>
          <div class="card-title">新增跟进备注</div>
        </template>

        <div class="log-editor">
          <el-input v-model="nextLog" type="textarea" :rows="4" placeholder="记录这次联系结果、下一步动作或判断依据" />
          <div class="log-actions">
            <el-button type="primary" :loading="savingLog" @click="handleAddLog">保存备注</el-button>
          </div>
        </div>
      </el-card>

      <el-card shadow="never">
        <template #header>
          <div class="card-title">跟进日志</div>
        </template>

        <div v-if="record.logs.length === 0" class="empty-state">
          还没有跟进日志，先记录第一条动作。
        </div>

        <div v-else class="log-list">
          <div v-for="item in record.logs" :key="item.id" class="log-item">
            <div class="log-meta">
              <strong>{{ item.action }}</strong>
              <span>{{ item.createdAt }}</span>
            </div>
            <p>{{ item.content }}</p>
          </div>
        </div>
      </el-card>
    </template>
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

.content-grid {
  display: grid;
  gap: 20px;
  grid-template-columns: 1.2fr 1fr;
}

.card-title {
  font-weight: 700;
  color: #111827;
}

.meta-list {
  display: grid;
  gap: 14px;
}

.meta-item {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding-bottom: 14px;
  border-bottom: 1px solid rgba(15, 23, 35, 0.08);
}

.meta-item span {
  color: rgba(17, 24, 39, 0.55);
}

.meta-item strong {
  color: #111827;
  text-align: right;
}

.source-object {
  display: grid;
  justify-items: end;
  gap: 4px;
}

.status-block {
  display: flex;
  gap: 12px;
  align-items: center;
}

.notes-box {
  margin-top: 18px;
  padding: 16px;
  border-radius: 12px;
  background: #f8fafc;
}

.notes-title {
  margin-bottom: 8px;
  font-size: 13px;
  color: rgba(17, 24, 39, 0.55);
}

.message-box {
  margin: 0;
  line-height: 1.8;
  white-space: pre-line;
  color: #111827;
}

.log-editor {
  display: grid;
  gap: 12px;
}

.log-actions {
  display: flex;
  justify-content: flex-end;
}

.empty-state {
  color: rgba(17, 24, 39, 0.55);
}

.log-list {
  display: grid;
  gap: 16px;
}

.log-item {
  padding: 16px;
  border-radius: 12px;
  background: #f8fafc;
}

.log-item p {
  margin: 8px 0 0;
  line-height: 1.7;
  color: #111827;
  white-space: pre-line;
}

.log-meta {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
  color: rgba(17, 24, 39, 0.55);
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
