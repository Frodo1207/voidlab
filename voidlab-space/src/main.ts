import "./styles.css";
import { extractMarkdownHeadings, renderMarkdownWithOptions } from "./markdown";

type FrontmatterMap = Record<string, string>;

type EntryRecord = {
  spaceSlug: string;
  sectionSlug: string;
  sectionLabel: string;
  sourcePath: string;
  sourceDirectory: string;
  fileName: string;
  title: string;
  slug: string;
  summary: string;
  content: string;
  readMinutes: number;
  coverUrl: string;
};

type SpaceRecord = {
  slug: string;
  title: string;
  summary: string;
  introContent: string;
  introSourcePath: string;
  coverUrl: string;
  icon: string;
  themeTint: string;
  entries: EntryRecord[];
};

const SECTION_ORDER = [
  "nav",
  "intro",
  "foundation",
  "loop",
  "build",
  "practice",
  "career",
  "system"
];

const SECTION_LABELS: Record<string, string> = {
  nav: "导航",
  intro: "认知",
  foundation: "基础",
  loop: "闭环",
  build: "构建",
  practice: "实战",
  career: "表达",
  system: "系统"
};

const rawFiles = import.meta.glob("../content/entries/**/*.md", {
  query: "?raw",
  import: "default",
  eager: true
}) as Record<string, string>;

const assetFiles = import.meta.glob("../content/**/*.{png,jpg,jpeg,gif,webp,svg,avif}", {
  import: "default",
  eager: true
}) as Record<string, string>;

function humanizeSlug(value: string) {
  return value
    .replace(/^\d+[-_]?/, "")
    .replace(/[-_]+/g, " ")
    .trim();
}

