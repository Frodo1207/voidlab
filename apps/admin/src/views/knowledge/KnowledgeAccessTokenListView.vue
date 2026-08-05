<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  createKnowledgeAccessToken,
  listKnowledgeAccessTokens,
  listKnowledgeSpaces,
  updateKnowledgeAccessTokenStatus
} from "../../services/knowledgeService";
import type { KnowledgeAccessTokenRecord, KnowledgeSpaceRecord } from "../../types";

const route = useRoute();
const router = useRouter();

const records = ref<KnowledgeAccessTokenRecord[]>([]);
const spaces = ref<KnowledgeSpaceRecord[]>([]);
const loading = ref(false);
const creating = ref(false);
const createDialogVisible = ref(false);
const revealDialogVisible = ref(false);
const createdTokenValue = ref("");

const createForm = reactive({
  accessLevel: "basic" as "basic" | "pro" | "vip",
  spaceIds: [] as number[],
  name: "",
  expiresAt: ""
});

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

const stats = computed(() => ({
  total: records.value.length,
  active: records.value.filter((item) => item.isActive).length,
  inactive: records.value.filter((item) => !item.isActive).length,
  permanent: records.value.filter((item) => !item.expiresAt).length,
  vip: records.value.filter((item) => item.accessLevel === "vip").length
}));

function statusTagType(isActive: boolean) {
  return isActive ? "success" : "info";
}

function accessLevelLabel(level: KnowledgeAccessTokenRecord["accessLevel"]) {
  switch (level) {
    case "basic":
      return "基础令牌";
    case "pro":
      return "专题包令牌";
    case "vip":
      return "全局令牌";
    default:
      return level;
  }
}

function scopeSummary(record: KnowledgeAccessTokenRecord) {
  if (record.scopeType === "all_published") {
    return "全部已发布 Space";
  }

  const labels = record.spaceIds
    .map((spaceId) => spaces.value.find((item) => item.id === spaceId)?.title ?? `Space #${spaceId}`)
    .join(" / ");

  return labels || "未绑定 Space";
}

async function loadSpaces() {
  spaces.value = await listKnowledgeSpaces();
}

async function loadRecords() {
  loading.value = true;
  try {
    records.value = await listKnowledgeAccessTokens(selectedSpaceId.value);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载知识库访问令牌失败");
  } finally {
    loading.value = false;
  }
}

async function loadData() {
  try {
    await loadSpaces();
    await loadRecords();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "初始化令牌管理页失败");
  }
}

function openCreateDialog() {
  createForm.accessLevel = selectedSpaceId.value ? "basic" : "basic";
  createForm.spaceIds = selectedSpaceId.value ? [selectedSpaceId.value] : [];
  createForm.name = "";
  createForm.expiresAt = "";
  createDialogVisible.value = true;
}

async function handleCreate() {
  creating.value = true;
  try {
    const normalizedSpaceIds = Array.from(new Set(createForm.spaceIds.filter((item) => item > 0)));
    const result = await createKnowledgeAccessToken({
      space_id: normalizedSpaceIds[0],
      space_ids: normalizedSpaceIds,
      name: createForm.name,
      access_level: createForm.accessLevel,
      expires_at: createForm.expiresAt
    });
    records.value = [result.record, ...records.value];
    createdTokenValue.value = result.token;
    createDialogVisible.value = false;
    revealDialogVisible.value = true;
    ElMessage.success("知识库访问令牌已创建");
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "创建知识库访问令牌失败");
  } finally {
    creating.value = false;
  }
}

async function handleToggle(record: KnowledgeAccessTokenRecord, nextValue: boolean) {
  try {
    await ElMessageBox.confirm(
      nextValue ? `确认启用 ${record.name} 吗？` : `确认停用 ${record.name} 吗？`,
      nextValue ? "启用令牌" : "停用令牌",
      { type: nextValue ? "info" : "warning" }
    );
    const updated = await updateKnowledgeAccessTokenStatus(record.id, nextValue);
    records.value = records.value.map((item) => (item.id === updated.id ? updated : item));
    ElMessage.success(nextValue ? "令牌已启用" : "令牌已停用");
  } catch {
    await loadRecords();
  }
}

function handleSpaceChange(spaceId: number | undefined) {
  void router.replace({
    name: "knowledge-access-tokens",
    query: spaceId ? { spaceId: String(spaceId) } : {}
  });
}

async function copyToken() {
  try {
    await navigator.clipboard.writeText(createdTokenValue.value);
    ElMessage.success("令牌已复制");
  } catch {
    ElMessage.error("复制失败，请手动复制");
  }
}

