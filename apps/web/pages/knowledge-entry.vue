<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import KnowledgeTokenDialog from "../components/KnowledgeTokenDialog.vue";
import KnowledgeSidebar from "../components/KnowledgeSidebar.vue";
import SiteHeader from "../components/SiteHeader.vue";
import { useKnowledgeBase } from "../composables/useKnowledgeBase";
import { useKnowledgeSidebar } from "../composables/useKnowledgeSidebar";
import { renderMarkdown, extractMarkdownHeadings } from "../src/markdown";

const route = useRoute();
const router = useRouter();
const {
  getEntryBySlug,
  getEntriesBySpace,
  getSpaceBySlug,
  canReadEntry,
  isSpaceUnlocked,
  resolveKnowledgeAssetUrl,
  loadKnowledgeSpaceBySlug,
  loadKnowledgeEntryBySlug,
  unlockSpace
} = useKnowledgeBase();
const { openSidebarDrawer, closeSidebarDrawer, sidebarDrawerOpen } = useKnowledgeSidebar();

const spaceSlug = computed(() => String(route.params.spaceSlug ?? ""));
const entrySlug = computed(() => String(route.params.entrySlug ?? ""));
const tokenDialogOpen = ref(false);
const tokenError = ref("");
const pageLoading = ref(true);
const pageReady = ref(false);
const pageError = ref("");
const mobileTocOpen = ref(false);

const space = computed(() => getSpaceBySlug(spaceSlug.value));
const entry = computed(() => getEntryBySlug(spaceSlug.value, entrySlug.value));
const canRead = computed(() => canReadEntry(spaceSlug.value, entrySlug.value));
const hasRenderableContent = computed(() => Boolean(entry.value?.contentMarkdown?.trim()));
const renderedContent = computed(() =>
  entry.value
    ? renderMarkdown(entry.value.contentMarkdown, {
        imageResolver: (src) => resolveKnowledgeAssetUrl(spaceSlug.value, src)
      })
    : ""
);
const orderedEntries = computed(() =>
  getEntriesBySpace(spaceSlug.value)
    .filter((item) => item.status === "published")
    .slice()
    .sort((left, right) => {
      if (left.sortOrder === right.sortOrder) {
        return left.id - right.id;
      }
      return left.sortOrder - right.sortOrder;
    })
);
const currentEntryIndex = computed(() => orderedEntries.value.findIndex((item) => item.slug === entrySlug.value));
const previousEntry = computed(() =>
  currentEntryIndex.value > 0 ? orderedEntries.value[currentEntryIndex.value - 1] : null
);
const nextEntry = computed(() =>
  currentEntryIndex.value >= 0 && currentEntryIndex.value < orderedEntries.value.length - 1
    ? orderedEntries.value[currentEntryIndex.value + 1]
    : null
);
const hasAdjacentEntries = computed(() => Boolean(previousEntry.value || nextEntry.value));
const formattedUpdatedAt = computed(() => formatUpdatedAt(entry.value?.updatedAt || ""));
const headings = computed(() => extractMarkdownHeadings(entry.value?.contentMarkdown || ""));
const desktopViewportWidth = ref(0);
const activeHeadingId = ref("");
let headingObserver: IntersectionObserver | null = null;
let mobileTocOriginalBodyOverflow = "";
let mobileTocOriginalBodyPosition = "";
let mobileTocOriginalBodyTop = "";
let mobileTocOriginalBodyWidth = "";
let mobileTocOriginalHtmlOverflow = "";
let mobileTocLockedScrollY = 0;

const desktopSidebarStyle = computed(() => {
  const width = desktopViewportWidth.value;
  if (!width) {
    return { left: "0px", width: "16rem" };
  }

  const containerWidth = Math.min(width, 1440);
  const gutter = width > 1280 ? 48 : width >= 768 ? 32 : 16;
  const left = Math.max((width - containerWidth) / 2 + gutter, 0);

  return {
    left: `${left}px`,
    width: "16rem"
  };
});

const desktopTocStyle = computed(() => {
  const width = desktopViewportWidth.value;
  if (!width) {
    return { right: "0px", width: "10rem" };
  }

  const containerWidth = Math.min(width, 1440);
  const gutter = width > 1280 ? 48 : width >= 768 ? 32 : 16;
  const right = Math.max((width - containerWidth) / 2 + gutter, 0);

  return {
    right: `${right}px`,
    width: "10rem"
  };
});

