import { apiRequest } from "./apiClient";
import type {
  ContactChannelConfig,
  FeaturedContentSlotsConfig,
  FooterConfig,
  GlobalCtaConfig,
  HomeBannerConfig,
  HomeFeaturedConfig,
  SiteConfigKey,
  SiteConfigRecord
} from "../types";

type SiteConfigResponse<T> = {
  id: number;
  config_key: SiteConfigKey;
  config_value: T;
  updated_by: number;
  updated_at: string;
};

function mapSiteConfigRecord<T>(record: SiteConfigResponse<T>): SiteConfigRecord<T> {
  return {
    id: record.id,
    configKey: record.config_key,
    configValue: record.config_value,
    updatedBy: record.updated_by,
    updatedAt: record.updated_at
  };
}

export async function listSiteConfigs() {
  const records = await apiRequest<SiteConfigResponse<unknown>[]>("/api/v1/site-configs");
  return records.map((record) => mapSiteConfigRecord(record));
}

export async function updateSiteConfig<T>(key: SiteConfigKey, value: T) {
  const record = await apiRequest<SiteConfigResponse<T>>(`/api/v1/site-configs/${key}`, {
    method: "PUT",
    body: JSON.stringify({
      config_value: value
    })
  });

  return mapSiteConfigRecord(record);
}

export const defaultHomeBannerConfig: HomeBannerConfig = {
  titleText: "探索 AI 的边界",
  subtitle: "我们构建 AI 资产，不做玩具。",
  primaryCtaLabel: "查看社区活动",
  primaryCtaPath: "/events",
  secondaryCtaLabel: "联系我们",
  secondaryCtaPath: "/contact",
  statusLabel: "系统状态 // 社区指标"
};

export const defaultHomeFeaturedConfig: HomeFeaturedConfig = {
  communityCount: "1000",
  communityCountSuffix: "+",
  eventCount: "10",
  eventCountSuffix: "+",
  eventsDescription: "",
  buildersDescription: "社区里不只有活动和内容，还有一批可以被连接、被协作，也能直接发起合作的人。",
  insightsDescription: "剥离噪音，只提供值得关注的行业信号与落地经验。"
};

export const defaultContactChannelsConfig: ContactChannelConfig[] = [
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

export const defaultFooterConfig: FooterConfig = {
  slogan: "连接共创者，重塑 AI 资产。",
  navLinks: [
    { label: "首页", path: "/" },
    { label: "活动", path: "/events" },
    { label: "资讯", path: "/insights" },
    { label: "联系我们", path: "/contact" }
  ],
  legalText: "VOIDLAB.AI © 2026. All rights reserved."
};

export const defaultGlobalCtaConfig: GlobalCtaConfig = {
  eyebrow: "NEXT ACTION",
  title: "准备把想法变成活动、内容或合作项目？",
  description: "如果你想和 VOIDLAB 一起办活动、发起合作或咨询 AI 落地方案，可以直接从这里进入下一步。",
  primaryLabel: "提交合作意向",
  primaryPath: "/contact",
  secondaryLabel: "查看活动",
  secondaryPath: "/events"
};

export const defaultFeaturedContentSlotsConfig: FeaturedContentSlotsConfig = {
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
