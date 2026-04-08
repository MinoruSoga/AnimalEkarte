# Phase 5: LIFF App（新規プロジェクト）

## TASK-RES-040: プロジェクトセットアップ

**概要**: `liff-app/` に Vite + React 19 + TypeScript + Tailwind プロジェクト作成。

**構成**:
```
liff-app/
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
│   ├── types/
│   │   └── models.ts
│   └── lib/
│       └── liff-config.ts
├── index.html
├── vite.config.ts
├── tailwind.config.ts
├── tsconfig.json
├── package.json
└── Dockerfile
```

**LIFF SDK初期化とエラーハンドリング**:
```
1. liff.init({ liffId }) を呼び出し
2. 成功 → liff.isLoggedIn() チェック
   - 未ログイン → liff.login() でLINEログインへリダイレクト
3. ログイン済み → liff.getProfile() でユーザー情報取得
4. liff.getFriendship() で友だち追加チェック
   - 未友だち → 友だち追加を促す画面を表示

エラー時のフォールバック:
- liff.init() 失敗 → 「LINEアプリからアクセスしてください」エラー画面表示
- liff.getProfile() 失敗 → お名前欄を空にしてフォーム表示（手入力）
- ネットワークエラー → リトライボタン付きエラー画面
- 開発時: LIFF_MOCK=true 環境変数でモックプロフィルを返す
```

**タイムゾーン**: 全日時処理はJST固定。サーバー・クライアント共にAsia/Tokyoを前提とする。

**完了条件**:
- [ ] `docker compose up` で起動
- [ ] LIFF SDK初期化（成功/失敗の両パス）
- [ ] 未ログイン時のLINEログインリダイレクト
- [ ] 友だち未追加時の案内画面
- [ ] 開発時モック（LIFF_MOCK=true）
- [ ] ルーティング設定

---

## TASK-RES-041: STEP 1 — お客様情報入力

**UI仕様**: 仕様書セクション15.2 STEP 1

5フィールド: お名前, 電話番号, 飼い主名, ペットの名前と種類, 診察内容

**完了条件**:
- [ ] 5フィールドの入力フォーム
- [ ] 2回目以降: 前回入力値プリフィル（GET /api/liff/profile）
- [ ] バリデーション（必須チェック）
- [ ] プログレスドット表示

---

## TASK-RES-042: STEP 2 — コース選択

**UI仕様**: 仕様書セクション2.2 STEP 3（コース選択）

`is_internal = false` のコースのみ表示。リスト形式（右シェブロン付き）。

**完了条件**:
- [ ] コースリスト表示
- [ ] コース名 + コメント表示
- [ ] 曜日オプション表示

---

## TASK-RES-043: STEP 3 — スタッフ選択

**UI仕様**: 仕様書セクション2.2 STEP 2（スタッフ選択）

選択コースで絞り込み。表示フォーマット: `{施設名}　{名前}({肩書})`

**完了条件**:
- [ ] 選択コースの非対応スタッフを除外
- [ ] 「選択せずに予約」オプション（設定依存）
- [ ] スタッフリスト表示

---

## TASK-RES-044: STEP 4 — 日付選択

**UI仕様**: 仕様書セクション2.2 STEP 4

**完了条件**:
- [ ] 月間カレンダー表示
- [ ] 選択不可日グレーアウト
- [ ] 前月/翌月ナビゲーション
- [ ] 曜日色分け（日=赤、土=青）

---

## TASK-RES-045: STEP 5 — 時間選択

**UI仕様**: 仕様書セクション2.2 STEP 5

**完了条件**:
- [ ] 空き時間枠リスト表示
- [ ] 「空きのある時間のみ表示」注記

---

## TASK-RES-046: STEP 6 — 要望入力

**UI仕様**: 仕様書セクション2.2 STEP 6

**完了条件**:
- [ ] テキストエリア（プレースホルダー付き）
- [ ] 任意入力（スキップ可能）

---

## TASK-RES-047: STEP 7 — 確認画面

**UI仕様**: 仕様書セクション2.2 STEP 7

**表示項目**: お名前, 電話番号, 飼い主名, ペット情報, 診察内容, コース, スタッフ, 日時, 時間, 要望

**完了条件**:
- [ ] 全入力内容のテーブル表示
- [ ] 「まだ予約は完了していません」警告
- [ ] 確認/戻るボタン

---

## TASK-RES-048: STEP 8 — 完了画面 + エラー画面

**完了条件**:
- [ ] 予約完了メッセージ
- [ ] 同日/同月予約制限エラー画面
- [ ] メンテナンス中画面（Stopped時）

---

## TASK-RES-049: トップページ + 予約確認・キャンセル

**UI仕様**: 仕様書セクション2.0 + セクション9.3

**完了条件**:
- [ ] トップページ（アコーディオン3セクション + CTA2ボタン）
- [ ] マイ予約一覧
- [ ] 予約キャンセル（確認ダイアログ付き）
