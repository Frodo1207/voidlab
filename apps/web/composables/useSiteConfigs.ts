import { computed, ref } from "vue";
import { resolvePublicApiPath } from "../src/runtimeConfig";

export type HomeBannerConfig = {
  titleText: string;
  subtitle: string;
  primaryCtaLabel: string;
  primaryCtaPath: string;
  secondaryCtaLabel: string;
  secondaryCtaPath: string;
  statusLabel: string;
};

export type HomeFeaturedConfig = {
  communityCount: string;
  communityCountSuffix: string;
  eventCount: string;
  eventCountSuffix: string;
  eventsDescription: string;
  buildersDescription: string;
  insightsDescription: string;
};

export type ContactChannelConfig = {
  title: string;
  desc: string;
  account: string;
  buttonText: string;
  link: string;
};

export type FooterNavItemConfig = {
  label: string;
  path: string;
};

export type FooterConfig = {
  slogan: string;
  navLinks: FooterNavItemConfig[];
  legalText: string;
};

export type GlobalCtaConfig = {
  eyebrow: string;
  title: string;
  description: string;
  primaryLabel: string;
  primaryPath: string;
  secondaryLabel: string;
  secondaryPath: string;
};

export type FeaturedContentSlotsConfig = {
  eventsTitle: string;
  eventsViewAllLabel: string;
  eventsLimit: number;
  buildersTitle: string;
  buildersViewAllLabel: string;
  buildersLimit: number;
  insightsTitle: string;
  insightsViewAllLabel: string;
  insightsLimit: number;
};

const defaultHomeBanner: HomeBannerConfig = {
  titleText: "探索 AI 的边界",
  subtitle: "我们构建 AI 资产，不做玩具。",
  primaryCtaLabel: "连接社交网络",
  primaryCtaPath: "/builders",
  secondaryCtaLabel: "查看社区活动",
  secondaryCtaPath: "/events",
  statusLabel: "系统状态 // 社区指标"
};

const defaultHomeFeatured: HomeFeaturedConfig = {
  communityCount: "1000",
  communityCountSuffix: "+",
  eventCount: "10",
  eventCountSuffix: "+",
  eventsDescription: "",
  buildersDescription: "社区里不只有活动和内容，还有一批可以被连接、被协作，也能直接发起合作的人。",
  insightsDescription: "剥离噪音，只提供值得关注的行业信号与落地经验。"
};

const defaultContactChannels: ContactChannelConfig[] = [
  {
    title: "官方小红书",
    desc: "关注 VOID LAB 的活动回顾、现场照片、社区动态和新活动发布。",
    account: "小红书号 VOIDLAB_AI",
    buttonText: "关注小红书",
    link: "#"
  },
  {
    title: "加入社区群",
    desc: "加入 VOID LAB 微信群，获取活动通知、结识伙伴和线下聚会信息。",
    account: "企业微信入群二维码",
    buttonText: "查看二维码",
    link: "#"
  },
  {
    title: "官方邮箱",
    desc: "合作、赞助、学校/社区组织、媒体和正式事项联系。",
    account: "JOIN@VOIDLAB.AI",
    buttonText: "发邮件",
    link: "mailto:join@voidlab.ai"
  },
  {
    title: "官方小助手",
    desc: "咨询、报名、入群、活动问题和成员对接支持。",
    account: "企业微信客服链接",
    buttonText: "添加小助手",
    link: "#"
  }
];

const defaultFooterConfig: FooterConfig = {
  slogan: "连接共创者，重塑 AI 资产。",
  navLinks: [
    { label: "首页", path: "/" },
    { label: "活动", path: "/events" },
    { label: "社交网络", path: "/builders" },
    { label: "资讯", path: "/insights" },
    { label: "联系我们", path: "/contact" }
  ],
  legalText: "VOIDLAB.AI © 2026. All rights reserved."
};

const defaultGlobalCta: GlobalCtaConfig = {
  eyebrow: "NEXT ACTION",
  title: "准备把想法变成活动、内容或合作项目？",
  description: "如果你想和 VOIDLAB 一起办活动、发起合作、加入网络或咨询 AI 落地方案，可以直接从这里进入下一步。",
  primaryLabel: "提交合作意向",
  primaryPath: "/contact",
  secondaryLabel: "查看活动",
  secondaryPath: "/events"
};

