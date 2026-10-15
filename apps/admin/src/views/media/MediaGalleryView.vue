<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import type { FormInstance } from "element-plus";
import type { MediaAssetRecord, MediaGalleryItemRecord, MediaGalleryRecord } from "../../types";
import { listMediaAssets } from "../../services/mediaService";
import {
  addMediaGalleryItems,
  createMediaGallery,
  deleteMediaGallery,
  deleteMediaGalleryItem,
  listMediaGalleries,
  listMediaGalleryItems,
  updateMediaGallery,
  updateMediaGalleryItem
} from "../../services/mediaGalleryService";

const galleries = ref<MediaGalleryRecord[]>([]);
const galleryItems = ref<MediaGalleryItemRecord[]>([]);
const mediaAssets = ref<MediaAssetRecord[]>([]);
const selectedGalleryId = ref<number | null>(null);
const loading = ref(false);
const itemLoading = ref(false);
const assetLoading = ref(false);
const galleryDialogVisible = ref(false);
const assetDialogVisible = ref(false);
const savingGallery = ref(false);
const addingAssets = ref(false);
const selectedAssetIds = ref<number[]>([]);
const galleryFormRef = ref<FormInstance>();

const galleryForm = reactive({
  id: 0,
  title: "",
  slug: "",
  description: "",
  isDefault: false
});

const selectedGallery = computed(() =>
  galleries.value.find((item) => item.id === selectedGalleryId.value) ?? null
);

const imageAssets = computed(() => mediaAssets.value.filter((asset) => isImageAsset(asset)));

const selectableAssets = computed(() => {
  const used = new Set(galleryItems.value.map((item) => item.mediaAssetId));
  return imageAssets.value.filter((asset) => !used.has(asset.id));
});

function handleAssetSelectionChange(rows: MediaAssetRecord[]) {
  selectedAssetIds.value = rows.map((row) => row.id);
}

async function loadGalleries(preserveSelection = true) {
  loading.value = true;
  try {
    galleries.value = await listMediaGalleries();
    if (!galleries.value.length) {
      selectedGalleryId.value = null;
      galleryItems.value = [];
      return;
    }

    if (!preserveSelection || !selectedGalleryId.value || !galleries.value.some((item) => item.id === selectedGalleryId.value)) {
      selectedGalleryId.value = galleries.value[0].id;
    }

    await loadGalleryItems();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载图库失败");
  } finally {
    loading.value = false;
  }
}

async function loadGalleryItems() {
  if (!selectedGalleryId.value) {
    galleryItems.value = [];
    return;
  }

  itemLoading.value = true;
  try {
    galleryItems.value = await listMediaGalleryItems(selectedGalleryId.value);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载图库图片失败");
  } finally {
    itemLoading.value = false;
  }
}

async function loadMediaAssets() {
  assetLoading.value = true;
  try {
    mediaAssets.value = await listMediaAssets();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载媒体资源失败");
  } finally {
    assetLoading.value = false;
  }
}

function openCreateGalleryDialog() {
  galleryForm.id = 0;
  galleryForm.title = "";
  galleryForm.slug = "";
  galleryForm.description = "";
  galleryForm.isDefault = false;
  galleryDialogVisible.value = true;
}

function openEditGalleryDialog(record: MediaGalleryRecord) {
  galleryForm.id = record.id;
  galleryForm.title = record.title;
  galleryForm.slug = record.slug;
  galleryForm.description = record.description;
  galleryForm.isDefault = record.isDefault;
  galleryDialogVisible.value = true;
}

async function saveGallery() {
  if (!galleryFormRef.value) {
    return;
  }

  try {
    await galleryFormRef.value.validate();
  } catch {
    return;
  }

  savingGallery.value = true;
  try {
    const payload = {
      title: galleryForm.title.trim(),
      slug: galleryForm.slug.trim(),
      description: galleryForm.description.trim(),
      isDefault: galleryForm.isDefault
    };

    const record =
      galleryForm.id > 0
        ? await updateMediaGallery(galleryForm.id, payload)
        : await createMediaGallery(payload);

    selectedGalleryId.value = record.id;
    galleryDialogVisible.value = false;
    ElMessage.success(galleryForm.id > 0 ? "图库已更新" : "图库已创建");
    await loadGalleries(false);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "保存图库失败");
  } finally {
    savingGallery.value = false;
  }
}

