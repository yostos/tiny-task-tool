# ttt (Tiny Task Tool) Demo
# VHS tape file for recording demo GIF

# Output settings
Output demo.gif
Set Width 800
Set Height 600
Set Shell fish
Set FontSize 14
Set TypingSpeed 50ms

# Start with clean prompt
Sleep 500ms

# Launch ttt
Type "ttt"
Sleep 1s
Enter
Sleep 1s

# Show help dialog
Type "h"
Sleep 2s

# Close help
Escape
Sleep 1s

# Enter edit mode (opens $EDITOR)
Type "e"
Sleep 1.5s

# In vim: go to end of file and add new section
Type "G"
Sleep 300ms
Type "o"
Enter
Type "## Project B"
Escape
Type "o"
Sleep 300ms

# Add first task
Type "- [ ] Prepare slides for Friday's tech talk"
Enter
Sleep 300ms

# Add subtasks (indented)
Tab
Type "- [ ] Create Outline"
Enter
Sleep 300ms
Type "- [ ] Add code examples"
Enter
Sleep 300ms
Type "- [ ] Practice run-through"
Escape
Sleep 500ms

# Add another task (back to main level)
Type "G"
Type "o"
Backspace 1
Type '- [ ] Read "Designing Data-Intensive Applications" Chapter 5'
Escape
Sleep 500ms

# Navigate up to mark a task complete
Type "gg"
Type "f["
Type "j"
Type "i"
Sleep 300ms

# Move to checkbox and mark complete
Right
Right
Backspace
Type "x"
Escape
Sleep 500ms

# Save and quit vim
Type ":wq"
Sleep 300ms
Enter
Sleep 5s

# Quit ttt
Type "q"
Sleep 2s
