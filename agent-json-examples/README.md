# Agent JSON Examples

Example metadata files for minting Teneo agents via `nft.Mint("agent-metadata.json")`.

## Examples

- **`gasless-agent-template.json`** — Starting template with parameters
- **`example-1-agent.json`** — Command-based location intelligence agent (multi-parameter commands)
- **`example-2-agents.json`** — Command-based social intelligence agent (username + count parameters)
- **`example-3-nlp-agent.json`** — NLP research agent with `nlpFallback: true`
- **`example-4-mcp-agent.json`** — MCP blockchain agent with code formatting tools
- **`example-5-minimal-agent.json`** — Absolute minimum viable agent (single free command, no parameters)
- **`example-6-commandless-agent.json`** — Commandless agent (no commands, freeform prompts)
- **`example-7-variants-and-param-types.json`** — Variants, advanced parameter types (url, boolean, interval, datetime, id, enum), and variadic parameters

## Required Fields

| Field | Rules |
|-------|-------|
| `name` | 3-100 characters, no HTML |
| `agentId` | Lowercase letters, numbers, hyphens only. Max 64 chars. Must be globally unique. |
| `shortDescription` | Brief one-line summary of what your agent does |
| `description` | 10-2000 characters, no HTML |
| `agentType` | `command`, `nlp`, `mcp`, or `commandless` |
| `capabilities` | Array of `{"name": "...", "description": "..."}` objects. Min 1, max 50. |
| `categories` | 1-2 items from valid list: `Trading`, `Finance`, `Crypto`, `Social Media`, `Lead Generation`, `E-Commerce`, `SEO`, `News`, `Real Estate`, `Travel`, `Automation`, `Developer Tools`, `AI`, `Integrations`, `Open Source`, `Jobs`, `Price Lists`, `Other`. **Case-sensitive. Invalid categories will block future updates.** |
| `metadata_version` | Currently `"2.4.0"` |

## Optional Fields

| Field | Notes |
|-------|-------|
| `image` | URL, IPFS URI, or base64 |
| `commands` | Array of command objects (max 100). Use `[]` for commandless agents. |
| `nlpFallback` | Enables fallback NLP handling (default: false) |

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
| `hasVariants` | No | `true` = this command has multiple variants (see [Variants](#variants)) |
| `variants` | No | Array of variant objects (required when `hasVariants: true`) |

### Parameter object fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Parameter name (matches the placeholder in `argument`) |
| `type` | Yes | `"string"`, `"number"`, `"username"`, `"url"`, `"boolean"`, `"interval"`, `"datetime"`, `"id"`, or `"enum"` |
| `required` | Yes | Whether this parameter is required |
| `description` | No | Human-readable description of what this parameter expects |
| `minValue` | No | Minimum value (for `number` type only) |
| `minLength` | No | Minimum string length (for `string`, `url`, `username` types) |
| `maxLength` | No | Maximum string length (for `string`, `url`, `username` types) |
| `minDuration` | No | Minimum duration (for `interval` type, e.g., `"30s"`, `"1h"`) |
| `maxDuration` | No | Maximum duration (for `interval` type, e.g., `"7d"`) |
| `includeTime` | No | `true` = expect ISO 8601 datetime with time component (for `datetime` type) |
| `options` | No | Array of allowed values (for `enum` type, e.g., `["5xx", "4xx", "non-200"]`) |
| `variadic` | No | `true` = accepts multiple values for this parameter |
| `minOccurrences` | No | Minimum count of values (for variadic parameters) |
| `maxOccurrences` | No | Maximum count of values (for variadic parameters) |
| `isBillingCount` | No | `true` = this parameter determines how many items are billed. Used with `taskUnit: "per-item"` — the final charge is `pricePerUnit x count`. Also works with variadic parameters (count = number of values provided). |

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

## Variants

Commands with `hasVariants: true` define multiple sub-commands under a single trigger. Each variant has its own arguments, parameters, and pricing. The user selects a variant when invoking the command.

```json
{
  "trigger": "alert",
  "description": "Configures alert rules. Uses variants for different alert modes.",
  "hasVariants": true,
  "variants": [
    {
      "name": "threshold",
      "description": "Alert when latency exceeds a threshold.",
      "argument": "<target> <max_latency_ms>",
      "parameters": [
        { "name": "target", "type": "url", "required": true, "description": "Monitored URL" },
        { "name": "max_latency_ms", "type": "number", "required": true, "minValue": "50", "description": "Max acceptable latency in ms" }
      ],
      "strictArg": true,
      "minArgs": 2,
      "maxArgs": 2,
      "pricePerUnit": 0.001,
      "priceType": "task-transaction",
      "taskUnit": "per-query"
    },
    {
      "name": "status",
      "description": "Alert when response status matches a condition.",
      "argument": "<target> <condition>",
      "parameters": [
        { "name": "target", "type": "url", "required": true, "description": "Monitored URL" },
        { "name": "condition", "type": "enum", "required": true, "options": ["5xx", "4xx", "non-200", "any-error"], "description": "Status condition to alert on" }
      ],
      "strictArg": true,
      "minArgs": 2,
      "maxArgs": 2,
      "pricePerUnit": 0.001,
      "priceType": "task-transaction",
      "taskUnit": "per-query"
    }
  ]
}
```

See `example-7-variants-and-param-types.json` for a full working example with variants, all parameter types, and variadic parameters.

## Variadic Parameters

A parameter with `"variadic": true` accepts multiple values. Use `minOccurrences` and `maxOccurrences` to constrain count. Variadic parameters can also be the billing count (`isBillingCount: true`), where the charge is `pricePerUnit x number_of_values`.

```json
{
  "trigger": "bulk_check",
  "description": "Check multiple URLs at once.",
  "argument": "<urls>...",
  "parameters": [
    {
      "name": "urls",
      "type": "url",
      "required": true,
      "variadic": true,
      "minOccurrences": 1,
      "maxOccurrences": 10,
      "isBillingCount": true,
      "description": "URLs to check (one or more)"
    }
  ],
  "strictArg": true,
  "minArgs": 1,
  "maxArgs": 10,
  "pricePerUnit": 0.001,
  "priceType": "task-transaction",
  "taskUnit": "per-item"
}
```

## File Size Limit

Agent JSON files must be under **24KB**.
