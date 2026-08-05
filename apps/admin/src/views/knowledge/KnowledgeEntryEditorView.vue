<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { useRoute, useRouter } from "vue-router";
import MediaPicker from "../../components/MediaPicker.vue";
import {
  createKnowledgeEntry,
  getKnowledgeEntryById,
  importKnowledgeEntryMarkdown,
  listKnowledgeAssets,
  listKnowledgeSpaces,
  updateKnowledgeEntry,
  uploadKnowledgeAsset
} from "../../services/knowledgeService";
import type { ContentStatus, KnowledgeAssetRecord, KnowledgeSpaceRecord } from "../../types";

const route = useRoute();
const router = useRouter();

const entryId = computed(() => {
  const raw = route.params.id;
  return raw ? Number(raw) : null;
});

const initialSpaceId = computed(() => {
  const raw = route.query.spaceId;
  if (typeof raw !== "string" || raw.trim() === "") {
    return null;
  }

  const parsed = Number(raw);
  return Number.isFinite(parsed) ? parsed : null;
});

const pageTitle = computed(() => (entryId.value ? "编辑知识文档" : "新建知识文档"));
const loading = ref(false);
const saving = ref(false);
const markdownImporting = ref(false);
const assetLoading = ref(false);
const assetUploading = ref(false);
const currentStatus = ref<ContentStatus>("draft");
const spaces = ref<KnowledgeSpaceRecord[]>([]);
const assets = ref<KnowledgeAssetRecord[]>([]);
const fileInputRef = ref<HTMLInputElement | null>(null);
const markdownInputRef = ref<HTMLInputElement | null>(null);

const statusOptionsByCurrent: Record<ContentStatus, ContentStatus[]> = {
  draft: ["draft", "published"],
  published: ["published", "draft", "archived"],
  archived: ["archived", "draft"]
};

const form = reactive({
  spaceId: initialSpaceId.value ?? 0,
  title: "",
  slug: "",
  sectionName: "",
  sortOrder: 0,
  estimatedReadMinutes: 8,
  publicSummary: "",
  contentMarkdown: "",
  coverUrl: "",
  isPreview: false,
  status: "draft" as ContentStatus
});

const selectedSpace = computed(() =>
  spaces.value.find((item) => item.id === form.spaceId) ?? null
);

const availableStatusOptions = computed(() => {
  if (!entryId.value) {
    return ["draft", "published"] satisfies ContentStatus[];
  }

  return statusOptionsByCurrent[currentStatus.value];
});

async function loadSpaces() {
  spaces.value = await listKnowledgeSpaces();
}

async function loadDetail() {
  if (!entryId.value) {
    return;
  }

  loading.value = true;
  try {
    const detail = await getKnowledgeEntryById(entryId.value);
    if (!detail) {
      ElMessage.warning("未找到知识文档，已保留为空白表单");
      return;
    }

    form.spaceId = detail.spaceId;
    form.title = detail.title;
    form.slug = detail.slug;
    form.sectionName = detail.sectionName;
    form.sortOrder = detail.sortOrder;
    form.estimatedReadMinutes = detail.estimatedReadMinutes;
    form.publicSummary = detail.publicSummary;
    form.contentMarkdown = detail.contentMarkdown;
    form.coverUrl = detail.coverUrl;
    form.isPreview = detail.isPreview;
    form.status = detail.status;
    currentStatus.value = detail.status;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载知识文档详情失败");
  } finally {
    loading.value = false;
  }
}

async function loadAssetsForSpace(spaceId: number) {
  if (!spaceId) {
    assets.value = [];
    return;
  }

  assetLoading.value = true;
  try {
    assets.value = await listKnowledgeAssets(spaceId);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载知识资源失败");
  } finally {
    assetLoading.value = false;
  }
}

function openFilePicker() {
  if (!form.spaceId) {
    ElMessage.warning("请先选择所属 Space，再上传受控图片资源");
    return;
  }

  fileInputRef.value?.click();
}

function openMarkdownImportPicker() {
  markdownInputRef.value?.click();
}

