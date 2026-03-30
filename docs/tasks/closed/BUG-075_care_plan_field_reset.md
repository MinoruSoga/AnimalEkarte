# BUG-075: CarePlanDialog でタイプ変更後も旧フィールド値が残留・送信される

## 概要
入院治療プランの CarePlanDialog でタイプ（種別）を変更しても、
旧条件の unitPrice・masterId・category が formData に残留したまま送信される。
「マスタ連動中」バッジも消えない。

## 再現手順
1. 入院ページで「プラン追加」ダイアログを開く
2. タイプを選択し、単価等を入力
3. タイプを別の種別に変更
4. 保存
5. → 旧タイプの unitPrice・masterId が送信データに含まれる

## 期待する動作
- タイプ変更時に関連フィールド（unitPrice、masterId、category）をクリアする
- 「マスタ連動中」バッジはマスタ連動タイプのみに表示

## 実装場所
- `frontend/src/features/hospitalization/` または `medical-records/` の CarePlanDialog コンポーネント
- `CarePlanDialog.tsx` L128 付近のタイプ変更ハンドラを修正

## 優先度
Medium

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-075
- テスト確認日: 2026-03-30
