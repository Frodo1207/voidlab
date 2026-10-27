<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import marketLogo from "../../../assets/logo/white-bg.svg";
import { getNetworkToken, networkRequest, projectFromLegacyOpportunity, type NetworkOpportunity, type NetworkProfile, type NetworkProject, type NetworkProjectTiming } from "../composables/useBuilderMarket";

// 模拟 Discussions 数据结构
type DiscussionCategory = "Announcements" | "Q&A" | "Ideas" | "General" | "Show and tell";

interface DiscussionThread {
  id: number;
  category: DiscussionCategory;
  title: string;
  author: { name: string; avatar?: string };
  created_at: string;
  replies_count: number;
  is_solved?: boolean;
  is_pinned?: boolean;
  tags?: string[];
  last_activity: string;
}

const mockDiscussions = ref<DiscussionThread[]>([
  {
    id: 1,
    category: "Announcements",
    title: "Welcome to AI 工作流平台 MVP Discussions!",
    author: { name: "Demo Builder" },
    created_at: "2026-10-20",
    replies_count: 5,
    is_pinned: true,
    last_activity: "2 days ago"
  },
  {
    id: 2,
    category: "Q&A",
    title: "如何配置自定义的采集来源？",
    author: { name: "Debug User" },
    created_at: "2026-10-21",
    replies_count: 2,
    is_solved: true,
    tags: ["feature", "question"],
    last_activity: "1 day ago"
  },
  {
    id: 3,
    category: "Ideas",
    title: "建议增加支持飞书 webhook 分发",
    author: { name: "1" },
    created_at: "2026-10-22",
    replies_count: 0,
    tags: ["enhancement"],
    last_activity: "4 hours ago"
  }
]);

const currentDiscussionCategory = ref<DiscussionCategory | "All">("All");

const filteredDiscussions = computed(() => {
  if (currentDiscussionCategory.value === "All") {
    return mockDiscussions.value;
  }
  return mockDiscussions.value.filter(d => d.category === currentDiscussionCategory.value);
});

const route = useRoute();
const router = useRouter();
const slug = computed(() => typeof route.params.slug === "string" ? route.params.slug : "");

const isAuthed = ref(false);
onMounted(() => {
  isAuthed.value = Boolean(getNetworkToken());
});

const loading = ref(true);
const error = ref("");
const project = ref<NetworkProject | null>(null);
const selectedOpportunityID = ref<number | null>(null);
const teamProfiles = ref<NetworkProfile[]>([]);

