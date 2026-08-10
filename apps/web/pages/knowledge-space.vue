<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import KnowledgeTokenDialog from "../components/KnowledgeTokenDialog.vue";
import KnowledgeSidebar from "../components/KnowledgeSidebar.vue";
import SiteHeader from "../components/SiteHeader.vue";
import { useKnowledgeBase } from "../composables/useKnowledgeBase";
import { useKnowledgeSidebar } from "../composables/useKnowledgeSidebar";
import { renderMarkdown } from "../src/markdown";
import { resolveUploadsUrl } from "../src/runtimeConfig";

const route = useRoute();
const router = useRouter();
const {
  getEntriesBySpace,
  getSpaceAccessState,
  getSpaceBySlug,
  isSpaceUnlocked,
  resolveKnowledgeAssetUrl,
  loadKnowledgeSpaceBySlug,
  unlockSpace,
  lockSpace
} = useKnowledgeBase();
const { openSidebarDrawer } = useKnowledgeSidebar();

const spaceSlug = computed(() => String(route.params.spaceSlug ?? ""));
const tokenDialogOpen = ref(false);
const tokenError = ref("");
const pageLoading = ref(true);
const pageReady = ref(false);
const pageError = ref("");
const desktopViewportWidth = ref(0);

const space = computed(() => getSpaceBySlug(spaceSlug.value));
const entries = computed(() => getEntriesBySpace(spaceSlug.value));
const groupedEntries = computed(() => {
  const groups = new Map<string, typeof entries.value>();

  entries.value.forEach((entry) => {
    const current = groups.get(entry.sectionName) ?? [];
    current.push(entry);
    groups.set(entry.sectionName, current);
  });

  return Array.from(groups.entries()).map(([sectionName, records]) => ({
    sectionName,
    records
  }));
});
const unlocked = computed(() => isSpaceUnlocked(spaceSlug.value));
const accessState = computed(() => getSpaceAccessState(spaceSlug.value));
const renderedIntro = computed(() =>
  space.value
    ? renderMarkdown(space.value.introMarkdown, {
        imageResolver: (src) => resolveKnowledgeAssetUrl(space.value!.slug, src)
      })
    : ""
);
const coverStyle = computed(() => {
  if (!space.value?.coverUrl) {
    return {
      backgroundImage:
        "linear-gradient(135deg, #eef2f6 0%, #f8fafc 45%, #ffffff 100%)"
    };
  }

  return {
    backgroundImage: `linear-gradient(rgba(255,255,255,0.12), rgba(255,255,255,0.02)), url('${resolveUploadsUrl(space.value.coverUrl)}')`,
    backgroundSize: "cover",
    backgroundPosition: "center"
  };
});

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

function updateDesktopViewportWidth() {
  if (typeof window === "undefined") {
    return;
  }
  desktopViewportWidth.value = window.innerWidth;
}

