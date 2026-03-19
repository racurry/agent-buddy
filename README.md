# Agent Buddy

Shimming together agentic coding stuff. Patching stupid holes.

Features:

- Installs skills from claude plugin marketplace repo into `~/.agents/skills`
- Creates `.mcpb` bundles for Claude Desktop from uvx-based MCP servers on PyPI

## Usage

```sh
# Install skills
agent-buddy install racurry/neat-little-package     # Install skills from a shared skill repo into ~/.agents/skills
agent-buddy install racurry/neat-little-package --ref main   # Or pin a branch, tag, or commit explicitly
agent-buddy install racurry/neat-little-package --only mr-sparkle/config   # Use plugin/path selectors when names collide

# Create MCP bundles for Claude Desktop
agent-buddy mcpb --from .mcp.json                   # Bundle all uvx servers from a .mcp.json file
agent-buddy mcpb --from .mcp.json mcp-obsidian       # Bundle a specific server by name
agent-buddy mcpb mcp-obsidian --env OBSIDIAN_API_KEY  # Bundle a PyPI package with env var config
agent-buddy mcpb mcp-obsidian -o my-obsidian.mcpb    # Custom output path
```

## Contrib

```sh
brew install goreleaser           # Install goreleaser for building and releasing the tool
```