const devMockOpportunities: NetworkOpportunity[] = [
  {
    id: 1001,
    is_demo: true,
    profile_id: 1,
    slug: "agent-workflow-platform-mvp",
    title: "AI 工作流平台 MVP：需要一位 Vue 工程化伙伴",
    summary: "我们正在把资讯运营中的内容采集、人工审核、发布与渠道分发，整理成一套可配置、可追踪、可复用的工作流。MVP 阶段会先跑通从来源接入到内容发布的完整链路，并在约 2 周内交付一个可供真实运营场景试用的版本。",
    opportunity_type: "collaboration",
    looking_for_roles: ["前端开发", "产品工程"],
    required_skills: ["Vue 3", "TypeScript", "Vite"],
    engagement_mode: "按阶段合作",
    budget_hint: "预算面议",
    duration_hint: "2–3 周",
    city_scope: "不限",
    remote_allowed: true,
    status: "open",
    expires_at: "2026-12-31",
    owner_name: "林舟",
    team_intro: "我们是一个偏工程的产品小队，目标是在最短路径内把闭环跑通：可用 > 完美。需要你把工程化、组件抽象与交互体验一次性拉到可持续迭代的水位。",
    team_members: [
      {
        profile_id: 1,
        display_name: "林舟",
        primary_role: "产品工程师",
        role: "发起人 / 产品工程",
        contribution: "定义 MVP 边界、产品结构、核心流程与验收标准",
        sort_order: 0
      },
      {
        profile_id: 2,
        display_name: "陈以南",
        primary_role: "产品设计师",
        role: "交互 / 信息架构",
        contribution: "把流程拆成可复用的信息结构，确保关键页面可交付",
        sort_order: 1
      },
      {
        profile_id: 3,
        display_name: "周行知",
        primary_role: "平台工程师",
        role: "工程支持",
        contribution: "参与联调、质量检查与上线验证",
        sort_order: 2
      }
    ],
    project_plan: {
      milestones: [
        {
          title: "Milestone 0：需求冻结与工程底座",
          description: "把 MVP 边界和数据/接口契约冻结下来，建立前端工程基座（路由、状态、表单规范、错误/加载态）。",
          order: 0,
          status: "in_progress",
          timing: { kind: "duration", duration_value: 3, duration_unit: "天", label: "尽快完成以进入开发" },
          goals: ["明确用户故事与验收条件", "确定页面信息架构与导航", "确定 API 字段与错误模型"],
          deliverables: ["页面骨架与路由", "可复用表单/输入规范", "统一的 Loading/Empty/Error UI"],
          acceptance_criteria: ["关键路径可点击通", "表单校验与提交有明确反馈", "请求失败不会导致页面崩溃"],
          owner_member_ids: [1, 2]
        },
        {
          title: "Milestone 1：工作流闭环跑通",
          description: "把“采集-审核-发布-分发”打通最小闭环，优先保证可用，允许后续再补高级能力。",
          order: 1,
          status: "planned",
          timing: { kind: "duration", duration_value: 7, duration_unit: "天", label: "以可用闭环为唯一目标" },
          goals: ["能创建工作流", "能执行一次完整流程", "能看到清晰的状态与结果"],
          deliverables: ["工作流创建/编辑 UI", "执行与状态展示 UI", "基础错误回滚/重试策略"],
          acceptance_criteria: ["闭环至少能稳定跑通 3 次", "异常时能提示用户下一步动作", "关键状态可追踪"],
          owner_member_ids: [1, 3]
        },
        {
          title: "Milestone 2：体验打磨与上线内测",
          description: "补齐空状态、权限与多端适配，完成内测上线。",
          order: 2,
          status: "planned",
          timing: { kind: "flexible", label: "根据内测反馈滚动推进" },
          goals: ["提升可读性和容错", "完成移动端可用", "上线内测版本"],
          deliverables: ["完善的 Empty/Loading/Error", "移动端适配", "内测发布 checklist"],
          acceptance_criteria: ["移动端关键路径可用", "核心页面无明显布局问题", "内测用户能完成一次完整流程"],
          owner_member_ids: [2, 3]
        }
      ],
      sections: [
        {
          type: "context",
          title: "背景",
          body: "目前资讯运营包含来源维护、内容整理、人工校对、发布排期和多渠道分发等连续环节。大量工作依靠人工在不同工具之间切换，流程容易断点，状态也难以追踪。我们希望把这些重复劳动产品化，让团队能够看清每条内容走到哪一步、由谁处理、最终发布到了哪里。",
          items: [
            "采集规则可以按来源配置，来源变化时不需要重做整条流程。",
            "审核过程有明确状态、处理人和结果，待处理事项一眼可见。",
            "发布和分发记录可查询、可回放，便于核对结果和排查问题。"
          ],
          sort_order: 0
        },
        {
          type: "scope",
          title: "范围与边界（MVP）",
          body: "第一阶段聚焦一条真实、完整、可验证的运营链路。先验证流程设计和团队协作方式，再根据内测反馈决定扩展方向。",
          items: [
            "先交付 1 套端到端工作流模板，覆盖采集、审核、发布和分发。",
            "先接入 1 个分发渠道，并保留清楚的执行记录与失败反馈。",
            "保留必要的人工确认节点，不以全自动化作为 MVP 的前置条件。",
            "暂不扩展复杂的多租户、权限矩阵和大量渠道适配。"
          ],
          sort_order: 1
        },
        {
          type: "outcome",
          title: "阶段目标与验收",
          body: "目标是在约 2 周内完成闭环并启动小范围内测。验收重点是运营人员能否独立完成一次流程，而不是功能数量。",
          items: [
            "运营人员可以配置来源并发起一次内容采集。",
            "内容可以经过审核状态流转，并明确显示待处理和处理结果。",
            "审核通过的内容可以发布到首个渠道，并查询执行记录。",
            "遇到失败时能定位所在步骤，并知道如何重试或人工处理。"
          ],
          sort_order: 2
        },
        {
          type: "execution",
          title: "执行方式",
          body: "",
          items: ["前端按组件模块拆分交付", "每个模块先可用再迭代", "关键节点做一次联调演示"],
          sort_order: 3
        }
      ]
    }
  },
  {
    id: 1002,
    is_demo: true,
    profile_id: 2,
    slug: "buildernetwork-discovery-ranking",
    title: "BuilderNetwork 发现页：信息架构与排序策略设计",
    summary: "优化“找项目/找人”的发现体验，输出可解释的排序规则与落地可实现的 UI 方案。",
    opportunity_type: "advisor",
    looking_for_roles: ["产品设计", "策略"],
    required_skills: ["信息架构", "增长实验"],
    engagement_mode: "每周 2 次同步",
    budget_hint: "按周结算",
    duration_hint: "4 周",
    city_scope: "上海 / 远程",
    remote_allowed: true,
    status: "open",
    expires_at: "2026-12-31",
    owner_name: "陈以南"
  },
  {
    id: 1003,
    is_demo: true,
    profile_id: 3,
    slug: "observability-dashboard",
    title: "LLM 应用观测面板：埋点、追踪与成本看板",
    summary: "需要把请求链路、错误率、延迟与 Token 成本做成一页可用的观测面板，用于持续迭代。",
    opportunity_type: "freelance",
    looking_for_roles: ["全栈开发", "数据可视化"],
    required_skills: ["Go", "PostgreSQL", "可视化"],
    engagement_mode: "短期外包",
    budget_hint: "¥15k–30k",
    duration_hint: "1–2 周",
    city_scope: "不限",
    remote_allowed: true,
    status: "open",
    expires_at: "2026-12-31",
    owner_name: "周行知"
  },
  {
    id: 1004,
    is_demo: true,
    profile_id: 4,
    slug: "agent-sandbox-automation",
    title: "Agent 自动化沙盒：把脚本打包成可复用 Skill",
    summary: "希望搭一套“脚本 → 参数化 → 审计 → 发布”的轻量流水线，方便团队复用自动化能力并可追踪效果。",
    opportunity_type: "collaboration",
    looking_for_roles: ["后端开发", "平台工程"],
    required_skills: ["Go", "CLI 工具", "权限模型"],
    engagement_mode: "按模块合作",
    budget_hint: "预算面议",
    duration_hint: "2–4 周",
    city_scope: "不限",
    remote_allowed: true,
    status: "open",
    expires_at: "2026-12-31",
    owner_name: "许白"
  },
  {
    id: 1005,
    is_demo: true,
    profile_id: 5,
    slug: "multimodal-commerce-ops",
    title: "多模态电商运营 Agent：从素材到上架的半自动流水线",
    summary: "把“图片/短视频素材整理 → 文案生成 → 上架表单”串成可回放的工作流，优先能跑通一个德国站点品类。",
    opportunity_type: "freelance",
    looking_for_roles: ["AI 工程", "增长运营"],
    required_skills: ["LLM 应用", "工作流编排", "电商数据"],
    engagement_mode: "短期外包",
    budget_hint: "¥20k–50k",
    duration_hint: "3–5 周",
    city_scope: "远程",
    remote_allowed: true,
    status: "open",
    expires_at: "2026-12-31",
    owner_name: "沈知夏"
  }
];

const devMockProjects: NetworkProject[] = devMockOpportunities.map((opportunity) => ({
  id: opportunity.id + 10000,
  is_demo: true,
  owner_profile_id: opportunity.profile_id,
  slug: opportunity.slug,
  title: opportunity.title,
  summary: opportunity.summary,
  status: "published",
  project_plan: opportunity.project_plan,
  team_intro: opportunity.team_intro || "",
  owner_name: opportunity.owner_name,
  members: opportunity.team_members || [],
  opportunities: [{ ...opportunity, project_id: opportunity.id + 10000 }],
  created_at: opportunity.created_at || "",
  updated_at: opportunity.updated_at || opportunity.created_at || ""
}));

const activeOpportunity = computed(() =>
  project.value?.opportunities.find((item) => item.id === selectedOpportunityID.value)
  || project.value?.opportunities[0]
  || null
);
const projectRoles = computed(() => [...new Set(project.value?.opportunities.flatMap((item) => item.looking_for_roles || []) || [])]);
const projectSkills = computed(() => [...new Set(project.value?.opportunities.flatMap((item) => item.required_skills || []) || [])]);

