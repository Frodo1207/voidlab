#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DB_PATH="${1:-$ROOT_DIR/data/sqlite/voidlab.db}"

if ! command -v sqlite3 >/dev/null 2>&1; then
  echo "sqlite3 is required to seed local data" >&2
  exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "python3 is required to seed local data" >&2
  exit 1
fi

if [ ! -f "$DB_PATH" ]; then
  echo "Database not found: $DB_PATH" >&2
  echo "Start the API once to bootstrap the local SQLite schema." >&2
  exit 1
fi

export DB_PATH

python3 - << 'EOF'
import sqlite3
import os
import sys
import datetime
import hashlib

db_path = os.environ.get('DB_PATH')
conn = sqlite3.connect(db_path)
cursor = conn.cursor()

now = datetime.datetime.utcnow().strftime("%Y-%m-%d %H:%M:%S")

def sha256_hash(text):
    return hashlib.sha256(text.encode('utf-8')).hexdigest()

# Clean up
cursor.execute("DELETE FROM knowledge_access_pass_spaces")
cursor.execute("DELETE FROM knowledge_access_passes")
cursor.execute("DELETE FROM knowledge_entries")
cursor.execute("DELETE FROM knowledge_spaces")
cursor.execute("DELETE FROM articles")
cursor.execute("DELETE FROM events")
cursor.execute("DELETE FROM builders")

# Insert Spaces
spaces = [
    (1, "Agent Builder Playbook", "agent-builder-playbook", "A comprehensive guide to building autonomous agents.", "Playbook", "🤖", "bg-emerald-100", "directory_only", "Explore chapters below", 
    """# Agent Builder Playbook

Welcome to the ultimate guide for building autonomous AI agents. This playbook is designed to take you from foundational concepts to advanced production deployments.

## What's Inside

- **Core Concepts**: Understanding ReAct, Plan-and-Execute, and memory structures.
- **Tools & Environments**: Giving your agent hands and eyes.
- **Production Readiness**: Logging, tracing, and handling failures gracefully.

![Agent Architecture](https://images.unsplash.com/photo-1620712943543-bcc4688e7485?q=80&w=800&auto=format&fit=crop)
""", 
    "Enter the access token to unlock full chapters.", "https://images.unsplash.com/photo-1673847402636-8e50334df531?q=80&w=2000&auto=format&fit=crop", "published", 1, 1, now, now),
    
    (2, "Design System", "design-system", "Our core design principles and component guidelines.", "Design", "✨", "bg-orange-100", "directory_only", "UI components", 
    """# VOIDLAB Design System

This space contains all our design decisions, component specifications, and brand guidelines.
""", 
    "Enter token", "https://images.unsplash.com/photo-1561070791-2526d30994b5?q=80&w=2000&auto=format&fit=crop", "published", 1, 1, now, now)
]

cursor.executemany("""
    INSERT INTO knowledge_spaces 
    (id, title, slug, description, cover_label, icon, theme_tint, visibility_mode, directory_summary, intro_markdown, token_hint, cover_url, status, created_by, updated_by, created_at, updated_at) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
""", spaces)

