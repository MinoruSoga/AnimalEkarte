# Backend Instructions

backend を変更する前に次を読む。

1. [Go/Gin Backend Guidelines](../.claude/rules/go-gin-backend-guidelines.md) — Go/Gin 公式由来の正本
2. [Backend Coding Rules](CODING_RULES.md) — backend 作業用サマリー
3. [Backend Application Invariants](../.claude/refs/backend-application-invariants.md) — clinic/owner/pet/staff 分離
4. 変更対象に最も近い `CLAUDE.md` と関連 ADR/OpenAPI

業務能力、workflow、data ownership、自動化を変更する場合は、実装前に[Product Philosophy](../docs/product-philosophy.md)と該当するproduct specも読む。

## Architecture

- Go/Gin公式は Handler → Service → Repository、Clean Architecture、layer-first/domain-first を規定しない。
- package は現在の folder tree ではなく、凝集性、利用者、依存方向、変更単位を根拠に設計する。
- interface は一般に利用側で最小に定義し、実装側は concrete type を返す。
- 新しい抽象化や package 分割は、実際の利用箇所が生じてから導入する。
- 既存 package 名は現状の実装説明であり、新規 code の固定 template ではない。
- projectのtargetは、[ADR-006](../docs/architecture/adr/006-backend-domain-package-boundaries.md)で採用したdomain/capability-firstのmodular monolithである。route、use case、transaction、persistence、testをvertical sliceで変更する。
- BE9-2B以降、既存未移行codeの保守・安全修正と移動に必要なcompatibility変更を除き、新規production実装を`internal/handler`、`internal/service`、`internal/repository`へ追加しない。新規実装はADR-006のtarget domain packageへ置き、現在地は[`BE-refactor.md`](../BE-refactor.md)で確認する。
- domain内に`handler`、`service`、`repository` subpackageを機械的に作らない。実際のconsumer、依存方向、変更周期が分かれる場合だけ分離する。
- business factごとにsource of truthとwrite ownerを1つにする。`appointments`とそのlifecycleは`reservation`がwrite ownerであり、他domainはbusiness intentを表すconsumer-side interfaceまたは明示的orchestrationを通す。owner外へ任意field更新APIを公開しない。BE9-2E-0で収束済みの実装と回帰gateは`BE-refactor.md`で追跡する。
- cross-domain write、row/advisory lock、master・担当者・LINE顧客検証に必要なtransactionまたはrepositoryが欠ける場合はfail-closedとし、部分writeへ進まない。LIFFの明示staffは`is_active=true`かつ`reservation_visible=true`もwrite transaction内で検証する。通常カルテ削除は対象行lock下で見積依存を再確認してからdraftだけを原子的に許可し、見積Createと同じ親行lockで直列化する。fail-closedなclinical/financial監査はdependency欠落でもbusiness writeをrollbackする。
- compatibility facadeは薄いdelegate/type aliasに限定し、business ruleやwrite実装を複製しない。
- 自動化には停止、失敗通知、監査、手動fallback、idempotencyまたは明示的retry policyを設ける。

## Required safety

- boundary で input を検証し、authentication、authorization、resource ownership を分ける。
- request Context を DB/外部 API へ伝播し、Context を struct に保存しない。
- clinic-scoped data は全 read/write/delete path で認証済み clinic に制約する。
- appointmentに紐づく通常カルテは一般診療予約に限定し、1予約1 active record、予約日時由来のJST日付、依存失敗時のfail-closedを維持する。
- 1つのbusiness graphを構成する複数rowのwriteは同じtransactionで原子的に扱い、commit済みの成功を後段の再取得errorで失敗応答へ反転させない。
- error response と log に secret、credential、個人情報、内部詳細を出さない。
- OpenAPI、migration、security ADR との contract を維持する。

## Commands and verification

- host で `go` command を直接実行しない。Docker の scoped command を使う。
- full `go test ./...`、full lint、DB reset/migration apply は自動実行しない。
- 変更 package に限定した test、format、lint を優先する。
- docs-only change は runtime verification 不要。link、reference、format drift を確認する。

P1–P18 は廃止された project 固有 checklist であり、レビュー基準に使わない。レビューは [Go/Gin Backend Review](../.claude/refs/go-gin-backend-review.md) を使う。
