<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref } from "vue";
import { useRouter } from "vue-router";
import { loadInsights, useInsightsFeed } from "../composables/useInsightsFeed";
import SiteHeader from "../components/SiteHeader.vue";

const router = useRouter();
const { insights, newsInsights, loading, error } = useInsightsFeed();

const searchQuery = ref("");

// 从 insights 里分离出一些数据用于不同区块展示
const filteredInsights = computed(() => {
  const keyword = searchQuery.value.trim().toLowerCase();
  if (!keyword) {
    return insights.value;
  }

  return insights.value.filter((item) => {
    const haystack = [
      item.title,
      item.summary,
      item.category,
      item.audience,
      item.tags.join(" ")
    ].join(" ").toLowerCase();

    return haystack.includes(keyword);
  });
});

const carouselItems = computed(() => {
  const source = filteredInsights.value.filter((item) => item.featured);
  const target = source.length > 0 ? source : filteredInsights.value;
  return target.slice(0, 4);
});

const picksInsights = computed(() => {
  const source = filteredInsights.value.filter((item) => item.featured);
  const target = source.length > 1 ? source : filteredInsights.value;
  return target.slice(1, 5);
});

const deepDives = computed(() => filteredInsights.value.slice(0, 6));

const carouselIndex = ref(0);
let carouselTimer: number | null = null;

function startCarousel() {
  if (carouselItems.value.length <= 1) {
    return;
  }
  stopCarousel();
  carouselTimer = window.setInterval(() => {
    carouselIndex.value = (carouselIndex.value + 1) % carouselItems.value.length;
  }, 4000);
}

function stopCarousel() {
  if (carouselTimer) {
    clearInterval(carouselTimer);
    carouselTimer = null;
  }
}

function setCarouselIndex(idx: number) {
  carouselIndex.value = idx;
  startCarousel();
}

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

onMounted(() => {
  document.title = "VOID LAB | 资讯";
  void loadInsights();
  startCarousel();
});

onBeforeUnmount(() => {
  stopCarousel();
});
</script>

