# CasaOS-UserService

> ⚠️ **UNOFFICIAL FORK — NOT THE OFFICIAL RELEASE.** This is a **TofuGG** community
> fork of the CasaOS UserService, not affiliated with or endorsed by the official
> CasaOS / IceWhaleTech team. It is provided **AS-IS** with **no warranty** and
> **no official support**. **USE AT YOUR OWN RISK.**

[![Go Reference](https://pkg.go.dev/badge/github.com/IceWhaleTech/CasaOS-UserService.svg)](https://pkg.go.dev/github.com/IceWhaleTech/CasaOS-UserService) [![Go Report Card](https://goreportcard.com/badge/github.com/IceWhaleTech/CasaOS-UserService)](https://goreportcard.com/report/github.com/IceWhaleTech/CasaOS-UserService) [![goreleaser](https://github.com/IceWhaleTech/CasaOS-UserService/actions/workflows/release.yml/badge.svg)](https://github.com/IceWhaleTech/CasaOS-UserService/actions/workflows/release.yml) [![codecov](https://codecov.io/gh/IceWhaleTech/CasaOS-UserService/branch/main/graph/badge.svg?token=4GWJIF6FDD)](https://codecov.io/gh/IceWhaleTech/CasaOS-UserService)

User Service provides user management functionalities to CasaOS.

## Security Hardening

| # | Fix |
|---|-----|
| 11 | **JWT skipper removed** — no more loopback (`127.0.0.1` / `::1`) auth bypass on any route |
| 11 | **IP extractor hardened** — `echo.ExtractIPDirect()`; client-supplied `X-Forwarded-For` / `X-Real-IP` headers are never trusted |
| 11 | **Query-token fallback removed** — JWT is read from the `Authorization` header only, except the statically-rendered `<img>` asset routes (`/v1/users/avatar`, `/v1/users/current/image/:key`) which still accept a short-lived `?token=` for browser compatibility |
| 11 | **CORS tightened** — origins restricted to localhost/`127.0.0.1`, `AllowCredentials: false` |
| 13 | **Resilience** — malformed avatar uploads return `400` instead of killing the service (`log.Fatal` banned in request handlers); custom-config writes are nil-guarded when the message bus is absent; gateway management bootstrap retries (30×/500 ms) instead of panicking on a stale marker file |

> **Note:** build requires codegen first (`go generate ./...`); the generated `codegen/` packages are intentionally git-ignored.



## publish api to npm

### edit version in package.json

### run
```bash
yarn

yarn start
```

### publish

Manual publish
```bash
yarn publish
```

Auto publish
```bash 
git push origin dev**
```