const defaultFeaturedContentSlots: FeaturedContentSlotsConfig = {
  eventsTitle: "社区活动",
  eventsViewAllLabel: "查看全部 [全部活动]",
  eventsLimit: 5,
  buildersTitle: "社交网络",
  buildersViewAllLabel: "查看全部网络 [全部成员]",
  buildersLimit: 6,
  insightsTitle: "资讯中心",
  insightsViewAllLabel: "查看全部资讯 [全部内容]",
  insightsLimit: 5
};

type SiteConfigMap = {
  home_banner?: Partial<HomeBannerConfig>;
  home_featured?: Partial<HomeFeaturedConfig>;
  contact_channels?: ContactChannelConfig[];
  footer_config?: Partial<FooterConfig>;
  global_cta?: Partial<GlobalCtaConfig>;
  featured_content_slots?: Partial<FeaturedContentSlotsConfig>;
};

type ApiEnvelope<T> = {
  code: number;
  message: string;
  data: T;
};

const siteConfigState = ref<SiteConfigMap>({});
const loadingState = ref(false);
const loadedState = ref(false);
const errorState = ref("");

let loadPromise: Promise<SiteConfigMap> | null = null;

async function requestPublicConfigs() {
  const response = await fetch(resolvePublicApiPath("/api/v1/public/site-configs"));
  const envelope = (await response.json()) as ApiEnvelope<SiteConfigMap>;

  if (!response.ok || envelope.code !== 0) {
    throw new Error(envelope.message || "加载站点配置失败");
  }

  return envelope.data;
}

export async function loadSiteConfigs(force = false) {
  if (loadedState.value && !force) {
    return siteConfigState.value;
  }

  if (loadPromise && !force) {
    return loadPromise;
  }

  loadingState.value = true;
  errorState.value = "";

  loadPromise = requestPublicConfigs()
    .then((data) => {
      siteConfigState.value = data;
      loadedState.value = true;
      return siteConfigState.value;
    })
    .catch((error: unknown) => {
      errorState.value = error instanceof Error ? error.message : "加载站点配置失败";
      throw error;
    })
    .finally(() => {
      loadingState.value = false;
      loadPromise = null;
    });

  return loadPromise;
}

export function useSiteConfigs() {
  if (!loadedState.value && !loadingState.value) {
    void loadSiteConfigs();
  }

  const homeBanner = computed<HomeBannerConfig>(() => ({
    ...defaultHomeBanner,
    ...(siteConfigState.value.home_banner ?? {})
  }));

  const homeFeatured = computed<HomeFeaturedConfig>(() => ({
    ...defaultHomeFeatured,
    ...(siteConfigState.value.home_featured ?? {})
  }));

  const contactChannels = computed<ContactChannelConfig[]>(() => {
    const channels = siteConfigState.value.contact_channels;
    return Array.isArray(channels) && channels.length > 0 ? channels : defaultContactChannels;
  });

  const footerConfig = computed<FooterConfig>(() => {
    const config = siteConfigState.value.footer_config ?? {};
    return {
      ...defaultFooterConfig,
      ...config,
      navLinks: Array.isArray(config.navLinks) && config.navLinks.length > 0 ? config.navLinks : defaultFooterConfig.navLinks
    };
  });

  const globalCta = computed<GlobalCtaConfig>(() => ({
    ...defaultGlobalCta,
    ...(siteConfigState.value.global_cta ?? {})
  }));

  const featuredContentSlots = computed<FeaturedContentSlotsConfig>(() => {
    const config = siteConfigState.value.featured_content_slots ?? {};
    return {
      ...defaultFeaturedContentSlots,
      ...config,
      eventsLimit: normalizePositiveNumber(config.eventsLimit, defaultFeaturedContentSlots.eventsLimit),
      buildersLimit: normalizePositiveNumber(config.buildersLimit, defaultFeaturedContentSlots.buildersLimit),
      insightsLimit: normalizePositiveNumber(config.insightsLimit, defaultFeaturedContentSlots.insightsLimit)
    };
  });

  return {
    homeBanner,
    homeFeatured,
    contactChannels,
    footerConfig,
    globalCta,
    featuredContentSlots,
    loading: computed(() => loadingState.value),
    loaded: computed(() => loadedState.value),
    error: computed(() => errorState.value),
    loadSiteConfigs
  };
}

function normalizePositiveNumber(value: unknown, fallback: number) {
  if (typeof value !== "number" || Number.isNaN(value) || value <= 0) {
    return fallback;
  }

  return Math.min(Math.floor(value), 12);
}
