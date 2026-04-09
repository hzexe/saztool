# Integrations

This document explains how to wrap `saztool` for MCP or skill-based workflows.

## 1. Turn saztool into an MCP server

`saztool` is a normal CLI today. To expose it as MCP, create a thin MCP wrapper that maps CLI commands to MCP tools.

Suggested MCP tools:
- `normalize`
- `show`
- `search`

Suggested tool mapping:
- MCP `normalize(input, outDir?)` -> `saztool normalize <input> [-out outDir]`
- MCP `show(bundleDir, sessionId, bodyPreview?)` -> `saztool show ...`
- MCP `search(bundleDir, query, in?, beforeId?, afterId?, context?, bodyPreview?)` -> `saztool search ...`

### Why a wrapper is better than exposing shell directly
- stable typed inputs
- easier schema discovery by AI tools
- no shell quoting confusion
- easier permission control

## 2. Turn saztool into an OpenClaw skill

A skill should not re-implement the parser. It should:
- decide when `saztool` is the right tool
- explain the recommended workflow
- optionally include helper scripts for repetitive commands

A useful saztool skill would trigger on phrases like:
- analyze Fiddler SAZ
- search capture
- inspect session id
- normalize packet capture
- 查 SAZ
- 查抓包
- 看某个 session

Recommended workflow for a skill:
1. normalize the archive once
2. use search to narrow to relevant sessions
3. use show on specific session ids
4. only then move to manual reverse engineering or browser-side debugging

## 3. Recommended MCP tool surface

If you wrap it for MCP, expose these inputs clearly:

### normalize
- `input`
- `outDir`

### show
- `bundleDir`
- `sessionId`
- `bodyPreview`

### search
- `bundleDir`
- `query`
- `in`
- `beforeId`
- `afterId`
- `context`
- `bodyPreview`
- `output`

For wrappers, prefer `output: "json"` so downstream agents can parse a stable machine-readable result.

## 4. AI-facing guidance

When an AI agent sees `saztool`, it should understand:
- SAZ is the source archive format
- normalized bundles are derived artifacts, not original evidence
- `show` is for one session
- `search` is the discovery command
- Fiddler id order and timeline order are different concepts

## 5. Practical OpenClaw path

For OpenClaw specifically, a good progression is:
1. use the CLI directly while behavior is still evolving
2. standardize on `--output json` for machine-driven flows
3. add a thin skill or MCP wrapper once the contract is stable

See also:
- [`CLI.md`](./CLI.md)
- [`OPENCLAW_TEMPLATES.md`](./OPENCLAW_TEMPLATES.md)

## 6. Good packaging pattern

If you later publish an MCP wrapper, keep this repo focused on the core CLI and create a companion repo like:
- `saztool-mcp`

This keeps the parsing/search core stable and reusable across:
- direct CLI use
- OpenClaw skills
- MCP servers
- other automation
