<script setup lang="ts">
import { computed, reactive, ref, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import type { AgentScope, AgentTokenRecord } from "../../types";
import { createAgentToken, listAgentTokens, updateAgentTokenStatus } from "../../services/agentTokenService";

const records = ref<AgentTokenRecord[]>([]);
const loading = ref(false);
const creating = ref(false);
const createDialogVisible = ref(false);
const revealDialogVisible = ref(false);
const createdTokenValue = ref("");

const createForm = reactive<{
  name: string;
  scopes: AgentScope[];
}>({
  name: "",
  scopes: ["articles:write"]
});

const scopeOptions: { value: AgentScope; label: string; description: string }[] = [
  { value: "articles:read", label: "文章读取", description: "读取文章列表和详情" },
  { value: "articles:write", label: "文章写入", description: "创建、更新、发布文章" },
  { value: "events:read", label: "活动读取", description: "读取活动列表和详情" },
  { value: "events:write", label: "活动写入", description: "创建、更新活动" },
  { value: "builders:read", label: "Builder 读取", description: "读取 Builder 列表和详情" },
  { value: "builders:write", label: "Builder 写入", description: "创建、更新 Builder" },
  { value: "knowledge:read", label: "知识库读取", description: "读取知识空间、目录与文档" },
  { value: "knowledge:write", label: "知识库写入", description: "创建、更新知识空间、文档和资源" },
  { value: "knowledge_tokens:read", label: "知识令牌读取", description: "读取知识库 Space 访问令牌列表" },
  { value: "knowledge_tokens:write", label: "知识令牌写入", description: "创建、启停知识库 Space 访问令牌" },
  { value: "media:read", label: "媒体读取", description: "读取媒体资源库" },
  { value: "media:write", label: "媒体上传", description: "上传媒体资源" }
];

const stats = computed(() => ({
  total: records.value.length,
  active: records.value.filter((item) => item.isActive).length,
  inactive: records.value.filter((item) => !item.isActive).length,
  used: records.value.filter((item) => item.lastUsedAt).length
}));

async function loadRecords() {
  loading.value = true;
  try {
    records.value = await listAgentTokens();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载 Agent Tokens 失败");
  } finally {
    loading.value = false;
  }
}

async function handleCreate() {
  creating.value = true;
  try {
    const result = await createAgentToken(createForm);
    records.value = [result.record, ...records.value];
    createdTokenValue.value = result.token;
    createDialogVisible.value = false;
    revealDialogVisible.value = true;
    createForm.name = "";
    createForm.scopes = ["articles:write"];
    ElMessage.success("Agent Token 已创建");
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "创建 Agent Token 失败");
  } finally {
    creating.value = false;
  }
}

async function confirmToggle(record: AgentTokenRecord, nextValue: boolean) {
  try {
    await ElMessageBox.confirm(
      nextValue ? `确认启用 ${record.name} 吗？` : `确认停用 ${record.name} 吗？`,
      nextValue ? "启用 Token" : "停用 Token",
      { type: nextValue ? "info" : "warning" }
    );
    const updated = await updateAgentTokenStatus(record.id, nextValue);
    records.value = records.value.map((item) => (item.id === updated.id ? updated : item));
    ElMessage.success(nextValue ? "Token 已启用" : "Token 已停用");
  } catch {
    await loadRecords();
  }
}

function scopeLabel(scope: AgentScope) {
  return scopeOptions.find((item) => item.value === scope)?.label ?? scope;
}

async function copyToken() {
  try {
    await navigator.clipboard.writeText(createdTokenValue.value);
    ElMessage.success("Token 已复制");
  } catch {
    ElMessage.error("复制失败，请手动复制");
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
        <h2>Agent Tokens</h2>
        <p>为 Skill 或自动化 Agent 生成受限身份，按 scope 控制它能操作哪些后台资源。</p>
      </div>
      <div class="toolbar-actions">
        <el-button @click="loadRecords" :loading="loading">刷新列表</el-button>
        <el-button type="primary" @click="createDialogVisible = true">新建 Agent Token</el-button>
      </div>
    </div>

    <div class="stats-grid">
      <el-card shadow="never">
        <div class="stat-label">Token 总数</div>
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
        <div class="stat-label">曾使用</div>
        <div class="stat-value">{{ stats.used }}</div>
      </el-card>
    </div>

    <el-card shadow="never">
      <el-table :data="records" v-loading="loading">
        <el-table-column prop="name" label="名称" min-width="180" />
        <el-table-column label="Scopes" min-width="320">
          <template #default="{ row }">
            <div class="scope-list">
              <el-tag v-for="scope in row.scopes" :key="scope" size="small" effect="plain">{{ scopeLabel(scope) }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="120">
          <template #default="{ row }">
            <el-tag :type="row.isActive ? 'success' : 'info'">{{ row.isActive ? "启用中" : "已停用" }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="lastUsedAt" label="最近使用" min-width="150">
          <template #default="{ row }">
            {{ row.lastUsedAt || "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" min-width="150" />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-switch
              :model-value="row.isActive"
              inline-prompt
              active-text="启用"
              inactive-text="停用"
              @change="confirmToggle(row, $event)"
            />
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="createDialogVisible" title="新建 Agent Token" width="620px">
      <el-form label-position="top">
        <el-form-item label="Token 名称">
          <el-input v-model="createForm.name" placeholder="例如 content-agent-prod" />
        </el-form-item>
        <el-form-item label="授权范围">
          <el-select v-model="createForm.scopes" multiple class="full-width" placeholder="选择允许的操作范围">
            <el-option
              v-for="item in scopeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            >
              <div class="scope-option">
                <strong>{{ item.label }}</strong>
                <span>{{ item.description }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">创建 Token</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="revealDialogVisible" title="Token 已生成" width="680px">
      <p class="dialog-copy">这个明文 token 只会显示这一次，建议马上保存到你的 Skill 或环境变量里。</p>
      <el-input :model-value="createdTokenValue" type="textarea" :rows="3" readonly />
      <template #footer>
        <el-button @click="copyToken">复制 Token</el-button>
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

.scope-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.scope-option {
  display: grid;
  gap: 4px;
}

.scope-option span,
.dialog-copy {
  color: rgba(17, 24, 39, 0.6);
  font-size: 12px;
}

.full-width {
  width: 100%;
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
