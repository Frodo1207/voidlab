<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useInsightsFeed } from "../composables/useInsightsFeed";
import SiteHeader from "../components/SiteHeader.vue";
import { renderMarkdown } from "../src/markdown";

const route = useRoute();
const router = useRouter();
const { insights, loadInsightBySlug } = useInsightsFeed();

const slug = computed(() => String(route.params.slug ?? ""));
const insight = ref<Awaited<ReturnType<typeof loadInsightBySlug>>>(null);
const loading = ref(true);
const error = ref("");
const relatedArticles = computed(() => insights.value.filter((item) => item.slug !== slug.value && !item.isNews).slice(0, 6));
const renderedContent = computed(() => {
  if (!insight.value?.rawContent) {
    return "";
  }

  return renderMarkdown(insight.value.rawContent);
});

async function syncInsightDetail() {
  loading.value = true;
  error.value = "";

  try {
    const detail = await loadInsightBySlug(slug.value);
    if (!detail) {
      router.replace("/insights");
      return;
    }

    insight.value = detail;
    document.title = `VOID LAB | ${detail.title}`;
  } catch (loadError) {
    error.value = loadError instanceof Error ? loadError.message : "加载文章失败";
    router.replace("/insights");
  } finally {
    loading.value = false;
  }
}

watch(slug, () => {
  void syncInsightDetail();
}, { immediate: true });

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}
</script>

<template>
  <div class="relative z-10 min-h-screen bg-white text-[#111111] font-sans selection:bg-black selection:text-white pt-16">
    <!-- Header -->
    <SiteHeader theme="light" activePath="/insights" />

    <div v-if="loading" class="mx-auto max-w-[1440px] px-6 py-12 md:px-10 lg:px-14 text-center text-gray-400">
      正在加载文章详情...
    </div>

    <div v-else-if="insight" class="mx-auto max-w-[1440px] px-6 py-12 md:px-10 lg:px-14">
      <div class="grid grid-cols-1 lg:grid-cols-[1fr_380px] gap-12 lg:gap-24">
        
        <!-- Left Column: Main Article -->
        <article class="flex flex-col">
          <!-- Breadcrumbs -->
          <div class="text-sm text-gray-400 mb-8 font-medium">
            <RouterLink to="/insights" class="hover:text-black transition-colors">VOIDLAB</RouterLink> / 
            <span class="hover:text-black transition-colors cursor-pointer">{{ insight.category }}</span> / 
            <span class="text-black">文章详情</span>
          </div>

          <!-- Title -->
          <h1 class="text-4xl md:text-5xl lg:text-[56px] font-bold leading-[1.15] tracking-tight text-black">
            {{ insight.title }}
          </h1>

          <!-- Meta: Date & Socials -->
          <div class="mt-8 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div class="text-sm text-gray-500 font-medium">
              发布时间：{{ insight.publishedAt }}
            </div>
            <div class="flex items-center gap-4 text-black">
              <!-- X (Twitter) Icon -->
              <button class="hover:text-gray-500 transition-colors">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"></path></svg>
              </button>
              <!-- Telegram Icon -->
              <button class="hover:text-gray-500 transition-colors">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="M11.944 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0a12 12 0 0 0-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 0 1 .171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.895-1.056-.68-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.48.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.888-.662 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z"></path></svg>
              </button>
              <!-- Facebook Icon -->
              <button class="hover:text-gray-500 transition-colors">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.469h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.469h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"></path></svg>
              </button>
              <!-- Link Icon -->
              <button class="hover:text-gray-500 transition-colors">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path></svg>
              </button>
            </div>
          </div>

          <!-- Tags -->
          <div class="mt-8 flex flex-wrap gap-3">
            <span
              v-for="tag in insight.tags"
              :key="tag"
              class="border border-gray-200 rounded-full px-4 py-1.5 text-sm text-[#0b7f67] font-medium flex items-center gap-1 cursor-pointer hover:bg-gray-50 transition-colors"
            >
              {{ tag }}
              <span class="text-gray-400 font-normal ml-0.5">></span>
            </span>
          </div>

          <!-- Summary Block -->
          <div v-if="insight.summary" class="mt-10 rounded-[24px] border border-gray-200 bg-[#f7f7f5] px-6 py-6 md:px-8">
            <div class="text-xs font-semibold uppercase tracking-[0.24em] text-gray-400">
              Summary
            </div>
            <p class="mt-4 text-[17px] leading-8 text-[#222222]">
              {{ insight.summary }}
            </p>
          </div>

          <!-- Main Content -->
          <div v-if="renderedContent" class="markdown-body mt-12" v-html="renderedContent"></div>
          <div v-else class="mt-12 text-[17px] leading-8 text-gray-700">
            <p v-for="(paragraph, index) in insight.content" :key="index" class="mb-6">
              {{ paragraph }}
            </p>
          </div>

          <!-- End of Article Source -->
          <div class="mt-16 pt-8 border-t border-gray-100 flex items-center justify-between text-sm">
            <span class="text-gray-500">来源：<span class="font-medium text-black">{{ insight.sourceName }}</span></span>
            <a :href="insight.sourceUrl" target="_blank" class="text-blue-600 hover:underline">查看原文</a>
          </div>
        </article>

        <!-- Right Column: Sidebar -->
        <aside class="relative">
          <div class="sticky top-28">
            <h3 class="text-[22px] font-bold text-black mb-6">相关文章</h3>
            
            <div class="bg-[#f8f9fa] rounded-[24px] p-6 lg:p-8">
              <ul class="flex flex-col divide-y divide-gray-200/60">
                <li 
                  v-for="article in relatedArticles" 
                  :key="article.slug"
                  class="py-5 first:pt-0 last:pb-0 group cursor-pointer"
                >
                  <RouterLink :to="`/insights/${article.slug}`" class="flex gap-3">
                    <div class="mt-2 w-1.5 h-1.5 rounded-full bg-black shrink-0 group-hover:scale-150 transition-transform"></div>
                    <div>
                      <h4 class="text-[15px] font-semibold leading-[1.6] text-black group-hover:text-blue-600 transition-colors line-clamp-3">
                        {{ article.title }}
                      </h4>
                    </div>
                  </RouterLink>
                </li>
              </ul>
            </div>
          </div>
        </aside>

      </div>

      <!-- Footer -->
      <footer class="mt-24 border-t border-gray-100 py-8 text-sm text-gray-400">
        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>© 2026 VOIDLAB.AI / 资讯平台</div>
          <button type="button" class="text-left transition-colors hover:text-black md:text-right" @click="scrollToTop">
            回到顶部
          </button>
        </div>
      </footer>
    </div>

    <div v-else class="mx-auto max-w-[1440px] px-6 py-12 md:px-10 lg:px-14 text-center text-[#b42318]">
      {{ error || "文章不存在，正在返回资讯列表..." }}
    </div>
  </div>
