# BE-038: トリミングオプション 組合せ可否（combinable）が保存されない

## 重大度
**Medium** — UIで「不可」に設定しても DB は `combinable=true` のまま

## 症状

1. トリミングオプション新規登録フォームで「組合せ可否」ボタンを「可」→「不可」に切り替える
2. 保存ボタンをクリック → 「登録しました」トースト表示
3. DB確認: `combinable=true` のまま（`false` に変わっていない）
4. 編集フォームを再度開くと「可」と表示される

## 再現確認済み

- テストオプション_全項目入力（id=6, clinic_id=3）
- UIで「不可」に設定 → 保存 → DB: `combinable=t`

## 調査ポイント

1. フロントエンド: 組合せ可否ボタンのトグル状態がAPIリクエストに正しく含まれているか確認
   - `features/trimming-master/` のフォームコンポーネントを確認
   - POSTリクエストボディに `combinable: false` が含まれているか
2. バックエンド: `CreateTrimmingOption` / `UpdateTrimmingOption` ハンドラで `combinable` フィールドが正しく処理されているか確認

## 発見

マスタ設定ページ全ページ 登録テスト中に ChromeMCP で発見（2026-03-16）
