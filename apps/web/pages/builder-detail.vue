<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useBuilderNetwork } from "../composables/useBuilderNetwork";
import { submitBuilderInquiry } from "../composables/useLeadIntake";
import SiteHeader from "../components/SiteHeader.vue";

const route = useRoute();
const router = useRouter();
const { builders, loadBuilderBySlug } = useBuilderNetwork();

const slug = computed(() => String(route.params.slug ?? ""));
const builder = ref<Awaited<ReturnType<typeof loadBuilderBySlug>>>(null);
const otherBuilders = computed(() => builders.value.filter((item) => item.slug !== slug.value).slice(0, 4));
const loading = ref(true);
const error = ref("");
const submitting = ref(false);
const submitSuccess = ref("");
const submitError = ref("");
const form = reactive({
  name: "",
  contact: "",
  message: ""
});

async function syncBuilderDetail() {
  loading.value = true;
  error.value = "";

  try {
    const detail = await loadBuilderBySlug(slug.value);
    if (!detail) {
      router.replace("/builders");
      return;
    }

    builder.value = detail;
    document.title = `VOID LAB | ${detail.name}`;
  } catch (loadError) {
    error.value = loadError instanceof Error ? loadError.message : "加载 Builder 详情失败";
    router.replace("/builders");
  } finally {
    loading.value = false;
  }
}

watch(slug, () => {
  void syncBuilderDetail();
}, { immediate: true });

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

