package service

import (
        "strings"
        "testing"
)

func TestParseMarkdownImportWithFrontmatter(t *testing.T) {
        svc := &KnowledgeService{}
        input := `---
title: Founder Knowledge OS
slug: founder-knowledge-os
section_name: Operating System
summary: A compact guide for founder-facing knowledge ops.
estimated_read_minutes: 6
is_preview: true
status: published
---

# Founder Knowledge OS

This is the body paragraph.
`

        result, err := svc.ParseMarkdownImport("founder-knowledge-os.md", []byte(input))
        if err != nil {
                t.Fatalf("expected no error, got %v", err)
        }

        if result.Title != "Founder Knowledge OS" {
                t.Fatalf("unexpected title: %s", result.Title)
        }
        if result.Slug != "founder-knowledge-os" {
                t.Fatalf("unexpected slug: %s", result.Slug)
        }
        if result.SectionName != "Operating System" {
                t.Fatalf("unexpected section: %s", result.SectionName)
        }
        if result.PublicSummary != "A compact guide for founder-facing knowledge ops." {
                t.Fatalf("unexpected summary: %s", result.PublicSummary)
        }
        if result.EstimatedReadMinutes != 6 {
                t.Fatalf("unexpected read minutes: %d", result.EstimatedReadMinutes)
        }
        if !result.IsPreview {
                t.Fatalf("expected preview to be true")
        }
        if result.Status != "published" {
                t.Fatalf("unexpected status: %s", result.Status)
        }
        if strings.Contains(result.ContentMarkdown, "# Founder Knowledge OS") {
                t.Fatalf("expected leading h1 to be trimmed from content")
        }
}

func TestParseMarkdownImportWithoutFrontmatter(t *testing.T) {
        svc := &KnowledgeService{}
        input := `# Restructure Blogs Into Course

Turn long-form blog posts into a course structure with modules, checkpoints, and assets.

## Module One

Start here.
`

        result, err := svc.ParseMarkdownImport("restructure blogs into course.md", []byte(input))
        if err != nil {
                t.Fatalf("expected no error, got %v", err)
        }

        if result.Title != "Restructure Blogs Into Course" {
                t.Fatalf("unexpected title: %s", result.Title)
        }
        if result.Slug != "restructure-blogs-into-course" {
                t.Fatalf("unexpected slug: %s", result.Slug)
        }
        if result.PublicSummary == "" {
                t.Fatalf("expected derived summary")
        }
        if result.EstimatedReadMinutes < 1 {
                t.Fatalf("expected positive read minutes, got %d", result.EstimatedReadMinutes)
        }
        if !strings.Contains(result.ContentMarkdown, "## Module One") {
                t.Fatalf("expected remaining body content to be preserved")
        }
}
