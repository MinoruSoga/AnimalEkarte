# Phase 4: 管理画面フロントエンド

## TASK-RES-030: Feature scaffolding

**概要**: `features/reservations/` の基本構造とルーティング。

**対象ファイル（すべて新規）**:
```
frontend/src/features/reservations/
├── index.ts
├── api/
│   ├── get-reservation-settings.ts
│   ├── update-reservation-settings.ts
│   ├── get-reservation-courses.ts
│   ├── create-reservation-course.ts
│   ├── ... (他のAPI関数)
│   └── hooks.ts                  # React Query hooks
├── routes/
│   ├── ReservationSettings.tsx
│   ├── ReservationCourses.tsx
│   ├── ReservationStaffs.tsx
│   ├── StaffSchedule.tsx
│   ├── ReservationCalendar.tsx
│   └── ReservationPageEditor.tsx
└── hooks/
    └── use-reservation-form.ts
```

**ルーティング**: `frontend/src/app/router.tsx` に追加。サイドメニューに「LINE予約管理」セクション。

**完了条件**:
- [ ] ルーティング登録
- [ ] サイドメニュー表示
- [ ] RBAC: `reservation:read` / `reservation:write`

---

## TASK-RES-031: コース設定画面

**対象**: `ReservationCourses.tsx`
**UI仕様**: 仕様書セクション4 + セクション13.4

**機能**:
- 一覧テーブル（状態・Action・名称・略称・時間・is_internal・並び順）
- 新規登録モーダル
- 編集モーダル
- 削除（確認ダイアログ付き）
- 有効/休止トグル
- 並び順上下ボタン

**完了条件**:
- [ ] CRUD全操作
- [ ] `is_internal` フラグの表示・設定
- [ ] 並び順変更

---

## TASK-RES-032: スタッフ設定画面

**対象**: `ReservationStaffs.tsx`
**UI仕様**: 仕様書セクション5

**機能**:
- 一覧テーブル（状態・Action・肩書・名前・施設名・種別・非対応コース・並び順）
- 新規登録モーダル（非対応コースのマルチセレクト含む）
- 編集モーダル
- 削除、有効/休止トグル、並び順

**完了条件**:
- [ ] CRUD全操作
- [ ] `staff_type` (doctor/nurse/resource) 選択
- [ ] `facility_name` 入力
- [ ] 非対応コースのマルチセレクト

---

## TASK-RES-033: 基本設定画面

**対象**: `ReservationSettings.tsx`
**UI仕様**: 仕様書セクション6.3

**機能**:
- Running/Stopped トグル
- 休業曜日選択
- 営業時間設定（一括 or 曜日別）
- 休憩時間設定（複数）
- 各種制限値入力
- 追加入力フィールド定義（JSONBエディタ）
- LINE連携設定（channel ID, LIFF ID等）

**完了条件**:
- [ ] 全設定項目の表示・編集・保存
- [ ] 追加入力フィールドの動的追加・削除・並び替え

---

## TASK-RES-034: 個人設定画面

**対象**: `StaffSchedule.tsx`
**UI仕様**: 仕様書セクション7

**機能**:
- スタッフ選択ドロップダウン
- 月間スケジュール表示（青バー=勤務/グレーバー=休み）
- 日付クリック → 個人設定モーダル（休日チェック・受付時間・中断時間）
- 変更保存 / 設定削除 / キャンセル

**完了条件**:
- [ ] ガントチャート風表示（基本設定 + 個人上書きマージ）
- [ ] 個人設定の作成・更新・削除
- [ ] 中断時間の複数入力

---

## TASK-RES-035: 予約状況画面（★最も複雑な管理画面）

**対象**: `ReservationCalendar.tsx`
**UI仕様**: 仕様書セクション8

**機能**:
- **月表示**: カレンダーグリッド。セル内に予約サマリ。日付クリック→日表示。
- **日表示**: タイムテーブル（30分刻み × スタッフ列）。予約ブロック表示。
- **予約クリック** → キャンセル確認ダイアログ
- **空きセルクリック** → 手動予約入力モーダル
- **手動入力フォーム**: 日付・時間・スタッフ・コース・ユーザー名・電話番号
- 印刷対応

**完了条件**:
- [ ] 月表示（予約サマリ付き）
- [ ] 日表示（30分グリッド × スタッフ列）
- [ ] 月→日遷移
- [ ] 予約キャンセル
- [ ] 手動予約入力（バリデーションなし）
- [ ] 印刷対応

---

## TASK-RES-036: ページ編集画面

**対象**: `ReservationPageEditor.tsx`
**UI仕様**: 仕様書セクション6.2

**機能**: 5フィールドの編集フォーム（ヘッダ記載・予約通知・キャンセル通知・プライバシーポリシー）

**完了条件**:
- [ ] 全フィールドの編集・保存
