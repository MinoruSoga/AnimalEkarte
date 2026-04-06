# BUG-156: 薬剤・診断・ケージマスタで create=F なのに「新規登録」ボタンが表示される

## 概要
`/settings/medicine`、`/settings/diagnosis`、`/settings/cage` で
`master-medical` / `master-hospitalization` の `create=false` にも関わらず
「新規登録」ボタンや操作メニューが表示される。API は 403 でブロック済み。

BUG-124/132/152 と同根の問題 — カスタムレイアウトのマスタページに権限チェックが適用されていない。

## 再現手順
1. RBAC検証用グループを全リソース view-only に設定
2. `vet@example.com` でログイン
3. `/settings/medicine` にアクセス
4. **結果**: 「閲覧のみ」バッジ + 「新規登録」ボタン（青）が共存

## ブラウザテスト結果（全26ページ）

| ページ | バッジ | Create ボタン |
|--------|--------|-------------|
| ダッシュボード〜問診 (23ページ) | ✅ | ✅ hidden |
| **薬剤** | ✅ | **❌ visible** |
| **診断** | ✅ | **❌ visible** |
| **ケージ** | ✅ | **❌ visible** |

## 修正方針
これらのページは `MasterCRUDPage` ではなくカスタムコンポーネントを使用しているため、
`usePermission` による create ボタン制御が漏れている。
各コンポーネントで `canCreate` を確認してボタンを条件表示する。

## API テスト結果
全 7 エンドポイント（POST/PATCH/DELETE × 3リソース）→ 全て **403** ✅

## 優先度
**Low** — API で 403。UI 表示のみの問題。

## 関連チケット
- BUG-124/132/152（修正済み）: 他ページの同様の問題

## 関連ファイル
- `frontend/src/features/master/routes/MedicineSettings.tsx`
- `frontend/src/features/master/routes/DiagnosisSettings.tsx`
- `frontend/src/features/master/routes/CageSettings.tsx`