async function handleDeleteGallery(record: MediaGalleryRecord) {
  try {
    await ElMessageBox.confirm(`确认删除图库「${record.title}」吗？图库内关联图片会一起移除，但原始媒体资源仍保留。`, "删除图库", {
      type: "warning"
    });
    await deleteMediaGallery(record.id);
    if (selectedGalleryId.value === record.id) {
      selectedGalleryId.value = null;
    }
    ElMessage.success("图库已删除");
    await loadGalleries(false);
  } catch (error) {
    if (error !== "cancel") {
      ElMessage.error(error instanceof Error ? error.message : "删除图库失败");
    }
  }
}

async function selectGallery(id: number) {
  selectedGalleryId.value = id;
  await loadGalleryItems();
}

function handleGalleryRowClick(row: MediaGalleryRecord) {
  void selectGallery(row.id);
}

async function openAddAssetsDialog() {
  if (!selectedGallery.value) {
    return;
  }
  if (!mediaAssets.value.length) {
    await loadMediaAssets();
  }
  selectedAssetIds.value = [];
  assetDialogVisible.value = true;
}

async function handleAddAssets() {
  if (!selectedGalleryId.value || !selectedAssetIds.value.length) {
    ElMessage.warning("请先选择至少一张图片");
    return;
  }

  addingAssets.value = true;
  try {
    galleryItems.value = await addMediaGalleryItems(selectedGalleryId.value, selectedAssetIds.value);
    selectedAssetIds.value = [];
    assetDialogVisible.value = false;
    ElMessage.success("图片已加入图库");
    await loadGalleries();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加入图库失败");
  } finally {
    addingAssets.value = false;
  }
}

async function handleSaveItem(item: MediaGalleryItemRecord) {
  if (!selectedGalleryId.value) {
    return;
  }

  try {
    galleryItems.value = await updateMediaGalleryItem(selectedGalleryId.value, item.id, {
      altText: item.altText,
      sortOrder: item.sortOrder,
      isActive: item.isActive
    });
    ElMessage.success("图片设置已保存");
    await loadGalleries();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "保存图片设置失败");
  }
}

async function handleDeleteItem(item: MediaGalleryItemRecord) {
  if (!selectedGalleryId.value) {
    return;
  }

  try {
    await ElMessageBox.confirm(`确认把「${item.fileName}」从当前图库移除吗？`, "移除图片", {
      type: "warning"
    });
    await deleteMediaGalleryItem(selectedGalleryId.value, item.id);
    ElMessage.success("图片已移出图库");
    await loadGalleryItems();
    await loadGalleries();
  } catch (error) {
    if (error !== "cancel") {
      ElMessage.error(error instanceof Error ? error.message : "移除图片失败");
    }
  }
}

function isImageAsset(asset: MediaAssetRecord) {
  const lowerName = asset.fileName.toLowerCase();
  return (
    asset.contentType.toLowerCase().startsWith("image/") ||
    [".jpg", ".jpeg", ".png", ".webp", ".gif", ".svg"].some((suffix) => lowerName.endsWith(suffix))
  );
}

onMounted(async () => {
  await Promise.all([loadGalleries(false), loadMediaAssets()]);
});
</script>