watch(
  spaceSlug,
  async (nextSpaceSlug) => {
    if (!nextSpaceSlug) {
      router.replace("/knowledge");
      return;
    }

    pageLoading.value = true;
    pageReady.value = false;
    pageError.value = "";

    try {
      const payload = await loadKnowledgeSpaceBySlug(nextSpaceSlug, true);
      if (!payload?.space) {
        router.replace("/knowledge");
        return;
      }

      document.title = `VOIDLAB | ${payload.space.title}`;
      pageReady.value = true;
    } catch (error: unknown) {
      pageError.value = error instanceof Error ? error.message : "加载知识空间失败";
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

function handleUnlock(token: string) {
  if (!space.value) {
    return;
  }

  void unlockSpace(space.value.slug, token)
    .then((result) => {
      if (!result.success) {
        tokenError.value = result.message;
        return;
      }

      tokenDialogOpen.value = false;
      tokenError.value = "";
    })
    .catch((error: unknown) => {
      tokenError.value = error instanceof Error ? error.message : "令牌验证失败";
    });
}

function handleRetry() {
  if (!spaceSlug.value) {
    return;
  }

  void loadKnowledgeSpaceBySlug(spaceSlug.value, true)
    .then((payload) => {
      if (!payload?.space) {
        router.replace("/knowledge");
        return;
      }

      document.title = `VOIDLAB | ${payload.space.title}`;
      pageReady.value = true;
      pageError.value = "";
    })
    .catch((error: unknown) => {
      pageError.value = error instanceof Error ? error.message : "加载知识空间失败";
    })
    .finally(() => {
      pageLoading.value = false;
    });
}

function handleLockSpace() {
  if (!space.value) {
    return;
  }

  lockSpace(space.value.slug);
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

onMounted(() => {
  updateDesktopViewportWidth();
  window.addEventListener("resize", updateDesktopViewportWidth);
});

onBeforeUnmount(() => {
  if (typeof window === "undefined") {
    return;
  }
  window.removeEventListener("resize", updateDesktopViewportWidth);
});
</script>

<template>
  <div class="relative z-10 min-h-screen bg-[#f6f5f1] text-[#333333] font-sans selection:bg-[#cce2ff] pt-16 flex flex-col">
    <SiteHeader theme="light" activePath="/knowledge" />

    <!-- Mobile: follow the reference screenshot -->
    <div v-if="space && pageReady" class="md:hidden sticky top-16 z-20 bg-[#f6f5f1]/95 backdrop-blur-md border-b border-[#eaeaea]">
      <div class="mx-auto w-full max-w-4xl px-6 py-4 flex items-center justify-between text-[14px] text-[#787774]">
        <RouterLink to="/knowledge" class="flex items-center gap-3 hover:text-[#37352f] transition-colors">
          <span class="text-[18px] leading-none">‹</span>
          <span class="text-[14px] font-medium">回到知识库</span>
        </RouterLink>
        <button
          type="button"
          class="inline-flex items-center gap-2 text-[14px] font-medium text-[#37352f] transition-colors hover:text-black"
          @click="openSidebarDrawer"
        >
          目录 <span class="text-[#9ca3af]">›</span>
        </button>
      </div>
    </div>

    <div v-if="pageLoading" class="mx-auto flex min-h-[50vh] w-full max-w-5xl items-center justify-center px-6 text-[#787774]">
      正在加载知识空间...
    </div>

    <div v-else-if="pageError" class="mx-auto flex min-h-[50vh] w-full max-w-3xl flex-col items-center justify-center gap-4 px-6 text-center">
      <div class="text-4xl">⚠️</div>
      <div>
        <h1 class="text-2xl font-semibold text-[#37352f]">知识空间暂时打不开</h1>
        <p class="mt-2 text-[15px] text-[#787774]">{{ pageError }}</p>
      </div>
      <button
        type="button"
        class="rounded border border-[#e9e9e7] bg-transparent px-4 py-2 text-[14px] font-medium text-[#37352f] transition-colors hover:bg-black/5"
        @click="handleRetry"
      >
        重试
      </button>
    </div>

    <div
      v-else-if="space && pageReady"
      class="relative flex w-full flex-1 mx-auto max-w-[1440px] px-4 md:px-8 xl:px-12"
    >
      <div class="hidden md:block w-64 flex-shrink-0"></div>
      <KnowledgeSidebar embedded :desktop-fixed-style="desktopSidebarStyle" />

      <main class="flex-1 min-w-0 flex flex-col">
        <!-- Notion-like Cover -->
        <div
          class="h-48 md:h-64 w-full bg-gradient-to-br flex-shrink-0"
          :class="space.coverUrl ? [] : space.themeTint"
          :style="coverStyle"
        ></div>

        <div class="w-full max-w-[980px] px-6 pb-24 md:px-8 xl:px-10">
          <!-- Icon and Title -->
          <div class="relative -mt-12 md:-mt-16 mb-6">
            <div class="inline-flex h-24 w-24 md:h-32 md:w-32 items-center justify-center rounded text-6xl md:text-7xl">
              {{ space.icon }}
            </div>
          </div>

          <h1 class="text-4xl font-black tracking-tight text-[#111111] md:text-5xl relative pb-2 inline-block after:content-[''] after:absolute after:left-0 after:bottom-0 after:w-[1.5em] after:h-[5px] after:bg-[#c4f000]" style="font-family: 'Arial Black', Impact, Inter, 'Heiti SC', 'Microsoft YaHei', sans-serif;">
            {{ space.title }}
          </h1>

          <div class="mt-4 flex flex-wrap gap-4 text-[14px] text-[#787774] border-b border-[#e9e9e7] pb-6">
            <span>最后更新: {{ space.lastUpdatedAt }}</span>
            <span>·</span>
            <span class="font-medium" :class="unlocked ? 'text-[#0f7b6c]' : 'text-[#d97706]'">
              {{
                accessState === "public"
                  ? "全部公开"
                  : unlocked
                    ? "已解锁正文"
                    : "正文需令牌解锁"
              }}
            </span>
          </div>

          <!-- Overview / Intro -->
          <div class="mt-8 markdown-knowledge text-[16px] leading-relaxed text-[#37352f]" v-html="renderedIntro"></div>

          <!-- Unlock Banner -->
          <div class="mt-8 mb-12 flex flex-col sm:flex-row sm:items-center justify-between gap-4 rounded bg-[#f7f7f5] px-5 py-4 border border-[#e9e9e7]">
            <div class="text-[14px]">
              <span class="font-semibold">
                {{
                  accessState === "public"
                    ? "🌍 Space 公开访问"
                    : unlocked
                      ? "🎉 Space 已解锁"
                      : "🔒 需要访问令牌"
                }}
              </span>
              <span class="ml-2 text-[#787774]">
                {{
                  accessState === "public"
                    ? "这个 Space 的目录和正文都可以直接阅读。"
                    : unlocked
                      ? "你可以连续阅读全部正文内容。"
                      : space.tokenHint
                }}
              </span>
            </div>
            <button
              v-if="accessState !== 'public'"
              type="button"
              class="text-[14px] font-medium transition-colors whitespace-nowrap"
              :class="unlocked ? 'hover:text-[#eb5757] text-[#787774]' : 'hover:text-[#0f7b6c] text-[#37352f]'"
              @click="unlocked ? handleLockSpace() : openUnlockDialog()"
            >
              {{ unlocked ? "清除令牌" : "输入令牌" }}
            </button>
          </div>

          <!-- Directory / Database List -->
          <div class="space-y-10">
            <div v-for="group in groupedEntries" :key="group.sectionName">
              <h3 class="mb-3 text-[18px] font-semibold text-[#37352f]">{{ group.sectionName }}</h3>
              
              <div class="flex flex-col border-t border-[#e9e9e7]">
                <RouterLink
                  v-for="entry in group.records"
                  :key="entry.slug"
                  :to="`/knowledge/${space.slug}/${entry.slug}`"
                  class="group flex items-center justify-between border-b border-[#e9e9e7] py-3 transition-colors hover:bg-[#f7f7f5]"
                >
                  <div class="flex items-center gap-3">
                    <span class="text-xl">📄</span>
                    <span class="text-[15px] font-medium text-[#37352f] group-hover:underline underline-offset-4 decoration-[#e9e9e7]">
                      {{ entry.title }}
                    </span>
                  </div>
                  
                  <div class="flex items-center gap-3 text-[13px]">
                    <span class="text-[#9ca3af] hidden sm:inline">{{ entry.estimatedReadMinutes }} min read</span>
                    <span
                      class="rounded px-2 py-0.5"
                      :class="
                        entry.isPreview
                          ? 'bg-[#e0f2fe] text-[#1e3a8a]'
                          : accessState === 'public'
                            ? 'bg-[#ede9fe] text-[#5b21b6]'
                            : unlocked
                              ? 'bg-[#dcfce7] text-[#14532d]'
                              : 'bg-[#ffedd5] text-[#9a3412]'
                      "
                    >
                      {{ entry.isPreview ? "Preview" : accessState === "public" ? "Public" : unlocked ? "Unlocked" : "Locked" }}
                    </span>
                  </div>
                </RouterLink>
              </div>
            </div>
          </div>
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
:deep(.markdown-knowledge) {
  font-size: 16px;
  line-height: 1.7;
  color: #37352f;
}

:deep(.markdown-knowledge h1),
:deep(.markdown-knowledge h2),
:deep(.markdown-knowledge h3) {
  margin-top: 2rem;
  margin-bottom: 1rem;
  color: #37352f;
  font-weight: 900;
  font-family: "Arial Black", "Impact", "Inter", "Heiti SC", "Microsoft YaHei", sans-serif;
  letter-spacing: -0.02em;
}

:deep(.markdown-knowledge h1) { font-size: 2.25rem; }
:deep(.markdown-knowledge h2) { font-size: 1.5rem; }
:deep(.markdown-knowledge h3) { font-size: 1.25rem; }

:deep(.markdown-knowledge h4),
:deep(.markdown-knowledge h5),
:deep(.markdown-knowledge h6) {
  margin-top: 1.5rem;
  margin-bottom: 0.75rem;
  color: #333333;
  font-weight: 800;
  font-family: "Arial Black", "Impact", "Inter", "Heiti SC", "Microsoft YaHei", sans-serif;
  letter-spacing: -0.01em;
}

:deep(.markdown-knowledge h4) { font-size: 1.15rem; }
:deep(.markdown-knowledge h5) { font-size: 1.05rem; }
:deep(.markdown-knowledge h6) { font-size: 0.95rem; color: #5f5e58; }

:deep(.markdown-knowledge h1) {
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

:deep(.markdown-knowledge h1::after) {
  content: "";
  position: absolute;
  left: 0;
  bottom: 0;
  width: 1.5em;
  height: 5px;
  background-color: #c4f000;
}

:deep(.markdown-knowledge h2) {
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

:deep(.markdown-knowledge h3) {
  position: relative;
  padding-left: 1.25rem;
  margin-top: 2rem;
  margin-bottom: 1rem;
  color: #333333;
  font-weight: 800;
  font-family: "Arial Black", "Impact", "Inter", "Heiti SC", "Microsoft YaHei", sans-serif;
  letter-spacing: -0.01em;
}

:deep(.markdown-knowledge h2::before),
:deep(.markdown-knowledge h3::before) {
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

:deep(.markdown-knowledge p) {
  color: #37352f;
  line-height: 1.6;
  margin-bottom: 0.5rem;
}

:deep(.markdown-knowledge li) {
  color: #37352f;
  line-height: 1.65;
  margin: 0.35rem 0;
  padding-left: 0.15rem;
}

:deep(.markdown-knowledge li > p) {
  margin: 0;
  display: inline;
}

:deep(.markdown-knowledge a) {
  color: #0f7b6c;
  text-decoration: underline;
  text-decoration-thickness: 1px;
  text-underline-offset: 2px;
}

:deep(.markdown-knowledge a:hover) {
  color: #0b5f54;
}

:deep(.markdown-knowledge ul),
:deep(.markdown-knowledge ol) {
  padding-left: 1.25rem;
  list-style-position: outside;
  margin: 0.25rem 0 0.5rem 0;
}

:deep(.markdown-knowledge ul) { list-style-type: disc; }
:deep(.markdown-knowledge ol) { list-style-type: decimal; }

:deep(.markdown-knowledge ul ul) { list-style-type: circle; }
:deep(.markdown-knowledge ul ul ul) { list-style-type: square; }

:deep(.markdown-knowledge ul ul),
:deep(.markdown-knowledge ul ol),
:deep(.markdown-knowledge ol ul),
:deep(.markdown-knowledge ol ol) {
  margin-top: 0.25rem;
  margin-bottom: 0.25rem;
}

:deep(.markdown-knowledge blockquote) {
  margin: 1rem 0;
  border-left: 3px solid #d0d0ca;
  padding: 0.2rem 1rem;
  color: #787774;
  background: #f7f7f5;
  border-radius: 8px;
}

:deep(.markdown-knowledge code) {
  border-radius: 3px;
  background: #f1f1ef;
  padding: 0.2rem 0.4rem;
  color: #eb5757;
  font-size: 0.85em;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}

:deep(.markdown-knowledge .md-codeblock) {
  margin: 1rem 0 1.25rem 0;
  position: relative;
}

:deep(.markdown-knowledge .md-codeblock[data-lang]::before) {
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

:deep(.markdown-knowledge .md-codeblock-copy) {
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

:deep(.markdown-knowledge .md-codeblock-copy:hover) {
  background: #f9f5ed;
  color: #37352f;
}

:deep(.markdown-knowledge .md-codeblock pre) {
  overflow-x: auto;
  border-radius: 12px;
  background: #f7f7f5;
  padding: 1rem;
  padding-top: 2.25rem;
  color: #37352f;
  margin: 0;
  border: 1px solid #e9e9e7;
}

:deep(.markdown-knowledge .md-codeblock code) {
  background: transparent;
  padding: 0;
  color: inherit;
  font-size: 0.92em;
  line-height: 1.7;
}

:deep(.markdown-knowledge .md-codeblock pre code) {
  white-space: pre;
}

/* Highlight.js Theme (Atom One Light variant) */
:deep(.markdown-knowledge .hljs-comment),
:deep(.markdown-knowledge .hljs-quote) {
  color: #a0a1a7;
  font-style: italic;
}
:deep(.markdown-knowledge .hljs-doctag),
:deep(.markdown-knowledge .hljs-keyword),
:deep(.markdown-knowledge .hljs-formula) {
  color: #a626a4;
}
:deep(.markdown-knowledge .hljs-section),
:deep(.markdown-knowledge .hljs-name),
:deep(.markdown-knowledge .hljs-selector-tag),
:deep(.markdown-knowledge .hljs-deletion),
:deep(.markdown-knowledge .hljs-subst) {
  color: #e45649;
}
:deep(.markdown-knowledge .hljs-literal) {
  color: #0184bb;
}
:deep(.markdown-knowledge .hljs-string),
:deep(.markdown-knowledge .hljs-regexp),
:deep(.markdown-knowledge .hljs-addition),
:deep(.markdown-knowledge .hljs-attribute),
:deep(.markdown-knowledge .hljs-meta-string) {
  color: #50a14f;
}
:deep(.markdown-knowledge .hljs-built_in),
:deep(.markdown-knowledge .hljs-class .hljs-title) {
  color: #c18401;
}
:deep(.markdown-knowledge .hljs-attr),
:deep(.markdown-knowledge .hljs-variable),
:deep(.markdown-knowledge .hljs-template-variable),
:deep(.markdown-knowledge .hljs-type),
:deep(.markdown-knowledge .hljs-selector-class),
:deep(.markdown-knowledge .hljs-selector-attr),
:deep(.markdown-knowledge .hljs-selector-pseudo),
:deep(.markdown-knowledge .hljs-number) {
  color: #986801;
}
:deep(.markdown-knowledge .hljs-symbol),
:deep(.markdown-knowledge .hljs-bullet),
:deep(.markdown-knowledge .hljs-link),
:deep(.markdown-knowledge .hljs-meta),
:deep(.markdown-knowledge .hljs-selector-id),
:deep(.markdown-knowledge .hljs-title) {
  color: #4078f2;
}
:deep(.markdown-knowledge .hljs-emphasis) {
  font-style: italic;
}
:deep(.markdown-knowledge .hljs-strong) {
  font-weight: bold;
}
:deep(.markdown-knowledge .hljs-link) {
  text-decoration: underline;
}

:deep(.markdown-knowledge table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1.25rem 0;
  font-size: 0.95em;
}

:deep(.markdown-knowledge th) {
  background-color: #f7f7f5;
  font-weight: 600;
  text-align: left;
  padding: 0.75rem 1rem;
  border: 1px solid #e9e9e7;
  color: #5f5e58;
}

:deep(.markdown-knowledge td) {
  padding: 0.75rem 1rem;
  border: 1px solid #e9e9e7;
  color: #37352f;
}

:deep(.markdown-knowledge tr:nth-child(even)) {
  background-color: rgba(249, 245, 237, 0.4);
}

:deep(.markdown-knowledge img) {
  display: block;
  max-width: 100%;
  height: auto;
  border-radius: 10px;
  border: 1px solid #e9e9e7;
  margin: 1.25rem auto;
}
</style>
