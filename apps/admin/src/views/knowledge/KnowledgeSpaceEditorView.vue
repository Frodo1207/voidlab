<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { useRoute, useRouter } from "vue-router";
import MediaPicker from "../../components/MediaPicker.vue";
import {
  createKnowledgeSpace,
  getKnowledgeSpaceById,
  updateKnowledgeSpace
} from "../../services/knowledgeService";
import type { ContentStatus, KnowledgeVisibilityMode } from "../../types";

const route = useRoute();
const router = useRouter();

const spaceId = computed(() => {
  const raw = route.params.id;
  return raw ? Number(raw) : null;
});

const pageTitle = computed(() => (spaceId.value ? "编辑知识空间" : "新建知识空间"));
const loading = ref(false);
const saving = ref(false);
const currentStatus = ref<ContentStatus>("draft");
const statusOptionsByCurrent: Record<ContentStatus, ContentStatus[]> = {
  draft: ["draft", "published"],
  published: ["published", "draft", "archived"],
  archived: ["archived", "draft"]
};

const visibilityOptions: Array<{ value: KnowledgeVisibilityMode; label: string; desc: string }> = [
  { value: "directory_only", label: "目录开放", desc: "访客可看目录，正文需要 Space Token" },
  { value: "public", label: "全部公开", desc: "目录和正文都直接可读" },
  { value: "private_hidden", label: "完全隐藏", desc: "前台不直接开放该 Space" }
];

const form = reactive({
  title: "",
  slug: "",
  description: "",
  coverLabel: "",
  icon: "",
  themeTint: "#7c3aed",
  visibilityMode: "directory_only" as KnowledgeVisibilityMode,
  directorySummary: "",
  introMarkdown: "",
  tokenHint: "",
  coverUrl: "",
  status: "draft" as ContentStatus
});

const availableStatusOptions = computed(() => {
  if (!spaceId.value) {
    return ["draft", "published"] satisfies ContentStatus[];
  }

  return statusOptionsByCurrent[currentStatus.value];
});

async function loadDetail() {
  if (!spaceId.value) {
    return;
  }

  loading.value = true;
  try {
    const detail = await getKnowledgeSpaceById(spaceId.value);
    if (!detail) {
      ElMessage.warning("未找到知识空间，已保留为空白表单");
      return;
    }

    form.title = detail.title;
    form.slug = detail.slug;
    form.description = detail.description;
    form.coverLabel = detail.coverLabel;
    form.icon = detail.icon;
    form.themeTint = detail.themeTint || "#7c3aed";
    form.visibilityMode = detail.visibilityMode;
    form.directorySummary = detail.directorySummary;
    form.introMarkdown = detail.introMarkdown;
    form.tokenHint = detail.tokenHint;
    form.coverUrl = detail.coverUrl;
    form.status = detail.status;
    currentStatus.value = detail.status;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载知识空间详情失败");
  } finally {
    loading.value = false;
  }
}

async function handleSave(status?: ContentStatus) {
  if (status) {
    form.status = status;
  }

  saving.value = true;

  try {
    const payload = {
      title: form.title,
      slug: form.slug,
      description: form.description,
      cover_label: form.coverLabel,
      icon: form.icon,
      theme_tint: form.themeTint,
      visibility_mode: form.visibilityMode,
      directory_summary: form.directorySummary,
      intro_markdown: form.introMarkdown,
      token_hint: form.tokenHint,
      cover_url: form.coverUrl,
      status: form.status
    };

    if (spaceId.value) {
      await updateKnowledgeSpace(spaceId.value, payload);
    } else {
      await createKnowledgeSpace(payload);
    }

    ElMessage.success(`${pageTitle.value}已保存为 ${form.status}`);
    await router.push({ name: "knowledge-spaces" });
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "保存知识空间失败");
  } finally {
    saving.value = false;
  }
}

onMounted(() => {
  void loadDetail();
});
</script>

<template>
  <div class="view">
    <div class="toolbar">
      <div>
        <h2>{{ pageTitle }}</h2>
        <p>维护项目级知识空间的封面、目录说明、解锁方式和发布状态。</p>
      </div>
      <div class="actions">
        <el-button @click="router.push({ name: 'knowledge-spaces' })">返回列表</el-button>
        <el-button :loading="saving" @click="handleSave('draft')">保存草稿</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave('published')">
          发布 Space
        </el-button>
        <el-button
          v-if="spaceId && availableStatusOptions.includes('archived')"
          type="warning"
          :loading="saving"
          @click="handleSave('archived')"
        >
          归档 Space
        </el-button>
      </div>
    </div>

    <el-card shadow="never" v-loading="loading">
      <el-form label-position="top" class="form-grid">
        <el-form-item label="项目名称">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="Slug">
          <el-input v-model="form.slug" />
        </el-form-item>
        <el-form-item label="Icon">
          <el-input v-model="form.icon" placeholder="如：AI / KB / ∆" />
        </el-form-item>
        <el-form-item label="Theme Tint">
          <div class="tint-field">
            <span class="tint-dot" :style="{ background: form.themeTint || '#7c3aed' }" />
            <el-input v-model="form.themeTint" placeholder="#7c3aed" />
          </div>
        </el-form-item>
        <el-form-item label="封面 Label">
          <el-input v-model="form.coverLabel" placeholder="如：Project Knowledge" />
        </el-form-item>
        <el-form-item label="可见性">
          <el-select v-model="form.visibilityMode">
            <el-option
              v-for="item in visibilityOptions"
              :key="item.value"
              :label="`${item.label} · ${item.value}`"
              :value="item.value"
            />
          </el-select>
          <div class="field-hint">
            {{ visibilityOptions.find((item) => item.value === form.visibilityMode)?.desc }}
          </div>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option v-for="item in availableStatusOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="Token 提示语">
          <el-input v-model="form.tokenHint" placeholder="用于前台解锁弹窗提示" />
        </el-form-item>
        <el-form-item label="摘要说明" class="full">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="目录摘要" class="full">
          <el-input v-model="form.directorySummary" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="封面资源" class="full">
          <MediaPicker v-model="form.coverUrl" title="Space 封面" />
        </el-form-item>
        <el-form-item label="开场介绍 Markdown" class="full">
          <el-input
            v-model="form.introMarkdown"
            type="textarea"
            :rows="12"
            placeholder="这里用于前台 Space 页面顶部的项目介绍。"
          />
        </el-form-item>
      </el-form>
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

.tint-field {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  gap: 10px;
  align-items: center;
}

.tint-dot {
  width: 18px;
  height: 18px;
  border-radius: 999px;
  border: 1px solid rgba(17, 24, 39, 0.12);
}

.field-hint {
  margin-top: 6px;
  font-size: 12px;
  color: rgba(17, 24, 39, 0.55);
}
</style>
