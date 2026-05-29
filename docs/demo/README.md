# Demo assets

Reproducible [VHS](https://github.com/charmbracelet/vhs) recordings used in the
project README. The GIFs/PNGs are committed; the browsed tree is generated on
demand at `/tmp/rummage-demo` (a neutral path, so the breadcrumb in the
recordings doesn't leak a real home directory).

## Regenerate

```bash
make demos
```

That builds the `basic` example, regenerates the fixture (via `fixture.sh`)
before each tape, runs every `*.tape`, and removes the fixture afterward.
Outputs land here.

Requires `vhs` on your PATH (which needs `ffmpeg` and `ttyd`):

```bash
go install github.com/charmbracelet/vhs@latest
```

You can also run a single recording from the repo root:

```bash
bash docs/demo/fixture.sh      # create the fixture at /tmp/rummage-demo first
vhs docs/demo/hero.tape        # then record one tape
rm -rf /tmp/rummage-demo       # cleanup
```

## Tapes

| Tape | Output | Shows |
|---|---|---|
| `hero.tape` | `hero.gif` | Core loop: preview pane + fuzzy search/highlight + select |
| `preview.tape` | `preview.gif`, `preview.png` | Preview pane (file head & directory listing) |
| `multi.tape` | `multi.gif`, `multi.png` | Multi-select persisting across directories |
| `interactive.tape` | `interactive.gif`, `interactive.png` | In-picker create (file + trailing-slash folder) |

All tapes drive a single `bin/basic` binary with different flags
(`-preview`, `-mode multi`, `-interactive`) against `-dir /tmp/rummage-demo`,
so the recordings are deterministic. Theme/size are set consistently at the top
of each tape — tweak there to restyle.
