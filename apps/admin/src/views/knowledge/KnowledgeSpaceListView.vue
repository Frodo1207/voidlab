<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { deleteKnowledgeSpace, listKnowledgeSpaces } from "../../services/knowledgeService";
import type { KnowledgeSpaceRecord } from "../../types";

const router = useRouter();
const records = ref<KnowledgeSpaceRecord[]>([]);
const loading = ref(false);

function statusTagType(status: KnowledgeSpaceRecord["status"]) {
  if (status === "published") return "success";
  if (status === "archived") return "info";
  return "warning";
}

function visibilityTagType(mode: KnowledgeSpaceRecord["visibilityMode"]) {
  if (mode === "public") return "success";
  if (mode === "private_hidden") return "danger";
  return "warning";
}

async function loadRecords() {
  loading.value = true;
  try {
    records.value = await listKnowledgeSpaces();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载知识空间失败");
  } finally {
    loading.value = false;
  }
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm("删除后会一并移除该 Space 下的文档与访问数据，确定继续吗？", "删除知识空间", {
    type: "warning"
  });

  try {
    await deleteKnowledgeSpace(id);
    ElMessage.success("知识空间已删除");
    await loadRecords();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "删除知识空间失败");
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
        <h2>知识库 Space</h2>
        <p>按项目维护知识空间，控制目录开放、正文解锁和整体发布状态。</p>
      </div>
      <el-button type="primary" @click="router.push({ name: 'knowledge-space-create' })">
        新建 Space
      </el-button>
    </div>

    <el-card shadow="never">
      <el-table :data="records" v-loading="loading">
        <el-table-column prop="title" label="项目名称" min-width="240" />
        <el-table-column prop="slug" label="Slug" min-width="180" />
        <el-table-column label="可见性" min-width="140">
          <template #default="{ row }">
            <el-tag :type="visibilityTagType(row.visibilityMode)">
              {{ row.visibilityMode }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="条目数" min-width="110">
          <template #default="{ row }">{{ row.entryCount }}</template>
        </el-table-column>
        <el-table-column label="章节数" min-width="110">
          <template #default="{ row }">{{ row.sectionCount }}</template>
        </el-table-column>
        <el-table-column label="状态" min-width="120">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updatedAt" label="更新时间" min-width="160" />
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              @click="router.push({ name: 'knowledge-entries', query: { spaceId: String(row.id) } })"
            >
              管理文档
            </el-button>
            <el-button
              link
              type="primary"
              @click="router.push({ name: 'knowledge-access-tokens', query: { spaceId: String(row.id) } })"
            >
              管理令牌
            </el-button>
            <el-button
              link
              type="primary"
              @click="router.push({ name: 'knowledge-space-edit', params: { id: row.id } })"
            >
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
