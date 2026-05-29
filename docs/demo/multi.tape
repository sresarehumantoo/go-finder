# rummage — multi-select demo
# Selections persist across directory navigation (toggle with space/tab, confirm
# with enter). Also emits a still (docs/demo/multi.png).
#
# Regenerate:  make demos   (or:  vhs docs/demo/multi.tape from the repo root)

Output docs/demo/multi.gif

Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"

Hide
Type "bin/basic -mode multi -dir /tmp/rummage-demo"
Enter
Sleep 1.5s
Show

Sleep 1s
# Toggle two top-level files. NOTE: toggle (space) auto-advances the cursor by
# one, so the second hop is +2, not +3: 0 ->(3) config.go ->4 ->(2) main.go.
Down@450ms 3
Space
Sleep 600ms
Down@450ms 2
Space
Sleep 700ms
Screenshot docs/demo/multi.png
Sleep 400ms

# Jump up to internal/ (index 2) and open it — earlier selections persist.
Up@400ms 5
Sleep 500ms
Right
Sleep 1s
# Toggle both files inside (cache.go, store.go) to show cross-directory state.
Space
Sleep 600ms
Space
Sleep 800ms

# Enter on a file with selections present confirms the whole set (4 across 2 dirs).
Enter
Sleep 2s
