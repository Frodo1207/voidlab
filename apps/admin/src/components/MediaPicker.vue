<script setup lang="ts">
import { computed, ref } from "vue";
import { ElMessage } from "element-plus";
import type { MediaAssetRecord } from "../types";
import { listMediaAssets, uploadMediaAsset } from "../services/mediaService";

const props = withDefaults(
  defineProps<{
    modelValue: string;
    title?: string;
  }>(),
  {
    title: "封面资源"
  }
);

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const dialogVisible = ref(false);
const loading = ref(false);
const uploading = ref(false);
const assets = ref<MediaAssetRecord[]>([]);
const fileInputRef = ref<HTMLInputElement | null>(null);

const hasValue = computed(() => Boolean(props.modelValue));

async function openDialog() {
  dialogVisible.value = true;
  await loadAssets();
}

async function loadAssets() {
  loading.value = true;
  try {
    assets.value = await listMediaAssets();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载媒体资源失败");
  } finally {
    loading.value = false;
  }
}

function selectAsset(asset: MediaAssetRecord) {
  emit("update:modelValue", asset.objectUrl);
  dialogVisible.value = false;
}

function clearValue() {
  emit("update:modelValue", "");
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
    const asset = await uploadMediaAsset(file);
    assets.value = [asset, ...assets.value];
    emit("update:modelValue", asset.objectUrl);
    ElMessage.success("资源上传成功并已设为当前封面");
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "上传资源失败");
  } finally {
    uploading.value = false;
    input.value = "";
  }
}
</script>

<template>
  <div class="media-picker">
    <div class="picker-header">
      <div>
        <div class="picker-title">{{ title }}</div>
        <div class="picker-hint">支持选择已有媒体或直接上传新资源</div>
      </div>
      <div class="picker-actions">
        <el-button @click="openDialog">选择媒体</el-button>
        <el-button type="primary" :loading="uploading" @click="openFilePicker">上传新图</el-button>
        <el-button v-if="hasValue" text type="danger" @click="clearValue">清空</el-button>
      </div>
    </div>

    <input ref="fileInputRef" class="hidden-input" type="file" accept="image/*" @change="handleFileChange" />

    <div v-if="hasValue" class="preview-card">
      <el-image :src="modelValue" fit="cover" class="preview-image">
        <template #error>
          <div class="preview-fallback">Preview unavailable</div>
        </template>
      </el-image>
      <div class="preview-meta">
        <div class="preview-label">当前封面</div>
        <el-link :href="modelValue" target="_blank" type="primary">{{ modelValue }}</el-link>
      </div>
    </div>
    <div v-else class="empty-state">还没有选择封面资源</div>

    <el-dialog v-model="dialogVisible" title="选择媒体资源" width="920px">
      <div class="dialog-toolbar">
        <div class="dialog-tip">从媒体库中选择已有资源，或先上传新资源再选择。</div>
        <el-button type="primary" :loading="uploading" @click="openFilePicker">上传新图</el-button>
      </div>

      <el-table :data="assets" v-loading="loading" empty-text="还没有媒体资源">
        <el-table-column label="预览" width="120">
          <template #default="{ row }">
            <el-image :src="row.objectUrl" fit="cover" class="asset-thumb">
              <template #error>
                <div class="asset-fallback">FILE</div>
              </template>
            </el-image>
          </template>
        </el-table-column>
        <el-table-column prop="fileName" label="文件名" min-width="220" />
        <el-table-column prop="contentType" label="类型" min-width="140" />
        <el-table-column prop="fileSizeLabel" label="大小" min-width="100" />
        <el-table-column prop="createdAt" label="上传时间" min-width="160" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="selectAsset(row)">使用</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<style scoped>
.media-picker {
  display: grid;
  gap: 14px;
}

.picker-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 16px;
}

.picker-title {
  font-size: 14px;
  font-weight: 600;
  color: #111827;
}

.picker-hint {
  margin-top: 4px;
  font-size: 12px;
  color: rgba(17, 24, 39, 0.55);
}

.picker-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.preview-card {
  display: grid;
  gap: 14px;
  grid-template-columns: 220px minmax(0, 1fr);
  padding: 16px;
  border: 1px solid rgba(17, 24, 39, 0.08);
  border-radius: 16px;
  background: #fafbfc;
}

.preview-image {
  width: 220px;
  height: 128px;
  border-radius: 12px;
  overflow: hidden;
  background: #e5e7eb;
}

.preview-meta {
  display: grid;
  align-content: start;
  gap: 8px;
}

.preview-label {
  font-size: 13px;
  font-weight: 600;
  color: #111827;
}

.preview-fallback,
.asset-fallback,
.empty-state {
  display: grid;
  place-items: center;
  color: rgba(17, 24, 39, 0.55);
  background: #eef2f7;
}

.preview-fallback {
  width: 100%;
  height: 100%;
}

.empty-state {
  min-height: 96px;
  border: 1px dashed rgba(17, 24, 39, 0.18);
  border-radius: 16px;
}

.dialog-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}

.dialog-tip {
  font-size: 13px;
  color: rgba(17, 24, 39, 0.6);
}

.asset-thumb {
  width: 88px;
  height: 56px;
  border-radius: 10px;
  overflow: hidden;
  background: #eef2f7;
}

.asset-fallback {
  width: 88px;
  height: 56px;
  font-size: 12px;
}

.hidden-input {
  display: none;
}
</style>
