# 優先修復キュー

coordinatorは本文を修復せず、後続ユニットへ渡す。Severity は S3=利用判断を誤らせる説明、S2=運用/性能/証拠の誤読、S1=案内・表現改善（S4は緊急の漏えい等、本調査では該当なし）。Impact は4=現場/安全に広く影響、3=複数経路、2=限定読者、1=局所。Score = Severity × Impact の降順、同点はID順。これは文書修復の優先度であり、製品障害の深刻度判定ではない。

`queued` は子の修復待ち、`external-scope` は docs 外の別承認依存、`index-resolved` はこのcoordinatorで入口だけ修正済み。全件の本文・runtime修復完了を意味しない。action は update / delete / merge-link / rewrite-structure / adr-correction のいずれか。delete と adr-correction は今回確定対象なし。

| ID | Severity | Impact | Score | 子 / 状態 | 推奨処置 | 根拠と実施すべき修復 |
|---|---|---|---|---|---|---|
| RQ-001 | S3 | 4 | 12 | spec / resolved | merge-link | `docs/spec/screens/34-lstep-delivery-monitor.md:48` は死亡後に配信対象から外れると説明するが、`docs/spec/line/lstep-integration.md:72` はbest-effortと既知のgapを明記。`backend/internal/lstep/lstep_delivery_trigger_state.go:11` の最終除外確認にはpet status読取りがない。画面文書を既存の安全制約ownerへリンクし、保証範囲を揃える。コードの安全欠陥が解消したとは書かない |
| RQ-002 | S3 | 4 | 12 | delivery+manual / resolved | update | `docs/delivery/OPERATION_MANUAL.md:9` が委ねるmanual内で、`frontend/src/features/manual/content/workflows/01-new-owner-first-visit.md:111` は確定で会計自動作成、`:112` は会計待ち自動遷移と説明。他方 `frontend/src/features/manual/content/screens/05-accounting.md:68` は会計作成時pull、`docs/spec/screens/99-medical-record-flow.md:41` は会計別作成。まず参照先の説明矛盾として登録。2026-09-07: frontend manual（workflows/screens）を実装整合。カルテ確定は billing 非作成、会計作成時 pull。退院は create_accounting 選択時のみ同時作成。D-254 FAQ/スクショ突合は別依存 |
| RQ-003 | S2 | 3 | 6 | spec / resolved | update | `docs/spec/screens/34-lstep-delivery-monitor.md:60` の無条件no-N+1と `docs/spec/line/lstep-integration.md:95` のdegraded fallback説明が相違。`backend/internal/lstep/lstep_delivery_trigger_batch.go:106`、`:143`、`:153` にbulk失敗後の個別読取りfallback。通常経路のbulk要件と現存する劣化経路を分け、既存要件を撤回する仕様変更にしない |
| RQ-004 | S2 | 3 | 6 | ops / resolved | merge-link | `docs/ops/testing/UAT-DOMAIN-STATUS.md:54`、`:73`、`:98`、`:140`、`:160`、`:204` の9リンク先がローカルに不在。`:27` はgitignored原証跡の非配布を既に明記。リンク表示を「保管者から取得する証跡識別子」と区別するか、承認済み保管先への参照へ変更する。原証跡の不在をUAT未実施/完了に読み替えない。全対象は `EVIDENCE/INVENTORY-SCAN.json` |
| RQ-005 | S2 | 3 | 6 | ops / index-resolved | merge-link | 開始時の `docs/ops/deploy/README.md:46` は実装契約だけを案内し、`docs/ops/deploy/LAB_DEVICE_AGENT_MACOS.md:11`、`:15`、`:18` の配布前提への入口がない。本ユニットでREADMEに同文書への1リンクを追加。署名・token供給・STG/PROD配布の解決や検証はしていない |
| RQ-006 | S2 | 2 | 4 | ops / resolved | rewrite-structure | `docs/ops/deploy/STG-CONTINUOUS-OPERATIONS.md:235` はデプロイ直後ログの保存先を一時領域とする。`docs/ops/CLAUDE.md:25` はActions run/変更チケット/承認済み運用記録への証跡を要求。一時採取と永続的な正式記録への保存を分け、既存承認済み記録先へリンクする。新しいwikiサービスを発明しない |
| RQ-007 | S1 | 3 | 3 | spec / resolved | merge-link | `docs/spec/screens/README.md:24` の検査機器は独立仕様へのリンクがなく、`docs/spec/screens/settings/README.md:23` 以降の表にも検査機器マスタの案内がない。実在routeは `frontend/src/app/routes/clinical-care-routes.tsx:221`、`frontend/src/app/routes/settings-routes.tsx:462`。既存ADR/実装契約を案内する最小リンクを優先し、重複本文を増やさない |
| RQ-008 | S1 | 2 | 2 | ops / resolved | update | `docs/ops/testing/UAT-ENV-SETUP.md:15` の `scripts/check-uat-env.sh` は文書相対なら実在する（`docs/ops/testing/scripts/check-uat-env.sh:1`）。repo-root相対と誤解されないリンク/表記へ整える。helperが不存在という指摘は棄却。advisoryの境界を維持し、helperを実行しない |
| RQ-009 | S1 | 2 | 2 | work / resolved | update | `docs/work/decisions/fable-po-recommendation.md:11` のNeeds Human表現を、`:5` の2026-08-20照合と `:6` の外部未確認に明示的に結び付ける。「当時」の意味を強める表現修正であり、Linear現在状態の誤りを確認したものではない |
| RQ-010 | S1 | 2 | 2 | product-philosophy / resolved | update | `docs/product-philosophy.md:41` の「誤操作を防がない」という説明例が、`:162` の補助的な明示同意の位置付けより強く読める可能性。既存の安全原則に沿って例の限定を検討する。対話確認を一律禁止する新要件や、臨床安全の削減を発明しない |
| RQ-011 | S1 | 2 | 2 | architecture / resolved | merge-link | `docs/architecture/overview.md:111` はwrite ownershipをADR-006へ案内し、`docs/architecture/model-write-owner-catalog.md:3`、`:14` は型別owner表を担う。役割を明示する必要が残ればoverview/indexに型別catalogへのリンクを補う。ADRの採択理由・BE9履歴を現行表へ書き換えない |

