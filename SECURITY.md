# Security Policy

## Supported versions

The latest minor release on the `0.x` line receives security updates.
Once `1.0.0` ships, the two most recent minor releases will be
supported.

| Version  | Supported          |
| -------- | ------------------ |
| 0.1.x    | :white_check_mark: |

## Reporting a vulnerability

Please **do not** report security issues via public GitHub issues.

Instead, email the maintainer with:

- A description of the issue and its impact
- Steps to reproduce (or a proof-of-concept)
- The version of goanim affected
- Any suggested mitigation

You should receive an acknowledgement within 72 hours. If you don't,
please open a non-sensitive GitHub issue noting only that you sent a
report (so we know mail delivery failed) — do not include details.

## Scope

goanim is a rendering library. Realistic attack surfaces include:

- **Input parsing** — LaTeX source passed to `mathx.NewEquation`,
  manifest JSON read by data-driven example renderers
  (`examples/*_synced`), and any user-supplied font path. These run
  in-process; a crash or unbounded-memory issue is a valid report.
- **ffmpeg invocation** — the video encoder spawns `ffmpeg` via
  `os/exec`. Arguments are not currently user-controlled, but
  please report any path that lets a caller inject flags.
- **Disk I/O** — the LaTeX cache lives in `~/.cache/goanim/latex/`
  (or `$GOANIM_CACHE_DIR`). Report any path-traversal that lets a
  caller write outside that directory.

Out of scope: vulnerabilities in `ffmpeg` itself or transitive
dependencies — please report those upstream.
