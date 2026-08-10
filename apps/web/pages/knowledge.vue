<script setup lang="ts">
import { onMounted } from "vue";
import SiteHeader from "../components/SiteHeader.vue";
import { useKnowledgeBase } from "../composables/useKnowledgeBase";
import { resolveUploadsUrl } from "../src/runtimeConfig";

const { spaces, getSpaceAccessState, getSpaceStats } = useKnowledgeBase();

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function spaceCardCoverStyle(coverUrl: string) {
  const resolved = resolveUploadsUrl(coverUrl || "");
  if (!resolved) {
    return {
      backgroundImage:
        "linear-gradient(135deg, #eef2f6 0%, #f8fafc 45%, #ffffff 100%)"
    };
  }

  return {
    backgroundImage: `linear-gradient(rgba(255,255,255,0.08), rgba(255,255,255,0.18)), url('${resolved}')`,
    backgroundSize: "cover",
    backgroundPosition: "center"
  };
}

onMounted(() => {
  document.title = "VOIDLAB | 知识库";
});
</script>

<template>
  <div class="relative z-10 min-h-screen bg-[#f6f5f1] text-[#333333] font-sans selection:bg-[#cce2ff] pt-16">
    <SiteHeader theme="light" activePath="/knowledge" />

    <main class="mx-auto max-w-5xl px-6 py-16 md:px-10 lg:py-24">
      <div class="mb-12">
        <h1 class="text-4xl font-black tracking-tight text-[#111111] md:text-5xl relative pb-2 inline-block after:content-[''] after:absolute after:left-0 after:bottom-0 after:w-[1.5em] after:h-[5px] after:bg-[#c4f000]" style="font-family: 'Arial Black', Impact, Inter, 'Heiti SC', 'Microsoft YaHei', sans-serif;">
          Knowledge Base
        </h1>
      </div>

      <div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
        <RouterLink
          v-for="space in spaces"
          :key="space.slug"
          :to="`/knowledge/${space.slug}`"
          class="group flex flex-col overflow-hidden rounded-lg border border-[#eaeaea] bg-white transition-all hover:shadow-[0_4px_12px_rgba(0,0,0,0.05)]"
        >
          <!-- Cover -->
          <div
            class="h-28 w-full bg-gradient-to-br"
            :class="space.coverUrl ? [] : space.themeTint"
            :style="spaceCardCoverStyle(space.coverUrl)"
          ></div>
          
          <div class="relative flex flex-col p-4 pb-6">
            <!-- Icon -->
            <div class="absolute -top-10 flex h-14 w-14 items-center justify-center rounded bg-white text-3xl shadow-sm border border-[#eaeaea]">
              {{ space.icon }}
            </div>
            
            <div class="mt-6 flex items-center justify-between">
              <h3 class="text-[17px] font-black tracking-tight text-[#111111]" style="font-family: 'Arial Black', Impact, Inter, 'Heiti SC', 'Microsoft YaHei', sans-serif;">{{ space.title }}</h3>
            </div>
            <p class="mt-1 text-sm text-[#777777] line-clamp-2">
              {{ space.description }}
            </p>
            
            <div class="mt-4 flex items-center gap-3 text-[13px] text-[#999999]">
              <div class="flex items-center gap-1.5">
                <span class="inline-block h-4 w-4 text-center">📄</span>
                <span>{{ space.entryCount }} 篇内容</span>
              </div>
              <span>·</span>
              <span
                :class="
                  getSpaceAccessState(space.slug) === 'public'
                    ? 'text-[#6d28d9]'
                    : getSpaceAccessState(space.slug) === 'unlocked'
                      ? 'text-[#0f7b6c]'
                      : 'text-[#d97706]'
                "
              >
                {{
                  getSpaceAccessState(space.slug) === "public"
                    ? "公开"
                    : getSpaceAccessState(space.slug) === "unlocked"
                      ? "已解锁"
                      : "需令牌"
                }}
              </span>
            </div>
          </div>
        </RouterLink>
      </div>
    </main>

    <footer class="border-t border-[#eaeaea] py-8 text-sm text-[#999999]">
      <div class="mx-auto flex max-w-5xl flex-col gap-4 px-6 md:flex-row md:items-center md:justify-between md:px-10">
        <div>© 2026 VOIDLAB.AI</div>
        <button type="button" class="text-left transition-colors hover:text-[#111111] md:text-right" @click="scrollToTop">
          回到顶部
        </button>
      </div>
    </footer>
  </div>
</template>
