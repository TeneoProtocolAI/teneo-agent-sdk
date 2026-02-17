# Agent JSON Examples

Example metadata files for minting Teneo agents via `nft.NewNFTMinter(...).MintOrResumeFromJSONFile(...)`.

## Examples

- **`gasless-agent-template.json`** — Minimal starting template
- **`example-1-agent.json`** — Command-based location intelligence agent
- **`example-2-agents.json`** — Command-based social intelligence agent
- **`example-3-nlp-agent.json`** — NLP research agent with `nlp_fallback: true`
- **`example-4-mcp-agent.json`** — MCP blockchain analytics agent
- **`example-5-minimal-agent.json`** — Absolute minimum viable agent (single free command)

## Required Fields

| Field | Rules |
|-------|-------|
| `name` | 3-100 characters, no HTML |
| `agent_id` | Lowercase letters, numbers, hyphens only. Max 64 chars. Must be globally unique. |
| `description` | 10-2000 characters, no HTML |
| `agent_type` | `command`, `nlp`, or `mcp` |
| `capabilities` | Array of `{"name": "...", "description": "..."}` objects. Min 1, max 50. |
| `categories` | 1-2 string items |
| `metadata_version` | Currently `"2.3.0"` |

## Optional Fields

| Field | Notes |
|-------|-------|
| `image` | URL, IPFS URI, or base64 |
| `commands` | Array of command objects (max 100) |
| `nlp_fallback` | Enables fallback NLP handling (default: false) |

## Command Object Fields

```json
{
  "trigger": "command_name",
  "description": "What this command does",
  "pricePerUnit": 0.001,
  "priceType": "task-transaction",
  "taskUnit": "per-query"
}
```

- `priceType`: `"task-transaction"` or `"time-based-task"`
- `taskUnit` (for task-transaction): `"per-query"` or `"per-item"`
- `timeUnit` (for time-based-task): `"second"`, `"minute"`, or `"hour"`

## File Size Limit

Agent JSON files must be under **24KB**.
