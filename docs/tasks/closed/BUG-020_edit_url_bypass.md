# BUG-020: can_edit=false のユーザーが編集フォームを直接URLアクセス可能

## 概要
`can_edit=false` の権限を持つユーザーでも、編集フォームのURLに直接アクセスすることで
フォームが表示される。バックエンドのアクセス制御が未実装。

## 再現手順
1. 山田花子（vet@example.com、一般グループ、edit権限なし）でログイン
2. 編集フォームのURL（例: `/owners/1/edit`）に直接アクセス
3. → フォームが表示され編集可能（本来は操作ができないべき）

## 期待する動作
- フロントエンド: 権限なしの場合は編集フォームへのアクセスをブロック
- バックエンド: PATCH/PUT API で権限チェックを実装（HTTP 403 Forbidden）

## 実装場所
- `frontend/src/app/router.tsx`: 編集ルートに権限ガードを追加
- `backend/internal/middleware/` または各ハンドラ: can_edit チェック追加

## 優先度
High（セキュリティ）

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-020
- テスト確認日: 2026-03-30
