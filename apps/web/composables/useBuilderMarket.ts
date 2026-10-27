import { resolvePublicApiPath } from "../src/runtimeConfig";

const TOKEN_KEY = "voidlab.network.token";

export type NetworkProfile = {
  id: number;
  slug: string;
  is_demo?: boolean;
  display_name: string;
  primary_role: string;
  headline: string;
  bio: string;
  city: string;
  skills: string[];
  collaboration_modes: string[];
  availability: string;
  available_from: string;
};

export const devMockProfiles: NetworkProfile[] = [
  {
    id: 99001,
    slug: "lin-mu-agent-product-engineer",
    is_demo: true,
    display_name: "林沐",
    primary_role: "AI 产品工程师",
    headline: "把想法变成能持续迭代的 AI 产品，关注 Agent 工作流和真实用户体验。",
    bio: "参与过从原型验证到小规模上线的 AI 产品项目，擅长在工程约束和用户需求之间找到可交付的第一步。",
    city: "上海",
    skills: ["Python", "LLM 应用", "Agent 工作流", "产品原型"],
    collaboration_modes: ["远程", "项目制"],
    availability: "open",
    available_from: "近期可开始"
  },
  {
    id: 99002,
    slug: "zhou-kewei-platform-engineer",
    is_demo: true,
    display_name: "周可为",
    primary_role: "平台工程师",
    headline: "专注服务稳定性、开发者体验与 AI 基础设施，让团队更快交付。",
    bio: "熟悉云原生平台、部署流水线和可观测性建设，擅长从重复运维工作中提炼平台能力。",
    city: "深圳",
    skills: ["Go", "Kubernetes", "Docker", "可观测性"],
    collaboration_modes: ["远程", "按阶段合作"],
    availability: "limited",
    available_from: "下月起"
  },
  {
    id: 99003,
    slug: "xu-xinghe-product-designer",
    is_demo: true,
    display_name: "许星河",
    primary_role: "产品设计师",
    headline: "将复杂工具做得清晰易用，关注 AI 产品的信息架构和交互细节。",
    bio: "长期参与 B 端工具与数据产品设计，习惯和产品、研发一起定义问题并快速验证方案。",
    city: "杭州",
    skills: ["产品设计", "Figma", "信息架构", "原型验证"],
    collaboration_modes: ["远程", "顾问支持"],
    availability: "open",
    available_from: "近期可开始"
  },
  {
    id: 99004,
    slug: "chen-yinuo-frontend-engineer",
    is_demo: true,
    display_name: "陈一诺",
    primary_role: "前端工程师",
    headline: "用稳定的组件体系和细致的交互，把复杂业务界面做得顺手可靠。",
    bio: "熟悉 Vue 与 React 项目交付，关注页面性能、可访问性和前后端协作效率。",
    city: "成都",
    skills: ["Vue 3", "TypeScript", "设计系统", "前端性能"],
    collaboration_modes: ["远程", "项目制"],
    availability: "open",
    available_from: "本周可开始"
  },
  {
    id: 99005,
    slug: "tang-yue-growth-operator",
    is_demo: true,
    display_name: "唐越",
    primary_role: "增长运营",
    headline: "擅长把内容、渠道和数据串成可以复盘的增长实验。",
    bio: "做过早期产品冷启动与内容运营，希望加入重视长期口碑和真实用户价值的团队。",
    city: "北京",
    skills: ["内容增长", "用户研究", "数据分析", "渠道运营"],
    collaboration_modes: ["远程", "顾问支持"],
    availability: "limited",
    available_from: "两周后"
  },
  {
    id: 99006,
    slug: "gao-yuan-data-engineer",
    is_demo: true,
    display_name: "高远",
    primary_role: "数据工程师",
    headline: "搭建可信的数据管道与指标体系，让团队更快看见产品真实表现。",
    bio: "熟悉数据采集、清洗、分析和可视化，关注数据质量与业务问题之间的连接。",
    city: "苏州",
    skills: ["SQL", "Python", "数据管道", "指标体系"],
    collaboration_modes: ["远程", "短期项目"],
    availability: "unavailable",
    available_from: "暂不开放"
  },
  {
    id: 99007,
    slug: "fang-zhiye-technical-cofounder",
    is_demo: true,
    display_name: "方知野",
    primary_role: "技术联合创始人",
    headline: "从 0 到 1 构建产品与团队，寻找值得长期投入的问题。",
    bio: "有全栈交付和早期团队协作经验，愿意共同打磨产品方向、技术方案与最小商业闭环。",
    city: "远程",
    skills: ["全栈开发", "技术架构", "产品验证", "团队协作"],
    collaboration_modes: ["远程", "联合创业"],
    availability: "open",
    available_from: "可协商"
  },
  {
    id: 99008,
    slug: "liang-xiaochuan-llm-engineer",
    is_demo: true,
    display_name: "梁晓川",
    primary_role: "LLM 工程师",
    headline: "关注检索增强、模型评测和成本控制，让大模型能力进入可靠的产品流程。",
    bio: "做过知识库问答与模型应用集成，喜欢用小规模实验验证方案，再逐步完善工程质量。",
    city: "南京",
    skills: ["RAG", "模型评测", "向量检索", "Python"],
    collaboration_modes: ["远程", "按模块合作"],
    availability: "open",
    available_from: "近期可开始"
  }
];

