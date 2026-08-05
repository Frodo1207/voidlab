<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { useAuthStore } from "../../stores/auth";

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const form = reactive({
  username: "admin",
  password: "admin"
});

const loading = ref(false);

async function handleSubmit() {
  loading.value = true;

  try {
    await authStore.login(form.username, form.password);
    ElMessage.success("登录成功，已进入后台骨架");
    const redirect = typeof route.query.redirect === "string" ? route.query.redirect : "/";
    await router.push(redirect);
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : "登录失败");
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-panel">
      <div class="eyebrow">VOIDLAB.AI / ADMIN</div>
      <h1>Phase 3 运营后台</h1>
      <p>
        这一版已经进入基础权限控制阶段。不同角色会看到不同菜单，并且只能访问各自负责的运营模块。
      </p>

      <div class="account-tips">
        <div><strong>管理员</strong> <code>admin / admin</code></div>
        <div><strong>内容编辑</strong> <code>editor / editor</code></div>
        <div><strong>运营</strong> <code>ops / ops</code></div>
      </div>

      <el-form label-position="top" class="login-form" @submit.prevent="handleSubmit">
        <el-form-item label="账号">
          <el-input v-model="form.username" placeholder="请输入账号" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" show-password placeholder="请输入密码" />
        </el-form-item>
        <el-button type="primary" :loading="loading" class="login-button" @click="handleSubmit">
          进入后台
        </el-button>
      </el-form>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  background:
    radial-gradient(circle at top, rgba(66, 255, 209, 0.14), transparent 35%),
    #09111b;
}

.login-panel {
  width: min(520px, 100%);
  padding: 36px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 24px 80px rgba(0, 0, 0, 0.22);
}

.eyebrow {
  color: #0f8f73;
  font-size: 12px;
  letter-spacing: 0.16em;
}

h1 {
  margin: 12px 0 12px;
  font-size: 36px;
  color: #111827;
}

p {
  margin: 0 0 28px;
  line-height: 1.8;
  color: rgba(17, 24, 39, 0.62);
}

.account-tips {
  display: grid;
  gap: 10px;
  padding: 14px 16px;
  margin-bottom: 24px;
  border-radius: 16px;
  background: rgba(15, 23, 35, 0.04);
  color: rgba(17, 24, 39, 0.72);
  font-size: 13px;
}

.account-tips code {
  font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
  font-size: 12px;
}

.login-form {
  margin-top: 12px;
}

.login-button {
  width: 100%;
  height: 44px;
  margin-top: 8px;
}
</style>