<template>
  <div class="view">
    <div class="toolbar">
      <div>
        <h2>图库管理</h2>
        <p>先把现有能正常显示的图片收成图库，后面 daily news 和文章封面都可以从这里稳定取图。</p>
      </div>
      <div class="toolbar-actions">
        <el-button @click="$router.push('/media')">返回媒体库</el-button>
        <el-button type="primary" @click="openCreateGalleryDialog">新建图库</el-button>
      </div>
    </div>

    <div class="layout">
      <el-card shadow="never" class="gallery-panel">
        <template #header>
          <div class="panel-header">
            <span>图库列表</span>
            <el-tag v-if="galleries.length" type="info">{{ galleries.length }} 个</el-tag>
          </div>
        </template>

        <el-table :data="galleries" v-loading="loading" @row-click="handleGalleryRowClick">
          <el-table-column label="封面" width="110">
            <template #default="{ row }">
              <el-image v-if="row.coverUrl" :src="row.coverUrl" fit="cover" class="thumb" />
              <div v-else class="thumb thumb-placeholder">无封面</div>
            </template>
          </el-table-column>
          <el-table-column prop="title" label="名称" min-width="180" />
          <el-table-column prop="slug" label="标识" min-width="160" />
          <el-table-column prop="itemCount" label="图片数" width="90" />
          <el-table-column prop="updatedAt" label="更新时间" min-width="160" />
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click.stop="openEditGalleryDialog(row)">编辑</el-button>
              <el-button link type="danger" @click.stop="handleDeleteGallery(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <el-card shadow="never" class="item-panel">
        <template #header>
          <div class="panel-header">
            <div>
              <div class="detail-title">{{ selectedGallery?.title ?? "请选择图库" }}</div>
              <div class="detail-desc">{{ selectedGallery?.description || "当前图库里的图片会作为后续封面池使用。" }}</div>
            </div>
            <div class="toolbar-actions" v-if="selectedGallery">
              <el-tag v-if="selectedGallery.isDefault" type="success">基础图库</el-tag>
              <el-button type="primary" @click="openAddAssetsDialog">加入现有图片</el-button>
            </div>
          </div>
        </template>

        <el-empty v-if="!selectedGallery" description="先从左侧选一个图库" />
        <el-table v-else :data="galleryItems" v-loading="itemLoading">
          <el-table-column label="预览" width="120">
            <template #default="{ row }">
              <el-image :src="row.objectUrl" fit="cover" class="thumb" />
            </template>
          </el-table-column>
          <el-table-column prop="fileName" label="文件名" min-width="220" />
          <el-table-column label="Alt 文案" min-width="220">
            <template #default="{ row }">
              <el-input v-model="row.altText" placeholder="可留空，后面可做 SEO / 图像描述" />
            </template>
          </el-table-column>
          <el-table-column label="顺序" width="120">
            <template #default="{ row }">
              <el-input-number v-model="row.sortOrder" :min="1" :step="10" />
            </template>
          </el-table-column>
          <el-table-column label="启用" width="90">
            <template #default="{ row }">
              <el-switch v-model="row.isActive" />
            </template>
          </el-table-column>
          <el-table-column prop="fileSizeLabel" label="大小" width="110" />
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="handleSaveItem(row)">保存</el-button>
              <el-button link type="danger" @click="handleDeleteItem(row)">移除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <el-dialog v-model="galleryDialogVisible" :title="galleryForm.id > 0 ? '编辑图库' : '新建图库'" width="560px">
      <el-form ref="galleryFormRef" :model="galleryForm" label-width="96px">
        <el-form-item label="名称" prop="title" required>
          <el-input v-model="galleryForm.title" placeholder="例如：基础图库 / 科技场景图库" />
        </el-form-item>
        <el-form-item label="标识">
          <el-input v-model="galleryForm.slug" placeholder="可留空，系统会根据名称生成" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="galleryForm.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="默认图库">
          <el-switch v-model="galleryForm.isDefault" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="galleryDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingGallery" @click="saveGallery">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="assetDialogVisible" title="把现有图片加入图库" width="960px">
      <el-table
        :data="selectableAssets"
        v-loading="assetLoading"
        @selection-change="handleAssetSelectionChange"
      >
        <el-table-column type="selection" width="48" />
        <el-table-column label="预览" width="120">
          <template #default="{ row }">
            <el-image :src="row.objectUrl" fit="cover" class="thumb" />
          </template>
        </el-table-column>
        <el-table-column prop="fileName" label="文件名" min-width="220" />
        <el-table-column prop="contentType" label="类型" min-width="160" />
        <el-table-column prop="fileSizeLabel" label="大小" width="120" />
        <el-table-column label="地址" min-width="240">
          <template #default="{ row }">
            <el-link :href="row.objectUrl" target="_blank" type="primary">{{ row.objectUrl }}</el-link>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
        <el-button @click="assetDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="addingAssets" @click="handleAddAssets">加入图库</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.view {
  display: grid;
  gap: 20px;
}

.toolbar,
.panel-header,
.toolbar-actions,
.layout {
  display: flex;
  gap: 16px;
}

.toolbar,
.panel-header {
  justify-content: space-between;
  align-items: center;
}

.toolbar-actions {
  align-items: center;
}

.layout {
  align-items: flex-start;
}

.gallery-panel {
  width: 42%;
}

.item-panel {
  flex: 1;
}

.detail-title,
h2 {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
}

.detail-title {
  font-size: 18px;
}

.detail-desc,
p {
  margin: 8px 0 0;
  color: rgba(17, 24, 39, 0.6);
}

.detail-desc {
  font-size: 13px;
}

.thumb {
  width: 88px;
  height: 60px;
  border-radius: 10px;
  overflow: hidden;
  background: #eef2ff;
}

.thumb-placeholder {
  display: grid;
  place-items: center;
  color: rgba(17, 24, 39, 0.45);
  font-size: 12px;
}
</style>
