<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  deleteKnowledgeEntry,
  listKnowledgeEntries,
  listKnowledgeSpaces
} from "../../services/knowledgeService";
import type { KnowledgeEntryRecord, KnowledgeSpaceRecord } from "../../types";

const route = useRoute();
const router = useRouter();

const spaces = ref<KnowledgeSpaceRecord[]>([]);
const records = ref<KnowledgeEntryRecord[]>([]);
const loading = ref(false);

const selectedSpaceId = computed<number | undefined>(() => {
  const raw = route.query.spaceId;
  if (typeof raw !== "string" || raw.trim() === "") {
    return undefined;
  }

  const parsed = Number(raw);
  return Number.isFinite(parsed) ? parsed : undefined;
});

const selectedSpaceLabel = computed(() => {
  const record = spaces.value.find((item) => item.id === selectedSpaceId.value);
  return record?.title ?? "全部 Space";
});

function statusTagType(status: KnowledgeEntryRecord["status"]) {
  if (status === "published") return "success";
  if (status === "archived") return "info";
  return "warning";
}

async function loadSpaces() {
  spaces.value = await listKnowledgeSpaces();
}

async function loadEntries() {
  loading.value = true;
  try {
    records.value = await listKnowledgeEntries(selectedSpaceId.value);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载知识文档失败");
  } finally {
    loading.value = false;
  }
}

async function loadData() {
  try {
    await loadSpaces();
    await loadEntries();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载知识库数据失败");
  }
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm("删除后不可恢复，确定继续吗？", "删除知识文档", {
    type: "warning"
  });

  try {
    await deleteKnowledgeEntry(id);
    ElMessage.success("知识文档已删除");
    await loadEntries();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "删除知识文档失败");
  }
}

function handleSpaceChange(spaceId: number | undefined) {
  void router.replace({
    name: "knowledge-entries",
    query: spaceId ? { spaceId: String(spaceId) } : {}
  });
}

watch(
  () => route.query.spaceId,
  () => {
    void loadEntries();
  }
);

onMounted(() => {
  void loadData();
});
</script>

<template>
  <div class="view">
    <div class="toolbar">
      <div>
        <h2>知识文档 Entry</h2>
        <p>当前筛选：{{ selectedSpaceLabel }}。这里维护正文、预览权限和目录排序。</p>
      </div>
      <div class="toolbar-actions">
        <el-select
          :model-value="selectedSpaceId"
          clearable
          placeholder="按 Space 筛选"
          class="space-filter"
          @update:model-value="handleSpaceChange"
        >
          <el-option v-for="item in spaces" :key="item.id" :label="item.title" :value="item.id" />
        </el-select>
        <el-button
          type="primary"
          @click="router.push({ name: 'knowledge-entry-create', query: selectedSpaceId ? { spaceId: String(selectedSpaceId) } : {} })"
        >
          新建 Entry
        </el-button>
      </div>
    </div>

    <el-card shadow="never">
      <el-table :data="records" v-loading="loading">
        <el-table-column prop="title" label="标题" min-width="240" />
        <el-table-column label="所属 Space" min-width="180">
          <template #default="{ row }">
            {{ spaces.find((item) => item.id === row.spaceId)?.title ?? row.spaceSlug ?? "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="sectionName" label="章节" min-width="160" />
        <el-table-column prop="sortOrder" label="排序" min-width="90" />
        <el-table-column prop="estimatedReadMinutes" label="阅读时长" min-width="110" />
        <el-table-column label="预览" min-width="100">
          <template #default="{ row }">
            <el-tag :type="row.isPreview ? 'success' : 'info'">
              {{ row.isPreview ? "是" : "否" }}
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
            <el-button
              link
              type="primary"
              @click="router.push({ name: 'knowledge-entry-edit', params: { id: row.id } })"
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

.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.space-filter {
  width: 240px;
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
