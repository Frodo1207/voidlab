<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useKnowledgeBase } from "../composables/useKnowledgeBase";
import { useKnowledgeSidebar } from "../composables/useKnowledgeSidebar";

const props = withDefaults(
  defineProps<{
    embedded?: boolean;
    desktopFixedStyle?: Record<string, string | number>;
  }>(),
  {
    embedded: false,
    desktopFixedStyle: undefined
  }
);

const route = useRoute();
const { getSpaceBySlug, getEntriesBySpace, isSpaceUnlocked, loadKnowledgeSpaceBySlug } = useKnowledgeBase();
const { sidebarCollapsed, toggleSidebar, sidebarDrawerOpen, closeSidebarDrawer } = useKnowledgeSidebar();

const spaceSlug = computed(() => String(route.params.spaceSlug ?? ""));
const space = computed(() => getSpaceBySlug(spaceSlug.value));
const entries = computed(() => getEntriesBySpace(spaceSlug.value));

const collapsedSections = ref<Record<string, boolean>>({});

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

function isSectionOpen(sectionName: string) {
  return collapsedSections.value[sectionName] !== true;
}

function toggleSection(sectionName: string) {
  collapsedSections.value = {
    ...collapsedSections.value,
    [sectionName]: !isSectionOpen(sectionName)
  };
}

const desktopCollapsed = computed(() => (props.embedded ? false : sidebarCollapsed.value));

let originalBodyOverflow = "";
let originalBodyPosition = "";
let originalBodyTop = "";
let originalBodyWidth = "";
let originalHtmlOverflow = "";
let lockedScrollY = 0;
watch(
  sidebarDrawerOpen,
  (open) => {
    if (typeof document === "undefined" || typeof window === "undefined") {
      return;
    }
    if (open) {
      lockedScrollY = window.scrollY;
      originalBodyOverflow = document.body.style.overflow;
      originalBodyPosition = document.body.style.position;
      originalBodyTop = document.body.style.top;
      originalBodyWidth = document.body.style.width;
      originalHtmlOverflow = document.documentElement.style.overflow;
      document.documentElement.style.overflow = "hidden";
      document.body.style.overflow = "hidden";
      document.body.style.position = "fixed";
      document.body.style.top = `-${lockedScrollY}px`;
      document.body.style.width = "100%";
      return;
    }
    document.documentElement.style.overflow = originalHtmlOverflow;
    document.body.style.overflow = originalBodyOverflow;
    document.body.style.position = originalBodyPosition;
    document.body.style.top = originalBodyTop;
    document.body.style.width = originalBodyWidth;
    window.scrollTo(0, lockedScrollY);
  },
  { immediate: true }
);

watch(
  () => route.fullPath,
  () => {
    // 路由切换时自动收起移动端目录抽屉，保持体验一致
    if (sidebarDrawerOpen.value) {
      closeSidebarDrawer();
    }
  }
);

onBeforeUnmount(() => {
  if (typeof document === "undefined") {
    return;
  }
  document.documentElement.style.overflow = originalHtmlOverflow;
  document.body.style.overflow = originalBodyOverflow;
  document.body.style.position = originalBodyPosition;
  document.body.style.top = originalBodyTop;
  document.body.style.width = originalBodyWidth;
});

watch(
  spaceSlug,
  (nextSpaceSlug) => {
    if (!nextSpaceSlug) {
      return;
    }

    void loadKnowledgeSpaceBySlug(nextSpaceSlug);
  },
  { immediate: true }
);
</script>

