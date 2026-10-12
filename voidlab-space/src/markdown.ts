import MarkdownIt from "markdown-it";
import hljs from "highlight.js";

hljs.registerAliases(["vue"], { languageName: "xml" });

const markdown = new MarkdownIt({
  html: false,
  linkify: true,
  breaks: true,
  highlight(code, lang) {
    const normalizedLang = lang === "vue" ? "xml" : lang;
    if (normalizedLang && hljs.getLanguage(normalizedLang)) {
      try {
        return hljs.highlight(code, { language: normalizedLang }).value;
      } catch {
        return markdown.utils.escapeHtml(code);
      }
    }
    return markdown.utils.escapeHtml(code);
  }
});

const originalFenceRule =
  markdown.renderer.rules.fence ??
  ((tokens, idx, options, env, self) => self.renderToken(tokens, idx, options));

markdown.renderer.rules.fence = (tokens, idx, options, env, self) => {
  const token = tokens[idx];
  const lang = (token.info || "").trim().split(/\s+/)[0];
  const rendered = originalFenceRule(tokens, idx, options, env, self);
  return `<div class="md-codeblock" data-lang="${markdown.utils.escapeHtml(lang)}">${rendered}</div>`;
};

function slugifyHeading(value: string) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[`~!@#$%^&*()+=,./?<>:;"'\\|[\]{}]+/g, "")
    .replace(/\s+/g, "-");
}

markdown.renderer.rules.heading_open = (tokens, idx, options, _env, self) => {
  const inlineToken = tokens[idx + 1];
  const headingText = inlineToken?.type === "inline" ? inlineToken.content : "";
  const headingId = slugifyHeading(headingText);
  if (headingId) {
    tokens[idx].attrSet("id", headingId);
  }
  return self.renderToken(tokens, idx, options);
};

export function renderMarkdown(markdownText: string) {
  return markdown.render(markdownText);
}

export type MarkdownRenderOptions = {
  imageResolver?: (src: string) => string;
};

export function renderMarkdownWithOptions(markdownText: string, options?: MarkdownRenderOptions) {
  if (!options?.imageResolver) {
    return markdown.render(markdownText);
  }

  const imageResolver = options.imageResolver;
  const originalImageRule =
    markdown.renderer.rules.image ??
    ((tokens, idx, renderOptions, _env, self) => self.renderToken(tokens, idx, renderOptions));

  markdown.renderer.rules.image = (tokens, idx, renderOptions, env, self) => {
    const token = tokens[idx];
    const src = token.attrGet("src");
    if (src) {
      token.attrSet("src", imageResolver(src));
    }
    return originalImageRule(tokens, idx, renderOptions, env, self);
  };

  const rendered = markdown.render(markdownText);
  markdown.renderer.rules.image = originalImageRule;
  return rendered;
}

export function extractMarkdownHeadings(markdownText: string) {
  return markdownText
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => /^#{1,3}\s+/.test(line))
    .map((line) => {
      const [, hashes, rawTitle] = line.match(/^(#{1,3})\s+(.+)$/) ?? [];
      const title = rawTitle?.trim() ?? "";
      return {
        level: hashes?.length ?? 1,
        title,
        id: slugifyHeading(title)
      };
    })
    .filter((item) => item.title);
}
