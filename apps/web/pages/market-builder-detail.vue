<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import marketLogo from "../../../assets/logo/white-bg.svg";
import {
  devMockProfiles,
  getNetworkToken,
  networkRequest,
  type NetworkProfile,
  type NetworkWork
} from "../composables/useBuilderMarket";

const route = useRoute();
const router = useRouter();
const slug = computed(() => typeof route.params.slug === "string" ? route.params.slug : "");

const isAuthed = ref(false);
onMounted(() => {
  isAuthed.value = Boolean(getNetworkToken());
});

const loading = ref(true);
const error = ref("");
const worksError = ref("");
const builder = ref<NetworkProfile | null>(null);
const builderWorks = ref<NetworkWork[]>([]);
const devMockWorks: NetworkWork[] = [
  {
    id: 99101,
    profile_id: 99002,
    slug: "agent-workflow-console",
    title: "Agent Workflow Console",
    summary: "面向多 Agent 任务编排的可观测工作台，支持运行状态追踪与人工介入。",
    body: "",
    cover_url: "",
    demo_url: "",
    repo_url: "",
    tags: ["TypeScript", "Agent", "Workflow"],
    status: "published",
    author_name: "周可为",
    members: [{ profile_id: 99001, display_name: "林沐", role: "产品工程师", contribution: "负责交互原型、工作流编排与前端体验。", sort_order: 1, is_owner: false }]
  },
  {
    id: 99102,
    profile_id: 99002,
    slug: "rag-evaluation-kit",
    title: "RAG Evaluation Kit",
    summary: "轻量化检索增强生成评测工具，帮助团队比较检索策略与回答质量。",
    body: "",
    cover_url: "",
    demo_url: "",
    repo_url: "",
    tags: ["Python", "RAG", "Evaluation"],
    status: "published",
    author_name: "周可为",
    members: [{ profile_id: 99001, display_name: "林沐", role: "AI 工程师", contribution: "搭建评测流程，并设计可复用的检索实验模板。", sort_order: 1, is_owner: false }]
  },
  {
    id: 99103,
    profile_id: 99002,
    slug: "prompt-trace-dashboard",
    title: "Prompt Trace Dashboard",
    summary: "把提示词版本、模型调用和用户反馈放进同一条可回放链路。",
    body: "",
    cover_url: "",
    demo_url: "",
    repo_url: "",
    tags: ["Vue", "LLM Ops", "Analytics"],
    status: "published",
    author_name: "周可为",
    members: [{ profile_id: 99001, display_name: "林沐", role: "全栈工程师", contribution: "实现调用链路回放、版本对比和反馈标注。", sort_order: 1, is_owner: false }]
  },
  {
    id: 99104,
    profile_id: 99002,
    slug: "builder-network-starter",
    title: "Builder Network Starter",
    summary: "用于早期协作社区的个人资料与项目发现体验原型。",
    body: "",
    cover_url: "",
    demo_url: "",
    repo_url: "",
    tags: ["Vue 3", "Go", "Product"],
    status: "published",
    author_name: "周可为",
    members: [{ profile_id: 99001, display_name: "林沐", role: "产品工程师", contribution: "从用户流程到 API 联调，完成首版闭环验证。", sort_order: 1, is_owner: false }]
  }
];

