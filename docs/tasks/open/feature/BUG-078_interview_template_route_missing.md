# BUG-078: 問診テンプレートルートが router.tsx 未登録

## 概要
`paths.ts` に `/settings/interview-template` パスが定義済みだが、`router.tsx` にルートが登録されていない。
設定画面から問診テンプレート管理画面にアクセスできない（404 または画面が表示されない）。

## 再現手順
1. `/settings/interview-template` に直接アクセス
2. → 404 またはルートなしのブランク表示

## 期待する動作
- 問診テンプレート管理画面が表示される
- サイドバーのリンクからも遷移できる

## 実装場所
- `frontend/src/app/router.tsx` に `/settings/interview-template` ルートを追加
- 対応するページコンポーネントが未実装の場合は作成が必要

## 優先度
Medium

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-078
- `frontend/src/lib/paths.ts` に `interviewTemplate` パスは定義済み
- テスト確認日: 2026-03-30
