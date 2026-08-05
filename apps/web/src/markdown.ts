import MarkdownIt from "markdown-it";
import hljs from "highlight.js";

// Register Vue alias to XML so that it highlights correctly
hljs.registerAliases(["vue"], { languageName: "xml" });

const markdown = new MarkdownIt({
  html: false,
  linkify: true,
  breaks: true,
  highlight: function (str, lang) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(str, { language: lang }).value;
      } catch (__) {}
    }
    return ''; // use external default escaping
  }
});

const originalFenceRule =
  markdown.renderer.rules.fence ??
  ((tokens, idx, renderOptions, env, self) => self.renderToken(tokens, idx, renderOptions));

const originalCodeBlockRule =
  markdown.renderer.rules.code_block ??
  ((tokens, idx, renderOptions, env, self) => self.renderToken(tokens, idx, renderOptions));

markdown.renderer.rules.fence = (tokens, idx, renderOptions, env, self) => {
  const token = tokens[idx];
  const lang = (token.info || "").trim().split(/\s+/)[0];
  
  let hljsLang = lang;
  if (lang === "vue") {
    hljsLang = "xml";
  }

  let rendered = "";
  if (hljsLang && hljs.getLanguage(hljsLang)) {
    try {
      rendered = hljs.highlight(token.content, { language: hljsLang }).value;
    } catch (__) {
      rendered = markdown.utils.escapeHtml(token.content);
    }
  } else {
    rendered = markdown.utils.escapeHtml(token.content);
  }
  
  const escapedLang = markdown.utils.escapeHtml(lang);
  const escapedCode = markdown.utils.escapeHtml(token.content);
  
  return `<div class="md-codeblock"${escapedLang ? ` data-lang="${escapedLang}"` : ""}>
    <button class="md-codeblock-copy" data-code="${escapedCode}" onclick="navigator.clipboard.writeText(this.getAttribute('data-code')).then(() => { const t = this.innerText; this.innerText = 'Copied!'; setTimeout(() => this.innerText = t, 2000); })">Copy</button>
    <pre><code class="${escapedLang ? `language-${escapedLang}` : ""}">${rendered}</code></pre>
  </div>`;
};

markdown.renderer.rules.code_block = (tokens, idx, renderOptions, env, self) => {
  const rendered = originalCodeBlockRule(tokens, idx, renderOptions, env, self);
  return `<div class="md-codeblock">${rendered}</div>`;
};

export type MarkdownRenderOptions = {
  imageResolver?: (src: string) => string;
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
    const token = tokens[idx];
    token.attrSet("id", headingId);
  }

  return self.renderToken(tokens, idx, options);
};

function extractPlainText(markdownText: string) {
  return markdownText
    .replace(/```[\s\S]*?```/g, " ")
    .replace(/`([^`]+)`/g, "$1")
    .replace(/!\[([^\]]*)\]\(([^)]+)\)/g, "$1")
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, "$1")
    .replace(/^#{1,6}\s+/gm, "")
    .replace(/^>\s?/gm, "")
    .replace(/^[-*+]\s+/gm, "")
    .replace(/^\d+\.\s+/gm, "")
    .replace(/\*\*([^*]+)\*\*/g, "$1")
    .replace(/\*([^*]+)\*/g, "$1")
    .replace(/~~([^~]+)~~/g, "$1")
    .replace(/\r/g, "")
    .replace(/\n{2,}/g, "\n")
    .trim();
}

export function renderMarkdown(markdownText: string, options?: MarkdownRenderOptions) {
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

export function extractMarkdownExcerpt(markdownText: string, maxLength = 180) {
  const plainText = extractPlainText(markdownText);
  if (plainText.length <= maxLength) {
    return plainText;
  }

  return `${plainText.slice(0, maxLength).trim()}...`;
}

export function extractMarkdownHeadings(markdownText: string) {
  return markdownText
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter((line) => /^#{1,3}\s+/.test(line))
    .map((line) => {
      const [, rawHashes, rawTitle] = line.match(/^(#{1,3})\s+(.+)$/) ?? [];
      const title = rawTitle?.trim() ?? "";
      return {
        level: rawHashes?.length ?? 1,
        title,
        id: slugifyHeading(title)
      };
    })
    .filter((item) => item.title);
}