async function load() {
  loading.value = true;
  error.value = "";
  worksError.value = "";
  builderWorks.value = [];
  try {
    let hit: NetworkProfile | undefined;
    try {
      hit = await networkRequest<NetworkProfile>(
        `/api/v1/public/network/profiles/${encodeURIComponent(slug.value)}`
      );
    } catch {
      try {
        const profiles = await networkRequest<NetworkProfile[]>("/api/v1/public/network/profiles");
        hit = profiles.find((profile) => profile.slug === slug.value);
      } catch (listCause) {
        if (!import.meta.env.DEV) throw listCause;
      }
    }
    if (!hit && import.meta.env.DEV) hit = devMockProfiles.find((profile) => profile.slug === slug.value);
    if (!hit) {
      await router.replace("/market/builders");
      return;
    }
    builder.value = hit;
    document.title = `BuilderNetwork | ${hit.display_name}`;
    if (hit.is_demo) {
      if (hit.id === 99001) builderWorks.value = devMockWorks;
      return;
    }

    try {
      const works = await networkRequest<NetworkWork[]>("/api/v1/public/network/works");
      builderWorks.value = works.filter((work) =>
        work.profile_id === hit.id || work.members?.some((member) => member.profile_id === hit.id)
      );
    } catch (cause) {
      worksError.value = cause instanceof Error ? cause.message : "作品暂时无法加载";
    }
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : "加载 Builder 详情失败";
  } finally {
    loading.value = false;
  }
}

watch(slug, () => {
  void load();
}, { immediate: true });

const availabilityLabel = computed(() => {
  if (!builder.value) return "";
  if (builder.value.availability === "unavailable") return "暂不接单";
  if (builder.value.availability === "limited") return "有限档期";
  return "开放合作";
});

const availabilityDescription = computed(() => {
  if (builder.value?.availability === "unavailable") return "目前暂不接收新的合作申请";
  if (builder.value?.availability === "limited") return "档期有限，欢迎先介绍合作内容";
  return "目前开放新的合作沟通";
});

function roleInWork(work: NetworkWork) {
  const member = work.members?.find((item) => item.profile_id === builder.value?.id);
  if (member?.is_owner || work.profile_id === builder.value?.id) return "项目发起人";
  return member?.role || "项目成员";
}

function contributionInWork(work: NetworkWork) {
  const member = work.members?.find((item) => item.profile_id === builder.value?.id);
  return member?.contribution || "";
}

function externalLink(value: string) {
  try {
    const url = new URL(value);
    return url.protocol === "http:" || url.protocol === "https:" ? url.href : "";
  } catch {
    return "";
  }
}

