<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import SiteHeader from "../components/SiteHeader.vue";
import { submitContactLead } from "../composables/useLeadIntake";
import { useSiteConfigs } from "../composables/useSiteConfigs";

const { contactChannels, loadSiteConfigs } = useSiteConfigs();

onMounted(async () => {
  document.title = "联系我们 | VOIDLAB";
  try {
    await loadSiteConfigs();
  } catch {
    // Keep built-in defaults when configs are unavailable.
  }
});

const submitting = ref(false);
const submitSuccess = ref("");
const submitError = ref("");
const form = reactive({
  name: "",
  contact: "",
  message: ""
});

const contactIcons = [
  `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="2" width="16" height="20" rx="2" ry="2"></rect><line x1="12" y1="18" x2="12.01" y2="18"></line></svg>`,
  `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>`,
  `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path><polyline points="22,6 12,13 2,6"></polyline></svg>`,
  `<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path></svg>`
];

const contacts = computed(() =>
  contactChannels.value.map((item, index) => ({
    ...item,
    btnText: item.buttonText,
    icon: contactIcons[index % contactIcons.length]
  }))
);

async function handleSubmit() {
  submitting.value = true;
  submitError.value = "";
  submitSuccess.value = "";

  try {
    await submitContactLead({
      name: form.name,
      contact: form.contact,
      message: form.message
    });

    submitSuccess.value = "已提交，我们会尽快联系你。";
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
  <div class="relative z-10 min-h-screen bg-[#08080a] text-white font-sans selection:bg-[#42ffd1] selection:text-black">
    <SiteHeader theme="dark" activePath="/contact" />

    <div class="mx-auto max-w-[1440px] px-6 lg:px-14 pt-32 pb-24">
      
      <!-- Top Title Area -->
      <div class="flex flex-col lg:flex-row lg:items-end justify-between gap-12 mb-20 border-b border-white/10 pb-16">
        <div>
          <div class="text-[#42ffd1] font-mono text-sm tracking-widest mb-6 uppercase">联系我们</div>
          <h1 class="text-5xl md:text-6xl lg:text-[80px] font-bold tracking-tight">联系 VOID LAB。</h1>
        </div>
        <p class="text-gray-400 text-lg max-w-lg leading-relaxed">
          想加入社区、合作办活动、赞助支持、学校/社区合作，或咨询 VOID LAB，都可以先从这里联系。
        </p>
      </div>

      <!-- 4 Columns Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-0 border-l border-white/10">
        <div 
          v-for="(item, index) in contacts" 
          :key="index"
          class="border-r border-b border-white/10 p-10 xl:p-12 flex flex-col group hover:bg-[#111218] transition-colors"
        >
          <!-- Icon -->
          <div class="w-14 h-14 rounded-full border border-white/20 flex items-center justify-center text-gray-400 mb-12 group-hover:text-white group-hover:border-white/40 transition-colors" v-html="item.icon"></div>
          
          <!-- Content -->
          <h3 class="text-2xl font-bold mb-4">{{ item.title }}</h3>
          <p class="text-gray-400 leading-relaxed mb-8 flex-1">{{ item.desc }}</p>
          
          <!-- Bottom Area -->
          <div class="mt-auto">
            <div class="font-mono text-xs text-gray-500 mb-6 uppercase tracking-wider">{{ item.account }}</div>
            <a 
              :href="item.link" 
              class="inline-block px-6 py-2.5 rounded-full border border-white/20 text-sm font-bold hover:bg-white hover:text-black transition-all"
            >
              {{ item.btnText }}
            </a>
          </div>
        </div>
      </div>

      <section class="mt-20 grid grid-cols-1 lg:grid-cols-[1fr_1.1fr] gap-10 items-start">
        <div class="border border-white/10 rounded-3xl p-8 md:p-10 bg-[#0d0f14]">
          <div class="text-[#42ffd1] font-mono text-xs tracking-[0.25em] uppercase mb-4">Direct Intake</div>
          <h2 class="text-3xl md:text-4xl font-bold tracking-tight">直接提交合作或咨询意向</h2>
          <p class="mt-5 text-gray-400 leading-relaxed">
            如果你已经有明确需求，不用再跳转邮箱。直接在这里留下姓名、联系方式和诉求，系统会自动进入 VOIDLAB 的 Leads 跟进池。
          </p>
          <div class="mt-8 grid gap-4 text-sm text-gray-400">
            <div class="flex items-start gap-3">
              <span class="text-[#42ffd1] font-mono">01</span>
              <span>适合活动合作、赞助支持、学校社群联动、媒体采访和正式咨询。</span>
            </div>
            <div class="flex items-start gap-3">
              <span class="text-[#42ffd1] font-mono">02</span>
              <span>提交后会以 `contact` 来源进入后台 Leads，方便团队统一跟进。</span>
            </div>
          </div>
        </div>

        <form class="border border-white/10 rounded-3xl p-8 md:p-10 bg-[#111218]" @submit.prevent="handleSubmit">
          <div class="grid gap-5">
            <div>
              <label class="block text-sm font-mono uppercase tracking-[0.2em] text-gray-500 mb-3">姓名</label>
              <input
                v-model="form.name"
                type="text"
                required
                class="w-full rounded-2xl border border-white/10 bg-[#0b0c11] px-4 py-3 text-white outline-none transition-colors focus:border-[#42ffd1]/60"
                placeholder="怎么称呼你"
              />
            </div>

            <div>
              <label class="block text-sm font-mono uppercase tracking-[0.2em] text-gray-500 mb-3">联系方式</label>
              <input
                v-model="form.contact"
                type="text"
                required
                class="w-full rounded-2xl border border-white/10 bg-[#0b0c11] px-4 py-3 text-white outline-none transition-colors focus:border-[#42ffd1]/60"
                placeholder="微信 / 邮箱 / 手机号"
              />
            </div>

            <div>
              <label class="block text-sm font-mono uppercase tracking-[0.2em] text-gray-500 mb-3">需求说明</label>
              <textarea
                v-model="form.message"
                required
                rows="6"
                class="w-full rounded-2xl border border-white/10 bg-[#0b0c11] px-4 py-3 text-white outline-none transition-colors focus:border-[#42ffd1]/60"
                placeholder="简单说下你想合作什么、想咨询什么，或者想从 VOIDLAB 获得什么帮助"
              />
            </div>

            <div v-if="submitError" class="rounded-2xl border border-[#7f1d1d] bg-[#2a1116] px-4 py-3 text-sm text-[#fca5a5]">
              {{ submitError }}
            </div>

            <div v-if="submitSuccess" class="rounded-2xl border border-[#14532d] bg-[#0f2217] px-4 py-3 text-sm text-[#86efac]">
              {{ submitSuccess }}
            </div>

            <button
              type="submit"
              :disabled="submitting"
              class="inline-flex items-center justify-center rounded-2xl bg-[#42ffd1] px-6 py-3 text-sm font-bold text-black transition-all hover:bg-white disabled:cursor-not-allowed disabled:opacity-60"
            >
              {{ submitting ? "提交中..." : "提交意向" }}
            </button>
          </div>
        </form>
      </section>

    </div>
  </div>
</template>
