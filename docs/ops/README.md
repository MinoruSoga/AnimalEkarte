# ops/ — 運用系（デプロイ・CI・テスト・インフラ）

> **目的**: システムを「どう動かすか」の運用系ドキュメントの索引を提供する。
> **読者**: 全開発者・DevOps・AI エージェント。
> **タイミング**: デプロイ・CI 変更・テスト実施・インフラ作業の前。

編集時のルールは [CLAUDE.md](CLAUDE.md) を参照。技術設計は [../architecture/](../architecture/README.md)、仕様は [../spec/](../spec/README.md)。

## 索引

```text
docs/ops/
├── README.md                  … 本索引
├── CLAUDE.md                  … 編集時ルール（インフラ SSOT・運用原則）
├── deploy/                    … デプロイハブ・runbooks/（個別手順）・rehearsals/（検証）・seeds/（DB/CSV）・stg/・accounts/・lab-devices/
├── testing/                   … テストアーキテクチャ・受入 scenarios・E2E・環境セットアップ
├── infra/                     … Cloudflare 基盤の現行契約と構成図・AWS 廃止履歴への正本ポインタ
├── ci-policy.md               … CI ワークフローの決定事項記録
├── agent-harness.md           … AI 設定スコープ・作業完了条件・検証と失敗の還元
├── coverage-policy.md         … テストカバレッジ ratchet 方式の運用ポリシー
├── backlog-spreadsheet.md     … Q&A バックログスプレッドシートの運用ルール
└── plane-ticket-conventions.md … Plane チケット起票規約（タイトル・ラベル・本文テンプレート）
```

| ドキュメント | 内容 | いつ読むか |
|:---|:---|:---|
| [deploy/](deploy/README.md) | デプロイハブ（環境 URL・Cloudflare デプロイ・障害時判断） | デプロイ・リリース作業前 |
| [deploy/runbooks/](deploy/runbooks/README.md) | 個別実行手順書（リリース前検証・scheduler運用・資格情報ローテーション・シークレット棚卸し） | 該当オペレーション実施時 |
| [deploy/LAB_DEVICE_CONNECTIVITY.md](deploy/lab-devices/LAB_DEVICE_CONNECTIVITY.md) | 検査機器の実装契約。操作手順は old_db `docs/lab-go/hospital-field-pack/手元テスト手順.md` | 院内で機器を新カルテへ取り込むとき |
| [testing/](testing/README.md) | テストアーキテクチャ・受入 scenarios（項目単位）・E2E・環境セットアップ | テスト実施・品質検証時 |
| [ci-policy.md](ci-policy.md) | CI ワークフローの決定事項記録（Actions バージョンピン方針等） | .github/workflows/ 変更前 |
| [agent-harness.md](agent-harness.md) | AI設定のスコープ、作業完了条件、検証と失敗の還元 | AI設定変更・自律タスク実行前 |
| [coverage-policy.md](coverage-policy.md) | テストカバレッジ ratchet 方式の運用ポリシー | カバレッジゲート調整時 |
| [backlog-spreadsheet.md](backlog-spreadsheet.md) | Q&A バックログスプレッドシートの運用ルール | クライアント Q&A シート操作前 |
| [plane-ticket-conventions.md](plane-ticket-conventions.md) | Plane チケット起票規約（タイトル・ラベル・本文テンプレート・UUID 対応表） | Plane work item 作成・更新前 |
| [infra/architecture.md](infra/architecture.md) | インフラ構成図・ネットワーク・セキュリティ設計（`../architecture/overview.md` の domain/capability-first 構成とは別物） | インフラ構成の調査・変更前 |
| [infra/](infra/README.md) | Cloudflare 基盤の現行リポジトリ契約と、AWS 廃止履歴への正本ポインタ | インフラ変更・履歴調査前 |

## GitHub Actions workflow → 文書オーナー

実装された操作・trigger・条件の正本は各 `.github/workflows/*.yml` 自体であり、右列はその workflow の evidence/policy オーナー（runner 契約の詳細は YAML 側に残り、リンク先文書が runner 詳細を複製するわけではない）。下表は既存ポリシー・runbook への navigation であり、実行・稼働・合格の主張ではない（2026-10-11 棚卸し・静的対応のみ）。

