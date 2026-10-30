# Security Policy

## Supported Versions

| Version              | Bug fixes        | Security fixes                   |
|----------------------|------------------|----------------------------------|
| 5.x                  | Yes              | Yes                              |
| 4.x, latest release  | Until 2026-12-31 | Until 2027-12-31                 |
| 4.x, older releases  | No               | No: upgrade to the latest 4.x    |

Routine v4 bug fixes end on 2026-12-31. Free, public, security-only fixes for v4
continue until 2027-12-31. The plan for v4 after 2027-12-31 will be published by
2027-06-30; nothing beyond 2027 is promised today.

Security fixes cover confirmed vulnerabilities in Echo, including dependency and
compatibility changes needed to fix them. They ship as tagged releases with a
GitHub security advisory, to everyone at the same time. For v4 they ship as new
patch releases of the latest v4 version; older 4.x releases
do not get backports. Echo v4 requires Go 1.26 or later; build with a Go release
that the Go team still supports.

Client IP defaults in the
[proxy](https://github.com/labstack/echo/security/advisories/GHSA-99jh-6h7p-pp36)
and [rate limiter](https://github.com/labstack/echo/security/advisories/GHSA-246p-cpwv-v3jq)
advisories are fixed in v4.16.3. Upgrade older v4 releases to the latest v4 release.
Without `Echo#IPExtractor`, v4.16.3 and v5 use the connection peer and ignore
forwarded client IP headers. Applications behind proxies must configure a header
extractor that trusts only their own proxies; otherwise, clients behind one proxy
share its address for rate limiting, access controls, and logging. See the
[v4 migration example](https://github.com/labstack/echo/blob/v4.16.3/README.md#client-ip-security-migration-since-v4163).
Before v4.16.3, configuring an appropriate `Echo#IPExtractor` remains the workaround.

Paid [Echo Support](https://echo.labstack.com/pro#support) buys maintainer time.
It does not buy earlier or private fixes.

## Reporting a Vulnerability

Report vulnerabilities privately through GitHub:
https://github.com/labstack/echo/security/advisories/new

Do not open public issues, pull requests, or discussions for security problems.