async function load() {
  loading.value = true;
  error.value = "";
  try {
    let hit: NetworkProject | null = null;
    try {
      hit = await networkRequest<NetworkProject>(`/api/v1/public/network/projects/${encodeURIComponent(slug.value)}`);
    } catch (cause) {
      if (!import.meta.env.DEV) throw cause;
      try {
        const projects = await networkRequest<NetworkProject[]>("/api/v1/public/network/projects");
        const source = [...projects, ...devMockProjects.filter((mock) => !projects.some((item) => item.slug === mock.slug))];
        hit = source.find((item) => item.slug === slug.value) || null;
      } catch {
        const opportunities = await networkRequest<NetworkOpportunity[]>("/api/v1/public/network/opportunities");
        const legacy = opportunities.find((item) => item.slug === slug.value);
        hit = legacy ? projectFromLegacyOpportunity(legacy) : devMockProjects.find((item) => item.slug === slug.value) || null;
      }
    }
    if (!hit) {
      await router.replace("/market/projects");
      return;
    }
    project.value = { ...hit, opportunities: hit.opportunities || [], members: hit.members || [] };
    selectedOpportunityID.value = project.value.opportunities[0]?.id || null;
    if (project.value.members.length && !project.value.is_demo) {
      teamProfiles.value = await networkRequest<NetworkProfile[]>("/api/v1/public/network/profiles").catch(() => []);
    } else {
      teamProfiles.value = [];
    }
    document.title = `BuilderNetwork | ${hit.title}`;
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "加载项目详情失败";
  } finally {
    loading.value = false;
  }
}

function profileForTeamMember(profileID: number) {
  return teamProfiles.value.find((profile) => profile.id === profileID);
}

function teamMemberName(profileID: number) {
  const member = project.value?.members.find((item) => item.profile_id === profileID);
  return profileForTeamMember(profileID)?.display_name || member?.display_name || "项目成员";
}

function timingText(timing?: NetworkProjectTiming | null) {
  if (!timing) return "";
  if (timing.label) return timing.label;
  if (timing.kind === "fixed_dates") return [timing.start_at, timing.end_at].filter(Boolean).join(" - ");
  if (timing.kind === "duration") return `${timing.duration_value} ${timing.duration_unit}`;
  if (timing.kind === "relative") return "按推进顺序";
  if (timing.kind === "flexible") return "弹性安排";
  return "";
}

function milestoneStatusLabel(status: string) {
  return ({ planned: "计划中", in_progress: "进行中", completed: "已完成", blocked: "受阻" } as Record<string, string>)[status] || status;
}

function projectStatusLabel(status: string) {
  return ({ published: "公开项目", archived: "已归档", draft: "草稿" } as Record<string, string>)[status] || status || "状态待更新";
}

function projectTypeLabel(type: string) {
  return ({ collaboration: "项目合作", advisor: "顾问合作", freelance: "短期外包", hiring: "招募团队" } as Record<string, string>)[type] || type || "合作项目";
}

const activeTab = ref<"overview" | "plan" | "team" | "discussions">("overview");

watch(slug, () => {
  void load();
}, { immediate: true });

const goApply = async () => {
  if (!activeOpportunity.value) return;
  await router.push({ path: "/market/projects", query: { apply_opportunity: String(activeOpportunity.value.id) } });
};
</script>