function updateDesktopViewportWidth() {
  if (typeof window === "undefined") {
    return;
  }
  desktopViewportWidth.value = window.innerWidth;
}

function disconnectHeadingObserver() {
  headingObserver?.disconnect();
  headingObserver = null;
}

function updateActiveHeadingFromScroll() {
  if (typeof document === "undefined") {
    return;
  }

  const visibleHeadings = headings.value
    .map((heading) => document.getElementById(heading.id))
    .filter((element): element is HTMLElement => Boolean(element))
    .filter((element) => element.getBoundingClientRect().top <= 140);

  const currentHeading = visibleHeadings[visibleHeadings.length - 1];
  activeHeadingId.value = currentHeading?.id ?? headings.value[0]?.id ?? "";
}

function setupHeadingObserver() {
  if (typeof window === "undefined" || typeof document === "undefined") {
    return;
  }

  disconnectHeadingObserver();

  if (!headings.value.length) {
    activeHeadingId.value = "";
    return;
  }

  window.requestAnimationFrame(() => {
    const elements = headings.value
      .map((heading) => document.getElementById(heading.id))
      .filter((element): element is HTMLElement => Boolean(element));

    if (!elements.length) {
      activeHeadingId.value = headings.value[0]?.id ?? "";
      return;
    }

    headingObserver = new IntersectionObserver(
      () => {
        updateActiveHeadingFromScroll();
      },
      {
        root: null,
        rootMargin: "-120px 0px -55% 0px",
        threshold: [0, 0.1, 0.25, 0.5, 1]
      }
    );

    elements.forEach((element) => {
      headingObserver?.observe(element);
    });

    updateActiveHeadingFromScroll();
  });
}

watch(
  [spaceSlug, entrySlug, () => isSpaceUnlocked(spaceSlug.value)],
  async ([nextSpaceSlug, nextEntrySlug]) => {
    if (!nextSpaceSlug || !nextEntrySlug) {
      router.replace("/knowledge");
      return;
    }

    const hasCachedEntry = Boolean(getEntryBySlug(nextSpaceSlug, nextEntrySlug));
    pageLoading.value = true;
    activeHeadingId.value = "";
    // 只有首屏没有缓存时才显示整页 loading，路由内跳转保持页面“丝滑”不闪白
    if (!hasCachedEntry) {
      pageReady.value = false;
    }
    pageError.value = "";

    try {
      const tocPayload = await loadKnowledgeSpaceBySlug(nextSpaceSlug, true);
      if (!tocPayload?.space) {
        router.replace("/knowledge");
        return;
      }

      const currentEntry = getEntryBySlug(nextSpaceSlug, nextEntrySlug);
      if (!currentEntry) {
        router.replace(`/knowledge/${nextSpaceSlug}`);
        return;
      }

      if (currentEntry.isPreview || isSpaceUnlocked(nextSpaceSlug)) {
        await loadKnowledgeEntryBySlug(nextSpaceSlug, nextEntrySlug, true);
      }

      const finalEntry = getEntryBySlug(nextSpaceSlug, nextEntrySlug);
      document.title = finalEntry ? `VOIDLAB | ${finalEntry.title}` : `VOIDLAB | ${tocPayload.space.title}`;
      pageReady.value = true;
    } catch (error: unknown) {
      if (error instanceof Error && error.message === "knowledge entry is locked") {
        pageReady.value = true;
        return;
      }
      pageError.value = error instanceof Error ? error.message : "加载知识文档失败";
    } finally {
      pageLoading.value = false;
    }
  },
  { immediate: true }
);

function openUnlockDialog() {
  tokenError.value = "";
  tokenDialogOpen.value = true;
}

function handleOpenGlobalDirectory() {
  closeMobileToc();
  openSidebarDrawer();
}

function handleOpenMobileToc() {
  if (sidebarDrawerOpen.value) {
    closeSidebarDrawer();
  }
  mobileTocOpen.value = true;
}

function closeMobileToc() {
  mobileTocOpen.value = false;
}

