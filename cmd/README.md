# Agent Buddy

Shimming together agentic coding stuff.

Features:
- Installs skills from claude plugin marketplace repo into `~/.agents/skills`

## Usage

```sh
agent-buddy install racurry/neat-little-package     # Install skills from a shared skill repo into ~/.agents/skills
agent-buddy install racurry/neat-little-package --ref main   # Or pin a branch, tag, or commit explicitly
agent-buddy install racurry/neat-little-package --only mr-sparkle/config   # Use plugin/path selectors when names collide
```

## Contrib

```sh
brew install goreleaser           # Install goreleaser for building and releasing the tool
```
