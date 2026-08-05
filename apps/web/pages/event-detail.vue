<script setup lang="ts">
import { computed, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  eventActionLabel,
  eventStatusLabel,
  useEventArchive
} from "../composables/useEventArchive";
import { submitEventRsvp } from "../composables/useLeadIntake";
import SiteHeader from "../components/SiteHeader.vue";

const route = useRoute();
const router = useRouter();
const { loadEventBySlug } = useEventArchive();

const slug = computed(() => String(route.params.slug ?? ""));
const event = ref<Awaited<ReturnType<typeof loadEventBySlug>>>(null);
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

async function syncEventDetail() {
  loading.value = true;
  error.value = "";

  try {
    const detail = await loadEventBySlug(slug.value);
    if (!detail) {
      router.replace("/events");
      return;
    }

    event.value = detail;
    document.title = `${detail.title} | VOIDLAB`;
  } catch (loadError) {
    error.value = loadError instanceof Error ? loadError.message : "加载活动失败";
    router.replace("/events");
  } finally {
    loading.value = false;
  }
}

watch(slug, () => {
  void syncEventDetail();
}, { immediate: true });

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

async function handleSubmit() {
  if (!event.value) {
    return;
  }

  submitting.value = true;
  submitError.value = "";
  submitSuccess.value = "";

  try {
    await submitEventRsvp(event.value.id, {
      name: form.name,
      contact: form.contact,
      message: form.message
    });

    submitSuccess.value = "报名意向已提交，我们会尽快和你确认。";
    form.name = "";
    form.contact = "";
    form.message = "";
  } catch (submitErr) {
    submitError.value = submitErr instanceof Error ? submitErr.message : "提交失败，请稍后重试";
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="relative z-10 min-h-screen bg-[#f8f9fa] text-[#37352f] font-sans selection:bg-[#cce2ff] pt-16" style="background-image: radial-gradient(#d1d5db 1px, transparent 1px); background-size: 24px 24px;">
    
    <!-- Header -->
    <SiteHeader theme="light" activePath="/events" />

    <div v-if="loading" class="mx-auto max-w-[900px] px-6 py-12 md:py-20 text-center text-[#787774]">
      正在加载活动详情...
    </div>

    <div v-else-if="event" class="mx-auto max-w-[900px] px-6 py-12 md:py-20">
      
      <!-- Event Header Cover -->
      <div class="w-full h-[30vh] max-h-[300px] overflow-hidden bg-[#f1f1ef] mb-12 rounded-lg">
        <img :src="event.cover" :alt="event.title" class="w-full h-full object-cover" />
      </div>

      <article class="flex flex-col">
        <!-- Icon & Title -->
        <div class="mb-8">
          <div class="text-[78px] leading-none mb-6 -mt-24 relative z-10 drop-shadow-sm">🗓️</div>
          <h1 class="text-4xl md:text-[40px] font-bold leading-[1.2] tracking-tight text-[#37352f]">
            {{ event.title }}
          </h1>
        </div>

        <!-- Properties (Notion Style) -->
        <div class="flex flex-col gap-3 py-6 border-y border-[#ededed] mb-10">
          
          <!-- Status -->
          <div class="flex items-center text-[14px]">
            <div class="w-[140px] text-[#787774] flex items-center gap-2">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
              状态
            </div>
            <div class="flex-1">
              <span 
                class="px-2 py-1 rounded text-[14px] font-medium inline-flex items-center gap-1.5"
                :class="{
                  'bg-[#fdebec] text-[#1c3829]': event.status === 'live',
                  'bg-[#e3e2e0] text-[#32302c]': event.status === 'done',
                  'bg-[#e8f3f8] text-[#1e3250]': event.status === 'next'
                }"
              >
                <span class="w-1.5 h-1.5 rounded-full" :class="{
                  'bg-red-500': event.status === 'live',
                  'bg-gray-500': event.status === 'done',
                  'bg-blue-500': event.status === 'next'
                }"></span>
                {{ eventStatusLabel(event.status) }}
              </span>
            </div>
          </div>

          <!-- Type -->
          <div class="flex items-center text-[14px]">
            <div class="w-[140px] text-[#787774] flex items-center gap-2">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
              类型
            </div>
            <div class="flex-1">
              <span class="px-2 py-1 rounded bg-[#f1f0f2] text-[#37352f] text-[14px] font-medium">
                {{ event.type }}
              </span>
            </div>
          </div>

          <!-- Time -->
          <div class="flex items-center text-[14px]">
            <div class="w-[140px] text-[#787774] flex items-center gap-2">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
              日期
            </div>
            <div class="flex-1 text-[#37352f]">
              {{ event.time }}
            </div>
          </div>

          <!-- Location -->
          <div class="flex items-center text-[14px]">
            <div class="w-[140px] text-[#787774] flex items-center gap-2">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path><circle cx="12" cy="10" r="3"></circle></svg>
              地点
            </div>
            <div class="flex-1 text-[#37352f]">
              <span class="underline decoration-gray-300 underline-offset-4">{{ event.location }} ({{ event.city }})</span>
            </div>
          </div>

        </div>

        <!-- Action Button -->
        <div class="mb-12" v-if="event.status !== 'done'">
          <div class="rounded-2xl border border-[#d9e7fb] bg-[#f8fbff] p-5 md:p-6">
            <div class="flex items-center gap-2 text-[#1b6fc2] text-[14px] font-semibold">
              {{ eventActionLabel(event.status) }}
            </div>
            <p class="mt-3 text-[14px] text-[#5b6473]">
              留下你的姓名和联系方式，系统会把这次活动意向直接写入 VOIDLAB Leads，方便我们后续确认席位和沟通细节。
            </p>

            <form class="mt-5 grid gap-4" @submit.prevent="handleSubmit">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <input
                  v-model="form.name"
                  type="text"
                  required
                  class="w-full rounded-xl border border-[#d7dde5] bg-white px-4 py-3 text-[#37352f] outline-none focus:border-[#2383e2]"
                  placeholder="你的姓名"
                />
                <input
                  v-model="form.contact"
                  type="text"
                  required
                  class="w-full rounded-xl border border-[#d7dde5] bg-white px-4 py-3 text-[#37352f] outline-none focus:border-[#2383e2]"
                  placeholder="微信 / 邮箱 / 手机号"
                />
              </div>

              <textarea
                v-model="form.message"
                rows="4"
                class="w-full rounded-xl border border-[#d7dde5] bg-white px-4 py-3 text-[#37352f] outline-none focus:border-[#2383e2]"
                placeholder="补充一下你的参与目标、团队情况，或者想了解的问题"
              />

              <div v-if="submitError" class="rounded-xl border border-[#f3d3d6] bg-[#fff7f7] px-4 py-3 text-[14px] text-[#b42318]">
                {{ submitError }}
              </div>

              <div v-if="submitSuccess" class="rounded-xl border border-[#ccebd5] bg-[#f0fdf4] px-4 py-3 text-[14px] text-[#166534]">
                {{ submitSuccess }}
              </div>

              <div>
                <button
                  type="submit"
                  :disabled="submitting"
                  class="px-4 py-2 rounded shadow-sm font-medium transition-colors text-[14px] flex items-center gap-2 bg-[#2383e2] text-white hover:bg-[#1b6fc2] disabled:opacity-60 disabled:cursor-not-allowed"
                >
                  {{ submitting ? "提交中..." : eventActionLabel(event.status) }}
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="5" y1="12" x2="19" y2="12"></line><polyline points="12 5 19 12 12 19"></polyline></svg>
                </button>
              </div>
            </form>
          </div>
        </div>

        <!-- Main Content -->
        <div class="prose prose-p:text-[#37352f] prose-headings:text-[#37352f] max-w-none leading-[1.6]">
          <p class="text-[18px] text-[#787774] mb-8 font-medium italic border-l-4 border-[#e3e2e0] pl-4">
            {{ event.summary }}
          </p>
          
          <h2 class="text-2xl font-bold mb-4">活动详情</h2>
          <p class="whitespace-pre-line">{{ event.content || "活动详细内容正在整理中。" }}</p>
          
          <h3 class="text-xl font-bold mt-8 mb-4">活动亮点</h3>
          <ul class="list-disc pl-6 marker:text-[#e3e2e0]">
            <li class="pl-2">沉浸式的现场体验与极客氛围</li>
            <li class="pl-2">拒绝理论，只讲真实跑通的案例与代码</li>
            <li class="pl-2">高净值的参与者网络，现场建联</li>
          </ul>
          
          <h3 class="text-xl font-bold mt-8 mb-4">适合人群</h3>
          <p>无论你是开发者、产品经理，还是寻求 AI 破局的企业主，只要你关注“落地”而非“概念”，这里都适合你。</p>

          <hr class="my-10 border-[#ededed]" />

          <div class="bg-[#f1f1ef] rounded p-4 text-[14px] flex items-start gap-3">
            <span class="text-lg">💡</span>
            <div class="text-[#37352f]">
              <strong>需要帮助或企业合作？</strong><br>
              请发送邮件至 <a href="mailto:join@voidlab.ai" class="underline decoration-[#787774] underline-offset-2">join@voidlab.ai</a>
            </div>
          </div>
        </div>
      </article>

      <!-- Footer -->
      <footer class="mt-20 border-t border-[#ededed] py-8 text-[14px] text-[#787774]">
        <div class="flex items-center justify-between">
          <div>© 2026 VOIDLAB.AI</div>
          <button type="button" class="hover:text-[#37352f] transition-colors" @click="scrollToTop">
          回到顶部
          </button>
        </div>
      </footer>
    </div>

    <div v-else class="mx-auto max-w-[900px] px-6 py-12 md:py-20 text-center text-[#b42318]">
      {{ error || "活动不存在，正在返回活动列表..." }}
    </div>
  </div>
</template>