function handleUnlock(token: string) {
  if (!space.value) {
    return;
  }

  void unlockSpace(space.value.slug, token)
    .then(async (result) => {
      if (!result.success) {
        tokenError.value = result.message;
        return;
      }

      tokenDialogOpen.value = false;
      tokenError.value = "";

      if (space.value && entry.value) {
        await loadKnowledgeEntryBySlug(space.value.slug, entry.value.slug, true);
      }
    })
    .catch((error: unknown) => {
      tokenError.value = error instanceof Error ? error.message : "令牌验证失败";
    });
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function formatUpdatedAt(value: string) {
  if (!value) {
    return "";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "long",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false
  }).format(date);
}

onMounted(() => {
  updateDesktopViewportWidth();
  window.addEventListener("resize", updateDesktopViewportWidth);
  setupHeadingObserver();
});

onBeforeUnmount(() => {
  if (typeof window === "undefined") {
    return;
  }
  window.removeEventListener("resize", updateDesktopViewportWidth);
  disconnectHeadingObserver();
  document.documentElement.style.overflow = mobileTocOriginalHtmlOverflow;
  document.body.style.overflow = mobileTocOriginalBodyOverflow;
  document.body.style.position = mobileTocOriginalBodyPosition;
  document.body.style.top = mobileTocOriginalBodyTop;
  document.body.style.width = mobileTocOriginalBodyWidth;
});

watch([headings, renderedContent, pageReady], () => {
  if (!pageReady.value) {
    return;
  }
  setupHeadingObserver();
});

watch(mobileTocOpen, (open) => {
  if (typeof document === "undefined" || typeof window === "undefined") {
    return;
  }

  if (open) {
    mobileTocLockedScrollY = window.scrollY;
    mobileTocOriginalBodyOverflow = document.body.style.overflow;
    mobileTocOriginalBodyPosition = document.body.style.position;
    mobileTocOriginalBodyTop = document.body.style.top;
    mobileTocOriginalBodyWidth = document.body.style.width;
    mobileTocOriginalHtmlOverflow = document.documentElement.style.overflow;
    document.documentElement.style.overflow = "hidden";
    document.body.style.overflow = "hidden";
    document.body.style.position = "fixed";
    document.body.style.top = `-${mobileTocLockedScrollY}px`;
    document.body.style.width = "100%";
    return;
  }

  document.documentElement.style.overflow = mobileTocOriginalHtmlOverflow;
  document.body.style.overflow = mobileTocOriginalBodyOverflow;
  document.body.style.position = mobileTocOriginalBodyPosition;
  document.body.style.top = mobileTocOriginalBodyTop;
  document.body.style.width = mobileTocOriginalBodyWidth;
  window.scrollTo(0, mobileTocLockedScrollY);
});

watch(
  () => route.fullPath,
  () => {
    if (mobileTocOpen.value) {
      closeMobileToc();
    }
  }
);
</script>