<template>
  <div class="page">
    <header class="market-nav-full">
      <div class="nav-inner">
        <div class="nav-primary">
          <RouterLink class="market-brand" to="/market">
            <img class="market-logo" :src="marketLogo" alt="VOIDLAB Logo" width="32" height="32" />
            <span>VOIDLAB <b>NETWORK</b></span>
          </RouterLink>
          <nav class="market-links" aria-label="市场导航">
            <RouterLink to="/market">发现</RouterLink>
            <RouterLink to="/market/builders">找 Builder</RouterLink>
            <RouterLink to="/market/projects">找项目</RouterLink>
          </nav>
        </div>
        <div class="market-actions">
          <template v-if="isAuthed">
            <RouterLink class="quiet-link" to="/market/workspace">工作台</RouterLink>
            <RouterLink class="publish-link" to="/market/publish">＋ 发布需求</RouterLink>
          </template>
          <template v-else>
            <RouterLink class="account-link" to="/market/register">加入</RouterLink>
          </template>
        </div>
      </div>
    </header>

    <main class="main">
      <div v-if="loading" class="state">正在加载项目详情…</div>
      <div v-else-if="error" class="state error">{{ error }}</div>
      <div v-else-if="project" class="detail">
        <div class="breadcrumb">
          <RouterLink to="/market/projects">全部项目</RouterLink>
          <span>/</span>
          <span>{{ project.title }}</span>
        </div>

        <section class="project-repo-header">
          <div class="repo-heading">
            <div class="repo-title-block">
              <div class="repo-kicker">
                <span class="repo-mark" aria-hidden="true">◈</span>
                <span>{{ project.owner_name || "VOIDLAB" }}</span>
                <span class="repo-divider">/</span>
                <h1 class="repo-project-name">{{ project.title }}</h1>
                <span class="repo-visibility">公开项目</span>
              </div>
              <div class="repo-badges">
                <span v-if="project.is_demo" class="project-demo-badge">本地演示数据</span>
                <span v-if="activeOpportunity" class="hero-type">{{ projectTypeLabel(activeOpportunity.opportunity_type) }}</span>
                <span class="repo-status">{{ projectStatusLabel(project.status) }}</span>
                <span v-if="activeOpportunity" class="hero-remote">{{ activeOpportunity.remote_allowed ? "支持远程" : (activeOpportunity.city_scope || "线下协作") }}</span>
              </div>
              <p class="repo-summary">{{ project.summary }}</p>
              <label v-if="project.opportunities.length > 1" class="project-opportunity-select">
                <span>申请合作机会</span>
                <select v-model.number="selectedOpportunityID" aria-label="选择申请的合作机会">
                  <option v-for="opportunity in project.opportunities" :key="opportunity.id" :value="opportunity.id">
                    {{ opportunity.title }}
                  </option>
                </select>
              </label>
              <div class="repo-facts">
                <span>{{ activeOpportunity?.engagement_mode || "合作方式待沟通" }}</span>
                <span>{{ activeOpportunity?.duration_hint || "周期面议" }}</span>
                <span>{{ activeOpportunity?.budget_hint || "预算未公开" }}</span>
              </div>
            </div>
            <div class="hero-actions">
              <RouterLink class="button-secondary" to="/market/projects">全部项目</RouterLink>
              <button class="button-primary" type="button" :disabled="project.is_demo || !activeOpportunity" @click="goApply">
                {{ project.is_demo ? "仅供预览" : activeOpportunity ? "提交合作申请" : "暂无开放机会" }}
              </button>
            </div>
          </div>
        </section>

        <div class="project-tabs">
          <button type="button" :class="['tab-item', { active: activeTab === 'overview' }]" @click="activeTab = 'overview'">Overview</button>
          <button type="button" :class="['tab-item', { active: activeTab === 'plan' }]" @click="activeTab = 'plan'">Plan</button>
          <button type="button" :class="['tab-item', { active: activeTab === 'team' }]" @click="activeTab = 'team'">Team</button>
        </div>

        <div v-show="activeTab === 'overview'" class="tab-pane">
          <div class="overview-layout">
            <section class="project-introduction">
              <div class="section-heading">
                <p class="eyebrow">PROJECT OVERVIEW</p>
                <h2>项目介绍</h2>
              </div>
              <p class="introduction-copy">{{ project.summary || "项目发起方暂未补充项目介绍。" }}</p>

              <div
                v-for="section in (project.project_plan?.sections || []).filter((item) => ['context', 'scope', 'outcome'].includes(item.type)).slice(0, 3)"
                :key="`intro-${section.sort_order}-${section.title}`"
                class="introduction-detail"
              >
                <h3>{{ section.title }}</h3>
                <p v-if="section.body">{{ section.body }}</p>
                <ul v-if="section.items?.length">
                  <li v-for="item in section.items" :key="item">{{ item }}</li>
                </ul>
              </div>

              <div class="intro-footer">
                <span>由 {{ project.owner_name || "项目发起方" }} 发布</span>
                <button class="text-action" type="button" @click="activeTab = 'plan'">查看项目计划 <span>→</span></button>
              </div>
            </section>

            <aside class="overview-sidebar">
              <section class="sidebar-card">
                <h2>About</h2>
                <p class="sidebar-description">{{ project.summary || "项目发起方暂未补充项目介绍。" }}</p>
                <div class="about-facts">
                  <div><span class="fact-icon">◉</span><span>{{ projectStatusLabel(project.status) }}</span></div>
                  <div v-if="activeOpportunity"><span class="fact-icon">◇</span><span>{{ projectTypeLabel(activeOpportunity.opportunity_type) }}</span></div>
                  <div v-if="activeOpportunity"><span class="fact-icon">⌁</span><span>{{ activeOpportunity.remote_allowed ? "支持远程协作" : (activeOpportunity.city_scope || "线下协作") }}</span></div>
                  <div v-if="activeOpportunity?.engagement_mode"><span class="fact-icon">↗</span><span>{{ activeOpportunity.engagement_mode }}</span></div>
                  <div v-if="activeOpportunity?.duration_hint"><span class="fact-icon">◷</span><span>{{ activeOpportunity.duration_hint }}</span></div>
                  <div v-if="activeOpportunity?.budget_hint"><span class="fact-icon">￥</span><span>{{ activeOpportunity.budget_hint }}</span></div>
                </div>
                <button class="sidebar-apply" type="button" :disabled="project.is_demo || !activeOpportunity" @click="goApply">
                  {{ project.is_demo ? "仅供预览" : activeOpportunity ? "提交合作申请" : "暂无开放机会" }} <span>↗</span>
                </button>
              </section>

              <section v-if="projectRoles.length || projectSkills.length" class="sidebar-card">
                <h2>正在寻找</h2>
                <div v-if="projectRoles.length" class="sidebar-tag-group">
                  <span v-for="role in projectRoles" :key="`role-${role}`" class="project-tag role-tag">{{ role }}</span>
                </div>
                <div v-if="projectSkills.length" class="sidebar-tag-group">
                  <span v-for="skill in projectSkills" :key="`skill-${skill}`" class="project-tag">{{ skill }}</span>
                </div>
              </section>

              <section v-if="project.members.length" class="sidebar-card contributor-card">
                <div class="sidebar-card-heading">
                  <h2>项目成员</h2>
                  <button class="text-action" type="button" @click="activeTab = 'team'">全部 <span>→</span></button>
                </div>
                <div class="contributor-list">
                  <RouterLink
                    v-for="member in project.members.slice(0, 5)"
                    :key="member.profile_id"
                    :to="profileForTeamMember(member.profile_id)?.slug ? `/market/builders/${profileForTeamMember(member.profile_id)?.slug}` : '/market/builders'"
                    class="contributor-row"
                  >
                    <span class="contributor-avatar">{{ (profileForTeamMember(member.profile_id)?.display_name || member.display_name || "B").slice(0, 1) }}</span>
                    <span class="contributor-copy">
                      <strong>{{ profileForTeamMember(member.profile_id)?.display_name || member.display_name || "Builder" }}</strong>
                      <small>{{ member.role }}</small>
                    </span>
                    <span class="contributor-arrow">↗</span>
                  </RouterLink>
                </div>
              </section>
            </aside>
          </div>
        </div>

        <div v-show="activeTab === 'plan'" class="tab-pane">
          <section v-if="project.project_plan?.milestones?.length" class="content-section milestone-section">
            <div class="section-heading">
              <p class="eyebrow">PROJECT PLAN</p>
              <h2>推进计划</h2>
              <p>按项目自身节奏拆分的推进节点；时间安排由项目发起方定义。</p>
            </div>
            <div class="milestone-timeline">
              <article
                v-for="(milestone, index) in project.project_plan.milestones"
                :key="`${milestone.order}-${milestone.title}`"
                class="milestone-item"
                :class="`status-${milestone.status}`"
              >
                <div class="milestone-rail" aria-hidden="true">
                  <span class="milestone-marker">{{ milestone.status === "completed" ? "✓" : String(index + 1).padStart(2, "0") }}</span>
                </div>
                <div class="milestone-content">
                  <div class="milestone-meta">
                    <span class="milestone-status">{{ milestoneStatusLabel(milestone.status) }}</span>
                    <span v-if="timingText(milestone.timing)" class="milestone-timing">{{ timingText(milestone.timing) }}</span>
                    <span v-for="ownerID in milestone.owner_member_ids || []" :key="ownerID" class="milestone-owner">负责人 · {{ teamMemberName(ownerID) }}</span>
                  </div>
                  <h3>{{ milestone.title }}</h3>
                  <p v-if="milestone.description" class="milestone-description">{{ milestone.description }}</p>
                  <div class="milestone-details">
                    <div v-if="milestone.goals?.length"><strong>目标</strong><ul><li v-for="goal in milestone.goals" :key="goal">{{ goal }}</li></ul></div>
                    <div v-if="milestone.deliverables?.length"><strong>交付物</strong><ul><li v-for="deliverable in milestone.deliverables" :key="deliverable">{{ deliverable }}</li></ul></div>
                    <div v-if="milestone.acceptance_criteria?.length"><strong>完成标准</strong><ul><li v-for="criteria in milestone.acceptance_criteria" :key="criteria">{{ criteria }}</li></ul></div>
                  </div>
                </div>
              </article>
            </div>
          </section>

          <div v-if="!project.project_plan?.milestones?.length" class="empty-state">
            <p>该项目暂未公开推进计划。</p>
          </div>
        </div>

        <div v-show="activeTab === 'team'" class="tab-pane">
          <section v-if="project.team_intro || project.members.length" class="content-section team-section">
            <div class="section-heading">
              <p class="eyebrow">PROJECT TEAM</p>
              <h2>项目团队</h2>
            </div>
            <p v-if="project.team_intro" class="team-intro">{{ project.team_intro }}</p>
            <div v-if="project.members.length" class="team-grid">
              <article v-for="member in project.members" :key="member.profile_id" class="team-card">
                <span class="team-avatar" aria-hidden="true">
                  {{ (profileForTeamMember(member.profile_id)?.display_name || member.display_name || "B").slice(0, 1) }}
                </span>
                <RouterLink v-if="profileForTeamMember(member.profile_id)?.slug" :to="`/market/builders/${profileForTeamMember(member.profile_id)?.slug}`" class="team-name">
                  {{ profileForTeamMember(member.profile_id)?.display_name || member.display_name }}
                </RouterLink>
                <strong v-else class="team-name">{{ profileForTeamMember(member.profile_id)?.display_name || member.display_name || "Builder" }}</strong>
                <span class="team-role">{{ member.role }}</span>
                <p v-if="member.contribution">{{ member.contribution }}</p>
                <small>{{ profileForTeamMember(member.profile_id)?.primary_role || member.primary_role || profileForTeamMember(member.profile_id)?.headline || "" }}</small>
              </article>
            </div>
          </section>
          <div v-if="!project.team_intro && !project.members.length" class="empty-state">
            <p>该项目暂未公开团队信息。</p>
          </div>
        </div>

        <div v-show="activeTab === 'discussions'" class="tab-pane">
          <div class="discussions-layout">
            <aside class="discussions-sidebar">
              <div class="search-box">
                <input type="text" placeholder="Search discussions..." class="search-input" />
              </div>
              <nav class="category-nav">
                <button 
                  :class="['category-btn', { active: currentDiscussionCategory === 'All' }]"
                  @click="currentDiscussionCategory = 'All'"
                >
                  💬 View all discussions
                </button>
                <div class="category-group">
                  <h3>Categories</h3>
                  <button 
                    :class="['category-btn', { active: currentDiscussionCategory === 'Announcements' }]"
                    @click="currentDiscussionCategory = 'Announcements'"
                  >📢 Announcements</button>
                  <button 
                    :class="['category-btn', { active: currentDiscussionCategory === 'General' }]"
                    @click="currentDiscussionCategory = 'General'"
                  >💬 General</button>
                  <button 
                    :class="['category-btn', { active: currentDiscussionCategory === 'Ideas' }]"
                    @click="currentDiscussionCategory = 'Ideas'"
                  >💡 Ideas</button>
                  <button 
                    :class="['category-btn', { active: currentDiscussionCategory === 'Q&A' }]"
                    @click="currentDiscussionCategory = 'Q&A'"
                  >🙏 Q&A</button>
                  <button 
                    :class="['category-btn', { active: currentDiscussionCategory === 'Show and tell' }]"
                    @click="currentDiscussionCategory = 'Show and tell'"
                  >🙌 Show and tell</button>
                </div>
              </nav>
            </aside>
            <main class="discussions-main">
              <div class="discussions-header">
                <h2>{{ currentDiscussionCategory === 'All' ? 'All discussions' : currentDiscussionCategory }}</h2>
                <div class="discussions-actions">
                  <button class="button-secondary">New discussion</button>
                </div>
              </div>
              <div class="discussions-list">
                <article v-for="discussion in filteredDiscussions" :key="discussion.id" class="discussion-item">
                  <div class="discussion-icon">
                    <span v-if="discussion.is_pinned" title="Pinned">📌</span>
                    <span v-else-if="discussion.category === 'Q&A' && discussion.is_solved" title="Solved" class="text-green-500">✅</span>
                    <span v-else-if="discussion.category === 'Q&A'" title="Unanswered">❔</span>
                    <span v-else>💬</span>
                  </div>
                  <div class="discussion-content">
                    <h3 class="discussion-title">
                      <a href="#">{{ discussion.title }}</a>
                      <span v-for="tag in discussion.tags" :key="tag" class="discussion-tag">{{ tag }}</span>
                    </h3>
                    <div class="discussion-meta">
                      <span class="author">{{ discussion.author.name }}</span>
                      <span class="category-badge">{{ discussion.category }}</span>
                      <span class="time">Last active {{ discussion.last_activity }}</span>
                    </div>
                  </div>
                  <div class="discussion-stats">
                    <span class="replies">
                      <svg viewBox="0 0 16 16" width="16" height="16" fill="currentColor" style="display:inline-block;vertical-align:middle;margin-right:4px;"><path d="M1.5 2.75a.25.25 0 0 1 .25-.25h8.5a.25.25 0 0 1 .25.25v5.5a.25.25 0 0 1-.25.25h-3.5a.75.75 0 0 0-.53.22L3.5 11.44V9.25a.75.75 0 0 0-.75-.75h-1a.25.25 0 0 1-.25-.25v-5.5Z"></path></svg>
                      {{ discussion.replies_count }}
                    </span>
                  </div>
                </article>
                <div v-if="filteredDiscussions.length === 0" class="empty-state">
                  <p>No discussions found in this category.</p>
                </div>
              </div>
            </main>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.page {
  min-height: 100vh;
  background: #FFFFFF;
  color: #0F172A;
}