function titleCaseWords(value: string) {
  const readable = humanizeSlug(value);
  if (!readable) {
    return value;
  }
  return readable
    .split(/\s+/)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

function titleFromBody(markdown: string) {
  const matched = markdown.match(/^#\s+(.+)\s*$/m);
  return matched?.[1]?.trim() ?? "";
}

function parseFrontmatter(raw: string) {
  if (!raw.startsWith("---\n")) {
    return { frontmatter: {} as FrontmatterMap, body: raw.trim() };
  }

  const lines = raw.split("\n");
  let endIndex = -1;
  for (let index = 1; index < lines.length; index += 1) {
    if (lines[index].trim() === "---") {
      endIndex = index;
      break;
    }
  }

  if (endIndex < 0) {
    return { frontmatter: {} as FrontmatterMap, body: raw.trim() };
  }

  const frontmatter: FrontmatterMap = {};
  for (const line of lines.slice(1, endIndex)) {
    const separator = line.indexOf(":");
    if (separator <= 0) {
      continue;
    }
    const key = line.slice(0, separator).trim().toLowerCase();
    const value = line.slice(separator + 1).trim().replace(/^['"]|['"]$/g, "");
    if (key && value) {
      frontmatter[key] = value;
    }
  }

  return {
    frontmatter,
    body: lines.slice(endIndex + 1).join("\n").trim()
  };
}

function extractSummary(markdown: string) {
  const lines = markdown
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)
    .filter((line) => !line.startsWith("#") && !line.startsWith("```"));
  return lines[0] ?? "";
}

function countReadUnits(markdown: string) {
  const latinWords = markdown.match(/[A-Za-z0-9_]+/g)?.length ?? 0;
  const hanCount = Array.from(markdown).filter((char) => /\p{Script=Han}/u.test(char)).length;
  return latinWords + hanCount;
}

function readMinutes(markdown: string) {
  const units = countReadUnits(markdown);
  return Math.max(1, Math.ceil(units / 220));
}

function normalizeTitle(value: string) {
  return value.trim() || "未命名文档";
}

function normalizeContentPath(value: string) {
  return value.replace(/\\/g, "/").replace(/^\.\.\//, "");
}

function dirname(path: string) {
  const normalized = normalizeContentPath(path);
  const index = normalized.lastIndexOf("/");
  return index >= 0 ? normalized.slice(0, index) : normalized;
}

function resolveRelativePath(baseDir: string, relativePath: string) {
  const baseParts = baseDir.split("/").filter(Boolean);
  const relativeParts = relativePath.split("/").filter(Boolean);
  const merged = [...baseParts];
  for (const part of relativeParts) {
    if (part === ".") {
      continue;
    }
    if (part === "..") {
      merged.pop();
      continue;
    }
    merged.push(part);
  }
  return merged.join("/");
}

function buildAssetUrlMap() {
  const map = new Map<string, string>();
  for (const [sourcePath, url] of Object.entries(assetFiles)) {
    const normalized = normalizeContentPath(sourcePath);
    map.set(normalized, url);
    map.set(normalized.replace(/^content\//, ""), url);
  }
  return map;
}

const assetUrlMap = buildAssetUrlMap();

function resolveAssetUrl(sourcePath: string, relativeTo = "") {
  const trimmed = sourcePath.trim();
  if (!trimmed) {
    return "";
  }

  if (
    trimmed.startsWith("http://") ||
    trimmed.startsWith("https://") ||
    trimmed.startsWith("data:")
  ) {
    return trimmed;
  }

  const normalized = trimmed.replace(/\\/g, "/");
  const candidates = new Set<string>();
  if (normalized.startsWith("/")) {
    candidates.add(normalized.slice(1));
  } else if (normalized.startsWith("content/")) {
    candidates.add(normalized);
  } else if (relativeTo) {
    candidates.add(resolveRelativePath(relativeTo, normalized));
  }

  candidates.add(normalized.replace(/^\.\//, ""));
  if (!normalized.startsWith("content/")) {
    candidates.add(`content/${normalized.replace(/^\.\//, "")}`);
    candidates.add(`content/assets/${normalized.replace(/^\.\//, "")}`);
  }

  for (const candidate of candidates) {
    const direct = assetUrlMap.get(candidate);
    if (direct) {
      return direct;
    }
  }

  return trimmed;
}

function compareSection(left: string, right: string) {
  const leftIndex = SECTION_ORDER.indexOf(left);
  const rightIndex = SECTION_ORDER.indexOf(right);
  if (leftIndex === -1 && rightIndex === -1) {
    return left.localeCompare(right);
  }
  if (leftIndex === -1) {
    return 1;
  }
  if (rightIndex === -1) {
    return -1;
  }
  return leftIndex - rightIndex;
}

function buildContentModel() {
  const entries: EntryRecord[] = [];
  const spaceMeta = new Map<
    string,
    {
      title: string;
      summary: string;
      introContent: string;
      introSourcePath: string;
      coverUrl: string;
      icon: string;
      themeTint: string;
    }
  >();

  for (const [sourcePath, raw] of Object.entries(rawFiles)) {
    const normalized = normalizeContentPath(sourcePath);
    const segments = normalized.split("/");
    const entriesIndex = segments.lastIndexOf("entries");
    const relativeSegments = segments.slice(entriesIndex + 1);
    const [spaceSlug = "default"] = relativeSegments;
    const { frontmatter, body } = parseFrontmatter(raw);
    const sourcePathForView = normalized.replace(/^content\//, "content/");

    if (relativeSegments.length === 2 && relativeSegments[1] === "_index.md") {
      spaceMeta.set(spaceSlug, {
        title: normalizeTitle(frontmatter.title || titleFromBody(body) || titleCaseWords(spaceSlug)),
        summary: frontmatter.public_summary || frontmatter.summary || extractSummary(body),
        introContent: body,
        introSourcePath: sourcePathForView,
        coverUrl: resolveAssetUrl(frontmatter.cover_url || frontmatter.cover || "", dirname(sourcePathForView)),
        icon: frontmatter.icon || "📘",
        themeTint: frontmatter.theme_tint || "soft-paper"
      });
      continue;
    }

    const [, sectionSlug = "misc", fileName = "untitled.md"] = relativeSegments;
    const sourceDirectory = dirname(sourcePathForView);
    const title = normalizeTitle(frontmatter.title || titleFromBody(body) || humanizeSlug(fileName.replace(/\.md$/i, "")));
    const slug = (frontmatter.slug || fileName.replace(/\.md$/i, "").replace(/^\d+[-_]?/, "")).trim();
    entries.push({
      spaceSlug,
      sectionSlug,
      sectionLabel: frontmatter.section_name || SECTION_LABELS[sectionSlug] || humanizeSlug(sectionSlug),
      sourcePath: sourcePathForView,
      sourceDirectory,
      fileName,
      title,
      slug,
      summary: frontmatter.public_summary || frontmatter.summary || extractSummary(body),
      content: body,
      readMinutes: Number(frontmatter.estimated_read_minutes || frontmatter.read_minutes) || readMinutes(body),
      coverUrl: resolveAssetUrl(frontmatter.cover_url || frontmatter.cover || "", sourceDirectory)
    });
  }

  entries.sort((left, right) => {
    if (left.spaceSlug !== right.spaceSlug) {
      return left.spaceSlug.localeCompare(right.spaceSlug);
    }
    const sectionCompare = compareSection(left.sectionSlug, right.sectionSlug);
    if (sectionCompare !== 0) {
      return sectionCompare;
    }
    return left.fileName.localeCompare(right.fileName);
  });

  const spacesMap = new Map<string, SpaceRecord>();
  for (const entry of entries) {
    const meta = spaceMeta.get(entry.spaceSlug);
    const current = spacesMap.get(entry.spaceSlug) ?? {
      slug: entry.spaceSlug,
      title: meta?.title || titleCaseWords(entry.spaceSlug),
      summary: meta?.summary || "",
      introContent: meta?.introContent || "",
      introSourcePath: meta?.introSourcePath || "",
      coverUrl: meta?.coverUrl || "",
      icon: meta?.icon || "📘",
      themeTint: meta?.themeTint || "soft-paper",
      entries: []
    };
    current.entries.push(entry);
    spacesMap.set(entry.spaceSlug, current);
  }

  for (const [spaceSlug, meta] of spaceMeta.entries()) {
    if (!spacesMap.has(spaceSlug)) {
      spacesMap.set(spaceSlug, {
        slug: spaceSlug,
        title: meta.title || titleCaseWords(spaceSlug),
        summary: meta.summary,
        introContent: meta.introContent,
        introSourcePath: meta.introSourcePath,
        coverUrl: meta.coverUrl,
        icon: meta.icon,
        themeTint: meta.themeTint,
        entries: []
      });
    }
  }

  return Array.from(spacesMap.values()).sort((left, right) => left.slug.localeCompare(right.slug));
}

function getSelection(spaces: SpaceRecord[]) {
  const rawHash = window.location.hash.replace(/^#\/?/, "");
  const [hashSpace, hashEntry] = rawHash.split("/").filter(Boolean);
  const fallbackSpace = spaces[0];
  const currentSpace =
    spaces.find((space) => space.slug === hashSpace) ??
    fallbackSpace ??
    null;

  const currentEntry = hashEntry
    ? currentSpace?.entries.find((entry) => entry.slug === hashEntry) ?? null
    : null;

  const mode: "space" | "entry" = currentEntry ? "entry" : "space";
  if (currentSpace) {
    const expectedHash = currentEntry ? `#/${currentSpace.slug}/${currentEntry.slug}` : `#/${currentSpace.slug}`;
    if (window.location.hash !== expectedHash) {
      window.history.replaceState({}, "", expectedHash);
    }
  }

  return { currentSpace, currentEntry, mode };
}

function escapeHtml(value: unknown) {
  return String(value ?? "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function renderTocLinks(markdownText: string) {
  const headings = extractMarkdownHeadings(markdownText);
  if (!headings.length) {
    return `<div class="toc-empty">当前文档还没有可提取的标题目录。</div>`;
  }

  return headings
    .map(
      (heading) => `
        <button class="toc-link level-${heading.level}" data-scroll-target="${escapeHtml(heading.id)}" type="button">
          ${escapeHtml(heading.title)}
        </button>
      `
    )
    .join("");
}

function renderEntryPreview(entry: EntryRecord, markdownText: string) {
  return renderMarkdownWithOptions(markdownText, {
    imageResolver: (src) => resolveEntryMarkdownImage(entry, src)
  });
}

function sectionEntryCountLabel(space: SpaceRecord) {
  const sectionCount = new Set(space.entries.map((entry) => entry.sectionSlug)).size;
  return `${space.entries.length} 篇 / ${sectionCount} 个分组`;
}

function groupEntriesBySection(space: SpaceRecord) {
  const grouped = new Map<string, EntryRecord[]>();
  for (const entry of space.entries) {
    const list = grouped.get(entry.sectionSlug) ?? [];
    list.push(entry);
    grouped.set(entry.sectionSlug, list);
  }
  return Array.from(grouped.entries()).sort(([left], [right]) => compareSection(left, right));
}

function resolveEntryMarkdownImage(entry: EntryRecord, src: string) {
  return resolveAssetUrl(src, entry.sourceDirectory);
}

function coverStyle(space: SpaceRecord) {
  if (space.coverUrl) {
    return `style="background-image: linear-gradient(rgba(255,255,255,0.12), rgba(255,255,255,0.04)), url('${escapeHtml(space.coverUrl)}'); background-size: cover; background-position: center;"`;
  }
  return "";
}

function renderSpaceHome(space: SpaceRecord) {
  const groupedEntries = groupEntriesBySection(space);
  const renderedIntro = renderMarkdownWithOptions(space.introContent || "", {
    imageResolver: (src) => resolveAssetUrl(src, dirname(space.introSourcePath || "content/entries"))
  });

  return `
    <main class="preview-panel preview-space-home">
      <section class="space-cover ${space.themeTint}" ${coverStyle(space)}>
        <div class="space-cover-inner">
          <div class="space-icon">${escapeHtml(space.icon)}</div>
          <div class="space-cover-copy">
            <div class="eyebrow">SPACE HOME</div>
            <h1>${escapeHtml(space.title)}</h1>
            <div class="preview-stats">
              <span>${space.entries.length} 篇内容</span>
              <span>·</span>
              <span>${new Set(space.entries.map((entry) => entry.sectionSlug)).size} 个分组</span>
            </div>
            ${space.summary ? `<p class="preview-summary">${escapeHtml(space.summary)}</p>` : ""}
          </div>
        </div>
      </section>

      <section class="space-home-grid">
        <article class="space-intro markdown-entry markdown-space-intro">
          ${renderedIntro || `<p>这个 space 还没有写 <code>_index.md</code>，你可以先补一篇空间首页说明。</p>`}
        </article>

        <section class="space-directory-card">
          <div class="eyebrow">空间目录</div>
          ${groupedEntries
            .map(
              ([sectionSlug, sectionEntries]) => `
                <div class="space-directory-group">
                  <div class="space-directory-title">${escapeHtml(sectionEntries[0]?.sectionLabel || sectionSlug)}</div>
                  ${sectionEntries
                    .map(
                      (entry) => `
                        <a class="space-directory-item" href="#/${entry.spaceSlug}/${entry.slug}">
                          <div class="space-directory-main">
                            <span class="space-directory-doc">📄</span>
                            <span>${escapeHtml(entry.title)}</span>
                          </div>
                          <div class="space-directory-meta">
                            <span>${entry.readMinutes} 分钟</span>
                            <span>›</span>
                          </div>
                        </a>
                      `
                    )
                    .join("")}
                </div>
              `
            )
            .join("")}
        </section>
      </section>
    </main>
  `;
}

function renderEntryView(space: SpaceRecord, entry: EntryRecord) {
  const rendered = renderEntryPreview(entry, entry.content);

  return `
    <main class="preview-panel entry-mode">
      <div class="entry-content">
        ${entry.coverUrl ? `<div class="entry-cover" style="background-image: linear-gradient(rgba(255,255,255,0.1), rgba(255,255,255,0.02)), url('${escapeHtml(entry.coverUrl)}');"></div>` : ""}

        <div class="preview-head preview-head-grid">
          <div class="preview-meta">
            <div class="eyebrow">SPACE PREVIEW</div>
            <div class="breadcrumb-row">
              <a href="#/${space.slug}">${escapeHtml(space.title)}</a>
              <span>›</span>
              <span>${escapeHtml(entry.sectionLabel)}</span>
            </div>
            <h1>${escapeHtml(entry.title)}</h1>
            <div class="preview-stats">
              <span>${escapeHtml(space.title)}</span>
              <span>·</span>
              <span>${escapeHtml(entry.sectionLabel)}</span>
              <span>·</span>
              <span>${entry.readMinutes} 分钟</span>
            </div>
            ${entry.summary ? `<p class="preview-summary">${escapeHtml(entry.summary)}</p>` : ""}
          </div>
          <div class="source-meta-card">
            <div class="source-meta-label">源文件路径</div>
            <code>${escapeHtml(entry.sourcePath)}</code>
          </div>
        </div>

        <article class="markdown-entry">${rendered}</article>
      </div>

      <aside class="toc-panel">
        <div class="toc-sticky">${renderTocLinks(entry.content)}</div>
      </aside>
    </main>
  `;
}

function renderEmptyState() {
  return `
    <div class="empty-shell">
      <div class="empty-card">
        <div class="eyebrow">VOIDLAB SPACE</div>
        <h1>这里已经是你的内容源空间了</h1>
        <p>把 Markdown 放进 <code>content/entries/&lt;space-slug&gt;/&lt;section&gt;/</code>，页面会自动刷新并显示最终预览效果。</p>
        <div class="empty-path">当前约定目录：<code>voidlab-space/content/entries</code></div>
      </div>
    </div>
  `;
}

function renderApp() {
  const app = document.querySelector<HTMLDivElement>("#app");
  if (!app) {
    return;
  }

  const spaces = buildContentModel();
  if (spaces.length === 0) {
    app.innerHTML = renderEmptyState();
    return;
  }

  const { currentSpace, currentEntry, mode } = getSelection(spaces);
  if (!currentSpace) {
    app.innerHTML = renderEmptyState();
    return;
  }

  const groupedEntries = groupEntriesBySection(currentSpace);

  app.innerHTML = `
    <div class="shell">
      <aside class="sidebar">
        <div class="sidebar-header">
          <div class="sidebar-title">目录</div>
          <div class="sidebar-current-space">${escapeHtml(currentSpace.title)}</div>
        </div>

        <div class="sidebar-section-label">Space</div>
        <div class="space-switcher">
          ${spaces
            .map(
              (space) => `
                <a class="space-chip ${space.slug === currentSpace.slug ? "is-active" : ""}" href="#/${space.slug}">
                  <span>${escapeHtml(space.title)}</span>
                  <small>${escapeHtml(sectionEntryCountLabel(space))}</small>
                </a>
              `
            )
            .join("")}
        </div>

        <div class="sidebar-section-label">文章</div>
        <div class="entry-groups">
            ${groupedEntries
            .map(
              ([sectionSlug, sectionEntries]) => `
                <section class="entry-group">
                  <div class="group-title">${escapeHtml(sectionEntries[0]?.sectionLabel || sectionSlug)}</div>
                  ${sectionEntries
                    .map(
                      (entry) => `
                        <a class="entry-link ${entry.slug === currentEntry?.slug ? "is-active" : ""}" href="#/${entry.spaceSlug}/${entry.slug}">
                          <span>${escapeHtml(entry.title)}</span>
                          <small>${entry.readMinutes} 分钟</small>
                        </a>
                      `
                    )
                    .join("")}
                </section>
              `
            )
            .join("")}
        </div>

        <div class="sidebar-foot">
          <div>源文件</div>
          <code>${escapeHtml(mode === "space" ? currentSpace.introSourcePath || "content/entries/<space>/_index.md" : currentEntry?.sourcePath || "")}</code>
        </div>
      </aside>

      ${
        mode === "space" || !currentEntry
          ? renderSpaceHome(currentSpace)
          : renderEntryView(currentSpace, currentEntry)
      }
    </div>
  `;

  const bindTocLinks = () => {
    app.querySelectorAll<HTMLElement>("[data-scroll-target]").forEach((element) => {
      element.addEventListener("click", () => {
        const targetId = element.dataset.scrollTarget;
        if (!targetId) {
          return;
        }
        const target = document.getElementById(targetId);
        target?.scrollIntoView({ behavior: "smooth", block: "start" });
      });
    });
  };

  bindTocLinks();
}

window.addEventListener("hashchange", renderApp);
renderApp();
