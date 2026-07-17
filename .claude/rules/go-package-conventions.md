---
description: backend の Go パッケージ配置・命名のコード規約（恒久正本）。層優先×ドメインサブパッケージ、stutter 禁止、consumer 側 interface、lint 1階層制約。
alwaysApply: false
globs: ["backend/**/*.go"]
---

# Go パッケージ配置・命名規約（正本）

> **正本宣言**: 本ファイル（`.claude/rules/go-package-conventions.md`）がコード規約の恒久正本。`.agents/rules/` 側は sync-agents-skills.sh による生成ミラー（直接編集禁止）。移行**計画**（BE8 手順・地雷・ドメイン一覧）はリポジトリ直下 `BE-refactor.md` — 対応完了後に削除される使い捨て文書であり、規約は本ファイルに残る。
> 決定日: 2026-07-17（要件責任者: 曽我）

## 目標構成（層優先 × ドメインサブパッケージ）

```
backend/internal/
  handler/                 # 現状フラット維持（分割判断は BE-refactor.md BE8-7）
  service/
    <domain>/              # 例: service/reservation/ — 新規ドメインはここに作る
  repository/
    repohelpers/           # 共有 clinic-scope / DBOrTx ヘルパ
    <domain>/              # 例: repository/paymentmethod/（先行分割済みドメイン多数）
  model/                   # 単一パッケージ維持（FK 相互参照のため分割しない — 決定事項）
  middleware/ infra/ config/ ...  # 健全・変更不要
```

不採用と決定済みの構成（再提案しない）: ドメイン優先の全面転換（`internal/<domain>/{handler,service,repository}`）／`pkg/` 新設。理由は BE-refactor.md §2-3（削除後は git 履歴）。

## 配置規則

| 層 | 新規コードの配置 | 備考 |
|---|---|---|
| handler | `internal/handler/`（フラット直下） | 分割は保留中。ルート登録 = `handler.go` / `master_routes.go` |
| service | `internal/service/<domain>/` **サブパッケージ** | フラット直下への新規追加は禁止 |
| repository | `internal/repository/<domain>/` **サブパッケージ** | 先例: `paymentmethod/`。共有ヘルパは `repohelpers` |
| model | `internal/model/`（単一パッケージ） | 分割しない — 決定事項 |

- 既存フラットファイルは実装変更で触るときにサブパッケージへ移す（strangler）。**一斉移動・移動と同時の公開型リネームは禁止**（diff 爆発防止。リネームは別コミット）。

## 命名規則

- パッケージ名 = **単数形・全小文字・アンダースコアなし**のドメイン名（`trimmingcoursetype` 形式）。`util` / `common` / `helpers` 単独名は禁止（`repohelpers` は既存例外）。
- **stutter 禁止**: 識別子でパッケージ名を繰り返さない — `reservation.NewRepository` であって `reservation.NewReservationRepository` ではない（新規コードに適用。既存型は移動時にリネームしない）。

## ドメイン間参照（import cycle 回避）

- service ↔ service のドメイン間参照は **consumer 側で小文字ローカル interface** を定義して受ける（in-repo 先例: `reservation_service.go` の `reservationTypeFinder`）。
- cycle が出たら移動を巻き戻すのではなく interface 抽出で解決する。

## lint 制約（違反するとサイレント緑になる）

- **サブパッケージ内にさらにディレクトリを掘らない**。repository の走査 lint 群（preload_clinic_scope / audit_tx_inventory / dbortx_inventory）は `go:embed *.go */*.go` の **1 階層走査**であり、2 階層目のファイルは臨床安全 lint から不可視になる。
- service 側を走査する同種 lint は現存しない（安全網は repository のみ — 分割時は BE-refactor.md BE8-5 の手順に従う）。

## 出典

- [Go 公式 module layout](https://go.dev/doc/modules/layout) — server logic は `internal/` 配下のドメイン名パッケージ
- [Google Go Style: Best Practices](https://google.github.io/styleguide/go/best-practices) — util 禁止・stutter 禁止・分割基準
- 関連: `.claude/refs/go-language.md` §8（Go 作業時ロードの要約）／各層 `CLAUDE.md`（編集時自動ロードの要約）
