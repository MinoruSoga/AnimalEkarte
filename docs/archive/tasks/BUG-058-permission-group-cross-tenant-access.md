# BUG-058: 権限グループの取得・更新・削除で clinic_id チェックなし（クロステナントアクセス可能）

**作成日**: 2026-03-29
**ステータス**: Closed（Superseded by TASK-049 / BE-082）
> 権限グループを company 単位にフラット化する（TASK-049）ことで、クロスクリニックアクセスの問題が構造的に解消される。本バグは TASK-049 完了後に無効となる。
**優先度**: Critical
**領域**: Backend
**関連**: BUG-056, TASK-048

---

## 背景・問題

`GET/PUT/DELETE /v1/permission-groups/:id` および `POST /v1/permission-groups/:id/rules` において、
ハンドラーが `clinic_id` を検証していない。

**クリニック A のユーザーが、ID を推測してクリニック B の権限グループを取得・変更・削除できる。**

マルチテナントの根幹を脅かす CRITICAL バグ。

---

## 解決方針

TASK-049 で権限グループのスコープを `clinic_id` → `company_id` にフラット化することで構造的に解消。
`company` 単位のグループになれば、クリニック間での横断アクセスの概念自体がなくなる。

---

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BUG-058（BE） | Backend | Superseded by TASK-049/BE-082 — 個別修正は不要 |

詳細は `backend/issues/closed/BUG-058-permission-group-cross-tenant-access.md` を参照。

---

## 依存・関連

- **TASK-049 / BE-082**: 権限スコープ company 単位移行（本バグを構造的に解消する）
- **BUG-056 / BE-080**: バックエンド認可ミドルウェア（`RequirePermission` / `RequireClinicAdmin`）
