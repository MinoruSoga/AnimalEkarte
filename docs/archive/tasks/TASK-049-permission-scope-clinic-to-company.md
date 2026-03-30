# TASK-049: 権限スコープを clinic 単位から company 単位へ変更

**作成日**: 2026-03-29
**ステータス**: Open
**優先度**: High
**領域**: Backend + Frontend / Architecture

---

## 背景・決定経緯

現在の権限グループは `clinic_id` でスコープされているが、
「院A では管理者・院B では一般スタッフ」という運用は不要であることが確認された。

`company` はシングルトン（1社1デプロイ）であるため、
権限グループを company 単位で定義し全院で共有する設計が正しい。

### この変更で解消されるバグ

| チケット | 内容 | 状態 |
|---------|------|------|
| BUG-058 | 権限グループのクロステナントアクセス | **本タスクで不要になるため Closed** |
| BUG-059 | switchClinic 後に JWT clinic_id が乖離 | **本タスクで不要になるため Closed** |

---

## 変更の概要

### 変更前

```
permission_groups.clinic_id = 1  →「院1の管理者グループ」
permission_groups.clinic_id = 2  →「院2の管理者グループ」（同じ定義を複製）

/me レスポンス:
permissions: {
  "1": { "medical-records": { view: true, create: true, edit: true, delete: false } },  // clinic 1
  "2": { "medical-records": { view: true, create: true, edit: true, delete: false } },  // clinic 2
}

hasPermission(resource, action):
  → user.permissions[currentClinicId][resource][action]
```

### 変更後

```
permission_groups.company_id = 1  →「管理者グループ」（全院共通）

/me レスポンス:
permissions: {
  "medical-records": { view: true, create: true, edit: true, delete: false },
  "owners":          { view: true, create: false, edit: false, delete: false },
  ...
}

hasPermission(resource, action):
  → user.permissions[resource][action]  // currentClinicId 不要
```

---

## 注意: clinic_id はデータ分離には引き続き必要

**権限スコープの変更はあくまで「誰が何をできるか」の判定部分のみ。**

カルテ・会計・予約等のデータは引き続き `clinic_id` で分離する。
JWT の `clinic_id` も「どの院のデータを操作するか」の用途で残す。

---

## 実装スコープ

### BE-082: バックエンドの権限スコープ変更

- `permission_groups.clinic_id` → `company_id` への DB 変更
- `findEffectivePermissions()` SQL から clinic_id フィルタを削除
- `/me` レスポンスの `permissions` フィールドをフラット化
- `permission_group_handler.go` の extractClinicID → extractCompanyID
- seed データの更新

### FE-139: フロントエンドの権限スコープ変更

- `AuthUser.permissions` 型を `Record<string, ResourcePermissions>` にフラット化
- `hasPermission()` から `currentClinicId` スコープを削除
- `/me` レスポンス型の更新

---

## 完了条件

- [ ] 権限グループが company 単位で定義されている
- [ ] 同一ユーザーの権限が全院で同一になっている
- [ ] カルテ・会計等のデータは従来通り clinic_id で分離されている
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec frontend npm run build` 成功
- [ ] `docker compose exec backend go test ./... -v` 成功
- [ ] `docker compose exec frontend npm run test:run` 成功

---

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BE-082 | Backend | DB + ハンドラー + サービス + リポジトリの clinic_id → company_id 変更 |
| FE-139 | Frontend | AuthUser 型 + hasPermission + usePermission の clinic スコープ削除 |
