# Security Policy

## Supported Versions

rummage follows semantic versioning. Security fixes are made on the latest
minor release; older releases do not receive backports.

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |
| Older   | :x:                |

## Reporting a Vulnerability

Please **do not** open a public issue for security reports.

Use [GitHub's private vulnerability reporting](https://github.com/rummage-dev/rummage/security/advisories/new)
to file a report. This routes the disclosure privately to the maintainers and
allows coordinated remediation before public disclosure.

Include in the report:

- A description of the issue and the impact (what an attacker can do)
- Steps to reproduce, or a minimal proof-of-concept
- The version (`go list -m github.com/rummage-dev/rummage`) and Go
  toolchain version (`go version`) you observed it on
- Any suggested remediation, if you have one

You can expect an initial response within five business days. Once a fix is
ready we will coordinate a release and credit the reporter in the advisory
unless you prefer to remain anonymous.

## Scope

In scope:

- The `finder` Go package itself
- Example programs under `examples/`

Out of scope:

- Vulnerabilities in upstream dependencies (please report those to the
  respective project — we will pick up fixes via Dependabot)
- Issues that require an already-compromised terminal or local code execution
  beyond the trust boundary of a normal CLI tool
