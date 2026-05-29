# rummage — preview pane demo
# The side pane previews the highlighted entry: file head for files, child
# listing for directories. Also emits a still (docs/demo/preview.png).
#
# Regenerate:  make demos   (or:  vhs docs/demo/preview.tape from the repo root)

Output docs/demo/preview.gif

Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"

Hide
Type "bin/basic -preview -dir /tmp/rummage-demo"
Enter
Sleep 1.5s
Show

Sleep 1s
# Walk down to a file (README.md, index 9) — preview shows its head.
Down@750ms 9
Sleep 1s
Screenshot docs/demo/preview.png
Sleep 500ms

# Back up to a directory (internal/, index 2) and open it — preview shows listings.
Up@500ms 7
Sleep 800ms
Right
Sleep 1500ms
Escape
Sleep 1s
