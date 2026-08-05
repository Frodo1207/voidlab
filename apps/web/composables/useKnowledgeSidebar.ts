import { computed, ref } from "vue";

const STORAGE_KEY = "voidlab-knowledge-sidebar-collapsed";

const sidebarCollapsedState = ref(false);
const sidebarDrawerOpenState = ref(false);
let sidebarPreferenceLoaded = false;
let viewportVarInstalled = false;

function ensureSidebarPreferenceLoaded() {
  if (sidebarPreferenceLoaded || typeof window === "undefined") {
    return;
  }

  sidebarPreferenceLoaded = true;

  try {
    sidebarCollapsedState.value = window.localStorage.getItem(STORAGE_KEY) === "1";
  } catch {
    sidebarCollapsedState.value = false;
  }
}

function persistSidebarPreference() {
  if (typeof window === "undefined") {
    return;
  }

  try {
    window.localStorage.setItem(STORAGE_KEY, sidebarCollapsedState.value ? "1" : "0");
  } catch {
    // ignore localStorage errors
  }
}

function ensureViewportHeightVarInstalled() {
  if (viewportVarInstalled) {
    return;
  }
  viewportVarInstalled = true;

  if (typeof window === "undefined" || typeof document === "undefined") {
    return;
  }

  const update = () => {
    const vv = window.visualViewport;
    const height = vv?.height ?? window.innerHeight;
    document.documentElement.style.setProperty("--vvh", `${height * 0.01}px`);
  };

  update();
  window.addEventListener("resize", update, { passive: true });
  window.addEventListener("orientationchange", update, { passive: true });
  if (window.visualViewport) {
    window.visualViewport.addEventListener("resize", update, { passive: true });
    window.visualViewport.addEventListener("scroll", update, { passive: true });
  }
}

export function useKnowledgeSidebar() {
  ensureSidebarPreferenceLoaded();
  ensureViewportHeightVarInstalled();

  function setSidebarCollapsed(value: boolean) {
    sidebarCollapsedState.value = value;
    persistSidebarPreference();
  }

  function toggleSidebar() {
    setSidebarCollapsed(!sidebarCollapsedState.value);
  }

  function openSidebarDrawer() {
    sidebarDrawerOpenState.value = true;
    // Safari sometimes updates visual viewport height only after first scroll.
    // Force-refresh the CSS var when opening to avoid the initial gray gap.
    if (typeof window !== "undefined" && typeof document !== "undefined") {
      const vv = window.visualViewport;
      const height = vv?.height ?? window.innerHeight;
      document.documentElement.style.setProperty("--vvh", `${height * 0.01}px`);
      window.setTimeout(() => {
        const vv2 = window.visualViewport;
        const height2 = vv2?.height ?? window.innerHeight;
        document.documentElement.style.setProperty("--vvh", `${height2 * 0.01}px`);
      }, 0);
    }
  }

  function closeSidebarDrawer() {
    sidebarDrawerOpenState.value = false;
  }

  function toggleSidebarDrawer() {
    sidebarDrawerOpenState.value = !sidebarDrawerOpenState.value;
  }

  return {
    sidebarCollapsed: computed(() => sidebarCollapsedState.value),
    setSidebarCollapsed,
    toggleSidebar,
    sidebarDrawerOpen: computed(() => sidebarDrawerOpenState.value),
    openSidebarDrawer,
    closeSidebarDrawer,
    toggleSidebarDrawer
  };
}
