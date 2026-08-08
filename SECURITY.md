# Security Policy

Osmedeus is a workflow engine that runs commands on your machine. That is what it is for. This
document draws the line between the code execution that is **intentional** and the code execution
that is a **vulnerability**, so operators know what they are running and reporters know what is
worth reporting.

The narrative version of the operator guidance lives at
[docs.osmedeus.org/others/security-warning](https://docs.osmedeus.org/others/security-warning).
This file is the canonical security model for the repository.

---

## Supported versions

Security fixes land on `main` and ship in the next release. Only the latest release is supported —
if you are running an older tag, upgrade with `osmedeus update` before reporting.

---

## Reporting a vulnerability

Report through [GitHub Security Advisories](https://github.com/j3ssie/osmedeus/security/advisories/new).

Please do not open a public issue, and please do not disclose publicly until a fix is available.
Include the endpoint or workflow involved, a reproduction, and what an attacker gains. A working
proof of concept gets a fix out much faster than a description of one.

Before reporting, check the [Intentional by design](#intentional-by-design) list below. Reports that
amount to "the workflow engine ran the command in the workflow" are closed as working-as-intended.

---

## The trust model

Osmedeus executes code because the **operator** asked it to. The operator is the person who runs the
binary, writes the workflows, and holds the API credentials. Everything the operator authors is
trusted; everything that arrives from outside is not.

| Input | Trust | Rule |
|-------|-------|------|
| Workflow YAML the operator wrote or installed | Trusted | May execute anything |
| The embedded binary registry (`public/presets/`) | Trusted | Its `valide-command` may be executed |
| CLI flags and local config | Trusted | Operator is at the keyboard |
| A registry supplied through the API (`registry_url`) | **Untrusted** | Never execute its entries |
| Scan targets, and anything a tool discovers about them | **Untrusted** | Never reaches a shell as code |
| Tool output parsed back in (SARIF, nmap XML, httpx JSON) | **Untrusted** | Data only |
| Unauthenticated HTTP requests | **Untrusted** | Must not cause execution |

**The test for whether something is a bug:** did the operator ask for this code to run? Osmedeus
running a command from a workflow the operator installed is the product. The same command running
because someone else triggered it, because a scan target was named a certain way, or because an
endpoint documented as read-only executed something — that is a vulnerability.

---

## Intentional by design

These are not vulnerabilities. They are the tool working.

**Workflows execute arbitrary code.** `bash`, `remote-bash`, `function`, `agent`, and `agent-acp`
steps run commands, scripts, and LLM-driven tool calls with the privileges of the Osmedeus process.
Never run a workflow you have not read. This is the same posture as Airflow, Argo, GitHub Actions,
and Jenkins.

**Utility functions execute code.** `osmedeus func e '<expr>'` and the functions API evaluate
expressions through a JavaScript runtime that includes `exec_python()`, `exec_ts()`, `tmux_run()`,
and `ssh_exec()`. Anyone who can call these can run commands. That is the feature.

**Installing binaries executes commands.** `osmedeus install` and
`POST /osm/api/registry-install` download binaries and run their install commands — including from a
registry you point at with `registry_url`. Installing *is* executing; a caller who can reach the
install endpoint and supply a registry can run commands by design. The CLI prints a security warning
before it does this. Treat install access as equivalent to shell access, and see the read-only
carve-out in the next section.

**Distributed workers execute dispatched work.** A worker that joins a master runs what the master
sends. `ssh_exec()` and the rsync helpers reach configured hosts. Only join masters you control.

**`--no-auth` disables authentication.** It exists for isolated development. Using it on a reachable
interface is an operator error, not a bug.

**Webhook triggers are unauthenticated.** When `server.enable_trigger_via_webhook` is on,
`/osm/api/webhook-runs/{uuid}/trigger` starts a run for anyone holding the UUID (plus the optional
auth key). The CLI warns about this when it prints a webhook URL. The UUID is the credential.

**Scans look like attacks.** Osmedeus sends traffic that IDS/WAF products will flag. Get
authorization for every target before you scan it.

---

## What *is* a vulnerability

Report these:

- **Execution from a read-only surface.** Any `GET`, or any endpoint documented as returning
  information, that causes a command to run. Read-only means read-only.
- **Execution from untrusted data.** A scan target, a hostname a tool discovered, a filename in an
  archive, or a field in parsed tool output reaching a shell as code.
- **Execution from an untrusted registry.** Entries loaded from a caller-supplied `registry_url`
  being executed anywhere other than an explicit install request.
- **Authentication bypass.** Reaching an authenticated endpoint without credentials, forging a JWT,
  or defeating the API key check.
- **Credential leakage.** `GITHUB_API_KEY`, cloud provider keys, or LLM API keys being sent to a host
  other than the intended one — including via URL parsing tricks, redirects, or log output.
- **Path traversal.** API parameters reading or writing outside the workspace, or archive extraction
  escaping its destination directory.
- **SSRF where the operator supplied no URL.** The operator pointing Osmedeus at an internal host is
  the feature; a request to an internal host they did not name is not.
- **Privilege escalation** between workspaces, users, or workers.

---

## Hardening a deployment

Fresh installs generate random values for `auth_api_key` (32 chars), `jwt.secret_signing_key`
(64 chars), and the default user password (12 chars), so there are no shipped default credentials.
To rotate them:

```bash
osmedeus config set server.password "$(openssl rand -hex 12)"
osmedeus config set server.jwt.secret_signing_key "$(openssl rand -hex 32)"
osmedeus config set server.auth_api_key "$(openssl rand -hex 24)"
```

| Area | Do this |
|------|---------|
| Exposure | Never put the API on the public internet. Bind to localhost or a private interface; reach it over a VPN or SSH tunnel |
| Transport | Terminate TLS at a reverse proxy in front of the server |
| Auth | Keep `server.enabled_auth_api` on and use the `x-osm-api-key` header for automation |
| Privileges | Run as a dedicated non-root user with the minimum filesystem access it needs |
| Workflows | Review before running; keep them in version control; `osmedeus workflow validate <name>` |
| Binaries | Install from the embedded registry or a registry you host; prefer Nix builds for reproducibility |
| Database | PostgreSQL with TLS in production; encrypt backups |
| Monitoring | Enable logging and audit access to the API |

### Known limitations

Be aware of these when deciding how to expose the server:

- **The browser session is not CSRF-hardened.** The `osmedeus_session` cookie is `SameSite=Lax` and
  CORS reflects any origin with credentials allowed. A top-level navigation from another site will
  carry the cookie. Prefer API-key auth for anything scripted, and do not leave a dashboard session
  open in a browser you also use for general browsing while the server is reachable.
- **The session cookie is readable by JavaScript** (`HTTPOnly=false`, so the UI can read login
  state) and is not marked `Secure` — another reason to terminate TLS at a proxy and keep the
  server off shared networks.

---

## For contributors

Invariants to preserve when touching these areas. Breaking one is a vulnerability, not a style
issue:

1. **`GET` handlers never execute.** If a handler can reach `exec.Command`, `sh -c`, or an installer
   command path, it must be a `POST` that says what it does.
2. **Only a trusted registry may be executed.** Use `installer.IsBinaryInstalled` for the embedded
   registry and `installer.IsBinaryInstalledNoExec` when the source came from a caller. If you add a
   code path that runs anything out of a `BinaryEntry`, gate it on the registry's provenance.
3. **Match hosts, never substrings, before attaching a token.** Use `installer.IsGitHubURL`, which
   compares the parsed hostname. `strings.Contains(url, "github.com")` sends your token to
   `evil.tld/?x=github.com`.
4. **Target data is data.** Interpolating `{{Target}}` or a discovered hostname into a shell command
   is the highest-risk pattern in the codebase. Quote it, or pass it through a file.
5. **Bound reads from the network.** Wrap response bodies in `io.LimitReader`; a caller-supplied URL
   should not be able to exhaust memory.
6. **New endpoints go in `docs/api/`** with their auth requirements and any execution side effects
   stated explicitly.

---

## Disclaimer

**Osmedeus is for authorized security testing only.** Unauthorized use may violate the law where you
live. By using it you accept that:

- **You need authorization.** Explicit permission before scanning any target, every time.
- **You are responsible.** For legal compliance and for every consequence of running this tool.
- **There is no warranty.** Provided "AS IS". The authors are not liable for damages, claims, or
  legal trouble arising from its use.
- **It executes code by design.** Review workflows before you run them.
- **Third-party tools have their own terms.** Comply with the licenses of everything Osmedeus
  integrates.
