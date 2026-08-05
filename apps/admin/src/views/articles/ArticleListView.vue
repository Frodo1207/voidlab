<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import type { ArticleRecord } from "../../types";
import { deleteArticle, listArticles } from "../../services/articleService";

const router = useRouter();
const records = ref<ArticleRecord[]>([]);
const loading = ref(false);

function statusTagType(status: ArticleRecord["status"]) {
  if (status === "published") return "success";
  if (status === "archived") return "info";
  return "warning";
}

async function loadRecords() {
  loading.value = true;
  try {
    records.value = await listArticles();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载文章列表失败");
  } finally {
    loading.value = false;
  }
}

async function handleDelete(id: number) {
  await ElMessageBox.confirm("删除后不可恢复，确定继续吗？", "删除文章", {
    type: "warning"
  });

  try {
    await deleteArticle(id);
    ElMessage.success("文章已删除");
    await loadRecords();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "删除文章失败");
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
        <h2>资讯管理</h2>
        <p>这里先承接 Phase 1 的文章列表、编辑入口和状态字段。</p>
      </div>
      <el-button type="primary" @click="router.push({ name: 'article-create' })">新建文章</el-button>
    </div>

    <el-card shadow="never">
      <el-table :data="records" v-loading="loading">
        <el-table-column prop="title" label="标题" min-width="260" />
        <el-table-column prop="category" label="分类" min-width="140" />
        <el-table-column prop="audience" label="受众" min-width="140" />
        <el-table-column label="精选" min-width="100">
          <template #default="{ row }">
            <el-tag :type="row.featured ? 'success' : 'info'">
              {{ row.featured ? "是" : "否" }}
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
            <el-button link type="primary" @click="router.push({ name: 'article-edit', params: { id: row.id } })">
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