function insertMarkdownSnippet(snippet: string) {
  form.contentMarkdown = form.contentMarkdown.trim()
    ? `${form.contentMarkdown.trimEnd()}\n\n${snippet}`
    : snippet;
  ElMessage.success("已插入 Markdown 引用");
}

async function handleMarkdownFileChange(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];

  if (!file) {
    return;
  }

  try {
    if (form.title || form.publicSummary || form.contentMarkdown) {
      await ElMessageBox.confirm("导入 Markdown 会覆盖当前标题、摘要、正文等字段，确定继续吗？", "导入 Markdown", {
        type: "warning"
      });
    }
  } catch {
    input.value = "";
    return;
  }

  markdownImporting.value = true;
  try {
    const result = await importKnowledgeEntryMarkdown(file);
    form.title = result.title;
    form.slug = result.slug;
    form.sectionName = result.sectionName;
    form.estimatedReadMinutes = result.estimatedReadMinutes;
    form.publicSummary = result.publicSummary;
    form.contentMarkdown = result.contentMarkdown;
    form.coverUrl = result.coverUrl;
    form.isPreview = result.isPreview;
    form.status = result.status;
    ElMessage.success("Markdown 已解析并回填到表单");
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "导入 Markdown 失败");
  } finally {
    markdownImporting.value = false;
    input.value = "";
  }
}

async function handleAssetFileChange(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];

  if (!file || !form.spaceId) {
    input.value = "";
    return;
  }

  assetUploading.value = true;
  try {
    const result = await uploadKnowledgeAsset(form.spaceId, file);
    assets.value = [result.asset, ...assets.value];
    insertMarkdownSnippet(result.markdownSnippet);
    ElMessage.success("知识图片上传成功，并已插入正文");
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "上传知识图片失败");
  } finally {
    assetUploading.value = false;
    input.value = "";
  }
}

async function handleSave(status?: ContentStatus) {
  if (status) {
    form.status = status;
  }

  saving.value = true;

  try {
    const payload = {
      space_id: form.spaceId,
      title: form.title,
      slug: form.slug,
      section_name: form.sectionName,
      sort_order: form.sortOrder,
      estimated_read_minutes: form.estimatedReadMinutes,
      public_summary: form.publicSummary,
      content_markdown: form.contentMarkdown,
      cover_url: form.coverUrl,
      is_preview: form.isPreview,
      status: form.status
    };

    if (entryId.value) {
      await updateKnowledgeEntry(entryId.value, payload);
    } else {
      await createKnowledgeEntry(payload);
    }

    ElMessage.success(`${pageTitle.value}已保存为 ${form.status}`);
    await router.push({
      name: "knowledge-entries",
      query: form.spaceId ? { spaceId: String(form.spaceId) } : {}
    });
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "保存知识文档失败");
  } finally {
    saving.value = false;
  }
}

watch(
  () => form.spaceId,
  (spaceId) => {
    if (spaceId > 0) {
      void loadAssetsForSpace(spaceId);
      return;
    }

    assets.value = [];
  },
  { immediate: true }
);

onMounted(async () => {
  try {
    await loadSpaces();
    await loadDetail();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "初始化知识文档表单失败");
  }
});
</script>

