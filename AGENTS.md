# Hermes Coder — Project Context

## Architecture

Hermes Agent (brain) ↔ OpenCode (hands) via hermes-opencode-plugin.

## OpenCode Agent Team

Define in `.opencode/agents/`. Dispatch with `opencode run --agent <name>`

| Agent | Role | Permissions | Dispatch When |
|-------|------|-------------|---------------|
| `orchestrator` | Default primary agent. Team lead — plans, delegates, debriefs | edit: ask, bash: ask, task: allow | User gives a high-level request. Breaks down work, routes to sub-agents, produces debrief. |
| `frontend-dev` | Frontend implementation | edit: allow, bash: allow | TypeScript/React/Vue/CSS features, components, styling |
| `backend-dev` | Backend implementation | edit: allow, bash: allow | APIs, services, databases, business logic (Go/Python/Node/Rust) |
| `code-reviewer` | Code quality review (read-only) | edit: deny, bash: deny | After any implementation — checks correctness, maintainability, perf |
| `security-auditor` | Security audit (read-only) | edit: deny, bash: limited | Auth, payments, PII, APIs, deps — OWASP top 10 scanning |
| `designer` | UI/UX design | edit: allow, bash: deny | Before frontend work — bespoke designs, no AI-looking UIs |

## Workflow

1. User → orchestrator: high-level request
2. orchestrator plans, delegates to sub-agents
3. Coding agents implement
4. code-reviewer / security-auditor reviews
5. orchestrator debriefs user with summary + approval requests

## Coding Conventions

- TypeScript over JS, typed function components, CSS variables, mobile-first responsive
- Parameterized queries, structured logging, OpenAPI docs, proper error handling
- WCAG AA accessibility, semantic HTML, prefers-reduced-motion support
- All auth/PII/payment code gets security review
- No destructive filesystem ops, pushes to protected branches, or new deps without approval

---