<template>
  <div class="relative z-10 min-h-screen overflow-x-hidden bg-white text-[#111111] font-sans pt-16">
    <!-- Header -->
    <SiteHeader theme="light" activePath="/insights" />

    <div class="mx-auto max-w-[1440px] overflow-x-hidden px-6 py-12 md:px-10 lg:px-14">
      
      <!-- Hero Section -->
      <section class="grid grid-cols-1 lg:grid-cols-[1fr_1.1fr] gap-12 lg:gap-20 items-center">
        <!-- Left: Text & Search -->
        <div class="flex flex-col">
          <h1 class="text-5xl md:text-6xl font-extrabold leading-[1.1] tracking-tight">
            VOIDLAB 资讯
          </h1>
          <h2 class="mt-4 text-3xl md:text-4xl font-bold leading-snug">
            看清 AI 变化，<br />
            找到可行动的机会
          </h2>
          <p class="mt-6 text-gray-500 text-lg leading-relaxed max-w-lg">
            追踪 AI 趋势，识别异常信号，更快理解技术变化，把值得跟进的机会筛出来。
          </p>
          
          <div class="mt-10 relative max-w-md">
            <input 
              v-model="searchQuery"
              type="text" 
              placeholder="搜索关键词"
              class="w-full bg-gray-50 border border-gray-100 rounded-full py-4 pl-6 pr-12 text-base focus:outline-none focus:ring-2 focus:ring-black/5 transition-shadow"
            />
            <button class="absolute right-5 top-1/2 -translate-y-1/2 text-gray-400 hover:text-black">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="11" cy="11" r="8"></circle>
                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
              </svg>
            </button>
          </div>
        </div>

        <!-- Right: Featured Hero Card Carousel -->
        <div class="relative flex h-[360px] w-full items-center justify-center overflow-hidden perspective-1000 md:h-[460px]" @mouseenter="stopCarousel" @mouseleave="startCarousel">

          <div v-if="loading && carouselItems.length === 0" class="absolute inset-0 flex items-center justify-center rounded-3xl bg-gray-50 text-gray-400">
            正在加载资讯...
          </div>

          <div v-else-if="error && carouselItems.length === 0" class="absolute inset-0 flex items-center justify-center rounded-3xl bg-[#fff7f7] text-[#b42318]">
            {{ error }}
          </div>
          
          <div 
            v-for="(item, index) in carouselItems"
            :key="item.slug"
            class="absolute w-[85%] md:w-[80%] h-full rounded-3xl overflow-hidden shadow-2xl transition-all duration-500 ease-[cubic-bezier(0.25,1,0.5,1)] cursor-pointer"
            :class="{
              'z-30 scale-100 translate-x-0 opacity-100': index === carouselIndex,
              'z-20 scale-[0.85] -translate-x-[30%] opacity-50 hover:opacity-70': index === (carouselIndex - 1 + carouselItems.length) % carouselItems.length,
              'z-20 scale-[0.85] translate-x-[30%] opacity-50 hover:opacity-70': index === (carouselIndex + 1) % carouselItems.length,
              'z-10 scale-[0.75] translate-x-0 opacity-0 pointer-events-none': index !== carouselIndex && index !== (carouselIndex - 1 + carouselItems.length) % carouselItems.length && index !== (carouselIndex + 1) % carouselItems.length
            }"
            @click="index === carouselIndex ? router.push(`/insights/${item.slug}`) : setCarouselIndex(index)"
          >
            <div class="absolute inset-0 bg-gray-900">
              <img 
                :src="item.cover" 
                alt="Cover" 
                class="w-full h-full object-cover opacity-80 group-hover:opacity-100 transition-opacity duration-500"
              />
              <div class="absolute inset-0 bg-gradient-to-t from-black/90 via-black/40 to-transparent"></div>
            </div>
            
            <!-- Logo Badge -->
            <div class="absolute top-6 left-6 flex items-center gap-2">
              <div class="grid h-5 w-5 grid-cols-2 gap-[1px]">
                <div class="bg-white"></div>
                <div class="bg-white/40"></div>
                <div class="bg-white/40"></div>
                <div class="bg-white"></div>
              </div>
              <span class="text-white font-bold tracking-widest text-sm">VOIDLAB</span>
            </div>

            <!-- Content -->
            <div class="absolute bottom-8 left-8 right-8 transition-all duration-500" :class="index === carouselIndex ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'">
              <div class="inline-block border border-white/20 bg-white/10 backdrop-blur-md rounded-lg px-3 py-1.5 mb-4">
                <span class="text-white text-xs font-semibold tracking-wider uppercase">{{ item.category }}</span>
              </div>
              <h3 class="text-white text-2xl md:text-3xl lg:text-4xl font-bold leading-tight line-clamp-2">
                {{ item.title }}
              </h3>
              <p class="mt-3 text-gray-300 text-sm md:text-base line-clamp-2 max-w-lg">
                {{ item.summary }}
              </p>
            </div>
          </div>

          <!-- Carousel dots -->
          <div class="absolute -bottom-10 left-0 right-0 flex justify-center gap-2">
            <button
              v-for="(_, index) in carouselItems"
              :key="index"
              @click="setCarouselIndex(index)"
              class="h-1.5 rounded-full transition-all duration-300"
              :class="index === carouselIndex ? 'w-6 bg-black' : 'w-1.5 bg-gray-300 hover:bg-gray-400'"
            ></button>
          </div>
        </div>
      </section>

      <!-- Picks Section -->
      <section id="picks" class="mt-32">
        <div class="flex items-end justify-between mb-8">
          <h2 class="text-2xl md:text-3xl font-bold">精选推荐</h2>
          <div class="flex items-center gap-4 text-gray-500 text-sm">
            <button class="hover:text-black transition-colors"><</button>
            <span class="font-medium text-black">1 <span class="text-gray-400 font-normal">/ 4</span></span>
            <button class="hover:text-black transition-colors">></button>
          </div>
        </div>

        <div v-if="picksInsights.length === 0" class="rounded-2xl border border-gray-100 bg-gray-50 px-6 py-10 text-center text-gray-400">
          暂无精选推荐内容。
        </div>

        <div v-else class="flex gap-6 overflow-x-auto pb-8 hide-scrollbar snap-x snap-mandatory">
          <RouterLink
            v-for="item in picksInsights"
            :key="item.slug"
            :to="`/insights/${item.slug}`"
            class="group flex-none w-[280px] md:w-[320px] flex flex-col gap-4 snap-start"
          >
            <!-- Image Box -->
            <div class="relative w-full aspect-[4/3] rounded-2xl overflow-hidden bg-gray-100">
              <img 
                :src="item.cover" 
                alt="Cover" 
                class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
              />
              <div class="absolute inset-0 bg-black/10 group-hover:bg-transparent transition-colors"></div>
              <div class="absolute top-3 right-3 flex items-center gap-1.5">
                <div class="grid h-3.5 w-3.5 grid-cols-2 gap-[1px]">
                  <div class="bg-[var(--color-turquoise)]"></div>
                  <div class="bg-white/40"></div>
                  <div class="bg-white/40"></div>
                  <div class="bg-white"></div>
                </div>
                <span class="text-white text-[10px] font-bold tracking-widest">VL</span>
              </div>
              <div class="absolute bottom-0 left-0 right-0 p-4 bg-gradient-to-t from-black/80 to-transparent">
                <span class="text-[var(--color-turquoise)] text-xl font-black italic leading-none block transform -skew-x-6">
                  {{ item.category === '模型更新' ? 'MODELS' : item.category === '产品拆解' ? 'PRODUCT' : 'TRENDS' }}
                </span>
                <span class="text-white text-[10px] font-bold uppercase tracking-widest">{{ item.category }}</span>
              </div>
            </div>
            
            <!-- Text Content -->
            <div>
              <h3 class="text-lg font-bold leading-snug line-clamp-2 group-hover:text-blue-600 transition-colors">
                {{ item.title }}
              </h3>
              <p class="mt-2 text-sm text-gray-500 line-clamp-2 leading-relaxed">
                {{ item.summary }}
              </p>
            </div>
          </RouterLink>
        </div>
      </section>

      <!-- Two Columns: Sector & News -->
      <section class="mt-20 grid grid-cols-1 lg:grid-cols-[1fr_400px] gap-12 lg:gap-20">
        
        <!-- Left: Deep Dives -->
        <div id="themes">
          <div class="flex items-center gap-2 mb-8 group cursor-pointer w-max">
            <h2 class="text-2xl font-bold">专题深读</h2>
            <span class="text-gray-400 group-hover:text-black transition-colors transform group-hover:translate-x-1">→</span>
          </div>

          <div class="flex flex-col gap-10">
            <div v-if="deepDives.length === 0" class="rounded-2xl border border-gray-100 bg-gray-50 px-6 py-10 text-center text-gray-400">
              暂无专题深读内容。
            </div>

            <RouterLink
              v-for="item in deepDives"
              :key="item.slug"
              :to="`/insights/${item.slug}`"
              class="group flex flex-col sm:flex-row gap-6 items-start"
            >
              <!-- Thumbnail -->
              <div class="relative w-full sm:w-[260px] shrink-0 aspect-[16/10] rounded-xl overflow-hidden bg-gray-100">
                <img 
                  :src="item.cover" 
                  alt="Cover" 
                  class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
                />
                <div class="absolute top-2 right-2 flex items-center gap-1">
                  <div class="w-2.5 h-2.5 bg-black"></div>
                  <span class="text-white text-[8px] font-bold tracking-widest drop-shadow-md">VL</span>
                </div>
              </div>
              
              <!-- Content -->
              <div class="flex flex-col h-full py-1">
                <h3 class="text-xl font-bold leading-snug group-hover:text-blue-600 transition-colors line-clamp-2">
                  {{ item.title }}
                </h3>
                <p class="mt-3 text-sm text-gray-500 line-clamp-2 leading-relaxed">
                  研究对象：{{ item.tags.join('、') }}<br/>
                  {{ item.summary }}
                </p>
                <div class="mt-auto pt-4 text-xs text-gray-400 font-medium">
                  {{ item.publishedAt }}
                </div>
              </div>
            </RouterLink>
          </div>
          
          <div class="mt-12 flex items-center gap-2 group cursor-pointer w-max">
            <h2 class="text-2xl font-bold">编辑观点</h2>
            <span class="text-gray-400 group-hover:text-black transition-colors transform group-hover:translate-x-1">→</span>
          </div>
        </div>

        <!-- Right: News -->
        <aside id="news">
          <div class="flex items-center gap-2 mb-8 group cursor-pointer w-max">
            <h2 class="text-2xl font-bold">实时快讯</h2>
            <span class="text-gray-400 group-hover:text-black transition-colors transform group-hover:translate-x-1">→</span>
          </div>

          <div class="bg-[#f8f9fa] rounded-[2rem] p-8">
            <ul class="flex flex-col divide-y divide-gray-200/60">
              <li v-if="newsInsights.length === 0" class="py-2 text-sm text-gray-400">
                暂无快讯内容。
              </li>

              <li 
                v-for="news in newsInsights" 
                :key="news.slug"
                class="py-5 first:pt-0 last:pb-0 group cursor-pointer"
              >
                <div class="flex gap-3">
                  <div class="mt-1.5 w-1.5 h-1.5 rounded-full bg-black shrink-0 group-hover:scale-150 transition-transform"></div>
                  <div>
                    <h4 class="text-sm font-semibold leading-relaxed group-hover:text-blue-600 transition-colors">
                      {{ news.title }}
                    </h4>
                    <div class="mt-2 flex items-center gap-1.5 text-xs text-gray-400">
                      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <circle cx="12" cy="12" r="10"></circle>
                        <polyline points="12 6 12 12 16 14"></polyline>
                      </svg>
                      {{ news.publishedAt }}
                    </div>
                  </div>
                </div>
              </li>
            </ul>
          </div>
        </aside>

      </section>

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
  </div>
</template>

<style scoped>
.hide-scrollbar::-webkit-scrollbar {
  display: none;
}
.hide-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
.perspective-1000 {
  perspective: 1000px;
}
.rotate-y-6 {
  transform: rotateY(-6deg);
}
</style>
