# FE-075: 物販・その他マスタ管理画面の追加

## 背景

BE-046 で物販マスタ API が実装された後、マスタ管理画面（CRUD）が必要。

## 依存

- BE-046（物販マスタ CRUD API）が完了していること

## 要件

1. `features/master/` に物販マスタ管理画面を追加
2. 既存のマスタ管理画面（medicines, procedures 等）のパターンに準拠
3. 一覧表示（DataTable）: name, category, unit_price, tax_rate, is_active
4. 新規作成・編集モーダル
5. 削除（論理削除）
6. 並べ替え（ドラッグ or 矢印ボタン）
7. カテゴリフィルタ（food / goods / other）

## Figmaデザイン

未定。デザイン確定後に実装する。

## 受入条件

- [ ] マスタ管理画面で CRUD 操作が可能
- [ ] カテゴリ別フィルタ
- [ ] 並べ替え
- [ ] router.tsx にルート追加
