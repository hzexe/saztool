# SAZ normalized bundle format

## Purpose

This bundle format is designed for AI-assisted analysis of Fiddler SAZ archives while preserving provenance.

## Core rules

1. The original `.saz` archive remains untouched.
2. `request.raw.txt` and `response.raw.txt` are verbatim copies from Fiddler `raw/*_c.txt` and `raw/*_s.txt`.
3. `response.body.decoded.txt` is canonical normalized text only after:
   - transfer decoding (`chunked` removal)
   - content decoding (`gzip` / `deflate` / `br`)
   - charset decoding
4. Canonical normalized output does **not** pretty-print JSON.
5. Session identity is preserved by `sessionId`.
6. Session order is preserved by `ordinal` and `manifest.fiddlerSessionOrder`.

## Session order semantics

Current version uses ascending Fiddler session id as the canonical session order.

Why:
- SAZ `raw/` filenames already encode session ids
- In common Fiddler captures, these ids reflect capture order
- This gives deterministic ordering without requiring XML-only parsing in v0.1

Future versions may additionally expose `timelineOrder` derived from `raw/*_m.xml` timestamps.

## AI guidance

If an AI reads this bundle, it should treat:
- `request.raw.txt` and `response.raw.txt` as source evidence
- `meta.json` as provenance and transform log
- `response.body.decoded.txt` as canonical searchable text for textual bodies
- `search` derivatives, when added later, as non-canonical helper views
