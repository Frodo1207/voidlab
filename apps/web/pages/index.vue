<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { useInsightsFeed } from "../composables/useInsightsFeed";
import { useSiteConfigs } from "../composables/useSiteConfigs";
import {
  eventStatusBadgeClass,
  eventStatusLabel,
  useEventArchive
} from "../composables/useEventArchive";
import SiteHeader from "../components/SiteHeader.vue";

const { featuredEvents, loadEvents } = useEventArchive();
const { featuredInsights, loadInsights } = useInsightsFeed();
const { homeBanner, homeFeatured, footerConfig, featuredContentSlots, loadSiteConfigs } = useSiteConfigs();
const router = useRouter();

const typedTitle = ref("");
const showSubtitle = ref(false);
const showCtas = ref(false);
const showStats = ref(false);
const titleDone = computed(() => typedTitle.value === homeBanner.value.titleText);
const homepageEvents = computed(() => featuredEvents.value.slice(0, featuredContentSlots.value.eventsLimit));
const homepageInsights = computed(() => featuredInsights.value.slice(0, featuredContentSlots.value.insightsLimit));

let typeTimer: number | null = null;
let startTypeTimer: number | null = null;

// 移除原有的复杂滚动计算，因为 footer 现在是正常流式布局
// function updateFooterSpacer() { ... }
// function handleScroll() { ... }

