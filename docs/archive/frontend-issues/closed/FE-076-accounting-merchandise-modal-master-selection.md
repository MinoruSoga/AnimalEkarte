# FE-076: 会計「物販・その他追加」モーダルをマスタ選択方式に変更

## 背景

現在の「物販・その他追加」モーダルは品目名・単価を手動入力する方式。
BE-046 の物販マスタ化に伴い、マスタから選択する方式に変更する。

## 依存

- BE-046（物販マスタ CRUD API）が完了していること

## 要件

1. `AccountingDetail.tsx` の `ItemListCard` 内モーダルを改修
2. 現在の手動入力フォーム（品目名テキスト入力 + 単価入力）を廃止
3. マスタから選択する方式に変更:
   - カテゴリ選択（food / goods / other）でフィルタ
   - マスタ一覧から品目を選択（クリックで追加）
   - 選択時に name, unit_price, tax_rate, category が自動セット
4. 手動入力は不可（マスタ選択のみ）
5. API: `GET /v1/merchandise-items?category=xxx` でマスタ取得

## 変更対象ファイル

- `frontend/src/features/accounting/routes/AccountingDetail.tsx` の `ItemListCard` コンポーネント
- `frontend/src/features/accounting/api/` に merchandise API hook 追加

## Figmaデザイン

未定。既存の `MasterSelectModal` コンポーネントの使用を検討。

## 受入条件

- [ ] モーダルでマスタから品目を選択可能
- [ ] カテゴリフィルタ機能
- [ ] 選択時に name, unit_price, tax_rate, category が自動セット
- [ ] 手動入力フィールドの廃止
