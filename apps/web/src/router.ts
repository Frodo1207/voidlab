import { createRouter, createWebHistory } from "vue-router";
import HomePage from "../pages/index.vue";
import BuilderDetailPage from "../pages/builder-detail.vue";
import BuildersPage from "../pages/builders.vue";
import EventDetailPage from "../pages/event-detail.vue";
import EventsPage from "../pages/events.vue";
import InsightDetailPage from "../pages/insight-detail.vue";
import InsightsPage from "../pages/insights.vue";
import ContactPage from "../pages/contact.vue";
import KnowledgePage from "../pages/knowledge.vue";
import KnowledgeSpacePage from "../pages/knowledge-space.vue";
import KnowledgeEntryPage from "../pages/knowledge-entry.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      component: HomePage
    },
    {
      path: "/events",
      component: EventsPage
    },
    {
      path: "/insights",
      component: InsightsPage
    },
    {
      path: "/builders",
      component: BuildersPage
    },
    {
      path: "/builders/:slug",
      component: BuilderDetailPage
    },
    {
      path: "/events/:slug",
      component: EventDetailPage
    },
    {
      path: "/insights/:slug",
      component: InsightDetailPage
    },
    {
      path: "/contact",
      component: ContactPage
    },
    {
      path: "/knowledge",
      component: KnowledgePage
    },
    {
      path: "/knowledge/:spaceSlug",
      component: KnowledgeSpacePage
    },
    {
      path: "/knowledge/:spaceSlug/:entrySlug",
      component: KnowledgeEntryPage
    }
  ],
  scrollBehavior() {
    return { top: 0 };
  }
});

export default router;
