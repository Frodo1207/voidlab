<script setup lang="ts">
import { computed, defineAsyncComponent, onBeforeUnmount, onMounted, reactive, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import FaultyTerminal from "../components/FaultyTerminal.vue";
import MarketAccordionGallery, { type MarketAccordionGalleryItem } from "../components/MarketAccordionGallery.vue";
import marketLogo from "../../../assets/logo/white-bg.svg";
import {
  clearNetworkToken,
  devMockProfiles,
  getNetworkToken,
  networkRequest,
  NetworkRequestError,
  postJSON,
  projectFromLegacyOpportunity,
  saveNetworkSession,
  type NetworkContactRequest,
  type NetworkAgentToken,
  type NetworkAgentTokenCreateResult,
  type NetworkOpportunity,
  type NetworkOpportunityMember,
  type NetworkProfile,
  type NetworkProject,
  type NetworkProjectPlanSection,
  type NetworkSession,
  type NetworkWork
} from "../composables/useBuilderMarket";

interface SkillPackDownload {
  value: string;
  label: string;
  kind: string;
  description: string;
  href: string;
}

interface SkillPackGroup {
  value: string;
  label: string;
  items: SkillPackDownload[];
}

const skillPackGroups: SkillPackGroup[] = [
  {
    value: "setup",
    label: "START HERE",
    items: [
      {
        value: "readme",
        label: "安装与鉴权说明",
        kind: "README",
        description: "了解安装方式、邀请码注册与 Agent Token 配置。",
        href: "/downloads/voidlab-buildernetwork-agent/README.md"
      }
    ]
  },
  {
    value: "hosts",
    label: "AGENT HOSTS",
    items: [
      {
        value: "trae",
        label: "Trae Skill",
        kind: "SKILL.MD",
        description: "在 Trae 中通过自然语言管理 BuilderNetwork。",
        href: "/downloads/voidlab-buildernetwork-agent/trae/voidlab-buildernetwork-agent/SKILL.md"
      },
      {
        value: "codex",
        label: "Codex Bridge",
        kind: "AGENTS.MD",
        description: "为 Codex 与兼容 Agent 配置市场 API 工作流。",
        href: "/downloads/voidlab-buildernetwork-agent/codex/AGENTS.md"
      },
      {
        value: "claude",
        label: "Claude Code Bridge",
        kind: "CLAUDE.MD",
        description: "在 Claude Code 终端中接入 BuilderNetwork。",
        href: "/downloads/voidlab-buildernetwork-agent/claude-code/CLAUDE.md"
      }
    ]
  }
];
const expandedSkillGroups = ref(["setup", "hosts"]);
const selectedSkillValue = ref("trae");
const selectedSkillDownload = computed(() =>
  skillPackGroups.flatMap((group) => group.items).find((item) => item.value === selectedSkillValue.value)
  ?? skillPackGroups[1].items[0]
);
const isSkillGroupExpanded = (value: string) => expandedSkillGroups.value.includes(value);
const toggleSkillGroup = (value: string) => {
  expandedSkillGroups.value = isSkillGroupExpanded(value)
    ? expandedSkillGroups.value.filter((item) => item !== value)
    : [...expandedSkillGroups.value, value];
};
const skillBranchPath = (index: number) => {
  const y = index * 58 + 29;
  return `M 4 0 V ${y - 8} Q 4 ${y} 12 ${y} H 30`;
};
const skillTrunkPath = (count: number) => `M 4 0 V ${Math.max(0, count * 58 - 29)}`;

const popularCoverBySlug: Record<string, string> = {
  "agent-workflow-platform-mvp": "/market/covers/agent-workflow-platform-mvp.jpg",
  "buildernetwork-discovery-ranking": "/market/covers/buildernetwork-discovery-ranking.jpg",
  "observability-dashboard": "/market/covers/observability-dashboard.jpg",
  "agent-sandbox-automation": "/market/covers/agent-sandbox-automation.jpg",
  "multimodal-commerce-ops": "/market/covers/multimodal-commerce-ops.jpg"
};
const coverForProjectSlug = (slug: string, index: number) => {
  const mapped = popularCoverBySlug[slug];
  if (mapped) return mapped;
  const defaults = [
    "/market/covers/agent-workflow-platform-mvp.jpg",
    "/market/covers/buildernetwork-discovery-ranking.jpg",
    "/market/covers/observability-dashboard.jpg",
    "/market/covers/agent-sandbox-automation.jpg",
    "/market/covers/multimodal-commerce-ops.jpg"
  ];
  return defaults[index % defaults.length];
};

const route = useRoute();
const router = useRouter();
const ShapeWaves = defineAsyncComponent(() => import("../components/ShapeWaves.vue"));

const section = computed(() => typeof route.params.section === "string" ? route.params.section : "discover");

const isAuthed = ref(false);
const isAuthPage = computed(() => ["login", "register"].includes(section.value));

onMounted(() => {
  if (getNetworkToken()) {
    isAuthed.value = true;
  }
  window.addEventListener("keydown", focusProjectSearchShortcut);
});
onBeforeUnmount(() => window.removeEventListener("keydown", focusProjectSearchShortcut));
const profiles = ref<NetworkProfile[]>([]);
const builderDirectoryProfiles = computed(() =>
  import.meta.env.DEV ? [...devMockProfiles, ...profiles.value] : profiles.value
);
const works = ref<NetworkWork[]>([]);
const myWorks = ref<NetworkWork[]>([]);
const projects = ref<NetworkProject[]>([]);
const opportunities = ref<NetworkOpportunity[]>([]);
const myOpportunities = ref<NetworkOpportunity[]>([]);
const builderSearch = ref("");
const builderRoleFilter = ref("all");
const builderAvailableOnly = ref(false);
const builderRemoteOnly = ref(false);
const builderSort = ref("default");
const projectSearch = ref("");
const projectTypeFilter = ref("all");
const projectRoleFilter = ref("");
const projectRemoteOnly = ref(false);
const projectSort = ref("newest");

const builderRoleOptions = computed(() => {
  const roleCounts = new Map<string, number>();
  builderDirectoryProfiles.value.forEach((person) => {
    if (person.primary_role) roleCounts.set(person.primary_role, (roleCounts.get(person.primary_role) || 0) + 1);
  });
  return [...roleCounts.entries()]
    .sort(([leftRole, leftCount], [rightRole, rightCount]) => {
      const leftNumeric = /^\d+$/.test(leftRole.trim());
      const rightNumeric = /^\d+$/.test(rightRole.trim());
      if (leftNumeric !== rightNumeric) return leftNumeric ? 1 : -1;
      return rightCount - leftCount || leftRole.localeCompare(rightRole, "zh-CN");
    })
    .slice(0, 5)
    .map(([role, count]) => ({ role, count }));
});

const visibleProfiles = computed(() => {
  const query = builderSearch.value.trim().toLocaleLowerCase();
  const filtered = builderDirectoryProfiles.value.filter((person) => {
    const searchableText = [
      person.display_name,
      person.primary_role,
      person.headline,
      person.bio,
      person.city,
      ...(person.skills || []),
      ...(person.collaboration_modes || [])
    ].join(" ").toLocaleLowerCase();
    return (!query || searchableText.includes(query))
      && (builderRoleFilter.value === "all" || person.primary_role === builderRoleFilter.value)
      && (!builderAvailableOnly.value || person.availability !== "unavailable")
      && (!builderRemoteOnly.value || person.city === "远程" || person.collaboration_modes?.some((mode) => mode.includes("远程")));
  });
  if (builderSort.value === "name") {
    return filtered.sort((left, right) => left.display_name.localeCompare(right.display_name, "zh-CN"));
  }
  return filtered;
});

const hasBuilderFilters = computed(() =>
  Boolean(
    builderSearch.value.trim()
    || builderRoleFilter.value !== "all"
    || builderAvailableOnly.value
    || builderRemoteOnly.value
  )
);

function resetBuilderFilters() {
  builderSearch.value = "";
  builderRoleFilter.value = "all";
  builderAvailableOnly.value = false;
  builderRemoteOnly.value = false;
  builderSort.value = "default";
}

const projectRoleOptions = computed(() => {
  const roleCounts = new Map<string, number>();
  projects.value.forEach((project) => {
    project.opportunities.forEach((item) => {
      (item.looking_for_roles || []).forEach((role) => {
        if (role) roleCounts.set(role, (roleCounts.get(role) || 0) + 1);
      });
    });
  });
  return [...roleCounts.entries()]
    .sort(([leftRole, leftCount], [rightRole, rightCount]) =>
      rightCount - leftCount || leftRole.localeCompare(rightRole, "zh-CN")
    )
    .slice(0, 5)
    .map(([role]) => role);
});

const visibleProjects = computed(() => {
  const query = projectSearch.value.trim().toLocaleLowerCase();
  const filtered = projects.value.filter((project) => {
    const projectOpportunities = project.opportunities || [];
    const roles = [...new Set(projectOpportunities.flatMap((item) => item.looking_for_roles || []))];
    const skills = [...new Set(projectOpportunities.flatMap((item) => item.required_skills || []))];
    const searchableText = [
      project.title,
      project.summary,
      project.owner_name,
      ...roles,
      ...skills
    ].join(" ").toLocaleLowerCase();
    return (!query || searchableText.includes(query))
      && (projectTypeFilter.value === "all" || projectOpportunities.some((item) => item.opportunity_type === projectTypeFilter.value))
      && (!projectRoleFilter.value || roles.includes(projectRoleFilter.value))
      && (!projectRemoteOnly.value || projectOpportunities.some((item) => item.remote_allowed));
  });

  if (projectSort.value === "title") {
    return filtered.sort((left, right) => left.title.localeCompare(right.title, "zh-CN"));
  }
  return filtered.sort((left, right) => {
    const leftDate = Date.parse(left.updated_at || left.created_at || "") || 0;
    const rightDate = Date.parse(right.updated_at || right.created_at || "") || 0;
    return rightDate - leftDate;
  });
});

function resetProjectFilters() {
  projectSearch.value = "";
  projectTypeFilter.value = "all";
  projectRoleFilter.value = "";
  projectRemoteOnly.value = false;
  projectSort.value = "newest";
}

function focusProjectSearchShortcut(event: KeyboardEvent) {
  if (event.key !== "/" || !["projects", "builders"].includes(section.value)) return;
  const target = event.target;
  if (target instanceof HTMLElement && (target.isContentEditable || ["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName))) return;
  event.preventDefault();
  document.getElementById(section.value === "projects" ? "project-search" : "builder-search")?.focus();
}

function projectTypeLabel(type: string) {
  return ({
    collaboration: "项目合作",
    freelance: "短期外包",
    advisor: "顾问支持",
    cofounder: "联合创始"
  } as Record<string, string>)[type] || type || "合作机会";
}

function requestStatusLabel(request: NetworkContactRequest) {
  if (request.status === "accepted") return "已同意";
  if (request.status === "pending") {
    return request.from_profile_id === profile.value?.id ? "等待回复" : "待处理";
  }
  if (request.status === "declined") return "已婉拒";
  if (request.status === "withdrawn") return "已撤回";
  return request.status;
}

function requestExchange(request: NetworkContactRequest) {
  return request.is_demo ? devMockContactExchanges[request.id] : exchange.value[request.id];
}

function exchangeContactFields(contact?: Record<string, string>) {
  return [
    { label: "邮箱", value: contact?.email || "" },
    { label: "微信", value: contact?.wechat || "" },
    { label: "电话", value: contact?.phone || "" }
  ].filter((field) => field.value);
}

function projectTypeCount(type: string) {
  return type === "all"
    ? projects.value.length
    : projects.value.filter((project) => project.opportunities.some((item) => item.opportunity_type === type)).length;
}

const hasProjectFilters = computed(() =>
  Boolean(projectSearch.value || projectTypeFilter.value !== "all" || projectRoleFilter.value || projectRemoteOnly.value)
);

const applyProfileId = computed(() => {
  const raw = route.query.apply_profile;
  if (typeof raw !== "string" || !raw) return null;
  const id = Number(raw);
  return Number.isFinite(id) ? id : null;
});

const applyOpportunityId = computed(() => {
  const raw = route.query.apply_opportunity;
  if (typeof raw !== "string" || !raw) return null;
  const id = Number(raw);
  return Number.isFinite(id) ? id : null;
});

watch([applyProfileId, profiles], async ([id, nextProfiles]) => {
  if (!id) return;
  const target = nextProfiles.find((item) => item.id === id);
  if (!target || target.is_demo) return;
  await openApplication("profile", target.id, target.display_name);
  await router.replace({ path: route.path });
}, { immediate: true });

watch([applyOpportunityId, opportunities], async ([id, nextOpportunities]) => {
  if (!id) return;
  const target = nextOpportunities.find((item) => item.id === id);
  if (!target) return;
  await openApplication("opportunity", target.id, target.title);
  await router.replace({ path: route.path });
}, { immediate: true });

async function applyToProject(project: NetworkProject) {
  if (project.is_demo || !project.opportunities.length) return;
  if (project.opportunities.length > 1) {
    await router.push(`/market/projects/${project.slug}`);
    return;
  }
  const opportunity = project.opportunities[0];
  await openApplication("opportunity", opportunity.id, opportunity.title || project.title);
}

const popularGalleryItems = computed<MarketAccordionGalleryItem[]>(() =>
  [...projects.value.filter((item) => item.is_demo), ...projects.value.filter((item) => !item.is_demo)]
    .slice(0, 5)
    .map((project, index) => {
      const opportunity = project.opportunities[0];
      return {
          id: project.id,
          title: project.title,
          summary: project.summary,
          meta: project.is_demo ? "演示项目" : opportunity?.remote_allowed ? "REMOTE" : (opportunity?.city_scope || "ONSITE"),
          tags: [...new Set(project.opportunities.flatMap((item) => [...(item.looking_for_roles || []), ...(item.required_skills || [])]))],
          href: `/market/projects/${project.slug}`,
          image: coverForProjectSlug(project.slug, index),
          image_alt: `封面：${project.title}`
        };
      })
);
const profile = ref<NetworkProfile | null>(null);
const incoming = ref<NetworkContactRequest[]>([]);
const outgoing = ref<NetworkContactRequest[]>([]);
const agentTokens = ref<NetworkAgentToken[]>([]);
const devMockIncomingRequests = (targetProfileId: number): NetworkContactRequest[] => [
  {
    id: -91001,
    is_demo: true,
    from_profile_id: 99002,
    target_type: "profile",
    target_id: targetProfileId,
    message: "你好，我正在搭建一个面向小团队的 Agent 工作流，希望找一位熟悉产品落地的伙伴一起打磨首版。看了你的主页，想先约一次线上交流。",
    status: "pending",
    progress_status: "",
    exchange_revoked: false,
    applicant_name: "林舟",
    applicant_role: "AI 产品工程师",
    share_email: false,
    share_wechat: false,
    share_phone: false,
    related_work_ids: []
  },
  {
    id: -91002,
    is_demo: true,
    from_profile_id: 99003,
    target_type: "profile",
    target_id: targetProfileId,
    message: "我们在做企业知识库检索与自动化分发，正在寻找有全栈经验的伙伴。想先聊聊项目范围、协作方式和每周投入。",
    status: "pending",
    progress_status: "",
    exchange_revoked: false,
    applicant_name: "沈知夏",
    applicant_role: "全栈开发者",
    share_email: true,
    share_wechat: false,
    share_phone: false,
    related_work_ids: []
  },
  {
    id: -91003,
    is_demo: true,
    from_profile_id: 99004,
    target_type: "profile",
    target_id: targetProfileId,
    message: "我已同意这次合作申请，可以通过双方授权的联系方式直接联系。",
    status: "accepted",
    progress_status: "",
    exchange_revoked: false,
    applicant_name: "陈以南",
    applicant_role: "独立开发者",
    share_email: false,
    share_wechat: true,
    share_phone: false,
    related_work_ids: []
  }
];
const devMockOutgoingRequests = (profileId: number, profileName: string): NetworkContactRequest[] => [
  {
    id: -92001,
    is_demo: true,
    from_profile_id: profileId,
    target_type: "opportunity",
    target_id: 99101,
    target_label: "Agent 知识库协作计划",
    message: "你好，我们有兴趣参与知识库检索和自动化分发的首版建设，可以承担前端集成与工作流编排。希望进一步了解目前的计划和协作节奏。",
    status: "pending",
    progress_status: "",
    exchange_revoked: false,
    applicant_name: profileName,
    applicant_role: "AI Builder",
    share_email: false,
    share_wechat: false,
    share_phone: false,
    related_work_ids: []
  },
  {
    id: -92002,
    is_demo: true,
    from_profile_id: profileId,
    target_type: "profile",
    target_id: 99005,
    target_label: "周行知 · 全栈开发者",
    message: "你好，看到你关注 Agent 工作流和产品落地。我正在做一个面向独立开发者的效率工具，希望找伙伴一起验证需求并完成早期版本。",
    status: "pending",
    progress_status: "",
    exchange_revoked: false,
    applicant_name: profileName,
    applicant_role: "AI Builder",
    share_email: true,
    share_wechat: false,
    share_phone: false,
    related_work_ids: []
  },
  {
    id: -92003,
    is_demo: true,
    from_profile_id: profileId,
    target_type: "profile",
    target_id: 99004,
    target_label: "陈以南 · 独立开发者",
    message: "对方已同意申请，双方授权的联系方式现已开放。",
    status: "accepted",
    progress_status: "",
    exchange_revoked: false,
    applicant_name: profileName,
    applicant_role: "AI Builder",
    share_email: false,
    share_wechat: true,
    share_phone: false,
    related_work_ids: []
  }
];
const devMockContactExchanges: Record<number, { applicant: Record<string, string>; recipient: Record<string, string> }> = {
  [-91003]: {
    applicant: { email: "chen.yinan@example.test", wechat: "chen_yinan_demo" },
    recipient: { email: "local.builder@example.test", wechat: "local_builder_demo" }
  },
  [-92003]: {
    applicant: { email: "local.builder@example.test", wechat: "local_builder_demo" },
    recipient: { email: "chen.yinan@example.test", wechat: "chen_yinan_demo" }
  }
};
const workspaceTab = ref<"overview" | "profile" | "works" | "opportunities" | "messages" | "settings">("overview");
const pendingIncomingRequests = computed(() => incoming.value.filter((request) => request.status === "pending"));
const profileCompletion = computed(() => {
  if (!profile.value) return 0;
  const fields = [
    profile.value.display_name,
    profile.value.primary_role,
    profile.value.headline,
    profile.value.bio,
    profile.value.city || profile.value.collaboration_modes?.length,
    profile.value.skills?.length
  ];
  return Math.round(fields.filter(Boolean).length / fields.length * 100);
});
const loading = ref(false);
const error = ref("");
const notice = ref("");
const warning = ref("");
const inviteValid = ref(false);
const selectedTarget = ref<{ type: "profile" | "opportunity"; id: number; title: string } | null>(null);
const applicationMessage = ref("");
const applicationWorkIDs = ref<number[]>([]);
const applicationBusy = ref(false);
const agentTokenBusy = ref(false);
const agentTokenName = ref("");
const agentTokenScopes = ref<string[]>([]);
const issuedAgentToken = ref("");
const editingWorkId = ref<number | null>(null);
const editingOpportunityId = ref<number | null>(null);
const exchange = ref<Record<number, { applicant: Record<string, string>; recipient: Record<string, string> }>>({});
const exchangeErrors = ref<Record<number, string>>({});
const contacts = reactive({ email: "", wechat: "", phone: "", share_email: false, share_wechat: false, share_phone: false });
const applicationConsent = reactive({ share_email: false, share_wechat: false, share_phone: false });
const agentScopeGroups = [
  {
    title: "个人资料",
    scopes: [
      { value: "network_profiles:read", label: "读取个人资料", description: "读取你自己的公开资料" },
      { value: "network_profiles:write", label: "编辑个人资料", description: "更新你的公开资料与档期" }
    ]
  },
  {
    title: "作品",
    scopes: [
      { value: "network_works:read", label: "读取作品", description: "查看你发布的作品" },
      { value: "network_works:write", label: "管理作品", description: "创建、编辑或下架你的作品" }
    ]
  },
  {
    title: "项目",
    scopes: [
      { value: "network_projects:read", label: "读取项目", description: "读取你管理的项目及关联机会" },
      { value: "network_projects:write", label: "管理项目", description: "创建、编辑项目计划与团队成员" }
    ]
  },
  {
    title: "合作机会",
    scopes: [
      { value: "network_opportunities:read", label: "读取机会", description: "查看你发布的合作机会" },
      { value: "network_opportunities:write", label: "管理机会", description: "创建、编辑或更新机会状态" }
    ]
  },
  {
    title: "联系申请",
    scopes: [
      { value: "network_contact_requests:read", label: "读取申请", description: "查看你收到和发出的申请" },
      { value: "network_contact_requests:write", label: "管理申请", description: "发起申请或撤回待处理申请" }
    ]
  },
  {
    title: "联系方式",
    scopes: [
      { value: "network_contact_exchange:read", label: "读取联系方式", description: "读取自己的联系方式及已授权交换结果" },
      { value: "network_contact_exchange:write", label: "管理联系方式授权", description: "更新自己的联系方式与授权，或撤回已建立的交换" }
    ]
  }
];

const registerForm = reactive({
  invite_code: "",
  email: "",
  password: "",
  display_name: "",
  primary_role: "",
  headline: ""
});
const loginForm = reactive({ email: "", password: "" });
const profileForm = reactive({
  display_name: "",
  primary_role: "",
  headline: "",
  bio: "",
  city: "",
  skills: "",
  collaboration_modes: "",
  availability: "open",
  available_from: ""
});
const workForm = reactive({
  title: "",
  summary: "",
  body: "",
  cover_url: "",
  demo_url: "",
  repo_url: "",
  tags: "",
  status: "published"
});
type OpportunityMilestoneForm = {
  title: string;
  description: string;
  status: "planned" | "in_progress" | "completed" | "blocked";
  owner_member_ids: number[];
  timing_kind: "fixed_dates" | "duration" | "relative" | "flexible" | "unspecified";
  timing_label: string;
  timing_start_at: string;
  timing_end_at: string;
  timing_duration_value: number;
  timing_duration_unit: string;
  goals_text: string;
  deliverables_text: string;
  acceptance_criteria_text: string;
};
type OpportunitySectionForm = NetworkProjectPlanSection & { items_text: string };
const opportunityForm = reactive({
  title: "",
  summary: "",
  opportunity_type: "collaboration",
  looking_for_roles: "",
  required_skills: "",
  engagement_mode: "项目合作",
  budget_hint: "",
  duration_hint: "",
  city_scope: "",
  remote_allowed: true,
  status: "open",
  expires_at: "",
  team_intro: "",
  team_members: [] as NetworkOpportunityMember[],
  milestones: [] as OpportunityMilestoneForm[],
  plan_sections: [] as OpportunitySectionForm[]
});

function newMilestone(): OpportunityMilestoneForm {
  return {
    title: "",
    description: "",
    status: "planned",
    owner_member_ids: [],
    timing_kind: "unspecified",
    timing_label: "",
    timing_start_at: "",
    timing_end_at: "",
    timing_duration_value: 1,
    timing_duration_unit: "",
    goals_text: "",
    deliverables_text: "",
    acceptance_criteria_text: ""
  };
}

function newPlanSection(): OpportunitySectionForm {
  return { type: "context", title: "", body: "", items: [], items_text: "", sort_order: 0 };
}

function splitFormItems(value: string) {
  return value.split(/\n/).map((item) => item.trim()).filter(Boolean);
}

function addMilestone() {
  opportunityForm.milestones.push(newMilestone());
}

function addPlanSection() {
  opportunityForm.plan_sections.push(newPlanSection());
}

function addOpportunityTeamMember() {
  const available = profiles.value.find((profile) =>
    !opportunityForm.team_members.some((member) => member.profile_id === profile.id)
  );
  if (!available) return;
  opportunityForm.team_members.push({
    profile_id: available.id,
    role: available.primary_role || "",
    contribution: "",
    sort_order: opportunityForm.team_members.length
  });
}

function buildProjectPlan() {
  const milestones = opportunityForm.milestones
    .filter((milestone) => milestone.title.trim())
    .map((milestone, index) => {
      const timing = milestone.timing_kind === "unspecified"
        ? undefined
        : milestone.timing_kind === "fixed_dates"
          ? { kind: milestone.timing_kind, start_at: milestone.timing_start_at, end_at: milestone.timing_end_at, label: milestone.timing_label }
          : milestone.timing_kind === "duration"
            ? { kind: milestone.timing_kind, duration_value: milestone.timing_duration_value, duration_unit: milestone.timing_duration_unit, label: milestone.timing_label }
            : { kind: milestone.timing_kind, label: milestone.timing_label };
      return {
        title: milestone.title.trim(),
        description: milestone.description.trim(),
        order: index,
        status: milestone.status,
        ...(timing ? { timing } : {}),
        goals: splitFormItems(milestone.goals_text),
        deliverables: splitFormItems(milestone.deliverables_text),
        acceptance_criteria: splitFormItems(milestone.acceptance_criteria_text),
        owner_member_ids: milestone.owner_member_ids.filter((profileID) =>
          opportunityForm.team_members.some((member) => member.profile_id === profileID)
        )
      };
    });
  const sections = opportunityForm.plan_sections
    .filter((section) => section.title.trim())
    .map((section, index) => ({
      type: section.type,
      title: section.title.trim(),
      body: section.body?.trim() || "",
      items: splitFormItems(section.items_text),
      sort_order: index
    }));
  return milestones.length || sections.length ? { milestones, sections } : null;
}

const sectionTitles: Record<string, { eyebrow: string; title: string; desc: string }> = {
  discover: { eyebrow: "BUILDER NETWORK", title: "让作品成为连接的起点", desc: "在这里发现正在创造的人，也把你正在做的事带到更多人面前。" },
  builders: { eyebrow: "PEOPLE", title: "找 Builder", desc: "通过真实作品、技能和档期，找到适合一起做事的人。" },
  projects: { eyebrow: "OPPORTUNITIES", title: "找项目", desc: "浏览正在招募的合作机会，用清晰的理由发起联系。" },
  publish: { eyebrow: "PUBLISH", title: "发布合作需求", desc: "讲清目标、需要的角色和合作方式，让合适的人主动找到你。" },
  workspace: { eyebrow: "YOUR SPACE", title: "个人空间", desc: "展示你的 Builder 身份，查看合作动态，并管理正在进行的事。" },
  register: { eyebrow: "INVITE ONLY", title: "加入 BuilderNetwork", desc: "当前为邀请制社区。使用有效邀请码创建你的市场身份。" },
  login: { eyebrow: "WELCOME BACK", title: "登录市场账户", desc: "使用 BuilderNetwork 账户继续你的协作。" }
};
const devMockOpportunities: NetworkOpportunity[] = [
  {
    id: 1001,
    is_demo: true,
    profile_id: 1,
    slug: "agent-workflow-platform-mvp",
    title: "AI 工作流平台 MVP：需要一位 Vue 工程化伙伴",
    summary: "把“内容采集-审核-发布-分发”做成可复用工作流，目标是 2 周内打通闭环并上线内测。",
    opportunity_type: "collaboration",
    created_at: "2026-10-20T09:00:00Z",
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
        role: "发起人 / 产品工程",
        contribution: "定义 MVP 边界、产品结构、核心流程与验收标准",
        sort_order: 0
      },
      {
        profile_id: 2,
        role: "交互 / 信息架构",
        contribution: "把流程拆成可复用的信息结构，确保关键页面可交付",
        sort_order: 1
      },
      {
        profile_id: 3,
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
          body: "我们希望把资讯运营的重复工作流产品化，让内容生产更可控、可复用、可追踪。",
          items: ["采集来源可配置", "审核需要明确状态流转", "发布与分发需要可回放"],
          sort_order: 0
        },
        {
          type: "scope",
          title: "范围与边界（MVP）",
          body: "",
          items: ["先跑通 1 套工作流模板", "先支持 1 个分发渠道", "先保证可用，不追求全自动智能化"],
          sort_order: 1
        },
        {
          type: "execution",
          title: "执行方式",
          body: "",
          items: ["前端按组件模块拆分交付", "每个模块先可用再迭代", "关键节点做一次联调演示"],
          sort_order: 2
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
    created_at: "2026-10-18T09:00:00Z",
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
    created_at: "2026-10-16T09:00:00Z",
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
    created_at: "2026-10-14T09:00:00Z",
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
    created_at: "2026-10-12T09:00:00Z",
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
const pageMeta = computed(() => section.value === "publish" && editingOpportunityId.value
  ? { eyebrow: "EDIT OPPORTUNITY", title: "编辑合作需求", desc: "更新需求内容、截止日期和当前招募状态。" }
  : sectionTitles[section.value] || sectionTitles.discover);
watch(() => registerForm.invite_code, () => {
  inviteValid.value = false;
  notice.value = "";
});
watch(() => route.query.invite, (invite) => {
  if (typeof invite !== "string" || !invite.trim()) return;
  registerForm.invite_code = invite.trim();
  void verifyInvite();
}, { immediate: true });

async function loadMarketData() {
  if (!["discover", "builders", "projects"].includes(section.value)) return;
  loading.value = true;
  error.value = "";
  try {
    const [nextProfiles, nextWorks] = await Promise.all([
      networkRequest<NetworkProfile[]>("/api/v1/public/network/profiles"),
      networkRequest<NetworkWork[]>("/api/v1/public/network/works")
    ]);
    let nextProjects: NetworkProject[];
    try {
      nextProjects = await networkRequest<NetworkProject[]>("/api/v1/public/network/projects");
    } catch (cause) {
      if (!import.meta.env.DEV) throw cause;
      const legacyOpportunities = await networkRequest<NetworkOpportunity[]>("/api/v1/public/network/opportunities");
      nextProjects = legacyOpportunities.map(projectFromLegacyOpportunity);
    }
    profiles.value = nextProfiles;
    works.value = nextWorks;
    const liveProjects = nextProjects.map((project) => ({
      ...project,
      opportunities: project.opportunities || []
    }));
    const demoProjects: NetworkProject[] = import.meta.env.DEV
      ? devMockOpportunities
          .filter((mock) => !liveProjects.some((item) => item.slug === mock.slug))
          .map((mock) => {
            const projectID = mock.id + 10000;
            return {
              id: projectID,
              is_demo: true,
              owner_profile_id: mock.profile_id,
              slug: mock.slug,
              title: mock.title,
              summary: mock.summary,
              status: "published",
              project_plan: mock.project_plan,
              team_intro: mock.team_intro || "",
              owner_name: mock.owner_name,
              members: mock.team_members || [],
              opportunities: [{ ...mock, project_id: projectID }],
              created_at: mock.created_at || "",
              updated_at: mock.updated_at || mock.created_at || ""
            };
          })
      : [];
    projects.value = [...liveProjects, ...demoProjects].filter((project) => project.opportunities.length > 0);
    opportunities.value = projects.value.flatMap((project) => project.opportunities);
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "市场内容加载失败";
  } finally {
    loading.value = false;
  }
}

async function loadWorkspace() {
  error.value = "";
  if (!getNetworkToken()) {
    profile.value = null;
    return;
  }
  try {
    const results = await Promise.allSettled([
      networkRequest<NetworkProfile>("/api/v1/network/me", {}, true),
      networkRequest<NetworkContactRequest[]>("/api/v1/network/contact-requests?direction=incoming", {}, true),
      networkRequest<NetworkContactRequest[]>("/api/v1/network/contact-requests?direction=outgoing", {}, true),
      networkRequest<{
        contacts: { email: string; wechat: string; phone: string };
        share_email: boolean;
        share_wechat: boolean;
        share_phone: boolean;
      }>("/api/v1/network/me/contacts", {}, true),
      networkRequest<NetworkWork[]>("/api/v1/network/works", {}, true),
      networkRequest<NetworkOpportunity[]>("/api/v1/network/opportunities", {}, true),
      networkRequest<NetworkAgentToken[]>("/api/v1/network/agent-tokens", {}, true)
    ]);
    const [profileResult, incomingResult, outgoingResult, contactsResult, worksResult, opportunitiesResult, tokensResult] = results;
    if (profileResult.status === "rejected") {
      const cause = profileResult.reason;
      if (cause instanceof NetworkRequestError && [401, 403].includes(cause.status)) {
        clearNetworkToken();
        isAuthed.value = false;
        profile.value = null;
        error.value = "登录状态已失效，请重新登录。";
        return;
      }
      throw cause;
    }

    const nextProfile = profileResult.value;
    profile.value = nextProfile;
    if (incomingResult.status === "fulfilled") {
      incoming.value = incomingResult.value;
      if (import.meta.env.DEV && nextProfile.display_name === "Local Builder" && !incoming.value.length) {
        incoming.value = devMockIncomingRequests(nextProfile.id);
      }
    }
    if (outgoingResult.status === "fulfilled") {
      outgoing.value = outgoingResult.value;
      if (import.meta.env.DEV && nextProfile.display_name === "Local Builder" && !outgoing.value.length) {
        outgoing.value = devMockOutgoingRequests(nextProfile.id, nextProfile.display_name);
      }
    }
    if (worksResult.status === "fulfilled") myWorks.value = worksResult.value;
    if (opportunitiesResult.status === "fulfilled") myOpportunities.value = opportunitiesResult.value;
    if (tokensResult.status === "fulfilled") agentTokens.value = tokensResult.value;
    Object.assign(profileForm, {
      display_name: nextProfile.display_name,
      primary_role: nextProfile.primary_role,
      headline: nextProfile.headline,
      bio: nextProfile.bio,
      city: nextProfile.city,
      skills: (nextProfile.skills || []).join(", "),
      collaboration_modes: (nextProfile.collaboration_modes || []).join(", "),
      availability: nextProfile.availability,
      available_from: nextProfile.available_from
    });
    if (contactsResult.status === "fulfilled") {
      const contactData = contactsResult.value;
      Object.assign(contacts, contactData.contacts, {
        share_email: contactData.share_email,
        share_wechat: contactData.share_wechat,
        share_phone: contactData.share_phone
      });
    }

    const acceptedRequests = [...incoming.value, ...outgoing.value].filter(
      (request) => request.status === "accepted" && !request.exchange_revoked && !request.is_demo
    );
    const acceptedRequestIDs = new Set(acceptedRequests.map((request) => request.id));
    exchange.value = Object.fromEntries(
      Object.entries(exchange.value).filter(([id]) => acceptedRequestIDs.has(Number(id)))
    );
    exchangeErrors.value = Object.fromEntries(
      Object.entries(exchangeErrors.value).filter(([id]) => acceptedRequestIDs.has(Number(id)))
    );
    const exchangeResults = await Promise.allSettled(
      acceptedRequests.map((request) =>
        networkRequest<{ applicant: Record<string, string>; recipient: Record<string, string> }>(
          `/api/v1/network/contact-exchanges/${request.id}`,
          {},
          true
        )
      )
    );
    exchangeResults.forEach((result, index) => {
      const requestID = acceptedRequests[index].id;
      if (result.status === "fulfilled") {
        exchange.value[requestID] = result.value;
        delete exchangeErrors.value[requestID];
      } else {
        exchangeErrors.value[requestID] = "联系方式暂时无法读取，请稍后刷新重试。";
      }
    });

    const unavailableModules = [
      incomingResult.status === "rejected" ? "收到的申请" : "",
      outgoingResult.status === "rejected" ? "发出的申请" : "",
      contactsResult.status === "rejected" ? "联系方式授权" : "",
      worksResult.status === "rejected" ? "作品" : "",
      opportunitiesResult.status === "rejected" ? "合作机会" : "",
      tokensResult.status === "rejected" ? "Agent Token" : ""
    ].filter(Boolean);
    if (unavailableModules.length) {
      warning.value = `登录成功，但以下模块暂时不可用：${unavailableModules.join("、")}。`;
    } else {
      notice.value = "登录成功，欢迎回来。";
    }
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "工作台加载失败";
  }
}

async function loadForSection() {
  notice.value = "";
  error.value = "";
  warning.value = "";
  if (["discover", "builders", "projects"].includes(section.value)) {
    await loadMarketData();
  } else if (section.value === "publish") {
    try {
      profiles.value = await networkRequest<NetworkProfile[]>("/api/v1/public/network/profiles");
    } catch (cause) {
      warning.value = cause instanceof Error ? `公开 Builder 列表暂不可用：${cause.message}` : "公开 Builder 列表暂不可用。";
    }
  } else if (section.value === "workspace") {
    await loadWorkspace();
  }
}

watch(section, () => void loadForSection());
onMounted(() => void loadForSection());

async function verifyInvite() {
  error.value = "";
  try {
    const result = await postJSON<{ valid: boolean }>("/api/v1/auth/invitations/verify", { code: registerForm.invite_code });
    inviteValid.value = result.valid;
    notice.value = result.valid ? "邀请码有效，可以继续填写注册信息。" : "邀请码无效、已用完或已过期。";
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "邀请码校验失败";
  }
}

async function register() {
  if (!inviteValid.value) {
    error.value = "请先验证有效邀请码。";
    return;
  }
  error.value = "";
  try {
    const session = await postJSON<NetworkSession>("/api/v1/auth/register", {
      invite_code: registerForm.invite_code,
      email: registerForm.email,
      password: registerForm.password,
      profile: {
        display_name: registerForm.display_name,
        primary_role: registerForm.primary_role,
        headline: registerForm.headline
      }
    });
    saveNetworkSession(session);
    isAuthed.value = true;
    await router.push("/market/workspace");
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "注册失败";
  }
}

async function login() {
  error.value = "";
  try {
    const session = await postJSON<NetworkSession>("/api/v1/auth/network/login", loginForm);
    saveNetworkSession(session);
    isAuthed.value = true;
    await router.push("/market/workspace");
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "登录失败";
  }
}

async function publishOpportunity() {
  if (!getNetworkToken()) {
    await router.push("/market/login");
    return;
  }
  error.value = "";
  try {
    const payload = {
      ...opportunityForm,
      looking_for_roles: opportunityForm.looking_for_roles.split(/[,，]/).map((item) => item.trim()).filter(Boolean),
      required_skills: opportunityForm.required_skills.split(/[,，]/).map((item) => item.trim()).filter(Boolean),
      expires_at: opportunityForm.expires_at ? new Date(`${opportunityForm.expires_at}T23:59:59`).toISOString() : "",
      project_plan: buildProjectPlan(),
      team_intro: opportunityForm.team_intro.trim(),
      team_members: opportunityForm.team_members.map((member, index) => ({ ...member, sort_order: index }))
    };
    if (editingOpportunityId.value) {
      await networkRequest(`/api/v1/network/opportunities/${editingOpportunityId.value}`, {
        method: "PUT",
        body: JSON.stringify(payload)
      }, true);
      editingOpportunityId.value = null;
      await loadWorkspace();
      notice.value = "合作机会已更新。";
    } else {
      await postJSON("/api/v1/network/opportunities", payload, true);
      await router.push("/market/projects");
    }
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "发布失败";
  }
}

async function saveProfile() {
  error.value = "";
  try {
    profile.value = await networkRequest<NetworkProfile>("/api/v1/network/me", {
      method: "PUT",
      body: JSON.stringify({
        ...profileForm,
        skills: profileForm.skills.split(/[,，]/).map((item) => item.trim()).filter(Boolean),
        collaboration_modes: profileForm.collaboration_modes.split(/[,，]/).map((item) => item.trim()).filter(Boolean)
      })
    }, true);
    notice.value = "个人资料已更新。";
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "保存资料失败";
  }
}

async function saveWork() {
  if (!getNetworkToken()) {
    await router.push("/market/login");
    return;
  }
  error.value = "";
  try {
    const payload = {
      ...workForm,
      tags: workForm.tags.split(/[,，]/).map((item) => item.trim()).filter(Boolean)
    };
    if (editingWorkId.value) {
      await networkRequest(`/api/v1/network/works/${editingWorkId.value}`, {
        method: "PUT",
        body: JSON.stringify(payload)
      }, true);
      notice.value = "作品已更新。";
    } else {
      await postJSON("/api/v1/network/works", payload, true);
      notice.value = "作品已发布到发现广场。";
    }
    editingWorkId.value = null;
    Object.assign(workForm, { title: "", summary: "", body: "", cover_url: "", demo_url: "", repo_url: "", tags: "", status: "published" });
    await loadWorkspace();
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "作品发布失败";
  }
}

async function openApplication(type: "profile" | "opportunity", id: number, title: string) {
  if (!getNetworkToken()) {
    await router.push("/market/login");
    return;
  }
  error.value = "";
  try {
    const [contactData, ownWorks] = await Promise.all([
      networkRequest<{
        contacts: { email: string; wechat: string; phone: string };
        share_email: boolean;
        share_wechat: boolean;
        share_phone: boolean;
      }>("/api/v1/network/me/contacts", {}, true),
      networkRequest<NetworkWork[]>("/api/v1/network/works", {}, true)
    ]);
    myWorks.value = ownWorks;
    Object.assign(contacts, contactData.contacts, {
      share_email: contactData.share_email,
      share_wechat: contactData.share_wechat,
      share_phone: contactData.share_phone
    });
    Object.assign(applicationConsent, { share_email: false, share_wechat: false, share_phone: false });
    applicationWorkIDs.value = [];
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "读取授权设置失败";
    return;
  }
  selectedTarget.value = { type, id, title };
  applicationMessage.value = "";
}

async function submitApplication() {
  if (!selectedTarget.value) return;
  applicationBusy.value = true;
  error.value = "";
  try {
    await postJSON("/api/v1/network/contact-requests", {
      target_type: selectedTarget.value.type,
      target_id: selectedTarget.value.id,
      message: applicationMessage.value,
      related_work_ids: applicationWorkIDs.value,
      ...applicationConsent
    }, true);
    selectedTarget.value = null;
    notice.value = "申请已发送。对方接受后，双方才能查看已授权的联系方式。";
    if (section.value === "workspace") await loadWorkspace();
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "申请发送失败";
  } finally {
    applicationBusy.value = false;
  }
}

function limitRelatedWorks() {
  if (applicationWorkIDs.value.length > 3) {
    applicationWorkIDs.value = applicationWorkIDs.value.slice(0, 3);
    notice.value = "一次申请最多关联 3 个作品。";
  }
}

async function processRequest(id: number, action: "accept" | "decline" | "withdraw") {
  error.value = "";
  const demoRequest = [...incoming.value, ...outgoing.value].find((request) => request.id === id && request.is_demo);
  if (demoRequest) {
    demoRequest.status = action === "withdraw" ? "withdrawn" : action === "accept" ? "accepted" : "declined";
    notice.value = action === "withdraw"
      ? "演示申请已在本地预览中撤回。"
      : action === "accept"
        ? "演示申请已在本地预览中接受。"
        : "演示申请已在本地预览中婉拒。";
    return;
  }
  try {
    await postJSON(`/api/v1/network/contact-requests/${id}/${action}`, {}, true);
    await loadWorkspace();
    if (action === "accept") notice.value = "已同意申请。双方授权的联系方式已显示在申请中。";
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "申请处理失败";
  }
}

async function showExchange(id: number) {
  delete exchangeErrors.value[id];
  try {
    exchange.value[id] = await networkRequest(`/api/v1/network/contact-exchanges/${id}`, {}, true);
  } catch (cause) {
    exchangeErrors.value[id] = cause instanceof Error ? cause.message : "联系方式读取失败";
  }
}

async function revokeExchange(id: number) {
  error.value = "";
  try {
    await postJSON(`/api/v1/network/contact-exchanges/${id}/revoke`, {}, true);
    delete exchange.value[id];
    await loadWorkspace();
    notice.value = "本次联系方式交换已撤回，双方之后无法继续从平台查看。已复制的信息无法远程收回。";
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "撤回授权失败";
  }
}

function editWork(work: NetworkWork) {
  editingWorkId.value = work.id;
  Object.assign(workForm, {
    title: work.title,
    summary: work.summary,
    body: work.body,
    cover_url: work.cover_url,
    demo_url: work.demo_url,
    repo_url: work.repo_url,
    tags: (work.tags || []).join(", "),
    status: work.status
  });
}

async function updateWorkStatus(work: NetworkWork, status: string) {
  error.value = "";
  try {
    await networkRequest(`/api/v1/network/works/${work.id}`, {
      method: "PUT",
      body: JSON.stringify({ ...work, status })
    }, true);
    await loadWorkspace();
    notice.value = status === "archived" ? "作品已下架，可随时重新发布。" : "作品状态已更新。";
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "作品状态更新失败";
  }
}

function editOpportunity(opportunity: NetworkOpportunity) {
  editingOpportunityId.value = opportunity.id;
  Object.assign(opportunityForm, {
    title: opportunity.title,
    summary: opportunity.summary,
    opportunity_type: opportunity.opportunity_type,
    looking_for_roles: (opportunity.looking_for_roles || []).join(", "),
    required_skills: (opportunity.required_skills || []).join(", "),
    engagement_mode: opportunity.engagement_mode,
    budget_hint: opportunity.budget_hint,
    duration_hint: opportunity.duration_hint,
    city_scope: opportunity.city_scope,
    remote_allowed: opportunity.remote_allowed,
    status: opportunity.status,
    expires_at: opportunity.expires_at ? opportunity.expires_at.slice(0, 10) : "",
    team_intro: opportunity.team_intro || "",
    team_members: (opportunity.team_members || []).map((member) => ({ ...member })),
    milestones: (opportunity.project_plan?.milestones || []).map((milestone) => ({
      title: milestone.title,
      description: milestone.description,
      status: milestone.status || "planned",
      owner_member_ids: milestone.owner_member_ids || [],
      timing_kind: milestone.timing?.kind || "unspecified",
      timing_label: milestone.timing?.label || "",
      timing_start_at: milestone.timing?.start_at || "",
      timing_end_at: milestone.timing?.end_at || "",
      timing_duration_value: milestone.timing?.duration_value || 1,
      timing_duration_unit: milestone.timing?.duration_unit || "",
      goals_text: (milestone.goals || []).join("\n"),
      deliverables_text: (milestone.deliverables || []).join("\n"),
      acceptance_criteria_text: (milestone.acceptance_criteria || []).join("\n")
    })),
    plan_sections: (opportunity.project_plan?.sections || []).map((section) => ({
      ...section,
      items_text: (section.items || []).join("\n")
    }))
  });
  void router.push("/market/publish");
}

async function updateOpportunityStatus(opportunity: NetworkOpportunity, status: string) {
  error.value = "";
  try {
    await networkRequest(`/api/v1/network/opportunities/${opportunity.id}`, {
      method: "PUT",
      body: JSON.stringify({ ...opportunity, status })
    }, true);
    await loadWorkspace();
    notice.value = "合作机会状态已更新。";
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "合作机会状态更新失败";
  }
}

async function saveContacts() {
  error.value = "";
  try {
    await networkRequest("/api/v1/network/me/contacts", {
      method: "PUT",
      body: JSON.stringify({
        contacts: { email: contacts.email, wechat: contacts.wechat, phone: contacts.phone },
        share_email: contacts.share_email,
        share_wechat: contacts.share_wechat,
        share_phone: contacts.share_phone
      })
    }, true);
    notice.value = "联系方式与交换授权已保存。";
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "保存失败";
  }
}

async function createAgentToken() {
  if (!agentTokenName.value.trim()) {
    error.value = "请填写 Agent 名称。";
    return;
  }
  if (!agentTokenScopes.value.length) {
    error.value = "至少选择一个权限范围。";
    return;
  }
  agentTokenBusy.value = true;
  error.value = "";
  try {
    const result = await postJSON<NetworkAgentTokenCreateResult>("/api/v1/network/agent-tokens", {
      name: agentTokenName.value,
      scopes: agentTokenScopes.value
    }, true);
    issuedAgentToken.value = result.token;
    agentTokenName.value = "";
    agentTokenScopes.value = [];
    await loadWorkspace();
    notice.value = "Token 已创建。明文只展示这一次，请立即复制并妥善保存。";
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "Agent Token 创建失败";
  } finally {
    agentTokenBusy.value = false;
  }
}

async function updateAgentTokenStatus(token: NetworkAgentToken) {
  error.value = "";
  try {
    const updated = await networkRequest<NetworkAgentToken>(`/api/v1/network/agent-tokens/${token.id}/status`, {
      method: "PUT",
      body: JSON.stringify({ is_active: !token.is_active })
    }, true);
    agentTokens.value = agentTokens.value.map((item) => item.id === updated.id ? updated : item);
    notice.value = updated.is_active ? "Token 已重新启用。" : "Token 已停用，新请求会立即失效。";
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "Token 状态更新失败";
  }
}

async function copyIssuedAgentToken() {
  try {
    await navigator.clipboard.writeText(issuedAgentToken.value);
    notice.value = "Token 已复制。请勿将它发送给不可信的人。";
  } catch {
    error.value = "无法访问剪贴板，请手动复制上方 Token。";
  }
}

function logout() {
  clearNetworkToken();
  isAuthed.value = false;
  profile.value = null;
  agentTokens.value = [];
  issuedAgentToken.value = "";
  void router.push("/market");
}
</script>

<template>
  <main class="market-page">
    <section v-if="isAuthPage" class="auth-split-layout">
      <aside class="auth-visual-side">
        <FaultyTerminal />
        <div class="terminal-overlay"></div>
        <div class="terminal-content">
          <RouterLink class="brand-logo" to="/market"><img class="market-logo" :src="marketLogo" alt="VOIDLAB Logo" width="32" height="32" /><span>VOIDLAB <b>NETWORK</b></span></RouterLink>
          <div class="terminal-text-container">
            <p class="auth-kicker">BUILD OPENLY. CONNECT INTENTIONALLY.</p>
            <h1>让作品被看见。<br />让合作发生。</h1>
            <p class="auth-brand-desc">在这里展示你的能力，<br />找到值得一起创造的伙伴。</p>
          </div>
          <div class="auth-visual-footer"><span>BUILDER NETWORK</span><span>人 · 作品 · 合作</span></div>
        </div>
      </aside>
      <div class="auth-form-side">
        <div class="auth-form-container">
          <RouterLink class="auth-back" to="/market">← 返回市场</RouterLink>
          <nav class="auth-tabs" aria-label="账户入口">
            <RouterLink to="/market/login" :class="{ active: section === 'login' }">登录</RouterLink>
            <RouterLink to="/market/register" :class="{ active: section === 'register' }">邀请码注册</RouterLink>
          </nav>
          <h2>{{ section === "login" ? "欢迎回来" : "创建你的 Builder 身份" }}</h2>
          <p class="auth-intro">{{ section === "login" ? "登录后，继续你的作品与合作。" : "社区当前为邀请制，请使用运营提供的邀请码。" }}</p>
          <p v-if="error" class="feedback error-feedback" role="alert">{{ error }}</p>
          <p v-if="notice" class="feedback success-feedback" role="status">{{ notice }}</p>
          <p v-if="warning" class="feedback warning-feedback" role="status">{{ warning }}</p>
          <form v-if="section === 'login'" class="clean-form" @submit.prevent="login">
            <label class="clean-form-group">邮箱<input v-model="loginForm.email" type="email" required autocomplete="email" placeholder="you@example.com" /></label>
            <label class="clean-form-group">密码<input v-model="loginForm.password" type="password" required autocomplete="current-password" placeholder="输入你的密码" /></label>
            <button class="button-primary clean-submit" type="submit">登录 BuilderNetwork</button>
            <p class="auth-form-note">还没有账户？ <RouterLink to="/market/register">使用邀请码加入</RouterLink></p>
          </form>
          <form v-else class="clean-form" @submit.prevent="register">
            <div class="auth-invite-row">
              <label class="clean-form-group">邀请码<input v-model="registerForm.invite_code" required autocomplete="off" placeholder="输入专属邀请码" /></label>
              <button class="auth-verify" type="button" @click="verifyInvite">{{ inviteValid ? "已验证" : "验证邀请码" }}</button>
            </div>
            <label class="clean-form-group">邮箱<input v-model="registerForm.email" type="email" required autocomplete="email" placeholder="you@example.com" /></label>
            <label class="clean-form-group">密码<input v-model="registerForm.password" type="password" required minlength="10" autocomplete="new-password" placeholder="至少 10 位" /></label>
            <div class="clean-form-row">
              <label class="clean-form-group">姓名 / 常用称呼<input v-model="registerForm.display_name" required placeholder="如何称呼你" /></label>
              <label class="clean-form-group">核心职业身份<input v-model="registerForm.primary_role" required placeholder="例如：全栈开发者" /></label>
            </div>
            <label class="clean-form-group">一句话介绍<input v-model="registerForm.headline" placeholder="你正在做什么，或希望和谁合作" /></label>
            <p class="auth-form-note">联系方式默认私密，仅在双方同意后按授权交换。</p>
            <button class="button-primary clean-submit" type="submit">创建市场账户</button>
          </form>
        </div>
      </div>
    </section>
  <div v-if="section === 'discover'" class="hero-full-light">
    <!-- Fullscreen Hero Background -->
    <div class="hero-full-bg-waves">
      <ShapeWaves
        text=""
        :text-size="0.56"
        color="#777777"
        hover-color="#0BCB78"
        background-color="#09090b"
        :dot-size="0.78"
        :cell-size="12"
        :fade="0"
        :glow="0.28"
        :speed="0.8"
        :flow="0.45"
        :interactive="true"
        shapes="mixed"
      />
    </div>

    <!-- Integrated Header for Hero -->
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

    <main class="hero-brand-lockup">
      <h1>Builder<wbr />Network</h1>
      <p>VOIDLAB</p>
    </main>
  </div>

  <div
    v-if="!isAuthPage"
    class="market-shell"
    :class="{
      'is-discover-shell': section === 'discover',
      'is-projects-shell': section === 'projects',
      'is-builders-shell': section === 'builders',
      'is-workspace-shell': section === 'workspace'
    }"
  >

    <template v-if="section !== 'discover'">
      <header class="market-nav-full market-nav-solid">
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
      <section class="page-heading">
        <p class="eyebrow">{{ pageMeta.eyebrow }}</p>
        <h1>{{ pageMeta.title }}</h1>
        <p class="page-desc">{{ pageMeta.desc }}</p>
      </section>

      <p v-if="error" class="feedback error-feedback" role="alert">{{ error }}</p>
      <p v-if="notice" class="feedback success-feedback" role="status">{{ notice }}</p>
      <p v-if="warning" class="feedback warning-feedback" role="status">{{ warning }}</p>
    </template>

    <template v-if="section === 'discover'">
      <div class="discover-content-start"></div>

        <section class="market-intro-panel" aria-label="项目介绍">
          <div class="market-intro-inner">
            <div class="market-intro-copy">
              <p class="eyebrow">ABOUT VOIDLAB</p>
              <h2>一个邀请制的<br />Builder 协作网络</h2>
              <p class="market-intro-lead">BuilderNetwork 让作品成为身份，让合作申请带着理由发生。你可以在网页工作台管理资料、作品与合作机会，也可以把市场能力装进你的 Agent。</p>
            </div>
            <div class="market-intro-points">
              <div class="market-intro-point">
                <strong>邀请制加入</strong>
                <p>只需一条邀请码，创建公开身份，开始被发现。</p>
              </div>
              <div class="market-intro-point">
                <strong>隐私优先</strong>
                <p>联系方式默认私密，必须在双方同意后才会交换。</p>
              </div>
              <div class="market-intro-point">
                <strong>Agent 优先</strong>
                <p>通过 Skill / API 把发现、管理与申请处理接入你的日常工作流。</p>
              </div>
            </div>
          </div>
        </section>

        <section class="skill-download-panel">
          <div class="skill-panel-text">
            <p class="eyebrow">AGENT SKILL PACK</p>
            <h2>把市场装进<br />你的 <span>Agent</span></h2>
            <p>安装 BuilderNetwork Skill，把发现合作机会、管理作品与协作申请等能力接入你的日常 Agent 工作流。</p>

            <div v-if="selectedSkillDownload" class="skill-selected-download" aria-live="polite">
              <div class="skill-selected-meta">
                <span>SELECTED PACKAGE</span>
                <span>{{ selectedSkillDownload.kind }}</span>
              </div>
              <strong>{{ selectedSkillDownload.label }}</strong>
              <p>{{ selectedSkillDownload.description }}</p>
              <a class="skill-download-cta" :href="selectedSkillDownload.href" download>
                <span>下载所选文件</span>
                <span aria-hidden="true">↓</span>
              </a>
            </div>
          </div>

          <nav class="skill-branch-menu" aria-label="选择要下载的 BuilderNetwork Skill">
            <div v-for="(group, groupIndex) in skillPackGroups" :key="group.value" class="skill-branch-group">
              <button
                class="skill-branch-heading"
                type="button"
                :aria-expanded="isSkillGroupExpanded(group.value)"
                @click="toggleSkillGroup(group.value)"
              >
                <span class="skill-branch-index">0{{ groupIndex + 1 }}</span>
                <span>{{ group.label }}</span>
                <span class="skill-branch-toggle" aria-hidden="true">{{ isSkillGroupExpanded(group.value) ? "−" : "+" }}</span>
              </button>

              <div v-if="isSkillGroupExpanded(group.value)" class="skill-branch-children">
                <svg
                  class="skill-branch-lines"
                  width="32"
                  :height="group.items.length * 58"
                  :viewBox="`0 0 32 ${group.items.length * 58}`"
                  preserveAspectRatio="none"
                  aria-hidden="true"
                >
                  <path class="skill-branch-trunk" :d="skillTrunkPath(group.items.length)" />
                  <path
                    v-for="(item, itemIndex) in group.items"
                    :key="item.value"
                    class="skill-branch-path"
                    :class="{ 'is-active': selectedSkillValue === item.value }"
                    :d="skillBranchPath(itemIndex)"
                  />
                </svg>
                <button
                  v-for="item in group.items"
                  :key="item.value"
                  class="skill-branch-item"
                  :class="{ 'is-active': selectedSkillValue === item.value }"
                  type="button"
                  :aria-pressed="selectedSkillValue === item.value"
                  @click="selectedSkillValue = item.value"
                >
                  <span class="skill-branch-copy">
                    <strong>{{ item.label }}</strong>
                    <small>{{ item.description }}</small>
                  </span>
                  <span class="skill-branch-kind">{{ item.kind }}</span>
                </button>
              </div>
            </div>
          </nav>
        </section>

        <section class="popular-projects-panel">
          <div class="popular-projects-inner">
            <div class="popular-projects-heading">
              <div><p class="eyebrow">POPULAR PROJECTS</p><h2>热门项目</h2></div>
              <RouterLink to="/market/projects">浏览全部项目 ↗</RouterLink>
            </div>
            <div v-if="loading" class="popular-project-empty">正在加载合作项目…</div>
            <div v-else-if="popularGalleryItems.length" class="popular-project-gallery">
              <MarketAccordionGallery :items="popularGalleryItems" :default-index="2" />
            </div>
            <div v-else class="popular-project-empty">暂时没有开放中的合作项目。</div>
          </div>
        </section>

        <section class="discover-people-section">
          <div class="discover-people-inner">
            <div class="section-head compact"><div><p class="eyebrow">PEOPLE TO KNOW</p><h2>活跃 Builder</h2></div><RouterLink to="/market/builders">全部 Builder ↗</RouterLink></div>
            <div v-if="profiles.length" class="discover-people-grid" :class="{ 'is-sparse': profiles.length < 3 }">
              <RouterLink v-for="person in profiles.slice(0, 3)" :key="person.id" class="discover-person-row" :to="`/market/builders/${person.slug}`">
                <span class="person-avatar">{{ person.display_name.slice(0, 1) }}</span>
                <div class="person-details"><strong>{{ person.display_name }}</strong><span>{{ person.primary_role }} · {{ person.city || "远程" }}</span></div>
                <span class="availability" :class="person.availability">{{ person.availability === "unavailable" ? "暂不接单" : person.availability === "limited" ? "有限档期" : "开放合作" }}</span>
              </RouterLink>
            </div>
            <div v-else-if="!loading" class="empty-inline discover-people-empty">社区成员即将加入</div>
          </div>
        </section>
      </template>

      <template v-else-if="section === 'builders'">
        <div class="builder-discovery">
          <div class="project-search-bar">
            <label class="project-search-field">
              <span aria-hidden="true">⌕</span>
              <input id="builder-search" v-model="builderSearch" type="search" placeholder="搜索姓名、角色、技能或城市…" aria-label="搜索 Builder 姓名、角色、技能或城市" />
              <kbd>/</kbd>
            </label>
            <RouterLink to="/market/register" class="button-primary project-publish-button">＋ 完善我的档案</RouterLink>
          </div>

          <div class="project-discovery-layout">
            <aside class="project-filter-sidebar" aria-label="Builder 筛选">
              <div class="project-filter-heading">
                <strong>筛选 Builder</strong>
                <button v-if="hasBuilderFilters" type="button" @click="resetBuilderFilters">清除</button>
              </div>

              <fieldset class="project-filter-group">
                <legend>专业角色</legend>
                <label class="project-filter-option">
                  <input v-model="builderRoleFilter" type="radio" name="builder-role" value="all" />
                  <span>全部角色</span>
                  <small>{{ builderDirectoryProfiles.length }}</small>
                </label>
                <label v-for="role in builderRoleOptions" :key="role.role" class="project-filter-option">
                  <input v-model="builderRoleFilter" type="radio" name="builder-role" :value="role.role" />
                  <span>{{ role.role }}</span>
                  <small>{{ role.count }}</small>
                </label>
              </fieldset>

              <fieldset class="project-filter-group">
                <legend>合作偏好</legend>
                <label class="project-filter-option checkbox-option">
                  <input v-model="builderAvailableOnly" type="checkbox" />
                  <span>当前可合作</span>
                </label>
                <label class="project-filter-option checkbox-option">
                  <input v-model="builderRemoteOnly" type="checkbox" />
                  <span>支持远程</span>
                </label>
              </fieldset>
            </aside>

            <section class="project-results" aria-label="Builder 搜索结果">
              <div class="project-results-toolbar">
                <p><strong>{{ visibleProfiles.length }}</strong> 位公开 Builder<span v-if="hasBuilderFilters"> / 共 {{ builderDirectoryProfiles.length }} 位</span></p>
                <label class="project-sort-control">
                  <span>排序</span>
                  <select v-model="builderSort" aria-label="Builder 排序">
                    <option value="default">默认顺序</option>
                    <option value="name">姓名 A-Z</option>
                  </select>
                </label>
              </div>

              <div class="project-list">
                <article v-for="person in visibleProfiles" :key="person.id" class="project-card">
                  <RouterLink class="project-main project-main-link builder-main-link" :to="`/market/builders/${person.slug}`">
                    <div class="project-result-meta">
                      <span class="project-owner-mark builder-avatar-mark">{{ person.display_name.slice(0, 1) }}</span>
                      <strong>{{ person.display_name }}</strong>
                      <span class="project-meta-divider">·</span>
                      <span>{{ person.primary_role }}</span>
                      <span v-if="person.is_demo" class="builder-demo-label">本地演示</span>
                      <span class="project-open-state builder-availability-state" :class="person.availability">
                        <i></i>{{ person.availability === "unavailable" ? "暂不接单" : person.availability === "limited" ? "有限档期" : "开放合作" }}
                      </span>
                    </div>
                    <h3>{{ person.headline || `${person.primary_role}，正在寻找合作` }}</h3>
                    <p>{{ person.bio || person.headline || "正在寻找值得投入的合作。" }}</p>
                    <div class="project-facts">
                      <span v-if="person.city && person.city !== '远程'">{{ person.city }}</span>
                      <span v-else-if="!person.city">远程协作</span>
                      <span v-for="mode in (person.collaboration_modes || []).slice(0, 2)" :key="mode">{{ mode }}</span>
                    </div>
                    <div class="project-result-tags">
                      <span v-for="skill in (person.skills || []).slice(0, 4)" :key="skill">{{ skill }}</span>
                    </div>
                  </RouterLink>
                  <button
                    class="button-secondary apply-button"
                    :disabled="person.is_demo"
                    :title="person.is_demo ? '本地演示档案不能提交合作申请' : undefined"
                    @click="openApplication('profile', person.id, person.display_name)"
                  >{{ person.is_demo ? "演示档案" : "申请合作" }}</button>
                </article>
                <div v-if="!visibleProfiles.length && builderDirectoryProfiles.length && hasBuilderFilters" class="empty-state project-empty-state">
                  <strong>没有找到符合条件的 Builder</strong>
                  <p>试试调整关键词或筛选条件。</p>
                  <button class="button-secondary" type="button" @click="resetBuilderFilters">清除筛选</button>
                </div>
                <div v-else-if="!builderDirectoryProfiles.length && !loading" class="empty-state project-empty-state">
                  <strong>目前还没有公开 Builder</strong>
                  <p>邀请你的第一位协作者加入。</p>
                </div>
              </div>
            </section>
          </div>
        </div>
      </template>

      <template v-else-if="section === 'projects'">
        <div class="project-discovery">
          <div class="project-search-bar">
            <label class="project-search-field">
              <span aria-hidden="true">⌕</span>
              <input id="project-search" v-model="projectSearch" type="search" placeholder="搜索项目、技能或角色…" aria-label="搜索项目、技能或角色" />
              <kbd>/</kbd>
            </label>
            <RouterLink to="/market/publish" class="button-primary project-publish-button">＋ 发布需求</RouterLink>
          </div>

          <div class="project-discovery-layout">
            <aside class="project-filter-sidebar" aria-label="项目筛选">
              <div class="project-filter-heading">
                <strong>筛选项目</strong>
                <button v-if="hasProjectFilters" type="button" @click="resetProjectFilters">清除</button>
              </div>

              <fieldset class="project-filter-group">
                <legend>合作类型</legend>
                <label class="project-filter-option">
                  <input v-model="projectTypeFilter" type="radio" name="project-type" value="all" />
                  <span>全部类型</span>
                  <small>{{ projectTypeCount("all") }}</small>
                </label>
                <label v-for="type in ['collaboration', 'freelance', 'advisor', 'cofounder']" :key="type" class="project-filter-option">
                  <input v-model="projectTypeFilter" type="radio" name="project-type" :value="type" />
                  <span>{{ projectTypeLabel(type) }}</span>
                  <small>{{ projectTypeCount(type) }}</small>
                </label>
              </fieldset>

              <fieldset class="project-filter-group">
                <legend>参与方式</legend>
                <label class="project-filter-option checkbox-option">
                  <input v-model="projectRemoteOnly" type="checkbox" />
                  <span>支持远程</span>
                </label>
              </fieldset>

              <fieldset v-if="projectRoleOptions.length" class="project-filter-group">
                <legend>需要的角色</legend>
                <label class="project-filter-option">
                  <input v-model="projectRoleFilter" type="radio" name="project-role" value="" />
                  <span>全部角色</span>
                </label>
                <label v-for="role in projectRoleOptions" :key="role" class="project-filter-option">
                  <input v-model="projectRoleFilter" type="radio" name="project-role" :value="role" />
                  <span>{{ role }}</span>
                </label>
              </fieldset>
            </aside>

            <section class="project-results" aria-label="项目搜索结果">
              <div class="project-results-toolbar">
                <p><strong>{{ visibleProjects.length }}</strong> 个开放项目<span v-if="hasProjectFilters"> / 共 {{ projects.length }} 个</span></p>
                <label class="project-sort-control">
                  <span>排序</span>
                  <select v-model="projectSort" aria-label="项目排序">
                    <option value="newest">最新发布</option>
                    <option value="title">项目名称</option>
                  </select>
                </label>
              </div>

              <div class="project-list">
                <article v-for="item in visibleProjects" :key="item.id" class="project-card">
                  <RouterLink class="project-main project-main-link" :to="`/market/projects/${item.slug}`">
                    <div class="project-result-meta">
                      <span class="project-owner-mark">{{ (item.owner_name || "V").slice(0, 1) }}</span>
                      <strong>{{ item.owner_name || "项目发起方" }}</strong>
                      <span class="project-meta-divider">·</span>
                      <span>{{ projectTypeLabel(item.opportunities[0]?.opportunity_type || "") }}</span>
                      <span v-if="item.is_demo" class="project-demo-badge">演示数据</span>
                      <span v-if="item.opportunities.some((opportunity) => opportunity.remote_allowed)" class="remote-label">支持远程</span>
                      <span class="project-open-state"><i></i> 招募中</span>
                    </div>
                    <h3>{{ item.title }}</h3>
                    <p>{{ item.summary }}</p>
                    <div class="project-facts">
                      <span>{{ item.opportunities.length }} 个开放机会</span>
                      <span>{{ item.opportunities[0]?.engagement_mode || "合作方式待沟通" }}</span>
                      <span>{{ item.opportunities[0]?.duration_hint || "周期面议" }}</span>
                      <span>{{ item.opportunities[0]?.budget_hint || "预算未公开" }}</span>
                    </div>
                    <div class="project-result-tags">
                      <span v-for="role in [...new Set(item.opportunities.flatMap((opportunity) => opportunity.looking_for_roles || []))].slice(0, 3)" :key="`role-${role}`" class="role-tag">{{ role }}</span>
                      <span v-for="skill in [...new Set(item.opportunities.flatMap((opportunity) => opportunity.required_skills || []))].slice(0, 4)" :key="`skill-${skill}`">{{ skill }}</span>
                    </div>
                  </RouterLink>
                  <button
                    class="button-secondary apply-button"
                    :disabled="item.is_demo || !item.opportunities.length"
                    :title="item.is_demo ? '演示项目不可提交真实申请' : undefined"
                    @click="applyToProject(item)"
                  >{{ item.is_demo ? "仅供预览" : item.opportunities.length > 1 ? "选择机会" : "提交申请" }}</button>
                </article>
                <div v-if="!visibleProjects.length && !loading" class="empty-state project-empty-state">
                  <strong>没有找到匹配的项目</strong>
                  <p>试试调整关键词或筛选条件。</p>
                  <button v-if="hasProjectFilters" class="button-secondary" type="button" @click="resetProjectFilters">清除筛选条件</button>
                  <span v-else>目前没有开放中的合作机会，你可以发布需求，邀请合适的人加入。</span>
                </div>
              </div>
            </section>
          </div>
        </div>
      </template>

      <template v-else-if="section === 'publish'">
        <form class="form-panel publish-form" @submit.prevent="publishOpportunity">
          <label>需求标题<input v-model="opportunityForm.title" required maxlength="100" placeholder="例如：为 AI 教育产品寻找一位前端伙伴" /></label>
          <div class="form-two-col">
            <label>合作类型<select v-model="opportunityForm.opportunity_type"><option value="collaboration">项目合作</option><option value="freelance">短期外包</option><option value="advisor">顾问支持</option><option value="cofounder">联合创始</option></select></label>
            <label>合作方式<input v-model="opportunityForm.engagement_mode" placeholder="远程 / 兼职 / 按阶段合作" /></label>
          </div>
          <label>项目背景与目标<textarea v-model="opportunityForm.summary" required rows="5" maxlength="3000" placeholder="介绍现在要解决的问题、项目阶段和希望达成的结果。" /></label>
          <label>需要的角色<input v-model="opportunityForm.looking_for_roles" required placeholder="用逗号分隔，例如：前端开发, AI 产品经理" /></label>
          <label>技能要求<input v-model="opportunityForm.required_skills" placeholder="例如：Vue、LLM、数据可视化" /></label>
          <div class="form-two-col">
            <label>预算说明<input v-model="opportunityForm.budget_hint" placeholder="例如：按项目报价 / 面议 / 暂不公开" /></label>
            <label>预计周期<input v-model="opportunityForm.duration_hint" placeholder="例如：4–6 周" /></label>
          </div>
          <div class="form-two-col">
            <label>地点范围<input v-model="opportunityForm.city_scope" placeholder="例如：上海 / 不限" /></label>
            <label class="checkbox-label"><input v-model="opportunityForm.remote_allowed" type="checkbox" /> 支持远程协作</label>
          </div>
          <div class="form-two-col">
            <label>截止日期<input v-model="opportunityForm.expires_at" type="date" /></label>
            <label>发布状态<select v-model="opportunityForm.status"><option value="open">立即招募</option><option value="draft">先保存为草稿</option><template v-if="editingOpportunityId"><option value="paused">暂停招募</option><option value="filled">已找到合作者</option><option value="closed">关闭需求</option></template></select></label>
          </div>
          <section class="project-plan-editor">
            <div class="plan-editor-head">
              <div><p class="eyebrow">PROJECT CONTEXT</p><h2>项目计划与团队 <span>可选</span></h2></div>
              <p>按项目实际情况填写。里程碑不限定周期，计划章节也可以留空。</p>
            </div>
            <label>团队介绍<textarea v-model="opportunityForm.team_intro" rows="3" maxlength="2000" placeholder="介绍团队背景或项目当前已有的协作基础；没有就留空。" /></label>
            <div class="plan-editor-subhead">
              <strong>项目成员</strong>
              <button class="quiet-button" type="button" :disabled="profiles.length <= opportunityForm.team_members.length" @click="addOpportunityTeamMember">＋ 添加成员</button>
            </div>
            <div v-if="opportunityForm.team_members.length" class="plan-editor-list">
              <div v-for="(member, index) in opportunityForm.team_members" :key="`${member.profile_id}-${index}`" class="team-member-editor">
                <label>Builder
                  <select v-model.number="member.profile_id">
                    <option v-for="profile in profiles" :key="profile.id" :value="profile.id">{{ profile.display_name }} · {{ profile.primary_role }}</option>
                  </select>
                </label>
                <label>项目角色<input v-model="member.role" maxlength="100" placeholder="例如：产品负责人" /></label>
                <label>职责 / 贡献<input v-model="member.contribution" maxlength="1000" placeholder="可留空" /></label>
                <button class="remove-inline" type="button" aria-label="移除成员" @click="opportunityForm.team_members.splice(index, 1)">移除</button>
              </div>
            </div>
            <p v-else class="plan-editor-hint">只添加已公开的 Builder Profile。尚未确定成员时可以留空。</p>

            <div class="plan-editor-subhead">
              <strong>推进里程碑</strong>
              <button class="quiet-button" type="button" @click="addMilestone">＋ 添加里程碑</button>
            </div>
            <div v-if="opportunityForm.milestones.length" class="plan-editor-list">
              <article v-for="(milestone, index) in opportunityForm.milestones" :key="index" class="milestone-editor">
                <div class="plan-editor-subhead">
                  <strong>节点 {{ index + 1 }}</strong>
                  <button class="remove-inline" type="button" @click="opportunityForm.milestones.splice(index, 1)">移除</button>
                </div>
                <label>节点名称<input v-model="milestone.title" required maxlength="120" placeholder="例如：完成首轮用户验证" /></label>
                <label>目标说明<textarea v-model="milestone.description" rows="2" maxlength="2000" placeholder="描述这个节点要推进什么。" /></label>
                <div class="form-two-col">
                  <label>状态<select v-model="milestone.status"><option value="planned">计划中</option><option value="in_progress">进行中</option><option value="completed">已完成</option><option value="blocked">受阻</option></select></label>
                  <label>时间表达<select v-model="milestone.timing_kind"><option value="unspecified">不指定时间</option><option value="fixed_dates">明确日期</option><option value="duration">周期长度</option><option value="relative">相对节点</option><option value="flexible">弹性安排</option></select></label>
                </div>
                <fieldset class="milestone-owners">
                  <legend>负责人（可选）</legend>
                  <label v-for="member in opportunityForm.team_members" :key="member.profile_id" class="owner-option">
                    <input v-model="milestone.owner_member_ids" type="checkbox" :value="member.profile_id" />
                    <span>{{ profiles.find((profile) => profile.id === member.profile_id)?.display_name || member.display_name || "Builder" }} · {{ member.role }}</span>
                  </label>
                  <p v-if="!opportunityForm.team_members.length" class="plan-editor-hint">先添加项目成员后，才能指定负责人。</p>
                </fieldset>
                <div v-if="milestone.timing_kind === 'fixed_dates'" class="form-two-col">
                  <label>开始<input v-model="milestone.timing_start_at" type="date" /></label>
                  <label>结束<input v-model="milestone.timing_end_at" type="date" /></label>
                </div>
                <div v-else-if="milestone.timing_kind === 'duration'" class="form-two-col">
                  <label>长度<input v-model.number="milestone.timing_duration_value" type="number" min="1" /></label>
                  <label>单位<input v-model="milestone.timing_duration_unit" placeholder="天 / 次迭代 / 月" /></label>
                </div>
                <label v-if="['relative', 'flexible', 'fixed_dates', 'duration'].includes(milestone.timing_kind)">时间说明<input v-model="milestone.timing_label" maxlength="120" placeholder="例如：完成原型后启动" /></label>
                <div class="form-two-col">
                  <label>目标（每行一项）<textarea v-model="milestone.goals_text" rows="3" placeholder="可留空；每行一项" /></label>
                  <label>交付物（每行一项）<textarea v-model="milestone.deliverables_text" rows="3" placeholder="可留空；每行一项" /></label>
                </div>
                <label>完成标准（每行一项）<textarea v-model="milestone.acceptance_criteria_text" rows="2" placeholder="可留空；每行一项" /></label>
              </article>
            </div>
            <p v-else class="plan-editor-hint">可按阶段、实验轮次、发布节点等方式定义，不需要填写也不会影响发布。</p>

            <div class="plan-editor-subhead">
              <strong>补充计划章节</strong>
              <button class="quiet-button" type="button" @click="addPlanSection">＋ 添加章节</button>
            </div>
            <div v-if="opportunityForm.plan_sections.length" class="plan-editor-list">
              <article v-for="(section, index) in opportunityForm.plan_sections" :key="index" class="section-editor">
                <div class="plan-editor-subhead">
                  <strong>章节 {{ index + 1 }}</strong>
                  <button class="remove-inline" type="button" @click="opportunityForm.plan_sections.splice(index, 1)">移除</button>
                </div>
                <div class="form-two-col">
                  <label>章节类型<select v-model="section.type"><option value="context">背景</option><option value="scope">范围边界</option><option value="approach">执行方式</option><option value="resources">已有资源</option><option value="risks">风险</option><option value="communication">沟通节奏</option><option value="custom">其他</option></select></label>
                  <label>标题<input v-model="section.title" required maxlength="100" placeholder="例如：实验设计" /></label>
                </div>
                <label>内容<textarea v-model="section.body" rows="3" maxlength="5000" placeholder="这个章节按需填写，不适用就不要添加。" /></label>
                <label>要点（每行一项）<textarea v-model="section.items_text" rows="3" placeholder="可选；每行一项" /></label>
              </article>
            </div>
            <p v-else class="plan-editor-hint">只有需要额外说明时再添加章节；没有默认必填的 communication 或风险模块。</p>
          </section>
          <div class="form-footer"><span>{{ editingOpportunityId ? "保存后会更新现有需求。" : "发布后可在工作台查看申请并随时更新招募状态。" }}</span><div><button v-if="editingOpportunityId" class="quiet-button" type="button" @click="editingOpportunityId = null; void router.push('/market/workspace')">取消</button><button class="button-primary" type="submit">{{ editingOpportunityId ? "保存需求修改" : "发布到找项目" }}</button></div></div>
        </form>
      </template>

      <template v-else-if="section === 'register'">
        <form class="form-panel auth-panel" @submit.prevent="register">
          <div class="invite-row"><label>邀请码<input v-model="registerForm.invite_code" required autocomplete="off" placeholder="输入邀请代码" /></label><button class="button-secondary" type="button" @click="verifyInvite">验证邀请码</button></div>
          <span v-if="inviteValid" class="verified-mark">邀请码有效</span>
          <label>邮箱<input v-model="registerForm.email" type="email" required autocomplete="email" placeholder="you@example.com" /></label>
          <label>密码<input v-model="registerForm.password" type="password" required minlength="10" autocomplete="new-password" placeholder="至少 10 位" /></label>
          <div class="form-two-col">
            <label>姓名 / 常用称呼<input v-model="registerForm.display_name" required placeholder="你希望别人如何称呼你" /></label>
            <label>核心职业身份<input v-model="registerForm.primary_role" required placeholder="例如：全栈开发者" /></label>
          </div>
          <label>一句话介绍<input v-model="registerForm.headline" placeholder="你正在做什么，或希望和谁合作" /></label>
          <p class="privacy-note">注册即代表你同意社区规范。邮箱与联系方式不会出现在公开主页；只有在你主动申请并获接受后，才按授权交换。</p>
          <button class="button-primary full-button" type="submit">创建市场账户</button>
          <p class="form-switch">已有账户？ <RouterLink to="/market/login">登录</RouterLink></p>
        </form>
      </template>

      <template v-else-if="section === 'login'">
        <form class="form-panel auth-panel" @submit.prevent="login">
          <label>邮箱<input v-model="loginForm.email" type="email" required autocomplete="email" placeholder="you@example.com" /></label>
          <label>密码<input v-model="loginForm.password" type="password" required autocomplete="current-password" placeholder="输入密码" /></label>
          <button class="button-primary full-button" type="submit">登录</button>
          <p class="form-switch">还没有账户？ <RouterLink to="/market/register">使用邀请码加入</RouterLink></p>
        </form>
      </template>

      <template v-else-if="section === 'workspace'">
        <div v-if="!profile" class="empty-state workspace-gate">
          <p>登录后管理你的资料、申请与联系方式授权。</p>
          <RouterLink class="button-primary" to="/market/login">登录市场账户</RouterLink>
        </div>
        <template v-else>
          <section class="workspace-profile-hero">
            <div class="workspace-avatar" aria-hidden="true">{{ profile.display_name.slice(0, 2) }}</div>
            <div class="workspace-profile-copy">
              <p class="workspace-kicker">YOUR BUILDER SPACE <span>个人空间</span></p>
              <h2>{{ profile.display_name }}</h2>
              <p class="workspace-profile-headline">{{ profile.primary_role }}<template v-if="profile.headline"> · {{ profile.headline }}</template></p>
              <div class="workspace-profile-meta">
                <span>⌖ {{ profile.city || "远程协作" }}</span>
                <span class="workspace-availability" :class="profile.availability"><i></i>{{ profile.availability === "unavailable" ? "暂不接单" : profile.availability === "limited" ? "有限档期" : "开放合作" }}</span>
                <span v-if="profile.skills?.length">{{ profile.skills.slice(0, 3).join(" · ") }}</span>
              </div>
            </div>
            <div class="workspace-profile-actions">
              <button class="workspace-bell" type="button" aria-label="查看合作申请通知" @click="workspaceTab = 'messages'">
                <span class="workspace-bell-icon" aria-hidden="true"></span><b v-if="pendingIncomingRequests.length">{{ pendingIncomingRequests.length }}</b>
              </button>
              <button class="workspace-edit-profile" type="button" @click="workspaceTab = 'profile'">编辑主页 <span aria-hidden="true">↗</span></button>
              <button class="workspace-logout" type="button" @click="logout">退出登录</button>
            </div>
          </section>
          <nav class="workspace-tabs" aria-label="个人空间导航">
            <button type="button" :class="{ active: workspaceTab === 'overview' }" @click="workspaceTab = 'overview'">总览</button>
            <button type="button" :class="{ active: workspaceTab === 'profile' }" @click="workspaceTab = 'profile'">个人资料</button>
            <button type="button" :class="{ active: workspaceTab === 'works' }" @click="workspaceTab = 'works'">我的作品 <span>{{ myWorks.length }}</span></button>
            <button type="button" :class="{ active: workspaceTab === 'opportunities' }" @click="workspaceTab = 'opportunities'">合作机会 <span>{{ myOpportunities.length }}</span></button>
            <button type="button" :class="{ active: workspaceTab === 'messages' }" @click="workspaceTab = 'messages'">消息通知 <span v-if="pendingIncomingRequests.length" class="workspace-tab-alert">{{ pendingIncomingRequests.length }}</span></button>
            <button type="button" :class="{ active: workspaceTab === 'settings' }" @click="workspaceTab = 'settings'">隐私与 Agent</button>
          </nav>

          <section v-if="workspaceTab === 'overview'" class="workspace-overview">
            <div class="workspace-stats">
              <button type="button" @click="workspaceTab = 'works'"><span>公开作品</span><strong>{{ myWorks.filter((work) => work.status === "published").length }}</strong><small>查看与管理作品 <i>↗</i></small></button>
              <button type="button" @click="workspaceTab = 'opportunities'"><span>招募中的机会</span><strong>{{ myOpportunities.filter((item) => item.status === "open").length }}</strong><small>管理合作需求 <i>↗</i></small></button>
              <button type="button" @click="workspaceTab = 'messages'"><span>待处理申请</span><strong>{{ pendingIncomingRequests.length }}</strong><small>查看消息通知 <i>↗</i></small></button>
              <button type="button" @click="workspaceTab = 'settings'"><span>Agent 凭证</span><strong>{{ agentTokens.filter((token) => token.is_active).length }}</strong><small>管理 API 权限 <i>↗</i></small></button>
            </div>

            <div class="workspace-overview-grid">
              <div class="workspace-overview-main">
                <section class="workspace-card workspace-notifications">
                  <div class="workspace-card-heading">
                    <div><p class="eyebrow">INBOX</p><h2>消息通知 <span v-if="pendingIncomingRequests.length">{{ pendingIncomingRequests.length }}</span></h2></div>
                    <button type="button" @click="workspaceTab = 'messages'">查看全部 <i>↗</i></button>
                  </div>
                  <article v-for="request in pendingIncomingRequests.slice(0, 4)" :key="request.id" class="workspace-notification-row">
                    <div class="workspace-notification-avatar">{{ request.applicant_name.slice(0, 1) }}</div>
                    <div class="workspace-notification-copy">
                      <strong>{{ request.applicant_name }} <span>向你发起了合作申请</span><i v-if="request.is_demo" class="workspace-demo-badge">演示</i></strong>
                      <p>{{ request.message || "对方暂未填写申请说明。" }}</p>
                      <small>{{ request.applicant_role || "Builder" }} · 等待你处理</small>
                    </div>
                    <div class="workspace-notification-actions">
                      <button class="notification-accept" type="button" @click="processRequest(request.id, 'accept')">接受</button>
                      <button type="button" @click="workspaceTab = 'messages'">查看</button>
                    </div>
                  </article>
                  <div v-if="!pendingIncomingRequests.length" class="workspace-inbox-empty">
                    <span aria-hidden="true">✳</span>
                    <div><strong>暂时没有待处理的合作申请</strong><p>新的合作申请会优先出现在这里；所有往来申请都可在消息通知中查看。</p></div>
                  </div>
                </section>

                <section class="workspace-card workspace-recent-work">
                  <div class="workspace-card-heading">
                    <div><p class="eyebrow">YOUR WORK</p><h2>最近的作品</h2></div>
                    <button type="button" @click="workspaceTab = 'works'">管理作品 <i>↗</i></button>
                  </div>
                  <article v-for="work in myWorks.slice(0, 3)" :key="work.id" class="workspace-mini-row">
                    <span class="workspace-mini-icon">W</span>
                    <div><strong>{{ work.title }}</strong><small>{{ work.status === "published" ? "公开发布" : work.status === "draft" ? "草稿" : "已下架" }}<template v-if="work.tags?.length"> · {{ work.tags.slice(0, 2).join(" / ") }}</template></small></div>
                    <span class="workspace-row-arrow">↗</span>
                  </article>
                  <div v-if="!myWorks.length" class="workspace-small-empty"><span>作品还在路上？</span><button type="button" @click="workspaceTab = 'works'">添加你的第一个作品 <i>↗</i></button></div>
                </section>
              </div>

              <aside class="workspace-overview-side">
                <section class="workspace-card workspace-completion">
                  <div class="workspace-card-heading"><div><p class="eyebrow">PROFILE</p><h2>主页完整度</h2></div><strong>{{ profileCompletion }}%</strong></div>
                  <div class="workspace-completion-track"><span :style="{ width: `${profileCompletion}%` }"></span></div>
                  <p>补充技能、介绍和档期，让合作者更容易判断是否适合一起做事。</p>
                  <button type="button" @click="workspaceTab = 'profile'">完善个人资料 <i>↗</i></button>
                </section>
                <section class="workspace-quick-actions">
                  <p class="eyebrow">QUICK ACTIONS</p><h2>接下来做什么？</h2>
                  <button type="button" @click="workspaceTab = 'works'"><span>＋</span><div><strong>发布一个作品</strong><small>让你的能力有迹可循</small></div><i>↗</i></button>
                  <RouterLink to="/market/publish"><span>＋</span><div><strong>发布合作需求</strong><small>邀请合适的人加入</small></div><i>↗</i></RouterLink>
                  <button type="button" @click="workspaceTab = 'settings'"><span>⌘</span><div><strong>连接你的 Agent</strong><small>配置 Skill 与 API 权限</small></div><i>↗</i></button>
                </section>
              </aside>
            </div>
          </section>

          <section v-if="workspaceTab === 'profile'" class="workspace-section">
            <div class="section-head"><div><p class="eyebrow">BUILDER PROFILE</p><h2>个人资料与档期</h2></div></div>
            <form class="form-panel workspace-form" @submit.prevent="saveProfile">
              <div class="form-two-col">
                <label>姓名 / 常用称呼<input v-model="profileForm.display_name" required /></label>
                <label>核心职业身份<input v-model="profileForm.primary_role" required /></label>
              </div>
              <label>一句话介绍<input v-model="profileForm.headline" /></label>
              <label>个人介绍<textarea v-model="profileForm.bio" rows="3" maxlength="2000" /></label>
              <div class="form-two-col">
                <label>城市<input v-model="profileForm.city" placeholder="可留空，表示远程/不限" /></label>
                <label>当前档期<select v-model="profileForm.availability"><option value="open">开放合作</option><option value="limited">有限档期</option><option value="unavailable">暂不接单</option></select></label>
              </div>
              <label>最早可合作时间<input v-model="profileForm.available_from" placeholder="例如：2026-11" /></label>
              <label>技能标签<input v-model="profileForm.skills" placeholder="用逗号分隔，例如：Vue, Go, LLM" /></label>
              <label>合作方式<input v-model="profileForm.collaboration_modes" placeholder="用逗号分隔，例如：远程, 项目制" /></label>
              <button class="button-secondary" type="submit">保存个人资料</button>
            </form>
          </section>
          <section v-if="workspaceTab === 'works'" class="workspace-section">
            <div class="section-head"><div><p class="eyebrow">YOUR WORK</p><h2>发布作品</h2></div></div>
            <form class="form-panel workspace-form" @submit.prevent="saveWork">
              <label>作品标题<input v-model="workForm.title" required maxlength="100" placeholder="你做过什么，或正在做什么" /></label>
              <label>作品摘要<input v-model="workForm.summary" required maxlength="300" placeholder="用一两句话讲清楚作品价值" /></label>
              <label>作品说明<textarea v-model="workForm.body" rows="4" placeholder="背景、你的贡献、使用的技术或结果" /></label>
              <div class="form-two-col">
                <label>Demo 链接<input v-model="workForm.demo_url" type="url" placeholder="https://" /></label>
                <label>代码 / 介绍链接<input v-model="workForm.repo_url" type="url" placeholder="https://" /></label>
              </div>
              <div class="form-two-col">
                <label>封面图片 URL<input v-model="workForm.cover_url" type="url" placeholder="可后续从图库选择" /></label>
                <label>标签<input v-model="workForm.tags" placeholder="用逗号分隔" /></label>
              </div>
              <label>发布状态<select v-model="workForm.status"><option value="published">公开发布</option><option value="draft">保存草稿</option></select></label>
              <div class="form-footer">
                <button v-if="editingWorkId" class="quiet-button" type="button" @click="editingWorkId = null; Object.assign(workForm, { title: '', summary: '', body: '', cover_url: '', demo_url: '', repo_url: '', tags: '', status: 'published' })">取消编辑</button>
                <button class="button-primary" type="submit">{{ editingWorkId ? "保存作品修改" : "发布作品" }}</button>
              </div>
            </form>
            <div v-if="myWorks.length" class="my-work-list">
              <article v-for="work in myWorks" :key="work.id" class="my-work-row">
                <div class="work-row-title"><strong>{{ work.title }}</strong><span>{{ work.status === "published" ? "已发布" : work.status === "draft" ? "草稿" : "已下架" }}</span></div>
                <div class="work-row-actions">
                  <button class="quiet-button" type="button" @click="editWork(work)">编辑</button>
                  <button v-if="work.status !== 'archived'" class="quiet-button" type="button" @click="updateWorkStatus(work, 'archived')">下架</button>
                  <button v-else class="quiet-button" type="button" @click="updateWorkStatus(work, 'published')">重新发布</button>
                </div>
              </article>
            </div>
            <div v-else class="empty-inline">发布后的作品会出现在这里，可随时编辑、下架或重新发布。</div>
          </section>
          <section v-if="workspaceTab === 'opportunities'" class="workspace-section">
            <div class="section-head"><div><p class="eyebrow">YOUR OPPORTUNITIES</p><h2>我的合作需求</h2></div><RouterLink to="/market/publish">＋ 发布新需求</RouterLink></div>
            <div v-if="myOpportunities.length" class="my-work-list">
              <article v-for="opportunity in myOpportunities" :key="opportunity.id" class="my-opportunity-row">
                <div class="work-row-title"><strong>{{ opportunity.title }}</strong><span>{{ opportunity.status }}</span></div>
                <p>{{ opportunity.summary }}</p>
                <div class="opportunity-manage-actions">
                  <button class="quiet-button" type="button" @click="editOpportunity(opportunity)">编辑内容</button>
                  <select :value="opportunity.status" aria-label="合作机会状态" @change="updateOpportunityStatus(opportunity, ($event.target as HTMLSelectElement).value)">
                    <option value="draft">草稿</option><option value="open">招募中</option><option value="paused">暂停</option><option value="filled">已找到合作者</option><option value="closed">已关闭</option>
                  </select>
                </div>
              </article>
            </div>
            <div v-else class="empty-inline">你还没有发布合作需求。</div>
          </section>
          <section v-if="workspaceTab === 'settings'" class="workspace-section">
            <div class="section-head"><div><p class="eyebrow">CONTACT PRIVACY</p><h2>联系方式与授权</h2></div></div>
            <p class="privacy-note">分享设置限定可分享字段；发起申请时仍需逐次选择，未勾选的联系方式不会因对方接受而展示。</p>
            <form class="contact-form" @submit.prevent="saveContacts">
              <label>邮箱<input v-model="contacts.email" type="email" /></label>
              <label>微信<input v-model="contacts.wechat" /></label>
              <label>手机号<input v-model="contacts.phone" /></label>
              <div class="share-options"><label><input v-model="contacts.share_email" type="checkbox" /> 同意在对方接受申请后交换邮箱</label><label><input v-model="contacts.share_wechat" type="checkbox" /> 同意交换微信</label><label><input v-model="contacts.share_phone" type="checkbox" /> 同意交换手机号</label></div>
              <button class="button-secondary" type="submit">保存联系方式设置</button>
            </form>
          </section>
          <section v-if="workspaceTab === 'settings'" class="workspace-section">
            <div class="section-head"><div><p class="eyebrow">AGENT ACCESS</p><h2>Agent / API 接入</h2></div></div>
            <p class="privacy-note agent-intro">为你使用的 Agent 单独创建凭证，并只授予完成任务所需的权限。接受或婉拒申请仍需你本人在市场会话中确认。</p>
            <div class="agent-skill-pack">
              <div>
                <strong>下载用户侧 Skill 包</strong>
                <p>提供 Trae、Codex、Claude Code 三种宿主文件，用户下载后即可把自己的 Agent 接到 BuilderNetwork。</p>
              </div>
              <div class="agent-skill-actions">
                <a class="quiet-button" href="/downloads/voidlab-buildernetwork-agent/README.md" download>安装说明</a>
                <a class="quiet-button" href="/downloads/voidlab-buildernetwork-agent/trae/voidlab-buildernetwork-agent/SKILL.md" download>Trae Skill</a>
                <a class="quiet-button" href="/downloads/voidlab-buildernetwork-agent/codex/AGENTS.md" download>Codex</a>
                <a class="quiet-button" href="/downloads/voidlab-buildernetwork-agent/claude-code/CLAUDE.md" download>Claude Code</a>
              </div>
            </div>
            <form class="form-panel agent-token-form" @submit.prevent="createAgentToken">
              <label>凭证名称<input v-model="agentTokenName" required maxlength="80" placeholder="例如：个人资料助手" /></label>
              <div v-for="group in agentScopeGroups" :key="group.title" class="agent-scope-group">
                <h3>{{ group.title }}</h3>
                <label v-for="scope in group.scopes" :key="scope.value" class="agent-scope-option">
                  <input v-model="agentTokenScopes" type="checkbox" :value="scope.value" />
                  <span><strong>{{ scope.label }}</strong><small>{{ scope.description }}</small></span>
                </label>
              </div>
              <div class="agent-form-footer">
                <span>权限可以随时停用。私密联系方式权限与 Profile 权限分开控制。</span>
                <button class="button-primary" type="submit" :disabled="agentTokenBusy">{{ agentTokenBusy ? "正在创建…" : "创建 Agent Token" }}</button>
              </div>
            </form>
            <div v-if="issuedAgentToken" class="agent-token-reveal">
              <div><strong>新 Token 仅显示这一次</strong><p>关闭此区域后无法再次查看；如遗失，请停用并创建新的凭证。</p></div>
              <code>{{ issuedAgentToken }}</code>
              <div class="agent-token-actions">
                <button class="button-secondary" type="button" @click="copyIssuedAgentToken">复制 Token</button>
                <button class="quiet-button" type="button" @click="issuedAgentToken = ''">我已保存，关闭</button>
              </div>
            </div>
            <div v-if="agentTokens.length" class="agent-token-list">
              <article v-for="token in agentTokens" :key="token.id" class="agent-token-row">
                <div class="agent-token-main">
                  <div class="agent-token-title"><strong>{{ token.name }}</strong><span :class="{ inactive: !token.is_active }">{{ token.is_active ? "启用中" : "已停用" }}</span></div>
                  <small>最近使用：{{ token.last_used_at || "尚未使用" }}</small>
                  <div class="tag-row"><span v-for="scope in token.scopes" :key="scope">{{ scope }}</span></div>
                </div>
                <button class="quiet-button" type="button" @click="updateAgentTokenStatus(token)">{{ token.is_active ? "立即停用" : "重新启用" }}</button>
              </article>
            </div>
            <div v-else class="empty-inline">还没有 Agent 凭证。创建后可将 Token 配置到你自己的 Agent 或 Skill 中。</div>
          </section>
          <div v-if="workspaceTab === 'messages'" class="workspace-columns workspace-message-columns">
            <section class="workspace-section">
              <div class="section-head"><div><p class="eyebrow">INCOMING</p><h2>收到的申请 <small>{{ incoming.length }}</small></h2></div></div>
              <article v-for="request in incoming" :key="request.id" class="request-card">
                <div class="request-card-head"><strong>{{ request.applicant_name }} <i v-if="request.is_demo" class="workspace-demo-badge">演示数据</i></strong><span class="request-status">{{ requestStatusLabel(request) }}</span></div>
                <p class="request-role">{{ request.applicant_role }}</p>
                <p>{{ request.message }}</p>
                <div v-if="request.status === 'pending'" class="request-actions"><button class="button-primary small-button" @click="processRequest(request.id, 'accept')">同意并交换联系方式</button><button class="quiet-button" @click="processRequest(request.id, 'decline')">婉拒</button></div>
                <template v-else-if="request.status === 'accepted'">
                  <div v-if="request.is_demo" class="workspace-demo-note">演示联系方式仅供预览，example.test 和 _demo 均为虚构值。</div>
                  <div v-if="!request.is_demo && request.exchange_revoked" class="workspace-demo-note">联系方式交换已撤回。</div>
                  <div v-else-if="requestExchange(request)" class="exchange-box">
                    <section class="exchange-party">
                      <strong>对方 · {{ request.applicant_name }}</strong>
                      <span v-for="field in exchangeContactFields(requestExchange(request)?.applicant)" :key="field.label"><b>{{ field.label }}</b>{{ field.value }}</span>
                      <span v-if="!exchangeContactFields(requestExchange(request)?.applicant).length" class="exchange-empty">对方暂无可展示的联系方式</span>
                    </section>
                    <section class="exchange-party">
                      <strong>你 · {{ profile?.display_name }}</strong>
                      <span v-for="field in exchangeContactFields(requestExchange(request)?.recipient)" :key="field.label"><b>{{ field.label }}</b>{{ field.value }}</span>
                      <span v-if="!exchangeContactFields(requestExchange(request)?.recipient).length" class="exchange-empty">你暂无可展示的联系方式</span>
                    </section>
                  </div>
                  <div v-else-if="!request.is_demo && exchangeErrors[request.id]" class="exchange-empty-state">{{ exchangeErrors[request.id] }} <button class="text-action" type="button" @click="showExchange(request.id)">重试</button></div>
                  <div v-else-if="!request.is_demo && !request.exchange_revoked" class="exchange-empty-state">正在读取双方授权的联系方式…</div>
                  <button v-if="!request.is_demo && !request.exchange_revoked" class="text-action exchange-revoke" type="button" @click="revokeExchange(request.id)">撤回联系方式共享</button>
                </template>
              </article>
              <div v-if="!incoming.length" class="empty-inline">暂时没有收到申请</div>
            </section>
            <section class="workspace-section">
              <div class="section-head"><div><p class="eyebrow">OUTGOING</p><h2>我发出的申请 <small>{{ outgoing.length }}</small></h2></div></div>
              <article v-for="request in outgoing" :key="request.id" class="request-card">
                <div class="request-card-head"><strong>{{ request.target_label || (request.target_type === "opportunity" ? "合作机会申请" : "Builder 合作申请") }} <i v-if="request.is_demo" class="workspace-demo-badge">演示数据</i></strong><span class="request-status">{{ requestStatusLabel(request) }}</span></div>
                <p>{{ request.message }}</p>
                <button v-if="request.status === 'pending'" class="quiet-button" @click="processRequest(request.id, 'withdraw')">撤回申请</button>
                <template v-if="request.status === 'accepted'">
                  <div v-if="request.is_demo" class="workspace-demo-note">演示联系方式仅供预览，example.test 和 _demo 均为虚构值。</div>
                  <div v-if="!request.is_demo && request.exchange_revoked" class="workspace-demo-note">联系方式交换已撤回。</div>
                  <div v-else-if="requestExchange(request)" class="exchange-box">
                    <section class="exchange-party">
                      <strong>你 · {{ profile?.display_name }}</strong>
                      <span v-for="field in exchangeContactFields(requestExchange(request)?.applicant)" :key="field.label"><b>{{ field.label }}</b>{{ field.value }}</span>
                      <span v-if="!exchangeContactFields(requestExchange(request)?.applicant).length" class="exchange-empty">你暂无可展示的联系方式</span>
                    </section>
                    <section class="exchange-party">
                      <strong>对方 · {{ request.target_label || "合作方" }}</strong>
                      <span v-for="field in exchangeContactFields(requestExchange(request)?.recipient)" :key="field.label"><b>{{ field.label }}</b>{{ field.value }}</span>
                      <span v-if="!exchangeContactFields(requestExchange(request)?.recipient).length" class="exchange-empty">对方暂无可展示的联系方式</span>
                    </section>
                  </div>
                  <div v-else-if="!request.is_demo && exchangeErrors[request.id]" class="exchange-empty-state">{{ exchangeErrors[request.id] }} <button class="text-action" type="button" @click="showExchange(request.id)">重试</button></div>
                  <div v-else-if="!request.is_demo && !request.exchange_revoked" class="exchange-empty-state">正在读取双方授权的联系方式…</div>
                  <button v-if="!request.is_demo && !request.exchange_revoked" class="text-action exchange-revoke" type="button" @click="revokeExchange(request.id)">撤回联系方式共享</button>
                </template>
              </article>
              <div v-if="!outgoing.length" class="empty-inline">你还没有发出合作申请</div>
            </section>
          </div>
        </template>
      </template>

      <section v-else class="empty-state">页面不存在。<RouterLink to="/market">返回发现广场</RouterLink></section>

      <footer class="market-footer"><span>VOIDLAB NETWORK</span><span>Build openly. Connect intentionally.</span></footer>
    </div>

    <div v-if="selectedTarget" class="modal-backdrop" @click.self="selectedTarget = null">
      <form class="application-modal" @submit.prevent="submitApplication">
        <button class="modal-close" type="button" aria-label="关闭" @click="selectedTarget = null">×</button>
        <p class="eyebrow">COLLABORATION REQUEST</p>
        <h2>申请联系</h2>
        <p class="modal-context">对象：{{ selectedTarget.title }}</p>
        <label>申请理由<textarea v-model="applicationMessage" required maxlength="500" rows="5" placeholder="介绍你为什么想合作、能带来什么，以及希望如何推进。" /></label>
        <fieldset v-if="contacts.email || contacts.wechat || contacts.phone" class="application-consent">
          <legend>申请人联系方式授权（可选）</legend>
          <label v-if="contacts.email && contacts.share_email"><input v-model="applicationConsent.share_email" type="checkbox" /> 对方接受后展示我的邮箱</label>
          <label v-if="contacts.wechat && contacts.share_wechat"><input v-model="applicationConsent.share_wechat" type="checkbox" /> 对方接受后展示我的微信</label>
          <label v-if="contacts.phone && contacts.share_phone"><input v-model="applicationConsent.share_phone" type="checkbox" /> 对方接受后展示我的手机号</label>
          <p>只显示你在工作台开启分享且本次主动勾选的字段。也可以不分享联系方式。</p>
        </fieldset>
        <div v-if="myWorks.some((work) => work.status === 'published')" class="application-works">
          <strong>关联作品（可选，最多 3 个）</strong>
          <label v-for="work in myWorks.filter((item) => item.status === 'published')" :key="work.id"><input v-model="applicationWorkIDs" type="checkbox" :value="work.id" @change="limitRelatedWorks" /> {{ work.title }}</label>
        </div>
        <p class="privacy-note">提交后对方能看到你的公开资料和申请理由。对方接受后，按你的授权交换联系方式。</p>
        <button class="button-primary full-button" type="submit" :disabled="applicationBusy">{{ applicationBusy ? "正在发送…" : "提交申请" }}</button>
      </form>
    </div>
  </main>
</template>

<style scoped>
/* SWISS STYLE / MAGAZINE / INFRASTRUCTURE VIBE */
.market-page {
  min-height: 100vh;
  background: #F8FAFC;
  color: #0F172A;
  font-family: "Helvetica Neue", Helvetica, Arial, "PingFang SC", sans-serif;
  -webkit-font-smoothing: antialiased;
}
.market-shell { max-width: 1540px; margin: 0 auto; padding: 0 max(40px, calc((100vw - 1540px) / 2)) 60px; }
.is-workspace-shell { max-width: 1800px; padding: 0 40px 60px; }
.is-projects-shell, .is-builders-shell { max-width: 1800px; padding: 0 40px 60px; }
.market-nav { height: 80px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid #E2E8F0; gap: 24px; }
.market-brand { display: flex; align-items: center; gap: 12px; color: #0F172A; text-decoration: none; font-size: 14px; letter-spacing: 0.15em; font-weight: 800; text-transform: uppercase; }
.market-brand b { color: #0BCB78; font-weight: 800; }
.market-logo { width: 32px; height: 32px; flex-shrink: 0; object-fit: contain; border-radius: 7px; }
.market-links, .market-actions { display: flex; align-items: center; gap: 32px; }
.market-links a, .quiet-link, .account-link { color: #64748B; text-decoration: none; font-size: 13px; font-weight: 600; letter-spacing: 0.05em; text-transform: uppercase; transition: color 0.2s; }
.market-links a:hover, .market-links a.router-link-active, .quiet-link:hover, .account-link:hover { color: #0F172A; }
.publish-link, .button-primary { background: #0BCB78; color: #FFFFFF; text-decoration: none; border: 0; padding: 12px 24px; font-size: 13px; font-weight: 700; letter-spacing: 0.05em; text-transform: uppercase; cursor: pointer; transition: background 0.2s; }
.publish-link:hover, .button-primary:hover { background: #08B96C; }
.button-secondary { display: inline-flex; align-items: center; justify-content: center; border: 1px solid #CBD5E1; color: #0F172A; background: transparent; text-decoration: none; padding: 11px 23px; font-size: 13px; font-weight: 700; letter-spacing: 0.05em; text-transform: uppercase; cursor: pointer; transition: border-color 0.2s, color 0.2s; }
.button-secondary:hover { border-color: #0F172A; color: #0F172A; }
.account-link { border: 1px solid #CBD5E1; padding: 11px 16px; }

/* TYPOGRAPHY & HEADINGS */
.page-heading { padding: 80px 0 60px; }
.eyebrow { margin: 0 0 16px; color: #0BCB78; font-size: 11px; letter-spacing: 0.2em; font-weight: 700; text-transform: uppercase; }
.page-heading h1 { font-size: clamp(42px, 6vw, 72px); line-height: 1.05; letter-spacing: -0.04em; margin: 0; font-weight: 800; color: #0F172A; }
.page-desc { margin: 24px 0 0; color: #475569; font-size: 18px; line-height: 1.6; max-width: 680px; }
.is-workspace-shell .page-heading { padding: 38px 0 24px; }
.is-workspace-shell .page-heading .eyebrow { margin-bottom: 10px; }
.is-workspace-shell .page-heading h1 { font-size: clamp(38px, 4vw, 50px); }
.is-workspace-shell .page-desc { margin-top: 12px; font-size: 15px; }
.is-workspace-shell .feedback { margin-bottom: 12px; padding: 11px 16px; font-size: 12px; }
.is-projects-shell .page-heading { padding: 48px 0 32px; }
.is-projects-shell .page-heading .eyebrow { margin-bottom: 12px; }
.is-projects-shell .page-heading h1 { font-size: clamp(44px, 5vw, 56px); }
.is-projects-shell .page-desc { margin-top: 14px; font-size: 16px; }
.is-builders-shell .page-heading { padding: 48px 0 32px; }
.is-builders-shell .page-heading .eyebrow { margin-bottom: 12px; }
.is-builders-shell .page-heading h1 { font-size: clamp(44px, 5vw, 56px); }
.is-builders-shell .page-desc { margin-top: 14px; font-size: 16px; }

/* FEEDBACK */
.feedback { padding: 16px 20px; margin: 0 0 24px; font-size: 14px; font-weight: 500; }
.error-feedback { color: #991B1B; background: #FEE2E2; border-left: 4px solid #DC2626; }
.success-feedback { color: #166534; background: #DCFCE7; border-left: 4px solid #16A34A; }
.warning-feedback { color: #92400E; background: #FEF3C7; border-left: 4px solid #F59E0B; }

/* FULLSCREEN LIGHT HERO */
.hero-full-light {
  position: relative;
  width: 100vw;
  min-height: 100vh;
  background: #09090B;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  margin-left: calc(-50vw + 50%); /* Break out of any parent container if needed, though we moved it out */
}

.hero-full-bg-waves {
  position: absolute;
  inset: 0;
  z-index: 1;
  overflow: hidden;
  pointer-events: none;
}
.hero-full-bg-waves :deep(.shape-waves-root) {
  pointer-events: none;
}
.hero-full-light::after {
  content: "";
  position: absolute;
  z-index: 1;
  inset: 0;
  pointer-events: none;
  background: rgba(9, 9, 11, 0.26);
}
.hero-full-bg-code {
  position: absolute;
  position: absolute;
  left: -2%;
  top: 30%;
  z-index: 1;
  opacity: 0.15;
  transform: perspective(1000px) rotateY(15deg);
  user-select: none;
}
.hero-full-bg-code pre {
  font-family: "JetBrains Mono", ui-monospace, monospace;
  font-size: 14px;
  line-height: 1.6;
  color: #94A3B8; /* Slate gray code */
  text-shadow: none; /* No shadow needed on white background */
}

/* Integrated Header for Fullscreen */
.market-nav-full {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 10;
  width: 100%;
  padding: 0 32px;
  box-sizing: border-box;
  background: transparent;
}
.market-nav-solid {
  position: sticky;
  top: 0;
  z-index: 10;
  width: 100vw;
  margin-left: calc(50% - 50vw);
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
}
.nav-primary {
  display: flex;
  align-items: center;
  gap: clamp(28px, 4vw, 68px);
}
.hero-full-light .market-brand { color: #F8FAFC; }
.hero-full-light .market-brand b { color: #0BCB78; }
.hero-full-light .market-links a,
.hero-full-light .quiet-link,
.hero-full-light .account-link { color: #A1A1AA; }
.hero-full-light .market-links a:hover,
.hero-full-light .market-links a.router-link-active,
.hero-full-light .quiet-link:hover { color: #0BCB78; }
.hero-full-light .account-link {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 40px;
  padding: 0 18px;
  border: 0;
  border-radius: 5px;
  background: #0BCB78;
  color: #111827;
  letter-spacing: .04em;
}
.hero-full-light .account-link:hover { background: #08B96C; color: #111827; }
.hero-full-light .publish-link { background: #0BCB78; color: #111827; }
.hero-full-light .publish-link:hover { background: #08B96C; }

/* Minimal market brand lockup */
.hero-brand-lockup {
  position: relative;
  z-index: 2;
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 24px 72px;
  text-align: center;
  box-sizing: border-box;
}
.hero-brand-lockup h1 {
  max-width: 100%;
  margin: 0;
  font-size: clamp(52px, 9vw, 124px);
  line-height: .98;
  letter-spacing: -.075em;
  font-weight: 760;
  color: #F8FAFC;
  font-family: "Helvetica Neue", Helvetica, Arial, "PingFang SC", sans-serif;
}
.hero-brand-lockup p {
  margin: 18px 0 0;
  color: #0BCB78;
  font-size: 12px;
  font-weight: 650;
  letter-spacing: .42em;
  text-indent: .42em;
}

/* Spacing adjustments for discover page */
.is-discover-shell {
  padding: 0;
}
.discover-content-start {
  height: 0;
}


/* AUTH SPLIT LAYOUT (50:50) */
.auth-split-layout {
  min-height: 100svh;
  display: grid;
  grid-template-columns: 1fr 1fr;
  background: #FFFFFF;
}

/* LEFT SIDE - FAULTY TERMINAL */
.auth-visual-side {
  min-height: 100svh;
  box-sizing: border-box;
  background-color: #0A0A0A;
  background-image: radial-gradient(circle at 20% 30%, rgba(11, 203, 120, 0.15) 0%, transparent 50%),
                    radial-gradient(circle at 80% 80%, rgba(11, 203, 120, 0.1) 0%, transparent 50%);
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 48px;
  color: #F8FAFC;
}

.terminal-overlay {
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    0deg,
    rgba(0, 0, 0, 0.15),
    rgba(0, 0, 0, 0.15) 1px,
    transparent 1px,
    transparent 2px
  );
  pointer-events: none;
  z-index: 1;
}

.terminal-content {
  position: relative;
  z-index: 2;
  flex: 1;
  min-height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.auth-visual-side .brand-logo {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 20px;
  font-weight: 700;
  letter-spacing: -0.02em;
  color: #FFFFFF;
  text-decoration: none;
}

.terminal-text-container {
  margin: auto 0;
}
.auth-kicker { font: 10px ui-monospace, monospace; letter-spacing: .16em; color: #81E6B0; }
.terminal-text-container h1 { font-size: clamp(36px, 4vw, 68px); line-height: 1.25; letter-spacing: -.04em; margin: 24px 0; }
.auth-brand-desc { color: #CCF6E0; font-size: 17px; line-height: 1.9; }
.auth-visual-footer { display: flex; justify-content: space-between; color: #A1A1AA; font-size: 11px; letter-spacing: .08em; }
.auth-back { color: #64748b; text-decoration: none; font-size: 13px; display: inline-block; margin-bottom: 40px; }
.auth-form-container h2 { font-size: 30px; letter-spacing: -.04em; margin: 0 0 12px; }
.auth-intro { color: #64748b; font-size: 14px; line-height: 1.7; margin: 0 0 28px; }
.auth-form-note { font-size: 12px; line-height: 1.7; color: #64748b; margin: 0; }
.auth-form-note a { color: #0BCB78; text-decoration: none; }
.auth-invite-row { display: flex; align-items: flex-end; gap: 16px; }
.auth-invite-row label { flex: 1; min-width: 0; }
.auth-verify { flex-shrink: 0; padding: 10px 0; border: 0; background: transparent; color: #0BCB78; cursor: pointer; }

/* Glitch Effect */
.glitch {
  font-size: 64px;
  font-weight: 900;
  text-transform: uppercase;
  position: relative;
  display: inline-block;
  color: #FFFFFF;
  margin-bottom: 24px;
  letter-spacing: -0.05em;
}
.glitch::before, .glitch::after {
  content: attr(data-text);
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  opacity: 0.8;
}
.glitch::before {
  left: 2px;
  text-shadow: -2px 0 #0BCB78;
  clip: rect(44px, 450px, 56px, 0);
  animation: glitch-anim 5s infinite linear alternate-reverse;
}
.glitch::after {
  left: -2px;
  text-shadow: -2px 0 #0BCB78;
  clip: rect(44px, 450px, 56px, 0);
  animation: glitch-anim2 5s infinite linear alternate-reverse;
}
@keyframes glitch-anim {
  0% { clip: rect(10px, 9999px, 83px, 0); }
  20% { clip: rect(40px, 9999px, 12px, 0); }
  40% { clip: rect(87px, 9999px, 56px, 0); }
  60% { clip: rect(32px, 9999px, 98px, 0); }
  80% { clip: rect(74px, 9999px, 34px, 0); }
  100% { clip: rect(12px, 9999px, 76px, 0); }
}
@keyframes glitch-anim2 {
  0% { clip: rect(65px, 9999px, 100px, 0); }
  20% { clip: rect(21px, 9999px, 84px, 0); }
  40% { clip: rect(93px, 9999px, 45px, 0); }
  60% { clip: rect(12px, 9999px, 67px, 0); }
  80% { clip: rect(48px, 9999px, 23px, 0); }
  100% { clip: rect(89px, 9999px, 12px, 0); }
}

.terminal-lines {
  font-family: ui-monospace, monospace;
  color: #0BCB78;
  font-size: 14px;
  line-height: 2;
}
.terminal-lines p { margin: 0; }
.blinking-cursor { animation: blink 1s step-end infinite; }
@keyframes blink { 50% { opacity: 0; } }

.footer-links {
  display: flex;
  gap: 24px;
  font-size: 12px;
  font-weight: 500;
  color: #64748B;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* RIGHT SIDE - CLEAN FORM */
.auth-form-side {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #FFFFFF;
  min-width: 0;
  padding: 48px 24px;
  box-sizing: border-box;
}

.auth-form-container {
  width: 100%;
  max-width: 400px;
  padding: 0;
}

.auth-tabs {
  display: flex;
  gap: 32px;
  margin-bottom: 28px;
}

.auth-tabs a {
  background: none;
  border: none;
  padding: 0 0 8px;
  font-size: 15px;
  text-decoration: none;
  font-weight: 800;
  color: #94A3B8;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.auth-tabs a.active {
  color: #0F172A;
  border-bottom-color: #0BCB78;
}

.clean-form {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.clean-form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 0;
  font-size: 12px;
  color: #64748b;
}

.clean-form-group label {
  font-size: 12px;
  font-weight: 700;
  color: #64748B;
}

.clean-form-group input, .clean-form-group select {
  width: 100%;
  min-width: 0;
  box-sizing: border-box;
  border: none;
  border-bottom: 1px solid #E2E8F0;
  padding: 8px 0;
  font-size: 16px;
  color: #0F172A;
  background: transparent;
  outline: none;
  transition: border-color 0.2s;
  border-radius: 0;
}

.clean-form-group input:focus, .clean-form-group select:focus {
  border-bottom-color: #0BCB78;
}

.clean-form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.clean-submit {
  margin-top: 16px;
  width: 100%;
  padding: 14px;
  border-radius: 4px;
  font-size: 15px;
}

/* RESPONSIVE SPLIT */
@media (max-width: 980px) {
  .auth-split-layout { grid-template-columns: 1fr; }
  .auth-visual-side { min-height: 300px; padding: 24px; }
  .terminal-text-container { margin: 32px 0; }
  .auth-form-side { padding: 40px 24px; }
  .glitch { font-size: 32px; margin-bottom: 12px; }
  .terminal-lines { display: none; }
  .footer-links { display: none; }
}

/* SECTIONS & GRIDS */
.section-head { display: flex; align-items: flex-end; justify-content: space-between; gap: 24px; margin: 80px 0 32px; border-bottom: 2px solid #0F172A; padding-bottom: 16px; }
.section-head h2 { font-size: 32px; letter-spacing: -0.04em; margin: 0; font-weight: 800; }
.section-head a, .text-action { color: #0BCB78; text-decoration: none; font-size: 13px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; }
.section-head .eyebrow { margin-bottom: 12px; }

/* CARDS - SWISS CLEAN */
.work-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 32px; }
.work-card, .project-card { background: #FFFFFF; border: 1px solid #E2E8F0; display: flex; flex-direction: column; transition: transform 0.2s, box-shadow 0.2s; }
.work-card:hover, .project-card:hover { transform: translateY(-2px); box-shadow: 0 12px 24px rgba(15, 23, 42, 0.04); }
.work-cover { height: 220px; background: #F1F5F9; overflow: hidden; border-bottom: 1px solid #E2E8F0; }
.work-cover img { width: 100%; height: 100%; object-fit: cover; display: block; }
.cover-abstract { display: flex; align-items: flex-end; justify-content: space-between; padding: 24px; background: linear-gradient(135deg, #F8FAFC, #E2E8F0); height: 100%; box-sizing: border-box; }
.cover-abstract span { font-size: 11px; letter-spacing: 0.2em; color: #64748B; font-weight: 700; }
.cover-abstract i { font-size: 32px; font-style: normal; color: #94A3B8; font-weight: 300; }
.work-card-body { padding: 24px; display: flex; flex-direction: column; flex: 1; }
.card-meta { font-size: 10px; letter-spacing: 0.15em; color: #64748B; margin: 0 0 12px; text-transform: uppercase; font-weight: 700; }
.work-card h3, .project-card h3 { font-size: 22px; letter-spacing: -0.03em; margin: 0 0 12px; font-weight: 800; color: #0F172A; }
.builder-name-link,
.project-title-link {
  color: inherit;
  text-decoration: none;
}
.builder-name-link:hover,
.project-title-link:hover {
  text-decoration: underline;
  text-decoration-color: rgba(11, 203, 120, 0.55);
  text-underline-offset: 3px;
}
.work-card-body > p:not(.card-meta) { font-size: 14px; line-height: 1.6; color: #475569; flex: 1; }
.tag-row { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 24px; }
.tag-row span { font-size: 11px; color: #334155; border: 1px solid #CBD5E1; padding: 4px 10px; font-weight: 600; background: #F8FAFC; }

/* SPLIT SECTIONS (PEOPLE / OPEN CALLS) */
.split-sections { display: grid; grid-template-columns: 1fr 1fr; gap: 64px; margin-top: 80px; }
.section-head.compact { margin: 0 0 24px; border-bottom-width: 1px; padding-bottom: 12px; }
.section-head.compact h2 { font-size: 24px; }
.person-row, .opportunity-row { padding: 16px 0; border-bottom: 1px solid #E2E8F0; display: flex; align-items: center; gap: 16px; }
.person-avatar { width: 48px; height: 48px; display: grid; place-items: center; background: #F1F5F9; color: #0F172A; font: 24px "Times New Roman", Times, serif; flex: none; border: 1px solid #CBD5E1; }
.person-details { display: flex; flex-direction: column; gap: 4px; flex: 1; }
.person-details strong { font-size: 15px; font-weight: 700; color: #0F172A; }
.person-details span { font-size: 12px; color: #64748B; }
.availability { font-size: 10px; color: #0F172A; border: 1px solid #0F172A; padding: 4px 10px; white-space: nowrap; font-weight: 700; text-transform: uppercase; }
.availability.limited { color: #B45309; border-color: #B45309; }
.availability.unavailable { color: #94A3B8; border-color: #94A3B8; }
.opportunity-row { color: inherit; text-decoration: none; transition: background 0.2s; margin: 0 -16px; padding: 16px; }
.opportunity-row:hover { background: #F8FAFC; }
.opportunity-symbol { width: 48px; height: 48px; border: 1px solid #CBD5E1; display: grid; place-items: center; color: #0F172A; font-size: 20px; font-weight: 300; background: #FFF; }
.opportunity-text { display: flex; flex-direction: column; gap: 6px; flex: 1; }
.opportunity-text strong { font-size: 15px; font-weight: 700; }
.opportunity-text small { font-size: 12px; color: #64748B; }
.row-arrow { color: #0F172A; font-size: 18px; }

/* EMPTY STATES */
.empty-state, .empty-inline { color: #64748B; border: 1px dashed #CBD5E1; padding: 40px; text-align: center; font-size: 14px; font-weight: 500; background: #F8FAFC; }
.empty-inline { padding: 24px 0; border: 0; background: transparent; text-align: left; }

/* LIST VIEWS */
.list-toolbar { display: flex; align-items: center; justify-content: space-between; color: #475569; font-size: 14px; font-weight: 600; padding: 0 0 24px; border-bottom: 2px solid #0F172A; margin-bottom: 32px; }
.full-button { width: 100%; margin-top: auto; }

.project-discovery { width: 100%; }
.project-search-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}
.project-search-field {
  display: flex;
  min-width: 0;
  height: 44px;
  flex: 1;
  align-items: center;
  gap: 12px;
  padding: 0 12px;
  border: 1px solid #CBD5E1;
  border-radius: 6px;
  background: #FFFFFF;
  color: #64748B;
}
.project-search-field:focus-within {
  border-color: #0BCB78;
  box-shadow: 0 0 0 3px rgba(11, 203, 120, .1);
}
.project-search-field > span { font-size: 22px; line-height: 1; }
.project-search-field input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: #0F172A;
  font: inherit;
  font-size: 13px;
}
.project-search-field input::placeholder { color: #94A3B8; }
.project-search-field kbd {
  flex: none;
  padding: 2px 6px;
  border: 1px solid #E2E8F0;
  border-radius: 4px;
  color: #64748B;
  font-size: 11px;
}
.project-publish-button { flex: none; }
.project-discovery-layout {
  display: grid;
  grid-template-columns: 236px minmax(0, 1fr);
  align-items: start;
  gap: 36px;
}
.project-filter-sidebar {
  position: sticky;
  top: 24px;
  min-width: 0;
  padding-right: 24px;
  border-right: 1px solid #E2E8F0;
}
.project-filter-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 22px;
  color: #0F172A;
  font-size: 14px;
}
.project-filter-heading button {
  border: 0;
  background: transparent;
  color: #0A9F5E;
  font: inherit;
  font-size: 11px;
  font-weight: 700;
  cursor: pointer;
}
.project-filter-group {
  display: grid;
  gap: 4px;
  min-width: 0;
  margin: 0 0 22px;
  padding: 0;
  border: 0;
}
.project-filter-group legend {
  margin-bottom: 9px;
  color: #64748B;
  font-size: 11px;
  font-weight: 800;
}
.project-filter-option {
  display: flex;
  min-height: 34px;
  align-items: center;
  gap: 9px;
  padding: 0 8px;
  border-radius: 5px;
  color: #334155;
  font-size: 12px;
  cursor: pointer;
}
.project-filter-option:hover { background: #F1F5F9; }
.project-filter-option input {
  width: 14px;
  height: 14px;
  flex: none;
  margin: 0;
  accent-color: #0BCB78;
}
.project-filter-option span { min-width: 0; flex: 1; }
.project-filter-option small {
  color: #94A3B8;
  font-size: 10px;
}
.project-results { min-width: 0; }
.project-results-toolbar {
  display: flex;
  min-height: 42px;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid #CBD5E1;
  color: #64748B;
  font-size: 12px;
}
.project-results-toolbar p { margin: 0; }
.project-results-toolbar strong { color: #0F172A; }
.project-sort-control {
  display: flex;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
}
.project-sort-control select {
  min-height: 30px;
  padding: 0 26px 0 9px;
  border: 1px solid #CBD5E1;
  border-radius: 5px;
  background: #FFFFFF;
  color: #334155;
  font: inherit;
  font-size: 11px;
}
.project-list { display: flex; flex-direction: column; }
.project-card {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 22px;
  padding: 17px 8px;
  border: 0;
  border-bottom: 1px solid #E2E8F0;
  background: transparent;
  box-shadow: none;
}
.project-results .project-card:hover {
  transform: none;
  background: #FBFDFC;
  box-shadow: none;
}
.project-main { min-width: 0; max-width: none; }
.project-main-link {
  display: block;
  color: inherit;
  text-decoration: none;
}
.project-main-link:hover h3 {
  text-decoration: underline;
  text-decoration-color: rgba(11, 203, 120, 0.55);
  text-underline-offset: 3px;
}
.project-result-meta { display: flex; align-items: center; flex-wrap: wrap; gap: 7px; color: #64748B; font-size: 11px; }
.project-owner-mark {
  display: grid;
  width: 20px;
  height: 20px;
  place-items: center;
  border: 1px solid #DCE5E0;
  border-radius: 50%;
  background: #F2FBF6;
  color: #087A49;
  font-size: 9px;
  font-weight: 800;
}
.project-result-meta strong { color: #334155; font-size: 11px; }
.project-meta-divider { color: #CBD5E1; }
.project-result-meta .remote-label {
  padding: 3px 7px;
  border-radius: 999px;
  background: #F8FAFC;
  font-size: 9px;
}
.project-demo-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 6px;
  border: 1px solid #CDEEDC;
  border-radius: 3px;
  background: #F0FAF4;
  color: #16814F;
  font-size: 9px;
  font-weight: 700;
}
.project-open-state { display: inline-flex; align-items: center; gap: 5px; margin-left: auto; color: #087A49; font-size: 10px; }
.project-open-state i { width: 6px; height: 6px; border-radius: 50%; background: #0BCB78; }
.project-card h3 { margin: 8px 0 5px; font-size: 18px; line-height: 1.4; }
.project-card p {
  display: -webkit-box;
  overflow: hidden;
  margin: 0;
  color: #475569;
  font-size: 13px;
  line-height: 1.65;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}
.project-facts {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 18px;
  margin-top: 9px;
  color: #64748B;
  font-size: 10px;
  font-weight: 600;
}
.project-result-tags { display: flex; flex-wrap: wrap; gap: 6px; margin-top: 9px; }
.project-result-tags span {
  padding: 4px 8px;
  border: 1px solid #DCE5E0;
  border-radius: 999px;
  background: #F2FBF6;
  color: #276749;
  font-size: 9px;
  font-weight: 700;
}
.project-result-tags span.role-tag {
  border-color: #E2E8F0;
  background: #F8FAFC;
  color: #475569;
}
.apply-button {
  align-self: center;
  min-height: 34px;
  padding: 0 12px;
  border: 1px solid #D0D7DE;
  border-radius: 5px;
  background: #F6F8FA;
  color: #334155;
  font-size: 11px;
  white-space: nowrap;
}
.apply-button:hover { border-color: #0BCB78; color: #087A49; }
.project-empty-state { margin-top: 20px; padding: 52px 20px; }
.project-empty-state strong { display: block; margin-bottom: 8px; color: #0F172A; }
.project-empty-state p { margin: 0 0 16px; }
.builder-avatar-mark { width: 30px; height: 30px; font-size: 12px; }
.builder-demo-label {
  padding: 3px 7px;
  border: 1px solid #DCE5E0;
  border-radius: 999px;
  background: #F7FBF8;
  color: #64748B;
  font-size: 9px;
}
.builder-main-link h3 { margin: 8px 0 5px; font-size: 16px; line-height: 1.45; }
.builder-availability-state.limited { color: #A16207; }
.builder-availability-state.limited i { background: #EAB308; }
.builder-availability-state.unavailable { color: #94A3B8; }
.builder-availability-state.unavailable i { background: #94A3B8; }
.apply-button:disabled { opacity: 0.7; cursor: not-allowed; }

/* FORMS - INFRASTRUCTURE CLEAN */
.form-panel { max-width: 800px; background: #FFFFFF; border: 1px solid #E2E8F0; padding: 48px; box-shadow: 0 10px 30px rgba(15,23,42,0.03); }
.publish-form { margin: 0 auto 80px; }
.auth-panel { max-width: 640px; margin: 0 auto 80px; }
.form-panel label { display: flex; flex-direction: column; gap: 10px; color: #0F172A; font-size: 13px; font-weight: 700; margin-bottom: 24px; text-transform: uppercase; letter-spacing: 0.05em; }
.form-panel input, .form-panel select, .form-panel textarea, .contact-form input { width: 100%; border: 1px solid #CBD5E1; background: #F8FAFC; color: #0F172A; padding: 14px 16px; font: inherit; outline: none; font-size: 15px; text-transform: none; font-weight: 400; transition: border-color 0.2s, background 0.2s; box-sizing: border-box; }
.form-panel input:focus, .form-panel select:focus, .form-panel textarea:focus, .contact-form input:focus { border-color: #0BCB78; background: #FFFFFF; box-shadow: 0 0 0 3px rgba(11, 203, 120, 0.1); }
.form-panel textarea { resize: vertical; line-height: 1.6; }
.form-two-col { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; }
.project-plan-editor {
  margin: 16px 0 32px;
  padding: clamp(20px, 3vw, 36px);
  border: 1px solid #DCE5E0;
  background: #F8FBF9;
}
.plan-editor-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 28px;
  padding-bottom: 20px;
  border-bottom: 1px solid #DCE5E0;
}
.plan-editor-head .eyebrow { margin-bottom: 8px; }
.plan-editor-head h2 { margin: 0; color: #0F172A; font-size: clamp(20px, 2.2vw, 28px); letter-spacing: -.03em; }
.plan-editor-head h2 span {
  margin-left: 8px;
  color: #64748B;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: .06em;
  text-transform: uppercase;
}
.plan-editor-head > p {
  max-width: 420px;
  margin: 0;
  color: #64748B;
  font-size: 13px;
  line-height: 1.7;
}
.plan-editor-subhead {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin: 28px 0 14px;
}
.plan-editor-subhead strong { color: #0F172A; font-size: 13px; font-weight: 800; }
.plan-editor-list { display: grid; gap: 14px; }
.team-member-editor {
  display: grid;
  grid-template-columns: minmax(150px, 1.1fr) minmax(140px, .8fr) minmax(180px, 1.2fr) auto;
  align-items: end;
  gap: 14px;
  padding: 18px;
  border: 1px solid #E2E8F0;
  background: #FFFFFF;
}
.team-member-editor label,
.milestone-editor label,
.section-editor label { margin-bottom: 14px; }
.team-member-editor .remove-inline { margin-bottom: 25px; }
.milestone-editor,
.section-editor {
  min-width: 0;
  padding: clamp(16px, 2.4vw, 26px);
  border: 1px solid #DCE5E0;
  background: #FFFFFF;
}
.milestone-editor > .plan-editor-subhead,
.section-editor > .plan-editor-subhead { margin-top: 0; }
.milestone-owners {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 18px;
  min-width: 0;
  margin: 0 0 20px;
  padding: 14px 16px;
  border: 1px solid #E2E8F0;
}
.milestone-owners legend {
  padding: 0 6px;
  color: #475569;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: .06em;
}
.form-panel .milestone-owners .owner-option {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  color: #334155;
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0;
  text-transform: none;
}
.form-panel .milestone-owners .owner-option input {
  width: 16px;
  height: 16px;
  padding: 0;
  accent-color: #0BCB78;
}
.plan-editor-hint {
  margin: 0;
  padding: 14px 16px;
  color: #64748B;
  background: #FFFFFF;
  border-left: 2px solid #0BCB78;
  font-size: 12px;
  line-height: 1.7;
}
.remove-inline {
  border: 0;
  padding: 8px 4px;
  color: #64748B;
  background: transparent;
  font: inherit;
  font-size: 11px;
  font-weight: 700;
  cursor: pointer;
}
.remove-inline:hover { color: #B91C1C; }
.checkbox-label { flex-direction: row !important; align-items: center; justify-content: flex-start; text-transform: none !important; font-weight: 500 !important; font-size: 14px !important; letter-spacing: 0 !important; }
.checkbox-label input, .share-options input { width: 18px; height: 18px; accent-color: #0BCB78; }
.form-footer { display: flex; align-items: center; justify-content: space-between; gap: 24px; border-top: 1px solid #E2E8F0; padding-top: 32px; margin-top: 16px; }
.form-footer span, .privacy-note { font-size: 13px; color: #64748B; line-height: 1.6; font-weight: 400; }
.invite-row { display: flex; align-items: flex-end; gap: 16px; }
.invite-row label { flex: 1; }
.invite-row button { margin-bottom: 24px; padding: 14px 24px; }
.verified-mark { display: inline-block; color: #16A34A; font-size: 13px; font-weight: 600; margin: -12px 0 20px; }
.privacy-note { margin: 16px 0; padding: 16px; background: #F1F5F9; border-left: 3px solid #CBD5E1; }
.form-switch { text-align: center; font-size: 14px; color: #64748B; margin: 32px 0 0; font-weight: 500; }
.form-switch a { color: #0BCB78; font-weight: 700; text-decoration: none; }
.form-switch a:hover { text-decoration: underline; }

/* BUILDER PERSONAL SPACE */
.workspace-profile-hero {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 24px;
  padding: 28px 0 32px;
  border-bottom: 1px solid #E2E8F0;
}
.workspace-avatar {
  display: grid;
  width: 88px;
  height: 88px;
  place-items: center;
  border: 4px solid #FFFFFF;
  border-radius: 50%;
  outline: 1px solid #CDEEDC;
  background: linear-gradient(145deg, #DDF8E9, #B8EBD0);
  color: #087B4C;
  font-size: 27px;
  font-weight: 800;
  letter-spacing: -.06em;
}
.workspace-profile-copy { min-width: 0; }
.workspace-kicker {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 0 0 8px;
  color: #0A9B5D;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: .15em;
}
.workspace-kicker span {
  padding: 3px 7px;
  border-radius: 3px;
  background: #E9F8EF;
  color: #16814F;
  font-size: 10px;
  letter-spacing: 0;
}
.workspace-profile-copy h2 {
  margin: 0;
  color: #10241B;
  font-size: clamp(26px, 3vw, 36px);
  font-weight: 800;
  letter-spacing: -.04em;
}
.workspace-profile-headline {
  overflow: hidden;
  margin: 6px 0 12px;
  color: #64748B;
  font-size: 14px;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.workspace-profile-meta { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 16px; color: #64748B; font-size: 12px; }
.workspace-availability { display: inline-flex; align-items: center; gap: 6px; color: #16814F; }
.workspace-availability i { width: 7px; height: 7px; border-radius: 50%; background: #0BCB78; }
.workspace-availability.limited { color: #A16207; }
.workspace-availability.limited i { background: #EAB308; }
.workspace-availability.unavailable { color: #64748B; }
.workspace-availability.unavailable i { background: #94A3B8; }
.workspace-profile-actions { display: flex; align-items: center; gap: 12px; }
.workspace-bell {
  position: relative;
  display: grid;
  width: 42px;
  height: 42px;
  place-items: center;
  border: 1px solid #DCE8E0;
  border-radius: 50%;
  background: #FFFFFF;
  cursor: pointer;
}
.workspace-bell:hover { border-color: #0BCB78; background: #F3FBF6; }
.workspace-bell-icon { position: relative; width: 13px; height: 14px; border: 1.7px solid #334155; border-radius: 8px 8px 4px 4px; }
.workspace-bell-icon::before { position: absolute; top: -4px; left: 4px; width: 3px; height: 3px; border-radius: 50%; background: #334155; content: ""; }
.workspace-bell-icon::after { position: absolute; bottom: -4px; left: 3px; width: 5px; height: 2px; border-radius: 0 0 4px 4px; background: #334155; content: ""; }
.workspace-bell b {
  position: absolute;
  top: -4px;
  right: -4px;
  display: grid;
  min-width: 17px;
  height: 17px;
  place-items: center;
  border: 2px solid #FFFFFF;
  border-radius: 12px;
  background: #0BCB78;
  color: #FFFFFF;
  font-size: 9px;
}
.workspace-edit-profile {
  min-height: 40px;
  padding: 0 14px;
  border: 1px solid #DCE8E0;
  border-radius: 4px;
  background: #FFFFFF;
  color: #1F3529;
  font: inherit;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}
.workspace-edit-profile:hover { border-color: #0BCB78; color: #07804B; }
.workspace-edit-profile span { margin-left: 8px; color: #0A9B5D; }
.workspace-logout { border: 0; background: transparent; color: #94A3B8; font: inherit; font-size: 11px; cursor: pointer; }
.workspace-logout:hover { color: #B91C1C; }
.workspace-tabs {
  display: flex;
  gap: 26px;
  overflow-x: auto;
  margin: 0 0 28px;
  border-bottom: 1px solid #E2E8F0;
  scrollbar-width: thin;
}
.workspace-tabs button {
  position: relative;
  display: inline-flex;
  flex: none;
  align-items: center;
  gap: 8px;
  min-height: 48px;
  padding: 0 2px;
  border: 0;
  background: transparent;
  color: #64748B;
  font: inherit;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}
.workspace-tabs button:hover, .workspace-tabs button.active { color: #10241B; }
.workspace-tabs button.active::after { position: absolute; right: 0; bottom: -1px; left: 0; height: 2px; background: #0BCB78; content: ""; }
.workspace-tabs button > span { color: #94A3B8; font-size: 11px; }
.workspace-tabs button > .workspace-tab-alert { display: grid; min-width: 18px; height: 18px; place-items: center; border-radius: 50%; background: #0BCB78; color: #FFFFFF; font-size: 10px; }
.workspace-overview { padding-bottom: 12px; }
.workspace-stats { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 14px; }
.workspace-stats button {
  display: flex;
  min-width: 0;
  flex-direction: column;
  align-items: flex-start;
  padding: 18px;
  border: 1px solid #E4EBE6;
  border-radius: 5px;
  background: #FFFFFF;
  text-align: left;
  cursor: pointer;
  transition: border-color .18s, transform .18s, box-shadow .18s;
}
.workspace-stats button:hover { transform: translateY(-2px); border-color: #A7E4C3; box-shadow: 0 8px 22px rgba(15, 50, 31, .06); }
.workspace-stats button > span { color: #64748B; font-size: 11px; font-weight: 600; }
.workspace-stats button > strong { margin: 8px 0 7px; color: #10241B; font-size: 30px; font-weight: 800; line-height: 1; }
.workspace-stats button > small { color: #7B8B80; font-size: 10px; }
.workspace-stats button > small i { color: #0A9B5D; font-style: normal; }
.workspace-overview-grid { display: grid; grid-template-columns: minmax(0, 1.65fr) minmax(260px, .85fr); align-items: start; gap: 20px; margin-top: 20px; }
.workspace-overview-main, .workspace-overview-side { display: grid; min-width: 0; gap: 20px; }
.workspace-card { min-width: 0; padding: 22px; border: 1px solid #E4EBE6; border-radius: 5px; background: #FFFFFF; }
.workspace-card-heading { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
.workspace-card-heading .eyebrow { margin-bottom: 7px; font-size: 9px; }
.workspace-card-heading h2 { display: flex; align-items: center; gap: 8px; margin: 0; color: #10241B; font-size: 18px; font-weight: 750; letter-spacing: -.02em; }
.workspace-card-heading h2 > span { display: grid; min-width: 18px; height: 18px; place-items: center; border-radius: 50%; background: #E9F8EF; color: #16814F; font-size: 10px; }
.workspace-card-heading > button { flex: none; border: 0; background: transparent; color: #0A9B5D; font: inherit; font-size: 11px; font-weight: 700; cursor: pointer; }
.workspace-card-heading > button i { margin-left: 3px; font-style: normal; }
.workspace-notification-row { display: grid; grid-template-columns: 38px minmax(0, 1fr) auto; align-items: start; gap: 12px; padding: 18px 0; border-bottom: 1px solid #EEF2EF; }
.workspace-notification-row:last-of-type { border-bottom: 0; }
.workspace-notification-avatar { display: grid; width: 36px; height: 36px; place-items: center; border-radius: 50%; background: #EDF7F0; color: #16814F; font-size: 13px; font-weight: 800; }
.workspace-notification-copy { min-width: 0; }
.workspace-notification-copy > strong { display: block; overflow: hidden; color: #23372A; font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.workspace-notification-copy > strong span { color: #64748B; font-weight: 500; }
.workspace-demo-badge { display: inline-flex; align-items: center; margin-left: 7px; padding: 2px 6px; border: 1px solid #CDEEDC; border-radius: 3px; background: #F0FAF4; color: #16814F; font-size: 9px; font-style: normal; font-weight: 700; vertical-align: 1px; }
.workspace-demo-note { margin-top: 14px; padding: 9px 12px; border-left: 2px solid #0BCB78; background: #F2FAF5; color: #547260; font-size: 11px; }
.workspace-notification-copy p { display: -webkit-box; overflow: hidden; margin: 5px 0; color: #64748B; font-size: 11px; line-height: 1.5; -webkit-box-orient: vertical; -webkit-line-clamp: 2; }
.workspace-notification-copy small { color: #94A3B8; font-size: 10px; }
.workspace-notification-actions { display: flex; gap: 8px; }
.workspace-notification-actions button { min-height: 28px; padding: 0 9px; border: 1px solid #E2E8E4; border-radius: 3px; background: #FFFFFF; color: #64748B; font: inherit; font-size: 10px; cursor: pointer; }
.workspace-notification-actions button.notification-accept { border-color: #0BCB78; background: #0BCB78; color: #FFFFFF; }
.workspace-notification-actions button.notification-accept:hover { background: #08B96C; }
.workspace-inbox-empty { display: flex; align-items: center; gap: 14px; padding: 27px 0 10px; }
.workspace-inbox-empty > span { display: grid; width: 38px; height: 38px; place-items: center; border-radius: 50%; background: #F0F8F3; color: #0A9B5D; font-size: 18px; }
.workspace-inbox-empty strong { color: #334155; font-size: 12px; }
.workspace-inbox-empty p { margin: 4px 0 0; color: #94A3B8; font-size: 11px; }
.workspace-mini-row { display: grid; grid-template-columns: 34px minmax(0, 1fr) auto; align-items: center; gap: 12px; padding: 14px 0; border-bottom: 1px solid #EEF2EF; }
.workspace-mini-row:last-of-type { border-bottom: 0; }
.workspace-mini-icon { display: grid; width: 32px; height: 32px; place-items: center; border-radius: 4px; background: #F0F8F3; color: #16814F; font-size: 11px; font-weight: 800; }
.workspace-mini-row > div { display: grid; gap: 4px; min-width: 0; }
.workspace-mini-row strong { overflow: hidden; color: #334155; font-size: 12px; text-overflow: ellipsis; white-space: nowrap; }
.workspace-mini-row small { overflow: hidden; color: #94A3B8; font-size: 10px; text-overflow: ellipsis; white-space: nowrap; }
.workspace-row-arrow { color: #94A3B8; font-size: 13px; }
.workspace-small-empty { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 18px 0 0; color: #94A3B8; font-size: 11px; }
.workspace-small-empty button { border: 0; background: transparent; color: #0A9B5D; font: inherit; font-size: 11px; font-weight: 700; cursor: pointer; }
.workspace-small-empty i { font-style: normal; }
.workspace-completion .workspace-card-heading > strong { color: #16814F; font-size: 19px; }
.workspace-completion-track { height: 5px; overflow: hidden; margin: 18px 0 12px; border-radius: 5px; background: #EAF0EC; }
.workspace-completion-track span { display: block; height: 100%; border-radius: inherit; background: #0BCB78; transition: width .25s ease; }
.workspace-completion > p { margin: 0 0 15px; color: #718078; font-size: 11px; line-height: 1.65; }
.workspace-completion > button { padding: 0; border: 0; background: transparent; color: #0A9B5D; font: inherit; font-size: 11px; font-weight: 700; cursor: pointer; }
.workspace-completion > button i { margin-left: 4px; font-style: normal; }
.workspace-quick-actions { padding: 22px; border-radius: 5px; background: #F2F8F4; }
.workspace-quick-actions .eyebrow { margin-bottom: 8px; font-size: 9px; }
.workspace-quick-actions h2 { margin: 0 0 11px; color: #10241B; font-size: 18px; letter-spacing: -.02em; }
.workspace-quick-actions > button, .workspace-quick-actions > a { display: grid; grid-template-columns: 28px minmax(0, 1fr) auto; align-items: center; gap: 10px; width: 100%; padding: 12px 0; border: 0; border-top: 1px solid #DFEAE2; background: transparent; color: inherit; text-align: left; text-decoration: none; cursor: pointer; }
.workspace-quick-actions > button > span, .workspace-quick-actions > a > span { display: grid; width: 26px; height: 26px; place-items: center; border-radius: 4px; background: #FFFFFF; color: #0A9B5D; font-size: 14px; }
.workspace-quick-actions > button > div, .workspace-quick-actions > a > div { display: grid; gap: 4px; }
.workspace-quick-actions strong { color: #334155; font-size: 11px; }
.workspace-quick-actions small { color: #829087; font-size: 10px; }
.workspace-quick-actions > button > i, .workspace-quick-actions > a > i { color: #94A3B8; font-size: 12px; font-style: normal; }
.workspace-quick-actions > button:hover strong, .workspace-quick-actions > a:hover strong { color: #07804B; }
.workspace-message-columns { margin-top: 8px; }
.workspace-message-columns .workspace-section { margin-top: 0; }
.workspace-section > .section-head { margin: 0 0 24px; }

/* PROFILE PANEL (WORKSPACE HEADER) */
.profile-panel { display: flex; align-items: center; justify-content: space-between; gap: 32px; background: #0F172A; color: #FFFFFF; padding: 48px; border: 1px solid #1E293B; }
.profile-panel h2 { font-size: 36px; margin: 0 0 12px; font-weight: 800; letter-spacing: -0.03em; }
.profile-panel p:last-child { margin: 0; color: #94A3B8; font-size: 15px; }
.profile-actions { display: flex; align-items: center; gap: 16px; }
.quiet-button { border: 0; background: transparent; color: #64748B; font-size: 13px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em; cursor: pointer; transition: color 0.2s; }
.quiet-button:hover { color: #0F172A; }
.profile-panel .quiet-button { color: #94A3B8; }
.profile-panel .quiet-button:hover { color: #FFFFFF; }

/* WORKSPACE LISTS & REQUESTS */
.workspace-section { margin-top: 32px; }
.agent-intro { max-width: 800px; margin: 0 0 24px; color: #475569; font-size: 15px; line-height: 1.6; }
.agent-token-form { max-width: none; display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 24px; }
.agent-token-form > label, .agent-form-footer { grid-column: 1 / -1; }
.agent-scope-group { border-top: 1px solid #E2E8F0; padding-top: 24px; }
.agent-scope-group h3 { font-size: 14px; color: #0F172A; margin: 0 0 16px; font-weight: 800; text-transform: uppercase; }
.agent-scope-option { display: flex !important; flex-direction: row !important; align-items: flex-start; gap: 12px !important; margin: 0 0 12px !important; padding: 12px; background: #F8FAFC; border: 1px solid #E2E8F0; color: #0F172A !important; text-transform: none !important; }
.agent-scope-option input { width: 16px !important; height: 16px !important; margin-top: 2px; accent-color: #0BCB78; }
.agent-scope-option span { display: flex; flex-direction: column; gap: 4px; }
.agent-scope-option strong { font-size: 14px; font-weight: 700; }
.agent-scope-option small { font-size: 12px; color: #64748B; line-height: 1.5; font-weight: 400; text-transform: none; }
.agent-form-footer { display: flex; align-items: center; justify-content: space-between; gap: 24px; border-top: 1px solid #E2E8F0; padding-top: 24px; }
.agent-form-footer span { font-size: 12px; color: #64748B; line-height: 1.6; font-weight: 500; }
.agent-token-form button:disabled { opacity: 0.6; cursor: wait; }

.agent-token-reveal { display: grid; grid-template-columns: 1fr auto; gap: 16px 32px; align-items: center; margin-top: 24px; padding: 24px; background: #F0FDF4; border: 1px solid #BBF7D0; }
.agent-token-reveal strong { font-size: 15px; color: #166534; font-weight: 700; }
.agent-token-reveal p { font-size: 13px; color: #166534; margin: 8px 0 0; }
.agent-token-reveal code { grid-column: 1 / -1; display: block; overflow-wrap: anywhere; padding: 16px; background: #FFFFFF; color: #0F172A; border: 1px solid #BBF7D0; font-family: ui-monospace, monospace; font-size: 14px; }
.agent-token-actions { grid-column: 1 / -1; display: flex; align-items: center; gap: 16px; }

.agent-token-list { margin-top: 24px; border-top: 1px solid #E2E8F0; }
.agent-token-row { display: flex; align-items: center; justify-content: space-between; gap: 24px; padding: 24px 0; border-bottom: 1px solid #E2E8F0; }
.agent-token-main { min-width: 0; }
.agent-token-title { display: flex; align-items: center; gap: 12px; margin-bottom: 8px; }
.agent-token-title strong { font-size: 16px; font-weight: 700; color: #0F172A; }
.agent-token-title span { font-size: 10px; color: #0BCB78; border: 1px solid #0BCB78; padding: 4px 8px; font-weight: 700; text-transform: uppercase; }
.agent-token-title span.inactive { color: #64748B; border-color: #CBD5E1; }
.agent-token-main > small { font-size: 12px; color: #64748B; font-weight: 500; }
.agent-token-main .tag-row { margin-top: 12px; }
.agent-skill-pack { display: flex; justify-content: space-between; gap: 24px; align-items: flex-start; padding: 32px; margin: 24px 0 32px; border: 1px solid #E2E8F0; background: #F8FAFC; }
.agent-skill-pack p { margin: 8px 0 0; color: #475569; line-height: 1.7; font-size: 14px; }
.agent-skill-actions { display: flex; flex-wrap: wrap; gap: 16px; justify-content: flex-end; }

.workspace-form { max-width: none; display: flex; flex-direction: column; gap: 0; }
.workspace-form .form-two-col { width: 100%; }
.workspace-form button { align-self: flex-start; }

.my-work-list { margin-top: 24px; border-top: 1px solid #E2E8F0; }
.my-work-row { display: flex; align-items: center; justify-content: space-between; gap: 24px; padding: 20px 0; border-bottom: 1px solid #E2E8F0; font-size: 14px; }
.my-work-row span { font-size: 11px; color: #64748B; font-weight: 700; text-transform: uppercase; }
.work-row-title, .work-row-actions, .opportunity-manage-actions { display: flex; align-items: center; gap: 16px; }
.work-row-title strong { font-size: 15px; font-weight: 700; color: #0F172A; }

.my-opportunity-row { padding: 24px 0; border-bottom: 1px solid #E2E8F0; }
.my-opportunity-row > p { font-size: 14px; color: #475569; line-height: 1.6; margin: 12px 0; }
.opportunity-manage-actions select { background: #F8FAFC; border: 1px solid #CBD5E1; color: #0F172A; padding: 10px 14px; font-size: 12px; font-weight: 600; outline: none; }
.opportunity-manage-actions select:focus { border-color: #0BCB78; }

.contact-form { display: grid; grid-template-columns: repeat(3, 1fr); gap: 24px; background: #FFFFFF; border: 1px solid #E2E8F0; padding: 32px; }
.contact-form > label { display: flex; flex-direction: column; gap: 8px; font-size: 12px; color: #0F172A; font-weight: 700; text-transform: uppercase; margin: 0; }
.contact-form input { padding: 12px 16px; }
.share-options { grid-column: 1 / -1; display: flex; flex-wrap: wrap; gap: 32px; margin-top: 8px; }
.share-options label { font-size: 13px; color: #475569; font-weight: 600; text-transform: none; flex-direction: row; align-items: center; gap: 8px; }
.contact-form button { grid-column: 1 / -1; justify-self: start; margin-top: 16px; }

.workspace-columns { display: grid; grid-template-columns: 1fr 1fr; gap: 48px; }
.section-head h2 small { font-size: 14px; color: #64748B; font-weight: 600; text-transform: uppercase; margin-left: 12px; }

/* REQUEST CARDS & CONTACT EXCHANGE */
.request-card { background: #FFFFFF; border: 1px solid #E2E8F0; padding: 24px; margin-bottom: 16px; transition: box-shadow 0.2s; }
.request-card:hover { box-shadow: 0 8px 16px rgba(15, 23, 42, 0.03); }
.request-card-head { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; margin-bottom: 12px; }
.request-card-head strong { font-size: 16px; font-weight: 800; color: #0F172A; }
.request-status { font-size: 11px; color: #0BCB78; font-weight: 700; text-transform: uppercase; border: 1px solid #AFEBCD; background: #E5F8EE; padding: 4px 10px; }
.request-role { font-size: 12px !important; color: #64748B !important; margin: 0 0 12px !important; font-weight: 600; }
.request-card > p:not(.request-role) { font-size: 14px; color: #475569; line-height: 1.6; margin-bottom: 16px; }
.request-actions { display: flex; align-items: center; gap: 16px; margin-top: 16px; }

.small-button { padding: 10px 16px; font-size: 12px; }

.exchange-box { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); margin-top: 16px; background: #F8FAFC; color: #0F172A; border: 1px solid #CBD5E1; font-size: 13px; font-weight: 500; }
.exchange-party { display: flex; min-width: 0; flex-direction: column; gap: 10px; padding: 16px; }
.exchange-party + .exchange-party { border-left: 1px solid #E2E8F0; }
.exchange-party > strong { color: #23372A; font-size: 12px; }
.exchange-party > span { display: flex; flex-wrap: wrap; gap: 8px; color: #334155; overflow-wrap: anywhere; }
.exchange-box b { font-weight: 700; color: #475569; margin-right: 4px; text-transform: uppercase; font-size: 11px; letter-spacing: 0.05em; }
.exchange-empty, .exchange-empty-state { color: #64748B !important; font-size: 12px; }
.exchange-empty-state { margin-top: 16px; padding: 12px 16px; background: #F8FAFC; border: 1px solid #E2E8F0; }
.exchange-revoke { margin-top: 12px; }

.workspace-gate { margin-bottom: 100px; }
.workspace-gate .button-primary { display: inline-block; margin-top: 24px; }
.market-footer { display: flex; justify-content: space-between; gap: 32px; border-top: 2px solid #0F172A; padding-top: 32px; margin-top: 100px; color: #64748B; font-size: 11px; letter-spacing: 0.15em; text-transform: uppercase; font-weight: 700; }

/* MODAL / DIALOG - SWISS CLEAN */
.modal-backdrop { position: fixed; inset: 0; z-index: 50; background: rgba(15, 23, 42, 0.8); backdrop-filter: blur(4px); display: grid; place-items: center; padding: 24px; }
.application-modal { position: relative; width: min(100%, 560px); max-height: calc(100vh - 48px); overflow: auto; background: #FFFFFF; border: 1px solid #E2E8F0; padding: 48px; box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25); }
.application-modal h2 { font-size: 32px; margin: 0 0 12px; font-weight: 800; letter-spacing: -0.03em; color: #0F172A; }
.modal-context { font-size: 14px; color: #475569; margin-bottom: 32px; line-height: 1.6; }
.application-modal label { display: flex; flex-direction: column; gap: 10px; color: #0F172A; font-size: 13px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; }
.application-modal textarea { width: 100%; background: #F8FAFC; border: 1px solid #CBD5E1; color: #0F172A; padding: 16px; font: inherit; font-size: 15px; resize: vertical; outline: none; transition: border-color 0.2s; box-sizing: border-box; }
.application-modal textarea:focus { border-color: #0BCB78; background: #FFFFFF; }
.application-consent, .application-works { display: flex; flex-direction: column; gap: 12px; margin: 24px 0; padding: 24px; border: 1px solid #E2E8F0; background: #F8FAFC; }
.application-consent legend, .application-works strong { font-size: 12px; color: #0F172A; font-weight: 800; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 8px; display: block; }
.application-modal .application-consent label, .application-modal .application-works label { display: flex; flex-direction: row; align-items: flex-start; gap: 12px; font-size: 14px; text-transform: none; font-weight: 500; letter-spacing: 0; color: #334155; }
.application-consent input, .application-works input { width: 18px; height: 18px; margin-top: 2px; accent-color: #0BCB78; }
.application-consent p { margin: 0; color: #64748B; font-size: 12px; line-height: 1.6; }
.modal-close { position: absolute; right: 24px; top: 24px; border: 0; background: none; color: #94A3B8; font-size: 32px; font-weight: 300; cursor: pointer; transition: color 0.2s; line-height: 1; }
.modal-close:hover { color: #0F172A; }


/* Branched Skill Pack */
.market-intro-panel {
  position: relative;
  left: 50%;
  width: 100vw;
  margin: 0;
  padding: clamp(64px, 7vw, 104px) max(40px, calc((100vw - 1540px) / 2));
  box-sizing: border-box;
  transform: translateX(-50%);
  background: #F8FAFC;
}
.market-intro-inner {
  width: 100%;
  max-width: 1540px;
  margin: 0 auto;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: clamp(36px, 6vw, 88px);
  align-items: start;
}
.market-intro-copy h2 {
  margin: 0 0 18px;
  font-size: clamp(36px, 4vw, 56px);
  line-height: 1.12;
  letter-spacing: -.05em;
  font-weight: 850;
  color: #0F172A;
}
.market-intro-lead {
  margin: 0;
  max-width: 560px;
  color: #475569;
  font-size: 15px;
  line-height: 1.9;
}
.market-intro-points {
  display: grid;
  grid-template-columns: 1fr;
  gap: 18px;
  border-top: 1px solid #CBD5E1;
  padding-top: 22px;
}
.market-intro-point strong {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: #0F172A;
  font-size: 14px;
  font-weight: 800;
  letter-spacing: -.01em;
}
.market-intro-point strong::before {
  content: "";
  width: 10px;
  height: 10px;
  border-radius: 2px;
  background: #0BCB78;
}
.market-intro-point p {
  margin: 8px 0 0;
  color: #64748B;
  font-size: 13px;
  line-height: 1.7;
}

.skill-download-panel {
  position: relative;
  left: 50%;
  width: 100vw;
  max-width: none;
  margin: 0;
  padding: clamp(64px, 7vw, 112px) max(40px, calc((100vw - 1540px) / 2));
  box-sizing: border-box;
  transform: translateX(-50%);
  background: #FFFFFF;
  border: 0;
  box-shadow: none;
  display: grid;
  grid-template-columns: minmax(0, .9fr) minmax(0, 1.1fr);
  gap: clamp(48px, 7vw, 124px);
  align-items: center;
}

.skill-panel-text p.eyebrow { margin-bottom: 16px; }
.skill-panel-text h2 {
  margin: 0 0 24px;
  font-size: clamp(42px, 4.2vw, 68px);
  line-height: 1.15;
  letter-spacing: -0.04em;
  font-weight: 800;
  color: #0F172A;
}
.skill-panel-text h2 span { color: #0BCB78; }
.skill-panel-text > p:not(.eyebrow) {
  margin: 0;
  max-width: 480px;
  color: #475569;
  line-height: 1.7;
  font-size: 15px;
}
.skill-selected-download {
  margin-top: 36px;
  padding-top: 22px;
  border-top: 1px solid #E2E8F0;
}
.skill-selected-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
  color: #64748B;
  font: 10px ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: .12em;
}
.skill-selected-meta span:last-child { color: #0BCB78; }
.skill-selected-download > strong {
  color: #0F172A;
  font-size: 17px;
  font-weight: 700;
}
.skill-selected-download > p {
  margin: 8px 0 18px;
  color: #64748B;
  font-size: 13px;
  line-height: 1.6;
}
.skill-download-cta {
  display: inline-flex;
  min-height: 42px;
  align-items: center;
  justify-content: space-between;
  gap: 36px;
  padding: 0 16px;
  background: #0BCB78;
  color: #06130D;
  font-size: 12px;
  font-weight: 750;
  text-decoration: none;
  transition: background .2s, transform .2s;
}
.skill-download-cta:hover {
  background: #08B96C;
  transform: translateY(-1px);
}
.skill-download-cta span:last-child { font-size: 18px; line-height: 1; }
.skill-branch-menu {
  min-width: 0;
  padding-left: clamp(24px, 3vw, 44px);
  border-left: 1px solid #E2E8F0;
}
.skill-branch-group + .skill-branch-group { margin-top: 24px; }
.skill-branch-heading {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 12px;
  padding: 6px 0 14px;
  border: 0;
  background: transparent;
  color: #334155;
  text-align: left;
  font: 650 11px ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: .14em;
  cursor: pointer;
}
.skill-branch-index { color: #0BCB78; }
.skill-branch-toggle {
  margin-left: auto;
  color: #64748B;
  font: 18px/1 ui-monospace, SFMono-Regular, Menlo, monospace;
}
.skill-branch-children {
  position: relative;
  padding-left: 30px;
}
.skill-branch-lines {
  position: absolute;
  top: 0;
  left: 0;
  overflow: visible;
  pointer-events: none;
}
.skill-branch-trunk,
.skill-branch-path {
  fill: none;
  stroke: #CBD5E1;
  stroke-width: 1.4;
  stroke-linecap: round;
  stroke-linejoin: round;
}
.skill-branch-path { transition: stroke .24s ease; }
.skill-branch-path.is-active { stroke: #0BCB78; }
.skill-branch-item {
  position: relative;
  display: flex;
  width: 100%;
  height: 58px;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 0 8px 0 12px;
  border: 0;
  border-radius: 3px;
  background: transparent;
  color: #475569;
  text-align: left;
  cursor: pointer;
  transition: color .2s, background .2s;
}
.skill-branch-item:hover { color: #0F172A; }
.skill-branch-item.is-active {
  background: linear-gradient(90deg, rgba(11, 203, 120, .11), rgba(11, 203, 120, .025) 90%);
  color: #0F172A;
}
.skill-branch-copy {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 4px;
}
.skill-branch-copy strong { font-size: 13px; font-weight: 650; }
.skill-branch-copy small {
  overflow: hidden;
  color: #64748B;
  font-size: 11px;
  line-height: 1.3;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.skill-branch-item.is-active .skill-branch-copy small { color: #475569; }
.skill-branch-kind {
  flex: none;
  color: #64748B;
  font: 9px ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: .08em;
}
.skill-branch-item.is-active .skill-branch-kind { color: #0BCB78; }

.popular-projects-panel {
  position: relative;
  left: 50%;
  width: 100vw;
  margin: 0;
  padding: clamp(64px, 7vw, 104px) max(40px, calc((100vw - 1540px) / 2));
  box-sizing: border-box;
  transform: translateX(-50%);
  background: #F8FAFC;
  border-top: 0;
}
.popular-projects-inner { width: 100%; max-width: 1540px; margin: 0 auto; }
.popular-projects-heading {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 32px;
  padding-bottom: 20px;
  border-bottom: 2px solid #0F172A;
}
.popular-projects-heading .eyebrow { margin: 0 0 10px; }
.popular-projects-heading h2 {
  margin: 0;
  color: #0F172A;
  font-size: clamp(34px, 3.6vw, 52px);
  font-weight: 800;
  letter-spacing: -.04em;
}
.popular-projects-heading > a {
  flex: none;
  color: #0BCB78;
  font-size: 13px;
  font-weight: 700;
  text-decoration: none;
}
.popular-project-gallery { margin-top: 18px; }
.popular-project-empty {
  padding: 28px 0;
  border-top: 1px solid #CBD5E1;
  color: #64748B;
  font-size: 14px;
}
.discover-people-section {
  position: relative;
  left: 50%;
  width: 100vw;
  margin: 0;
  padding: clamp(64px, 7vw, 104px) max(40px, calc((100vw - 1540px) / 2));
  box-sizing: border-box;
  transform: translateX(-50%);
  background: #FFFFFF;
  border-top: 0;
}
.discover-people-inner { width: 100%; max-width: 1540px; margin: 0 auto; }
.discover-people-section .section-head.compact { margin-bottom: 28px; }
.discover-people-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  column-gap: clamp(28px, 4vw, 64px);
}
.discover-people-grid.is-sparse { grid-template-columns: repeat(auto-fit, minmax(min(100%, 380px), 1fr)); }
.discover-person-row {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 16px;
  padding: 22px 0;
  border-top: 1px solid #CBD5E1;
  color: inherit;
  text-decoration: none;
}
.discover-person-row .person-details { min-width: 0; }
.discover-person-row .person-details strong,
.discover-person-row .person-details span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.discover-person-row .availability { margin-left: auto; }
.discover-people-empty { border-top: 1px solid #CBD5E1; padding: 28px 0; }
.is-discover-shell .market-footer {
  position: relative;
  left: 50%;
  width: 100vw;
  margin: 0;
  padding: 28px max(40px, calc((100vw - 1540px) / 2)) 36px;
  box-sizing: border-box;
  transform: translateX(-50%);
  border-top: 1px solid #CBD5E1;
  background: #F8FAFC;
}

@media (prefers-reduced-motion: reduce) {
  .skill-download-cta,
  .skill-branch-path,
  .skill-branch-item { transition: none; }
}

/* RESPONSIVE ADJUSTMENTS */
@media (max-width: 980px) {
  .market-shell { padding: 0 24px 40px; }
  .is-workspace-shell { padding: 0 24px 40px; }
  .is-projects-shell, .is-builders-shell { padding: 0 24px 40px; }
  .split-sections, .workspace-columns { grid-template-columns: 1fr; gap: 48px; }
  .workspace-overview-grid { grid-template-columns: 1fr; }
  .workspace-overview-side { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .workspace-message-columns { gap: 20px; }
  .work-grid { grid-template-columns: repeat(2, 1fr); }
  .market-intro-panel { padding: 64px 40px; }
  .market-intro-inner { grid-template-columns: 1fr; gap: 32px; }
  .skill-download-panel { grid-template-columns: 1fr; padding: 40px; gap: 40px; }
  .skill-branch-menu { padding: 28px 0 0; border-top: 1px solid #E2E8F0; border-left: 0; }
  .popular-projects-panel { padding: 64px 40px; }
  .discover-people-section { padding: 64px 40px; }
  .discover-people-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0 36px; }
  .discover-people-grid.is-sparse { grid-template-columns: repeat(auto-fit, minmax(min(100%, 320px), 1fr)); }
  .discover-person-row:nth-child(2n + 1) { padding-right: 18px; }
  .discover-person-row:nth-child(2n) { padding-left: 18px; }
  .discover-person-row:nth-child(n + 3) { border-top-color: #CBD5E1; }
  .discover-person-row:nth-child(3):last-child { grid-column: 1 / -1; max-width: calc((100% - 36px) / 2); margin: 0 auto; }
  .is-discover-shell .market-footer { padding: 28px 40px 36px; }
  .market-nav-full { padding: 0 24px; }
  .nav-primary { gap: 24px; }
  .market-nav-full .market-links { gap: 20px; }
  .project-discovery-layout { grid-template-columns: 200px minmax(0, 1fr); gap: 24px; }
  .project-filter-sidebar { padding-right: 16px; }
  .project-card { gap: 14px; padding-right: 4px; padding-left: 4px; }
  .apply-button { padding: 0 10px; }
  .team-member-editor { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .team-member-editor .remove-inline { justify-self: start; margin: -8px 0 0; }
  .plan-editor-head { align-items: flex-start; flex-direction: column; gap: 12px; }
}

@media (max-width: 640px) {
  .market-shell { padding: 0 16px 32px; }
  .is-workspace-shell { padding: 0 16px 32px; }
  .is-projects-shell, .is-builders-shell { padding: 0 16px 32px; }
  .market-brand { font-size: 12px; }
  .account-link, .quiet-link { display: none; }
  .market-actions { gap: 12px; }
  .market-nav-full { padding: 0 14px; }
  .market-nav-full .nav-inner { height: 68px; }
  .market-nav-full .nav-primary { gap: 12px; }
  .market-nav-full .market-brand { gap: 0; }
  .market-nav-full .market-brand span { display: none; }
  .market-nav-full .market-links { gap: 10px; }
  .market-nav-full .market-links a { font-size: 10px; letter-spacing: 0; }
  .market-nav-full .market-actions .account-link { display: inline-flex; min-height: 34px; padding: 0 12px; font-size: 10px; }
  .publish-link { padding: 10px 14px; font-size: 11px; }
  .market-links a { font-size: 12px; }
  .page-heading { padding: 48px 0 32px; }
  .page-heading h1 { font-size: 40px; }
  .is-projects-shell .page-heading { padding: 36px 0 24px; }
  .is-projects-shell .page-heading h1 { font-size: 40px; }
  .is-projects-shell .page-desc { margin-top: 12px; font-size: 15px; }
  .page-desc { font-size: 16px; }
  .hero-brand-lockup { padding: 32px 20px 56px; }
  .hero-brand-lockup h1 { max-width: 100%; font-size: clamp(44px, 12vw, 72px); }
  .hero-brand-lockup p { margin-top: 14px; font-size: 10px; }
  .work-grid, .form-two-col, .agent-token-form, .contact-form { grid-template-columns: 1fr; }
  .workspace-profile-hero { grid-template-columns: 62px minmax(0, 1fr); align-items: start; gap: 14px; padding: 22px 0; }
  .workspace-avatar { width: 58px; height: 58px; border-width: 3px; font-size: 19px; }
  .workspace-profile-copy h2 { font-size: 25px; }
  .workspace-profile-headline { max-width: 100%; font-size: 12px; }
  .workspace-profile-meta { gap: 7px 12px; font-size: 10px; }
  .workspace-profile-actions { grid-column: 1 / -1; justify-content: flex-start; }
  .workspace-logout { margin-left: auto; }
  .workspace-tabs { gap: 20px; margin-bottom: 20px; }
  .workspace-tabs button { min-height: 44px; font-size: 12px; }
  .exchange-box { grid-template-columns: 1fr; }
  .exchange-party + .exchange-party { border-top: 1px solid #E2E8F0; border-left: 0; }
  .workspace-stats { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 9px; }
  .workspace-stats button { padding: 14px; }
  .workspace-stats button > strong { font-size: 26px; }
  .workspace-overview-grid { gap: 12px; margin-top: 12px; }
  .workspace-overview-main, .workspace-overview-side { gap: 12px; }
  .workspace-overview-side { grid-template-columns: 1fr; }
  .workspace-card, .workspace-quick-actions { padding: 16px; }
  .workspace-notification-row { grid-template-columns: 32px minmax(0, 1fr); gap: 9px; }
  .workspace-notification-avatar { width: 30px; height: 30px; }
  .workspace-notification-actions { grid-column: 2; }
  .workspace-notification-copy > strong { white-space: normal; }
  .workspace-inbox-empty { align-items: flex-start; }
  .workspace-small-empty { align-items: flex-start; flex-direction: column; }
  .workspace-section { margin-top: 24px; }
  .workspace-section > .section-head { margin: 0 0 20px; }
  .workspace-message-columns { gap: 18px; }
  .workspace-message-columns .workspace-section { margin-top: 0; }
  .market-intro-panel { padding: 52px 24px; }
  .market-intro-copy h2 { font-size: 34px; }
  .popular-projects-panel { padding: 52px 24px; }
  .popular-projects-heading { align-items: flex-start; flex-direction: column; gap: 16px; }
  .discover-people-section { padding: 52px 24px; }
  .discover-people-grid { grid-template-columns: 1fr; }
  .discover-person-row,
  .discover-person-row:nth-child(2n + 1),
  .discover-person-row:nth-child(2n) { padding: 20px 0; }
  .discover-person-row:nth-child(3):last-child { grid-column: auto; max-width: none; margin: 0; }
  .is-discover-shell .market-footer { padding: 24px 24px 32px; }
  .section-head { margin-top: 56px; flex-direction: column; align-items: flex-start; gap: 16px; }
  .project-search-bar { align-items: stretch; flex-direction: column; gap: 10px; }
  .project-publish-button { min-height: 40px; justify-content: center; }
  .project-discovery-layout { grid-template-columns: 1fr; gap: 20px; }
  .project-filter-sidebar {
    position: static;
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0 12px;
    padding: 0 0 12px;
    border-right: 0;
    border-bottom: 1px solid #E2E8F0;
  }
  .project-filter-heading { grid-column: 1 / -1; margin-bottom: 12px; }
  .project-filter-group { margin-bottom: 8px; }
  .project-filter-group legend { margin-bottom: 5px; }
  .project-filter-option { min-height: 30px; padding: 0 4px; font-size: 11px; }
  .project-results-toolbar { align-items: center; }
  .project-card { grid-template-columns: 1fr; align-items: start; padding: 18px 2px; gap: 14px; }
  .project-results .project-card:hover { margin: 0 -6px; padding-right: 8px; padding-left: 8px; }
  .apply-button { justify-self: start; }
  .project-facts { gap: 12px 16px; }
  .form-panel, .profile-panel, .application-modal { padding: 24px; }
  .project-plan-editor { padding: 18px 14px; }
  .team-member-editor, .milestone-editor .form-two-col, .section-editor .form-two-col { grid-template-columns: 1fr; }
  .milestone-owners { align-items: flex-start; flex-direction: column; }
  .plan-editor-subhead { align-items: flex-start; }
  .skill-download-panel { padding: 24px; gap: 32px; }
  .skill-panel-text h2 { font-size: 36px; }
  .skill-branch-copy small { max-width: 220px; }
  .form-footer, .agent-form-footer, .agent-skill-pack, .profile-panel { align-items: flex-start; flex-direction: column; }
  .invite-row { flex-direction: column; align-items: stretch; gap: 0; }
  .invite-row button { margin: -8px 0 24px; }
  .agent-token-row { flex-direction: column; align-items: flex-start; gap: 16px; }
  .share-options { grid-column: auto; flex-direction: column; gap: 12px; }
  .market-footer { flex-direction: column; gap: 16px; }
}
</style>