</template>

<style scoped>
:deep(.markdown-body) {
  color: #252525;
  font-size: 17px;
  line-height: 1.95;
}

:deep(.markdown-body > :first-child) {
  margin-top: 0;
}

:deep(.markdown-body h1),
:deep(.markdown-body h2),
:deep(.markdown-body h3),
:deep(.markdown-body h4) {
  color: #111111;
  font-weight: 700;
  letter-spacing: -0.02em;
  line-height: 1.25;
  margin-top: 2.8rem;
  margin-bottom: 1rem;
}

:deep(.markdown-body h1) {
  font-size: 2rem;
}

:deep(.markdown-body h2) {
  font-size: 1.6rem;
}

:deep(.markdown-body h3) {
  font-size: 1.25rem;
}

:deep(.markdown-body p) {
  margin: 1rem 0;
  color: #3f3f46;
}

:deep(.markdown-body strong) {
  color: #111111;
  font-weight: 700;
}

:deep(.markdown-body a) {
  color: #2563eb;
  text-decoration: underline;
  text-underline-offset: 3px;
  word-break: break-word;
}

:deep(.markdown-body ul),
:deep(.markdown-body ol) {
  margin: 1.2rem 0;
  padding-left: 1.5rem;
  color: #3f3f46;
}

:deep(.markdown-body li) {
  margin: 0.45rem 0;
}

:deep(.markdown-body blockquote) {
  margin: 1.5rem 0;
  padding: 1rem 1.25rem;
  border-left: 4px solid #111111;
  background: #f7f7f5;
  color: #52525b;
  border-radius: 0 16px 16px 0;
}

:deep(.markdown-body hr) {
  margin: 2rem 0;
  border: none;
  border-top: 1px solid #e5e7eb;
}

:deep(.markdown-body code) {
  background: #f3f4f6;
  color: #111111;
  padding: 0.18rem 0.4rem;
  border-radius: 0.4rem;
  font-size: 0.9em;
}

:deep(.markdown-body pre) {
  margin: 1.5rem 0;
  padding: 1rem 1.1rem;
  background: #111111;
  color: #f8fafc;
  border-radius: 1rem;
  overflow-x: auto;
}

:deep(.markdown-body pre code) {
  background: transparent;
  color: inherit;
  padding: 0;
}

:deep(.markdown-body img) {
  width: 100%;
  border-radius: 1rem;
  margin: 1.5rem 0;
}

:deep(.markdown-body table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1.5rem 0;
  overflow: hidden;
  display: block;
}

:deep(.markdown-body th),
:deep(.markdown-body td) {
  border: 1px solid #e5e7eb;
  padding: 0.8rem 0.9rem;
  text-align: left;
}

:deep(.markdown-body th) {
  background: #f8fafc;
  color: #111111;
  font-weight: 600;
}
</style>
