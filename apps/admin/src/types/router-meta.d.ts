import "vue-router";

import type { UserRole } from "../types";

declare module "vue-router" {
  interface RouteMeta {
    public?: boolean;
    title?: string;
    roles?: UserRole[];
  }
}