# Insert Entries
entries = [
    (1, 1, "01-introduction", "01-introduction", "01. Fundamentals", 10, 5, "What is an agent?", 
    """# 1. Introduction to Agents

An autonomous agent is a system that can perceive its environment, make decisions, and take actions to achieve specific goals.

## Core Loop

The core loop of most LLM-based agents looks like this:

1. **Perceive**: Receive input from the user or environment.
2. **Think**: Reason about what to do next.
3. **Act**: Execute a tool or return a response.

### Example Code

Here is a simple example of an agent loop in TypeScript:

```typescript
async function runAgentLoop(task: string) {
  let isDone = false;
  let context = task;
  
  while (!isDone) {
    const thought = await llm.reason(context);
    
    if (thought.action === "FINISH") {
      isDone = true;
      return thought.result;
    }
    
    const observation = await tools.execute(thought.action, thought.params);
    context += `\nObservation: ${observation}`;
  }
}
```

## Types of Memory

Agents typically use different memory structures:
- **Short-term memory**: The current context window.
- **Long-term memory**: Vector databases or knowledge graphs.
- **Episodic memory**: Past experiences and conversations.

![Memory Structures](https://images.unsplash.com/photo-1677442136019-21780ecad995?q=80&w=800&auto=format&fit=crop)
""", 
    "", 0, "published", 1, 1, now, now),
    
    (2, 1, "02-tool-calling", "02-tool-calling", "02. Execution", 20, 8, "How to give agents tools.", 
    """# 2. Tool Calling

Tools are the hands of the agent.

## Defining a Tool

When defining a tool, the schema must be crystal clear to the LLM.

```json
{
  "name": "get_weather",
  "description": "Get the current weather for a location",
  "parameters": {
    "type": "object",
    "properties": {
      "location": {
        "type": "string",
        "description": "The city and state, e.g., San Francisco, CA"
      }
    },
    "required": ["location"]
  }
}
```

## Nesting and Complex Schemas

- Nested properties
  - Level 1
    - Level 2
      - Level 3

You can also use lists.

1. First step
2. Second step
   1. Sub step
   2. Sub step 2
""", 
    "", 0, "published", 1, 1, now, now),
    
    (3, 2, "color-palette", "color-palette", "Foundations", 10, 3, "Our color system.", 
    """# Color Palette

We use a minimal palette inspired by Brutalism and Notion.

- **Background**: `#f9f5ed` (Beige)
- **Border**: `#e9e5df`
- **Text Primary**: `#37352f`
- **Accent**: `#0f7b6c` (Green)
""", 
    "", 0, "published", 1, 1, now, now),

    (4, 2, "markdown-test", "markdown-test", "Foundations", 30, 10, "A comprehensive test of all markdown elements.", 
    """# Markdown UI Test Page

This page is designed to test how various Markdown elements render in our frontend knowledge base. It includes tables, multiple code languages, images, and complex nested structures.

## 1. Tables

Here is a comparison of different frontend frameworks we evaluated:

| Framework | Reactivity | Community Size | Learning Curve | 
| :--- | :---: | :---: | :--- |
| **Vue 3** | Proxy-based | Massive | Low |
| **React** | Virtual DOM | Massive | Medium |
| **Svelte** | Signals (v5) | Growing | Low |
| **Solid** | Compile-time | Niche | Low |

> **Note**: This table should render with proper borders, padding, and alternating row colors if configured in our CSS.

## 2. Code Snippets

We need to make sure syntax highlighting works for various languages.

### TypeScript / Vue

```vue
<script setup lang="ts">
import { ref, computed } from 'vue'

const count = ref(0)
const doubleCount = computed(() => count.value * 2)

function increment() {
  count.value++
}
</script>

<template>
  <button @click="increment">
    Count is: {{ count }}
  </button>
</template>
```

### Python (Backend Scripting)

```python
import sqlite3
import json

def fetch_users(db_path: str):
    with sqlite3.connect(db_path) as conn:
        cursor = conn.cursor()
        cursor.execute("SELECT id, username FROM users")
        return [{"id": row[0], "username": row[1]} for row in cursor.fetchall()]
```

### Go (API Service)

```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, VOIDLAB!")
    })
    http.ListenAndServe(":8080", nil)
}
```

## 3. Rich Typography & Media

Here is a paragraph with **bold text**, *italic text*, and ~~strikethrough~~. You can also use `inline code` to highlight variables like `const a = 1`.

### Unsplash Images

![Minimalist Architecture](https://images.unsplash.com/photo-1486406146926-c627a92ad1ab?q=80&w=2000&auto=format&fit=crop)

*Caption: Minimalist architecture often aligns with Brutalist design principles.*

## 4. Complex Lists

Let's test nested lists again, combining ordered and unordered lists:

1. First main step
   - Sub-task A
   - Sub-task B
     - Detail 1
     - Detail 2
2. Second main step
   1. Numbered sub-step 1
   2. Numbered sub-step 2
3. Third main step
""", 
    "", 0, "published", 1, 1, now, now)
]

cursor.executemany("""
    INSERT INTO knowledge_entries 
    (id, space_id, title, slug, section_name, sort_order, estimated_read_minutes, public_summary, content_markdown, cover_url, is_preview, status, created_by, updated_by, created_at, updated_at) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
""", entries)

# Insert Token
vip_token = "voidlab_vip_2026"
vip_hash = sha256_hash(vip_token)

