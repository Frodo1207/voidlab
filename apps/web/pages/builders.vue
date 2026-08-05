<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { loadBuilders, useBuilderNetwork, type BuilderRole } from "../composables/useBuilderNetwork";
import SiteHeader from "../components/SiteHeader.vue";

const { builders, cityCount, roleCount, loading, error } = useBuilderNetwork();

const selectedRole = ref<"all" | BuilderRole>("all");
const selectedCity = ref<"all" | string>("all");
const availableOnly = ref(true);

const filteredBuilders = computed(() =>
  builders.value.filter((builder) => {
    const matchRole = selectedRole.value === "all" || builder.role === selectedRole.value;
    const matchCity = selectedCity.value === "all" || builder.city === selectedCity.value;
    const matchAvailable = !availableOnly.value || builder.contactable;
    return matchRole && matchCity && matchAvailable;
  })
);

const roleOptions = computed(() => [
  { label: "全部", value: "all" as const },
  ...Array.from(new Set(builders.value.map((builder) => builder.role))).map((role) => ({
    label: role,
    value: role
  }))
]);

const cityOptions = computed(() => [
  { label: "全部城市", value: "all" },
  ...Array.from(new Set(builders.value.map((builder) => builder.city))).map((city) => ({
    label: city.toUpperCase(),
    value: city
  }))
]);

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

onMounted(() => {
  document.title = "VOID LAB | 社交网络";
  void loadBuilders();
});
</script>

