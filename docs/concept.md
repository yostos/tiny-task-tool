# ttt (Tiny Task Tool) - Project Concept

Updated: 2026-01-22.

## About This Document

This document describes the vision and design philosophy of ttt (Tiny Task Tool).
For detailed specifications, see [specification.md](specification.md).

## Vision and Motivation

### Why Build This

When working in the terminal, there are moments when you want to write down the clutter in your head—random thoughts and tasks.
While coding, you might think "I need to do that too," or multiple tasks get tangled in your mind.

In such moments, physical paper and pen would let you write things down, but reaching for a pen while working in the terminal is inconvenient.
On the other hand, existing todo apps have too many features and end up interrupting your thinking instead.

What I want is a "digital sheet of paper."
A simple scratch space that opens with a single command from the terminal.

### Problems with Existing Tools

#### 1. Issues with Todo Apps

Many todo apps are feature-rich: project management, priority settings, deadlines, tags, categories, and more.
While thinking "which project should this go in?" "what priority?" "what tag?", you forget what you wanted to write.

Using brain resources to manage notes defeats the purpose.

#### 2. Issues with Markdown Editors

Obsidian, Notion, Joplin, and similar tools are powerful, but they also focus on "managing notes."
You need to think about what to name the file, which folder to save it in, how to link things.

When all you want is to "jot down today's thoughts," there's too much to think about.

#### 3. The Problem of Leaving the Terminal

Many existing tools are GUI applications.
When working in the terminal, you have to open another window, launch an app, write, then come back.
This "context switch" breaks the flow of thought.

#### 4. Issues with Similar CLI Tools

Terminal-based Markdown tools exist, but as features like file selection and multiple file support are added, they tend to drift away from the simplicity of "one sheet of paper."

In pursuing convenience, options multiply, and eventually you're forced to decide "which file should I write in?"

### The Experience I Want to Achieve

The ideal is the simplicity of physical memo paper.

There's a sheet of paper on your desk. You quickly write down what comes to mind.
When a task is done, you cross it out. When the paper fills up, you throw it away.
That's it.

I want to recreate this in the terminal.

In the morning, you open the terminal. With one command, today's paper appears.
You look at your tasks and check off completed ones.
When you want to write something new, you use your familiar editor (Vim, Neovim, etc.).
Completed tasks flow into the archive with a single command.

Only the features needed for organizing thoughts exist—nothing more.
A quiet place to organize your thinking without leaving the terminal.

## Design Philosophy

### Core Ideas

- The "One Sheet of Paper" Constraint
  - Physical memo paper is one sheet. You don't spread out multiple papers to manage them. That's why it's simple, with no hesitation.
  - This tool is the same. There's always one file.
    The choice of "which file should I write in?" is forcibly eliminated.
- Focus on "Now"
  - A place to write out the noise in your head—ideas, tasks, worries.
  - Capture thoughts the moment they occur. That's the purpose.
  - Features to search or organize past notes are unnecessary.
  - Writing organizes your thoughts and lightens your mind.
- "Node-Oriented" Management
  - People perceive indented blocks of lines as a single unit.
  - This tool handles tasks not as lines, but as nodes (structures including parent-child relationships).
  - Completing a parent completes its children; archiving moves children together.
  - Implementation becomes complex, but aligning with human intuition takes priority.
- Follow Unix Philosophy
  - "Do one thing well."
  - This tool only does "viewing Markdown" and "archiving tasks."
  - Editing is left to external editors. Searching is left to grep. Version control is left to git.
  - We don't try to solve everything in one tool.

### Most Important Values

1. Simplicity: Removing features has more value than adding them.
2. Eliminating friction: Startup and exit are instant, never interrupting thought.
3. Text-oriented: Limited to simple Markdown notation.
4. Coexistence with familiar tools: Editing is done with the user's preferred editor.
5. Aim for the simplest possible key operations.
6. Editor and key operations are customizable via configuration file.

### What We Do

- View Markdown files.
- Detect newly completed tasks (`- [x]`) and automatically add `@done(date)`
- Archive completed tasks
- Invoke external editor

### What We Don't Do

- File specification (no file specification via command-line arguments; eliminate the choice of "which file to open")
- Multiple file management (violates the one sheet of paper principle)
- In-file editing (leave to external editor)
- Toggle task completion (leave to external editor)
- Custom Markdown rendering implementation (library usage is acceptable; low development priority)
- Rich feature additions (search, tags, metadata, etc.)
- Cloud sync or AI features

### Intentionally Excluded

These will never be implemented, based on design philosophy.

- File specification via command-line arguments
- Multiple file support
- In-file editing functionality
- Task completion toggle (leave to editor)
- Custom Markdown rendering implementation (library usage is acceptable)
- Search/filter functionality
- Tags or metadata
- Cloud sync
- AI features

### Trade-off Decision Criteria

When a feature addition is proposed, ask yourself:

- Does this break the "one sheet of paper" experience?
- Can this be left to external tools?
- Is this really needed for "today"?

When in doubt, choose to remove.

## Target Users

### Expected Users

- Developers, writers, researchers who often work in the terminal
- Have a preferred editor like Vim, Neovim, Visual Studio Code
- Understand the value of a "simple memo pad"
- Tired of feature-rich todo apps and note apps

## Positioning of This Tool

### Comparison with Other Tools

| Aspect | Typical Todo Apps | Markdown Note Apps | This Tool |
| --- | --- | --- | --- |
| Feature richness | Many: project mgmt, tags, deadlines | Links, search, folder mgmt | Only viewing and archiving |
| File management | Managed within app | Multiple files/folders | One Markdown file |
| Editing environment | Dedicated editor | Dedicated editor | External editor (Vim, etc.) |
| Usage environment | GUI (Web/Desktop) | GUI (Web/Desktop) | TUI (Terminal) |
| Learning cost | High (many features) | Medium to high | Low (just a few keys) |

### Reasons to Choose This Tool

- Don't want to leave the terminal: Opens with one command.
- Want to write with a familiar editor: Your plugins and keybindings work as-is.
- Want simplicity: Having fewer features is the value.
- Prefer plain text: It's a Markdown file, readable by any tool.
