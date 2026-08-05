<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import {
  loadEvents,
  useEventArchive,
  type EventStatus,
  type EventType
} from "../composables/useEventArchive";
import SiteHeader from "../components/SiteHeader.vue";

const { events, loading, error } = useEventArchive();

const selectedStatus = ref<"all" | EventStatus>("all");
const selectedType = ref<"all" | EventType>("all");
const selectedCity = ref<"all" | string>("all");
const availableOnly = ref(false);

const filteredEvents = computed(() =>
  events.value.filter((event) => {
    const matchStatus = selectedStatus.value === "all" || event.status === selectedStatus.value;
    const matchType = selectedType.value === "all" || event.type === selectedType.value;
    const matchCity = selectedCity.value === "all" || event.city === selectedCity.value;
    const matchAvailable = !availableOnly.value || event.status === "live" || event.status === "next";
    return matchStatus && matchType && matchCity && matchAvailable;
  })
);

const typeOptions = [
  { label: "全部类型", value: "all" as const },
  { label: "线下活动", value: "线下活动" as const },
  { label: "企业内训", value: "企业内训" as const },
  { label: "闭门研讨", value: "闭门研讨" as const },
  { label: "线上分享", value: "线上分享" as const }
];

const cityOptions = computed(() => [
  { label: "全部城市", value: "all" },
  ...Array.from(new Set(events.value.map((event) => event.city))).map((city: string) => ({
    label: city,
    value: city
  }))
]);

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

onMounted(() => {
  document.title = "活动 | VOIDLAB";
  void loadEvents();
});
</script>

