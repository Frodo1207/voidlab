<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { submitEventCheckin } from "../composables/useLeadIntake";
import { useEventArchive } from "../composables/useEventArchive";
import { normalizeMainlandPhone } from "../src/contactValidation";

const route = useRoute();
const router = useRouter();
const { loadEventBySlug } = useEventArchive();

const slug = computed(() => String(route.params.slug ?? "").trim());
const event = ref<Awaited<ReturnType<typeof loadEventBySlug>>>(null);
const loading = ref(true);
const pageError = ref("");
const submitting = ref(false);
const submitSuccess = ref("");
const submitError = ref("");
const mobileTouched = ref(false);
let redirectTimer: number | null = null;

const form = reactive({
  name: "",
  mobile: "",
  company: "",
  position: ""
});

const mobileFieldError = computed(() => {
  if (!mobileTouched.value || !form.mobile.trim()) {
    return "";
  }

  return normalizeMainlandPhone(form.mobile) ? "" : "请输入有效的 11 位手机号。";
});

watch(
  slug,
  async (nextSlug) => {
    if (!nextSlug) {
      router.replace("/events");
      return;
    }

    loading.value = true;
    pageError.value = "";

    try {
      const detail = await loadEventBySlug(nextSlug);
      if (!detail) {
        router.replace("/events");
        return;
      }
      event.value = detail;
      document.title = `${detail.title} 签到 | VOIDLAB`;
    } catch (error) {
      pageError.value = error instanceof Error ? error.message : "加载活动失败";
    } finally {
      loading.value = false;
    }
  },
  { immediate: true }
);

function clearRedirectTimer() {
  if (redirectTimer !== null) {
    window.clearTimeout(redirectTimer);
    redirectTimer = null;
  }
}

async function handleSubmit() {
  if (!event.value) {
    return;
  }

  mobileTouched.value = true;
  submitError.value = "";
  submitSuccess.value = "";
  clearRedirectTimer();

  const normalizedMobile = normalizeMainlandPhone(form.mobile);
  if (!normalizedMobile) {
    submitError.value = "请输入有效的 11 位手机号。";
    return;
  }

  submitting.value = true;

  try {
    await submitEventCheckin(event.value.id, {
      name: form.name.trim(),
      mobile: normalizedMobile,
      company: form.company.trim(),
      position: form.position.trim()
    });

    submitSuccess.value = event.value.signupSuccessMessage?.trim() || "签到成功！";
    form.name = "";
    form.mobile = "";
    form.company = "";
    form.position = "";
    mobileTouched.value = false;
    redirectTimer = window.setTimeout(() => {
      void router.replace("/");
    }, 1500);
  } catch (error) {
    submitError.value = error instanceof Error ? error.message : "提交失败，请稍后重试";
  } finally {
    submitting.value = false;
  }
}

onBeforeUnmount(() => {
  clearRedirectTimer();
});
</script>

<template>
  <div class="relative z-10 min-h-screen bg-[#0a0a0d] px-4 py-5 text-white sm:px-6">
    <div class="mx-auto w-full max-w-[520px]">
      <div class="mb-5 text-center">
        <router-link to="/" class="inline-flex items-center justify-center text-sm font-semibold tracking-[0.2em] text-white/45">
          VOIDLAB
        </router-link>
      </div>

      <div v-if="loading" class="rounded-3xl border border-white/10 bg-[#111218] px-5 py-10 text-center text-sm text-white/60">
        正在加载活动信息...
      </div>

      <div
        v-else-if="pageError || !event"
        class="rounded-3xl border border-[#7f1d1d] bg-[#201215] px-5 py-10 text-center text-sm text-[#fca5a5]"
      >
        {{ pageError || "活动不存在或暂时不可用。" }}
      </div>

      <div v-else class="space-y-4">
        <section class="rounded-3xl border border-white/10 bg-[#111218] px-5 py-6">
          <div class="text-[11px] uppercase tracking-[0.22em] text-[#42ffd1]">活动签到</div>
          <h1 class="mt-3 text-2xl font-bold leading-tight">{{ event.title }}</h1>
          <div class="mt-4 space-y-2 text-sm text-white/65">
            <p v-if="event.rawTime">时间：{{ event.time }}</p>
            <p>地点：{{ event.location || event.city || "待定" }}</p>
          </div>
        </section>

        <div
          v-if="submitSuccess"
          class="rounded-3xl border border-[#14532d] bg-[#0f2217] px-5 py-10 text-center"
        >
          <div class="text-2xl font-bold text-[#86efac]">签到成功！</div>
          <p class="mt-3 text-sm text-white/70">信息已提交，正在返回首页...</p>
        </div>

        <form v-else class="rounded-3xl border border-white/10 bg-[#111218] px-5 py-6" @submit.prevent="handleSubmit">
          <div class="space-y-4">
            <div>
              <label class="mb-2 block text-sm text-white/72">姓名</label>
              <input
                v-model="form.name"
                type="text"
                required
                class="w-full rounded-2xl border border-white/10 bg-[#0b0c11] px-4 py-3 text-base text-white outline-none transition-colors focus:border-[#42ffd1]/60"
                placeholder="请输入姓名"
              />
            </div>

            <div>
              <label class="mb-2 block text-sm text-white/72">手机号</label>
              <input
                v-model="form.mobile"
                type="tel"
                required
                autocomplete="tel-national"
                inputmode="numeric"
                maxlength="13"
                :class="[
                  'w-full rounded-2xl bg-[#0b0c11] px-4 py-3 text-base text-white outline-none transition-colors',
                  mobileFieldError ? 'border border-[#7f1d1d] focus:border-[#f87171]' : 'border border-white/10 focus:border-[#42ffd1]/60'
                ]"
                placeholder="请输入 11 位手机号"
                @blur="
                  mobileTouched = true;
                  form.mobile = normalizeMainlandPhone(form.mobile) || form.mobile.trim();
                "
              />
              <p v-if="mobileFieldError" class="mt-2 text-sm text-[#fca5a5]">{{ mobileFieldError }}</p>
            </div>

            <div>
              <label class="mb-2 block text-sm text-white/72">公司</label>
              <input
                v-model="form.company"
                type="text"
                required
                class="w-full rounded-2xl border border-white/10 bg-[#0b0c11] px-4 py-3 text-base text-white outline-none transition-colors focus:border-[#42ffd1]/60"
                placeholder="请输入公司名称"
              />
            </div>

            <div>
              <label class="mb-2 block text-sm text-white/72">岗位</label>
              <input
                v-model="form.position"
                type="text"
                required
                class="w-full rounded-2xl border border-white/10 bg-[#0b0c11] px-4 py-3 text-base text-white outline-none transition-colors focus:border-[#42ffd1]/60"
                placeholder="请输入岗位"
              />
            </div>

            <div v-if="submitError" class="rounded-2xl border border-[#7f1d1d] bg-[#2a1116] px-4 py-3 text-sm text-[#fca5a5]">
              {{ submitError }}
            </div>

            <button
              type="submit"
              :disabled="submitting"
              class="w-full rounded-2xl bg-[#42ffd1] px-5 py-3 text-base font-semibold text-black transition-colors hover:bg-[#7effde] disabled:cursor-not-allowed disabled:opacity-60"
            >
              {{ submitting ? "提交中..." : "完成签到" }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>
