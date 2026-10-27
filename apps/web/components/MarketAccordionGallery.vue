<script setup lang="ts">
import { ref, watch } from "vue";
import { RouterLink } from "vue-router";

export type MarketAccordionGalleryItem = {
  id: string | number;
  title: string;
  summary: string;
  meta: string;
  tags: string[];
  href: string;
  image: string;
  image_alt?: string;
};

const props = withDefaults(defineProps<{
  items: MarketAccordionGalleryItem[];
  defaultIndex?: number;
}>(), {
  defaultIndex: 0
});

const getDefaultIndex = () =>
  Math.max(0, Math.min(props.defaultIndex, Math.max(0, props.items.length - 1)));

const activeIndex = ref(getDefaultIndex());
watch(() => props.items.length, (len) => {
  if (!len) activeIndex.value = 0;
  else if (activeIndex.value > len - 1) activeIndex.value = len - 1;
});

const setActive = (index: number) => {
  activeIndex.value = index;
};

const resetActive = () => {
  activeIndex.value = getDefaultIndex();
};
</script>

<template>
  <div class="ag-root" role="list" aria-label="热门项目展示" @mouseleave="resetActive">
    <RouterLink
      v-for="(item, index) in items"
      :key="item.id"
      class="ag-panel"
      :class="{ active: index === activeIndex }"
      :to="item.href"
      role="listitem"
      @mouseenter="setActive(index)"
      @focus="setActive(index)"
      @click="setActive(index)"
    >
      <img class="ag-image" :src="item.image" :alt="item.image_alt || item.title" loading="lazy" decoding="async" />
      <div class="ag-overlay"></div>
      <div class="ag-content">
        <div class="ag-kicker">
          <span>0{{ index + 1 }}</span>
          <span>{{ item.meta }}</span>
        </div>
        <h3 class="ag-title">{{ item.title }}</h3>
        <p class="ag-summary">{{ item.summary }}</p>
        <div class="ag-tags">
          <span v-for="tag in item.tags.slice(0, 4)" :key="tag">{{ tag }}</span>
        </div>
      </div>
    </RouterLink>
  </div>
</template>

<style scoped>
.ag-root {
  display: flex;
  gap: 14px;
  width: 100%;
  height: 460px;
}
.ag-panel {
  position: relative;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  border-radius: 18px;
  text-decoration: none;
  color: #F8FAFC;
  outline: none;
  transform: translateZ(0);
  transition: flex 0.55s cubic-bezier(.2, .75, .1, 1), transform 0.35s ease;
}
.ag-panel.active {
  flex: 3.2;
}
.ag-panel:focus-visible {
  box-shadow: 0 0 0 3px rgba(11, 203, 120, 0.22);
}
.ag-image {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform: scale(1.06);
  transition: transform 0.8s cubic-bezier(.2, .75, .1, 1);
  filter: saturate(0.9) contrast(1.02);
}
.ag-panel.active .ag-image {
  transform: scale(1.1);
}
.ag-overlay {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(90deg, rgba(9, 9, 11, 0.18), rgba(9, 9, 11, 0.52) 48%, rgba(9, 9, 11, 0.78)),
    linear-gradient(0deg, rgba(9, 9, 11, 0.9), transparent 55%);
}
.ag-content {
  position: absolute;
  inset: 0;
  padding: 22px 22px 20px;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  gap: 10px;
}
.ag-kicker {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  font: 10px ui-monospace, SFMono-Regular, Menlo, monospace;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: rgba(248, 250, 252, 0.82);
}
.ag-kicker span:first-child {
  color: #0BCB78;
}
.ag-title {
  margin: 0;
  font-size: 22px;
  letter-spacing: -0.03em;
  line-height: 1.25;
  font-weight: 800;
  color: #FFFFFF;
  max-width: 52ch;
  text-wrap: balance;
}
.ag-summary {
  margin: 0;
  color: rgba(248, 250, 252, 0.82);
  font-size: 13px;
  line-height: 1.7;
  max-width: 64ch;
  display: -webkit-box;
  overflow: hidden;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
  opacity: 0;
  max-height: 0;
  transition: opacity 0.35s ease, max-height 0.55s cubic-bezier(.2, .75, .1, 1);
}
.ag-panel.active .ag-summary {
  opacity: 1;
  max-height: 90px;
}
.ag-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  opacity: 0;
  max-height: 0;
  overflow: hidden;
  transition: opacity 0.35s ease, max-height 0.55s cubic-bezier(.2, .75, .1, 1);
}
.ag-panel.active .ag-tags {
  opacity: 1;
  max-height: 40px;
}
.ag-tags span {
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  padding: 6px 10px;
  border-radius: 999px;
  border: 1px solid rgba(248, 250, 252, 0.24);
  background: rgba(9, 9, 11, 0.22);
  color: rgba(248, 250, 252, 0.86);
  backdrop-filter: blur(6px);
}

@media (max-width: 980px) {
  .ag-root {
    height: auto;
    flex-direction: column;
  }
  .ag-panel,
  .ag-panel.active {
    flex: none;
  }
  .ag-panel {
    height: 260px;
  }
  .ag-summary,
  .ag-tags {
    opacity: 1;
    max-height: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .ag-panel,
  .ag-image,
  .ag-summary,
  .ag-tags {
    transition: none;
  }
}
</style>