.market-nav-full {
  position: sticky;
  top: 0;
  left: 0;
  z-index: 10;
  width: 100%;
  padding: 0 32px;
  box-sizing: border-box;
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid #E2E8F0;
}

.nav-inner {
  width: 100%;
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
}

.nav-primary {
  display: flex;
  align-items: center;
  gap: clamp(28px, 4vw, 68px);
}

.market-brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  color: inherit;
  font-weight: 900;
  font-size: 14px;
  letter-spacing: 0.15em;
  text-transform: uppercase;
}

.market-brand b {
  color: #0BCB78;
  font-weight: 800;
}

.market-logo {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  object-fit: contain;
  border-radius: 7px;
}

.market-links {
  display: flex;
  align-items: center;
  gap: 32px;
}

.market-links,
.market-actions {
  font-size: 13px;
}

.market-links a,
.quiet-link,
.account-link {
  text-decoration: none;
  color: #64748B;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  transition: color 0.2s;
}

.market-links a:hover,
.market-links a.router-link-active,
.quiet-link:hover,
.account-link:hover {
  color: #0BCB78;
}

.market-actions {
  display: flex;
  align-items: center;
  gap: 32px;
}

.publish-link,
.account-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 40px;
  padding: 0 18px;
  border: 0;
  border-radius: 999px;
  background: #0BCB78;
  color: #111827;
  letter-spacing: .04em;
}

.publish-link:hover,
.account-link:hover {
  background: #08B96C;
  color: #111827;
}

.main {
  width: 100%;
  max-width: 1800px;
  box-sizing: border-box;
  margin: 0 auto;
  padding: 28px clamp(28px, 3.5vw, 64px) 80px;
}