| workflow | 文書オーナー（policy / evidence） |
|:---|:---|
| [actionlint.yml](../../.github/workflows/actionlint.yml) | [ci-policy.md](ci-policy.md) |
| [backend-deploy.yml](../../.github/workflows/backend-deploy.yml) | [deploy/CI-CD-PIPELINE.md](deploy/CI-CD-PIPELINE.md) |
| [chromatic.yml](../../.github/workflows/chromatic.yml) | [testing/TEST_ARCHITECTURE.md](testing/TEST_ARCHITECTURE.md) |
| [ci.yml](../../.github/workflows/ci.yml) | [ci-policy.md](ci-policy.md) |
| [e2e.yml](../../.github/workflows/e2e.yml) | [testing/E2E_TESTING_GUIDE.md](testing/E2E_TESTING_GUIDE.md) |
| [frontend-deploy.yml](../../.github/workflows/frontend-deploy.yml) | [deploy/CI-CD-PIPELINE.md](deploy/CI-CD-PIPELINE.md) |
| [performance-tests.yml](../../.github/workflows/performance-tests.yml) | [testing/INTEGRATION_TEST_PLAN.md](testing/INTEGRATION_TEST_PLAN.md) |
| [plane-merge-state.yml](../../.github/workflows/plane-merge-state.yml) | [plane-ticket-conventions.md](plane-ticket-conventions.md) |
| [security-scan.yml](../../.github/workflows/security-scan.yml) | [ci-policy.md](ci-policy.md) |
| [stg-migrate-direct.yml](../../.github/workflows/stg-migrate-direct.yml) | [deploy/STG_PLANETSCALE_SEED_RUNBOOK.md](deploy/stg/STG_PLANETSCALE_SEED_RUNBOOK.md) |
| [stg-smoke.yml](../../.github/workflows/stg-smoke.yml) | [deploy/CRUD-SMOKE-TEST.md](deploy/rehearsals/CRUD-SMOKE-TEST.md) |
| [storybook.yml](../../.github/workflows/storybook.yml) | [testing/TEST_ARCHITECTURE.md](testing/TEST_ARCHITECTURE.md) |
| [worker-secret-sync.yml](../../.github/workflows/worker-secret-sync.yml) | [infra/cloudflare-account-separation.md](infra/cloudflare-account-separation.md) |

補足（静的確認 2026-10-11、実装詳細は YAML 参照）:

- `chromatic.yml` — Storybook・ビジュアルテストの publish 用。token 未設定時は SKIP となり PASS 扱いではない（本 docs では設定有無は UNKNOWN）。
- `storybook.yml` — component catalog を gh-pages へ build/publish するもので、application deploy でも UAT 証明でもない。

## AI エージェント向け注記

- タスク台帳の実行 SoT は **Plane**（workspace `baritechllc`）。hub・issue の現在状態は外部情報のため、操作前に Plane / GitHub で人手確認する。root [`todo.md`](../../todo.md) は repo と強く結び付く未完了作業の入口であり、[`todo.md#human-lane`](../../todo.md#human-lane) は入口ポインタ。運用で発見した課題は Plane に起票する。旧 `q&a.html` / `STATUS.md` / `PO-todo.md` は削除済。（2026-09-29 訂正: 実行 SoT を Linear としていた → Linear は 2026-09-16 に閉鎖済み。現行 SoT は Plane で、移行記録は [../work/plane/plane-md-migration-20260923-receipt.md](../work/plane/plane-md-migration-20260923-receipt.md)。旧 BRT-* / LINMIG-* ID と linear.app リンクは履歴参照のみ）
- **外部 mapping（操作前に GitHub / Plane で検証必須。以下は履歴 pointer）:** #261 / BRT-51 の判定先として GitHub [#261](https://github.com/MinoruSoga/AnimalEkarte/issues/261) 本文（A/B/C）。`q&a.html` は repo に無い。B 薬量は BRT-39（旧 Linear ID、履歴参照のみ）。C 健診 seed は BRT-40（旧 Linear ID、履歴参照のみ）。検査 / #249 はスコープ外。
- migration/seed に触れる作業は `migration-seed-safety` スキル、リリース前チェックは `stg-release-readiness` スキルを先に読むこと。

Spec KitのCLI・Codex連携・正本参照・開始手順は [spec-kit.md](spec-kit.md)を参照する。
