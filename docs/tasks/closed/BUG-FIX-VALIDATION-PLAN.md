# バグ修正検証計画 - 2件

実装: 2026-04-02 16:10 JST
デプロイ: staging (CI/CD パイプライン実行中)

---

## BUG-EXAMINATION-DATE-INPUT-MISSING (HIGH)

### 実装箇所
- `frontend/src/features/examinations/routes/ExaminationForm.tsx:125-133`
- `useExaminationForm` hook: `formData.date` 既存フィールド活用

### 検証チェックリスト

- [ ] **検査日フィールドの表示**
  - URL: `http://localhost:3003/examinations/new?petId=<id>` または既存検査編集
  - 期待: 「検査日」ラベルと NotionDatePicker が表示される
  - 仕様: 検査種別・担当医 と同じ行（grid-cols-1 md:grid-cols-2）

- [ ] **日付選択（過去日）**
  - 操作: NotionDatePicker をクリック → 今日の日付を選択
  - 期待: フォームの `formData.date` が `YYYY-MM-DDTHH:MM:SSZ` フォーマットで更新
  - 確認: 検査詳細表示時に日付が正しく保存・表示される

- [ ] **日付制約（未来日防止）**
  - 操作: NotionDatePicker をクリック → 明日の日付は選択不可
  - 期待: `disabledDays={{ after: new Date() }}` で未来日が disabled 状態
  - 確認: 明日以降の日付がグレーアウト・選択不可

- [ ] **デフォルト動作**
  - 操作: 検査日未選択のまま保存
  - 期待: `new Date().toISOString()` で現在日時が自動設定される
  - 確認: 検査記録が作成され、日付が今日に設定される

- [ ] **確定状態での動作**
  - 操作: ステータス「確定」の検査を編集
  - 期待: NotionDatePicker が read-only または disabled 状態
  - 確認: 検査日の変更が不可能（保護機構動作）

---

## BUG-INVENTORY-CROSS-FIELD-VALIDATION (MEDIUM)

### 実装箇所
- `frontend/src/features/inventory/hooks/use-inventory-form.ts:51-68`
  - バリデーション: `minStockLevel > quantity` チェック
- `frontend/src/features/inventory/routes/InventoryForm.tsx:169-180`
  - エラー表示: FormFieldError コンポーネント追加

### 検証チェックリスト

- [ ] **バリデーションエラー表示**
  - 操作: 現在庫数=10, 最低在庫数=20 → 保存ボタンクリック
  - 期待: トーストエラー表示 「最低在庫数は現在庫数以下で設定してください」
  - 確認: StockInfoSection の minStockLevel フィールド下に赤いエラーメッセージ表示

- [ ] **等値許容テスト**
  - 操作: 現在庫数=10, 最低在庫数=10 → 保存
  - 期待: バリデーション通過、正常に保存
  - 確認: 在庫記録が作成され、現在庫数=最低在庫数=10 で保存される

- [ ] **正常値テスト**
  - 操作: 現在庫数=10, 最低在庫数=5 → 保存
  - 期待: バリデーション通過、正常に保存
  - 確認: 在庫記録が作成され、値が正しく保存される

- [ ] **境界値テスト**
  - 操作: 現在庫数=1, 最低在庫数=0 → 保存
  - 期待: バリデーション通過、正常に保存
  - 操作: 現在庫数=0, 最低在庫数=1 → 保存
  - 期待: エラー表示、保存失敗

- [ ] **编集時のバリデーション**
  - 操作: 既存在庫編集 → 現在庫数=20, 最低在庫数=30 に変更 → 保存
  - 期待: エラー表示 「最低在庫数は現在庫数以下で設定してください」
  - 確認: 編集が キャンセルされ、値が更新されない

- [ ] **エラー表示の UI/UX**
  - 確認: FormFieldError が minStockLevel フィールド直下に表示
  - 確認: エラーメッセージが赤色・読みやすいフォント
  - 確認: エラー表示時にフォーカスが該当フィールドに移動（accessibility）

---

## テスト環境

**ローカル開発環境:**
```bash
make up
# http://localhost:3003 でアクセス
```

**ステージング環境:**
- URL: https://stg.noah-karte.com
- デプロイ: GitHub Actions CI/CD パイプライン（自動実行）
- ステータス: 進行中...

---

## テスト実行手順

### 1. ローカル検証（開発環境）

```bash
# Docker コンテナ起動
make up

# ブラウザで検証
# http://localhost:3003/examinations/new?petId=1
# http://localhost:3003/inventory/1/edit
```

### 2. ステージング検証（CI/CD パイプライン完了後）

```bash
# GitHub Actions デプロイ完了待機
gh run list --workflow=deploy-staging.yml -L 1

# ステージング環境でテスト
# https://stg.noah-karte.com/examinations/new?petId=1
# https://stg.noah-karte.com/inventory/1/edit
```

### 3. テスト実行スクリプト（自動化可能）

```bash
# 後日: Playwright/Cypress で E2E テスト化
# cypress run --spec="cypress/e2e/bug-examination-date.cy.ts"
# cypress run --spec="cypress/e2e/bug-inventory-validation.cy.ts"
```

---

## チェックリスト統計

- **BUG-EXAMINATION-DATE-INPUT-MISSING**: 5 チェック項目
- **BUG-INVENTORY-CROSS-FIELD-VALIDATION**: 6 チェック項目
- **合計**: 11 チェック項目

---

## 完了予定日

- ローカル検証: 2026-04-02（本日）
- ステージング検証: 2026-04-02（CI/CD 完了後）
- 本番デプロイ: 2026-04-03 以降（QA 承認後）
