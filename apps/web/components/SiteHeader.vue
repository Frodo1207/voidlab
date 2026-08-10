<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue';
import greenLogo from '../../../assets/logo/green-bg.svg';
import whiteLogo from '../../../assets/logo/white-bg.svg';

const props = defineProps<{
  theme?: 'dark' | 'light';
  activePath?: string;
}>();

const isMobileMenuOpen = ref(false);
const isScrolled = ref(false);
const useGreenLogo = computed(() => props.activePath === '/' || props.activePath === '/builders');
const headerLogo = computed(() => (useGreenLogo.value ? greenLogo : whiteLogo));

const toggleMobileMenu = () => {
  isMobileMenuOpen.value = !isMobileMenuOpen.value;
  if (isMobileMenuOpen.value) {
    document.body.style.overflow = 'hidden';
  } else {
    document.body.style.overflow = '';
  }
};

const handleScroll = () => {
  isScrolled.value = window.scrollY > 20;
};

onMounted(() => {
  window.addEventListener('scroll', handleScroll);
});

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll);
  document.body.style.overflow = '';
});

const navLinks = [
  { name: '首页', path: '/', external: false },
  { name: '资讯', path: '/insights', external: true },
  { name: '知识库', path: '/knowledge', external: true },
  { name: '活动', path: '/events', external: true },
];
</script>

<template>
  <header
    class="fixed top-0 left-0 right-0 z-50 transition-all duration-300"
    :class="[
      props.theme === 'dark' 
        ? (isScrolled ? 'bg-[#050505]/80 border-b border-white/10 backdrop-blur-md' : 'bg-transparent border-b border-transparent')
        : (isScrolled ? 'bg-white/80 border-b border-gray-200 backdrop-blur-md' : 'bg-transparent border-b border-transparent'),
      isMobileMenuOpen && props.theme === 'dark' ? 'bg-[#050505] !border-transparent' : '',
      isMobileMenuOpen && props.theme === 'light' ? 'bg-white !border-transparent' : ''
    ]"
  >
    <div class="mx-auto max-w-[1440px] px-6 lg:px-14 h-16 flex items-center justify-between">
      <!-- Logo -->
      <a href="/" class="flex items-center gap-3 z-50" @click="isMobileMenuOpen && toggleMobileMenu()">
        <img
          :src="headerLogo"
          alt="VOIDLAB logo"
          width="32"
          height="32"
          class="h-8 w-8 rounded-[10px] object-cover shadow-sm"
        />
        <span 
          class="text-lg font-bold tracking-tight transition-colors"
          :class="props.theme === 'dark' ? 'text-white' : 'text-black'"
        >
          VOIDLAB
        </span>
      </a>

      <!-- Desktop Nav -->
      <nav class="hidden md:flex items-center gap-8">
        <a 
          v-for="link in navLinks" 
          :key="link.path"
          :href="link.path"
          :target="link.external ? '_blank' : '_self'"
          :rel="link.external ? 'noopener noreferrer' : ''"
          class="text-sm font-medium transition-colors relative py-1"
          :class="[
            props.theme === 'dark' 
              ? (activePath === link.path ? 'text-white' : 'text-white/60 hover:text-white')
              : (activePath === link.path ? 'text-black' : 'text-gray-500 hover:text-black')
          ]"
        >
          {{ link.name }}
          <span 
            v-if="activePath === link.path"
            class="absolute -bottom-[21px] left-0 right-0 h-[2px]"
            :class="props.theme === 'dark' ? 'bg-[#42ffd1]' : 'bg-black'"
          ></span>
        </a>
      </nav>

      <!-- Right Actions (Desktop) -->
      <div class="hidden md:flex items-center gap-4">
        <a 
          href="/contact"
          class="text-xs font-bold px-4 py-2 rounded-full transition-all duration-300"
          :class="props.theme === 'dark' ? 'bg-white/10 text-white hover:bg-white/20' : 'bg-black/5 text-black hover:bg-black/10'"
        >
          联系我们
        </a>
      </div>

      <!-- Mobile Menu Button -->
      <button 
        class="md:hidden z-50 p-2 -mr-2 flex flex-col justify-center items-center w-10 h-10 gap-1.5"
        @click="toggleMobileMenu"
        :aria-label="isMobileMenuOpen ? '关闭菜单' : '打开菜单'"
      >
        <span 
          class="w-5 h-[2px] transition-all duration-300 rounded-full"
          :class="[
            props.theme === 'dark' ? 'bg-white' : 'bg-black',
            isMobileMenuOpen ? 'rotate-45 translate-y-[8px]' : ''
          ]"
        ></span>
        <span 
          class="w-5 h-[2px] transition-all duration-300 rounded-full"
          :class="[
            props.theme === 'dark' ? 'bg-white' : 'bg-black',
            isMobileMenuOpen ? 'opacity-0' : ''
          ]"
        ></span>
        <span 
          class="w-5 h-[2px] transition-all duration-300 rounded-full"
          :class="[
            props.theme === 'dark' ? 'bg-white' : 'bg-black',
            isMobileMenuOpen ? '-rotate-45 -translate-y-[8px]' : ''
          ]"
        ></span>
      </button>
    </div>

  </header>

  <!-- Keep the mobile overlay outside the header so sticky/backdrop states
       on the header do not interfere with menu rendering on mobile browsers. -->
  <div
    class="fixed inset-0 top-16 z-40 transition-all duration-500 ease-[cubic-bezier(0.25,1,0.5,1)] md:hidden overflow-hidden flex flex-col"
    :class="[
      isMobileMenuOpen ? 'opacity-100 translate-y-0' : 'opacity-0 -translate-y-4 pointer-events-none',
      props.theme === 'dark' ? 'bg-[#050505]' : 'bg-white'
    ]"
  >
    <nav class="flex flex-col px-6 py-8 gap-6 flex-1">
      <a
        v-for="(link, index) in navLinks"
        :key="link.path"
        :href="link.path"
        :target="link.external ? '_blank' : '_self'"
        :rel="link.external ? 'noopener noreferrer' : ''"
        class="text-2xl font-bold tracking-tight transition-transform duration-300"
        :class="[
          props.theme === 'dark'
            ? (activePath === link.path ? 'text-[#42ffd1]' : 'text-white')
            : (activePath === link.path ? 'text-black' : 'text-gray-500'),
          isMobileMenuOpen ? 'translate-y-0 opacity-100' : 'translate-y-4 opacity-0'
        ]"
        :style="{ transitionDelay: `${index * 50}ms` }"
        @click="toggleMobileMenu"
      >
        {{ link.name }}
      </a>
    </nav>

    <div
      class="px-6 pb-12 transition-all duration-500 delay-300"
      :class="isMobileMenuOpen ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'"
    >
      <a
        href="/contact"
        class="flex items-center justify-center w-full py-4 rounded-xl font-bold text-sm"
        :class="props.theme === 'dark' ? 'bg-white text-black' : 'bg-black text-white'"
        @click="toggleMobileMenu"
      >
        联系我们
      </a>
    </div>
  </div>
</template>
