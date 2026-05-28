# Positioning & Roadmap

*Last updated: 2026-05-28*

This document captures where go-finder sits in the Go ecosystem, what makes it
worth choosing, and the sequenced plan for making it a library people can find
and rely on.

## The landscape

The Go terminal-picker space is dominated by Charm's ecosystem, but every
existing option is a *component* or a *binary* — none is a batteries-included,
importable picker library with a one-line API.

| Project | What it is | Multi-select | Search | Create/delete | Standalone API | File-or-folder |
|---|---|---|---|---|---|---|
| `charmbracelet/bubbles/filepicker` | Bubble Tea *component* (the default choice) | ❌ | ❌ | ❌ | ❌ embed-only | partial (`DirAllowed`, confusing — bubbles #659) |
| `charmbracelet/huh` filepicker field | Form *field* | ❌ | ❌ (open req #368) | ❌ | ❌ form-bound | ❌ |
| `charmbracelet/gum file` | CLI *binary*, not importable | ❌ | ❌ (open req #603) | ❌ | ✅ (as a binary) | partial |
| `ktr0731/go-fuzzyfinder` | fzf-style list finder (~500★) | ✅ `FindMulti` | ✅ fuzzy | ❌ | ✅ | ❌ (list finder, not a dir navigator) |
| `promptui` / `survey` / `go-prompt` | Older prompt libs | ❌ | ❌ | ❌ | ✅ | ❌ |
| **go-finder** | **Standalone picker library** | ✅ persists across dirs | ✅ live | ✅ | ✅ one-liner | ✅ `ModeAny` |

### Key takeaways

1. **The gap is real.** There is no popular Go library that gives you
   `path, err := finder.PickFile()` with a full picker behind it.
   `bubbles/filepicker` is deliberately minimal and embed-only; that is the seam
   go-finder slots into.
2. **Demand for our existing features is documented in Charm's own tracker.**
   Search/fuzzy-find (huh #368, gum #603), a real directory picker (bubbles
   #547/#659), and multi-select are all unmet asks elsewhere. We already ship
   live search, `ModeFolder`/`ModeAny`, and multi-select with cross-directory
   persistence.
3. **Charm wins adoption through composability.** Every bubbles package is a
   sub-model you embed. To win that audience we must offer an embeddable
   component story *in addition to* the one-liner.

## Differentiators (lead with these)

- **One import, one line.** `finder.PickFile()` / `PickFolder()` / `PickAny()` /
  `PickMultiple()` — no `Update`/`View` wiring required.
- **Multi-select that persists across directory navigation.**
- **Interactive create/delete** (no other picker library has this).
- **`ModeAny`** — select a file *or* a folder in one session.
- **WSL-aware path helpers** (`ToWindowsPath`/`ToWSLPath`/`IsWSL`).

## Gaps vs. the field (what to close)

These are the reasons someone might pick a competitor today:

1. **Substring search, not fuzzy.** `filterEntries` uses `strings.Contains`.
   The whole reason people want search is the fzf experience.
2. **No preview pane.** fzf, go-fuzzyfinder, ranger-likes all preview the
   highlighted item.
3. **Not embeddable.** `Model` implements `tea.Model` but assumes it owns the
   alt-screen; it can't be dropped into a larger TUI as a sub-model.
4. **No `io/fs` abstraction.** We call `os.*` directly. bubbles has an open ask
   (#815) for a mockable/embedded filesystem.
5. **Polish:** no icons/nerd-font support, no modtime column, no git-status
   tinting (bubbles only just added modtime in #807).

## Roadmap

All four agreed priorities, sequenced by leverage and dependency order.

### Phase 1 — Fuzzy search *(highest functional leverage, contained change)*

- Replace the `strings.Contains` path in `filterEntries` with scored fuzzy
  matching using **`sahilm/fuzzy`** — the same library `bubbles/list` uses for
  its default filter, so behavior is familiar to the Charm crowd and optimized
  for filenames.
- Rank filtered entries by score; preserve dirs-first as a secondary sort or
  expose a toggle.
- Optionally surface matched-character highlighting in `renderEntry` (sahilm
  returns matched indexes) — strong visual payoff for little cost.
- Add `WithFuzzySearch(bool)` (default on) so substring mode stays available.
- Tests: extend `search_test.go` with ranking/ordering and highlight cases.
- Alternative considered: `lithammer/fuzzysearch` (has `RankFind`) — rejected in
  favor of `sahilm/fuzzy` for ecosystem consistency and filename tuning.

### Phase 2 — `io/fs.FS` abstraction *(small surface, unlocks testing + Phase 3/4)*

- Introduce an internal filesystem interface; default to an `os`-backed impl,
  accept an `fs.FS` via `WithFS(fs.FS)`.
- Route `ReadDir` through `fs.ReadDir`; gate create/delete (write ops) on whether
  the FS supports them (degrade gracefully to read-only for `embed.FS`).
- Directly answers bubbles #815 and makes our own tests hermetic (drop the
  on-disk `createFile`/`createDir` helpers where possible).
- Do this before the preview pane so previews read through the same abstraction.

### Phase 3 — Preview pane *(the visible "wow" feature)*

- Right-hand pane previewing the highlighted entry: text head for files, child
  listing for directories, metadata (size, mode, modtime) for everything.
- Width-aware two-column layout in `view.go`; collapse below a min terminal
  width. Cap bytes read for large files; guard against binary files.
- `WithPreview(bool)` + a `PreviewFunc` hook so callers can customize (mirrors
  go-fuzzyfinder's `WithPreviewWindow`).
- Reads through the Phase 2 FS abstraction.

### Phase 4 — Embeddable component *(capture the "already building a TUI" audience)*

- Document and harden `Model` as a bubbles-style sub-model: no implicit
  alt-screen ownership, parent forwards `tea.Msg` to `Update`, calls `View`.
- Expose a "done/selected" signal the parent can poll (analogous to
  `filepicker.DidSelectFile`) instead of calling `tea.Quit` internally.
- Keep the one-liner API as a thin wrapper over the embeddable model.
- Add an `examples/embedded/` program showing it nested inside a larger TUI.

### Cross-cutting — discoverability

- Rewrite README to **lead with the one-liner pitch + the comparison table
  above**; that's the hook that wins search traffic.
- Submit to **awesome-go** (Command Line / Standard CLI sections).
- Keep `doc.go` + runnable `example_test.go` current so pkg.go.dev renders well.
- Consider polish items (icons, modtime column, git-status tinting) as fast
  follows once the four phases land.

## Sources

- [bubbles/filepicker (pkg.go.dev)](https://pkg.go.dev/github.com/charmbracelet/bubbles/filepicker)
- [bubbles #547 — directory picker](https://github.com/charmbracelet/bubbles/issues/547),
  [#659 — selecting a directory](https://github.com/charmbracelet/bubbles/issues/659),
  [#815 — pluggable filesystem](https://github.com/charmbracelet/bubbles/issues/815),
  [#807 — modtime](https://github.com/charmbracelet/bubbles/pull/807)
- [huh #368 — fuzzy find](https://github.com/charmbracelet/huh/issues/368),
  [gum #603 — filepicker fuzzy find](https://github.com/charmbracelet/gum/issues/603)
- [ktr0731/go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder) (`FindMulti`, `WithPreviewWindow`)
- [sahilm/fuzzy](https://github.com/sahilm/fuzzy),
  [lithammer/fuzzysearch](https://github.com/lithammer/fuzzysearch)
- [io/fs (pkg.go.dev)](https://pkg.go.dev/io/fs), [embed (pkg.go.dev)](https://pkg.go.dev/embed)
- [bubbles/list README — sub-model pattern](https://github.com/charmbracelet/bubbles/blob/master/list/README.md)
- [Awesome Go — Command Line](https://awesome-go.com/command-line/)
