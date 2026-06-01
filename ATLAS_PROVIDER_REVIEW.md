# Atlas Cloud Provider Review

## Scope

- Added explicit `atlas` provider support on top of the existing OpenAI-compatible LLM path.
- Enabled environment variable expansion for LLM provider config fields so local secrets can stay out of version control.
- Updated docs and config examples to show the Atlas Cloud integration path.

## Code Changes

- `internal/config/config.go`
  - Added `GetProviderByName()` for case-insensitive provider lookup.
  - Added `ResolveEnvVars()` for `provider`, `base_url`, `auth_token`, `model`, `system_prompt`, and `custom_headers`.
  - Added Atlas example provider block in the built-in config template.

- `internal/config/settings.go`
  - Run `cfg.LLM.ResolveEnvVars()` in both normal and strict config parsing.

- `internal/executor/llm_executor.go`
  - Persist step-level `llm_config.provider`.
  - Honor explicit provider selection instead of always using rotation.
  - Return a clear error when a requested provider is not configured.

- `internal/executor/agent_executor.go`
  - Reused the same explicit provider selection behavior for agent LLM calls.

## Docs And Examples

- `README.md`
  - Added an Atlas Cloud provider section.
  - Added the required UTM link:
    - `https://www.atlascloud.ai/?utm_source=github&utm_medium=link&utm_campaign=osmedeus`

- `docs/api/llm.mdx`
  - Added Atlas Cloud setup instructions and example env vars.

- `public/presets/osm-settings.example.yaml`
  - Added commented `atlas` provider example.

- `public/examples/osmedeus-base.example/osm-settings.yaml`
  - Added commented `atlas` provider example.

- `build/docker/.env.example`
  - Added Atlas Cloud env variable examples.

## Local Secret Handling

- Real Atlas key should be stored only in a local, ignored file or shell environment.
- Recommended local values:
  - `ATLASCLOUD_BASE_URL=https://api.atlascloud.ai/v1/chat/completions`
  - `ATLASCLOUD_API_KEY=<local-secret>`
  - `ATLASCLOUD_MODEL=deepseek-ai/DeepSeek-V3-0324`

## Verification

- Added config/provider coverage in:
  - `internal/config/config_test.go`
  - `internal/executor/llm_streaming_test.go`

- Completed API-level verification:
  - Targeted Go tests passed for provider lookup, provider override, env expansion, and streaming paths.
  - Real Atlas Cloud request succeeded with model `deepseek-ai/DeepSeek-V3-0324`.
  - Verified response from `POST https://api.atlascloud.ai/v1/chat/completions` returned HTTP 200 and content `atlas-ok`.
