<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import type { EventRecord } from "../../types";
import { deleteEvent, listEvents } from "../../services/eventService";

const router = useRouter();
const records = ref<EventRecord[]>([]);
const loading = ref(false);

function statusTagType(status: EventRecord["status"]) {
  if (status === "published") return "success";
  if (status === "archived") return "info";
  return "warning";
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
      <el-table :data="records" v-loading="loading">
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
</style>
