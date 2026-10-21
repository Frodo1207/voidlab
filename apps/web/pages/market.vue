<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import FaultyTerminal from "../components/FaultyTerminal.vue";
import marketLogo from "../../../assets/logo/white-bg.svg";
import {
  clearNetworkToken,
  getNetworkToken,
  networkRequest,
  NetworkRequestError,
  postJSON,
  saveNetworkSession,
  type NetworkContactRequest,
  type NetworkAgentToken,
  type NetworkAgentTokenCreateResult,
  type NetworkOpportunity,
  type NetworkProfile,
  type NetworkSession,
  type NetworkWork
} from "../composables/useBuilderMarket";

const route = useRoute();
const router = useRouter();

const isAuthed = ref(false);
const isAuthPage = computed(() => ["login", "register"].includes(section.value));

onMounted(() => {
  if (getNetworkToken()) {
    isAuthed.value = true;
  }
});

const section = computed(() => typeof route.params.section === "string" ? route.params.section : "discover");
const profiles = ref<NetworkProfile[]>([]);
const works = ref<NetworkWork[]>([]);
const myWorks = ref<NetworkWork[]>([]);
const opportunities = ref<NetworkOpportunity[]>([]);
const myOpportunities = ref<NetworkOpportunity[]>([]);
const profile = ref<NetworkProfile | null>(null);
const incoming = ref<NetworkContactRequest[]>([]);
const outgoing = ref<NetworkContactRequest[]>([]);
const agentTokens = ref<NetworkAgentToken[]>([]);
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
const progressDrafts = reactive<Record<number, string>>({});
const exchange = ref<Record<number, { applicant: Record<string, string>; recipient: Record<string, string> }>>({});
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
  expires_at: ""
});

const sectionTitles: Record<string, { eyebrow: string; title: string; desc: string }> = {
  discover: { eyebrow: "BUILDER NETWORK", title: "让作品成为连接的起点", desc: "在这里发现正在创造的人，也把你正在做的事带到更多人面前。" },
  builders: { eyebrow: "PEOPLE", title: "找 Builder", desc: "通过真实作品、技能和档期，找到适合一起做事的人。" },
  projects: { eyebrow: "OPPORTUNITIES", title: "找项目", desc: "浏览正在招募的合作机会，用清晰的理由发起联系。" },
  publish: { eyebrow: "PUBLISH", title: "发布合作需求", desc: "讲清目标、需要的角色和合作方式，让合适的人主动找到你。" },
  workspace: { eyebrow: "YOUR WORKSPACE", title: "我的工作台", desc: "管理个人档期、联系方式，以及你发出和收到的合作申请。" },
  register: { eyebrow: "INVITE ONLY", title: "加入 BuilderNetwork", desc: "当前为邀请制社区。使用有效邀请码创建你的市场身份。" },
  login: { eyebrow: "WELCOME BACK", title: "登录市场账户", desc: "使用 BuilderNetwork 账户继续你的协作。" }
};
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
    const [nextProfiles, nextWorks, nextOpportunities] = await Promise.all([
      networkRequest<NetworkProfile[]>("/api/v1/public/network/profiles"),
      networkRequest<NetworkWork[]>("/api/v1/public/network/works"),
      networkRequest<NetworkOpportunity[]>("/api/v1/public/network/opportunities")
    ]);
    profiles.value = nextProfiles;
    works.value = nextWorks;
    opportunities.value = nextOpportunities;
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
    if (incomingResult.status === "fulfilled") incoming.value = incomingResult.value;
    if (outgoingResult.status === "fulfilled") outgoing.value = outgoingResult.value;
    if (worksResult.status === "fulfilled") myWorks.value = worksResult.value;
    if (opportunitiesResult.status === "fulfilled") myOpportunities.value = opportunitiesResult.value;
    if (tokensResult.status === "fulfilled") agentTokens.value = tokensResult.value;
    for (const request of [...incoming.value, ...outgoing.value]) {
      if (request.progress_status) progressDrafts[request.id] = request.progress_status;
    }
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
      expires_at: opportunityForm.expires_at ? new Date(`${opportunityForm.expires_at}T23:59:59`).toISOString() : ""
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
  try {
    await postJSON(`/api/v1/network/contact-requests/${id}/${action}`, {}, true);
    await loadWorkspace();
    if (action === "accept") notice.value = "已接受申请。双方授权的联系方式现在可以查看。";
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "申请处理失败";
  }
}