<template>
  <div class="relative z-10 min-h-screen bg-[#f8f9fa] text-[#37352f] font-sans selection:bg-[#cce2ff] pt-16" style="background-image: radial-gradient(#d1d5db 1px, transparent 1px); background-size: 24px 24px;">
    
    <!-- Header -->
    <SiteHeader theme="light" activePath="/events" />

    <!-- Hero Section (Cobay Style with Banner) -->
    <section class="border-b border-gray-200">
      <div class="mx-auto max-w-[1440px] px-6 md:px-10 lg:px-14 py-20 lg:py-28 grid grid-cols-1 lg:grid-cols-[1fr_500px] gap-12 lg:gap-20 items-center">
        
        <!-- Left: Text -->
        <div>
          <h1 class="text-6xl md:text-7xl font-bold leading-[1.1] tracking-tight text-[#37352f]">
            VOIDLAB<br/>
            活动。
          </h1>
          <h2 class="mt-6 text-2xl md:text-3xl font-bold text-[#5a5955]">
            真实开发者的线下场域。
          </h2>
          <p class="mt-6 text-lg text-[#787774] leading-relaxed max-w-lg">
            VOIDLAB 活动由我们自己发起与组织，包含创始人聚会、工作坊和作品展示：让真正做事的人在线下分享项目、认识合作者，并把想法推进成可见成果。
          </p>
        </div>

        <!-- Right: Featured Card -->
        <div v-if="events.length > 0" class="relative rounded-2xl overflow-hidden border border-gray-200 bg-white shadow-sm group cursor-pointer" @click="$router.push(`/events/${events[0].slug}`)">
          <div class="h-[260px] w-full overflow-hidden">
            <img :src="events[0].cover" alt="Featured Event" class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-700" />
          </div>
          <div class="absolute inset-0 bg-gradient-to-t from-white via-white/80 to-transparent top-[100px]"></div>
          
          <div class="absolute bottom-0 left-0 right-0 p-6 flex flex-col justify-end">
            <h3 class="text-3xl font-bold text-[#37352f] mb-2">{{ events[0].city }}</h3>
            <div class="text-sm text-[#787774] mb-6 font-medium">
              {{ events[0].time.split(' ')[0] }} · {{ events[0].title }} · {{ events[0].city }}场
            </div>
            <button class="w-full py-3 rounded-lg border border-gray-300 text-sm font-bold text-[#37352f] group-hover:bg-gray-50 transition-colors flex items-center justify-center gap-2">
              查看活动
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="7" y1="17" x2="17" y2="7"></line><polyline points="7 7 17 7 17 17"></polyline></svg>
            </button>
          </div>
        </div>

      </div>
    </section>

    <!-- Main Content Area -->
    <main class="mx-auto max-w-[1440px] px-6 py-12 md:py-20 md:px-10 lg:px-14">
      
      <!-- Properties / Filters (Notion Board Style) -->
      <div class="border-t border-[#ededed] py-6 mb-8 flex flex-col gap-4">
        
        <div class="flex flex-wrap items-center gap-6 text-[14px]">
          <div class="flex items-center gap-2 w-[120px] text-[#787774]">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
            类型
          </div>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="option in typeOptions"
              :key="option.value"
              @click="selectedType = option.value"
              class="px-2.5 py-0.5 rounded text-[14px] transition-colors"
              :class="selectedType === option.value ? 'bg-[#f1f1ef] text-[#37352f] font-medium' : 'text-[#787774] hover:bg-[#f1f1ef] hover:text-[#37352f]'"
            >
              {{ option.label }}
            </button>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-6 text-[14px]">
          <div class="flex items-center gap-2 w-[120px] text-[#787774]">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path><circle cx="12" cy="10" r="3"></circle></svg>
            地点
          </div>
          <div class="flex flex-wrap gap-2">
            <button
              v-for="option in cityOptions"
              :key="option.value"
              @click="selectedCity = option.value"
              class="px-2.5 py-0.5 rounded text-[14px] transition-colors"
              :class="selectedCity === option.value ? 'bg-[#e9e5df] text-[#37352f] font-medium' : 'text-[#787774] hover:bg-[#e9e5df] hover:text-[#37352f]'"
            >
              {{ option.label }}
            </button>
          </div>
        </div>

      </div>

      <!-- Gallery View (Notion Style) -->
      <div v-if="loading && events.length === 0" class="rounded-lg border border-[#ededed] bg-white px-6 py-10 text-center text-[#787774]">
        正在加载活动...
      </div>

      <div v-else-if="error && events.length === 0" class="rounded-lg border border-[#f3d3d6] bg-[#fff7f7] px-6 py-10 text-center text-[#b42318]">
        {{ error }}
      </div>

      <div v-else-if="filteredEvents.length === 0" class="rounded-lg border border-[#ededed] bg-white px-6 py-10 text-center text-[#787774]">
        暂无符合条件的活动。
      </div>

      <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
        <RouterLink
          v-for="event in filteredEvents"
          :key="event.slug"
          :to="`/events/${event.slug}`"
          class="flex flex-col bg-white border border-[#ededed] rounded-lg overflow-hidden hover:bg-[#f9f9f8] transition-colors group cursor-pointer shadow-sm"
        >
          <!-- Image -->
          <div class="w-full h-[180px] border-b border-[#ededed] overflow-hidden bg-[#f1f1ef]">
            <img :src="event.cover" alt="Cover" class="w-full h-full object-cover" />
          </div>

          <!-- Info -->
          <div class="p-4 flex flex-col gap-2">
            <h3 class="text-[#37352f] text-[16px] font-semibold leading-snug truncate">{{ event.title }}</h3>
            
            <div class="flex flex-wrap gap-1 mt-1">
              <!-- Status Tag -->
              <span 
                class="px-1.5 py-0.5 rounded text-[12px] font-medium flex items-center gap-1"
                :class="{
                  'bg-[#fdebec] text-[#1c3829]': event.status === 'live',
                  'bg-[#e3e2e0] text-[#32302c]': event.status === 'done',
                  'bg-[#e8f3f8] text-[#1e3250]': event.status === 'next'
                }"
              >
                {{ event.status === 'live' ? '进行中' : event.status === 'next' ? '即将开始' : '已结束' }}
              </span>
              <!-- City Tag -->
              <span class="px-1.5 py-0.5 rounded bg-[#f1f1ef] text-[#37352f] text-[12px] font-medium">
                {{ event.city }}
              </span>
              <!-- Type Tag -->
              <span class="px-1.5 py-0.5 rounded bg-[#f1f0f2] text-[#37352f] text-[12px] font-medium">
                {{ event.type }}
              </span>
            </div>

            <div class="text-[#787774] text-[14px] mt-1">
              {{ event.time.split(' ')[0] }}
            </div>
          </div>
        </RouterLink>
      </div>

    </main>

    <!-- Footer -->
    <footer class="mt-20 py-8 text-[14px] text-[#787774]">
      <div class="mx-auto max-w-[1440px] px-6 md:px-10 lg:px-14 flex items-center justify-between">
        <div>© 2026 VOIDLAB.AI</div>
        <button type="button" class="hover:text-[#37352f] transition-colors" @click="scrollToTop">
          回到顶部
        </button>
      </div>
    </footer>

  </div>
</template>
