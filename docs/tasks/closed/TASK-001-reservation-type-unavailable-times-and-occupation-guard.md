# TASK-001: LINE予約 - 予約区分毎の予約不可時間 & 職種別出勤ガード

**作成日**: 2026-04-16
**ステータス**: Open
**依頼元**: ユーザー仕様追加依頼

---

## 概要

LINE予約において、予約区分（コース）毎に細かな受付制御を実現する2機能を追加する。
① 予約区分毎に予約不可時間帯を設定（曜日繰り返し or 特定日）
② 予約区分に職種を紐付け、その職種のスタッフが1人も出勤していない日は予約不可にする。

## 依頼内容（原文）

> LINE予約にて仕様追加です。
> 予約区分毎に予約不可時間を設定することは可能でしょうか？
> また、予約区分に職種を紐付けてその職種の職員が出勤していない日は予約不可にするような運用をできればと思います

## 仕様確認ログ

| # | 質問 | 回答 |
|---|------|------|
| 1 | 予約不可時間のパターン | C: 両方（曜日繰り返し + 特定日指定）に対応 |
| 2 | 予約区分と職種の対応数 | B: 1対多（M:N）— 複数職種を紐付け可能 |
| 3 | 「出勤している」の定義 | A: その職種のスタッフが1人でもいれば予約可 |
| 4 | 設定UIの配置場所 | デフォルト採用：カルテ側「予約区分マスタ」編集画面に追加 |
| 5 | 不可時間の最小単位 | デフォルト採用：30分単位 |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | DBスキーマ追加 + Go モデル + make codegen | DB/BE | BE-115 | - | [ ] |
| 2 | 管理API：予約不可時間 & 職種紐付け CRUD | BE | BE-116 | #1 | [ ] |
| 3 | LIFF予約可否ロジック拡張 | BE | BE-117 | #1 | [ ] |
| 4 | カルテ側：予約区分マスタ編集画面 UI 拡張 | FE | FE-252 | #2 | [ ] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: 管理画面の予約区分編集画面で「予約不可時間」を追加でき、「毎週月曜 12:00〜13:00」や「2026-05-10 09:00〜12:00」のように設定できる
- [ ] AC-2: 設定した不可時間帯はLINE予約の日時選択画面に反映され、当該時間が選択できない
- [ ] AC-3: 管理画面の予約区分編集画面で「対応職種」を1つ以上選択でき、複数職種を紐付けられる
- [ ] AC-4: 紐付けた職種のスタッフが1人も出勤しない日はLINE予約の日付選択画面でグレーアウト（選択不可）になる
- [ ] AC-5: 職種が紐付けられていない予約区分は従来通り動作する（職種ガード機能はオプション）
- [ ] AC-6: 予約不可時間も紐付け職種も未設定の予約区分は従来通り動作する

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| 予約不可時間の格納方式 | 専用テーブル `reservation_type_unavailable_times` | weekly/specific の2種別・複数設定を柔軟に管理。JSONB は型安全性が低く検索コスト高 | reservation_types への JSONB カラム追加 |
| 職種紐付けの格納方式 | 中間テーブル `reservation_type_occupations` | M:N。外部キー制約でデータ整合性を担保 | reservation_types への occupation_id 配列カラム |
| 出勤判定 | `shift_entries.shift_type NOT IN ('off', 'paid_leave')` | 既存 ShiftType の定義と一致。DB JOIN で完結 | カレンダー別途管理 |
| 不可時間の優先順位 | 特定日 > 曜日繰り返し | 特定日設定で例外対応可能にする | 両方単純加算 |

## 影響範囲

### DB
- 新テーブル: `reservation_type_unavailable_times` — 予約不可時間帯
- 新テーブル: `reservation_type_occupations` — 予約区分×職種 M:N 中間テーブル

### Backend
- `backend/migrations/001_init.sql` — 2テーブル追加
- `backend/internal/model/reservation_type.go` — 新モデル追加
- `backend/internal/handler/reservation_type_handler.go` — 不可時間・職種紐付け CRUD エンドポイント追加
- `backend/internal/service/reservation_type_service.go` — 同上サービス層
- `backend/internal/repository/reservation_type_repository.go` — 同上リポジトリ層
- `backend/internal/service/liff_service.go` — `GetAvailableDates`（職種ガード後処理）/ `GetAvailableTimes`（不可時間を DefaultBreaks に追加）に新ルール組み込み
- `backend/internal/service/timeslot_engine.go` — **変更なし**（既存の DefaultBreaks 機構を利用するため）

### Frontend
- `frontend/src/features/master/` — 予約区分編集画面に「予約不可時間」タブ + 「対応職種」セクション追加
- `frontend/src/types/generated/models.ts` — `make codegen` で自動更新

## 参照実装

- `features/owners/` — React 19 Action パターン（useActionState + SubmitButton）
- `backend/internal/handler/reservation_type_handler.go` — 既存の予約区分 CRUD ハンドラ
- `backend/internal/service/liff_service.go:GetAvailableDates()` — 既存の予約可否チェック

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| LIFF の `GetAvailableDates` / `GetAvailableTimes` のパフォーマンス低下 | 中 | shift_entries・reservation_type_occupations に適切なインデックスを追加（BE-115 で対応） |
| 職種設定なし予約区分の後方互換性 | 低 | 職種紐付けが0件の場合はガードをスキップする条件分岐を明示 |

## 未解決事項

- なし（仕様確認ログですべて解決済み）

## 実装順序

1. BE-115: DBスキーマ追加 + Go モデル + make codegen
2. BE-116: 管理 API（CRUD）実装
3. BE-117: LIFF 予約可否ロジック拡張
4. FE-252: カルテ側 UI 拡張

## 関連イシュー

- [BE-115](../../backend/issues/open/BE-115-reservation-type-unavailable-times-schema.md)
- [BE-116](../../backend/issues/open/BE-116-reservation-type-unavailable-times-api.md)
- [BE-117](../../backend/issues/open/BE-117-liff-availability-check-extension.md)
- [FE-252](../../frontend/issues/open/FE-252-reservation-type-master-ui-extension.md)
