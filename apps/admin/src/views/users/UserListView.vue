<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { getRoleLabel } from "../../permissions";
import { useAuthStore } from "../../stores/auth";
import type { ManagedUserRecord, UserRole } from "../../types";
import {
  createUser,
  listUsers,
  resetUserPassword,
  updateUserRole,
  updateUserStatus
} from "../../services/userService";

const authStore = useAuthStore();
const records = ref<ManagedUserRecord[]>([]);
const loading = ref(false);
const creating = ref(false);
const resetLoading = ref(false);
const createDialogVisible = ref(false);
const resetDialogVisible = ref(false);
const resettingUser = ref<ManagedUserRecord | null>(null);

const createForm = reactive({
  username: "",
  password: "",
  role: "editor" as UserRole
});

const resetForm = reactive({
  password: ""
});

const roleOptions: UserRole[] = ["admin", "editor", "ops"];

const stats = computed(() => ({
  total: records.value.length,
  active: records.value.filter((item) => item.isActive).length,
  admins: records.value.filter((item) => item.role === "admin").length,
  inactive: records.value.filter((item) => !item.isActive).length
}));

function isCurrentUser(user: ManagedUserRecord) {
  return user.id === authStore.user?.id;
}

async function loadRecords() {
  loading.value = true;
  try {
    records.value = await listUsers();
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "加载用户列表失败");
  } finally {
    loading.value = false;
  }
}

async function handleCreateUser() {
  creating.value = true;
  try {
    const user = await createUser(createForm);
    records.value = [...records.value, user];
    createDialogVisible.value = false;
    createForm.username = "";
    createForm.password = "";
    createForm.role = "editor";
    ElMessage.success("用户已创建");
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "创建用户失败");
  } finally {
    creating.value = false;
  }
}

async function handleRoleChange(user: ManagedUserRecord, role: UserRole) {
  try {
    const updated = await updateUserRole(user.id, role);
    replaceUser(updated);
    ElMessage.success("用户角色已更新");
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "更新角色失败");
    await loadRecords();
  }
}

async function handleToggleStatus(user: ManagedUserRecord, value: boolean) {
  try {
    const updated = await updateUserStatus(user.id, value);
    replaceUser(updated);
    ElMessage.success(value ? "用户已启用" : "用户已停用");
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "更新状态失败");
    await loadRecords();
  }
}

function openResetDialog(user: ManagedUserRecord) {
  resettingUser.value = user;
  resetForm.password = "";
  resetDialogVisible.value = true;
}

async function handleResetPassword() {
  if (!resettingUser.value) {
    return;
  }

  resetLoading.value = true;
  try {
    await resetUserPassword(resettingUser.value.id, resetForm.password);
    resetDialogVisible.value = false;
    resetForm.password = "";
    ElMessage.success("密码已重置");
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "重置密码失败");
  } finally {
    resetLoading.value = false;
  }
}

function replaceUser(nextUser: ManagedUserRecord) {
  records.value = records.value.map((item) => (item.id === nextUser.id ? nextUser : item));
}

async function confirmToggleStatus(user: ManagedUserRecord, nextValue: boolean) {
  try {
    await ElMessageBox.confirm(
      nextValue ? `确认启用用户 ${user.username} 吗？` : `确认停用用户 ${user.username} 吗？`,
      nextValue ? "启用用户" : "停用用户",
      {
        type: nextValue ? "info" : "warning"
      }
    );
    await handleToggleStatus(user, nextValue);
  } catch {
    await loadRecords();
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
        <h2>用户管理</h2>
        <p>管理后台账号、角色和启停用状态，让权限系统真正进入可运营阶段。</p>
      </div>
      <div class="toolbar-actions">
        <el-button @click="loadRecords" :loading="loading">刷新列表</el-button>
        <el-button type="primary" @click="createDialogVisible = true">新增用户</el-button>
      </div>
    </div>

    <div class="stats-grid">
      <el-card shadow="never">
        <div class="stat-label">用户总数</div>
        <div class="stat-value">{{ stats.total }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">启用中</div>
        <div class="stat-value">{{ stats.active }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">管理员</div>
        <div class="stat-value">{{ stats.admins }}</div>
      </el-card>
      <el-card shadow="never">
        <div class="stat-label">已停用</div>
        <div class="stat-value">{{ stats.inactive }}</div>
      </el-card>
    </div>

    <el-card shadow="never">
      <el-table :data="records" v-loading="loading">
        <el-table-column prop="username" label="账号" min-width="150" />
        <el-table-column prop="displayName" label="角色名" min-width="150" />
        <el-table-column label="系统角色" min-width="160">
          <template #default="{ row }">
            <el-select
              :model-value="row.role"
              size="small"
              :disabled="isCurrentUser(row)"
              @update:model-value="handleRoleChange(row, $event)"
            >
              <el-option v-for="role in roleOptions" :key="role" :label="getRoleLabel(role)" :value="role" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="140">
          <template #default="{ row }">
            <el-tag :type="row.isActive ? 'success' : 'info'">
              {{ row.isActive ? "启用中" : "已停用" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" min-width="160" />
        <el-table-column prop="updatedAt" label="更新时间" min-width="160" />
        <el-table-column label="操作" min-width="220" fixed="right">
          <template #default="{ row }">
            <div class="row-actions">
              <el-switch
                :model-value="row.isActive"
                :disabled="isCurrentUser(row)"
                inline-prompt
                active-text="启用"
                inactive-text="停用"
                @change="confirmToggleStatus(row, $event)"
              />
              <el-button link type="primary" @click="openResetDialog(row)">重置密码</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="createDialogVisible" title="新增后台用户" width="520px">
      <el-form label-position="top">
        <el-form-item label="用户名">
          <el-input v-model="createForm.username" placeholder="例如 partner-editor" />
        </el-form-item>
        <el-form-item label="初始密码">
          <el-input v-model="createForm.password" type="password" show-password placeholder="至少 4 位" />
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="createForm.role" class="full-width">
            <el-option v-for="role in roleOptions" :key="role" :label="getRoleLabel(role)" :value="role" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreateUser">创建用户</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="resetDialogVisible" title="重置用户密码" width="460px">
      <p class="dialog-copy">将为 <strong>{{ resettingUser?.username }}</strong> 设置新密码。</p>
      <el-form label-position="top">
        <el-form-item label="新密码">
          <el-input v-model="resetForm.password" type="password" show-password placeholder="至少 4 位" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="resetDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="resetLoading" @click="handleResetPassword">确认重置</el-button>
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
.toolbar-actions,
.row-actions {
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

.full-width {
  width: 100%;
}

.dialog-copy {
  margin: 0 0 8px;
  color: rgba(17, 24, 39, 0.65);
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
