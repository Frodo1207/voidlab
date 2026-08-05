<script setup lang="ts">
import { onMounted, ref } from "vue";
import { ElMessage } from "element-plus";
import type { MediaAssetRecord } from "../../types";
import { listMediaAssets, uploadMediaAsset } from "../../services/mediaService";

const records = ref<MediaAssetRecord[]>([]);
const loading = ref(false);
const uploading = ref(false);
const fileInputRef = ref<HTMLInputElement | null>(null);

async function loadRecords() {
  loading.value = true;
  try {
    records.value = await listMediaAssets();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载媒体列表失败");
  } finally {
    loading.value = false;
  }
}

function openFilePicker() {
  fileInputRef.value?.click();
}

async function handleFileChange(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];

  if (!file) {
    return;
  }

  uploading.value = true;
  try {
    await uploadMediaAsset(file);
    ElMessage.success("媒体资源上传成功");
    await loadRecords();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "上传媒体失败");
  } finally {
    uploading.value = false;
    input.value = "";
  }
}

function handleCopy(url: string) {
  void navigator.clipboard.writeText(url);
  ElMessage.success("文件 URL 已复制");
}

onMounted(() => {
  void loadRecords();
});
</script>

<template>
  <div class="view">
    <div class="toolbar">
      <div>
        <h2>媒体资源</h2>
        <p>当前已支持真实上传与媒体列表，可复制 URL 用于文章、活动和 Builder 封面。</p>
      </div>
      <el-button type="primary" :loading="uploading" @click="openFilePicker">上传资源</el-button>
    </div>

    <input ref="fileInputRef" class="hidden-input" type="file" accept="image/*" @change="handleFileChange" />

    <el-card shadow="never">
      <el-table :data="records" v-loading="loading">
        <el-table-column prop="fileName" label="文件名" min-width="220" />
        <el-table-column prop="contentType" label="类型" min-width="160" />
        <el-table-column prop="fileSizeLabel" label="大小" min-width="120" />
        <el-table-column prop="createdAt" label="上传时间" min-width="180" />
        <el-table-column label="访问地址" min-width="260">
          <template #default="{ row }">
            <el-link :href="row.objectUrl" target="_blank" type="primary">{{ row.objectUrl }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleCopy(row.objectUrl)">复制 URL</el-button>
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

.hidden-input {
  display: none;
}
</style>
