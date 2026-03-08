# The `claude-init` Script

The `bin/claude-init` script bootstraps the workflow into any project:

```bash
# Basic init (creates .claude/ directory structure + CLAUDE.md + .thoughts/)
claude-init

# Init with all agents, commands, and skills from your dotfiles
claude-init --all

# Init a specific project directory
claude-init --all ~/projects/my-app

# Init with .thoughts/ tracked in git (for team sharing)
claude-init --all --track-thoughts

# Update existing configs from dotfiles (preserves local changes)
claude-init --update

# Options
claude-init --force            # Overwrite existing .claude/
claude-init --no-claude-md     # Skip CLAUDE.md creation
claude-init --agents-only      # Only copy agents
claude-init --commands-only    # Only copy commands
claude-init --skills-only     # Only copy skills
claude-init --track-thoughts   # Don't gitignore .thoughts/ (track in git)
```

The script copies agents, commands, skills, and hooks from your global `~/.claude/` directory. Set the `DOTFILES_CLAUDE` environment variable to use a different source directory (e.g., `DOTFILES_CLAUDE=~/dotfiles/.claude claude-init --all`).

## RPI Binary

The `rpi` CLI binary is automatically built and installed during initialization:

- If `bin/rpi` doesn't exist in the source repo, it is built with `go build` (requires Go)
- The binary is installed to `~/.local/bin/rpi`
- If `~/.local/bin` is not in your PATH, the script adds it to your shell profile (`~/.zshrc`, `~/.bash_profile`, `~/.bashrc`, or `~/.profile`)
- Running `claude-init --update` rebuilds and reinstalls the binary
