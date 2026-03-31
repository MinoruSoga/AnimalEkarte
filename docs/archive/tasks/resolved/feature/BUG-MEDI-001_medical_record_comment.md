# BUG-MEDI-001: カルテコメント機能が未実装

## 概要
medical-records feature にコメントコンポーネントが存在しない。
カルテへのコメント投稿・編集・削除・一覧表示が全て未実装。

## 期待する動作
- カルテ詳細にコメントタブ or コメントセクションを表示
- 投稿者・投稿時刻の記録
- 自分のコメントのみ編集・削除可能
- コメントの時系列表示

## 実装場所
- `frontend/src/features/medical-records/` にコメントコンポーネント追加
- `backend/internal/` に comments エンドポイント追加（要DB設計確認）

## 優先度
Low

## 関連
- FUNCTIONAL_TEST_REPORT.md lines 2587-2589
- テスト確認日: 2026-03-30