## 保留・棄却した候補

- ADR-006の旧HEAD（`docs/architecture/adr/006-backend-domain-package-boundaries.md:202`）は `:194` の日付付き静的照合の出典。SHAが現在HEADと異なるだけでは誤りでない。adr-correctionは不要と判定した。
- `docs/spec/screens/README.md:93` 等の2026-09-06基準SHAも履歴根拠として保持する。古い日付だけを理由に最新SHAへ置換しない。
- `docs/spec/ui-design-compliance.md:3`、`:42`、`:44` はruntime日付と静的照合を既に区別し、未実行ルートをpendingとする。既知のE2E gapは別unitの依存であり、本調査で新しい文書誤りと認定しない。
- `docs/work/linear-f1-f6-mapping.md:5`、`:19`、`:51` はUNKNOWN・実装履歴・USER操作を既に分離している。章分割を必須とする指摘は棄却。外部現在状態を確認したとはしない。
- STGのcustom domainとworkers.devの現在稼働は未照会。`docs/ops/deploy/README.md:65` は切り分け手順を既に案内し、`backend/wrangler.jsonc:29` と `.github/workflows/backend-deploy.yml:22` はrepo設定/コメントに過ぎない。provider状態の断定修正はしない。
- `docs/work/README.md` の削除zoneに対するadr-correction提案は棄却（ADRではなく、既にCorpVaultとrepoを区別）。新たな削除確定対象はない。将来のdelete提案には理由・後継リンク・再作成条件が必須。

## 子への引渡し境界

各カテゴリは [CHILD-GOALS](CHILD-GOALS/) のallowlistだけを編集する。RQ-002のようなdocs外依存は別承認を必要とし、保守coordinatorやdelivery子が自動で吸収しない。実行状態は [LEDGER.md](LEDGER.md)、fact ownerの選択は [ROLE-MAP.md](ROLE-MAP.md) を参照する。
