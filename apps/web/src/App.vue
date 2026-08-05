<script setup lang="ts">
import { computed, watchEffect } from "vue";
import { useRoute } from "vue-router";
import SpaceBackground from "../components/SpaceBackground.vue";
import greenLogo from "../../../assets/logo/green-bg.svg";
import whiteLogo from "../../../assets/logo/white-bg.svg";

const route = useRoute();

const isLightTheme = computed(() => route.path.startsWith("/knowledge") || route.path.startsWith("/insights"));
const showSpaceBackground = computed(() => !isLightTheme.value);
const shouldUseGreenLogo = computed(() => route.path === "/" || route.path.startsWith("/builders"));

function ensureThemeColorMeta(): HTMLMetaElement {
  let meta = document.querySelector("meta[name='theme-color']") as HTMLMetaElement | null;
  if (!meta) {
    meta = document.createElement("meta");
    meta.name = "theme-color";
    document.head.appendChild(meta);
  }
  return meta;
}

function ensureIconLink(rel: string): HTMLLinkElement {
  let link = document.querySelector(`link[rel='${rel}']`) as HTMLLinkElement | null;
  if (!link) {
    link = document.createElement("link");
    link.rel = rel;
    document.head.appendChild(link);
  }
  return link;
}

watchEffect(() => {
  const light = isLightTheme.value;
  const favicon = shouldUseGreenLogo.value ? greenLogo : whiteLogo;

  if (light) {
    document.body.classList.add("theme-light");
  } else {
    document.body.classList.remove("theme-light");
  }

  const meta = ensureThemeColorMeta();
  meta.content = light ? "#fbfbfa" : "#121212";

  const icon = ensureIconLink("icon");
  icon.href = favicon;
  icon.type = "image/svg+xml";

  const shortcutIcon = ensureIconLink("shortcut icon");
  shortcutIcon.href = favicon;
  shortcutIcon.type = "image/svg+xml";

  const appleTouchIcon = ensureIconLink("apple-touch-icon");
  appleTouchIcon.href = favicon;
  appleTouchIcon.type = "image/svg+xml";
});
</script>

<template>
  <div>
    <SpaceBackground v-if="showSpaceBackground" />
    <RouterView />
  </div>
</template>
