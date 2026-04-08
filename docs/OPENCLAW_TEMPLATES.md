# OpenClaw templates

This file provides actionable templates for integrating `saztool` into OpenClaw workflows.

## 1. Skill template outline

A minimal OpenClaw skill can wrap the CLI without re-implementing parsing logic.

```md
---
name: saztool
description: Use saztool when the user wants to analyze Fiddler SAZ archives, inspect a session id, or search request/response/body content inside a normalized capture bundle.
---

# saztool

Use this skill when the user asks to:
- analyze a SAZ file
- inspect a capture session
- search within a Fiddler archive

## Recommended workflow
1. Run `saztool normalize <archive>` once
2. Run `saztool search <bundle> <query>` to find interesting sessions
3. Run `saztool show <bundle> <session-id>` on selected ids
```

## 2. MCP wrapper contract suggestion

If you create a separate `saztool-mcp` wrapper, expose these tools:

### `normalize`
Input:
```json
{
  "input": "capture.saz",
  "outDir": "capture.saz.norm"
}
```

### `show`
Input:
```json
{
  "bundleDir": "capture.saz.norm",
  "sessionId": 27,
  "bodyPreview": 800
}
```

### `search`
Input:
```json
{
  "bundleDir": "capture.saz.norm",
  "query": "sign",
  "in": "body",
  "beforeId": 200,
  "afterId": 100,
  "context": 2,
  "output": "json"
}
```

## 3. Recommended OpenClaw execution style

For tool-oriented automation:
- use `exec` for local CLI calls
- use `--output json` when downstream parsing matters
- prefer skill orchestration for multi-step analysis sessions

## 4. Why this matters

The CLI should remain the single source of truth for parsing, search, and normalization logic.
Wrappers and skills should stay thin.