async function handleSubmit() {
  if (!builder.value) {
    return;
  }

  submitting.value = true;
  submitError.value = "";
  submitSuccess.value = "";

  try {
    await submitBuilderInquiry(builder.value.slug, builder.value.name, {
      name: form.name,
      contact: form.contact,
      message: form.message
    });

    submitSuccess.value = "合作意向已提交，VOIDLAB 会尽快与你对接。";
    form.name = "";
    form.contact = "";
    form.message = "";
  } catch (error) {
    submitError.value = error instanceof Error ? error.message : "提交失败，请稍后重试";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="relative z-10 min-h-screen bg-[#08080a] text-white font-sans selection:bg-[#42ffd1] selection:text-black pt-16">
    
    <!-- Header -->
    <SiteHeader theme="dark" activePath="/builders" />

    <div v-if="loading" class="mx-auto max-w-[1440px] px-6 py-12 md:px-10 lg:px-14 text-center text-gray-400">
      正在加载 Builder 详情...
    </div>

    <div v-else-if="builder" class="mx-auto max-w-[1440px] px-6 py-12 md:px-10 lg:px-14">
      <div class="grid grid-cols-1 lg:grid-cols-[1fr_380px] gap-12 lg:gap-24">
        
        <!-- Left Column: Main Profile -->
        <article class="flex flex-col">
          <!-- Breadcrumbs -->
          <div class="text-xs text-gray-500 mb-12 font-mono uppercase tracking-wider">
            <RouterLink to="/builders" class="hover:text-white transition-colors">社交网络</RouterLink> <span class="mx-2">/</span>
            <span class="hover:text-white transition-colors cursor-pointer">{{ builder.role }}</span> <span class="mx-2">/</span> 
            <span class="text-white">成员档案</span>
          </div>

          <!-- Profile Header -->
          <div class="flex flex-col md:flex-row gap-8 items-start md:items-center relative">
            <!-- Glow behind avatar -->
            <div class="absolute top-0 left-0 w-40 h-40 bg-purple-600/20 blur-3xl rounded-full pointer-events-none"></div>
            
            <div class="w-32 h-32 md:w-40 md:h-40 rounded-2xl bg-[#1c1d25] border border-white/10 overflow-hidden shrink-0 relative z-10">
              <img v-if="builder.cover" :src="builder.cover" alt="Avatar" class="w-full h-full object-cover" />
              <div v-else class="w-full h-full flex items-center justify-center text-4xl font-bold text-gray-500">
                {{ builder.name.slice(0, 2).toUpperCase() }}
              </div>
            </div>

            <div class="flex flex-col relative z-10">
              <h1 class="text-4xl md:text-5xl font-bold leading-tight tracking-tight text-white">
                {{ builder.name }}
              </h1>
              <div class="mt-3 text-xl text-gray-400 font-mono">{{ builder.title }}</div>
              
              <div class="mt-6 flex flex-wrap items-center gap-3 font-mono text-xs">
                <span class="px-3 py-1.5 bg-purple-500/10 text-purple-400 border border-purple-500/20 rounded uppercase tracking-wider">
                  {{ builder.role }}
                </span>
                <span class="px-3 py-1.5 bg-[#1c1d25] text-gray-300 border border-white/5 rounded uppercase tracking-wider">
                  {{ builder.city }}
                </span>
              </div>
            </div>
          </div>

          <!-- Intro -->
          <div class="mt-12 text-xl md:text-2xl font-medium leading-relaxed text-gray-300">
            {{ builder.intro }}
          </div>

          <!-- Story -->
          <div class="mt-12 prose prose-lg prose-invert max-w-none text-gray-400 leading-[1.8]">
            <h2 class="text-2xl font-bold text-white mb-6">成员故事</h2>
            <p class="mb-6">
              {{ builder.story }}
            </p>
          </div>

          <!-- Tags Area -->
          <div class="mt-12 grid grid-cols-1 md:grid-cols-2 gap-10 border-t border-white/5 pt-10">
            <div>
              <h3 class="text-xs font-bold text-gray-500 uppercase tracking-widest mb-5 font-mono">关注方向</h3>
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="item in builder.focusAreas"
                  :key="item"
                  class="bg-[#111218] border border-white/5 rounded px-3 py-1.5 text-xs text-gray-400 font-mono uppercase tracking-wider"
                >
                  {{ item }}
                </span>
              </div>
            </div>

            <div>
              <h3 class="text-xs font-bold text-gray-500 uppercase tracking-widest mb-5 font-mono">能力栈</h3>
              <div class="flex flex-wrap gap-2">
                <span
                  v-for="item in builder.expertise"
                  :key="item"
                  class="bg-purple-500/5 border border-purple-500/20 rounded px-3 py-1.5 text-xs text-purple-400 font-mono uppercase tracking-wider"
                >
                  {{ item }}
                </span>
              </div>
            </div>
          </div>
        </article>

        <!-- Right Column: Sidebar -->
        <aside class="relative">
          <div class="sticky top-28 flex flex-col gap-8">
            
            <!-- Contact Card -->
            <div class="bg-[#111218] border border-white/5 rounded-2xl p-8">
              <h3 class="text-xl font-bold text-white mb-6">合作方式</h3>
              
              <div class="flex flex-col gap-3 mb-8">
                <div
                  v-for="mode in builder.collaborationModes"
                  :key="mode"
                  class="px-4 py-3 rounded-lg bg-[#1c1d25] border border-white/5 text-gray-300 font-mono text-xs uppercase tracking-wider"
                >
                  {{ mode }}
                </div>
              </div>

              <div class="border-t border-white/5 pt-6 mb-8">
                <div class="text-[10px] uppercase tracking-widest text-gray-500 font-bold mb-3 font-mono">当前状态</div>
                <p class="text-sm text-gray-400 leading-relaxed">
                  {{ builder.availabilityNote }}
                </p>
              </div>

              <div class="flex items-center justify-between mb-6 bg-[#1c1d25] rounded-lg p-4 border border-white/5">
                <div class="text-xs text-gray-400 font-mono uppercase">可提供服务</div>
                <div class="flex items-center text-white font-mono font-bold">
                  {{ builder.openFor }}
                </div>
              </div>

              <form class="grid gap-4" @submit.prevent="handleSubmit">
                <input
                  v-model="form.name"
                  type="text"
                  required
                  class="w-full rounded-lg border border-white/10 bg-[#0b0c11] px-4 py-3 text-white outline-none focus:border-purple-400"
                  placeholder="你的姓名"
                />
                <input
                  v-model="form.contact"
                  type="text"
                  required
                  class="w-full rounded-lg border border-white/10 bg-[#0b0c11] px-4 py-3 text-white outline-none focus:border-purple-400"
                  placeholder="微信 / 邮箱 / 手机号"
                />
                <textarea
                  v-model="form.message"
                  rows="4"
                  class="w-full rounded-lg border border-white/10 bg-[#0b0c11] px-4 py-3 text-white outline-none focus:border-purple-400"
                  placeholder="想合作什么、想一起做什么，或者你希望对方提供什么支持"
                />

                <div v-if="submitError" class="rounded-lg border border-[#7f1d1d] bg-[#2a1116] px-4 py-3 text-sm text-[#fca5a5]">
                  {{ submitError }}
                </div>

                <div v-if="submitSuccess" class="rounded-lg border border-[#14532d] bg-[#0f2217] px-4 py-3 text-sm text-[#86efac]">
                  {{ submitSuccess }}
                </div>

                <button
                  type="submit"
                  :disabled="submitting"
                  class="w-full py-4 rounded-lg font-bold text-white transition-all duration-300 bg-purple-600 hover:bg-purple-500 uppercase tracking-wider text-sm disabled:opacity-60 disabled:cursor-not-allowed"
                >
                  {{ submitting ? "提交中..." : "发起合作" }}
                </button>
              </form>
              <div class="text-center text-xs text-gray-500 mt-4 font-mono">
                需通过 VOIDLAB 发起联系
              </div>
            </div>

            <!-- Other Builders -->
            <div>
              <h3 class="text-sm font-bold text-gray-400 uppercase tracking-widest mb-5 font-mono">相似成员</h3>
              <ul class="flex flex-col gap-3">
                <li 
                  v-for="b in otherBuilders" 
                  :key="b.slug"
                  class="group"
                >
                  <RouterLink :to="`/builders/${b.slug}`" class="flex items-center gap-4 p-3 rounded-xl hover:bg-[#111218] border border-transparent hover:border-white/5 transition-all -mx-3">
                    <div class="w-10 h-10 rounded-lg bg-[#1c1d25] border border-white/5 overflow-hidden shrink-0">
                      <img v-if="b.cover" :src="b.cover" alt="" class="w-full h-full object-cover opacity-60 group-hover:opacity-100 transition-all" />
                    </div>
                    <div>
                      <h4 class="text-sm font-bold text-gray-200 group-hover:text-purple-400 transition-colors">
                        {{ b.name }}
                      </h4>
                      <div class="text-xs text-gray-500 mt-1 font-mono uppercase tracking-wider">{{ b.role }}</div>
                    </div>
                  </RouterLink>
                </li>
              </ul>
            </div>

          </div>
        </aside>

      </div>

      <!-- Footer -->
      <footer class="mt-20 border-t border-white/5 py-8 text-sm text-gray-500 font-mono">
        <div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>© 2026 VOIDLAB.AI / 社交网络</div>
          <button type="button" class="text-left transition-colors hover:text-white md:text-right" @click="scrollToTop">
            回到顶部
          </button>
        </div>
      </footer>
    </div>

    <div v-else class="mx-auto max-w-[1440px] px-6 py-12 md:px-10 lg:px-14 text-center text-[#fca5a5]">
      {{ error || "Builder 不存在，正在返回列表..." }}
    </div>
  </div>
</template>