<template>
  <div class="view">
    <div class="toolbar">
      <div>
        <h2>{{ pageTitle }}</h2>
        <p>支持 Markdown 正文、预览章节和受控图片引用，适合直接维护教程/项目文档。</p>
      </div>
      <div class="actions">
        <el-button
          @click="router.push({ name: 'knowledge-entries', query: form.spaceId ? { spaceId: String(form.spaceId) } : {} })"
        >
          返回列表
        </el-button>
        <el-button :loading="saving" @click="handleSave('draft')">保存草稿</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave('published')">
          发布 Entry
        </el-button>
        <el-button
          v-if="entryId && availableStatusOptions.includes('archived')"
          type="warning"
          :loading="saving"
          @click="handleSave('archived')"
        >
          归档 Entry
        </el-button>
      </div>
    </div>

    <el-card shadow="never" v-loading="loading">
      <el-form label-position="top" class="form-grid">
        <el-form-item label="所属 Space">
          <el-select v-model="form.spaceId" placeholder="请选择知识空间">
            <el-option v-for="item in spaces" :key="item.id" :label="item.title" :value="item.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option v-for="item in availableStatusOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="标题">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="Slug">
          <el-input v-model="form.slug" />
        </el-form-item>
        <el-form-item label="章节">
          <el-input v-model="form.sectionName" placeholder="如：Getting Started / API / Deployment" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sortOrder" :min="0" :step="1" />
        </el-form-item>
        <el-form-item label="阅读时长（分钟）">
          <el-input-number v-model="form.estimatedReadMinutes" :min="0" :step="1" />
        </el-form-item>
        <el-form-item label="开放预览">
          <el-switch v-model="form.isPreview" />
        </el-form-item>
        <el-form-item label="公开摘要" class="full">
          <el-input v-model="form.publicSummary" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="封面资源" class="full">
          <MediaPicker v-model="form.coverUrl" title="Entry 封面" />
        </el-form-item>
        <el-form-item label="正文 Markdown" class="full">
          <div class="markdown-field">
            <div class="markdown-toolbar">
              <div class="markdown-toolbar-copy">
                <strong>{{ selectedSpace?.title ?? "未选择 Space" }}</strong>
                <span>支持导入 `.md` 文件，受控图片会以 `knowledge-asset://ID` 的形式插入正文。</span>
              </div>
              <div class="markdown-toolbar-actions">
                <el-button :loading="markdownImporting" @click="openMarkdownImportPicker">导入 Markdown</el-button>
                <el-button :disabled="!form.spaceId" @click="openFilePicker">上传知识图片</el-button>
              </div>
            </div>
            <input
              ref="fileInputRef"
              class="hidden-input"
              type="file"
              accept="image/*"
              @change="handleAssetFileChange"
            />
            <input
              ref="markdownInputRef"
              class="hidden-input"
              type="file"
              accept=".md,.markdown,text/markdown,text/plain"
              @change="handleMarkdownFileChange"
            />
            <el-input
              v-model="form.contentMarkdown"
              type="textarea"
              :rows="18"
              placeholder="在这里直接维护 Markdown 正文。"
            />
          </div>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never">
      <template #header>
        <div class="asset-header">
          <div>
            <strong>Space 内图片资源</strong>
            <div class="asset-subtitle">
              仅当前 Space 可用，前台会通过受控接口校验 grant 后再读取。
            </div>
          </div>
          <el-button type="primary" plain :loading="assetUploading" :disabled="!form.spaceId" @click="openFilePicker">
            上传并插入
          </el-button>
        </div>
      </template>

      <el-table :data="assets" v-loading="assetLoading" empty-text="当前 Space 还没有受控图片资源">
        <el-table-column prop="fileName" label="文件名" min-width="260" />
        <el-table-column prop="contentType" label="类型" min-width="160" />
        <el-table-column prop="createdAt" label="上传时间" min-width="160" />
        <el-table-column label="Markdown" min-width="220">
          <template #default="{ row }">
            <code class="snippet">knowledge-asset://{{ row.id }}</code>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button
              link
              type="primary"
              @click="insertMarkdownSnippet(`![${row.fileName}](knowledge-asset://${row.id})`)"
            >
              插入引用
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
  align-items: flex-start;
  gap: 20px;
}

.actions {
  display: flex;
  gap: 12px;
}

h2 {
  margin: 0;
  font-size: 28px;
}

p {
  margin: 8px 0 0;
  color: rgba(17, 24, 39, 0.6);
}

.form-grid {
  display: grid;
  gap: 12px 16px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.full {
  grid-column: 1 / -1;
}

.markdown-field {
  display: grid;
  gap: 12px;
}

.markdown-toolbar,
.asset-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.markdown-toolbar-copy,
.asset-subtitle {
  color: rgba(17, 24, 39, 0.6);
  font-size: 13px;
}

.markdown-toolbar-copy {
  display: grid;
  gap: 4px;
}

.markdown-toolbar-actions {
  display: flex;
  gap: 10px;
}

.hidden-input {
  display: none;
}

.snippet {
  display: inline-block;
  padding: 4px 8px;
  border-radius: 8px;
  background: #f3f4f6;
  color: #111827;
}
</style>
