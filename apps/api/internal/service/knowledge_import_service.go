package service

import (
        "bytes"
        "errors"
        "path/filepath"
        "regexp"
        "strconv"
        "strings"
        "time"
        "unicode"
        "unicode/utf8"
)

type KnowledgeMarkdownImportResult struct {
        Title                string `json:"title"`
        Slug                 string `json:"slug"`
        SectionName          string `json:"section_name"`
        EstimatedReadMinutes int    `json:"estimated_read_minutes"`
        PublicSummary        string `json:"public_summary"`
        ContentMarkdown      string `json:"content_markdown"`
        CoverURL             string `json:"cover_url"`
        IsPreview            bool   `json:"is_preview"`
        Status               string `json:"status"`
}

var markdownWordRegexp = regexp.MustCompile(`[A-Za-z0-9_]+`)
var markdownLeadingHeadingRegexp = regexp.MustCompile(`(?m)^#\s+(.+?)\s*$`)

func (s *KnowledgeService) ParseMarkdownImport(fileName string, content []byte) (KnowledgeMarkdownImportResult, error) {
        if len(bytes.TrimSpace(content)) == 0 {
                return KnowledgeMarkdownImportResult{}, errors.New("markdown file is empty")
        }
        if !utf8.Valid(content) {
                return KnowledgeMarkdownImportResult{}, errors.New("markdown file must be utf-8 encoded")
        }

        normalized := strings.ReplaceAll(string(content), "\r\n", "\n")
        normalized = strings.TrimPrefix(normalized, "\uFEFF")
        normalized = strings.TrimSpace(normalized)

        frontmatter, body := parseMarkdownFrontmatter(normalized)
        body = strings.TrimSpace(body)

        title := firstNonEmpty(frontmatter["title"], extractMarkdownTitle(body), humanizeFileName(fileName))
        slug := firstNonEmpty(frontmatter["slug"], buildKnowledgeImportSlug(title), buildKnowledgeImportSlug(fileName))
        if slug == "" {
                slug = "entry-" + strconv.FormatInt(time.Now().Unix(), 10)
        }

        if extractedTitle := extractMarkdownTitle(body); extractedTitle != "" {
                body = trimLeadingMarkdownTitle(body)
        }
        body = strings.TrimSpace(body)

        summary := firstNonEmpty(
                frontmatter["public_summary"],
                frontmatter["summary"],
                extractMarkdownSummary(body),
        )

        result := KnowledgeMarkdownImportResult{
                Title:                title,
                Slug:                 slug,
                SectionName:          firstNonEmpty(frontmatter["section_name"], frontmatter["section"]),
                EstimatedReadMinutes: parseEstimatedReadMinutes(frontmatter["estimated_read_minutes"], frontmatter["read_minutes"], body),
                PublicSummary:        summary,
                ContentMarkdown:      body,
                CoverURL:             firstNonEmpty(frontmatter["cover_url"], frontmatter["cover"]),
                IsPreview:            parseBoolLike(frontmatter["is_preview"]),
                Status:               normalizeImportedStatus(frontmatter["status"]),
        }

        if result.Title == "" {
                return KnowledgeMarkdownImportResult{}, errors.New("could not determine title from markdown file")
        }

        return result, nil
}

func parseMarkdownFrontmatter(content string) (map[string]string, string) {
        if !strings.HasPrefix(content, "---\n") {
                return map[string]string{}, content
        }

        lines := strings.Split(content, "\n")
        closingIndex := -1
        for i := 1; i < len(lines); i++ {
                if strings.TrimSpace(lines[i]) == "---" {
                        closingIndex = i
                        break
                }
        }

        if closingIndex <= 0 {
                return map[string]string{}, content
        }

        metadata := make(map[string]string, closingIndex-1)
        for _, line := range lines[1:closingIndex] {
                separator := strings.Index(line, ":")
                if separator <= 0 {
                        continue
                }

                key := strings.ToLower(strings.TrimSpace(line[:separator]))
                value := strings.TrimSpace(line[separator+1:])
                value = strings.Trim(value, `"'`)
                if key != "" && value != "" {
                        metadata[key] = value
                }
        }

        return metadata, strings.Join(lines[closingIndex+1:], "\n")
}

func extractMarkdownTitle(content string) string {
        matches := markdownLeadingHeadingRegexp.FindStringSubmatch(content)
        if len(matches) < 2 {
                return ""
        }
        return strings.TrimSpace(matches[1])
}

