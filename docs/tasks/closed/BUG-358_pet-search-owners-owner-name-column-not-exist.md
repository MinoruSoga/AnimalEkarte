# BUG-358: ペット検索で `owners.owner_name` カラムが存在せず 500 エラー

## 概要

`pet_repository.go:45` のペット検索クエリが `owners.owner_name` を参照しているが、
`owners` テーブルのカラム名は `name`。検索文字列を入力するとペット一覧 API が 500 エラーになる。

## 再現手順

1. `GET /api/v1/pets?search=太郎` を実行
2. PostgreSQL エラー: `column owners.owner_name does not exist`
3. API が 500 Internal Server Error を返す

## 該当箇所

```go
// backend/internal/repository/pet_repository.go:45（バグ）
`(pets.name ILIKE ? ESCAPE '\' OR pets.name_kana ILIKE ? ESCAPE '\' OR owners.owner_name ILIKE ? ESCAPE '\')`
//                                                                      ^^^^^^^^^^^^^^^^^^^ owners.name が正しい
```

## 修正内容

```diff
- `(pets.name ILIKE ? ESCAPE '\' OR pets.name_kana ILIKE ? ESCAPE '\' OR owners.owner_name ILIKE ? ESCAPE '\')`
+ `(pets.name ILIKE ? ESCAPE '\' OR pets.name_kana ILIKE ? ESCAPE '\' OR owners.name ILIKE ? ESCAPE '\')`
```

## 優先度

**HIGH** — ペット検索が完全に動作しない。
