<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import type { LeadRecord, LeadSourceType, LeadStatus } from "../../types";
import { listLeads } from "../../services/leadService";

const router = useRouter();
const records = ref<LeadRecord[]>([]);
const loading = ref(false);

const statusMeta: Record<LeadStatus, { label: string; type: "" | "warning" | "success" | "info" | "danger" }> = {
  new: { label: "新线索", type: "warning" },
  contacted: { label: "已联系", type: "info" },
  following: { label: "跟进中", type: "" },
  converted: { label: "已转化", type: "success" },
  invalid: { label: "无效", type: "danger" }
};

const sourceMeta: Record<LeadSourceType, string> = {
  contact: "Contact",
  event: "活动报名",
  builder: "Builder 合作"
};

const groupedStats = computed(() => ({
  total: records.value.length,
  fresh: records.value.filter((item) => item.status === "new").length,
  following: records.value.filter((item) => item.status === "following").length,
  converted: records.value.filter((item) => item.status === "converted").length
}));

async function loadRecords() {
  loading.value = true;
  try {
    records.value = await listLeads();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载 Leads 列表失败");
  } finally {
    loading.value = false;
  }
}

function statusLabel(status: LeadStatus) {
  return statusMeta[status].label;
}

function statusType(status: LeadStatus) {
  return statusMeta[status].type;
}

function sourceLabel(sourceType: LeadSourceType) {
  return sourceMeta[sourceType];
}

onMounted(() => {
  void loadRecords();
});
</script>

<template>
  <div class="view">
    <div class="toolbar">
      <div>
        <h2>Leads 管理</h2>
        <p>把 Contact、活动报名和 Builder 合作意向统一沉淀到这里，方便后续跟进。</p>
      </div>
      <el-button @click="loadRecords">刷新列表</el-button>
    </div>

    <div class="stats-grid">
      <el-card shadow="never">
        <div class="stat-label">总线索</div>
        <div class="stat-value">{{ groupedStats.total }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">待处理</div>
        <div class="stat-value">{{ groupedStats.fresh }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">跟进中</div>
        <div class="stat-value">{{ groupedStats.following }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">已转化</div>
        <div class="stat-value">{{ groupedStats.converted }}</div>
      </el-card>
    </div>

    <el-card shadow="never">
      <el-table :data="records" v-loading="loading">
        <el-table-column prop="name" label="姓名" min-width="140" />
        <el-table-column prop="contact" label="联系方式" min-width="180" />
        <el-table-column label="来源" min-width="140">
          <template #default="{ row }">
            <el-tag effect="plain">{{ sourceLabel(row.sourceType) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="140">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" min-width="160" />
        <el-table-column label="操作" width="140" fixed="right">
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

h2 {
  margin: 0;
  font-size: 28px;
}

p {
  margin: 8px 0 0;
  color: rgba(17, 24, 39, 0.6);
}
</style>
