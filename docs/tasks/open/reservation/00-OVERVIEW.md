# LINE予約システム 実装タスク概要

> **仕様書**: `docs/line-reseavation.md` セクション15（実装確定仕様）
> **作成日**: 2026-04-08
> **ステータス**: 未着手

---

## 全体構成

```
Phase 1: DB・モデル基盤        ← 最初に着手（全ての土台）
Phase 2: バックエンド管理者API  ← 管理画面のデータ操作
Phase 3: バックエンド公開API    ← LIFF Appが呼ぶAPI
Phase 4: 管理画面フロントエンド ← 既存AnimalEkarteに統合
Phase 5: LIFF App             ← 新規プロジェクト（顧客向け予約UI）
Phase 6: LINE連携             ← LIFF認証・Push通知
Phase 7: 結合・デプロイ        ← 最後
```

## 依存関係

```
Phase 1 → Phase 2 → Phase 4
Phase 1 → Phase 3 → Phase 5
Phase 6 は Phase 3, 5 と並行可能
Phase 7 は全Phase完了後
```

## タスクサマリー

| Phase | タスク数 | 推定規模 | 詳細ファイル |
|---|---|---|---|
| Phase 1: DB・モデル | 3 | S | `01-PHASE1-DB.md` |
| Phase 2: 管理者API | 7 | M | `02-PHASE2-ADMIN-API.md` |
| Phase 3: 公開API | 6 | L（時間枠エンジンが核心） | `03-PHASE3-LIFF-API.md` |
| Phase 4: 管理画面FE | 7 | L（予約カレンダーが核心） | `04-PHASE4-ADMIN-FE.md` |
| Phase 5: LIFF App | 10 | L | `05-PHASE5-LIFF-APP.md` |
| Phase 6: LINE連携 | 2 | S | `06-PHASE6-LINE.md` |
| Phase 7: 結合・デプロイ | 4 | M | `07-PHASE7-DEPLOY.md` |
| **合計** | **39タスク** | |  |

## 全体方針

- **既存テーブル統合**: 予約データは既存 `reservation_appointments` に統合。スタッフは `staffs`、メニューは `service_types` を拡張。新規テーブルは4つのみ
- **カルテ自動連携**: LINE予約 → `reservation_appointments` INSERT → `medical_records.reservation_appointment_id` で参照可能
- **タイムゾーン**: 全日時処理はJST固定（Asia/Tokyo）
- **祝日**: `holiday_jp-go` ライブラリで判定（外部API依存なし）
- **楽観ロック**: SELECT FOR UPDATE（PostgreSQL行ロック）
- **データ移行**: v1は新規スタート。既存「予約 on ライン」のデータは移行しない（並行運用→切替）
- **マルチテナント**: 既存AnimalEkarteのclinic_id認可ミドルウェアを再利用
- **エラーレスポンス**: 既存プロジェクト規約（`RespondError` + `apperrors`）に準拠