<template>
  <div class="relative z-10 min-h-screen bg-[#f6f5f1] text-[#333333] font-sans selection:bg-[#cce2ff] pt-16 flex flex-col">
    <SiteHeader theme="light" activePath="/knowledge" />

    <div
      v-if="mobileTocOpen"
      class="fixed inset-0 z-[75] md:hidden"
      :style="{
        height: 'calc(var(--vvh, 1vh) * 100)',
        minHeight: '100vh'
      }"
      aria-modal="true"
      role="dialog"
    >
      <div class="absolute inset-0 bg-black/35" @click="closeMobileToc"></div>
      <transition
        appear
        enter-active-class="transition-transform duration-200 ease-out"
        enter-from-class="translate-x-full"
        enter-to-class="translate-x-0"
        leave-active-class="transition-transform duration-200 ease-in"
        leave-from-class="translate-x-0"
        leave-to-class="translate-x-full"
      >
        <aside
          v-if="headings.length > 0"
          class="absolute right-0 top-0 bottom-0 h-full w-[82vw] max-w-[320px] overflow-y-auto overscroll-contain border-l border-[#eaeaea] bg-[#f6f5f1]"
          :style="{
            WebkitOverflowScrolling: 'touch',
            paddingTop: 'env(safe-area-inset-top)',
            paddingBottom: 'env(safe-area-inset-bottom)'
          }"
        >
          <div class="sticky top-0 z-10 flex items-center justify-between border-b border-[#eaeaea] bg-[#f6f5f1] px-4 py-3">
            <div>
              <div class="text-[12px] font-medium tracking-[0.08em] text-[#999999] uppercase">本页目录</div>
              <div class="mt-1 text-[15px] font-bold text-[#333333]">{{ entry?.title }}</div>
            </div>
            <button
              type="button"
              class="inline-flex h-9 w-9 items-center justify-center rounded border border-[#eaeaea] bg-white text-[#777777] transition-colors hover:bg-[#f0f0f0] hover:text-[#333333]"
              title="关闭本页目录"
              @click="closeMobileToc"
            >
              ✕
            </button>
          </div>

          <div class="px-4 py-4">
            <div class="flex flex-col border-l border-[#eaeaea]">
              <a
                v-for="h in headings"
                :key="h.id"
                :href="`#${h.id}`"
                class="block border-l-2 py-2 text-[13px] transition-colors"
                :class="
                  activeHeadingId === h.id
                    ? 'border-[#6f8f43] text-[#6f8f43] font-semibold'
                    : 'border-transparent text-[#8a8a85] hover:text-[#111111]'
                "
                :style="{ paddingLeft: `${(h.level - 1) * 0.75 + 0.75}rem` }"
                :title="h.title"
                @click="closeMobileToc"
              >
                {{ h.title }}
              </a>
            </div>
          </div>
        </aside>
      </transition>
    </div>

    <div v-if="pageLoading && !pageReady" class="mx-auto flex min-h-[50vh] w-full max-w-5xl items-center justify-center px-6 text-[#787774]">
      正在加载知识文档...
    </div>

    <div v-else-if="pageError" class="mx-auto flex min-h-[50vh] w-full max-w-3xl flex-col items-center justify-center gap-4 px-6 text-center">
      <div class="text-4xl">⚠️</div>
      <div>
        <h1 class="text-2xl font-semibold text-[#37352f]">知识文档暂时打不开</h1>
        <p class="mt-2 text-[15px] text-[#787774]">{{ pageError }}</p>
      </div>
      <RouterLink
        :to="spaceSlug ? `/knowledge/${spaceSlug}` : '/knowledge'"
        class="rounded border border-[#e9e9e7] bg-transparent px-4 py-2 text-[14px] font-medium text-[#37352f] transition-colors hover:bg-black/5"
      >
        返回目录
      </RouterLink>
    </div>

    <div
      v-else-if="space && entry && pageReady"
      class="relative flex w-full flex-1 mx-auto max-w-[1440px] px-4 md:px-8 xl:px-12"
    >
      <div class="hidden md:block w-64 flex-shrink-0"></div>
      <KnowledgeSidebar embedded :desktop-fixed-style="desktopSidebarStyle" />

      <main class="flex-1 min-w-0 flex flex-col">
        <!-- Mobile subnav: follow the reference screenshot -->
        <div class="md:hidden sticky top-16 z-20 bg-[#f6f5f1]/95 backdrop-blur-md border-b border-[#eaeaea]">
          <div class="mx-auto flex w-full max-w-4xl items-center justify-between px-6 py-4 text-[14px] text-[#787774]">
            <button
              type="button"
              class="inline-flex items-center gap-2 text-[14px] font-medium text-[#37352f] transition-colors hover:text-black"
              @click="handleOpenGlobalDirectory"
            >
              目录 <span class="text-[#9ca3af]">‹</span>
            </button>
            <button
              type="button"
              class="inline-flex items-center gap-2 text-[14px] font-medium text-[#37352f] transition-colors hover:text-black"
              @click="handleOpenMobileToc"
            >
              本页目录 <span class="text-[#9ca3af]">›</span>
            </button>
          </div>
        </div>

        <div class="flex w-full items-start justify-start gap-12 px-6 py-12 md:px-8 xl:px-10 flex-1">
          <div class="w-full max-w-[760px] flex-1 min-w-0">
            <article>
              <h1 class="text-4xl font-black leading-[1.15] tracking-[-0.03em] text-[#111111] md:text-[3rem]">
                {{ entry.title }}
              </h1>

            <div class="mt-6 flex items-center gap-4 text-[13px] text-[#787774] border-b border-[#e9e9e7] pb-6">
              <span class="flex items-center gap-1.5">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
                {{ entry.estimatedReadMinutes }} min read
              </span>
              <span>·</span>
              <span
                class="rounded px-2 py-0.5 text-[13px]"
                :class="entry.isPreview ? 'bg-[#e0f2fe] text-[#1e3a8a]' : canRead ? 'bg-[#dcfce7] text-[#14532d]' : 'bg-[#ffedd5] text-[#9a3412]'"
              >
                {{ entry.isPreview ? "Preview" : canRead ? "Unlocked" : "Locked" }}
              </span>
            </div>

            <div v-if="canRead && hasRenderableContent" class="markdown-entry mt-10" v-html="renderedContent"></div>

            <div v-else-if="canRead" class="mt-10 rounded-md border border-[#e9e9e7] bg-[#f7f7f5] p-8">
              <div class="text-[13px] font-medium uppercase tracking-[0.08em] text-[#787774]">Preview</div>
              <h2 class="mt-3 text-xl font-semibold text-[#37352f]">这篇内容的正文还没有显示出来</h2>
              <p class="mt-3 text-[15px] leading-7 text-[#5f5e58]">
                当前页面不应该是一整块空白，所以这里先做了兜底展示。你现在至少能看到这篇内容的摘要，避免页面像坏掉一样。
              </p>
              <div v-if="entry.publicSummary" class="mt-5 rounded border border-[#e9e9e7] bg-white px-5 py-4 text-[15px] leading-7 text-[#37352f]">
                {{ entry.publicSummary }}
              </div>
            </div>

            <div v-else class="mt-10 rounded-md border border-[#e9e9e7] bg-[#f7f7f5] p-8 text-center">
              <div class="text-4xl mb-4">🔒</div>
              <h2 class="text-xl font-semibold text-[#37352f]">正文受控</h2>
              <p class="mt-2 text-[15px] text-[#787774]">
                这是一篇受控内容，输入访问令牌解锁整个 <strong>{{ space.title }}</strong> 后即可阅读。
              </p>
              <button
                type="button"
                class="mt-6 inline-flex items-center justify-center rounded border border-[#e9e9e7] bg-white px-4 py-2 text-[14px] font-medium text-[#37352f] transition-colors hover:bg-gray-50 shadow-sm"
                @click="openUnlockDialog"
              >
                输入令牌解锁
              </button>
            </div>

            <div v-if="formattedUpdatedAt" class="mt-14 text-[14px] text-[#787774]">
              最后更新：{{ formattedUpdatedAt }}
            </div>

            <div v-if="hasAdjacentEntries" class="mt-8 border-t border-[#e9e9e7] pt-8">
              <div class="grid gap-5 md:grid-cols-2">
                <RouterLink
                  v-if="previousEntry"
                  :to="`/knowledge/${space.slug}/${previousEntry.slug}`"
                  class="group rounded-2xl border border-[#eaeaea] bg-white px-5 py-5 md:px-6 md:py-7 transition-all duration-200 hover:border-[#cccccc] hover:bg-[#f0f0f0]"
                >
                  <div class="text-[13px] font-medium text-[#999999]">上一篇</div>
                  <div class="mt-2 text-[16px] md:mt-3 md:text-[24px] font-semibold leading-[1.35] tracking-[-0.01em] text-[#333333] transition-colors group-hover:text-[#111111]">
                    {{ previousEntry.title }}
                  </div>
                </RouterLink>
                <div v-else class="hidden md:block"></div>

                <RouterLink
                  v-if="nextEntry"
                  :to="`/knowledge/${space.slug}/${nextEntry.slug}`"
                  class="group rounded-2xl border border-[#eaeaea] bg-white px-5 py-5 md:px-6 md:py-7 text-right transition-all duration-200 hover:border-[#cccccc] hover:bg-[#f0f0f0]"
                >
                  <div class="text-[13px] font-medium text-[#999999]">下一篇</div>
                  <div class="mt-2 text-[16px] md:mt-3 md:text-[24px] font-semibold leading-[1.35] tracking-[-0.01em] text-[#333333] transition-colors group-hover:text-[#111111]">
                    {{ nextEntry.title }}
                  </div>
                </RouterLink>
              </div>
            </div>
          </article>
          </div>

          <div v-if="headings.length > 0" class="hidden xl:block w-40 flex-shrink-0"></div>
          <aside
            v-if="headings.length > 0"
            class="hidden xl:block fixed top-24 z-20 max-h-[calc(100vh-6rem)] overflow-y-auto hide-scrollbar pt-1 bg-[#f6f5f1]"
            :style="desktopTocStyle"
          >
            <div class="text-[11px] font-bold tracking-[0.08em] text-[#aaaaaa] mb-3 uppercase">本页目录</div>
            <div class="flex flex-col border-l border-[#eaeaea]">
              <a
                v-for="h in headings"
                :key="h.id"
                :href="`#${h.id}`"
                class="block border-l-2 py-1 text-[12px] transition-colors truncate"
                :class="
                  activeHeadingId === h.id
                    ? 'border-[#6f8f43] text-[#6f8f43] font-semibold'
                    : 'border-transparent text-[#8a8a85] hover:text-[#111111]'
                "
                :style="{ paddingLeft: `${(h.level - 1) * 0.75 + 0.75}rem` }"
                :title="h.title"
              >
                {{ h.title }}
              </a>
            </div>
          </aside>
        </div>

        <footer class="mt-auto border-t border-[#e9e9e7] py-8 text-sm text-[#9ca3af]">
          <div class="mx-auto flex w-full flex-col gap-4 px-6 md:flex-row md:items-center md:justify-between md:px-12 lg:px-24 max-w-4xl">
            <div>© 2026 VOIDLAB.AI</div>
            <button type="button" class="text-left transition-colors hover:text-[#37352f] md:text-right" @click="scrollToTop">
              回到顶部
            </button>
          </div>
        </footer>
      </main>
    </div>

    <KnowledgeTokenDialog
      v-if="space"
      v-model:open="tokenDialogOpen"
      :space-title="space.title"
      :hint="space.tokenHint"
      :error-message="tokenError"
      @submit="handleUnlock"
    />
  </div>
