# rummage — hero demo (README header GIF)
# Shows the one-take core loop: preview pane following the cursor, fuzzy search
# with live ranked highlighting, and one-keystroke select.
#
# Regenerate:  make demos      (or, from the repo root:  vhs docs/demo/hero.tape)
# Prereqs:     vhs + ffmpeg + ttyd, and `bin/basic` built (the make target does both).

Output docs/demo/hero.gif

Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 80ms

# Launch hidden so the GIF opens straight into the running TUI (no shell command,
# no compile flicker).
Hide
Type "bin/basic -preview -dir /tmp/rummage-demo"
Enter
Sleep 1.5s
Show

Sleep 1s
# Walk the cursor: the preview pane reflects each entry (dir listing -> file head).
Down@550ms 3
Sleep 900ms
Down@550ms 3
Sleep 900ms

# Fuzzy search: "srv" subsequence-matches server.go, ranked best-first, highlighted.
Type "/"
Sleep 400ms
Type@150ms "srv"
Sleep 1500ms
Enter
Sleep 700ms

# Select the match.
Enter
Sleep 2s
