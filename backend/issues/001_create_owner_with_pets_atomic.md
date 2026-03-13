# POST /v1/owners — 飼主登録時にペットを同時・アトミックに作成できるようにしたい

## 背景

飼主新規登録画面では、飼主登録と同時に複数のペットを入力できる。
現在のフロントエンドは、飼主登録後にペットを1件ずつ個別にAPIへ送っている。

```
POST /v1/owners          → 201 (owner_id 取得)
POST /v1/pets            → 201 (1匹目)
POST /v1/pets            → 201 (2匹目)
POST /v1/pets            → 500 ← ここで失敗すると飼主だけ作られてペットが残らない
```

この構造では、ペット作成が途中で失敗した場合に飼主だけが存在する中途半端な状態になる。

---

## 要望

`POST /v1/owners` のリクエストボディにオプションで `pets` 配列を受け取り、
飼主とペットをひとつのDBトランザクションで作成できるようにしてほしい。

---

## 仕様

### エンドポイント

```
POST /v1/owners
```

既存のエンドポイントに `pets` フィールドを追加する形で拡張する（後方互換あり）。

### リクエストボディ（変更箇所のみ）

```json
{
  "owner_name": "林 文昭",
  "owner_name_kana": "ハヤシ フミアキ",
  "phone": "090-1234-5678",

  "pets": [
    {
      "name": "ポチ",
      "animal_species_id": 1,
      "gender": "male",
      "birth_date": "2020-04-01",
      "pet_number": "P-001",
      "breed": "柴犬",
      "weight": 8.5,
      "status": "alive",
      "insurance_id": 3,
      "remarks": ""
    }
  ]
}
```

- `pets` は省略可能（既存の呼び出しに影響なし）
- `pets` の各フィールドは `POST /v1/pets` のリクエスト仕様に準ずる
- `owner_id` と `clinic_id` はサーバー側で自動設定するのでリクエストに含めない

### レスポンス（変更箇所のみ）

作成された飼主に `pets` を含めて返す。

```json
{
  "id": 42,
  "owner_name": "林 文昭",
  "pets": [
    {
      "id": 101,
      "name": "ポチ",
      "animal_species_id": 1,
      "animal_species": { "id": 1, "name": "犬" }
    }
  ]
}
```

### エラー仕様

- 飼主またはペットのいずれかのバリデーションエラー → `400 Bad Request`
- DB書き込み失敗 → トランザクションをロールバックし `500 Internal Server Error`
- ペットの `animal_species_id` が存在しない → `400 Bad Request`

---

## 現在の回避策

フロントエンドで `Promise.all` を使い、飼主作成後にペットを並列送信している。
アトミック性がないため、この実装は暫定対応。
