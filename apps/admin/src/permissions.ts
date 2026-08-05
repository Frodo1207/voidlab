import type { UserRole } from "./types";

export interface MenuItem {
  index: string;
  label: string;
  roles?: UserRole[];
}

export const menuItems: MenuItem[] = [
  { index: "/", label: "仪表盘", roles: ["admin", "editor", "ops"] },
  { index: "/users", label: "用户管理", roles: ["admin"] },
  { index: "/agent-tokens", label: "Agent Tokens", roles: ["admin"] },
  { index: "/articles", label: "资讯管理", roles: ["admin", "editor"] },
  { index: "/events", label: "活动管理", roles: ["admin", "editor"] },
  { index: "/builders", label: "Builder 管理", roles: ["admin", "editor"] },
  { index: "/knowledge/spaces", label: "知识库", roles: ["admin", "editor"] },
  { index: "/leads", label: "Leads 管理", roles: ["admin", "ops"] },
  { index: "/site-configs", label: "站点配置", roles: ["admin"] },
  { index: "/audit-logs", label: "审计日志", roles: ["admin"] },
  { index: "/media", label: "媒体资源", roles: ["admin", "editor"] }
];

export function hasRoleAccess(
  role: UserRole | null | undefined,
  allowedRoles?: readonly UserRole[]
): boolean {
  if (!allowedRoles?.length) {
    return true;
  }

  if (!role) {
    return false;
  }

  return allowedRoles.includes(role);
}

export function getRoleLabel(role: string | null | undefined): string {
  switch (role) {
    case "admin":
      return "管理员";
    case "editor":
      return "内容编辑";
    case "ops":
      return "运营";
    case "agent":
      return "Agent";
    default:
      return "未分配角色";
  }
}
