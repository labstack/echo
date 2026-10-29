# Echo Roadmap

> **DRAFT** — roadmap priorities are provisional; the published version and security support dates are commitments.
> Dates and priorities are owned by the Echo team. Open a discussion to propose changes.

This document exists so the community can see where Echo is heading. Echo is
**actively maintained**. We publish releases regularly across two supported
lines — see [README](./README.md) badges for the latest version and most recent commit.

## Version policy

| Line | Status | Support |
|------|--------|---------|
| `v5` | **Current** (since 2026-01-18) | New features, fixes, and improvements. |
| `v4` | Maintenance | **Bug fixes until 2026-12-31; security fixes until 2027-12-31** on the latest v4 release. No new features. See [SECURITY.md](./SECURITY.md). |

Upgrading from v4? See [API_CHANGES_V5.md](./API_CHANGES_V5.md).

Echo v5 supports the **three most recent Go releases** and may work with older versions. Echo v4 requires Go 1.26 or later. The version and security support dates in [SECURITY.md](./SECURITY.md) are a commitment; the rest of this roadmap is a draft.

## Now (in progress)

- Stabilizing the `v5` API surface through point releases.
- Documentation catch-up for v5 behavior changes (e.g. CORS / `RouteNotFound`
  behavior on groups — see #2950).
- Triaging and reducing the open issue / PR backlog.

## Next (under consideration)

These are frequently-requested items being discussed. Inclusion here is **not** a
commitment — each still needs design agreement before implementation:

- **Automatic `HEAD` for `GET` routes** (#2895; see #2949) — opt-in, likely via an
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
- See [README → Contributing](./README.md#contributing) for contribution guidelines.
