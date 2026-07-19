---
description: Go/Gin公式ガイドラインとアプリケーション安全不変条件に基づくbackendレビュー手順
---

# Go/Gin Backend Review

詳細な設計根拠は [go-gin-backend-guidelines.md](../rules/go-gin-backend-guidelines.md) を正本とする。この文書はレビュー時の実行用チェックリストであり、独自アーキテクチャを追加しない。

## 1. Package boundary

- package は凝集しており、名前が短く明確か。
- `util`、`common`、`misc` のような曖昧な package を増やしていないか。
- export は必要最小限か。package 名と export 名が stutter していないか。
- interface は利用側の必要なメソッドだけか。mock のためだけに先行作成されていないか。
- import cycle を避けるための不自然な global state や巨大 package を作っていないか。

配置を `handler/service/repository` や domain-first のどちらかへ機械的に矯正しない。変更理由、凝集性、依存方向、利用者を基準に判断する。

## 2. Request lifecycle

- request-scoped 処理は `c.Request.Context()` を下流へ伝播しているか。
- Context を struct に保存していないか。
- timeout/cancel 関数を確実に呼んでいるか。
- goroutine が元の `*gin.Context`、request body、response writer を保持していないか。
- goroutine に終了条件と error/panic の処理経路があるか。

## 3. HTTP boundary

- body/query/URI/header に合う binder を使い、binding error を処理しているか。
- 型、形式、長さ、範囲、列挙値を境界で検証しているか。
- authentication、authorization、resource ownership を別々に確認しているか。
- response を一度だけ書き、API contract と OpenAPI が一致しているか。
- internal model や秘密情報をそのまま返していないか。

## 4. Error and logging

- error を無視せず、必要なら `%w` で chain を保持しているか。
- 既知 error が安定した status/code に変換されるか。
- 未知 error は一般化した 500 となり、内部詳細を漏らさないか。
- 同じ error を複数箇所で重複ログしていないか。
- log に request correlation 情報はあるか。secret、token、個人情報はないか。
- panic recovery を通常の error 処理に使っていないか。

## 5. Database and security

- query は parameterized されているか。
- DB/外部 API は Context のキャンセルと deadline を受け取るか。
- transaction の commit/rollback と error path が明確か。
- CORS、cookie、CSRF、trusted proxy、rate limit が deployment/auth 方式に合っているか。
- [backend-application-invariants.md](backend-application-invariants.md) の clinic/owner/pet/staff 境界をすべての読み書きで維持しているか。

## 6. Server lifecycle

- production server に workload に合う timeout/limit があるか。
- SIGINT/SIGTERM から timeout 付き graceful shutdown が行われるか。
- `http.ErrServerClosed` を異常終了として扱っていないか。
- DB、worker、queue などの resource が安全な順序で閉じられるか。

## 7. Tests

- handler/middleware を `httptest` と最小 router で検証しているか。
- binding/validation/authentication/authorization/not-found/internal-error を含むか。
- dependency を差し替え可能で、global state に test が依存していないか。
- cancellation、concurrency、shutdown は変更 risk に応じて検証されているか。

## 判定時の禁止事項

- P1–P18 の旧番号を判定基準にしない。
- Handler → Service → Repository や Clean Architecture を Go/Gin 公式要件として強制しない。
- fixed directory tree、file size、interface 数、DI 場所を公式要件として指摘しない。
- GORM helper や project 独自 error helper を「Gin公式」と呼ばない。

指摘には、該当コード、実際の不具合または保守性リスク、根拠となる正本の節、最小の改善案を含める。
