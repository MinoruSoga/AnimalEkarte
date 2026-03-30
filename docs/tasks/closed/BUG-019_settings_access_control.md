# BUG-019: can_view=false でも設定ページにアクセス可能（フロントエンドUIコントロール漏れ）

## 概要
`master:can_view=false` の権限を持つ一般グループユーザーでも、
設定ページ（`/settings`）に直接URLでアクセスできる。
バックエンドのAPIアクセス制御はあるが、フロントエンドのルートガードが不完全。

## 再現手順
1. 山田花子（vet@example.com、一般グループ）でログイン
2. `/settings/animal-species` に直接URLでアクセス
3. → ページが表示される（本来は 403 またはリダイレクトであるべき）

## 期待する動作
- `master:can_view=false` のユーザーが設定ページにアクセスした場合、403エラーまたはダッシュボードへリダイレクト

## 実装場所
- `frontend/src/app/router.tsx` のルートガード（権限チェック）
- 設定系ルートに対して権限チェックを追加

## 優先度
Medium

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-019
- テスト確認日: 2026-03-30
