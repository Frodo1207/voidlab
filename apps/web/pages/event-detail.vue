<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  eventActionLabel,
  eventSignupStatusLabel,
  eventStatusLabel,
  useEventArchive
} from "../composables/useEventArchive";
import { submitEventRsvp } from "../composables/useLeadIntake";
import SiteHeader from "../components/SiteHeader.vue";
import { renderMarkdown } from "../src/markdown";

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
const showSuccessNotice = ref(false);
const form = reactive({
  name: "",
  wechat: "",
  phone: "",
  position: "",
  industry: "",
  message: ""
});
let successNoticeTimer: number | null = null;

function normalizeMainlandPhone(value: string) {
  const digits = value.replace(/\D/g, "");
  const normalized = digits.length === 13 && digits.startsWith("86") ? digits.slice(2) : digits;
  if (/^1[3-9]\d{9}$/.test(normalized)) {
    return normalized;
  }
  return "";
}

function clearSuccessNoticeTimer() {
  if (successNoticeTimer !== null) {
    window.clearTimeout(successNoticeTimer);
    successNoticeTimer = null;
  }
}

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

  if (event.value.signupMode === "external" && event.value.externalSignupUrl) {
    window.open(event.value.externalSignupUrl, "_blank", "noopener,noreferrer");
    return;
  }

  submitError.value = "";
  submitSuccess.value = "";

  const normalizedPhone = normalizeMainlandPhone(form.phone);
  if (!normalizedPhone) {
    submitError.value = "请输入有效的 11 位手机号。";
    return;
  }

  form.phone = normalizedPhone;
  submitting.value = true;

  try {
    const combinedMessage = [
      form.position ? `[职位] ${form.position}` : "",
      form.industry ? `[行业信息] ${form.industry}` : "",
      form.message ? `[备注] ${form.message}` : ""
    ].filter(Boolean).join("\n\n");

    const combinedContact = [
      form.wechat ? `微信: ${form.wechat}` : "",
      form.phone ? `手机: ${form.phone}` : ""
    ].filter(Boolean).join(" / ");

    await submitEventRsvp(event.value.id, {
      name: form.name,
      contact: combinedContact,
      message: combinedMessage
    });

    submitSuccess.value = event.value.signupSuccessMessage?.trim() || "报名意向已提交，我们会尽快和你确认。";
    form.name = "";
    form.wechat = "";
    form.phone = "";
    form.position = "";
    form.industry = "";
    form.message = "";
    closeSignupModal();
    showSuccessNotice.value = true;
    clearSuccessNoticeTimer();
    successNoticeTimer = window.setTimeout(() => {
      showSuccessNotice.value = false;
      submitSuccess.value = "";
      successNoticeTimer = null;
    }, 1800);
  } catch (submitErr) {
    submitError.value = submitErr instanceof Error ? submitErr.message : "提交失败，请稍后重试";
  } finally {
    submitting.value = false;
  }
}

const canUseInternalSignup = computed(() =>
  Boolean(event.value) &&
  event.value!.signupMode === "internal" &&
  event.value!.signupStatus === "open"
);

const showSignupPanel = computed(() =>
  Boolean(event.value) &&
  (event.value!.signupMode === "external" || event.value!.signupStatus !== "ended")
);

const signupPanelTitle = computed(() => {
  if (!event.value) return "";
  if (event.value.signupButtonLabel) return event.value.signupButtonLabel;
  if (event.value.signupMode === "external") return "前往报名";
  return eventActionLabel(event.value.status);
});

const signupHint = computed(() => {
  if (!event.value) return "";
  if (event.value.signupStatus === "open") {
    return "留下你的姓名和联系方式，系统会把这次活动意向直接写入 VOIDLAB Leads，方便我们后续确认席位和沟通细节。";
  }
  if (event.value.signupClosedReason) {
    return event.value.signupClosedReason;
  }
  return eventSignupStatusLabel(event.value.signupStatus);
});

const renderedContent = computed(() =>
  event.value?.content ? renderMarkdown(event.value.content) : ""
);

const isModalOpen = ref(false);

