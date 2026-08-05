<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import type { BuilderRecord } from "../../types";
import { deleteBuilder, listBuilders } from "../../services/builderService";

const router = useRouter();
const records = ref<BuilderRecord[]>([]);
const loading = ref(false);

function statusTagType(status: BuilderRecord["status"]) {
  if (status === "published") return "success";
  if (status === "archived") return "info";
  return "warning";
}

async function loadRecords() {
  loading.value = true;
  try {
    records.value = await listBuilders();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载 Builder 列表失败");
  } finally {
    loading.value = false;
  }
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm("删除后不可恢复，确定继续吗？", "删除 Builder", {
    type: "warning"
  });

  try {
    await deleteBuilder(id);
    ElMessage.success("Builder 已删除");
    await loadRecords();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "删除 Builder 失败");
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
        <h2>Builder 管理</h2>
        <p>先把人物档案的结构、标签和合作字段统一起来，方便后续接前台展示。</p>
      </div>
      <el-button type="primary" @click="router.push({ name: 'builder-create' })">新建 Builder</el-button>
    </div>

    <el-card shadow="never">
      <el-table :data="records" v-loading="loading">
        <el-table-column prop="name" label="姓名" min-width="180" />
        <el-table-column prop="role" label="角色" min-width="140" />
        <el-table-column prop="city" label="城市" min-width="120" />
        <el-table-column label="Featured" min-width="100">
          <template #default="{ row }">
            <el-tag :type="row.featured ? 'success' : 'info'">
              {{ row.featured ? "是" : "否" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="可联系" min-width="100">
          <template #default="{ row }">
            <el-tag :type="row.contactable ? 'success' : 'warning'">
              {{ row.contactable ? "开放" : "关闭" }}
            </el-tag>
          </template>
        </el-table-column>
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
            <el-button link type="primary" @click="router.push({ name: 'builder-edit', params: { id: row.id } })">
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
