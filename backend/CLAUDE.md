# Backend Instructions

backend を変更する前に次を読む。

1. [Go/Gin Backend Guidelines](../.claude/rules/go-gin-backend-guidelines.md) — Go/Gin 公式由来の正本
2. [Backend Coding Rules](CODING_RULES.md) — backend 作業用サマリー
3. [Backend Application Invariants](../.claude/refs/backend-application-invariants.md) — clinic/owner/pet/staff 分離
4. 変更対象に最も近い `CLAUDE.md` と関連 ADR/OpenAPI

## Architecture

- Go/Gin公式は Handler → Service → Repository、Clean Architecture、layer-first/domain-first を規定しない。
- package は現在の folder tree ではなく、凝集性、利用者、依存方向、変更単位を根拠に設計する。
- interface は一般に利用側で最小に定義し、実装側は concrete type を返す。
- 新しい抽象化や package 分割は、実際の利用箇所が生じてから導入する。
- 既存 package 名は現状の実装説明であり、新規 code の固定 template ではない。

## Required safety

- boundary で input を検証し、authentication、authorization、resource ownership を分ける。
- request Context を DB/外部 API へ伝播し、Context を struct に保存しない。
- clinic-scoped data は全 read/write/delete path で認証済み clinic に制約する。
- error response と log に secret、credential、個人情報、内部詳細を出さない。
- OpenAPI、migration、security ADR との contract を維持する。

## Commands and verification

- host で `go` command を直接実行しない。Docker の scoped command を使う。
- full `go test ./...`、full lint、DB reset/migration apply は自動実行しない。
- 変更 package に限定した test、format、lint を優先する。
- docs-only change は runtime verification 不要。link、reference、format drift を確認する。

P1–P18 は廃止された project 固有 checklist であり、レビュー基準に使わない。レビューは [Go/Gin Backend Review](../.claude/refs/go-gin-backend-review.md) を使う。
