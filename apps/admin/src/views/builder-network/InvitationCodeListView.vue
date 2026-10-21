<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import type { NetworkInviteRecord } from "../../types";
import {
  createNetworkInvite,
  listNetworkInvites,
  updateNetworkInviteStatus
} from "../../services/builderNetworkInviteService";

const records = ref<NetworkInviteRecord[]>([]);
const loading = ref(false);
const creating = ref(false);
const createDialogVisible = ref(false);
const revealDialogVisible = ref(false);
const createdCode = ref("");
const createForm = reactive({
  maxUses: 1,
  expiresAt: null as Date | null
});

const stats = computed(() => ({
  total: records.value.length,
  active: records.value.filter((item) => item.status === "active" && !isExpired(item)).length,
  used: records.value.reduce((total, item) => total + item.usedCount, 0),
  disabled: records.value.filter((item) => item.status === "disabled").length
}));

async function loadRecords() {
  loading.value = true;
  try {
    records.value = await listNetworkInvites();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载邀请码失败");
  } finally {
    loading.value = false;
  }
}

async function handleCreate() {
  creating.value = true;
  try {
    const result = await createNetworkInvite({
      maxUses: createForm.maxUses,
      expiresAt: createForm.expiresAt?.toISOString() ?? ""
    });
    createdCode.value = result.code;
    createDialogVisible.value = false;
    revealDialogVisible.value = true;
    createForm.maxUses = 1;
    createForm.expiresAt = null;
    await loadRecords();
    ElMessage.success("邀请码已创建");
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "创建邀请码失败");
  } finally {
    creating.value = false;
  }
}

async function toggleStatus(record: NetworkInviteRecord) {
  const nextStatus = record.status === "active" ? "disabled" : "active";
  const action = nextStatus === "disabled" ? "停用" : "启用";
  try {
    await ElMessageBox.confirm(
      `确认${action}该邀请码吗？${nextStatus === "disabled" ? "停用后，未使用次数也不能继续注册。" : ""}`,
      `${action}邀请码`,
      { type: nextStatus === "disabled" ? "warning" : "info" }
    );
    await updateNetworkInviteStatus(record.id, nextStatus);
    records.value = records.value.map((item) =>
      item.id === record.id ? { ...item, status: nextStatus } : item
    );
    ElMessage.success(`邀请码已${action}`);
  } catch (error) {
    if (error instanceof Error) {
      ElMessage.error(error.message);
    }
  }
}

async function copyCode() {
  try {
    await navigator.clipboard.writeText(createdCode.value);
    ElMessage.success("邀请码已复制");
  } catch {
    ElMessage.error("复制失败，请手动复制");
  }
}

function isExpired(record: NetworkInviteRecord): boolean {
  return Boolean(record.expiresAt && new Date(record.expiresAt.replace(" ", "T") + "Z").getTime() < Date.now());
}

function inviteStatusLabel(record: NetworkInviteRecord): string {
  if (record.status === "disabled") return "已停用";
  if (isExpired(record)) return "已过期";
  if (record.usedCount >= record.maxUses) return "已用完";
  return "可使用";
}

function inviteStatusType(record: NetworkInviteRecord): "success" | "info" | "warning" {
  if (record.status === "disabled") return "info";
  if (isExpired(record) || record.usedCount >= record.maxUses) return "warning";
  return "success";
}

function formatDate(value: string): string {
  if (!value) return "永不过期";
  return value.replace("T", " ").replace(/\.\d+Z$/, "").replace(/Z$/, " UTC");
}

onMounted(() => {
  void loadRecords();
});
</script>

<template>
  <div class="view">
    <div class="toolbar">
      <div>
        <h2>BuilderNetwork 邀请码</h2>
        <p>管理邀请制注册凭证。邀请码明文只在创建成功后展示一次，列表不保存也不回显明文。</p>
      </div>
      <div class="toolbar-actions">
        <el-button :loading="loading" @click="loadRecords">刷新列表</el-button>
        <el-button type="primary" @click="createDialogVisible = true">创建邀请码</el-button>
      </div>
    </div>

    <div class="stats-grid">
      <el-card shadow="never">
        <div class="stat-label">邀请码总数</div>
        <div class="stat-value">{{ stats.total }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">当前可用</div>
        <div class="stat-value">{{ stats.active }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">累计注册使用</div>
        <div class="stat-value">{{ stats.used }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">已停用</div>
        <div class="stat-value">{{ stats.disabled }}</div>
      </el-card>
    </div>

    <el-card shadow="never">
      <el-table :data="records" v-loading="loading" empty-text="暂无邀请码">
        <el-table-column prop="id" label="编号" width="90" />
        <el-table-column label="使用情况" min-width="160">
          <template #default="{ row }">
            <span>{{ row.usedCount }} / {{ row.maxUses }}</span>
            <el-progress
              :percentage="Math.min(100, Math.round((row.usedCount / row.maxUses) * 100))"
              :show-text="false"
              :stroke-width="6"
            />
          </template>
        </el-table-column>
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="inviteStatusType(row)">{{ inviteStatusLabel(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="有效期" min-width="180">
          <template #default="{ row }">{{ formatDate(row.expiresAt) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="180">
          <template #default="{ row }">{{ formatDate(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="邀请码明文" min-width="150">
          <template #default>
            <span class="muted">仅创建时展示</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              link
              :type="row.status === 'active' ? 'danger' : 'primary'"
              @click="toggleStatus(row)"
            >
              {{ row.status === "active" ? "停用" : "启用" }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="createDialogVisible" title="创建邀请码" width="520px">
      <el-form label-position="top">
        <el-form-item label="最大注册次数">
          <el-input-number v-model="createForm.maxUses" :min="1" :max="1000" />
          <div class="field-hint">同一个邀请码最多可完成 {{ createForm.maxUses }} 次注册。</div>
        </el-form-item>
        <el-form-item label="过期时间（可选）">
          <el-date-picker
            v-model="createForm.expiresAt"
            type="datetime"
            placeholder="不设置则永不过期"
            format="YYYY-MM-DD HH:mm"
            class="full-width"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">创建并查看邀请码</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="revealDialogVisible" title="邀请码已生成" width="560px" :close-on-click-modal="false">
      <p class="dialog-copy">请立即复制并发送给受邀用户。关闭后无法再次查看这段明文。</p>
      <el-input :model-value="createdCode" readonly />
      <template #footer>
        <el-button @click="copyCode">复制邀请码</el-button>
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

.stat-label,
.muted,
.field-hint,
.dialog-copy {
  color: rgba(17, 24, 39, 0.58);
  font-size: 13px;
}

.stat-value {
  margin-top: 8px;
  color: #111827;
  font-size: 32px;
  font-weight: 700;
}

.field-hint {
  margin-top: 8px;
}

.dialog-copy {
  margin: 0 0 16px;
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

@media (max-width: 900px) {
  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .toolbar {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