</template>

<style scoped>
:deep(.markdown-entry) {
  color: #37352f;
  font-size: 16px;
  line-height: 1.7;
}

:deep(.markdown-entry > :first-child) {
  margin-top: 0;
}

:deep(.markdown-entry h1),
:deep(.markdown-entry h2),
:deep(.markdown-entry h3) {
  margin-top: 2rem;
  margin-bottom: 1rem;
  color: #37352f;
  font-weight: 900;
  line-height: 1.3;
  font-family: "Arial Black", "Impact", "Inter", "Heiti SC", "Microsoft YaHei", sans-serif;
  letter-spacing: -0.02em;
}

:deep(.markdown-entry h1) { font-size: 2.25rem; }
:deep(.markdown-entry h2) { font-size: 1.5rem; }
:deep(.markdown-entry h3) { font-size: 1.25rem; }

:deep(.markdown-entry h4),
:deep(.markdown-entry h5),
:deep(.markdown-entry h6) {
  margin-top: 1.5rem;
  margin-bottom: 0.75rem;
  color: #333333;
  font-weight: 800;
  line-height: 1.35;
  font-family: "Arial Black", "Impact", "Inter", "Heiti SC", "Microsoft YaHei", sans-serif;
  letter-spacing: -0.01em;
}

:deep(.markdown-entry h4) { font-size: 1.15rem; }
:deep(.markdown-entry h5) { font-size: 1.05rem; }
:deep(.markdown-entry h6) { font-size: 0.95rem; color: #5f5e58; }

:deep(.markdown-entry h1) {
  position: relative;
  padding-bottom: 0.3em;
  padding-left: 0;
  margin-top: 2rem;
  margin-bottom: 1.5rem;
  color: #111111;
  font-weight: 900;
  font-family: "Arial Black", "Impact", "Inter", "Heiti SC", "Microsoft YaHei", sans-serif;
  letter-spacing: -0.02em;
}

:deep(.markdown-entry h1::after) {
  content: "";
  position: absolute;
  left: 0;
  bottom: 0;
  width: 1.5em;
  height: 5px;
  background-color: #c4f000;
}

:deep(.markdown-entry h2) {
  position: relative;
  padding-left: 1.25rem;
  padding-top: 1.5rem;
  margin-top: 2.5rem;
  margin-bottom: 1.25rem;
  border-top: 1px solid #eaeaea;
  color: #222222;
  font-weight: 800;
  font-family: "Arial Black", "Impact", "Inter", "Heiti SC", "Microsoft YaHei", sans-serif;
  letter-spacing: -0.01em;
}

:deep(.markdown-entry h3) {
  position: relative;
  padding-left: 1.25rem;
  margin-top: 2rem;
  margin-bottom: 1rem;
  color: #333333;
  font-weight: 800;
  font-family: "Arial Black", "Impact", "Inter", "Heiti SC", "Microsoft YaHei", sans-serif;
  letter-spacing: -0.01em;
}

:deep(.markdown-entry h2::before),
:deep(.markdown-entry h3::before) {
  content: "";
  position: absolute;
  left: 0;
  top: auto;
  margin-top: 0.4em;
  width: 0.75em;
  height: 0.75em;
  background-image: url("data:image/svg+xml,%3Csvg width='14' height='14' viewBox='0 0 14 14' fill='none' xmlns='http://www.w3.org/2000/svg'%3E%3Crect x='5' y='5' width='9' height='9' fill='%23eaeaea'/%3E%3Crect x='0' y='0' width='9' height='9' fill='%23c4f000'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: center;
  background-size: contain;
}

:deep(.markdown-entry p) {
  margin: 0.5rem 0 1rem 0;
}

:deep(.markdown-entry a) {
  color: #0f7b6c;
  text-decoration: underline;
  text-decoration-thickness: 1px;
  text-underline-offset: 2px;
}

:deep(.markdown-entry a:hover) {
  color: #0b5f54;
}

:deep(.markdown-entry hr) {
  border: none;
  border-top: 1px solid #e9e9e7;
  margin: 1.75rem 0;
}

:deep(.markdown-entry ul),
:deep(.markdown-entry ol) {
  margin: 0.5rem 0 1rem 0;
  padding-left: 1.25rem;
  list-style-position: outside;
}

:deep(.markdown-entry ul) {
  list-style-type: disc;
}

:deep(.markdown-entry ol) {
  list-style-type: decimal;
}

:deep(.markdown-entry li) {
  margin: 0.35rem 0;
  line-height: 1.65;
  padding-left: 0.15rem;
}

:deep(.markdown-entry li > p) {
  margin: 0;
  display: inline;
}

:deep(.markdown-entry ul ul) {
  list-style-type: circle;
}

:deep(.markdown-entry ul ul ul) {
  list-style-type: square;
}

:deep(.markdown-entry ul ul),
:deep(.markdown-entry ul ol),
:deep(.markdown-entry ol ul),
:deep(.markdown-entry ol ol) {
  margin-top: 0.25rem;
  margin-bottom: 0.25rem;
}

:deep(.markdown-entry input[type="checkbox"]) {
  width: 0.95rem;
  height: 0.95rem;
  margin-right: 0.5rem;
  accent-color: #0f7b6c;
  vertical-align: -2px;
}

:deep(.markdown-entry blockquote) {
  margin: 1rem 0;
  border-left: 3px solid #d0d0ca;
  padding: 0.2rem 1rem;
  color: #787774;
  background: #f7f7f5;
  border-radius: 8px;
}

:deep(.markdown-entry code) {
  border-radius: 3px;
  background: #f1f1ef;
  padding: 0.2rem 0.4rem;
  color: #eb5757;
  font-size: 0.85em;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}

:deep(.markdown-entry .md-codeblock) {
  margin: 1rem 0 1.25rem 0;
}

:deep(.markdown-entry .md-codeblock) {
  position: relative;
}

:deep(.markdown-entry .md-codeblock[data-lang]::before) {
  content: attr(data-lang);
  position: absolute;
  top: 0.55rem;
  left: 1rem;
  font-size: 12px;
  line-height: 1;
  letter-spacing: 0.04em;
  color: #9ca3af;
  text-transform: uppercase;
  pointer-events: none;
}

:deep(.markdown-entry .md-codeblock-copy) {
  position: absolute;
  top: 0.4rem;
  right: 0.5rem;
  background: #ffffff;
  border: 1px solid #e9e9e7;
  color: #787774;
  padding: 0.2rem 0.6rem;
  font-size: 11px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  z-index: 10;
  box-shadow: 0 1px 2px rgba(0,0,0,0.05);
}

:deep(.markdown-entry .md-codeblock-copy:hover) {
  background: #f9f5ed;
  color: #37352f;
}

:deep(.markdown-entry .md-codeblock pre) {
  overflow-x: auto;
  border-radius: 12px;
  background: #f7f7f5;
  padding: 1rem;
  padding-top: 2.25rem;
  color: #37352f;
  margin: 0;
  border: 1px solid #e9e9e7;
}

:deep(.markdown-entry .md-codeblock code) {
  background: transparent;
  padding: 0;
  color: inherit;
  font-size: 0.92em;
  line-height: 1.7;
}

:deep(.markdown-entry .md-codeblock pre code) {
  white-space: pre;
}

/* Highlight.js Theme (Atom One Light variant) */
:deep(.markdown-entry .hljs-comment),
:deep(.markdown-entry .hljs-quote) {
  color: #a0a1a7;
  font-style: italic;
}
:deep(.markdown-entry .hljs-doctag),
:deep(.markdown-entry .hljs-keyword),
:deep(.markdown-entry .hljs-formula) {
  color: #a626a4;
}
:deep(.markdown-entry .hljs-section),
:deep(.markdown-entry .hljs-name),
:deep(.markdown-entry .hljs-selector-tag),
:deep(.markdown-entry .hljs-deletion),
:deep(.markdown-entry .hljs-subst) {
  color: #e45649;
}
:deep(.markdown-entry .hljs-literal) {
  color: #0184bb;
}
:deep(.markdown-entry .hljs-string),
:deep(.markdown-entry .hljs-regexp),
:deep(.markdown-entry .hljs-addition),
:deep(.markdown-entry .hljs-attribute),
:deep(.markdown-entry .hljs-meta-string) {
  color: #50a14f;
}
:deep(.markdown-entry .hljs-built_in),
:deep(.markdown-entry .hljs-class .hljs-title) {
  color: #c18401;
}
:deep(.markdown-entry .hljs-attr),
:deep(.markdown-entry .hljs-variable),
:deep(.markdown-entry .hljs-template-variable),
:deep(.markdown-entry .hljs-type),
:deep(.markdown-entry .hljs-selector-class),
:deep(.markdown-entry .hljs-selector-attr),
:deep(.markdown-entry .hljs-selector-pseudo),
:deep(.markdown-entry .hljs-number) {
  color: #986801;
}
:deep(.markdown-entry .hljs-symbol),
:deep(.markdown-entry .hljs-bullet),
:deep(.markdown-entry .hljs-link),
:deep(.markdown-entry .hljs-meta),
:deep(.markdown-entry .hljs-selector-id),
:deep(.markdown-entry .hljs-title) {
  color: #4078f2;
}
:deep(.markdown-entry .hljs-emphasis) {
  font-style: italic;
}
:deep(.markdown-entry .hljs-strong) {
  font-weight: bold;
}
:deep(.markdown-entry .hljs-link) {
  text-decoration: underline;
}

:deep(.markdown-entry .md-codeblock pre::-webkit-scrollbar) {
  height: 8px;
}

:deep(.markdown-entry .md-codeblock pre::-webkit-scrollbar-thumb) {
  background: rgba(148, 163, 184, 0.35);
  border-radius: 999px;
}

:deep(.markdown-entry .md-codeblock pre::-webkit-scrollbar-thumb:hover) {
  background: rgba(148, 163, 184, 0.55);
}

:deep(.markdown-entry table) {
  display: block;
  width: 100%;
  overflow-x: auto;
  border-collapse: collapse;
  margin: 1rem 0;
}

:deep(.markdown-entry table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1.25rem 0;
  font-size: 0.95em;
}

:deep(.markdown-entry th) {
  background-color: #f7f7f5;
  font-weight: 600;
  text-align: left;
  padding: 0.75rem 1rem;
  border: 1px solid #e9e9e7;
  color: #5f5e58;
}

:deep(.markdown-entry td) {
  padding: 0.75rem 1rem;
  border: 1px solid #e9e9e7;
  color: #37352f;
}

:deep(.markdown-entry tr:nth-child(even)) {
  background-color: rgba(249, 245, 237, 0.4);
}

:deep(.markdown-entry img) {
  display: block;
  max-width: 100%;
  height: auto;
  border-radius: 10px;
  border: 1px solid #e9e9e7;
  margin: 1.25rem auto;
}

:deep(.markdown-entry th),
:deep(.markdown-entry td) {
  border: 1px solid #e9e9e7;
  padding: 0.5rem 0.75rem;
  text-align: left;
}

:deep(.markdown-entry th) {
  background: #f7f6f3;
  font-weight: 600;
}

:deep(.markdown-entry tbody tr:nth-child(odd)) {
  background: #fbfbfa;
}
</style>
