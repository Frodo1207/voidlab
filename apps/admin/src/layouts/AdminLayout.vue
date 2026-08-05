<script setup lang="ts">
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getRoleLabel, menuItems } from "../permissions";
import { useAuthStore } from "../stores/auth";

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const visibleMenuItems = computed(() =>
  menuItems.filter((item) => authStore.hasAnyRole(item.roles ?? []))
);

const activeMenu = computed(() => {
  if (route.path.startsWith("/users")) return "/users";
  if (route.path.startsWith("/agent-tokens")) return "/agent-tokens";
  if (route.path.startsWith("/articles")) return "/articles";
  if (route.path.startsWith("/events")) return "/events";
  if (route.path.startsWith("/builders")) return "/builders";
  if (route.path.startsWith("/knowledge")) return "/knowledge/spaces";
  if (route.path.startsWith("/leads")) return "/leads";
  if (route.path.startsWith("/site-configs")) return "/site-configs";
  if (route.path.startsWith("/audit-logs")) return "/audit-logs";
  if (route.path.startsWith("/media")) return "/media";
  return "/";
});

function handleLogout() {
  authStore.logout();
  void router.push({ name: "login" });
}
</script>

<template>
  <el-container class="admin-layout">
    <el-aside class="admin-aside" width="240px">
      <div class="brand">
        <div class="brand-mark">VOIDLAB.AI</div>
        <div class="brand-meta">Phase 4A Knowledge Base</div>
      </div>

      <el-menu
        :default-active="activeMenu"
        class="admin-menu"
        router
        background-color="#0f1723"
        text-color="rgba(255,255,255,0.72)"
        active-text-color="#42ffd1"
      >
        <el-menu-item v-for="item in visibleMenuItems" :key="item.index" :index="item.index">
          {{ item.label }}
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="admin-header">
        <div>
          <div class="header-title">{{ route.meta.title ?? "VOIDLAB 后台" }}</div>
          <div class="header-subtitle">内容、知识库、Leads 与站点配置已进入按角色运营阶段</div>
        </div>

        <div class="header-actions">
          <div class="operator">
            <strong>{{ authStore.user?.displayName ?? "Operator" }}</strong>
            <span>{{ getRoleLabel(authStore.user?.role) }}</span>
          </div>
          <el-button plain @click="handleLogout">退出登录</el-button>
        </div>
      </el-header>

      <el-main class="admin-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.admin-layout {
  min-height: 100vh;
  background: #f3f5f9;
}

.admin-aside {
  display: flex;
  flex-direction: column;
  background: #0f1723;
  color: #fff;
  border-right: 1px solid rgba(255, 255, 255, 0.06);
}

.brand {
  padding: 28px 24px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.brand-mark {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.brand-meta {
  margin-top: 8px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.52);
}

.admin-menu {
  flex: 1;
  border-right: none;
}

.admin-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 80px;
  padding: 0 28px;
  background: rgba(255, 255, 255, 0.82);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(15, 23, 35, 0.08);
}

.header-title {
  font-size: 20px;
  font-weight: 700;
  color: #111827;
}

.header-subtitle {
  margin-top: 4px;
  font-size: 13px;
  color: rgba(17, 24, 39, 0.55);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.operator {
  display: grid;
  gap: 4px;
  text-align: right;
}

.operator span {
  font-size: 12px;
  text-transform: uppercase;
  color: rgba(17, 24, 39, 0.55);
}

.admin-main {
  padding: 28px;
}
</style>
