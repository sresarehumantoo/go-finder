# rummage — interactive create demo
# Create files and folders without leaving the picker (n = new file, N = new
# folder; a trailing "/" in a new-file name makes a directory). Also emits a
# still (docs/demo/interactive.png).
#
# Regenerate:  make demos   (or:  vhs docs/demo/interactive.tape from the repo root)
#
# To also demo deletion, append after the last step:
#   Type "d"   Sleep 600ms   Enter   Sleep 1500ms   (confirms the delete prompt)

Output docs/demo/interactive.gif

Set Shell "bash"
Set FontSize 18
Set Width 1200
Set Height 700
Set Padding 20
Set Theme "Catppuccin Mocha"

Hide
Type "bin/basic -interactive -dir /tmp/rummage-demo"
Enter
Sleep 1.5s
Show

Sleep 1s
# New file.
Type "n"
Sleep 500ms
Type@120ms "draft.md"
Sleep 700ms
Enter
Sleep 1300ms
Screenshot docs/demo/interactive.png
Sleep 400ms

# New directory via the trailing-slash shortcut.
Type "n"
Sleep 500ms
Type@120ms "archive/"
Sleep 700ms
Enter
Sleep 1800ms
