# Phase 5: LIFF App（新規プロジェクト）

> **状態**: ✅ 全タスク完了（2026-04-08〜09）
>
> **実装場所**: `frontend/line-reserve/` ディレクトリ（既存 frontend とは独立した Vite プロジェクト）
> ※ 旧 `liff-app/` から移動済み（`4c46092a chore: liff-app/ 削除`）
>
> **Phase 8 追加機能**（2026-04-10）:
> - マルチペット選択（チェックボックス + 新規追加）
> - 紐付け済み飼主からの顧客情報自動入力
> - pet_name/pet_type → pets[] 配列に変更

## TASK-RES-040: プロジェクトセットアップ ✅

**実装済み構成**:
```
frontend/line-reserve/
├── src/
│   ├── main.tsx
│   ├── App.tsx
│   ├── pages/
│   │   ├── TopPage.tsx
│   │   ├── CustomerInfoPage.tsx   # STEP 1
│   │   ├── CourseSelectPage.tsx   # STEP 2
│   │   ├── StaffSelectPage.tsx    # STEP 3
│   │   ├── DateSelectPage.tsx     # STEP 4
│   │   ├── TimeSelectPage.tsx     # STEP 5
│   │   ├── RequestPage.tsx        # STEP 6
│   │   ├── ConfirmPage.tsx        # STEP 7
│   │   ├── CompletePage.tsx       # STEP 8
│   │   ├── MyReservationsPage.tsx
│   │   ├── ErrorPage.tsx
│   │   └── MaintenancePage.tsx
│   ├── components/
│   │   ├── ProgressDots.tsx
│   │   ├── ListItem.tsx
│   │   ├── Calendar.tsx
│   │   ├── PrimaryButton.tsx
│   │   └── BackButton.tsx
│   ├── hooks/
│   │   ├── use-liff.ts
│   │   └── use-reservation-flow.ts
│   ├── api/
│   │   └── liff-api.ts
│   └── lib/
│       └── liff-config.ts
├── index.html
├── vite.config.ts
├── tsconfig.json
├── package.json
└── Dockerfile
```

**完了条件**:
- [x] `docker compose up` で起動
- [x] LIFF SDK初期化（成功/失敗の両パス）
- [x] 未ログイン時のLINEログインリダイレクト
- [x] 友だち未追加時の案内画面
- [x] 開発時モック（LIFF_MOCK=true）
- [x] ルーティング設定

---

## TASK-RES-041: STEP 1 — お客様情報入力 ✅

**実装済みファイル**: `liff-app/src/pages/CustomerInfoPage.tsx`

**完了条件**:
- [x] 5フィールドの入力フォーム
- [x] 2回目以降: 前回入力値プリフィル（GET /api/liff/profile）
- [x] バリデーション（必須チェック）
- [x] プログレスドット表示

---

## TASK-RES-042: STEP 2 — コース選択 ✅

**実装済みファイル**: `liff-app/src/pages/CourseSelectPage.tsx`

**完了条件**:
- [x] コースリスト表示
- [x] コース名 + コメント表示
- [x] 曜日オプション表示

---

## TASK-RES-043: STEP 3 — スタッフ選択 ✅

**実装済みファイル**: `liff-app/src/pages/StaffSelectPage.tsx`

**完了条件**:
- [x] 選択コースの非対応スタッフを除外
- [x] 「選択せずに予約」オプション（設定依存）
- [x] スタッフリスト表示

---

## TASK-RES-044: STEP 4 — 日付選択 ✅

**実装済みファイル**: `liff-app/src/pages/DateSelectPage.tsx`

**完了条件**:
- [x] 月間カレンダー表示
- [x] 選択不可日グレーアウト
- [x] 前月/翌月ナビゲーション
- [x] 曜日色分け（日=赤、土=青）

---

## TASK-RES-045: STEP 5 — 時間選択 ✅

**実装済みファイル**: `liff-app/src/pages/TimeSelectPage.tsx`

**完了条件**:
- [x] 空き時間枠リスト表示
- [x] 「空きのある時間のみ表示」注記

---

## TASK-RES-046: STEP 6 — 要望入力 ✅

**実装済みファイル**: `liff-app/src/pages/RequestPage.tsx`

**完了条件**:
- [x] テキストエリア（プレースホルダー付き）
- [x] 任意入力（スキップ可能）

---

## TASK-RES-047: STEP 7 — 確認画面 ✅

**実装済みファイル**: `liff-app/src/pages/ConfirmPage.tsx`

**完了条件**:
- [x] 全入力内容のテーブル表示
- [x] 「まだ予約は完了していません」警告
- [x] 確認/戻るボタン

---

## TASK-RES-048: STEP 8 — 完了画面 + エラー画面 ✅

**実装済みファイル**:
- `liff-app/src/pages/CompletePage.tsx`
- `liff-app/src/pages/ErrorPage.tsx`
- `liff-app/src/pages/MaintenancePage.tsx`

**完了条件**:
- [x] 予約完了メッセージ
- [x] 同日/同月予約制限エラー画面
- [x] メンテナンス中画面（Stopped時）

---

## TASK-RES-049: トップページ + 予約確認・キャンセル ✅

**実装済みファイル**:
- `liff-app/src/pages/TopPage.tsx`
- `liff-app/src/pages/MyReservationsPage.tsx`

**完了条件**:
- [x] トップページ（アコーディオン3セクション + CTA2ボタン）
- [x] マイ予約一覧
- [x] 予約キャンセル（確認ダイアログ付き）
