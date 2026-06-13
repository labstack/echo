# Echo Roadmap

> **DRAFT** — this is a starting point for maintainers to edit, not a commitment.
> Dates and priorities are owned by the Echo team. Open a discussion to propose changes.

This document exists so the community can see where Echo is heading. Echo is
**actively maintained**. We publish releases regularly across two supported
lines — see [README](./README.md) badges for the latest version and most recent commit.

## Version policy

| Line | Status | Support |
|------|--------|---------|
| `v5` | **Current** (since 2026-01-18) | New features, fixes, and improvements. |
| `v4` | Maintenance / LTS | **Security and bug fixes until 2026-12-31.** No new features. |

Upgrading from v4? See [API_CHANGES_V5.md](./API_CHANGES_V5.md).

Echo supports the **latest four Go major releases** and may work with older versions.

## Now (in progress)

- Stabilizing the `v5` API surface through point releases.
- Documentation catch-up for v5 behavior changes (e.g. CORS / `RouteNotFound`
  behavior on groups — see #2950).
- Triaging and reducing the open issue / PR backlog.

## Next (under consideration)

These are frequently-requested items being discussed. Inclusion here is **not** a
commitment — each still needs design agreement before implementation:

- **Automatic `HEAD` for `GET` routes** (#2944, #2937) — opt-in, likely via an
  `OnAddRoute` hook so users keep control.
- **Rate limiter response metadata** — expose `Retry-After` / remaining quota
  through the store interface (#2961).
- **Real-IP / `Forwarded` header handling** improvements (#2744).
- **Proxy middleware** authorization-header handling (#2787).

## Later / exploratory

- Continued alignment with the Go standard library (`net/http`, `slog`).
- Reducing third-party surface where the stdlib now covers the need.

## How to influence the roadmap

- **Discuss before large PRs** — open a [Discussion](https://github.com/labstack/echo/discussions)
  or issue so we can agree on the design first.
- 👍 reactions on issues help us gauge demand.
- See [README → Contribute](./README.md#contribute) for contribution guidelines.
