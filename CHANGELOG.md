# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-05-28

The biggest feature release yet: go-finder goes from a one-line picker to the
batteries-included **and** embeddable terminal picker for Go, while staying fully
backward-compatible. Existing `PickFile`/`PickFolder`/`PickAny`/`PickMultiple`
code keeps working unchanged.

### Added

- **Fuzzy search** — `/` filters with scored fuzzy matching (via `sahilm/fuzzy`),
  ranked best-match-first, with matched characters highlighted as you type. Opt
  out with `WithFuzzySearch(false)` for plain substring matching.
- **Pluggable filesystem** — `WithFS(fs.FS)` browses any `io/fs.FS` (e.g.
  `embed.FS`, `fstest.MapFS`), not just the host OS. Introduces the `FileSystem`
  and `WritableFileSystem` interfaces. Custom filesystems are read-only, so
  interactive create/delete degrades gracefully.
- **Preview pane** — opt-in `WithPreview(true)` shows a side pane previewing the
  highlighted entry (file head, directory listing, or metadata). Customizable via
  `WithPreviewFunc` and the new `PreviewFunc` type.
- **Embeddable component** — `New(...Option)` returns a Bubble Tea sub-model you
  can nest in your own program; it reports completion via `DoneMsg` and `State()`
  (with `Done()`) instead of quitting. Also adds `WithEmbedded`.
- New styles: `Styles.Preview`, `Styles.PreviewBorder`, `Styles.Match`.
- New examples: `examples/preview` and `examples/embedded`; a `-preview` flag on
  `examples/basic`.
- README "Why go-finder?" comparison table and `docs/POSITIONING.md`.

### Fixed

- Preview rendering no longer jumps or corrupts the display while scrolling:
  list and preview columns are fixed-height, the list is sized from the actual
  (wrap-aware) help-bar height, and a new `clampView` hard-limits every frame to
  the terminal width × height so no line can wrap and the frame can never
  overflow the screen at any terminal size.

### Compatibility

No breaking changes. Module path remains `github.com/SREsAreHumanToo/go-finder`.
Requires Go 1.25+.

## [0.1.2] - 2026-05-27

- Packaging and tooling: `doc.go` with runnable godoc examples, `.golangci.yml`,
  `SECURITY.md`, GoReleaser, and Codecov; CI test step uses bash on Windows.

## [0.1.1] - 2026-05-05

- Maintenance and dependency updates.

## [0.1.0] - 2026-03-11

- Initial release: cross-platform terminal file/folder picker built on Bubble
  Tea. File, folder, any, and multi-select modes; live search; interactive
  create/delete; hidden-file toggle; symlink expansion; WSL path helpers; and
  fully customizable keybindings and styles.

[0.2.0]: https://github.com/SREsAreHumanToo/go-finder/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/SREsAreHumanToo/go-finder/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/SREsAreHumanToo/go-finder/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/SREsAreHumanToo/go-finder/releases/tag/v0.1.0
