# LINE予約システム 実装タスク概要

> **仕様書**: `docs/line-reseavation.md` セクション15（実装確定仕様）
> **作成日**: 2026-04-08
> **ステータス**: ✅ Phase 1〜6 実装完了 / Phase 7 結合テスト・デプロイ残

---

## 全体構成

```
Phase 1: DB・モデル基盤        ✅ 完了
Phase 2: バックエンド管理者API  ✅ 完了
Phase 3: バックエンド公開API    ✅ 完了
Phase 4: 管理画面フロントエンド ✅ 完了 → Phase 8 で電カル統合
Phase 5: LIFF App             ✅ 完了
Phase 6: LINE連携             ✅ 完了
Phase 8: 電カル統合            ✅ 完了（LINE管理ページ→マスタ設定統合）
Phase 7: 結合・デプロイ        🔲 結合テスト・本番デプロイ残
```

## 依存関係

```
Phase 1 → Phase 2 → Phase 4
Phase 1 → Phase 3 → Phase 5
Phase 6 は Phase 3, 5 と並行可能
Phase 7 は全Phase完了後
```

## タスクサマリー

| Phase | タスク数 | 推定規模 | 詳細ファイル | 状態 |
|---|---|---|---|---|
| Phase 1: DB・モデル | 3 | S | `01-PHASE1-DB.md` | ✅ 完了 |
| Phase 2: 管理者API | 7 | M | `02-PHASE2-ADMIN-API.md` | ✅ 完了 |
| Phase 3: 公開API | 6 | L（時間枠エンジンが核心） | `03-PHASE3-LIFF-API.md` | ✅ 完了 |
| Phase 4: 管理画面FE | 7 | L（予約カレンダーが核心） | `04-PHASE4-ADMIN-FE.md` | ✅ 完了 → Phase 8 で統合 |
| Phase 5: LIFF App | 10 | L | `05-PHASE5-LIFF-APP.md` | ✅ 完了 |
| Phase 6: LINE連携 | 2 | S | `06-PHASE6-LINE.md` | ✅ 完了 |
| Phase 8: 電カル統合 | - | L | (本ドキュメント参照) | ✅ 完了 |
| Phase 7: 結合・デプロイ | 4 | M | `07-PHASE7-DEPLOY.md` | 🔲 TASK-RES-071/072 残 |
| **合計** | **39タスク + Phase 8** | | | |

## 全体方針

- **既存テーブル統合**: 予約データは既存 `reservation_appointments` に統合。スタッフは `staffs`、メニューは `service_types` を拡張。新規テーブルは4つのみ
- **カルテ自動連携**: LINE予約 → `reservation_appointments` INSERT → `medical_records.reservation_appointment_id` で参照可能
- **タイムゾーン**: 全日時処理はJST固定（Asia/Tokyo）
- **祝日**: `holiday_jp-go` ライブラリで判定（外部API依存なし）
- **楽観ロック**: SELECT FOR UPDATE（PostgreSQL行ロック）
- **データ移行**: v1は新規スタート。既存「予約 on ライン」のデータは移行しない（並行運用→切替）
- **マルチテナント**: 既存AnimalEkarteのclinic_id認可ミドルウェアを再利用
- **エラーレスポンス**: 既存プロジェクト規約（`RespondError` + `apperrors`）に準拠