async function showExchange(id: number) {
  error.value = "";
  try {
    exchange.value[id] = await networkRequest(`/api/v1/network/contact-exchanges/${id}`, {}, true);
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "联系方式读取失败";
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

async function updateRequestProgress(request: NetworkContactRequest) {
  const progress = progressDrafts[request.id] || request.progress_status;
  if (!progress) {
    error.value = "请选择一个合作进度。";
    return;
  }
  error.value = "";
  try {
    await networkRequest(`/api/v1/network/contact-requests/${request.id}/progress`, {
      method: "PUT",
      body: JSON.stringify({ progress_status: progress })
    }, true);
    await loadWorkspace();
    notice.value = "合作进度已更新，双方都能看到。";
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "进度更新失败";
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
    expires_at: opportunity.expires_at ? opportunity.expires_at.slice(0, 10) : ""
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
    <div class="hero-full-bg-glow"></div>
    <div class="hero-full-bg-code" aria-hidden="true">
      <pre><code>type NetworkContactRequest struct {
    ID              int64   `json:"id"`
    FromProfileID   int64   `json:"from_profile_id"`
    TargetType      string  `json:"target_type"`
    TargetID        int64   `json:"target_id"`
    Message         string  `json:"message"`
    ShareEmail      bool    `json:"share_email"`
    ShareWeChat     bool    `json:"share_wechat"`
    Status          string  `json:"status"`
}

func (s *BuilderNetworkService) CreateContactRequest(
    accountID int64, input domain.NetworkContactRequest,
) (*domain.NetworkContactRequest, error) {
    profile, err := s.repo.GetProfileByAccountID(accountID)
    // ...
}</code></pre>
    </div>

    <!-- Integrated Header for Hero -->
    <header class="market-nav-full">
      <div class="nav-inner">
        <RouterLink class="market-brand" to="/market">
          <img class="market-logo" :src="marketLogo" alt="VOIDLAB Logo" width="32" height="32" />
          <span>VOIDLAB <b>NETWORK</b></span>
        </RouterLink>
        <nav class="market-links" aria-label="市场导航">
          <RouterLink to="/market">发现</RouterLink>
          <RouterLink to="/market/builders">找 Builder</RouterLink>
          <RouterLink to="/market/projects">找项目</RouterLink>
        </nav>
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

    <!-- Hero Content -->
    <div class="hero-full-content">
      <!-- Left side visual filler: Large typography & grids -->
      <div class="hero-full-visual-left" aria-hidden="true">
        <div class="giant-letter">V</div>
        <div class="engineering-grid">
          <div class="grid-line horizontal" style="top: 20%"></div>
          <div class="grid-line horizontal" style="top: 60%"></div>
          <div class="grid-line vertical" style="left: 30%"></div>
          <div class="grid-measure">
            <span>AXIS-Y: 48.092</span>
            <span>CAPACITY: 100%</span>
          </div>
        </div>
      </div>

      <!-- Right side text -->
      <div class="hero-full-text">
        <span class="hero-kicker">BUILDER NETWORK / 01</span>
        <h2>作品让能力被看见，<br />合作让想法<em>继续发生。</em></h2>
        <p>展示正在做的事，找到合适的伙伴，<br />也让真正想合作的人带着理由来联系你。</p>
        <div class="hero-buttons">
          <RouterLink class="button-primary-bold" to="/market/register">创建你的身份</RouterLink>
          <RouterLink class="button-secondary-outline" to="/market/projects">浏览合作机会</RouterLink>
        </div>
      </div>
    </div>

    <!-- Ticker -->
    <div class="hero-full-ticker">
      <div class="ticker-track">
        <span>agent identity <strong>ONLINE</strong></span>
        <span>AGENTS Service providers bidding <strong>LIVE</strong></span>
        <span>ESCROW + PAYOUTS <strong>READY</strong></span>
        <span>ONCHAIN OS <strong>ONLINE</strong></span>
        <span>Agent identity <strong>ONLINE</strong></span>
        <span>AGENTS Service providers bidding <strong>LIVE</strong></span>
        <span>ESCROW + PAYOUTS <strong>READY</strong></span>
        <span>ONCHAIN OS <strong>ONLINE</strong></span>
      </div>
    </div>
  </div>

  <div v-if="!isAuthPage" class="market-shell" :class="{'is-discover-shell': section === 'discover'}">

    <template v-if="section !== 'discover'">
      <header class="market-nav">
        <RouterLink class="market-brand" to="/market">
          <img class="market-logo" :src="marketLogo" alt="VOIDLAB Logo" width="32" height="32" />
          <span>VOIDLAB <b>NETWORK</b></span>
        </RouterLink>
        <nav class="market-links" aria-label="市场导航">
          <RouterLink to="/market">发现</RouterLink>
          <RouterLink to="/market/builders">找 Builder</RouterLink>
          <RouterLink to="/market/projects">找项目</RouterLink>
        </nav>
        <div class="market-actions">
          <template v-if="isAuthed">
            <RouterLink class="quiet-link" to="/market/workspace">工作台</RouterLink>
            <RouterLink class="publish-link" to="/market/publish">＋ 发布需求</RouterLink>
          </template>
          <template v-else>
            <RouterLink class="account-link" to="/market/register">加入</RouterLink>
          </template>
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

        <section class="skill-download-panel">
          <div class="skill-panel-text">
            <p class="eyebrow">AGENT SKILL PACK</p>
            <h2>把市场装进<br />你的终端</h2>
            <p>将 BuilderNetwork Skill 集成到你的日常 Agent 中。直接通过 API 发现合作机会、管理作品与档期，无需离开当前工作流。</p>
          </div>

          <div class="skill-download-grid">
            <a class="download-card" href="/downloads/voidlab-buildernetwork-agent/README.md" download>
              <div class="dl-icon">TXT</div>
              <div class="dl-info">
                <strong>安装与鉴权说明</strong>
                <span>README.md / 部署指南</span>
              </div>
              <div class="dl-arrow">&darr;</div>
            </a>

            <a class="download-card" href="/downloads/voidlab-buildernetwork-agent/trae/voidlab-buildernetwork-agent/SKILL.md" download>
              <div class="dl-icon">TRAE</div>
              <div class="dl-info">
                <strong>Trae Skill</strong>
                <span>SKILL.md / 原生指令支持</span>
              </div>
              <div class="dl-arrow">&darr;</div>
            </a>

            <a class="download-card" href="/downloads/voidlab-buildernetwork-agent/codex/AGENTS.md" download>
              <div class="dl-icon">CDX</div>
              <div class="dl-info">
                <strong>Codex Bridge</strong>
                <span>AGENTS.md / Cursor 桥接</span>
              </div>
              <div class="dl-arrow">&darr;</div>
            </a>

            <a class="download-card" href="/downloads/voidlab-buildernetwork-agent/claude-code/CLAUDE.md" download>
              <div class="dl-icon">CLD</div>
              <div class="dl-info">
                <strong>Claude Code Bridge</strong>
                <span>CLAUDE.md / CLI 终端接入</span>
              </div>
              <div class="dl-arrow">&darr;</div>
            </a>
          </div>
        </section>

        <div class="section-head">
          <div><p class="eyebrow">COMMUNITY WORK</p><h2>最近的作品</h2></div>
          <RouterLink to="/market/builders">发现创作者 ↗</RouterLink>
        </div>
        <div v-if="loading" class="empty-state">正在加载市场内容…</div>
        <div v-else-if="works.length" class="work-grid">
          <article v-for="work in works.slice(0, 3)" :key="work.id" class="work-card">
            <div v-if="work.cover_url" class="work-cover"><img :src="work.cover_url" :alt="work.title" /></div>
            <div v-else class="work-cover cover-abstract"><span>WORK / {{ String(work.id).padStart(2, "0") }}</span><i>↗</i></div>
            <div class="work-card-body">
              <p class="card-meta">{{ work.author_name || "COMMUNITY BUILDER" }}</p>
              <h3>{{ work.title }}</h3>
              <p>{{ work.summary }}</p>
              <div class="tag-row"><span v-for="tag in (work.tags || []).slice(0, 3)" :key="tag">{{ tag }}</span></div>
            </div>
          </article>
        </div>
        <div v-else class="empty-state">还没有公开作品。注册后发布你的第一个项目，让能力有迹可循。</div>

        <div class="split-sections">
          <section>
            <div class="section-head compact"><div><p class="eyebrow">PEOPLE TO KNOW</p><h2>活跃 Builder</h2></div><RouterLink to="/market/builders">全部 Builder ↗</RouterLink></div>
            <article v-for="person in profiles.slice(0, 3)" :key="person.id" class="person-row">
              <span class="person-avatar">{{ person.display_name.slice(0, 1) }}</span>
              <div class="person-details"><strong>{{ person.display_name }}</strong><span>{{ person.primary_role }} · {{ person.city || "远程" }}</span></div>
              <span class="availability" :class="person.availability">{{ person.availability === "unavailable" ? "暂不接单" : person.availability === "limited" ? "有限档期" : "开放合作" }}</span>
            </article>
            <div v-if="!profiles.length && !loading" class="empty-inline">社区成员即将加入</div>
          </section>
          <section>
            <div class="section-head compact"><div><p class="eyebrow">OPEN CALLS</p><h2>正在招募</h2></div><RouterLink to="/market/projects">全部需求 ↗</RouterLink></div>
            <RouterLink v-for="item in opportunities.slice(0, 3)" :key="item.id" class="opportunity-row" :to="`/market/projects`">
              <span class="opportunity-symbol">↗</span>
              <span class="opportunity-text"><strong>{{ item.title }}</strong><small>{{ item.looking_for_roles.join(" / ") }} · {{ item.remote_allowed ? "远程协作" : item.city_scope }}</small></span>
              <span class="row-arrow">→</span>
            </RouterLink>
            <div v-if="!opportunities.length && !loading" class="empty-inline">暂时没有开放中的合作机会</div>
          </section>
        </div>
      </template>

      <template v-else-if="section === 'builders'">
        <div class="list-toolbar"><span>{{ profiles.length }} 位公开 Builder</span><RouterLink to="/market/register" class="text-action">完善我的档案 ↗</RouterLink></div>
        <div class="builder-grid">
          <article v-for="person in profiles" :key="person.id" class="builder-card">
            <div class="builder-card-top"><span class="person-avatar large">{{ person.display_name.slice(0, 1) }}</span><span class="availability" :class="person.availability">{{ person.availability === "unavailable" ? "暂不接单" : person.availability === "limited" ? "有限档期" : "开放合作" }}</span></div>
            <p class="eyebrow">{{ person.primary_role }} <span v-if="person.city">/ {{ person.city }}</span></p>
            <h3>{{ person.display_name }}</h3>
            <p class="builder-headline">{{ person.headline || "正在寻找值得投入的合作。" }}</p>
            <div class="tag-row"><span v-for="tag in (person.skills || []).slice(0, 4)" :key="tag">{{ tag }}</span></div>
            <button class="button-secondary full-button" @click="openApplication('profile', person.id, person.display_name)">发起合作申请</button>
          </article>
          <div v-if="!profiles.length && !loading" class="empty-state wide">目前还没有公开 Builder，邀请你的第一位协作者加入。</div>
        </div>
      </template>

      <template v-else-if="section === 'projects'">
        <div class="list-toolbar"><span>{{ opportunities.length }} 个开放合作机会</span><RouterLink to="/market/publish" class="text-action">发布一个需求 ↗</RouterLink></div>
        <div class="project-list">
          <article v-for="item in opportunities" :key="item.id" class="project-card">
            <div class="project-main">
              <div class="project-topline"><span class="project-type">{{ item.opportunity_type }}</span><span v-if="item.remote_allowed" class="remote-label">REMOTE FRIENDLY</span></div>
              <h3>{{ item.title }}</h3>
              <p>{{ item.summary }}</p>
              <div class="project-facts"><span>需求方：{{ item.owner_name }}</span><span>{{ item.engagement_mode || "合作方式待沟通" }}</span><span>{{ item.budget_hint || "预算未公开" }}</span></div>
              <div class="tag-row"><span v-for="role in item.looking_for_roles" :key="role">{{ role }}</span><span v-for="skill in item.required_skills.slice(0, 4)" :key="skill">{{ skill }}</span></div>
            </div>
            <button class="button-primary apply-button" @click="openApplication('opportunity', item.id, item.title)">提交申请</button>
          </article>
          <div v-if="!opportunities.length && !loading" class="empty-state">目前没有开放中的合作机会。你可以发布需求，邀请合适的人加入。</div>
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
          <section class="profile-panel">
            <div><p class="eyebrow">YOUR PROFILE</p><h2>{{ profile.display_name }}</h2><p>{{ profile.primary_role }} · {{ profile.headline || "补充一句介绍，让合作者更了解你。" }}</p></div>
            <div class="profile-actions"><span class="availability" :class="profile.availability">{{ profile.availability === "unavailable" ? "暂不接单" : profile.availability === "limited" ? "有限档期" : "开放合作" }}</span><button class="quiet-button" @click="logout">退出登录</button></div>
          </section>
          <section class="workspace-section">
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
          <section class="workspace-section">
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
          <section class="workspace-section">
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
          <section class="workspace-section">
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
          <section class="workspace-section">
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
          <div class="workspace-columns">
            <section class="workspace-section">
              <div class="section-head"><div><p class="eyebrow">INCOMING</p><h2>收到的申请 <small>{{ incoming.length }}</small></h2></div></div>
              <article v-for="request in incoming" :key="request.id" class="request-card">
                <div class="request-card-head"><strong>{{ request.applicant_name }}</strong><span class="request-status">{{ request.status }}</span></div>
                <p class="request-role">{{ request.applicant_role }}</p>
                <p>{{ request.message }}</p>
                <div v-if="request.status === 'pending'" class="request-actions"><button class="button-primary small-button" @click="processRequest(request.id, 'accept')">接受并交换</button><button class="quiet-button" @click="processRequest(request.id, 'decline')">婉拒</button></div>
                <template v-else-if="request.status === 'accepted'">
                  <div class="progress-tracker">
                    <div class="progress-step" :class="{ completed: ['contacted', 'discussing', 'collaborating', 'completed', 'closed'].includes(request.progress_status) }">已联系</div>
                    <div class="progress-step" :class="{ completed: ['discussing', 'collaborating', 'completed', 'closed'].includes(request.progress_status), active: request.progress_status === 'contacted' }">沟通中</div>
                    <div class="progress-step" :class="{ completed: ['collaborating', 'completed', 'closed'].includes(request.progress_status), active: request.progress_status === 'discussing' }">合作中</div>
                    <div class="progress-step" :class="{ completed: ['completed', 'closed'].includes(request.progress_status), active: request.progress_status === 'collaborating' }">已结束</div>
                  </div>
                  <div class="progress-control"><span>更新状态</span><select v-model="progressDrafts[request.id]"><option value="">选择新进度</option><option value="contacted">已联系</option><option value="discussing">沟通中</option><option value="collaborating">合作中</option><option value="completed">已完成</option><option value="closed">已结束</option></select><button class="quiet-button" type="button" @click="updateRequestProgress(request)">保存</button></div>
                  <div class="request-actions"><button class="text-action" @click="showExchange(request.id)">查看已授权联系方式</button><button v-if="!request.exchange_revoked" class="quiet-button" @click="revokeExchange(request.id)">撤回本次授权</button><span v-else class="revoked-label">联系方式授权已撤回</span></div>
                </template>
                <div v-if="exchange[request.id]" class="exchange-box"><span v-for="(value, key) in exchange[request.id].applicant" :key="key"><b>{{ key }}</b> {{ value }}</span></div>
              </article>
              <div v-if="!incoming.length" class="empty-inline">暂时没有收到申请</div>
            </section>
            <section class="workspace-section">
              <div class="section-head"><div><p class="eyebrow">OUTGOING</p><h2>我发出的申请 <small>{{ outgoing.length }}</small></h2></div></div>
              <article v-for="request in outgoing" :key="request.id" class="request-card">
                <div class="request-card-head"><strong>{{ request.target_type === "opportunity" ? "合作机会申请" : "Builder 合作申请" }}</strong><span class="request-status">{{ request.status }}</span></div>
                <p>{{ request.message }}</p>
                <button v-if="request.status === 'pending'" class="quiet-button" @click="processRequest(request.id, 'withdraw')">撤回申请</button>
                <template v-if="request.status === 'accepted'">
                  <div class="progress-tracker">
                    <div class="progress-step" :class="{ completed: ['contacted', 'discussing', 'collaborating', 'completed', 'closed'].includes(request.progress_status) }">已联系</div>
                    <div class="progress-step" :class="{ completed: ['discussing', 'collaborating', 'completed', 'closed'].includes(request.progress_status), active: request.progress_status === 'contacted' }">沟通中</div>
                    <div class="progress-step" :class="{ completed: ['collaborating', 'completed', 'closed'].includes(request.progress_status), active: request.progress_status === 'discussing' }">合作中</div>
                    <div class="progress-step" :class="{ completed: ['completed', 'closed'].includes(request.progress_status), active: request.progress_status === 'collaborating' }">已结束</div>
                  </div>
                  <div class="progress-control"><span>更新状态</span><select v-model="progressDrafts[request.id]"><option value="">选择新进度</option><option value="contacted">已联系</option><option value="discussing">沟通中</option><option value="collaborating">合作中</option><option value="completed">已完成</option><option value="closed">已结束</option></select><button class="quiet-button" type="button" @click="updateRequestProgress(request)">保存</button></div>
                  <div class="request-actions"><button class="text-action" @click="showExchange(request.id)">查看已授权联系方式</button><button v-if="!request.exchange_revoked" class="quiet-button" @click="revokeExchange(request.id)">撤回本次授权</button><span v-else class="revoked-label">联系方式授权已撤回</span></div>
                </template>
                <div v-if="exchange[request.id]" class="exchange-box"><span v-for="(value, key) in exchange[request.id].recipient" :key="key"><b>{{ key }}</b> {{ value }}</span></div>
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
.market-shell { max-width: 1280px; margin: 0 auto; padding: 0 40px 60px; }
.market-nav { height: 80px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid #E2E8F0; gap: 24px; }
.market-brand { display: flex; align-items: center; gap: 12px; color: #0F172A; text-decoration: none; font-size: 14px; letter-spacing: 0.15em; font-weight: 800; text-transform: uppercase; }
.market-brand b { color: #2563EB; font-weight: 800; }
.market-logo { width: 32px; height: 32px; flex-shrink: 0; object-fit: contain; border-radius: 7px; }
.market-links, .market-actions { display: flex; align-items: center; gap: 32px; }
.market-links a, .quiet-link, .account-link { color: #64748B; text-decoration: none; font-size: 13px; font-weight: 600; letter-spacing: 0.05em; text-transform: uppercase; transition: color 0.2s; }
.market-links a:hover, .market-links a.router-link-active, .quiet-link:hover, .account-link:hover { color: #0F172A; }
.publish-link, .button-primary { background: #2563EB; color: #FFFFFF; text-decoration: none; border: 0; padding: 12px 24px; font-size: 13px; font-weight: 700; letter-spacing: 0.05em; text-transform: uppercase; cursor: pointer; transition: background 0.2s; }
.publish-link:hover, .button-primary:hover { background: #1D4ED8; }
.button-secondary { display: inline-flex; align-items: center; justify-content: center; border: 1px solid #CBD5E1; color: #0F172A; background: transparent; text-decoration: none; padding: 11px 23px; font-size: 13px; font-weight: 700; letter-spacing: 0.05em; text-transform: uppercase; cursor: pointer; transition: border-color 0.2s, color 0.2s; }
.button-secondary:hover { border-color: #0F172A; color: #0F172A; }
.account-link { border: 1px solid #CBD5E1; padding: 11px 16px; }

/* TYPOGRAPHY & HEADINGS */
.page-heading { padding: 80px 0 60px; }
.eyebrow { margin: 0 0 16px; color: #2563EB; font-size: 11px; letter-spacing: 0.2em; font-weight: 700; text-transform: uppercase; }
.page-heading h1 { font-size: clamp(42px, 6vw, 72px); line-height: 1.05; letter-spacing: -0.04em; margin: 0; font-weight: 800; color: #0F172A; }
.page-desc { margin: 24px 0 0; color: #475569; font-size: 18px; line-height: 1.6; max-width: 680px; }

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
  background: #FFFFFF;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  margin-left: calc(-50vw + 50%); /* Break out of any parent container if needed, though we moved it out */
}

/* Background elements (Light theme) */
.hero-full-bg-glow {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at 60% 40%, rgba(37, 99, 235, 0.05), transparent 50%),
    radial-gradient(circle at 30% 60%, rgba(147, 197, 253, 0.08), transparent 40%);
  z-index: 1;
  pointer-events: none;
}
.hero-full-bg-glow::after {
  content: "";
  position: absolute;
  inset: 0;
  background: url('data:image/svg+xml;utf8,<svg viewBox="0 0 200 200" xmlns="http://www.w3.org/2000/svg"><filter id="noiseFilter"><feTurbulence type="fractalNoise" baseFrequency="0.65" numOctaves="3" stitchTiles="stitch"/></filter><rect width="100%" height="100%" filter="url(%23noiseFilter)"/></svg>');
  opacity: 0.02; /* Very subtle noise for light theme */
  mix-blend-mode: multiply;
  pointer-events: none;
}

.hero-full-bg-code {
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
  position: relative;
  z-index: 10;
  width: 100%;
  padding: 0 40px;
  box-sizing: border-box;
}
.nav-inner {
  max-width: 1200px;
  margin: 0 auto;
  height: 80px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(15, 23, 42, 0.08); /* Subtle dark border */
}

/* Hero Content - Right Aligned */
.hero-full-content {
  position: relative;
  z-index: 2;
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 1fr;
  align-items: center;
  gap: 60px;
  max-width: 1280px;
  margin: 0 auto;
  width: 100%;
  padding: 0 40px;
  box-sizing: border-box;
}

/* Hero Left Visual Elements (Magazine / Engineering feel) */
.hero-full-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  align-items: center;
  gap: 40px;
}

.hero-full-visual-left {
  position: relative;
  height: 100%;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.giant-letter {
  font-size: clamp(200px, 30vw, 400px);
  font-family: "Helvetica Neue", Helvetica, Arial, sans-serif;
  font-weight: 800;
  color: #F1F5F9; /* Very subtle light gray */
  line-height: 0.8;
  letter-spacing: -0.05em;
  user-select: none;
  z-index: 1;
}

.engineering-grid {
  position: absolute;
  inset: 0;
  z-index: 2;
  pointer-events: none;
}

.grid-line {
  position: absolute;
  background: #E2E8F0;
}
.grid-line.horizontal {
  width: 100%;
  height: 1px;
}
.grid-line.vertical {
  height: 100%;
  width: 1px;
}

.grid-measure {
  position: absolute;
  bottom: 20%;
  left: 32%;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.grid-measure span {
  font-family: ui-monospace, monospace;
  font-size: 10px;
  color: #94A3B8;
  letter-spacing: 0.1em;
}

.hero-full-text p {
  color: #475569;
  font-size: 18px;
  line-height: 1.7;
  margin: 0 0 48px;
  font-weight: 400;
}
.hero-full-text h2 em {
  font-family: "Times New Roman", Times, serif;
  color: #64748B;
  font-weight: 400;
  font-style: italic;
}

.hero-full-text {
  max-width: 650px;
  padding-right: 40px;
}

.hero-full-text .hero-kicker {
  color: #2563EB; /* Blue kicker */
  font-size: 14px;
  letter-spacing: 0.1em;
  font-weight: 600;
  margin-bottom: 24px;
  display: inline-block;
  cursor: pointer;
  transition: color 0.2s;
}
.hero-full-text .hero-kicker:hover {
  color: #1D4ED8;
}

.hero-full-text h2 {
  font-size: clamp(48px, 6vw, 76px);
  line-height: 1.15;
  letter-spacing: -0.02em;
  margin: 0 0 48px;
  font-weight: 800;
  color: #0F172A; /* Dark Slate for text */
  font-family: "Helvetica Neue", Helvetica, Arial, "PingFang SC", sans-serif;
  text-shadow: none;
}

.hero-full-text .hero-buttons {
  display: flex;
  gap: 16px;
}

.button-primary-bold {
  background: #0F172A; /* Dark button on white bg */
  color: #FFFFFF;
  text-decoration: none;
  border: 0;
  padding: 16px 32px;
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0.05em;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}
.button-primary-bold:hover {
  background: #1E293B;
  transform: translateY(-1px);
  box-shadow: 0 10px 25px rgba(15, 23, 42, 0.15);
}

.button-secondary-outline {
  background: transparent;
  color: #0F172A;
  text-decoration: none;
  border: 1px solid #CBD5E1;
  padding: 15px 31px;
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.05em;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}
.button-secondary-outline:hover {
  background: #F8FAFC;
  border-color: #94A3B8;
}

/* Ticker at the bottom */
.hero-full-ticker {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  background: #F8FAFC; /* Light gray ticker bg */
  border-top: 1px solid #E2E8F0;
  padding: 16px 0;
  overflow: hidden;
  z-index: 2;
}
.ticker-track {
  display: flex;
  width: max-content;
  animation: ticker 30s linear infinite;
}
.ticker-track span {
  font-family: ui-monospace, monospace;
  font-size: 12px;
  color: #64748B;
  padding: 0 40px;
  letter-spacing: 0.05em;
  white-space: nowrap;
}
.ticker-track strong {
  color: #2563EB; /* Blue for strong words */
  font-weight: 700;
  margin-left: 8px;
}

@keyframes ticker {
  0% { transform: translateX(0); }
  100% { transform: translateX(-50%); }
}

/* Spacing adjustments for discover page */
.is-discover-shell {
  padding-top: 60px; /* Space after the full height hero */
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
  background-image: radial-gradient(circle at 20% 30%, rgba(37, 99, 235, 0.15) 0%, transparent 50%),
                    radial-gradient(circle at 80% 80%, rgba(147, 51, 234, 0.1) 0%, transparent 50%);
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
.auth-kicker { font: 10px ui-monospace, monospace; letter-spacing: .16em; color: #a6bbff; }
.terminal-text-container h1 { font-size: clamp(36px, 4vw, 68px); line-height: 1.25; letter-spacing: -.04em; margin: 24px 0; }
.auth-brand-desc { color: #c0cbe4; font-size: 17px; line-height: 1.9; }
.auth-visual-footer { display: flex; justify-content: space-between; color: #a6b4d1; font-size: 11px; letter-spacing: .08em; }
.auth-back { color: #64748b; text-decoration: none; font-size: 13px; display: inline-block; margin-bottom: 40px; }
.auth-form-container h2 { font-size: 30px; letter-spacing: -.04em; margin: 0 0 12px; }
.auth-intro { color: #64748b; font-size: 14px; line-height: 1.7; margin: 0 0 28px; }
.auth-form-note { font-size: 12px; line-height: 1.7; color: #64748b; margin: 0; }
.auth-form-note a { color: #2563eb; text-decoration: none; }
.auth-invite-row { display: flex; align-items: flex-end; gap: 16px; }
.auth-invite-row label { flex: 1; min-width: 0; }
.auth-verify { flex-shrink: 0; padding: 10px 0; border: 0; background: transparent; color: #2563eb; cursor: pointer; }

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
  text-shadow: -2px 0 #2563EB;
  clip: rect(44px, 450px, 56px, 0);
  animation: glitch-anim 5s infinite linear alternate-reverse;
}
.glitch::after {
  left: -2px;
  text-shadow: -2px 0 #9333EA;
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
  color: #38BDF8;
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
  border-bottom-color: #2563EB;
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
  border-bottom-color: #2563EB;
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
.section-head a, .text-action { color: #2563EB; text-decoration: none; font-size: 13px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; }
.section-head .eyebrow { margin-bottom: 12px; }

/* CARDS - SWISS CLEAN */
.work-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 32px; }
.work-card, .builder-card, .project-card { background: #FFFFFF; border: 1px solid #E2E8F0; display: flex; flex-direction: column; transition: transform 0.2s, box-shadow 0.2s; }
.work-card:hover, .builder-card:hover, .project-card:hover { transform: translateY(-2px); box-shadow: 0 12px 24px rgba(15, 23, 42, 0.04); }
.work-cover { height: 220px; background: #F1F5F9; overflow: hidden; border-bottom: 1px solid #E2E8F0; }
.work-cover img { width: 100%; height: 100%; object-fit: cover; display: block; }
.cover-abstract { display: flex; align-items: flex-end; justify-content: space-between; padding: 24px; background: linear-gradient(135deg, #F8FAFC, #E2E8F0); height: 100%; box-sizing: border-box; }
.cover-abstract span { font-size: 11px; letter-spacing: 0.2em; color: #64748B; font-weight: 700; }
.cover-abstract i { font-size: 32px; font-style: normal; color: #94A3B8; font-weight: 300; }
.work-card-body { padding: 24px; display: flex; flex-direction: column; flex: 1; }
.card-meta { font-size: 10px; letter-spacing: 0.15em; color: #64748B; margin: 0 0 12px; text-transform: uppercase; font-weight: 700; }
.work-card h3, .builder-card h3, .project-card h3 { font-size: 22px; letter-spacing: -0.03em; margin: 0 0 12px; font-weight: 800; color: #0F172A; }
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
.builder-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 32px; }
.builder-card { padding: 32px; }
.builder-card-top { display: flex; align-items: center; justify-content: space-between; margin-bottom: 32px; }
.person-avatar.large { width: 72px; height: 72px; font-size: 32px; }
.builder-card .eyebrow { color: #64748B; }
.builder-card h3 { font-size: 28px; }
.builder-headline { font-size: 14px; color: #475569; line-height: 1.6; min-height: 48px; margin-bottom: 24px; }
.full-button { width: 100%; margin-top: auto; }

.project-list { display: flex; flex-direction: column; gap: 24px; }
.project-card { display: flex; align-items: center; justify-content: space-between; gap: 40px; padding: 40px; }
.project-main { max-width: 800px; }
.project-topline { display: flex; align-items: center; gap: 16px; margin-bottom: 16px; }
.project-type { font-size: 11px; color: #2563EB; font-weight: 800; text-transform: uppercase; letter-spacing: 0.15em; }
.remote-label { font-size: 10px; color: #475569; border: 1px solid #CBD5E1; padding: 4px 8px; font-weight: 700; background: #F1F5F9; }
.project-card h3 { font-size: 28px; margin-bottom: 16px; }
.project-card p { font-size: 15px; color: #475569; line-height: 1.7; margin: 0; }
.project-facts { display: flex; flex-wrap: wrap; gap: 24px; margin-top: 24px; color: #64748B; font-size: 12px; font-weight: 600; text-transform: uppercase; }
.apply-button { white-space: nowrap; padding: 16px 32px; font-size: 14px; }

/* FORMS - INFRASTRUCTURE CLEAN */
.form-panel { max-width: 800px; background: #FFFFFF; border: 1px solid #E2E8F0; padding: 48px; box-shadow: 0 10px 30px rgba(15,23,42,0.03); }
.publish-form { margin: 0 auto 80px; }
.auth-panel { max-width: 640px; margin: 0 auto 80px; }
.form-panel label { display: flex; flex-direction: column; gap: 10px; color: #0F172A; font-size: 13px; font-weight: 700; margin-bottom: 24px; text-transform: uppercase; letter-spacing: 0.05em; }
.form-panel input, .form-panel select, .form-panel textarea, .contact-form input { width: 100%; border: 1px solid #CBD5E1; background: #F8FAFC; color: #0F172A; padding: 14px 16px; font: inherit; outline: none; font-size: 15px; text-transform: none; font-weight: 400; transition: border-color 0.2s, background 0.2s; box-sizing: border-box; }
.form-panel input:focus, .form-panel select:focus, .form-panel textarea:focus, .contact-form input:focus { border-color: #2563EB; background: #FFFFFF; box-shadow: 0 0 0 3px rgba(37,99,235,0.1); }
.form-panel textarea { resize: vertical; line-height: 1.6; }
.form-two-col { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; }
.checkbox-label { flex-direction: row !important; align-items: center; justify-content: flex-start; text-transform: none !important; font-weight: 500 !important; font-size: 14px !important; letter-spacing: 0 !important; }
.checkbox-label input, .share-options input { width: 18px; height: 18px; accent-color: #2563EB; }
.form-footer { display: flex; align-items: center; justify-content: space-between; gap: 24px; border-top: 1px solid #E2E8F0; padding-top: 32px; margin-top: 16px; }
.form-footer span, .privacy-note { font-size: 13px; color: #64748B; line-height: 1.6; font-weight: 400; }
.invite-row { display: flex; align-items: flex-end; gap: 16px; }
.invite-row label { flex: 1; }
.invite-row button { margin-bottom: 24px; padding: 14px 24px; }
.verified-mark { display: inline-block; color: #16A34A; font-size: 13px; font-weight: 600; margin: -12px 0 20px; }
.privacy-note { margin: 16px 0; padding: 16px; background: #F1F5F9; border-left: 3px solid #CBD5E1; }
.form-switch { text-align: center; font-size: 14px; color: #64748B; margin: 32px 0 0; font-weight: 500; }
.form-switch a { color: #2563EB; font-weight: 700; text-decoration: none; }
.form-switch a:hover { text-decoration: underline; }

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
.workspace-section { margin-top: 64px; }
.agent-intro { max-width: 800px; margin: 0 0 24px; color: #475569; font-size: 15px; line-height: 1.6; }
.agent-token-form { max-width: none; display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 24px; }
.agent-token-form > label, .agent-form-footer { grid-column: 1 / -1; }
.agent-scope-group { border-top: 1px solid #E2E8F0; padding-top: 24px; }
.agent-scope-group h3 { font-size: 14px; color: #0F172A; margin: 0 0 16px; font-weight: 800; text-transform: uppercase; }
.agent-scope-option { display: flex !important; flex-direction: row !important; align-items: flex-start; gap: 12px !important; margin: 0 0 12px !important; padding: 12px; background: #F8FAFC; border: 1px solid #E2E8F0; color: #0F172A !important; text-transform: none !important; }
.agent-scope-option input { width: 16px !important; height: 16px !important; margin-top: 2px; accent-color: #2563EB; }
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
.agent-token-title span { font-size: 10px; color: #2563EB; border: 1px solid #2563EB; padding: 4px 8px; font-weight: 700; text-transform: uppercase; }
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
.opportunity-manage-actions select, .progress-control select { background: #F8FAFC; border: 1px solid #CBD5E1; color: #0F172A; padding: 10px 14px; font-size: 12px; font-weight: 600; outline: none; }
.opportunity-manage-actions select:focus, .progress-control select:focus { border-color: #2563EB; }

.contact-form { display: grid; grid-template-columns: repeat(3, 1fr); gap: 24px; background: #FFFFFF; border: 1px solid #E2E8F0; padding: 32px; }
.contact-form > label { display: flex; flex-direction: column; gap: 8px; font-size: 12px; color: #0F172A; font-weight: 700; text-transform: uppercase; margin: 0; }
.contact-form input { padding: 12px 16px; }
.share-options { grid-column: 1 / -1; display: flex; flex-wrap: wrap; gap: 32px; margin-top: 8px; }
.share-options label { font-size: 13px; color: #475569; font-weight: 600; text-transform: none; flex-direction: row; align-items: center; gap: 8px; }
.contact-form button { grid-column: 1 / -1; justify-self: start; margin-top: 16px; }

.workspace-columns { display: grid; grid-template-columns: 1fr 1fr; gap: 48px; }
.section-head h2 small { font-size: 14px; color: #64748B; font-weight: 600; text-transform: uppercase; margin-left: 12px; }

/* REQUEST CARDS & PROGRESS TRACKER */
.request-card { background: #FFFFFF; border: 1px solid #E2E8F0; padding: 24px; margin-bottom: 16px; transition: box-shadow 0.2s; }
.request-card:hover { box-shadow: 0 8px 16px rgba(15, 23, 42, 0.03); }
.request-card-head { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; margin-bottom: 12px; }
.request-card-head strong { font-size: 16px; font-weight: 800; color: #0F172A; }
.request-status { font-size: 11px; color: #2563EB; font-weight: 700; text-transform: uppercase; border: 1px solid #BFDBFE; background: #EFF6FF; padding: 4px 10px; }
.request-role { font-size: 12px !important; color: #64748B !important; margin: 0 0 12px !important; font-weight: 600; }
.request-card > p:not(.request-role) { font-size: 14px; color: #475569; line-height: 1.6; margin-bottom: 16px; }
.request-actions { display: flex; align-items: center; gap: 16px; margin-top: 16px; }

/* PROGRESS BAR VISUALIZATION */
.progress-tracker { display: flex; align-items: center; justify-content: space-between; margin: 24px 0 16px; position: relative; }
.progress-tracker::before { content: ""; position: absolute; top: 50%; left: 0; right: 0; height: 2px; background: #E2E8F0; z-index: 1; transform: translateY(-50%); }
.progress-step { position: relative; z-index: 2; background: #FFFFFF; padding: 0 8px; font-size: 10px; font-weight: 700; color: #94A3B8; text-transform: uppercase; letter-spacing: 0.05em; display: flex; flex-direction: column; align-items: center; gap: 6px; }
.progress-step::before { content: ""; width: 12px; height: 12px; border-radius: 50%; background: #E2E8F0; border: 2px solid #FFFFFF; box-shadow: 0 0 0 1px #CBD5E1; transition: all 0.3s; }
.progress-step.active { color: #2563EB; }
.progress-step.active::before { background: #2563EB; box-shadow: 0 0 0 2px #BFDBFE; border-color: #FFF; }
.progress-step.completed { color: #0F172A; }
.progress-step.completed::before { background: #0F172A; box-shadow: 0 0 0 1px #0F172A; }

.progress-control { display: flex; align-items: center; gap: 16px; margin-top: 24px; padding-top: 24px; border-top: 1px dashed #E2E8F0; }
.progress-control > span { font-size: 12px; color: #64748B; font-weight: 600; text-transform: uppercase; }
.revoked-label { font-size: 12px; color: #94A3B8; font-weight: 600; font-style: italic; }
.small-button { padding: 10px 16px; font-size: 12px; }

.exchange-box { display: flex; flex-wrap: wrap; gap: 16px; margin-top: 16px; padding: 16px; background: #F8FAFC; color: #0F172A; border: 1px solid #CBD5E1; font-size: 13px; font-weight: 500; }
.exchange-box b { font-weight: 700; color: #475569; margin-right: 4px; text-transform: uppercase; font-size: 11px; letter-spacing: 0.05em; }

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
.application-modal textarea:focus { border-color: #2563EB; background: #FFFFFF; }
.application-consent, .application-works { display: flex; flex-direction: column; gap: 12px; margin: 24px 0; padding: 24px; border: 1px solid #E2E8F0; background: #F8FAFC; }
.application-consent legend, .application-works strong { font-size: 12px; color: #0F172A; font-weight: 800; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 8px; display: block; }
.application-modal .application-consent label, .application-modal .application-works label { display: flex; flex-direction: row; align-items: flex-start; gap: 12px; font-size: 14px; text-transform: none; font-weight: 500; letter-spacing: 0; color: #334155; }
.application-consent input, .application-works input { width: 18px; height: 18px; margin-top: 2px; accent-color: #2563EB; }
.application-consent p { margin: 0; color: #64748B; font-size: 12px; line-height: 1.6; }
.modal-close { position: absolute; right: 24px; top: 24px; border: 0; background: none; color: #94A3B8; font-size: 32px; font-weight: 300; cursor: pointer; transition: color 0.2s; line-height: 1; }
.modal-close:hover { color: #0F172A; }


/* SKILL DOWNLOAD PANEL - STRUCTURAL DESIGN */
.skill-download-panel {
  margin: 40px auto 60px;
  padding: 64px;
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  display: grid;
  grid-template-columns: 1fr 1.5fr;
  gap: 80px;
  align-items: center;
  max-width: 1280px;
}

.skill-panel-text p.eyebrow {
  margin-bottom: 16px;
}

.skill-panel-text h2 {
  margin: 0 0 24px;
  font-size: 42px;
  line-height: 1.15;
  letter-spacing: -0.04em;
  font-weight: 800;
  color: #0F172A;
}

.skill-panel-text p {
  margin: 0;
  color: #475569;
  line-height: 1.7;
  font-size: 16px;
}

.skill-download-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.download-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: #FFFFFF;
  border: 1px solid #E2E8F0;
  text-decoration: none;
  color: #0F172A;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.download-card:hover {
  border-color: #2563EB;
  box-shadow: 0 12px 24px -10px rgba(37, 99, 235, 0.15);
  transform: translateY(-2px);
}

.dl-icon {
  width: 48px;
  height: 48px;
  background: #F1F5F9;
  border: 1px solid #CBD5E1;
  display: grid;
  place-items: center;
  font-family: ui-monospace, monospace;
  font-size: 12px;
  font-weight: 800;
  color: #475569;
  flex-shrink: 0;
}

.download-card:hover .dl-icon {
  background: #2563EB;
  color: #FFFFFF;
  border-color: #2563EB;
}

.dl-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.dl-info strong {
  font-size: 14px;
  font-weight: 700;
  letter-spacing: -0.01em;
}

.dl-info span {
  font-size: 11px;
  color: #64748B;
  font-family: ui-monospace, monospace;
}

.dl-arrow {
  font-size: 20px;
  color: #CBD5E1;
  transition: color 0.2s;
}

.download-card:hover .dl-arrow {
  color: #2563EB;
}

/* RESPONSIVE ADJUSTMENTS */
@media (max-width: 980px) {
  .market-shell { padding: 0 24px 40px; }
  .split-sections, .workspace-columns { grid-template-columns: 1fr; gap: 48px; }
  .work-grid, .builder-grid { grid-template-columns: repeat(2, 1fr); }
  .skill-download-panel { grid-template-columns: 1fr; padding: 40px; gap: 40px; }
  .hero-full-content { padding: 0 24px; grid-template-columns: 1fr; text-align: center; }
  .hero-full-visual-left { display: none; }
  .hero-full-text { padding-right: 0; }
}

@media (max-width: 640px) {
  .market-shell { padding: 0 16px 32px; }
  .market-brand { font-size: 12px; }
  .account-link, .quiet-link { display: none; }
  .market-actions { gap: 12px; }
  .publish-link { padding: 10px 14px; font-size: 11px; }
  .market-links a { font-size: 12px; }
  .page-heading { padding: 48px 0 32px; }
  .page-heading h1 { font-size: 40px; }
  .page-desc { font-size: 16px; }
  .hero-full-content { padding: 0 24px; justify-content: center; text-align: center; }
  .hero-full-text { padding-right: 0; }
  .hero-full-text h2 { font-size: 44px; }
  .hero-full-bg-code { opacity: 0.08; left: -10%; }
  .work-grid, .builder-grid, .form-two-col, .agent-token-form, .contact-form { grid-template-columns: 1fr; }
  .section-head { margin-top: 56px; flex-direction: column; align-items: flex-start; gap: 16px; }
  .project-card { align-items: flex-start; flex-direction: column; padding: 24px; gap: 24px; }
  .project-facts { gap: 12px 16px; }
  .form-panel, .profile-panel, .application-modal { padding: 24px; }
  .form-footer, .agent-form-footer, .agent-skill-pack, .profile-panel { align-items: flex-start; flex-direction: column; }
  .invite-row { flex-direction: column; align-items: stretch; gap: 0; }
  .invite-row button { margin: -8px 0 24px; }
  .agent-token-row { flex-direction: column; align-items: flex-start; gap: 16px; }
  .share-options { grid-column: auto; flex-direction: column; gap: 12px; }
  .market-footer { flex-direction: column; gap: 16px; }
}
</style>
