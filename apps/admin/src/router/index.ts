import { createRouter, createWebHistory } from "vue-router";
import AdminLayout from "../layouts/AdminLayout.vue";
import LoginView from "../views/login/LoginView.vue";
import DashboardView from "../views/dashboard/DashboardView.vue";
import UserListView from "../views/users/UserListView.vue";
import AgentTokenListView from "../views/agent-tokens/AgentTokenListView.vue";
import ArticleListView from "../views/articles/ArticleListView.vue";
import ArticleEditorView from "../views/articles/ArticleEditorView.vue";
import EventListView from "../views/events/EventListView.vue";
import EventEditorView from "../views/events/EventEditorView.vue";
import BuilderListView from "../views/builders/BuilderListView.vue";
import BuilderEditorView from "../views/builders/BuilderEditorView.vue";
import KnowledgeSpaceListView from "../views/knowledge/KnowledgeSpaceListView.vue";
import KnowledgeSpaceEditorView from "../views/knowledge/KnowledgeSpaceEditorView.vue";
import KnowledgeEntryListView from "../views/knowledge/KnowledgeEntryListView.vue";
import KnowledgeEntryEditorView from "../views/knowledge/KnowledgeEntryEditorView.vue";
import KnowledgeAccessTokenListView from "../views/knowledge/KnowledgeAccessTokenListView.vue";
import MediaLibraryView from "../views/media/MediaLibraryView.vue";
import LeadListView from "../views/leads/LeadListView.vue";
import LeadDetailView from "../views/leads/LeadDetailView.vue";
import SiteConfigView from "../views/site-configs/SiteConfigView.vue";
import AuditLogListView from "../views/audit-logs/AuditLogListView.vue";
import { hasRoleAccess } from "../permissions";
import { useAuthStore } from "../stores/auth";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/login",
      name: "login",
      component: LoginView,
      meta: { public: true, title: "后台登录" }
    },
    {
      path: "/",
      component: AdminLayout,
      children: [
        {
          path: "",
          name: "dashboard",
          component: DashboardView,
          meta: { title: "仪表盘", roles: ["admin", "editor", "ops"] }
        },
        {
          path: "users",
          name: "users",
          component: UserListView,
          meta: { title: "用户管理", roles: ["admin"] }
        },
        {
          path: "agent-tokens",
          name: "agent-tokens",
          component: AgentTokenListView,
          meta: { title: "Agent Tokens", roles: ["admin"] }
        },
        {
          path: "articles",
          name: "articles",
          component: ArticleListView,
          meta: { title: "资讯管理", roles: ["admin", "editor"] }
        },
        {
          path: "articles/new",
          name: "article-create",
          component: ArticleEditorView,
          meta: { title: "新建文章", roles: ["admin", "editor"] }
        },
        {
          path: "articles/:id/edit",
          name: "article-edit",
          component: ArticleEditorView,
          props: true,
          meta: { title: "编辑文章", roles: ["admin", "editor"] }
        },
        {
          path: "events",
          name: "events",
          component: EventListView,
          meta: { title: "活动管理", roles: ["admin", "editor"] }
        },
        {
          path: "events/new",
          name: "event-create",
          component: EventEditorView,
          meta: { title: "新建活动", roles: ["admin", "editor"] }
        },
        {
          path: "events/:id/edit",
          name: "event-edit",
          component: EventEditorView,
          props: true,
          meta: { title: "编辑活动", roles: ["admin", "editor"] }
        },
        {
          path: "builders",
          name: "builders",
          component: BuilderListView,
          meta: { title: "Builder 管理", roles: ["admin", "editor"] }
        },
        {
          path: "builders/new",
          name: "builder-create",
          component: BuilderEditorView,
          meta: { title: "新建 Builder", roles: ["admin", "editor"] }
        },
        {
          path: "builders/:id/edit",
          name: "builder-edit",
          component: BuilderEditorView,
          props: true,
          meta: { title: "编辑 Builder", roles: ["admin", "editor"] }
        },
        {
          path: "knowledge/spaces",
          name: "knowledge-spaces",
          component: KnowledgeSpaceListView,
          meta: { title: "知识库 Space", roles: ["admin", "editor"] }
        },
        {
          path: "knowledge/spaces/new",
          name: "knowledge-space-create",
          component: KnowledgeSpaceEditorView,
          meta: { title: "新建知识空间", roles: ["admin", "editor"] }
        },
        {
          path: "knowledge/spaces/:id/edit",
          name: "knowledge-space-edit",
          component: KnowledgeSpaceEditorView,
          props: true,
          meta: { title: "编辑知识空间", roles: ["admin", "editor"] }
        },
        {
          path: "knowledge/entries",
          name: "knowledge-entries",
          component: KnowledgeEntryListView,
          meta: { title: "知识文档 Entry", roles: ["admin", "editor"] }
        },
        {
          path: "knowledge/entries/new",
          name: "knowledge-entry-create",
          component: KnowledgeEntryEditorView,
          meta: { title: "新建知识文档", roles: ["admin", "editor"] }
        },
        {
          path: "knowledge/entries/:id/edit",
          name: "knowledge-entry-edit",
          component: KnowledgeEntryEditorView,
          props: true,
          meta: { title: "编辑知识文档", roles: ["admin", "editor"] }
        },
        {
          path: "knowledge/access-tokens",
          name: "knowledge-access-tokens",
          component: KnowledgeAccessTokenListView,
          meta: { title: "知识库访问令牌", roles: ["admin"] }
        },
        {
          path: "media",
          name: "media",
          component: MediaLibraryView,
          meta: { title: "媒体资源", roles: ["admin", "editor"] }
        },
        {
          path: "leads",
          name: "leads",
          component: LeadListView,
          meta: { title: "Leads 管理", roles: ["admin", "ops"] }
        },
        {
          path: "leads/:id",
          name: "lead-detail",
          component: LeadDetailView,
          props: true,
          meta: { title: "Lead 详情", roles: ["admin", "ops"] }
        },
        {
          path: "site-configs",
          name: "site-configs",
          component: SiteConfigView,
          meta: { title: "站点配置", roles: ["admin"] }
        },
        {
          path: "audit-logs",
          name: "audit-logs",
          component: AuditLogListView,
          meta: { title: "审计日志", roles: ["admin"] }
        }
      ]
    }
  ]
});

router.beforeEach((to) => {
  const authStore = useAuthStore();

  if (to.meta.public) {
    if (authStore.isAuthenticated && to.name === "login") {
      return { name: "dashboard" };
    }

    return true;
  }

  if (!authStore.isAuthenticated) {
    return {
      name: "login",
      query: { redirect: to.fullPath }
    };
  }

  if (!hasRoleAccess(authStore.role, to.meta.roles)) {
    return { name: "dashboard" };
  }

  return true;
});

export default router;