watch(
  () => route.query.spaceId,
  () => {
    void loadRecords();
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
        <h2>知识库访问令牌</h2>
        <p>当前筛选：{{ selectedSpaceLabel }}。这里管理基础令牌、专题包令牌和全局令牌的启停状态。</p>
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
        <el-button @click="loadRecords" :loading="loading">刷新</el-button>
        <el-button type="primary" @click="openCreateDialog">新建令牌</el-button>
      </div>
    </div>

    <div class="stats-grid">
      <el-card shadow="never">
        <div class="stat-label">令牌总数</div>
        <div class="stat-value">{{ stats.total }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">启用中</div>
        <div class="stat-value">{{ stats.active }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">已停用</div>
        <div class="stat-value">{{ stats.inactive }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">永久令牌</div>
        <div class="stat-value">{{ stats.permanent }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">全局令牌</div>
        <div class="stat-value">{{ stats.vip }}</div>
      </el-card>
    </div>

    <el-card shadow="never">
      <el-table :data="records" v-loading="loading">
        <el-table-column label="令牌等级" min-width="150">
          <template #default="{ row }">
            <el-tag :type="row.accessLevel === 'vip' ? 'danger' : row.accessLevel === 'pro' ? 'warning' : 'success'">
              {{ accessLevelLabel(row.accessLevel) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="解锁范围" min-width="260">
          <template #default="{ row }">
            {{ scopeSummary(row) }}
          </template>
        </el-table-column>
        <el-table-column prop="name" label="令牌名称" min-width="220" />
        <el-table-column label="状态" min-width="120">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.isActive)">
              {{ row.isActive ? "启用中" : "已停用" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="过期时间" min-width="180">
          <template #default="{ row }">
            {{ row.expiresAt || "永不过期" }}
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" min-width="160" />
        <el-table-column prop="updatedAt" label="更新时间" min-width="160" />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-switch
              :model-value="row.isActive"
              inline-prompt
              active-text="启用"
              inactive-text="停用"
              @change="handleToggle(row, $event)"
            />
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="createDialogVisible" title="新建知识库访问令牌" width="620px">
      <el-form label-position="top">
        <el-form-item label="令牌等级">
          <el-radio-group v-model="createForm.accessLevel">
            <el-radio-button label="basic">基础</el-radio-button>
            <el-radio-button label="pro">专题包</el-radio-button>
            <el-radio-button label="vip">全局</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="createForm.accessLevel !== 'vip'" :label="createForm.accessLevel === 'basic' ? '绑定 1 个 Space' : '绑定多个 Space'">
          <el-select
            v-model="createForm.spaceIds"
            class="full-width"
            multiple
            :multiple-limit="createForm.accessLevel === 'basic' ? 1 : 0"
            collapse-tags
            placeholder="选择要解锁的知识空间"
          >
            <el-option v-for="item in spaces" :key="item.id" :label="item.title" :value="item.id" />
          </el-select>
          <div class="field-help">
            {{ createForm.accessLevel === "basic" ? "基础令牌只能解锁一个 Space。" : "专题包令牌可以覆盖多个指定 Space。" }}
          </div>
        </el-form-item>
        <el-form-item v-else label="全局范围">
          <el-alert type="success" :closable="false" show-icon title="全局令牌会解锁全部已发布知识库 Space。" />
        </el-form-item>
        <el-form-item label="令牌名称">
          <el-input v-model="createForm.name" placeholder="例如 founder-playbook-q4" />
        </el-form-item>
        <el-form-item label="过期时间（可选）">
          <el-input v-model="createForm.expiresAt" placeholder="例如 2026-12-31 23:59 或留空表示长期有效" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">创建令牌</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="revealDialogVisible" title="访问令牌已生成" width="680px">
      <p class="dialog-copy">这个明文令牌只会显示这一次，建议立即保存给需要解锁知识库的用户或流程。</p>
      <el-input :model-value="createdTokenValue" type="textarea" :rows="3" readonly />
      <template #footer>
        <el-button @click="copyToken">复制令牌</el-button>
        <el-button type="primary" @click="revealDialogVisible = false">我已保存</el-button>
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
.toolbar-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar {
  justify-content: space-between;
}

.space-filter {
  width: 240px;
}

.stats-grid {
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.stat-label {
  font-size: 13px;
  color: rgba(17, 24, 39, 0.55);
}

.stat-value {
  margin-top: 8px;
  font-size: 32px;
  font-weight: 700;
  color: #111827;
}

.full-width {
  width: 100%;
}

h2 {
  margin: 0;
  font-size: 28px;
}

p,
.dialog-copy {
  margin: 8px 0 0;
  color: rgba(17, 24, 39, 0.6);
}
</style>
