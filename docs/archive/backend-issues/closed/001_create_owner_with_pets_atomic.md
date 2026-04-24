---
status: closed
closed_at: 2026-03-13
commit: e1b3577
---

# POST /v1/owners — 飼主登録時にペットを同時・アトミックに作成できるようにしたい

## 背景

飼主新規登録画面では、飼主登録と同時に複数のペットを入力できる。
現在のフロントエンドは、飼主登録後にペットを1件ずつ個別にAPIへ送っていた。

```
POST /v1/owners          → 201 (owner_id 取得)
POST /v1/pets            → 201 (1匹目)
POST /v1/pets            → 201 (2匹目)
POST /v1/pets            → 500 ← ここで失敗すると飼主だけ作られてペットが残らない
```

この構造では、ペット作成が途中で失敗した場合に飼主だけが存在する中途半端な状態になった。

## 対応内容

- `POST /v1/owners` のリクエストボディに `pets[]` を追加（後方互換あり）
- `OwnerService.CreateWithPets()` でトランザクション内アトミック作成を実装
- `OwnerRepository` にトランザクション対応の `CreateWithPets()` を追加

## 完了条件

- [x] `POST /v1/owners` に `pets` 配列を渡すと飼主とペットがアトミックに作成される
- [x] `pets` 省略時は既存挙動（飼主のみ作成）と互換性あり
- [x] いずれかのペット作成失敗時にトランザクションがロールバックされる
- [x] テスト追加済み（`owner_service_test.go`）
