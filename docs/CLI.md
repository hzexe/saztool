# CLI reference

## Commands

## `normalize`

Purpose:
- Read a Fiddler SAZ archive
- Export a normalized bundle without mutating the original capture

Usage:

```bash
saztool normalize <file.saz> [-out output_dir]
```

Parameters:
- `<file.saz>`: input SAZ archive or Fiddler-exported archive file
- `-out`, `--out`: output directory; default is `<input>.norm`

Notes:
- `normalize file.saz -out outdir` and `normalize -out outdir file.saz` are both supported
- Output preserves Fiddler session ids and canonical id order

## `show`

Purpose:
- Print a structured summary for a single session id

Usage:

```bash
saztool show <bundle_dir> <session-id> [-body-preview N]
```

Parameters:
- `<bundle_dir>`: normalized bundle directory
- `<session-id>`: Fiddler session id to inspect
- `-body-preview`: max preview characters for decoded textual body

Displays:
- sessionId / ordinal / timelineOrdinal when available
- method / url / statusCode
- contentType / transferEncoding / contentEncoding / charset
- statusMarkers / decodeFailureReason / truncationReason when present
- transforms
- decoded body preview when available

## `search`

Purpose:
- Search normalized content and/or raw request/response content across sessions

Usage:

```bash
saztool search <bundle_dir> <query> [options]
```

Options:
- `--in body|meta|request|response|all`
  - limit search scope
  - default when omitted: `body + meta`
- `--before-id N`
  - only search sessions with `sessionId < N`
- `--after-id N`
  - only search sessions with `sessionId > N`
- `-C N`, `--context N`
  - show N surrounding lines for each textual match
- `--body-preview N`
  - max preview characters used in summary preview
- `--output plain|grep|json`
  - choose display format
  - `plain`: rich human-readable default
  - `grep`: compact grep-like lines
  - `json`: machine-readable structured output

Search result fields:
- `session=`: Fiddler session id
- `ordinal=`: canonical id-order position
- `where=`: matched sources
- `body_lines=`, `meta_lines=`, `request_lines=`, `response_lines=`
- `match source=... line=... text=...`
- `context line=... text=...`

### Output modes

#### `plain`
Best for interactive human inspection.

#### `grep`
Compact output intended for shell pipelines and quick scanning. Format is roughly:

```text
<sessionId>:<source>:<line>:<text>
```

#### `json`
Structured result array intended for scripts, MCP wrappers, or AI tooling.

## Defaults and semantics

- Canonical order = ascending Fiddler session id
- Timeline order = derived from `raw/*_m.xml` when timing data exists
- Search defaults to `body + meta` because this is the most useful low-noise default for analysis