function runTypeWriter(index = 0) {
  const titleText = homeBanner.value.titleText;

  if (index < titleText.length) {
    typedTitle.value = titleText.slice(0, index + 1);
    const delay = Math.random() * 150 + 100;
    typeTimer = window.setTimeout(() => runTypeWriter(index + 1), delay);
    return;
  }

  typedTitle.value = titleText;
  showSubtitle.value = true;
  window.setTimeout(() => {
    showCtas.value = true;
  }, 400);
  window.setTimeout(() => {
    showStats.value = true;
  }, 800);
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function routeHref(path: string) {
  return router.resolve(path).href;
}

onMounted(async () => {
  document.title = "VOID LAB | 探索 AI 的边界";
  void loadEvents();
  void loadInsights();
  try {
    await loadSiteConfigs();
  } catch {
    // Keep built-in defaults when public configs are unavailable.
  }
  startTypeTimer = window.setTimeout(() => runTypeWriter(), 600);
});

onBeforeUnmount(() => {
  if (typeTimer) window.clearTimeout(typeTimer);
  if (startTypeTimer) window.clearTimeout(startTypeTimer);
});
</script>

<template>
  <div>
    <SiteHeader theme="dark" activePath="/" />
    
    <div class="content-layer antialiased selection:bg-white selection:text-black pt-16">
      <section class="min-h-[calc(100vh-4rem)] flex flex-col justify-between p-6 md:p-12">
        
        <main class="flex-1 flex flex-col items-center md:items-end justify-center pointer-events-none text-center md:text-right mt-20 md:mt-0 relative z-10">
          <div class="bg-white text-black px-5 py-3 mb-6 inline-block pointer-events-auto">
            <h1 class="pixel-zh-title text-4xl md:text-6xl tracking-widest leading-none m-0" style="min-height: 1em;">
              {{ typedTitle }}
              <span v-if="!titleDone" class="typing-cursor-black"></span>
            </h1>
          </div>

          <div class="pixel-text text-xl md:text-2xl text-white/70 tracking-[0.2em] transition-opacity duration-1000 mt-2" :class="showSubtitle ? 'opacity-100' : 'opacity-0'">
            > {{ homeBanner.subtitle }}<span class="typing-cursor"></span>
          </div>

          <div class="flex flex-col md:flex-row gap-4 mt-12 transition-opacity duration-1000 pointer-events-auto" :class="showCtas ? 'opacity-100' : 'opacity-0'">
            <a :href="routeHref(homeBanner.primaryCtaPath)" target="_blank" rel="noreferrer" class="bg-[var(--color-turquoise)] text-black px-8 py-3 text-sm tracking-[0.2em] hover:bg-white hover:text-black transition-all font-bold pixel-zh-title pointer-events-auto">
              > {{ homeBanner.primaryCtaLabel }}
            </a>
            <a :href="routeHref(homeBanner.secondaryCtaPath)" target="_blank" rel="noreferrer" class="border border-white/30 text-white/70 px-8 py-3 text-sm tracking-[0.2em] hover:border-white hover:text-white transition-all pixel-zh-title">
              {{ homeBanner.secondaryCtaLabel }}
            </a>
          </div>

          <div class="mt-12 md:mt-16 bg-black/60 border border-[var(--color-turquoise)]/30 p-6 md:p-8 backdrop-blur-md transition-opacity duration-1000 hidden md:block relative w-full max-w-lg" :class="showStats ? 'opacity-100' : 'opacity-0'">
            <div class="absolute -top-3 left-6 bg-[var(--color-black)] px-2 text-[var(--color-turquoise)] text-xs tracking-widest pixel-text border border-[var(--color-turquoise)]/30">
              > {{ homeBanner.statusLabel }}
            </div>
            <div class="flex justify-between items-center mt-2">
              <div class="text-center md:text-left">
                <div class="text-[var(--color-turquoise)] text-5xl md:text-6xl pixel-text font-bold mb-2">{{ homeFeatured.communityCount }}<span class="text-3xl text-[var(--color-turquoise)]/60">{{ homeFeatured.communityCountSuffix }}</span></div>
                <div class="text-white/70 text-sm md:text-base tracking-widest pixel-zh-title">社区成员</div>
              </div>
              <div class="w-px h-16 bg-white/20"></div>
              <div class="text-center md:text-left">
                <div class="text-[var(--color-turquoise)] text-5xl md:text-6xl pixel-text font-bold mb-2">{{ homeFeatured.eventCount }}<span class="text-3xl text-[var(--color-turquoise)]/60">{{ homeFeatured.eventCountSuffix }}</span></div>
                <div class="text-white/70 text-sm md:text-base tracking-widest pixel-zh-title">举办活动</div>
              </div>
            </div>
          </div>
        </main>

        <footer class="w-full flex justify-between items-end mt-20 md:mt-0">
          <div class="text-white/40 text-xs font-light tracking-widest hidden md:block">
            <p>系统启动 [完成]</p>
          </div>
          <div class="flex flex-col items-center gap-2 cursor-pointer group pointer-events-auto mx-auto md:mx-0">
            <div class="pixel-text text-white/50 group-hover:text-[var(--color-turquoise)] transition-colors text-xl animate-bounce">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" class="stroke-current">
                <path d="M12 4V20M12 20L6 14M12 20L18 14" stroke-width="1" stroke-linecap="square" stroke-linejoin="miter" />
              </svg>
            </div>
            <span class="text-xs text-white/30 tracking-[0.2em] group-hover:text-white/80 transition-colors">向下浏览官网内容</span>
          </div>
        </footer>
      </section>

      <div class="w-full overflow-hidden border-y border-white/10 py-2 mt-20 relative z-10 bg-black/40 backdrop-blur-md pointer-events-auto">
        <div class="whitespace-nowrap flex w-max animate-marquee pixel-text text-xs tracking-widest text-[var(--color-turquoise)]/60">
          <span class="mx-6">> VOIDLAB 系统 . 在线</span>
          <span class="mx-6">智能体身份 . 已验证</span>
          <span class="mx-6">系统审计 . 已就绪</span>
          <span class="mx-6">工作流 . 已部署</span>
          <span class="mx-6">资产 . 已生成</span>
          <span class="mx-6">> VOIDLAB 系统 . 在线</span>
          <span class="mx-6">智能体身份 . 已验证</span>
          <span class="mx-6">系统审计 . 已就绪</span>
          <span class="mx-6">工作流 . 已部署</span>
          <span class="mx-6">资产 . 已生成</span>
        </div>
      </div>

      <section id="offerings" class="relative z-10 w-full border-t border-white/5 pointer-events-auto overflow-hidden bg-black/60 backdrop-blur-xl">
        <div class="absolute inset-0 bg-gradient-to-b from-[var(--color-turquoise)]/8 via-transparent to-transparent pointer-events-none"></div>
        <div class="relative z-10 w-full px-6 md:px-12 lg:px-16 py-20 md:py-24 mx-auto max-w-[1920px]">
          <div class="max-w-3xl mb-14">
            <div class="pixel-text text-xs md:text-sm text-[var(--color-turquoise)]/80 mb-4">> 我们提供什么</div>
            <h2 class="text-4xl md:text-5xl font-bold pixel-zh-title leading-tight text-white tracking-wide">
              服务介绍
            </h2>
            <p class="text-white/55 text-sm md:text-base leading-relaxed mt-5">
              从 AI 课程到活动落地，VOID LAB 提供 AI 商业化一条龙服务。
            </p>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mx-auto max-w-[1920px]">
            <!-- AI 课程 -->
            <div class="group relative bg-black/40 rounded-xl overflow-hidden flex flex-col h-[280px] md:h-[320px] p-6 md:p-8">
              <!-- 背景动效 -->
              <div class="absolute inset-0 pointer-events-none opacity-60 group-hover:opacity-100 transition-opacity duration-700">
                <div class="absolute -right-20 -bottom-20 w-[400px] h-[400px] mix-blend-screen animate-[spin_40s_linear_infinite]">
                   <div class="absolute top-10 right-10 w-56 h-56 bg-gradient-to-tr from-[#ff9f0a] to-[#bf5af2] rounded-full filter blur-[50px] animate-blob-morph opacity-40"></div>
                   <div class="absolute bottom-10 left-10 w-64 h-64 bg-gradient-to-bl from-[#ff375f] to-transparent rounded-full filter blur-[60px] animate-blob-morph animation-delay-2000 opacity-30"></div>
                </div>
              </div>
              <!-- 噪点纹理 -->
              <div class="absolute inset-0 bg-[url('data:image/svg+xml,%3Csvg viewBox=%220 0 200 200%22 xmlns=%22http://www.w3.org/2000/svg%22%3E%3Cfilter id=%22noise%22%3E%3CfeTurbulence type=%22fractalNoise%22 baseFrequency=%220.8%22 numOctaves=%224%22 stitchTiles=%22stitch%22/%3E%3C/filter%3E%3Crect width=%22100%25%22 height=%22100%25%22 filter=%22url(%23noise)%22/%3E%3C/svg%3E')] opacity-[0.15] mix-blend-overlay pointer-events-none"></div>

              <!-- 内容 -->
              <div class="relative z-10 flex flex-col h-full">
                <h3 class="text-2xl font-bold text-white mb-4 tracking-wide pixel-zh-title">AI 课程</h3>
                <ul class="text-xs md:text-sm text-white/90 leading-[1.8] space-y-2 max-w-sm">
                  <li class="flex items-start gap-2">
                    <span class="text-[var(--color-turquoise)] font-bold">></span>
                    <span>职场转行 AI 破局：重塑个人核心竞争力。</span>
                  </li>
                  <li class="flex items-start gap-2">
                    <span class="text-[var(--color-turquoise)] font-bold">></span>
                    <span>大学生启航计划：从 0 到 1 点亮 AI 技能树。</span>
                  </li>
                  <li class="flex items-start gap-2">
                    <span class="text-[var(--color-turquoise)] font-bold">></span>
                    <span>真实商业项目驱动：手摸手带做实战 Agent。</span>
                  </li>
                </ul>
                
                <div class="mt-auto pixel-text text-[9px] md:text-[10px] text-white leading-[1.4] whitespace-pre font-mono">
<span class="text-white">> edu.init_course</span>
  tracks   [career_shift, student, projects]
  status   enrolling
<span class="text-[var(--color-turquoise)]">✓</span> syllabus loaded
                </div>
              </div>
            </div>

            <!-- AI 咨询 -->
            <div class="group relative bg-black/40 rounded-xl overflow-hidden flex flex-col h-[280px] md:h-[320px] p-6 md:p-8">
              <div class="absolute inset-0 pointer-events-none opacity-60 group-hover:opacity-100 transition-opacity duration-700">
                <div class="absolute -right-10 top-10 w-[400px] h-[400px] mix-blend-screen animate-[spin_35s_linear_infinite_reverse]">
                   <div class="absolute top-0 right-20 w-64 h-64 bg-gradient-to-tr from-[#0a84ff] to-[#5e5ce6] rounded-full filter blur-[50px] animate-blob-morph opacity-40"></div>
                   <div class="absolute bottom-20 left-0 w-48 h-48 bg-gradient-to-bl from-[#32ade6] to-transparent rounded-full filter blur-[60px] animate-blob-morph animation-delay-4000 opacity-30"></div>
                </div>
              </div>
              <div class="absolute inset-0 bg-[url('data:image/svg+xml,%3Csvg viewBox=%220 0 200 200%22 xmlns=%22http://www.w3.org/2000/svg%22%3E%3Cfilter id=%22noise%22%3E%3CfeTurbulence type=%22fractalNoise%22 baseFrequency=%220.8%22 numOctaves=%224%22 stitchTiles=%22stitch%22/%3E%3C/filter%3E%3Crect width=%22100%25%22 height=%22100%25%22 filter=%22url(%23noise)%22/%3E%3C/svg%3E')] opacity-[0.15] mix-blend-overlay pointer-events-none"></div>

              <div class="relative z-10 flex flex-col h-full">
                <h3 class="text-2xl font-bold text-white mb-4 tracking-wide pixel-zh-title">AI 咨询</h3>
                <ul class="text-xs md:text-sm text-white/90 leading-[1.8] space-y-2 max-w-sm">
                  <li class="flex items-start gap-2">
                    <span class="text-[#0a84ff] font-bold">></span>
                    <span>企业级 AI 基建蓝图设计与规划。</span>
                  </li>
                  <li class="flex items-start gap-2">
                    <span class="text-[#0a84ff] font-bold">></span>
                    <span>量身定制个体生产力演化路线。</span>
                  </li>
                  <li class="flex items-start gap-2">
                    <span class="text-[#0a84ff] font-bold">></span>
                    <span>输出精确到算力与代码的重构方案。</span>
                  </li>
                </ul>
                
                <div class="mt-auto pixel-text text-[9px] md:text-[10px] text-white leading-[1.4] whitespace-pre font-mono">
<span class="text-white">> audit.init</span>
  target   enterprise_infra
  scope    full_stack
<span class="text-[var(--color-turquoise)]">✓</span> blueprint ready
                </div>
              </div>
            </div>

            <!-- AI 产品陪跑 -->
            <div class="group relative bg-black/40 rounded-xl overflow-hidden flex flex-col h-[280px] md:h-[320px] p-6 md:p-8">
              <div class="absolute inset-0 pointer-events-none opacity-60 group-hover:opacity-100 transition-opacity duration-700">
                <div class="absolute -left-20 -bottom-10 w-[400px] h-[400px] mix-blend-screen animate-[spin_45s_linear_infinite]">
                   <div class="absolute top-20 right-0 w-56 h-56 bg-gradient-to-tr from-[#ff375f] to-[#ff9f0a] rounded-full filter blur-[50px] animate-blob-morph opacity-40"></div>
                   <div class="absolute bottom-10 left-20 w-64 h-64 bg-gradient-to-bl from-[#ff453a] to-transparent rounded-full filter blur-[60px] animate-blob-morph animation-delay-2000 opacity-30"></div>
                </div>
              </div>
              <div class="absolute inset-0 bg-[url('data:image/svg+xml,%3Csvg viewBox=%220 0 200 200%22 xmlns=%22http://www.w3.org/2000/svg%22%3E%3Cfilter id=%22noise%22%3E%3CfeTurbulence type=%22fractalNoise%22 baseFrequency=%220.8%22 numOctaves=%224%22 stitchTiles=%22stitch%22/%3E%3C/filter%3E%3Crect width=%22100%25%22 height=%22100%25%22 filter=%22url(%23noise)%22/%3E%3C/svg%3E')] opacity-[0.15] mix-blend-overlay pointer-events-none"></div>

              <div class="relative z-10 flex flex-col h-full">
                <h3 class="text-2xl font-bold text-white mb-4 tracking-wide pixel-zh-title">AI 产品陪跑</h3>
                <ul class="text-xs md:text-sm text-white/90 leading-[1.8] space-y-2 max-w-sm">
                  <li class="flex items-start gap-2">
                    <span class="text-[#ff375f] font-bold">></span>
                    <span>不卖概念只交资产，手搓专属 Agent。</span>
                  </li>
                  <li class="flex items-start gap-2">
                    <span class="text-[#ff375f] font-bold">></span>
                    <span>提供从 Prompt 调优到私有化部署支持。</span>
                  </li>
                  <li class="flex items-start gap-2">
                    <span class="text-[#ff375f] font-bold">></span>
                    <span>深入业务流，让 AI 真正成为数字员工。</span>
                  </li>
                </ul>
                
                <div class="mt-auto pixel-text text-[9px] md:text-[10px] text-white leading-[1.4] whitespace-pre font-mono">
<span class="text-white">> task.deploy_agent</span>
  env      private_cloud
  assets   delivered
<span class="text-[var(--color-turquoise)]">✓</span> agent online . 0.2s eta
                </div>
              </div>
            </div>

            <!-- AI 活动落地 -->
            <div class="group relative bg-black/40 rounded-xl overflow-hidden flex flex-col h-[280px] md:h-[320px] p-6 md:p-8">
              <div class="absolute inset-0 pointer-events-none opacity-60 group-hover:opacity-100 transition-opacity duration-700">
                <div class="absolute right-0 top-20 w-[400px] h-[400px] mix-blend-screen animate-[spin_50s_linear_infinite_reverse]">
                   <div class="absolute top-10 right-20 w-64 h-64 bg-gradient-to-tr from-[#32ade6] to-[#30d158] rounded-full filter blur-[50px] animate-blob-morph opacity-30"></div>
                   <div class="absolute bottom-20 left-10 w-48 h-48 bg-gradient-to-bl from-[#42ffd1] to-transparent rounded-full filter blur-[60px] animate-blob-morph animation-delay-4000 opacity-20"></div>
                </div>
              </div>
              <div class="absolute inset-0 bg-[url('data:image/svg+xml,%3Csvg viewBox=%220 0 200 200%22 xmlns=%22http://www.w3.org/2000/svg%22%3E%3Cfilter id=%22noise%22%3E%3CfeTurbulence type=%22fractalNoise%22 baseFrequency=%220.8%22 numOctaves=%224%22 stitchTiles=%22stitch%22/%3E%3C/filter%3E%3Crect width=%22100%25%22 height=%22100%25%22 filter=%22url(%23noise)%22/%3E%3C/svg%3E')] opacity-[0.15] mix-blend-overlay pointer-events-none"></div>

              <div class="relative z-10 flex flex-col h-full">
                <h3 class="text-2xl font-bold text-white mb-4 tracking-wide pixel-zh-title">AI 活动落地</h3>
                <ul class="text-xs md:text-sm text-white/90 leading-[1.8] space-y-2 max-w-sm">
                  <li class="flex items-start gap-2">
                    <span class="text-[#32ade6] font-bold">></span>
                    <span>策划执行高影响力的线下 Workshop。</span>
                  </li>
                  <li class="flex items-start gap-2">
                    <span class="text-[#32ade6] font-bold">></span>
                    <span>为企业定制内训与闭门技术研讨会。</span>
                  </li>
                  <li class="flex items-start gap-2">
                    <span class="text-[#32ade6] font-bold">></span>
                    <span>提供完整的物料、讲师资源与 SOP。</span>
                  </li>
                </ul>
                
                <div class="mt-auto pixel-text text-[9px] md:text-[10px] text-white leading-[1.4] whitespace-pre font-mono">
<span class="text-white">> network.sync_event</span>
  type     workshop
  scale    enterprise
<span class="text-[var(--color-turquoise)]">✓</span> network connected
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section id="events" class="relative z-10 w-full border-t border-white/5 pointer-events-auto overflow-hidden py-12 bg-black/60 backdrop-blur-xl">
        <div class="relative z-10 w-full px-6 md:px-12 lg:px-16 mx-auto max-w-[1920px]">
          <div class="flex flex-col md:flex-row justify-between items-start md:items-end mb-12 gap-6">
            <div>
              <h2 class="text-3xl md:text-4xl font-bold pixel-zh-title leading-tight text-white tracking-wide">
                {{ featuredContentSlots.eventsTitle }}<br />
                <span class="text-lg md:text-xl text-[var(--color-turquoise)] pixel-text">> SYSTEM.EVENTS</span>
              </h2>
              <p v-if="homeFeatured.eventsDescription" class="pixel-text text-sm text-white/45 mt-4 max-w-2xl leading-relaxed">
                {{ homeFeatured.eventsDescription }}
              </p>
            </div>
            <a :href="routeHref('/events')" target="_blank" rel="noreferrer" class="pixel-text text-xs md:text-sm text-white/60 hover:text-[var(--color-turquoise)] transition-colors border-b border-transparent hover:border-[var(--color-turquoise)] pb-1">
              {{ featuredContentSlots.eventsViewAllLabel }}
            </a>
          </div>

          <div class="flex overflow-x-auto gap-6 pb-8 snap-x snap-mandatory hide-scrollbar">
            <a
              v-for="event in homepageEvents"
              :key="event.slug"
              :href="routeHref(`/events/${event.slug}`)"
              target="_blank"
              rel="noreferrer"
              class="relative bg-[#0a0a0a] border border-white/10 flex flex-col overflow-hidden rounded flex-shrink-0 w-[320px] md:w-[400px] h-[240px] snap-start cursor-pointer"
            >
              <div class="absolute inset-0 z-0">
                <img :src="event.cover" :alt="event.title" class="w-full h-full object-cover grayscale-[40%]" />
                <div class="absolute inset-0 bg-gradient-to-t from-[#0a0a0a] via-[#0a0a0a]/80 to-transparent z-10 pointer-events-none"></div>
              </div>

              <div class="absolute top-4 left-4 z-20 bg-black/80 px-2 py-1 rounded-sm border border-white/10 backdrop-blur-md" :class="eventStatusBadgeClass(event.status)">
                <span class="pixel-text text-[10px] tracking-wider">{{ eventStatusLabel(event.status) }}</span>
              </div>

              <div class="p-5 md:p-6 flex flex-col justify-end flex-1 relative z-20">
                <h3 class="text-base md:text-lg font-bold pixel-zh-title text-white line-clamp-2 leading-snug drop-shadow-md mb-3">{{ event.title }}</h3>
                
                <div class="space-y-2">
                  <div class="pixel-text text-[10px] text-white/60 flex items-center gap-2">
                    <span class="text-[var(--color-turquoise)] opacity-80">> [时间]</span>
                    <span class="tracking-wide text-white/80">{{ event.time }}</span>
                  </div>
                  <div class="pixel-text text-[10px] text-white/60 flex items-center gap-2">
                    <span class="text-[var(--color-turquoise)] opacity-80">> [地点]</span>
                    <span class="tracking-wide text-white/80 line-clamp-1">{{ event.location }}</span>
                  </div>
                </div>
              </div>
            </a>
          </div>
        </div>
      </section>

      <section id="insights" class="relative z-10 w-full border-t border-white/5 pointer-events-auto overflow-hidden py-16 bg-[#050505]">
        <div class="relative z-10 w-full px-6 md:px-12 lg:px-16 mx-auto max-w-[1920px]">
          <div class="flex flex-col md:flex-row justify-between items-start md:items-end mb-12 gap-6">
            <div>
              <h2 class="text-3xl md:text-4xl font-bold pixel-zh-title leading-tight text-white tracking-wide">
                {{ featuredContentSlots.insightsTitle }}<br />
                <span class="text-lg md:text-xl text-[var(--color-turquoise)] pixel-text">> 最新观察 / 信号筛选</span>
              </h2>
              <p class="pixel-text text-sm text-white/45 mt-4 max-w-2xl leading-relaxed">
                {{ homeFeatured.insightsDescription }}
              </p>
            </div>
            <a :href="routeHref('/insights')" target="_blank" rel="noreferrer" class="pixel-text text-xs md:text-sm text-white/60 hover:text-[var(--color-turquoise)] transition-colors border-b border-transparent hover:border-[var(--color-turquoise)] pb-1">
              {{ featuredContentSlots.insightsViewAllLabel }}
            </a>
          </div>

          <!-- 风琴卡片 Accordion 布局 -->
          <div class="flex flex-col lg:flex-row w-full h-[800px] lg:h-[420px] gap-2 lg:gap-4">
            <a
              v-for="(insight, index) in homepageInsights"
              :key="insight.slug"
              :href="routeHref(`/insights/${insight.slug}`)"
              target="_blank"
              rel="noreferrer"
              class="group relative flex-1 hover:flex-[4] lg:hover:flex-[4] transition-all duration-700 ease-[cubic-bezier(0.25,1,0.5,1)] bg-[#0a0a0a] rounded overflow-hidden flex flex-col lg:flex-row border border-white/5 hover:border-[var(--color-turquoise)]/30"
            >
              <!-- 侧边栏/顶部栏 (常驻) -->
              <div class="w-full lg:w-14 h-12 lg:h-full flex lg:flex-col items-center justify-between px-4 lg:p-3 lg:py-6 bg-white/[0.02] border-b lg:border-b-0 lg:border-r border-white/5 shrink-0 z-20">
                <div class="flex lg:flex-col items-center gap-3">
                  <div class="w-1.5 h-1.5 rounded-full bg-[var(--color-turquoise)]/50 group-hover:bg-[var(--color-turquoise)] group-hover:shadow-[0_0_8px_rgba(66,255,209,0.5)] transition-all duration-300 shrink-0"></div>
                  <span class="pixel-text text-[10px] text-white/50 group-hover:text-[var(--color-turquoise)]/80 transition-colors whitespace-nowrap lg:[writing-mode:vertical-rl] lg:rotate-180">
                    LOG_{{ String(index + 1).padStart(3, '0') }}
                  </span>
                </div>
                <span class="pixel-text text-[10px] text-white/30 whitespace-nowrap lg:[writing-mode:vertical-rl] lg:rotate-180">{{ insight.publishedAt }}</span>
              </div>

              <!-- 右侧动态区域 (Mobile 下是下方动态区域) -->
              <div class="relative flex-1 h-full overflow-hidden bg-[#0a0a0a]">
                
                <!-- 常驻背景图：展开时可见度更高，恢复色彩 -->
                <div 
                  class="absolute inset-0 bg-cover bg-center bg-no-repeat transition-all duration-700 opacity-40 mix-blend-luminosity filter grayscale group-hover:opacity-60 group-hover:mix-blend-normal group-hover:grayscale-0 group-hover:scale-105"
                  :style="`background-image: url('${insight.cover || ''}')`"
                ></div>

                <!-- 折叠态：大标题与数字 (仅在未 Hover 时显示) -->
                <div class="flex absolute inset-0 items-center justify-center opacity-100 group-hover:opacity-0 transition-opacity duration-300 pointer-events-none">
                  <!-- 底部深色遮罩，保证白色大标题可读 -->
                  <div class="absolute inset-0 bg-[#0a0a0a]/60 lg:bg-gradient-to-t lg:from-[#0a0a0a] lg:via-[#0a0a0a]/80 lg:to-transparent"></div>
                  
                  <span class="absolute text-white/5 text-[6rem] lg:text-[8rem] font-bold pixel-zh-title">{{ String(index + 1).padStart(2, '0') }}</span>
                  
                  <!-- Mobile: 横向标题, Desktop: 竖向标题 -->
                  <h3 class="z-10 lg:[writing-mode:vertical-rl] text-base lg:text-lg font-bold text-white/50 tracking-widest px-6 lg:px-4 text-center truncate lg:whitespace-nowrap lg:max-h-[80%] overflow-hidden w-full lg:w-auto">
                    {{ insight.title }}
                  </h3>
                </div>

                <!-- 展开态：完整内容 -->
                <div class="w-full absolute inset-0 p-5 md:p-8 flex flex-col justify-end opacity-0 group-hover:opacity-100 transition-opacity duration-500 lg:group-hover:delay-200 z-10 bg-gradient-to-t lg:bg-gradient-to-r from-[#0a0a0a] via-[#0a0a0a]/90 lg:via-[#0a0a0a]/90 to-transparent">
                  <div class="flex flex-col w-full lg:max-w-[450px] xl:max-w-[550px]">
                    <!-- 标签 -->
                    <div class="flex flex-wrap gap-2 mb-3 lg:mb-4">
                      <span class="pixel-text text-[9px] text-[var(--color-turquoise)] bg-[var(--color-turquoise)]/10 px-2 py-1 rounded-sm border border-[var(--color-turquoise)]/20">
                        {{ insight.category }}
                      </span>
                      <span class="pixel-text text-[9px] text-white/40 bg-white/5 px-2 py-1 rounded-sm">
                        {{ insight.audience }}
                      </span>
                    </div>
  
                    <!-- 标题 -->
                    <h3 class="text-base md:text-xl xl:text-2xl font-bold pixel-zh-title text-white/90 group-hover:text-white transition-colors mb-3 lg:mb-4 line-clamp-2 leading-snug whitespace-normal">
                      {{ insight.title }}
                    </h3>
  
                    <!-- 摘要 -->
                    <p class="text-xs md:text-sm text-white/60 leading-relaxed line-clamp-3 mb-4 lg:mb-6 whitespace-normal">
                      {{ insight.summary }}
                    </p>
  
                    <!-- 底部 Meta -->
                    <div class="flex items-center justify-between pt-3 lg:pt-4 border-t border-white/5">
                      <div class="flex items-center gap-2 hidden md:flex">
                        <span class="pixel-text text-[10px] text-white/30">> 标签：</span>
                        <span class="pixel-text text-[9px] text-white/50 truncate max-w-[150px] xl:max-w-[200px]">
                          {{ insight.tags.join(', ') }}
                        </span>
                      </div>
                      <span class="pixel-text text-[10px] text-[var(--color-turquoise)]/0 group-hover:text-[var(--color-turquoise)] transition-colors duration-300 transform translate-x-2 group-hover:translate-x-0 ml-auto">
                        查看详情 ->
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </a>
          </div>
        </div>
      </section>

      <section id="contact" class="relative z-10 w-full border-t border-white/5 pointer-events-auto py-16 bg-[#050505]">
        <div class="relative z-10 w-full px-6 md:px-12 lg:px-16 mx-auto max-w-[1920px]">
          <!-- Title & Description (Top) -->
          <div class="flex flex-col md:flex-row justify-between items-start md:items-center mb-12 gap-6 md:gap-12">
            <div class="md:w-1/2">
              <div class="pixel-text text-xs md:text-sm text-white/40 mb-4 tracking-widest uppercase">联系我们</div>
              <h2 class="text-4xl md:text-5xl font-bold pixel-zh-title leading-tight text-white tracking-wide">
                联系 VOID LAB<span class="text-white">。</span>
              </h2>
            </div>
            <div class="md:w-1/2">
              <p class="text-white/55 text-sm md:text-base leading-relaxed">
                想加入社区、合作办活动、赞助支持、学校/社区合作，或咨询 VOID LAB，都可以先从这里联系。
              </p>
            </div>
          </div>

          <!-- Bottom Grid -->
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 border border-white/10 rounded-sm overflow-hidden bg-[#0a0a0a]">
            <!-- Item 1: 小红书 -->
            <div class="p-8 border-b lg:border-b-0 md:border-r border-white/10 flex flex-col hover:bg-white/[0.02] transition-colors group">
              <div class="w-10 h-10 rounded-full border border-white/20 flex items-center justify-center mb-6 text-white/80 group-hover:border-white transition-colors">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path></svg>
              </div>
              <h3 class="text-lg font-bold text-white pixel-zh-title mb-3">官方小红书</h3>
              <p class="text-sm text-white/50 leading-relaxed mb-6 flex-1">
                关注 VOID LAB 的活动回顾、现场照片、社区动态和新活动发布。
              </p>
              <div class="text-xs text-white/80 mb-6 font-mono">小红书号 VOIDLAB_AI</div>
              <button class="w-fit px-6 py-2 rounded-full border border-white/20 text-sm text-white hover:bg-white hover:text-black transition-colors pixel-zh-title">
                关注小红书
              </button>
            </div>

            <!-- Item 2: 社区群 -->
            <div class="p-8 border-b lg:border-b-0 lg:border-r border-white/10 flex flex-col hover:bg-white/[0.02] transition-colors group">
              <div class="w-10 h-10 rounded-full border border-white/20 flex items-center justify-center mb-6 text-white/80 group-hover:border-white transition-colors">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
              </div>
              <h3 class="text-lg font-bold text-white pixel-zh-title mb-3">加入社区群</h3>
              <p class="text-sm text-white/50 leading-relaxed mb-6 flex-1">
                加入 VOID LAB 微信群，获取活动通知、结识伙伴和线下聚会信息。
              </p>
              <div class="text-xs text-white/80 mb-6 font-mono">企业微信入群二维码</div>
              <button class="w-fit px-6 py-2 rounded-full border border-white/20 text-sm text-white hover:bg-white hover:text-black transition-colors pixel-zh-title">
                查看二维码
              </button>
            </div>

            <!-- Item 3: 官方邮箱 -->
            <div class="p-8 border-b md:border-b-0 md:border-r border-white/10 flex flex-col hover:bg-white/[0.02] transition-colors group">
              <div class="w-10 h-10 rounded-full border border-white/20 flex items-center justify-center mb-6 text-white/80 group-hover:border-white transition-colors">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="4" width="20" height="16" rx="2"></rect><path d="m2 4 10 8 10-8"></path></svg>
              </div>
              <h3 class="text-lg font-bold text-white pixel-zh-title mb-3">官方邮箱</h3>
              <p class="text-sm text-white/50 leading-relaxed mb-6 flex-1">
                合作、赞助、学校/社区组织、媒体和正式事项联系。
              </p>
              <div class="text-xs text-white/80 mb-6 font-mono uppercase">join@voidlab.ai</div>
              <button class="w-fit px-6 py-2 rounded-full border border-white/20 text-sm text-white hover:bg-white hover:text-black transition-colors pixel-zh-title">
                发邮件
              </button>
            </div>

            <!-- Item 4: 官方小助手 -->
            <div class="p-8 flex flex-col hover:bg-white/[0.02] transition-colors group">
              <div class="w-10 h-10 rounded-full border border-white/20 flex items-center justify-center mb-6 text-white/80 group-hover:border-white transition-colors">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
              </div>
              <h3 class="text-lg font-bold text-white pixel-zh-title mb-3">官方小助手</h3>
              <p class="text-sm text-white/50 leading-relaxed mb-6 flex-1">
                咨询、报名、入群、活动问题和成员对接支持。
              </p>
              <div class="text-xs text-white/80 mb-6 font-mono">企业微信客服链接</div>
              <button class="w-fit px-6 py-2 rounded-full border border-white/20 text-sm text-white hover:bg-white hover:text-black transition-colors pixel-zh-title">
                添加小助手
              </button>
            </div>

          </div>
        </div>
      </section>

    <!-- 彻底移除多余的 spacer -->
    </div>

    <!-- 将 footer 改为相对流式布局，去除 fixed -->
    <footer id="footer" class="relative w-full z-[5] bg-[#050505] border-t border-white/10 pointer-events-auto text-white/60 py-12 px-6 md:px-12 lg:px-16">
      <div class="absolute inset-0 bg-[url('data:image/svg+xml,%3Csvg viewBox=%220 0 200 200%22 xmlns=%22http://www.w3.org/2000/svg%22%3E%3Cfilter id=%22noise%22%3E%3CfeTurbulence type=%22fractalNoise%22 baseFrequency=%220.8%22 numOctaves=%224%22 stitchTiles=%22stitch%22/%3E%3C/filter%3E%3Crect width=%22100%25%22 height=%22100%25%22 filter=%22url(%23noise)%22/%3E%3C/svg%3E')] opacity-[0.08] pointer-events-none mix-blend-overlay z-0"></div>
      
      <div class="relative z-10 mx-auto max-w-[1920px] flex flex-col md:flex-row justify-between items-start md:items-center gap-8">
        
        <!-- 左侧：Logo & Slogan -->
        <div class="flex flex-col gap-3">
          <div class="flex items-center gap-3">
            <div class="w-5 h-5 flex flex-wrap gap-[1px]">
              <div class="w-[9px] h-[9px] bg-[var(--color-turquoise)] opacity-90"></div>
              <div class="w-[9px] h-[9px] bg-white/30"></div>
              <div class="w-[9px] h-[9px] bg-white/30"></div>
              <div class="w-[9px] h-[9px] bg-white/80"></div>
            </div>
            <span class="pixel-text text-xl tracking-[0.2em] font-bold text-white">VOID LAB</span>
          </div>
          <p class="pixel-text text-xs text-white/40">
            {{ footerConfig.slogan }}
          </p>
        </div>

        <!-- 中间：导航链接 -->
        <div class="flex flex-col md:items-center w-full md:w-auto">
          <h4 class="text-white/30 pixel-text text-[10px] tracking-widest mb-3 uppercase">VOIDLAB.AI</h4>
          <ul class="flex flex-col gap-2 pixel-zh-title text-sm">
            <li v-for="item in footerConfig.navLinks" :key="item.label">
              <a :href="routeHref(item.path)" target="_blank" class="text-white/70 hover:text-white transition-colors">{{ item.label }}</a>
            </li>
          </ul>
        </div>

        <!-- 右侧：回到顶部悬浮按钮 (复刻图中右侧的绿色小方块按钮) -->
        <div class="self-end md:self-center">
          <button 
            @click="scrollToTop"
            class="w-12 h-12 bg-[var(--color-turquoise)] rounded-xl flex items-center justify-center cursor-pointer hover:scale-105 hover:shadow-[0_0_20px_rgba(66,255,209,0.3)] transition-all duration-300 group"
          >
            <!-- 内部的 4 瓣像素点阵 -->
            <div class="w-5 h-5 flex flex-wrap gap-[2px]">
              <div class="w-[9px] h-[9px] bg-black group-hover:opacity-80 transition-opacity"></div>
              <div class="w-[9px] h-[9px] bg-black group-hover:opacity-80 transition-opacity"></div>
              <div class="w-[9px] h-[9px] bg-black group-hover:opacity-80 transition-opacity"></div>
              <div class="w-[9px] h-[9px] bg-black group-hover:opacity-80 transition-opacity"></div>
            </div>
          </button>
        </div>

      </div>
      <div class="relative z-10 mx-auto mt-8 max-w-[1920px] border-t border-white/5 pt-6 text-xs text-white/35 pixel-text">
        {{ footerConfig.legalText }}
      </div>
    </footer>
  </div>
</template>

<style scoped>
@keyframes blob-morph {
  0%, 100% { border-radius: 60% 40% 30% 70% / 60% 30% 70% 40%; }
  50% { border-radius: 30% 60% 70% 40% / 50% 60% 30% 60%; }
}
.animate-blob-morph {
  animation: blob-morph 8s ease-in-out infinite alternate;
}
.animation-delay-2000 {
  animation-delay: 2s;
}
.animation-delay-4000 {
  animation-delay: 4s;
}
</style>