.state {
  padding: 56px 0;
  color: #64748B;
  font-weight: 700;
}

.state.error {
  color: #B91C1C;
}

.detail {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #64748B;
  font-size: 12px;
  font-weight: 700;
}

.breadcrumb a {
  color: inherit;
  text-decoration: none;
}

.breadcrumb a:hover {
  color: #0F172A;
}

.project-repo-header {
  padding: 8px 16px 28px;
  border-bottom: 1px solid #E2E8F0;
}

.repo-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 36px;
}

.repo-title-block {
  min-width: 0;
  flex: 1;
}

.repo-kicker {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 8px;
  color: #475569;
  font-size: 15px;
  font-weight: 700;
}

.repo-mark {
  display: grid;
  width: 28px;
  height: 28px;
  flex: none;
  place-items: center;
  border: 1px solid #DCE5E0;
  border-radius: 7px;
  background: #F2FBF6;
  color: #08B96C;
  font-size: 17px;
}

.repo-divider { color: #94A3B8; }

.repo-project-name {
  overflow: hidden;
  min-width: 0;
  margin: 0;
  color: #0BCB78;
  font-size: 16px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.repo-visibility,
.repo-status {
  flex: none;
  border: 1px solid #E2E8F0;
  border-radius: 999px;
  padding: 3px 9px;
  color: #64748B;
  font-size: 10px;
  font-weight: 700;
}

.repo-badges {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-top: 22px;
}

.hero-type,
.hero-remote {
  border: 1px solid #E2E8F0;
  border-radius: 999px;
  padding: 5px 9px;
  background: #F8FAFC;
  color: #475569;
  font-size: 10px;
  font-weight: 700;
}

.hero-type {
  border-color: #BDEBD4;
  background: #EFFBF5;
  color: #087A49;
}

.project-demo-badge {
  border: 1px solid #CDEEDC;
  border-radius: 999px;
  padding: 5px 9px;
  background: #F0FAF4;
  color: #16814F;
  font-size: 10px;
  font-weight: 700;
}

.repo-summary {
  max-width: 1040px;
  margin: 14px 0 0;
  color: #475569;
  font-size: 15px;
  line-height: 1.8;
}

.project-opportunity-select {
  display: flex;
  max-width: 680px;
  align-items: center;
  gap: 12px;
  margin-top: 14px;
  color: #475569;
  font-size: 12px;
  font-weight: 700;
}

.project-opportunity-select select {
  min-width: 0;
  flex: 1;
  padding: 8px 10px;
  border: 1px solid #DCE5E0;
  border-radius: 7px;
  background: #fff;
  color: #0F172A;
  font: inherit;
}

.repo-facts {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 22px;
  margin-top: 13px;
  color: #64748B;
  font-size: 11px;
  font-weight: 700;
}

.hero-actions {
  display: flex;
  flex: none;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 9px;
  padding-top: 2px;
}

.hero-actions .button-secondary,
.hero-actions .button-primary {
  display: inline-flex;
  min-height: 36px;
  align-items: center;
  justify-content: center;
  padding: 0 13px;
  border: 1px solid #D0D7DE;
  border-radius: 6px;
  background: #F6F8FA;
  color: #24292F;
  font-size: 12px;
  font-weight: 700;
  text-decoration: none;
  white-space: nowrap;
  cursor: pointer;
}

.hero-actions .button-secondary:hover {
  background: #EEF1F4;
}

.hero-actions .button-primary {
  border-color: #08B96C;
  background: #0BCB78;
  color: #0F172A;
}

.hero-actions .button-primary:hover {
  background: #08B96C;
}

.project-tabs {
  display: flex;
  gap: 32px;
  border-bottom: 1px solid #E2E8F0;
  margin-bottom: 40px;
  padding: 0 16px;
}

.tab-item {
  background: none;
  border: none;
  padding: 16px 0;
  font-size: 16px;
  font-weight: 700;
  color: #64748B;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.tab-item:hover {
  color: #0F172A;
}

.tab-item.active {
  color: #0BCB78;
  border-bottom-color: #0BCB78;
}

.tab-pane {
  animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}

.empty-state {
  padding: 60px 20px;
  text-align: center;
  color: #64748B;
  background: #F8FAFC;
  border-radius: 8px;
  font-size: 15px;
}

/* Discussions Layout */
.discussions-layout {
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: 32px;
  padding: 0 16px 40px;
}

.discussions-sidebar {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.search-input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #E2E8F0;
  border-radius: 6px;
  font-size: 14px;
  background: #F8FAFC;
}

.category-nav {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.category-group h3 {
  font-size: 12px;
  text-transform: uppercase;
  color: #94A3B8;
  margin-bottom: 8px;
  padding-left: 12px;
  font-weight: 700;
  letter-spacing: 0.05em;
}

.category-btn {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 8px 12px;
  border: none;
  background: transparent;
  color: #475569;
  font-size: 14px;
  text-align: left;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.category-btn:hover {
  background: #F1F5F9;
  color: #0F172A;
}

.category-btn.active {
  background: #E2E8F0;
  color: #0F172A;
  font-weight: 500;
}

.discussions-main {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.discussions-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 16px;
  border-bottom: 1px solid #E2E8F0;
}

.discussions-header h2 {
  font-size: 20px;
  font-weight: 600;
  margin: 0;
}

.discussion-item {
  display: flex;
  align-items: flex-start;
  padding: 16px;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  margin-bottom: 12px;
  background: #FFFFFF;
  transition: border-color 0.2s;
}

.discussion-item:hover {
  border-color: #CBD5E1;
}

.discussion-icon {
  font-size: 18px;
  margin-right: 16px;
  padding-top: 2px;
  opacity: 0.8;
}

.discussion-content {
  flex: 1;
  min-width: 0;
}

.discussion-title {
  margin: 0 0 8px 0;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.4;
}

.discussion-title a {
  color: #0F172A;
  text-decoration: none;
}

.discussion-title a:hover {
  color: #0BCB78;
}

.discussion-tag {
  display: inline-block;
  margin-left: 8px;
  padding: 2px 8px;
  font-size: 12px;
  background: #F1F5F9;
  color: #475569;
  border-radius: 12px;
  vertical-align: middle;
}

.discussion-meta {
  font-size: 13px;
  color: #64748B;
  display: flex;
  align-items: center;
  gap: 12px;
}

.category-badge {
  padding: 2px 8px;
  border: 1px solid #E2E8F0;
  border-radius: 12px;
  font-size: 12px;
}

.discussion-stats {
  display: flex;
  align-items: center;
  gap: 16px;
  color: #64748B;
  font-size: 13px;
  margin-left: 16px;
  padding-top: 4px;
}

@media (max-width: 768px) {
  .discussions-layout {
    grid-template-columns: 1fr;
  }
  
  .discussions-sidebar {
    border-bottom: 1px solid #E2E8F0;
    padding-bottom: 24px;
  }
  
  .project-tabs {
    overflow-x: auto;
    white-space: nowrap;
    -webkit-overflow-scrolling: touch;
    padding-bottom: 2px;
  }
}

.grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 16px;
}

.panel {
  border: 1px solid #E2E8F0;
  border-radius: 18px;
  padding: 22px;
  background: #FFFFFF;
}

.panel h2 {
  margin: 0 0 14px;
  font-size: 12px;
  letter-spacing: 0.14em;
  text-transform: uppercase;
  color: #64748B;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.tags span {
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  padding: 8px 10px;
  border-radius: 999px;
  border: 1px solid #E2E8F0;
  background: #FFFFFF;
  color: #0F172A;
}

.empty {
  color: #64748B;
  font-weight: 700;
  font-size: 13px;
}

.overview-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(300px, 380px);
  gap: clamp(32px, 5vw, 72px);
  align-items: start;
  padding: 0 16px 40px;
}

.project-introduction {
  min-width: 0;
  padding: 8px 0 30px;
}

.project-introduction .section-heading {
  margin-bottom: 22px;
  padding-bottom: 18px;
  border-bottom: 1px solid #E2E8F0;
}

.project-introduction .eyebrow,
.sidebar-card .eyebrow {
  margin: 0 0 8px;
  color: #64748B;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: .14em;
}

.introduction-copy {
  margin: 0;
  color: #334155;
  font-size: 17px;
  line-height: 1.9;
  white-space: pre-wrap;
}

.introduction-detail {
  margin-top: 30px;
  padding-bottom: 24px;
  border-bottom: 1px solid #E2E8F0;
}

.introduction-detail h3 {
  margin: 0 0 10px;
  color: #0F172A;
  font-size: 16px;
}

.introduction-detail p,
.introduction-detail li {
  color: #475569;
  font-size: 15px;
  line-height: 1.9;
}

.introduction-detail p { margin: 0; white-space: pre-wrap; }
.introduction-detail ul { margin: 10px 0 0; padding-left: 20px; }
.introduction-detail li::marker { color: #0BCB78; }

.intro-footer {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  align-items: center;
  gap: 14px;
  margin-top: 36px;
  padding-top: 18px;
  border-top: 1px solid #E2E8F0;
  color: #64748B;
  font-size: 11px;
}

.text-action {
  padding: 0;
  border: 0;
  background: transparent;
  color: #334155;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
}

.text-action:hover { color: #08B96C; }
.text-action span { color: #0BCB78; }

.overview-sidebar {
  display: grid;
  gap: 26px;
  min-width: 0;
  padding-top: 8px;
}

.sidebar-card {
  padding-bottom: 24px;
  border-bottom: 1px solid #E2E8F0;
}

.sidebar-card h2 {
  margin: 0 0 14px;
  color: #0F172A;
  font-size: 15px;
  letter-spacing: -.01em;
}

.sidebar-description {
  display: -webkit-box;
  overflow: hidden;
  margin: 0 0 16px;
  color: #475569;
  font-size: 13px;
  line-height: 1.7;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 4;
}

.about-facts {
  display: grid;
  gap: 12px;
}

.about-facts > div {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  color: #475569;
  font-size: 12px;
  line-height: 1.5;
}

.fact-icon {
  width: 16px;
  flex: none;
  color: #64748B;
  text-align: center;
}

.sidebar-apply {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: space-between;
  margin-top: 20px;
  padding: 11px 14px;
  border: 0;
  border-radius: 6px;
  background: #0BCB78;
  color: #0F172A;
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
}

.sidebar-apply:hover { background: #08B96C; }
.sidebar-apply span { font-size: 16px; }

.sidebar-tag-group {
  display: flex;
  flex-wrap: wrap;
  gap: 7px;
  margin-top: 9px;
}

.project-tag {
  display: inline-flex;
  padding: 5px 9px;
  border-radius: 999px;
  background: #EFF6F3;
  color: #276749;
  font-size: 10px;
  font-weight: 700;
}

.project-tag.role-tag {
  background: #F1F5F9;
  color: #334155;
}

.sidebar-card-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.sidebar-card-heading h2 { margin-bottom: 14px; }
.sidebar-card-heading .text-action { margin-top: -12px; font-size: 11px; }

.contributor-list { display: grid; gap: 12px; }

.contributor-row {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
  color: inherit;
  text-decoration: none;
}

.contributor-row:hover strong { color: #08B96C; }

.contributor-avatar {
  display: grid;
  width: 32px;
  height: 32px;
  flex: none;
  place-items: center;
  border: 1px solid #DCE5E0;
  border-radius: 50%;
  background: #F8FBF9;
  color: #0F172A;
  font-size: 12px;
  font-weight: 800;
}

.contributor-copy {
  display: grid;
  min-width: 0;
  gap: 3px;
}

.contributor-copy strong {
  overflow: hidden;
  color: #334155;
  font-size: 11px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.contributor-copy small {
  overflow: hidden;
  color: #64748B;
  font-size: 10px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.contributor-arrow {
  margin-left: auto;
  color: #94A3B8;
  font-size: 12px;
}

.content-section {
  margin-top: 18px;
  padding: clamp(22px, 4vw, 42px);
  border: 1px solid #E2E8F0;
  border-radius: 18px;
  background: #F8FAFC;
}
.content-section.milestone-section {
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
}
.content-section.team-section {
  padding: 0;
  border: 0;
  border-radius: 0;
  background: transparent;
}

.section-heading {
  max-width: 720px;
  margin-bottom: 28px;
}

.section-heading .eyebrow { margin-bottom: 10px; }
.section-heading h2 {
  margin: 0;
  color: #0F172A;
  font-size: clamp(24px, 3vw, 36px);
  letter-spacing: -.04em;
}
.section-heading > p:last-child:not(.eyebrow) {
  margin: 10px 0 0;
  color: #64748B;
  font-size: 13px;
  line-height: 1.7;
}

.milestone-timeline { display: flex; flex-direction: column; }
.milestone-item {
  display: grid;
  grid-template-columns: 48px minmax(0, 1fr);
  gap: 22px;
  position: relative;
  padding-bottom: 28px;
}
.milestone-rail {
  position: relative;
  display: flex;
  justify-content: center;
}
.milestone-rail::after {
  position: absolute;
  top: 38px;
  bottom: 0;
  left: 50%;
  width: 1px;
  background: #DCE5E0;
  content: "";
  transform: translateX(-50%);
}
.milestone-item:last-child .milestone-rail::after { display: none; }
.milestone-marker {
  z-index: 1;
  display: grid;
  width: 36px;
  height: 36px;
  place-items: center;
  border: 1px solid #CBD5E1;
  border-radius: 50%;
  background: #FFFFFF;
  color: #64748B;
  font: 700 11px ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: .02em;
}
.milestone-item.status-in_progress .milestone-marker {
  border-color: #0BCB78;
  background: #ECFDF5;
  color: #087A49;
  box-shadow: 0 0 0 5px #F0FDF4;
}
.milestone-item.status-completed .milestone-marker {
  border-color: #0BCB78;
  background: #0BCB78;
  color: #FFFFFF;
  font-family: inherit;
  font-size: 15px;
}
.milestone-item.status-blocked .milestone-marker {
  border-color: #FCA5A5;
  background: #FEF2F2;
  color: #B91C1C;
}
.milestone-content {
  min-width: 0;
  padding: 2px 0 24px;
  border-bottom: 1px solid #E2E8F0;
}
.milestone-item:last-child .milestone-content { padding-bottom: 0; border-bottom: 0; }
.milestone-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}
.milestone-meta span {
  padding: 5px 9px;
  border: 1px solid #DCE5E0;
  border-radius: 999px;
  color: #475569;
  background: #F8FBF9;
  font-size: 10px;
  font-weight: 700;
}
.milestone-meta .milestone-status { border-color: #CBD5E1; background: #F1F5F9; color: #475569; }
.milestone-item.status-in_progress .milestone-meta .milestone-status { border-color: #BBF7D0; background: #ECFDF5; color: #087A49; }
.milestone-item.status-completed .milestone-meta .milestone-status { border-color: #BBF7D0; background: #F0FDF4; color: #15803D; }
.milestone-item.status-blocked .milestone-meta .milestone-status { border-color: #FECACA; background: #FEF2F2; color: #B91C1C; }
.milestone-meta .milestone-timing,
.milestone-meta .milestone-owner { border-color: transparent; background: #F8FAFC; color: #64748B; }
.milestone-content h3,
.plan-section-card h3 {
  margin: 0;
  color: #0F172A;
  font-size: 19px;
  letter-spacing: -.025em;
}
.milestone-description,
.plan-section-card > p,
.team-card p {
  margin: 10px 0 0;
  color: #475569;
  font-size: 13px;
  line-height: 1.75;
  white-space: pre-wrap;
}
.milestone-details {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
  margin-top: 20px;
  padding-top: 18px;
  border-top: 1px solid #E2E8F0;
}
.milestone-details strong {
  color: #64748B;
  font-size: 10px;
  letter-spacing: .08em;
  text-transform: uppercase;
}
.milestone-details ul,
.plan-section-card ul {
  margin: 8px 0 0;
  padding-left: 17px;
  color: #334155;
  font-size: 12px;
  line-height: 1.75;
}
.plan-section-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}
.plan-section-card { min-width: 0; }
.section-type {
  display: inline-block;
  margin-bottom: 10px;
  color: #0BCB78;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: .12em;
  text-transform: uppercase;
}
.team-intro {
  max-width: 900px;
  margin: -8px 0 28px;
  color: #475569;
  font-size: 14px;
  line-height: 1.8;
  white-space: pre-wrap;
}
.team-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  column-gap: 36px;
}
.team-card {
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr);
  column-gap: 14px;
  row-gap: 5px;
  align-content: start;
  min-width: 0;
  padding: 20px 0;
  border-bottom: 1px solid #E2E8F0;
}
.team-card:nth-last-child(-n + 2) { border-bottom: 0; }
.team-avatar {
  display: grid;
  width: 38px;
  height: 38px;
  grid-row: 1 / span 4;
  place-items: center;
  border: 1px solid #DCE5E0;
  border-radius: 50%;
  background: #F2FBF6;
  color: #087A49;
  font-size: 13px;
  font-weight: 800;
}
.team-name {
  color: #0F172A;
  font-size: 16px;
  font-weight: 800;
  text-decoration: none;
}
a.team-name:hover {
  text-decoration: underline;
  text-decoration-color: #0BCB78;
  text-underline-offset: 3px;
}
.team-role {
  color: #0BCB78;
  font-size: 11px;
  font-weight: 700;
}
.team-card p { margin: 3px 0 0; }
.team-card small {
  margin-top: 3px;
  color: #64748B;
  font-size: 11px;
  line-height: 1.6;
}

@media (max-width: 980px) {
  .market-nav-full { padding: 0 24px; }
  .nav-primary { gap: 24px; }
  .market-links { gap: 20px; }
  .team-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .milestone-details { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .repo-heading {
    flex-direction: column;
    gap: 22px;
  }
  .hero-actions {
    padding-top: 0;
  }
  .overview-layout {
    grid-template-columns: minmax(0, 1fr) minmax(240px, 290px);
    gap: 28px;
    padding-right: 8px;
    padding-left: 8px;
  }
  .grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 600px) {
  .market-nav-full { padding: 0 14px; }
  .nav-inner { height: 68px; }
  .nav-primary { gap: 12px; }
  .market-brand { gap: 0; }
  .market-brand span { display: none; }
  .market-links { gap: 10px; }
  .market-links a { font-size: 10px; letter-spacing: 0; }
  .market-actions { gap: 12px; }
  .account-link { min-height: 34px; padding: 0 12px; font-size: 10px; }
  .main { padding: 20px 16px 56px; }
  .project-repo-header { padding: 0 4px 22px; }
  .repo-kicker { flex-wrap: wrap; font-size: 13px; }
  .repo-project-name { max-width: 100%; font-size: 14px; }
  .repo-badges { margin-top: 16px; }
  .repo-summary { font-size: 13px; }
  .hero-actions { width: 100%; }
  .hero-actions > * { flex: 1; justify-content: center; text-align: center; }
  .overview-layout {
    grid-template-columns: 1fr;
    gap: 26px;
    padding: 0 4px 24px;
  }
  .overview-sidebar { gap: 20px; }
  .project-introduction { padding-top: 0; }
  .introduction-copy { font-size: 15px; }
  .content-section { padding: 20px 16px; border-radius: 12px; }
  .content-section.team-section { padding: 0; }
  .milestone-item { grid-template-columns: 36px minmax(0, 1fr); gap: 14px; padding-bottom: 20px; }
  .milestone-marker { width: 30px; height: 30px; font-size: 10px; }
  .milestone-rail::after { top: 32px; }
  .milestone-content { padding-bottom: 18px; }
  .milestone-details,
  .plan-section-grid,
  .team-grid { grid-template-columns: 1fr; }
  .team-card:nth-last-child(-n + 2) { border-bottom: 1px solid #E2E8F0; }
  .team-card:last-child { border-bottom: 0; }
}
</style>
