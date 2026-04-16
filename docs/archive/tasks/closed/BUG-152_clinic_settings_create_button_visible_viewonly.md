# BUG-152: 医院設定ページで create=F なのに「新規登録」ボタンと「新しい医院を追加...」が表示される

## 概要
`/settings/clinic`（医院マスタ）ページで `hospital-settings: create=false` なのに
「新規登録」ボタンと「新しい医院を追加...」インラインボタンが表示される。
API は 403 で正しくブロックするが、UI が不整合。

## 脆弱性分類
- **UX 問題**（セキュリティ実害なし — API で 403）

## 再現手順
1. RBAC検証用グループ（全リソース view-only）のユーザーでログイン
2. `/settings/clinic` にアクセス
3. **結果**: 「閲覧のみ」バッジが表示されるが、「新規登録」ボタンと「新しい医院を追加...」も表示

## スクリーンショット
- 右上に「閲覧のみ」バッジ + 「新規登録」ボタン（青）が共存
- テーブル下に「+ 新しい医院を追加...」テキストリンク

## 原因
医院設定ページ（ClinicMasterSettings）は他のマスタページと異なるカスタムレイアウトで、
`MasterCRUDPage` の権限チェックを使用していない。

## 修正方針
`ClinicMasterSettings` コンポーネントで `usePermission(ResourceHospitalSettings)` を使い、
`canCreate` が false の場合はボタンを非表示にする。

## 優先度
**Low** — API で 403。UX 改善。

## 関連チケット
- BUG-132（修正済み）: 他ページの create ボタン表示漏れ

## 関連ファイル
- `frontend/src/features/hospital-settings/` — 医院設定コンポーネント
