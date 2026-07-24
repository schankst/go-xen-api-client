# Security

This is a personal fork of a XenAPI client library, not a security-critical
service. That said, if you find a real vulnerability (e.g. something that
could let a malicious XenAPI server compromise a client process, not just
a crash on malformed input), please open a private report via GitHub's
"Report a vulnerability" button under this repo's Security tab, rather
than a public issue, so there's time to fix it before it's public.

Dependency vulnerabilities are handled by Dependabot alerts on this repo;
as of v0.2.0 there are no external Go module dependencies left to alert on.