func trimLeadingMarkdownTitle(content string) string {
        content = strings.TrimSpace(content)
        if content == "" {
                return ""
        }

        firstLineEnd := strings.Index(content, "\n")
        if firstLineEnd < 0 {
                return ""
        }

        firstLine := strings.TrimSpace(content[:firstLineEnd])
        if !strings.HasPrefix(firstLine, "# ") {
                return content
        }

        trimmed := strings.TrimLeft(content[firstLineEnd+1:], "\n")
        return strings.TrimSpace(trimmed)
}

func extractMarkdownSummary(content string) string {
        if content == "" {
                return ""
        }

        paragraphs := strings.Split(content, "\n\n")
        for _, paragraph := range paragraphs {
                cleaned := normalizeMarkdownParagraph(paragraph)
                if cleaned == "" {
                        continue
                }
                if len([]rune(cleaned)) > 180 {
                        return string([]rune(cleaned)[:180]) + "..."
                }
                return cleaned
        }

        return ""
}

func normalizeMarkdownParagraph(paragraph string) string {
        lines := strings.Split(paragraph, "\n")
        cleanedLines := make([]string, 0, len(lines))
        for _, line := range lines {
                trimmed := strings.TrimSpace(line)
                if trimmed == "" {
                        continue
                }
                if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "```") {
                        continue
                }
                trimmed = strings.TrimLeft(trimmed, "-*0123456789.> ")
                trimmed = strings.ReplaceAll(trimmed, "`", "")
                trimmed = strings.ReplaceAll(trimmed, "**", "")
                trimmed = strings.ReplaceAll(trimmed, "__", "")
                trimmed = strings.ReplaceAll(trimmed, "*", "")
                trimmed = strings.ReplaceAll(trimmed, "_", "")
                trimmed = strings.TrimSpace(trimmed)
                if trimmed != "" {
                        cleanedLines = append(cleanedLines, trimmed)
                }
        }

        return strings.Join(cleanedLines, " ")
}

func parseEstimatedReadMinutes(values ...string) int {
        for _, value := range values[:len(values)-1] {
                parsed, err := strconv.Atoi(strings.TrimSpace(value))
                if err == nil && parsed > 0 {
                        return parsed
                }
        }

        content := values[len(values)-1]
        totalUnits := countMarkdownReadUnits(content)
        if totalUnits <= 0 {
                return 1
        }

        minutes := totalUnits / 220
        if totalUnits%220 != 0 {
                minutes++
        }
        if minutes < 1 {
                return 1
        }
        return minutes
}

func countMarkdownReadUnits(content string) int {
        latinWords := len(markdownWordRegexp.FindAllString(content, -1))
        hanCount := 0
        for _, r := range content {
                if unicode.Is(unicode.Han, r) {
                        hanCount++
                }
        }
        return latinWords + hanCount
}

func normalizeImportedStatus(status string) string {
        normalized := strings.TrimSpace(status)
        switch normalized {
        case "draft", "published", "archived":
                return normalized
        default:
                return "draft"
        }
}

func parseBoolLike(value string) bool {
        switch strings.ToLower(strings.TrimSpace(value)) {
        case "true", "1", "yes", "y", "on":
                return true
        default:
                return false
        }
}

func humanizeFileName(fileName string) string {
        base := strings.TrimSuffix(filepath.Base(strings.TrimSpace(fileName)), filepath.Ext(strings.TrimSpace(fileName)))
        base = strings.ReplaceAll(base, "_", " ")
        base = strings.ReplaceAll(base, "-", " ")
        base = strings.TrimSpace(base)
        return strings.Join(strings.Fields(base), " ")
}

func buildKnowledgeImportSlug(value string) string {
        base := strings.TrimSuffix(filepath.Base(strings.TrimSpace(value)), filepath.Ext(strings.TrimSpace(value)))
        base = strings.ToLower(base)

        var builder strings.Builder
        lastHyphen := false
        for _, r := range base {
                switch {
                case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
                        builder.WriteRune(r)
                        lastHyphen = false
                case r == '-' || r == '_' || unicode.IsSpace(r):
                        if builder.Len() > 0 && !lastHyphen {
                                builder.WriteByte('-')
                                lastHyphen = true
                        }
                }
        }

        return strings.Trim(builder.String(), "-")
}

func firstNonEmpty(values ...string) string {
        for _, value := range values {
                trimmed := strings.TrimSpace(value)
                if trimmed != "" {
                        return trimmed
                }
        }
        return ""
}
