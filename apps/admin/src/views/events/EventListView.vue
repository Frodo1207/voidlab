<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { listLeads } from "../../services/leadService";
import type { EventRecord, LeadRecord, LeadStatus } from "../../types";
import { deleteEvent, listEvents } from "../../services/eventService";

const router = useRouter();
const records = ref<EventRecord[]>([]);
const loading = ref(false);
const expandedRows = ref<number[]>([]);
const leadsByEventId = ref<Record<number, LeadRecord[]>>({});
const loadingLeadsByEventId = ref<Record<number, boolean>>({});

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

function statusTagType(status: EventRecord["status"]) {
  if (status === "published") return "success";
  if (status === "archived") return "info";
  return "warning";
}

function remainingSlots(row: EventRecord) {
  if (row.capacity <= 0) {
    return null;
  }

  return Math.max(row.capacity - row.signupCount, 0);
}

function leadStatusLabel(status: LeadStatus) {
  return leadStatusMeta[status].label;
}

function leadStatusType(status: LeadStatus) {
  return leadStatusMeta[status].type;
}

async function loadEventLeads(eventId: number) {
  if (loadingLeadsByEventId.value[eventId]) {
    return;
  }

  loadingLeadsByEventId.value = {
    ...loadingLeadsByEventId.value,
    [eventId]: true
  };

  try {
    const records = await listLeads({
      sourceType: "event",
      sourceId: eventId
    });

    leadsByEventId.value = {
      ...leadsByEventId.value,
      [eventId]: records
    };
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载报名详情失败");
  } finally {
    loadingLeadsByEventId.value = {
      ...loadingLeadsByEventId.value,
      [eventId]: false
    };
  }
}

function handleExpandChange(row: EventRecord, expanded: boolean) {
  if (expanded) {
    expandedRows.value = [row.id];
    if (!leadsByEventId.value[row.id]) {
      void loadEventLeads(row.id);
    }
    return;
  }

  expandedRows.value = expandedRows.value.filter((id) => id !== row.id);
}

async function loadRecords() {
  loading.value = true;
  try {
    records.value = await listEvents();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载活动列表失败");
  } finally {
    loading.value = false;
  }
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm("删除后不可恢复，确定继续吗？", "删除活动", {
    type: "warning"
  });

  try {
    await deleteEvent(id);
    ElMessage.success("活动已删除");
    await loadRecords();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "删除活动失败");
  }
}

onMounted(() => {
  void loadRecords();
});
</script>

<template>
  <div class="view">
    <div class="toolbar">
      <div>
        <h2>活动管理</h2>
        <p>承接活动主数据录入，包括时间、地点、类型和内容结构。</p>
      </div>
      <el-button type="primary" @click="router.push({ name: 'event-create' })">新建活动</el-button>
    </div>

    <el-card shadow="never">
      <el-table :data="records" v-loading="loading" row-key="id" :expand-row-keys="expandedRows" @expand-change="handleExpandChange">
        <el-table-column type="expand" width="52">
          <template #default="{ row }">
            <div class="expand-panel" v-loading="loadingLeadsByEventId[row.id]">
              <div class="expand-header">
                <div>
                  <div class="expand-title">报名详情</div>
                  <div class="expand-subtitle">在活动管理页直接看报名用户信息，不用再切到别处。</div>
                </div>
                <div class="expand-meta">
                  <span>总报名 {{ row.signupCount }} 人</span>
                  <span v-if="row.capacity > 0">名额 {{ row.capacity }}</span>
                  <span v-if="remainingSlots(row) !== null">剩余 {{ remainingSlots(row) }}</span>
                </div>
              </div>

              <div v-if="(leadsByEventId[row.id] ?? []).length === 0" class="expand-empty">
                这场活动暂时还没有报名记录。
              </div>

              <el-table v-else :data="leadsByEventId[row.id]" size="small">
                <el-table-column prop="name" label="姓名" min-width="120" />
                <el-table-column prop="contact" label="联系方式" min-width="180" />
                <el-table-column label="状态" min-width="120">
                  <template #default="{ row: lead }">
                    <el-tag :type="leadStatusType(lead.status)" size="small">
                      {{ leadStatusLabel(lead.status) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="createdAt" label="报名时间" min-width="160" />
                <el-table-column label="其他信息" min-width="320">
                  <template #default="{ row: lead }">
                    <div class="lead-message-cell">{{ lead.message || "无补充说明" }}</div>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="240" />
        <el-table-column prop="eventTime" label="时间" min-width="180" />
        <el-table-column prop="city" label="城市" min-width="120" />
        <el-table-column prop="eventType" label="类型" min-width="120" />
        <el-table-column label="状态" min-width="120">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="报名" min-width="140">
          <template #default="{ row }">
            <el-tag type="info">
              {{ row.signupStatus }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="报名人数" min-width="180">
          <template #default="{ row }">
            <div class="signup-count-cell">
              <strong>{{ row.signupCount }} 人</strong>
              <span v-if="row.capacity > 0"> / {{ row.capacity }} 名额</span>
            </div>
            <div v-if="remainingSlots(row) !== null" class="signup-subtext">
              剩余 {{ remainingSlots(row) }} 个名额
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="updatedAt" label="更新时间" min-width="160" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="router.push({ name: 'event-edit', params: { id: row.id } })">
              编辑
            </el-button>
            <el-button link type="danger" @click="handleDelete(row.id)">删除</el-button>
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

h2 {
  margin: 0;
  font-size: 28px;
}

p {
  margin: 8px 0 0;
  color: rgba(17, 24, 39, 0.6);
}

.signup-count-cell {
  font-size: 14px;
  color: #111827;
}

.signup-subtext {
  margin-top: 4px;
  font-size: 12px;
  color: rgba(17, 24, 39, 0.55);
}

.expand-panel {
  padding: 8px 8px 4px;
}

.expand-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 14px;
}

.expand-title {
  font-size: 15px;
  font-weight: 700;
  color: #111827;
}

.expand-subtitle {
  margin-top: 4px;
  font-size: 12px;
  color: rgba(17, 24, 39, 0.55);
}

.expand-meta {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  font-size: 12px;
  color: rgba(17, 24, 39, 0.62);
}

.expand-empty {
  border: 1px dashed rgba(17, 24, 39, 0.12);
  border-radius: 10px;
  padding: 18px;
  font-size: 13px;
  color: rgba(17, 24, 39, 0.55);
  background: #fafafa;
}

.lead-message-cell {
  white-space: pre-line;
  color: rgba(17, 24, 39, 0.72);
}
</style>
