# Agent JSON Examples

Example metadata files for minting Teneo agents via `nft.Mint("agent-metadata.json")`.

## Examples

- **`gasless-agent-template.json`** — Starting template with parameters
- **`example-1-agent.json`** — Command-based location intelligence agent (multi-parameter commands)
- **`example-2-agents.json`** — Command-based social intelligence agent (username + count parameters)
- **`example-3-nlp-agent.json`** — NLP research agent with `nlp_fallback: true`
- **`example-5-minimal-agent.json`** — Absolute minimum viable agent (single free command, no parameters)

## Required Fields

| Field | Rules |
|-------|-------|
| `name` | 3-100 characters, no HTML |
| `agent_id` | Lowercase letters, numbers, hyphens only. Max 64 chars. Must be globally unique. |
| `description` | 10-2000 characters, no HTML |
| `agent_type` | `command`, `nlp`, or `mcp` |
| `capabilities` | Array of `{"name": "...", "description": "..."}` objects. Min 1, max 50. |
| `categories` | 1-2 items from valid list: `Trading`, `Finance`, `Crypto`, `Social Media`, `Lead Generation`, `E-Commerce`, `SEO`, `News`, `Real Estate`, `Travel`, `Automation`, `Developer Tools`, `AI`, `Integrations`, `Open Source`, `Jobs`, `Price Lists`, `Other`. **Case-sensitive. Invalid categories will block future updates.** |
| `metadata_version` | Currently `"2.4.0"` |

## Optional Fields

| Field | Notes |
|-------|-------|
| `image` | URL, IPFS URI, or base64 |
| `commands` | Array of command objects (max 100) |
| `nlp_fallback` | Enables fallback NLP handling (default: false) |

## Command Object Fields

Every command must have `trigger`, `description`, and pricing. Use `parameters` to define what arguments your command accepts.

### Full command structure

```json
{
  "trigger": "timeline",
  "argument": "<username> <count>",
  "description": "Retrieves user's recent posts with engagement metrics.",
  "parameters": [
    {
      "name": "username",
      "type": "username",
      "required": true,
      "description": "Social media handle (without @)"
    },
    {
      "name": "count",
      "type": "number",
      "required": true,
      "minValue": "1",
      "description": "Number of posts to fetch",
      "isBillingCount": true
    }
  ],
  "strictArg": true,
  "minArgs": 2,
  "maxArgs": 2,
  "pricePerUnit": 0.001,
  "priceType": "task-transaction",
  "taskUnit": "per-item"
}
```

### Command fields

| Field | Required | Description |
|-------|----------|-------------|
| `trigger` | Yes | Command name the user types (e.g., `"timeline"`, `"search"`) |
| `description` | Yes | What this command does |
| `pricePerUnit` | Yes | USDC amount per task. `0` = free. |
| `priceType` | Yes | `"task-transaction"` (pay per use) or `"time-based-task"` (pay per time) |
| `taskUnit` | Yes | `"per-query"` (flat fee) or `"per-item"` (price x count) |
| `argument` | No | Template showing expected args (e.g., `"<username> <count>"`) |
| `parameters` | No | Array of parameter objects (see below). Use `[]` for no-argument commands. |
| `strictArg` | No | `true` = enforce argument count validation |
| `minArgs` | No | Minimum number of arguments required |
| `maxArgs` | No | Maximum number of arguments allowed |

### Parameter object fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Parameter name (matches the placeholder in `argument`) |
| `type` | Yes | `"string"`, `"number"`, or `"username"` |
| `required` | Yes | Whether this parameter is required |
| `description` | No | Human-readable description of what this parameter expects |
| `minValue` | No | Minimum value (for `number` type only) |
| `isBillingCount` | No | `true` = this parameter determines how many items are billed. Used with `taskUnit: "per-item"` — the final charge is `pricePerUnit x count`. |

### Pricing + parameters interaction

- **`per-query`**: User pays `pricePerUnit` once per command execution, regardless of parameters.
- **`per-item`** + **`isBillingCount`**: User pays `pricePerUnit x count`, where `count` comes from the parameter marked `isBillingCount: true`.

Example: `timeline elonmusk 50` with `pricePerUnit: 0.001` and `taskUnit: "per-item"` → user pays `0.001 x 50 = $0.05`.

### Common command patterns

**No arguments** (e.g., help, ping):
```json
{
  "trigger": "help",
  "description": "Lists all commands.",
  "parameters": [],
  "strictArg": true,
  "minArgs": 0,
  "maxArgs": 0,
  "pricePerUnit": 0,
  "priceType": "task-transaction",
  "taskUnit": "per-query"
}
```

**Single string argument** (e.g., lookup by ID):
```json
{
  "trigger": "post_stats",
  "argument": "<ID_or_URL>",
  "description": "Returns engagement metrics for a post.",
  "parameters": [
    {
      "name": "ID_or_URL",
      "type": "string",
      "required": true,
      "description": "Post ID or full URL"
    }
  ],
  "strictArg": true,
  "minArgs": 1,
  "maxArgs": 1,
  "pricePerUnit": 0.04,
  "priceType": "task-transaction",
  "taskUnit": "per-query"
}
```

**Username + count** (e.g., paginated social data):
```json
{
  "trigger": "followers",
  "argument": "<username> <count>",
  "description": "Retrieves user's followers list.",
  "parameters": [
    {
      "name": "username",
      "type": "username",
      "required": true,
      "description": "Social media handle"
    },
    {
      "name": "count",
      "type": "number",
      "required": true,
      "minValue": "1",
      "description": "Number of followers to return",
      "isBillingCount": true
    }
  ],
  "strictArg": true,
  "minArgs": 2,
  "maxArgs": 2,
  "pricePerUnit": 0.0005,
  "priceType": "task-transaction",
  "taskUnit": "per-item"
}
```

## File Size Limit

Agent JSON files must be under **24KB**.