function openSignupModal() {
  if (event.value?.signupMode === "external" && event.value.externalSignupUrl) {
    window.open(event.value.externalSignupUrl, "_blank", "noopener,noreferrer");
    return;
  }
  isModalOpen.value = true;
}

function closeSignupModal() {
  isModalOpen.value = false;
  submitError.value = "";
}

function closeSuccessNotice() {
  clearSuccessNoticeTimer();
  showSuccessNotice.value = false;
  submitSuccess.value = "";
}

onBeforeUnmount(() => {
  clearSuccessNoticeTimer();
});
</script>

<template>
  <div class="relative z-10 min-h-screen bg-[#f6f5f1] text-[#333333] font-sans selection:bg-[#cce2ff] pt-16">
    
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
        <div class="mb-12" v-if="showSignupPanel">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 rounded-2xl border border-[#eaeaea] bg-white p-5 md:p-6 shadow-sm">
            <div>
              <div class="flex items-center gap-2 text-[#6f8f43] text-[15px] font-semibold">
                {{ eventSignupStatusLabel(event.signupStatus) }}
              </div>
              <p class="mt-2 text-[14px] text-[#787774] max-w-xl">
                {{ signupHint }}
              </p>
            </div>
            
            <button
              type="button"
              class="group px-6 py-3 rounded-lg shadow-sm font-black transition-all text-[15px] flex items-center justify-center gap-2 bg-[#111111] text-[#c4f000] hover:bg-[#222222] hover:-translate-y-0.5 shrink-0"
              @click="openSignupModal"
            >
              {{ signupPanelTitle }}
              <svg class="transition-transform group-hover:translate-x-1" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="5" y1="12" x2="19" y2="12"></line><polyline points="12 5 19 12 12 19"></polyline></svg>
            </button>
          </div>
        </div>

        <!-- Main Content -->
        <div class="prose prose-p:text-[#333333] prose-headings:text-[#111111] max-w-none leading-[1.6]">
          <p v-if="event.summary" class="text-[16px] text-[#787774] mb-10 font-medium italic border-l-4 border-[#c4f000] pl-4">
            {{ event.summary }}
          </p>
          
          <div v-if="renderedContent" class="markdown-entry" v-html="renderedContent"></div>
          <div v-else class="text-center text-[#787774] py-12">
            活动详细内容正在整理中。
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

    <!-- Signup Modal -->
    <div v-if="isModalOpen" class="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 backdrop-blur-sm p-4">
      <div class="bg-[#f6f5f1] w-full max-w-md rounded-2xl shadow-xl overflow-hidden border border-[#e3e2e0]">
        <!-- Modal Header -->
        <div class="flex justify-between items-center px-6 py-4 border-b border-[#e3e2e0]">
          <h3 class="text-lg font-bold text-[#111111] flex items-center gap-2">
            <span class="text-[#c4f000]">●</span> {{ signupPanelTitle }}
          </h3>
          <button @click="closeSignupModal" class="text-[#787774] hover:text-[#111111] transition-colors">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
          </button>
        </div>
        
        <!-- Modal Body -->
        <div class="p-6">
          <p class="text-[14px] text-[#787774] mb-6">{{ signupHint }}</p>

          <form v-if="canUseInternalSignup" class="grid gap-4" @submit.prevent="handleSubmit">
            <div class="space-y-4">
              <div>
                <label class="block text-[13px] font-bold text-[#333333] mb-1">姓名 <span class="text-red-500">*</span></label>
                <input
                  v-model="form.name"
                  type="text"
                  required
                  class="w-full rounded-xl border border-[#d7dde5] bg-white px-4 py-3 text-[#333333] outline-none focus:border-[#c4f000] focus:ring-1 focus:ring-[#c4f000] transition-shadow"
                  placeholder="你的姓名"
                />
              </div>
              <div>
                <label class="block text-[13px] font-bold text-[#333333] mb-1">微信号 <span class="text-red-500">*</span></label>
                <input
                  v-model="form.wechat"
                  type="text"
                  required
                  class="w-full rounded-xl border border-[#d7dde5] bg-white px-4 py-3 text-[#333333] outline-none focus:border-[#c4f000] focus:ring-1 focus:ring-[#c4f000] transition-shadow"
                  placeholder="方便我们联系到你的微信号"
                />
              </div>
              <div>
                <label class="block text-[13px] font-bold text-[#333333] mb-1">手机号 <span class="text-red-500">*</span></label>
                <input
                  v-model="form.phone"
                  type="tel"
                  required
                  inputmode="numeric"
                  maxlength="13"
                  class="w-full rounded-xl border border-[#d7dde5] bg-white px-4 py-3 text-[#333333] outline-none focus:border-[#c4f000] focus:ring-1 focus:ring-[#c4f000] transition-shadow"
                  placeholder="11 位手机号"
                  @blur="form.phone = normalizeMainlandPhone(form.phone) || form.phone"
                />
              </div>
              <div>
                <label class="block text-[13px] font-bold text-[#333333] mb-1">职位</label>
                <input
                  v-model="form.position"
                  type="text"
                  class="w-full rounded-xl border border-[#d7dde5] bg-white px-4 py-3 text-[#333333] outline-none focus:border-[#c4f000] focus:ring-1 focus:ring-[#c4f000] transition-shadow"
                  placeholder="你的职位或角色（选填）"
                />
              </div>
              <div>
                <label class="block text-[13px] font-bold text-[#333333] mb-1">行业信息 <span class="text-red-500">*</span></label>
                <input
                  v-model="form.industry"
                  type="text"
                  required
                  class="w-full rounded-xl border border-[#d7dde5] bg-white px-4 py-3 text-[#333333] outline-none focus:border-[#c4f000] focus:ring-1 focus:ring-[#c4f000] transition-shadow"
                  placeholder="你所在的行业或公司"
                />
              </div>
              <div>
                <label class="block text-[13px] font-bold text-[#333333] mb-1">备注</label>
                <textarea
                  v-model="form.message"
                  rows="3"
                  class="w-full rounded-xl border border-[#d7dde5] bg-white px-4 py-3 text-[#333333] outline-none focus:border-[#c4f000] focus:ring-1 focus:ring-[#c4f000] transition-shadow"
                  placeholder="参与目标、团队情况或想了解的问题（选填）"
                />
              </div>
            </div>

            <div v-if="submitError" class="rounded-xl border border-[#f3d3d6] bg-[#fff7f7] px-4 py-3 text-[14px] text-[#b42318]">
              {{ submitError }}
            </div>
            <div class="mt-2 flex justify-end gap-3">
              <button
                type="button"
                @click="closeSignupModal"
                class="px-4 py-2 rounded-lg font-bold transition-colors text-[14px] bg-[#e3e2e0] text-[#333333] hover:bg-[#d1d0ce]"
              >
                取消
              </button>
              <button
                type="submit"
                :disabled="submitting"
                class="px-5 py-2 rounded-lg shadow-sm font-black transition-colors text-[14px] flex items-center gap-2 bg-[#111111] text-[#c4f000] hover:bg-[#222222] disabled:opacity-60 disabled:cursor-not-allowed"
              >
                {{ submitting ? "提交中..." : "确认提交" }}
                <svg v-if="!submitting" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="5" y1="12" x2="19" y2="12"></line><polyline points="12 5 19 12 12 19"></polyline></svg>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <div v-if="showSuccessNotice" class="fixed inset-0 z-[110] flex items-center justify-center bg-black/20 backdrop-blur-[2px] p-4">
      <div class="w-full max-w-sm rounded-2xl border border-[#dfe6d4] bg-[#f6f5f1] p-6 text-center shadow-xl">
        <div class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-[#c4f000]/20 text-[#6f8f43]">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"></path></svg>
        </div>
        <h3 class="mt-4 text-[20px] font-black text-[#111111]">提交成功</h3>
        <p class="mt-2 text-[14px] leading-6 text-[#66625c]">
          {{ submitSuccess }}
        </p>
        <button
          type="button"
          class="mt-5 inline-flex items-center justify-center rounded-lg bg-[#111111] px-5 py-2.5 text-[14px] font-black text-[#c4f000] transition-colors hover:bg-[#222222]"
          @click="closeSuccessNotice"
        >
          回到活动详情
        </button>
      </div>
    </div>
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