<template>
  <!-- Mobile: directory drawer (overlay) -->
  <div
    v-if="sidebarDrawerOpen"
    class="fixed inset-0 z-[70] md:hidden"
    :style="{
      height: 'calc(var(--vvh, 1vh) * 100)',
      minHeight: '100vh'
    }"
    aria-modal="true"
    role="dialog"
  >
    <div class="absolute inset-0 bg-black/35" @click="closeSidebarDrawer"></div>
    <transition
      appear
      enter-active-class="transition-transform duration-200 ease-out"
      enter-from-class="-translate-x-full"
      enter-to-class="translate-x-0"
      leave-active-class="transition-transform duration-200 ease-in"
      leave-from-class="translate-x-0"
      leave-to-class="-translate-x-full"
    >
      <aside
        v-if="space"
        class="absolute left-0 top-0 bottom-0 h-full w-[86vw] max-w-[360px] overflow-y-auto overscroll-contain border-r border-[#eaeaea] bg-[#f6f5f1]"
        :style="{
          WebkitOverflowScrolling: 'touch',
          paddingTop: 'env(safe-area-inset-top)',
          paddingBottom: 'env(safe-area-inset-bottom)'
        }"
      >
        <div class="sticky top-0 z-10 flex items-center justify-between border-b border-[#eaeaea] bg-[#f6f5f1] px-4 py-3">
          <div class="min-w-0">
            <div class="text-[12px] font-medium tracking-[0.08em] text-[#999999] uppercase">目录</div>
            <div class="mt-1 flex items-center gap-2 min-w-0">
              <span class="text-xl flex-shrink-0">{{ space.icon }}</span>
              <span class="truncate text-[15px] font-bold text-[#333333]">{{ space.title }}</span>
            </div>
          </div>
          <button
            type="button"
            class="inline-flex h-9 w-9 items-center justify-center rounded border border-[#eaeaea] bg-white text-[#777777] transition-colors hover:bg-[#f0f0f0] hover:text-[#333333]"
            title="关闭目录"
            @click="closeSidebarDrawer"
          >
            ✕
          </button>
        </div>

        <div class="px-4 py-4 space-y-6">
          <div v-for="group in groupedEntries" :key="group.sectionName">
            <button
              type="button"
              class="flex w-full items-center justify-between rounded px-3 py-2.5 text-left text-[15px] font-bold text-[#333333] hover:bg-[#f0f0f0]"
              @click="toggleSection(group.sectionName)"
            >
              <span class="truncate" :title="group.sectionName">{{ group.sectionName }}</span>
              <span class="text-[#999999]">{{ isSectionOpen(group.sectionName) ? "˅" : "›" }}</span>
            </button>
            <div v-show="isSectionOpen(group.sectionName)" class="mt-1 flex flex-col gap-1">
              <RouterLink
                v-for="entry in group.records"
                :key="entry.slug"
                :to="`/knowledge/${space.slug}/${entry.slug}`"
                class="rounded-md px-3 py-2 text-[13px] text-[#555555] hover:bg-[#eaeaea] transition-colors flex items-center gap-2"
                active-class="bg-[#c4f000] text-[#111111] font-bold shadow-sm"
                :title="entry.title"
              >
                <span class="truncate">{{ entry.title }}</span>
              </RouterLink>
            </div>
          </div>
        </div>
      </aside>
    </transition>
  </div>

  <aside
    v-if="space"
    class="hidden flex-col overflow-y-auto transition-[width] duration-200 md:flex"
    :class="[
      props.embedded
        ? 'fixed top-16 z-20 h-[calc(100vh-4rem)] bg-[#f6f5f1]'
        : 'fixed left-0 top-16 bottom-0 z-30 border-r border-[#eaeaea] bg-[#f6f5f1]',
      desktopCollapsed ? 'w-16' : 'w-64'
    ]"
    :style="props.embedded ? props.desktopFixedStyle : undefined"
  >
    <div
      v-if="!props.embedded"
      class="sticky top-0 z-10 flex items-center justify-end border-b px-3 py-2"
      :class="props.embedded ? 'border-transparent bg-[#f6f5f1]' : 'border-[#eaeaea] bg-[#f6f5f1]'"
    >
      <button
        type="button"
        class="inline-flex h-8 w-8 items-center justify-center rounded border border-[#eaeaea] bg-white text-[#777777] transition-colors hover:bg-[#f0f0f0] hover:text-[#333333]"
        :title="sidebarCollapsed ? '展开目录' : '收起目录'"
        @click="toggleSidebar"
      >
        <span class="text-[16px]">{{ sidebarCollapsed ? "›" : "‹" }}</span>
      </button>
    </div>

    <!-- Space Info -->
    <RouterLink
      :to="`/knowledge/${space.slug}`"
      class="flex items-center gap-3 border-b transition-colors hover:bg-white/40"
      :class="[props.embedded ? 'border-transparent' : 'border-[#eaeaea]', desktopCollapsed ? 'justify-center px-2 py-4' : props.embedded ? 'px-4 py-5' : 'p-4']"
      :title="space.title"
    >
      <div class="text-2xl flex-shrink-0">{{ space.icon }}</div>
      <div v-if="!desktopCollapsed" class="min-w-0 flex-1">
        <div class="text-[15px] font-bold text-[#333333] leading-tight truncate">{{ space.title }}</div>
        <div class="text-[12px] mt-1 flex items-center gap-1.5" :class="isSpaceUnlocked(space.slug) ? 'text-[#0f7b6c]' : 'text-[#d97706]'">
          <span class="inline-block w-1.5 h-1.5 rounded-full" :class="isSpaceUnlocked(space.slug) ? 'bg-[#0f7b6c]' : 'bg-[#d97706]'"></span>
          {{ isSpaceUnlocked(space.slug) ? '已解锁' : '需令牌' }}
        </div>
      </div>
    </RouterLink>

    <!-- Directory -->
    <div class="space-y-5" :class="desktopCollapsed ? 'p-2' : props.embedded ? 'px-3 py-2' : 'p-4'">
      <div v-for="group in groupedEntries" :key="group.sectionName">
        <button
          v-if="!desktopCollapsed"
          type="button"
          class="flex w-full items-center justify-between rounded px-3 py-2 text-left text-[13px] font-bold text-[#444444] hover:bg-white/35"
          @click="toggleSection(group.sectionName)"
        >
          <span class="truncate" :title="group.sectionName">{{ group.sectionName }}</span>
          <span class="text-[#999999]">{{ isSectionOpen(group.sectionName) ? "˅" : "›" }}</span>
        </button>

        <div v-show="desktopCollapsed || isSectionOpen(group.sectionName)" class="mt-1 flex flex-col gap-1">
          <RouterLink
            v-for="entry in group.records"
            :key="entry.slug"
            :to="`/knowledge/${space.slug}/${entry.slug}`"
            class="flex items-center gap-2 rounded-md text-[12px] leading-5 text-[#5c5c58] transition-colors hover:bg-white/45"
            :class="desktopCollapsed ? 'justify-center px-2 py-2 text-center' : 'px-3 py-2'"
            active-class="bg-[#c4f000] text-[#111111] font-bold shadow-sm"
            :title="entry.title"
          >
            <span v-if="desktopCollapsed" class="text-[14px] font-bold">{{ entry.title.charAt(0) }}</span>
            <span v-else class="truncate">{{ entry.title }}</span>
          </RouterLink>
        </div>
      </div>
    </div>
  </aside>
</template>
