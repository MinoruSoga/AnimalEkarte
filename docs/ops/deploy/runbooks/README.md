# runbooks/ — 個別実行手順書

> **目的**: 事故を防ぐための、機械的に実行できるオペレーション手順書の索引を提供する。
> **読者**: DevOps・AI エージェント。
> **タイミング**: 該当オペレーションの実施直前（手順の即興を禁止し、必ず本書を開いてから実行する）。

## 索引

| ランブック | 内容 | いつ使うか |
|:---|:---|:---|
| [STG_PRE_DEPLOY_READINESS_CHECK.md](STG_PRE_DEPLOY_READINESS_CHECK.md) | 本番反映前の最終検証（checksum・fresh-DB 適用・CI 波及・db_reset 要否） | main → staging PR 前・seed/migration を含むデプロイ前 |
| [SCHEDULER_OPERATIONS.md](SCHEDULER_OPERATIONS.md) | Durable schedulerのstatus・pause/resume・missing-slot catch-up・障害復旧 | 定期jobの監視・停止・手動復旧・release qualification時 |
| [SCHEDULER_BATCHES.md](SCHEDULER_BATCHES.md) | バッチ別仕様と本番 go-live 有効化判断 | go-live 前の batch 有効化判断時 |
| [BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md](BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md) | Cloudflare / PlanetScale 資格情報ローテーションと、退役済み AWS 手順の実行禁止境界 | 資格情報ローテーションの計画時、または過去の AWS 運用を調査する時 |
| [SEC_SECRETS_5_GITLEAKS_HISTORY_INVENTORY.md](SEC_SECRETS_5_GITLEAKS_HISTORY_INVENTORY.md) | git 履歴上のシークレット露出棚卸しとローテーション手順 | シークレットローテーション作業時 |
| [BREAK-HOURS-SHAPE-AUDIT.md](BREAK-HOURS-SHAPE-AUDIT.md) | `line_reservation_settings.break_hours` 形状監査（本番デプロイ前ゲート・使い切り） | 本番デプロイ前の監査 SQL 実行時 |
| [EMR-321_HOTEL_RESERVATION.md](EMR-321_HOTEL_RESERVATION.md) | ホテル予約区分・ケージ選択の医院別手動確認 | EMR-321 対象医院の手動確認時 |
| [LSTEP_WRITE_API_PAUSE.md](LSTEP_WRITE_API_PAUSE.md) | Lステップ Write API 一時停止（deploy gate + clinic gate の二重 gate） | Write 再有効化の判断時 |
| [LINE_RESERVATION_CLOSED_WEEKDAYS_RECOVERY.md](LINE_RESERVATION_CLOSED_WEEKDAYS_RECOVERY.md) | `line_reservation_settings.closed_weekdays` 破損時の認証付き API による休診曜日復旧 | closed_weekdays 破損で管理 UI 保存不能・LIFF 予約 fail-closed 発生時 |

## 運用ルール

- ランブックは「実施した通りに更新する」。手順と実態が乖離したまま放置しない。
- 使い切り（特定 PR・特定時点限り）のチェックリストはランブック化せず、完了後に削除して git 履歴に残す。
