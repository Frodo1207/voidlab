<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { useRoute, useRouter } from "vue-router";
import MediaPicker from "../../components/MediaPicker.vue";
import { createBuilder, getBuilderById, updateBuilder } from "../../services/builderService";
import type { ContentStatus } from "../../types";

const route = useRoute();
const router = useRouter();

const builderId = computed(() => {
  const raw = route.params.id;
  return raw ? Number(raw) : null;
});

const pageTitle = computed(() => (builderId.value ? "编辑 Builder" : "新建 Builder"));
const loading = ref(false);
const saving = ref(false);
const currentStatus = ref<ContentStatus>("draft");
const contentStatusOptions: Record<ContentStatus, ContentStatus[]> = {
  draft: ["draft", "published"],
  published: ["published", "draft", "archived"],
  archived: ["archived", "draft"]
};

const form = reactive({
  name: "",
  slug: "",
  title: "",
  city: "",
  role: "",
  intro: "",
  story: "",
  expertiseText: "",
  focusAreasText: "",
  collaborationModesText: "",
  coverUrl: "",
  contactable: false,
  featured: false,
  status: "draft" as ContentStatus
});

const availableStatusOptions = computed(() => {
  if (!builderId.value) {
    return ["draft", "published"] satisfies ContentStatus[];
  }

  return contentStatusOptions[currentStatus.value];
});

async function loadDetail() {
  if (!builderId.value) {
    return;
  }

  loading.value = true;
  try {
    const detail = await getBuilderById(builderId.value);
    if (!detail) {
      ElMessage.warning("未找到 Builder，已保留为空白表单");
      return;
    }

    form.name = detail.name;
    form.slug = detail.slug;
    form.title = detail.title;
    form.city = detail.city;
    form.role = detail.role;
    form.intro = detail.intro;
    form.story = detail.story;
    form.expertiseText = detail.expertise.join(", ");
    form.focusAreasText = detail.focusAreas.join(", ");
    form.collaborationModesText = detail.collaborationModes.join(", ");
    form.coverUrl = detail.coverUrl;
    form.contactable = detail.contactable;
    form.featured = detail.featured;
    form.status = detail.status;
    currentStatus.value = detail.status;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载 Builder 详情失败");
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
      name: form.name,
      slug: form.slug,
      title: form.title,
      city: form.city,
      role: form.role,
      intro: form.intro,
      story: form.story,
      expertise: form.expertiseText
        .split(",")
        .map((item) => item.trim())
        .filter(Boolean),
      focus_areas: form.focusAreasText
        .split(",")
        .map((item) => item.trim())
        .filter(Boolean),
      collaboration_modes: form.collaborationModesText
        .split(",")
        .map((item) => item.trim())
        .filter(Boolean),
      cover_url: form.coverUrl,
      contactable: form.contactable,
      featured: form.featured,
      status: form.status
    };

    if (builderId.value) {
      await updateBuilder(builderId.value, payload);
    } else {
      await createBuilder(payload);
    }

    ElMessage.success(`${pageTitle.value}已保存为 ${form.status}`);
    await router.push({ name: "builders" });
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "保存 Builder 失败");
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
        <p>Builder 模块会先统一人物介绍、能力项、合作方式和状态字段。</p>
      </div>
      <div class="actions">
        <el-button @click="router.push({ name: 'builders' })">返回列表</el-button>
        <el-button :loading="saving" @click="handleSave('draft')">保存草稿</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave('published')">发布 Builder</el-button>
        <el-button
          v-if="builderId && availableStatusOptions.includes('archived')"
          type="warning"
          :loading="saving"
          @click="handleSave('archived')"
        >
          归档 Builder
        </el-button>
      </div>
    </div>

    <el-card shadow="never" v-loading="loading">
      <el-form label-position="top" class="form-grid">
        <el-form-item label="姓名">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="Slug">
          <el-input v-model="form.slug" />
        </el-form-item>
        <el-form-item label="头衔">
          <el-input v-model="form.title" />
        </el-form-item>
        <el-form-item label="角色">
          <el-input v-model="form.role" />
        </el-form-item>
        <el-form-item label="城市">
          <el-input v-model="form.city" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option v-for="item in availableStatusOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="简介" class="full">
          <el-input v-model="form.intro" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="故事" class="full">
          <el-input v-model="form.story" type="textarea" :rows="8" />
        </el-form-item>
        <el-form-item label="能力标签" class="full">
          <el-input v-model="form.expertiseText" placeholder="逗号分隔" />
        </el-form-item>
        <el-form-item label="关注方向" class="full">
          <el-input v-model="form.focusAreasText" placeholder="逗号分隔" />
        </el-form-item>
        <el-form-item label="合作方式" class="full">
          <el-input v-model="form.collaborationModesText" placeholder="逗号分隔" />
        </el-form-item>
        <el-form-item label="封面资源" class="full">
          <MediaPicker v-model="form.coverUrl" title="Builder 封面" />
        </el-form-item>
        <el-form-item label="可联系">
          <el-switch v-model="form.contactable" />
        </el-form-item>
        <el-form-item label="Featured">
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