cursor.execute("""
    INSERT INTO knowledge_access_passes 
    (id, name, token_hash, access_level, scope_type, is_active, created_by, created_at, updated_at) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
""", (1, "VIP Test Pass", vip_hash, "vip", "all_published", 1, 1, now, now))


# Insert Articles
articles = [
    (1, "The Dawn of Autonomous Workflows", "dawn-of-autonomous", "How autonomous agents are reshaping our daily workflows.", "AI", "General", '["AI", "Workflow"]', "https://images.unsplash.com/photo-1620712943543-bcc4688e7485?q=80&w=800&auto=format&fit=crop", 
    """# The Dawn of Autonomous Workflows

We are entering a new era where software doesn't just assist us, it operates on our behalf.

## The Shift

In the past, we used software as a tool. Now, software is becoming an actor. This fundamentally changes how we approach productivity.

> "We are moving from software as a bicycle for the mind, to software as an engine for action."

![Workflow](https://images.unsplash.com/photo-1551288049-bebda4e38f71?q=80&w=800&auto=format&fit=crop)

### Key Drivers

1. **Better Reasoning**: LLMs can now plan and execute.
2. **Tool Integration**: APIs are everywhere.
3. **Context Windows**: We can feed massive amounts of context into a single prompt.
""", 
    "VOIDLAB", "https://voidlab.ai", 1, "published", now, 1, 1, now, now),
    
    (2, "Designing for Agents", "designing-for-agents", "UI/UX considerations when building agentic interfaces.", "Design", "Builders", '["Design", "UX"]', "https://images.unsplash.com/photo-1561070791-2526d30994b5?q=80&w=800&auto=format&fit=crop", 
    """# Designing for Agents

How do you design an interface for something that does the work for you?

- Transparency is key.
- Allow interruption.
- Make the "thinking" visible.
""", 
    "VOIDLAB", "https://voidlab.ai", 0, "published", now, 1, 1, now, now)
]

cursor.executemany("""
    INSERT INTO articles 
    (id, title, slug, summary, category, audience, tags_json, cover_url, content, source_name, source_url, featured, status, published_at, created_by, updated_by, created_at, updated_at) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
""", articles)

# Insert Events
events = [
    (1, "Agent Builders Meetup", "agent-builders-meetup-1", now, "Shanghai", "Shanghai", "Offline", "published", "A gathering of developers building autonomous agents.", "https://images.unsplash.com/photo-1540575467063-178a50c2df87?q=80&w=800&auto=format&fit=crop", 
    """# Agent Builders Meetup

Join us for an afternoon of deep technical dives into building reliable agents.

## Agenda

- 14:00 - Welcome & Intro
- 14:30 - Building Reliable Tools
- 15:30 - Memory Architecture
- 16:30 - Networking
""", 1, 1, now, now)
]

cursor.executemany("""
    INSERT INTO events 
    (id, title, slug, event_time, location, city, event_type, status, summary, cover_url, content, created_by, updated_by, created_at, updated_at) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
""", events)

# Insert Builders
builders = [
    (1, "Alice Chen", "alice-chen", "AI Researcher", "Shanghai", "Researcher", "Building the next generation of reasoning models.", "Alice has been working on NLP since 2018.", '["NLP", "Agents"]', '["Reasoning", "Planning"]', '["Open Source", "Advising"]', "Available for consulting", "Startups", 1, 1, "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=800&auto=format&fit=crop", "published", 1, 1, now, now),
    (2, "Bob Smith", "bob-smith", "Frontend Engineer", "Remote", "Engineer", "Crafting beautiful UI for AI.", "Bob specializes in creating intuitive interfaces for complex AI systems.", '["Vue", "Tailwind"]', '["UX", "Animation"]', '["Freelance"]', "Booking for Q3", "Projects", 1, 0, "https://images.unsplash.com/photo-1599566150163-29194dcaad36?q=80&w=800&auto=format&fit=crop", "published", 1, 1, now, now)
]

cursor.executemany("""
    INSERT INTO builders 
    (id, name, slug, title, city, role, intro, story, expertise_json, focus_areas_json, collaboration_modes_json, availability_note, open_for, contactable, featured, cover_url, status, created_by, updated_by, created_at, updated_at) 
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
""", builders)

conn.commit()

print("==================================================")
print("✅ Local data seeded successfully!")
print("==================================================")
print(f"VIP Testing Token: {vip_token}")
print("==================================================")
EOF
