# Security Policy

## Supported Versions

Security fixes are applied to the default branch and the current production deployment branch.

## Reporting a Vulnerability

Please report suspected vulnerabilities privately instead of opening a public issue.

If GitHub private vulnerability reporting is enabled for this repository, use that channel. Otherwise, contact the maintainers through the repository owner profile and include:

- Affected endpoint, workflow, or component
- Reproduction steps
- Expected impact
- Any relevant logs or request IDs

Do not include live credentials, production tokens, or personal data in reports.

## Operational Requirements

- Never commit `.env`, `.env.local`, `.env.atlas`, `frontend/.env`, database dumps, API keys, or private certificates.
- Run secret scanning before making the repository public.
- Use `GIN_MODE=release` in production.
- Set production `CORS_ALLOWED_ORIGINS` to explicit HTTPS origins only.
- Store production secrets in the hosting platform or server secret manager, not in source control.