export type NetworkWork = {
  id: number;
  profile_id: number;
  slug: string;
  title: string;
  summary: string;
  body: string;
  cover_url: string;
  demo_url: string;
  repo_url: string;
  tags: string[];
  status: string;
  author_name: string;
  members?: Array<{ profile_id: number; display_name?: string; role: string; contribution: string; sort_order: number; is_owner: boolean }>;
};

export type NetworkOpportunity = {
  id: number;
  project_id?: number;
  is_demo?: boolean;
  profile_id: number;
  slug: string;
  title: string;
  summary: string;
  opportunity_type: string;
  looking_for_roles: string[];
  required_skills: string[];
  engagement_mode: string;
  budget_hint: string;
  duration_hint: string;
  city_scope: string;
  remote_allowed: boolean;
  status: string;
  expires_at: string;
  created_at?: string;
  updated_at?: string;
  owner_name: string;
  project_plan?: NetworkProjectPlan | null;
  team_intro?: string;
  team_members?: NetworkOpportunityMember[];
};

export type NetworkProject = {
  id: number;
  is_demo?: boolean;
  owner_profile_id: number;
  slug: string;
  title: string;
  summary: string;
  status: string;
  project_plan?: NetworkProjectPlan | null;
  team_intro: string;
  owner_name: string;
  members: NetworkOpportunityMember[];
  opportunities: NetworkOpportunity[];
  created_at: string;
  updated_at: string;
};

export function projectFromLegacyOpportunity(opportunity: NetworkOpportunity): NetworkProject {
  const projectID = opportunity.project_id || opportunity.id + 100000;
  return {
    id: projectID,
    is_demo: opportunity.is_demo,
    owner_profile_id: opportunity.profile_id,
    slug: opportunity.slug,
    title: opportunity.title,
    summary: opportunity.summary,
    status: "published",
    project_plan: opportunity.project_plan,
    team_intro: opportunity.team_intro || "",
    owner_name: opportunity.owner_name,
    members: opportunity.team_members || [],
    opportunities: [{ ...opportunity, project_id: projectID }],
    created_at: opportunity.created_at || "",
    updated_at: opportunity.updated_at || opportunity.created_at || ""
  };
}

export type NetworkProjectPlan = {
  milestones: NetworkProjectMilestone[];
  sections: NetworkProjectPlanSection[];
};

export type NetworkProjectMilestone = {
  title: string;
  description: string;
  order: number;
  status: "planned" | "in_progress" | "completed" | "blocked";
  timing?: NetworkProjectTiming | null;
  goals: string[];
  deliverables: string[];
  acceptance_criteria: string[];
  owner_member_ids: number[];
};

export type NetworkProjectTiming = {
  kind: "fixed_dates" | "duration" | "relative" | "flexible" | "unspecified";
  label?: string;
  start_at?: string;
  end_at?: string;
  duration_value?: number;
  duration_unit?: string;
};

export type NetworkProjectPlanSection = {
  type: string;
  title: string;
  body?: string;
  items: string[];
  sort_order: number;
};

export type NetworkOpportunityMember = {
  profile_id: number;
  display_name?: string;
  slug?: string;
  primary_role?: string;
  headline?: string;
  role: string;
  contribution: string;
  sort_order: number;
};

export type NetworkContactRequest = {
  id: number;
  is_demo?: boolean;
  from_profile_id: number;
  target_type: string;
  target_id: number;
  target_label?: string;
  message: string;
  status: string;
  progress_status: string;
  exchange_revoked: boolean;
  applicant_name: string;
  applicant_role: string;
  share_email: boolean;
  share_wechat: boolean;
  share_phone: boolean;
  related_work_ids: number[];
};

export type NetworkSession = {
  token: string;
  account: { id: number; email: string };
  profile: NetworkProfile;
};

export type NetworkAgentToken = {
  id: number;
  profile_id: number;
  name: string;
  scopes: string[];
  is_active: boolean;
  last_used_at: string;
  created_at: string;
  updated_at: string;
};

export type NetworkAgentTokenCreateResult = {
  token: string;
  record: NetworkAgentToken;
};

type Envelope<T> = {
  code: number;
  message: string;
  data: T;
};

export class NetworkRequestError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly code: number,
    readonly path: string
  ) {
    super(message);
    this.name = "NetworkRequestError";
  }
}

export function getNetworkToken() {
  return localStorage.getItem(TOKEN_KEY) || "";
}

export function clearNetworkToken() {
  localStorage.removeItem(TOKEN_KEY);
}

export function saveNetworkSession(session: NetworkSession) {
  localStorage.setItem(TOKEN_KEY, session.token);
}

export async function networkRequest<T>(path: string, init: RequestInit = {}, authenticated = false): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (authenticated) {
    const token = getNetworkToken();
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
  }

  let response: Response;
  try {
    response = await fetch(resolvePublicApiPath(path), { ...init, headers });
  } catch {
    throw new Error("无法连接市场服务，请确认 API 已启动。");
  }
  let envelope: Envelope<T>;
  try {
    envelope = (await response.json()) as Envelope<T>;
  } catch {
    throw new NetworkRequestError("市场服务返回了无效响应，请稍后重试。", response.status, -1, path);
  }
  if (!response.ok || envelope.code !== 0) {
    throw new NetworkRequestError(envelope.message || "请求失败", response.status, envelope.code, path);
  }
  return envelope.data;
}

export function postJSON<T>(path: string, data: unknown, authenticated = false) {
  return networkRequest<T>(path, { method: "POST", body: JSON.stringify(data) }, authenticated);
}