<template>
  <div class="relative z-10 min-h-screen bg-[#08080a] text-white font-sans selection:bg-purple-500/30 pt-16">
    
    <!-- Header -->
    <SiteHeader theme="dark" activePath="/builders" />

    <!-- Hero Section -->
    <section class="relative overflow-hidden px-6 md:px-10 lg:px-14 py-16 lg:py-24 border-b border-white/5">
      <!-- Glow Background -->
      <div class="absolute top-1/2 right-0 lg:right-[10%] -translate-y-1/2 w-[600px] h-[600px] lg:w-[800px] lg:h-[800px] bg-[radial-gradient(circle,rgba(124,58,237,0.35)_0%,transparent_60%)] pointer-events-none blur-2xl"></div>
      
      <div class="relative z-10 mx-auto max-w-[1440px]">
        <h1 class="text-5xl md:text-6xl lg:text-7xl font-bold leading-[1.1] tracking-tight text-white max-w-3xl">
          找到真正能一起<br />
          把事做成的人
        </h1>
        
        <div class="mt-20 grid grid-cols-2 md:grid-cols-4 gap-8 md:gap-12">
          <div>
            <div class="text-gray-400 font-mono text-xs md:text-sm uppercase tracking-wider">入驻成员</div>
            <div class="mt-3 text-3xl md:text-4xl font-mono font-bold text-white">{{ builders.length }}</div>
          </div>
          <div>
            <div class="text-gray-400 font-mono text-xs md:text-sm uppercase tracking-wider">角色方向</div>
            <div class="mt-3 text-3xl md:text-4xl font-mono font-bold text-white">{{ roleCount }}</div>
          </div>
          <div>
            <div class="text-gray-400 font-mono text-xs md:text-sm uppercase tracking-wider">覆盖城市</div>
            <div class="mt-3 text-3xl md:text-4xl font-mono font-bold text-white">{{ cityCount }}</div>
          </div>
          <div>
            <div class="text-gray-400 font-mono text-xs md:text-sm uppercase tracking-wider">当前可联系</div>
            <div class="mt-3 text-3xl md:text-4xl font-mono font-bold text-white">{{ builders.filter((builder) => builder.contactable).length }}</div>
          </div>
        </div>
      </div>
    </section>

    <!-- Main Content -->
    <main class="mx-auto max-w-[1440px] px-6 py-12 md:px-10 lg:px-14">
      
      <!-- Filter Bar -->
      <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-6 mb-8">
        
        <!-- Tabs -->
        <div class="flex flex-wrap gap-2">
          <button
            v-for="option in roleOptions"
            :key="option.value"
            @click="selectedRole = option.value"
            class="px-4 py-2.5 rounded-md text-xs font-bold uppercase tracking-wider transition-colors"
            :class="selectedRole === option.value ? 'bg-white text-black' : 'bg-[#1c1d25] text-gray-400 hover:bg-[#252630] hover:text-white'"
          >
            {{ option.label }}
          </button>
        </div>

        <!-- Right Controls -->
        <div class="flex items-center gap-3">
          <label class="flex items-center gap-2 bg-[#1c1d25] px-4 py-2.5 rounded-md cursor-pointer hover:bg-[#252630] transition-colors">
            <input type="checkbox" v-model="availableOnly" class="accent-purple-500 w-4 h-4 rounded-sm border-gray-600 bg-transparent" />
            <span class="text-xs font-bold uppercase tracking-wider text-gray-300">仅看可联系</span>
          </label>

          <select 
            v-model="selectedCity"
            class="bg-[#1c1d25] text-gray-300 px-4 py-2.5 rounded-md outline-none text-xs font-bold uppercase tracking-wider border-r-[12px] border-transparent hover:bg-[#252630] transition-colors cursor-pointer"
          >
            <option v-for="option in cityOptions" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </div>
      </div>

      <!-- List -->
      <div v-if="loading && builders.length === 0" class="rounded-2xl border border-white/10 bg-[#111218] px-6 py-10 text-center text-gray-400">
        正在加载 Builder Network...
      </div>

      <div v-else-if="error && builders.length === 0" class="rounded-2xl border border-[#7f1d1d] bg-[#2a1116] px-6 py-10 text-center text-[#fca5a5]">
        {{ error }}
      </div>

      <div v-else-if="filteredBuilders.length === 0" class="rounded-2xl border border-white/10 bg-[#111218] px-6 py-10 text-center text-gray-400">
        当前筛选条件下还没有可展示的成员。
      </div>

      <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6">
        <RouterLink
          v-for="builder in filteredBuilders"
          :key="builder.slug"
          :to="`/builders/${builder.slug}`"
          class="flex flex-col sm:flex-row sm:items-center gap-5 bg-[#111218] hover:bg-[#161821] p-5 md:p-6 rounded-2xl transition-all group"
        >
          <!-- Avatar -->
          <div class="w-20 h-20 md:w-24 md:h-24 shrink-0 rounded-xl bg-[#1c1d25] overflow-hidden flex items-center justify-center relative">
            <img v-if="builder.cover" :src="builder.cover" alt="Avatar" class="w-full h-full object-cover opacity-80 group-hover:opacity-100 group-hover:scale-105 transition-all duration-500" />
            <div v-else class="text-2xl font-bold text-gray-500">
              {{ builder.name.slice(0, 2).toUpperCase() }}
            </div>
          </div>

          <!-- Info -->
          <div class="flex-1 min-w-0 flex flex-col justify-center">
            <h3 class="text-white text-xl font-bold">{{ builder.name }}</h3>
            <p class="text-gray-400 font-mono text-sm mt-2 truncate">{{ builder.intro }}</p>
            
            <div class="flex flex-wrap items-center gap-3 mt-3 text-xs font-mono">
              <span class="text-purple-400 flex items-center gap-1">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"></polygon></svg>
                {{ builder.role }}
              </span>
              <span class="text-gray-600">|</span>
              <span class="text-gray-300">{{ builder.city }}</span>
            </div>
          </div>

          <!-- Open For -->
          <div class="shrink-0 flex items-center text-white font-mono text-lg mt-4 sm:mt-0 sm:ml-auto">
            <span class="font-bold whitespace-nowrap">{{ builder.openFor }}</span>
          </div>
        </RouterLink>
      </div>

    </main>

    <!-- Footer -->
    <footer class="mt-20 border-t border-white/5 py-8 text-sm text-gray-500 font-mono">
      <div class="mx-auto max-w-[1440px] px-6 md:px-10 lg:px-14 flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>© 2026 VOIDLAB.AI / 社交网络</div>
        <button type="button" class="text-left transition-colors hover:text-white md:text-right" @click="scrollToTop">
          回到顶部
        </button>
      </div>
    </footer>

  </div>
</template>