const goApply = async () => {
  if (!builder.value || builder.value.is_demo) return;
  await router.push({ path: "/market/builders", query: { apply_profile: String(builder.value.id) } });
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
      <div v-if="loading" class="state">正在加载 Builder 详情…</div>
      <div v-else-if="error" class="state error">{{ error }}</div>
      <div v-else-if="builder" class="detail">
        <div class="breadcrumb">
          <RouterLink to="/market/builders">全部 Builder</RouterLink>
          <span>/</span>
          <span>{{ builder.display_name }}</span>
        </div>

        <div class="profile-layout">
          <aside class="profile-sidebar">
            <div class="profile-avatar">{{ builder.display_name.slice(0, 1) }}</div>
            <h1>{{ builder.display_name }}</h1>
            <p class="profile-handle">{{ builder.slug }}</p>
            <p class="profile-role">{{ builder.primary_role }}</p>
            <button v-if="!builder.is_demo" class="button-primary profile-action" type="button" @click="goApply">发起合作申请</button>
            <p v-else class="demo-profile-note">本地演示档案，仅用于预览。</p>
            <p class="profile-bio">{{ builder.bio || builder.headline || "该 Builder 暂未填写详细介绍。" }}</p>
            <div class="profile-facts">
              <span><i aria-hidden="true">⌖</i>{{ builder.city || "远程协作" }}</span>
              <span><i aria-hidden="true">◷</i>{{ availabilityLabel }}<template v-if="builder.available_from"> · {{ builder.available_from }}</template></span>
              <span v-if="builder.collaboration_modes?.length"><i aria-hidden="true">⌘</i>{{ builder.collaboration_modes.join(" · ") }}</span>
            </div>
          </aside>

          <div class="profile-content">
              <section class="repository-section">
                <div class="section-heading">
                  <div>
                    <p class="eyebrow">PUBLIC WORK</p>
                    <h2>公开作品</h2>
                  </div>
                  <span class="section-count">{{ builderWorks.length }} 个作品</span>
                </div>
                <p v-if="worksError" class="section-message">{{ worksError }}</p>
                <div v-else-if="builderWorks.length" class="repository-grid">
                  <article v-for="work in builderWorks" :key="work.id" class="repository-card">
                    <div class="repository-title-row">
                      <a v-if="externalLink(work.repo_url)" :href="externalLink(work.repo_url)" target="_blank" rel="noopener noreferrer">{{ work.title }} ↗</a>
                      <strong v-else>{{ work.title }}</strong>
                      <span>公开</span>
                    </div>
                    <p v-if="work.summary" class="repository-summary">{{ work.summary }}</p>
                    <p v-else class="repository-summary">尚未填写作品介绍。</p>
                    <p v-if="contributionInWork(work)" class="repository-contribution">
                      <strong>{{ roleInWork(work) }}</strong><span>{{ contributionInWork(work) }}</span>
                    </p>
                    <div class="repository-footer">
                      <div v-if="work.tags?.length" class="repository-tags">
                        <span v-for="tag in work.tags.slice(0, 3)" :key="tag"><i></i>{{ tag }}</span>
                      </div>
                      <a v-if="externalLink(work.demo_url)" :href="externalLink(work.demo_url)" target="_blank" rel="noopener noreferrer">Demo ↗</a>
                    </div>
                  </article>
                </div>
                <div v-else class="empty-portfolio">
                  <strong>暂未发布公开作品</strong>
                  <p>这个 Builder 的公开作品发布后会展示在这里。</p>
                </div>
              </section>

            <section class="skills-section">
              <div class="section-heading">
                <div><p class="eyebrow">SKILLS & COLLABORATION</p><h2>技能与合作</h2></div>
              </div>
              <div class="skills-block">
                <h3>技术栈与能力</h3>
                <div v-if="builder.skills?.length" class="tags">
                  <span v-for="skill in builder.skills" :key="skill">{{ skill }}</span>
                </div>
                <p v-else class="empty">暂未填写公开技术栈</p>
              </div>
              <div class="skills-block">
                <h3>合作方式</h3>
                <div v-if="builder.collaboration_modes?.length" class="tags">
                  <span v-for="mode in builder.collaboration_modes" :key="mode">{{ mode }}</span>
                </div>
                <p v-else class="empty">暂无合作方式说明</p>
              </div>
              <div class="skills-block">
                <h3>个人介绍</h3>
                <p class="profile-bio-full">{{ builder.bio || builder.headline || "该 Builder 暂未填写详细介绍。" }}</p>
              </div>
            </section>

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
  max-width: 1540px;
  margin: 0 auto;
  padding: 28px max(40px, calc((100vw - 1540px) / 2)) 80px;
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

.avatar {
  display: flex;
  width: 60px;
  height: 60px;
  flex: 0 0 60px;
  align-items: center;
  justify-content: center;
  border-radius: 18px;
  background: #0F172A;
  color: #FFFFFF;
  font-weight: 900;
  font-size: 24px;
}

.profile-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(270px, 340px);
  gap: clamp(24px, 5vw, 72px);
  align-items: center;
  padding: clamp(28px, 5vw, 64px);
  border: 1px solid #E2E8F0;
  border-top: 4px solid #0BCB78;
  background: linear-gradient(120deg, #FFFFFF 0%, #F8FAFC 100%);
}

.profile-identity {
  display: flex;
  align-items: flex-start;
  gap: 22px;
  min-width: 0;
}

.profile-identity .avatar {
  width: 76px;
  height: 76px;
  flex-basis: 76px;
  border-radius: 20px;
  font-size: 30px;
}

.identity-copy { min-width: 0; }
.identity-copy .eyebrow { margin: 0 0 12px; }
.identity-copy h1 {
  margin: 0;
  color: #0F172A;
  font-size: clamp(36px, 5vw, 58px);
  line-height: 1.05;
  letter-spacing: -.055em;
}
.identity-role {
  margin: 10px 0 0;
  color: #334155;
  font-size: 15px;
  font-weight: 700;
}
.identity-role span { padding: 0 4px; color: #0BCB78; }
.identity-headline {
  max-width: 680px;
  margin: 18px 0 0;
  color: #475569;
  font-size: 17px;
  line-height: 1.7;
}

.badges {
  margin-top: 18px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.badge {
  font-size: 11px;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  padding: 8px 10px;
  border-radius: 999px;
  background: rgba(11, 203, 120, 0.12);
  border: 1px solid rgba(11, 203, 120, 0.26);
  color: #065F46;
}

.badge.subtle {
  background: #F1F5F9;
  border-color: #E2E8F0;
  color: #334155;
}

.contact-card {
  padding: 24px;
  border: 1px solid #DCE5E0;
  background: #F8FBF9;
}
.contact-kicker {
  margin: 0 0 12px;
  color: #0BCB78;
  font: 700 10px ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: .12em;
}
.contact-card > strong { color: #0F172A; font-size: 19px; }
.contact-card > p:not(.contact-kicker) {
  margin: 8px 0 20px;
  color: #64748B;
  font-size: 13px;
  line-height: 1.6;
}
.contact-card .button-primary {
  display: flex;
  width: 100%;
  min-height: 48px;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
}
.contact-card .button-primary span { font-size: 18px; }
.contact-card small {
  display: block;
  margin-top: 14px;
  color: #64748B;
  font-size: 10px;
  line-height: 1.6;
}
.contact-card .demo-profile-note { margin-bottom: 0; color: #64748B; font-size: 12px; }

.builder-layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(270px, 330px);
  gap: clamp(24px, 4vw, 56px);
  align-items: start;
  margin-top: 24px;
}
.builder-main { min-width: 0; }
.portfolio-section,
.about-panel { padding: 28px 0; }
.section-heading {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 18px;
  margin-bottom: 22px;
  padding-bottom: 16px;
  border-bottom: 2px solid #0F172A;
}
.section-heading .eyebrow { margin: 0 0 8px; }
.section-heading h2,
.about-panel h2 {
  margin: 0;
  color: #0F172A;
  font-size: clamp(25px, 3vw, 34px);
  letter-spacing: -.04em;
}
.section-count {
  flex: none;
  color: #64748B;
  font-size: 11px;
  font-weight: 700;
}
.portfolio-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}
.portfolio-card {
  display: flex;
  min-width: 0;
  flex-direction: column;
  border: 1px solid #E2E8F0;
  background: #FFFFFF;
  transition: transform .2s, box-shadow .2s;
}
.portfolio-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 24px rgba(15, 23, 42, .06);
}
.portfolio-cover {
  position: relative;
  aspect-ratio: 16 / 9;
  overflow: hidden;
  background: #0F172A;
}
.portfolio-cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.portfolio-cover-empty {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 18px;
  box-sizing: border-box;
  background:
    radial-gradient(circle at 80% 20%, rgba(11, 203, 120, .25), transparent 35%),
    linear-gradient(135deg, #0F172A, #1E293B);
}
.portfolio-cover-empty span {
  color: #86EFAC;
  font: 10px ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: .12em;
}
.portfolio-cover-empty strong {
  color: #FFFFFF;
  font-size: 20px;
  letter-spacing: -.03em;
}
.portfolio-card-body {
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: flex-start;
  padding: 20px;
}
.portfolio-role {
  margin-bottom: 8px;
  color: #0BCB78;
  font-size: 10px;
  font-weight: 800;
  letter-spacing: .08em;
  text-transform: uppercase;
}
.portfolio-card h3 {
  margin: 0;
  color: #0F172A;
  font-size: 20px;
  line-height: 1.35;
  letter-spacing: -.025em;
}
.portfolio-summary,
.portfolio-contribution {
  color: #475569;
  font-size: 13px;
  line-height: 1.7;
}
.portfolio-summary { margin: 10px 0 0; }
.portfolio-contribution {
  display: grid;
  gap: 4px;
  margin: 16px 0 0;
  padding-top: 14px;
  border-top: 1px solid #E2E8F0;
}
.portfolio-contribution strong {
  color: #64748B;
  font-size: 10px;
  letter-spacing: .08em;
  text-transform: uppercase;
}
.portfolio-details { margin-top: 14px; color: #475569; font-size: 12px; line-height: 1.7; }
.portfolio-details summary { color: #0F172A; font-weight: 700; cursor: pointer; }
.portfolio-details p { margin: 8px 0 0; white-space: pre-wrap; }
.portfolio-card .tags { margin-top: 16px; }
.portfolio-card .tags span {
  padding: 5px 8px;
  border-radius: 2px;
  background: #F8FAFC;
  font-size: 9px;
}
.portfolio-links {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  margin-top: auto;
  padding-top: 18px;
}
.portfolio-links a {
  color: #0F172A;
  font-size: 11px;
  font-weight: 800;
  text-decoration: none;
}
.portfolio-links a:hover { color: #08B96C; }
.empty-portfolio {
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 28px;
  border: 1px dashed #CBD5E1;
  color: #475569;
  background: #F8FAFC;
}
.empty-mark {
  display: grid;
  width: 44px;
  height: 44px;
  flex: none;
  place-items: center;
  border: 1px solid #DCE5E0;
  color: #0BCB78;
  font: 700 16px ui-monospace, SFMono-Regular, Menlo, monospace;
}
.empty-portfolio strong { color: #0F172A; font-size: 14px; }
.empty-portfolio p { margin: 6px 0 0; font-size: 12px; line-height: 1.6; }
.section-message { color: #92400E; font-size: 13px; }
.about-panel {
  margin-top: 14px;
  padding-top: 28px;
  border-top: 1px solid #E2E8F0;
}
.about-panel .eyebrow { margin-bottom: 8px; }
.body {
  margin: 14px 0 0;
  color: #475569;
  font-size: 14px;
  line-height: 1.9;
  white-space: pre-wrap;
}

.builder-sidebar {
  position: sticky;
  top: 104px;
  display: grid;
  gap: 14px;
}
.sidebar-panel {
  padding: 22px;
  border: 1px solid #E2E8F0;
  background: #FFFFFF;
}
.sidebar-panel .eyebrow { margin-bottom: 8px; }
.sidebar-panel h2 {
  margin: 0 0 16px;
  color: #0F172A;
  font-size: 18px;
  letter-spacing: -.025em;
}
.sidebar-note {
  margin: 16px 0 0;
  color: #64748B;
  font-size: 11px;
  line-height: 1.6;
}
.availability-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid #E2E8F0;
  font-size: 11px;
}
.availability-row span { color: #64748B; }
.availability-row strong { color: #0F172A; text-align: right; }

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

@media (max-width: 920px) {
  .market-nav-full { padding: 0 24px; }
  .nav-primary { gap: 24px; }
  .market-links { gap: 20px; }
  .profile-hero {
    grid-template-columns: 1fr;
  }
  .builder-layout {
    grid-template-columns: 1fr;
  }
  .builder-sidebar {
    position: static;
    grid-template-columns: repeat(2, minmax(0, 1fr));
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
  .profile-hero { padding: 22px 18px; gap: 22px; }
  .profile-identity { gap: 14px; }
  .profile-identity .avatar { width: 54px; height: 54px; flex-basis: 54px; border-radius: 15px; font-size: 22px; }
  .identity-copy .eyebrow { margin-bottom: 8px; font-size: 9px; }
  .identity-copy h1 { font-size: 34px; }
  .identity-role { font-size: 12px; }
  .identity-headline { margin-top: 12px; font-size: 14px; }
  .badge { padding: 6px 8px; font-size: 9px; }
  .contact-card { padding: 18px; }
  .builder-layout { gap: 8px; margin-top: 10px; }
  .portfolio-section, .about-panel { padding: 20px 0; }
  .portfolio-grid, .builder-sidebar { grid-template-columns: 1fr; }
  .section-heading { align-items: flex-start; }
  .section-count { padding-top: 6px; }
  .portfolio-card-body { padding: 16px; }
}

.market-nav-full { padding-right: 40px; padding-left: 40px; }
.nav-inner { max-width: 1800px; margin: 0 auto; }
.main { max-width: 1800px; padding: 24px 40px 80px; }
.detail { gap: 18px; }
.breadcrumb { min-height: 22px; }
.profile-layout {
  display: grid;
  grid-template-columns: 260px minmax(0, 1fr);
  align-items: start;
  gap: 36px;
}
.profile-sidebar { min-width: 0; padding-top: 10px; }
.profile-avatar {
  display: grid;
  width: 220px;
  aspect-ratio: 1;
  place-items: center;
  border: 1px solid #DCE5E0;
  border-radius: 50%;
  background:
    radial-gradient(circle at 28% 22%, rgba(11, 203, 120, .18), transparent 35%),
    linear-gradient(145deg, #F8FAFC, #E8F7EF);
  color: #087A49;
  font-size: 96px;
  font-weight: 700;
  letter-spacing: -.08em;
}
.profile-sidebar h1 {
  margin: 20px 0 0;
  color: #0F172A;
  font-size: 28px;
  line-height: 1.2;
  letter-spacing: -.04em;
}
.profile-handle { margin: 3px 0 14px; color: #64748B; font-size: 18px; }
.profile-role { margin: 0 0 14px; color: #334155; font-size: 14px; font-weight: 600; }
.profile-action {
  width: 100%;
  min-height: 38px;
  border: 1px solid #CBD5E1;
  border-radius: 6px;
  background: #F8FAFC;
  color: #0F172A;
  font: inherit;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}
.profile-action:hover { border-color: #94A3B8; background: #F1F5F9; }
.profile-sidebar .demo-profile-note {
  margin: 0;
  padding: 10px 12px;
  border: 1px solid #E2E8F0;
  border-radius: 6px;
  color: #64748B;
  font-size: 11px;
  line-height: 1.5;
}
.profile-bio {
  margin: 16px 0;
  color: #334155;
  font-size: 14px;
  line-height: 1.55;
  white-space: pre-wrap;
}
.profile-facts {
  display: grid;
  gap: 10px;
  padding: 14px 0;
  border-top: 1px solid #E2E8F0;
  border-bottom: 1px solid #E2E8F0;
}
.profile-facts span { display: flex; align-items: flex-start; gap: 9px; color: #475569; font-size: 12px; line-height: 1.5; }
.profile-facts i { width: 14px; flex: none; color: #64748B; font-size: 14px; font-style: normal; text-align: center; }
.profile-content { min-width: 0; }
.repository-section { padding-top: 18px; }
.section-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}
.section-heading .eyebrow { margin: 0 0 5px; font-size: 9px; }
.section-heading h2,
.skills-section h2 {
  margin: 0;
  color: #0F172A;
  font-size: 18px;
  font-weight: 600;
  letter-spacing: -.02em;
}
.section-count { color: #64748B; font-size: 11px; }
.repository-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; }
.repository-card {
  display: flex;
  min-width: 0;
  min-height: 144px;
  flex-direction: column;
  padding: 16px;
  border: 1px solid #D8DEE4;
  border-radius: 6px;
  background: #FFFFFF;
}
.repository-title-row { display: flex; min-width: 0; align-items: center; gap: 12px; }
.repository-title-row a,
.repository-title-row strong {
  overflow: hidden;
  color: #087A49;
  font-size: 14px;
  font-weight: 700;
  text-overflow: ellipsis;
  text-decoration: none;
  white-space: nowrap;
}
.repository-title-row a:hover { text-decoration: underline; }
.repository-title-row > span {
  flex: none;
  margin-left: auto;
  padding: 2px 8px;
  border: 1px solid #D8DEE4;
  border-radius: 999px;
  color: #64748B;
  font-size: 10px;
}
.repository-summary {
  display: -webkit-box;
  overflow: hidden;
  min-height: 36px;
  margin: 8px 0 14px;
  color: #64748B;
  font-size: 11px;
  line-height: 1.6;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}
.repository-contribution { display: grid; gap: 3px; margin: -3px 0 12px; color: #64748B; font-size: 10px; line-height: 1.5; }
.repository-contribution strong { color: #334155; font-size: 10px; }
.repository-footer { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-top: auto; }
.repository-tags { display: flex; min-width: 0; flex-wrap: wrap; gap: 10px; }
.repository-tags span { display: inline-flex; align-items: center; gap: 5px; color: #64748B; font-size: 10px; }
.repository-tags i { width: 9px; height: 9px; border-radius: 50%; background: #0BCB78; }
.repository-footer > a { flex: none; color: #087A49; font-size: 10px; font-weight: 700; text-decoration: none; }
.repository-footer > a:hover { text-decoration: underline; }
.empty-portfolio {
  padding: 28px 18px;
  border: 1px dashed #CBD5E1;
  border-radius: 6px;
  background: #F8FAFC;
  color: #475569;
}
.empty-portfolio strong { color: #334155; font-size: 13px; }
.empty-portfolio p { margin: 6px 0 0; color: #64748B; font-size: 11px; }
.skills-section { padding-top: 18px; }
.skills-section > .section-heading { padding-bottom: 14px; border-bottom: 1px solid #E2E8F0; }
.skills-block { padding: 20px 0; border-bottom: 1px solid #E2E8F0; }
.skills-block h3 { margin: 0 0 12px; color: #334155; font-size: 13px; }
.skills-block .tags { gap: 8px; }
.skills-block .tags span { border-radius: 999px; background: #F8FAFC; font-size: 10px; text-transform: none; }
.profile-bio-full { margin: 0; color: #475569; font-size: 13px; line-height: 1.8; white-space: pre-wrap; }

@media (max-width: 920px) {
  .profile-layout { grid-template-columns: 220px minmax(0, 1fr); gap: 24px; }
  .profile-avatar { width: 180px; font-size: 78px; }
}

@media (max-width: 700px) {
  .market-nav-full { padding-right: 20px; padding-left: 20px; }
  .main { padding: 20px 20px 56px; }
  .profile-layout { grid-template-columns: 1fr; gap: 12px; }
  .profile-sidebar { display: grid; grid-template-columns: 76px minmax(0, 1fr); align-items: center; column-gap: 16px; padding: 0 0 18px; border-bottom: 1px solid #E2E8F0; }
  .profile-avatar { grid-row: span 3; width: 76px; font-size: 34px; }
  .profile-sidebar h1 { margin: 0; font-size: 24px; }
  .profile-handle { margin: 2px 0 0; font-size: 14px; }
  .profile-role { margin: 2px 0 0; font-size: 12px; }
  .profile-action, .profile-sidebar .demo-profile-note { grid-column: 1 / -1; margin-top: 14px; }
  .profile-bio, .profile-facts { grid-column: 1 / -1; }
  .profile-bio { margin: 14px 0; }
  .repository-grid { grid-template-columns: 1fr; }
}

@media (max-width: 420px) {
  .market-nav-full { padding-right: 14px; padding-left: 14px; }
  .main { padding-right: 16px; padding-left: 16px; }
}
</style>
