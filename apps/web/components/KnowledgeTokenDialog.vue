<script setup lang="ts">
import { computed, ref, watch } from "vue";

const props = defineProps<{
  open: boolean;
  spaceTitle: string;
  hint: string;
  errorMessage?: string;
  demoToken?: string;
}>();

const emit = defineEmits<{
  (event: "update:open", value: boolean): void;
  (event: "submit", token: string): void;
}>();

const token = ref("");

const visible = computed({
  get: () => props.open,
  set: (value: boolean) => emit("update:open", value)
});

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      token.value = "";
    }
  }
);

function handleSubmit() {
  emit("submit", token.value);
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="fixed inset-0 z-[70] flex items-center justify-center bg-black/20 px-4 backdrop-blur-sm"
      @click.self="visible = false"
    >
      <div class="w-full max-w-[440px] rounded-lg border border-[#e9e9e7] bg-white p-6 shadow-xl md:p-8">
        <div class="flex items-start justify-between gap-4">
          <div>
            <div class="text-[12px] font-medium uppercase tracking-wider text-[#9ca3af]">
              Unlock Space
            </div>
            <h3 class="mt-2 text-xl font-bold leading-tight text-[#37352f]">
              解锁 {{ spaceTitle }}
            </h3>
            <p class="mt-2 text-[14px] leading-relaxed text-[#787774]">
              {{ hint }}
            </p>
          </div>
          <button
            type="button"
            class="text-[#9ca3af] transition-colors hover:text-[#37352f]"
            @click="visible = false"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
          </button>
        </div>

        <div class="mt-6">
          <label class="block text-[13px] font-medium text-[#787774] mb-2">访问令牌</label>
          <input
            v-model="token"
            type="text"
            class="w-full rounded border border-[#e9e9e7] bg-[#f7f7f5] px-3 py-2 text-[15px] text-[#37352f] outline-none transition-colors focus:border-[#0f7b6c] focus:bg-white"
            placeholder="输入令牌..."
            @keyup.enter="handleSubmit"
          />
          <p v-if="errorMessage" class="mt-2 text-[13px] text-[#eb5757]">
            {{ errorMessage }}
          </p>
          <p v-else-if="demoToken" class="mt-2 text-[12px] text-[#9ca3af]">
            演示环境可输入：<code class="rounded bg-[#f1f1ef] px-1.5 py-0.5 text-[#eb5757] font-mono">{{ demoToken }}</code>
          </p>
        </div>

        <div class="mt-8 flex items-center justify-end gap-3">
          <button
            type="button"
            class="rounded px-4 py-2 text-[14px] font-medium text-[#787774] transition-colors hover:bg-gray-50"
            @click="visible = false"
          >
            取消
          </button>
          <button
            type="button"
            class="rounded bg-[#0f7b6c] px-4 py-2 text-[14px] font-medium text-white transition-colors hover:bg-[#0d685b]"
            @click="handleSubmit"
          >
            确认解锁
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>
