<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { useRoute, useRouter } from "vue-router";
import MediaPicker from "../../components/MediaPicker.vue";
import { createArticle, getArticleById, updateArticle } from "../../services/articleService";
import type { ContentStatus } from "../../types";

const route = useRoute();
const router = useRouter();

const articleId = computed(() => {
  const raw = route.params.id;
  return raw ? Number(raw) : null;
});

const pageTitle = computed(() => (articleId.value ? "编辑文章" : "新建文章"));
const loading = ref(false);
const saving = ref(false);
const currentStatus = ref<ContentStatus>("draft");
const contentStatusOptions: Record<ContentStatus, ContentStatus[]> = {
  draft: ["draft", "published"],
  published: ["published", "draft", "archived"],
  archived: ["archived", "draft"]
};

const form = reactive({
  title: "",
  slug: "",
  summary: "",
  category: "",
  audience: "",
  tagsText: "",
  coverUrl: "",
  content: "",
  sourceName: "",
  sourceUrl: "",
  featured: false,
  status: "draft" as ContentStatus
});

const availableStatusOptions = computed(() => {
  if (!articleId.value) {
    return ["draft", "published"] satisfies ContentStatus[];
  }

  return contentStatusOptions[currentStatus.value];
});

async function loadDetail() {
  if (!articleId.value) {
    return;
  }

  loading.value = true;
  try {
    const detail = await getArticleById(articleId.value);
    if (!detail) {
      ElMessage.warning("未找到文章，已保留为空白表单");
      return;
    }

    form.title = detail.title;
    form.slug = detail.slug;
    form.summary = detail.summary;
    form.category = detail.category;
    form.audience = detail.audience;
    form.tagsText = detail.tags.join(", ");
    form.coverUrl = detail.coverUrl;
    form.content = detail.content;
    form.sourceName = detail.sourceName;
    form.sourceUrl = detail.sourceUrl;
    form.featured = detail.featured;
    form.status = detail.status;
    currentStatus.value = detail.status;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载文章详情失败");
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
      summary: form.summary,
      category: form.category,
      audience: form.audience,
      tags: form.tagsText
        .split(",")
        .map((item) => item.trim())
        .filter(Boolean),
      cover_url: form.coverUrl,
      content: form.content,
      source_name: form.sourceName,
      source_url: form.sourceUrl,
      featured: form.featured,
      status: form.status
    };

    if (articleId.value) {
      await updateArticle(articleId.value, payload);
    } else {
      await createArticle(payload);
    }

    ElMessage.success(`${pageTitle.value}已保存为 ${form.status}`);
    await router.push({ name: "articles" });
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "保存文章失败");
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
        <p>当前为 Phase 1 表单骨架，字段结构已与设计文档对齐。</p>
      </div>
      <div class="actions">
        <el-button @click="router.push({ name: 'articles' })">返回列表</el-button>
        <el-button :loading="saving" @click="handleSave('draft')">保存草稿</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave('published')">发布文章</el-button>
        <el-button
          v-if="articleId && availableStatusOptions.includes('archived')"
          type="warning"
          :loading="saving"
          @click="handleSave('archived')"
        >
          归档文章
        </el-button>
      </div>
    </div>

    <el-card shadow="never" v-loading="loading">
      <el-form label-position="top" class="form-grid">
        <el-form-item label="标题">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="Slug">
          <el-input v-model="form.slug" />
        </el-form-item>
        <el-form-item label="分类">
          <el-input v-model="form.category" />
        </el-form-item>
        <el-form-item label="受众">
          <el-input v-model="form.audience" />
        </el-form-item>
        <el-form-item label="摘要" class="full">
          <el-input v-model="form.summary" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="标签（逗号分隔）" class="full">
          <el-input v-model="form.tagsText" />
        </el-form-item>
        <el-form-item label="封面资源" class="full">
          <MediaPicker v-model="form.coverUrl" title="文章封面" />
        </el-form-item>
        <el-form-item label="正文" class="full">
          <el-input v-model="form.content" type="textarea" :rows="10" />
        </el-form-item>
        <el-form-item label="来源名称">
          <el-input v-model="form.sourceName" />
        </el-form-item>
        <el-form-item label="来源链接">
          <el-input v-model="form.sourceUrl" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option v-for="item in availableStatusOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="是否精选">
          <el-switch v-model="form.featured" />
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
</style>
