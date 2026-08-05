<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { useRoute, useRouter } from "vue-router";
import MediaPicker from "../../components/MediaPicker.vue";
import { createEvent, getEventById, updateEvent } from "../../services/eventService";
import type { ContentStatus } from "../../types";

const route = useRoute();
const router = useRouter();

const eventId = computed(() => {
  const raw = route.params.id;
  return raw ? Number(raw) : null;
});

const pageTitle = computed(() => (eventId.value ? "编辑活动" : "新建活动"));
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
  city: "",
  location: "",
  eventType: "",
  eventTime: "",
  coverUrl: "",
  content: "",
  status: "draft" as ContentStatus
});

const availableStatusOptions = computed(() => {
  if (!eventId.value) {
    return ["draft", "published"] satisfies ContentStatus[];
  }

  return contentStatusOptions[currentStatus.value];
});

async function loadDetail() {
  if (!eventId.value) {
    return;
  }

  loading.value = true;
  try {
    const detail = await getEventById(eventId.value);
    if (!detail) {
      ElMessage.warning("未找到活动，已保留为空白表单");
      return;
    }

    form.title = detail.title;
    form.slug = detail.slug;
    form.summary = detail.summary;
    form.city = detail.city;
    form.location = detail.location;
    form.eventType = detail.eventType;
    form.eventTime = detail.eventTime;
    form.coverUrl = detail.coverUrl;
    form.content = detail.content;
    form.status = detail.status;
    currentStatus.value = detail.status;
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载活动详情失败");
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
      city: form.city,
      location: form.location,
      event_type: form.eventType,
      event_time: form.eventTime,
      cover_url: form.coverUrl,
      content: form.content,
      status: form.status
    };

    if (eventId.value) {
      await updateEvent(eventId.value, payload);
    } else {
      await createEvent(payload);
    }

    ElMessage.success(`${pageTitle.value}已保存为 ${form.status}`);
    await router.push({ name: "events" });
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "保存活动失败");
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
        <p>活动模块先把主数据录入和封面、状态、正文结构统一下来。</p>
      </div>
      <div class="actions">
        <el-button @click="router.push({ name: 'events' })">返回列表</el-button>
        <el-button :loading="saving" @click="handleSave('draft')">保存草稿</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave('published')">发布活动</el-button>
        <el-button
          v-if="eventId && availableStatusOptions.includes('archived')"
          type="warning"
          :loading="saving"
          @click="handleSave('archived')"
        >
          归档活动
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
        <el-form-item label="城市">
          <el-input v-model="form.city" />
        </el-form-item>
        <el-form-item label="地点">
          <el-input v-model="form.location" />
        </el-form-item>
        <el-form-item label="活动类型">
          <el-input v-model="form.eventType" />
        </el-form-item>
        <el-form-item label="活动时间">
          <el-input v-model="form.eventTime" placeholder="YYYY-MM-DD HH:mm" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="form.status">
            <el-option v-for="item in availableStatusOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="封面资源" class="full">
          <MediaPicker v-model="form.coverUrl" title="活动封面" />
        </el-form-item>
        <el-form-item label="摘要" class="full">
          <el-input v-model="form.summary" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="活动详情" class="full">
          <el-input v-model="form.content" type="textarea" :rows="10" />
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
