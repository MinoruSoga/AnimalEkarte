# BE-050: トリミング一覧 500エラー — body_weight スキャンエラー

## 問題

`/trimming` ページで「データの取得に失敗しました」エラーが表示。

```
trimming_repository.go:53
Scan error on column index 9, name "bw": converting driver.Value type string ("") to a float64: invalid syntax
```

## 根本原因

`trimming_repository.go` の `FindAll()` で `ownerID` フィルタ時に `pets` テーブルを JOIN しているが、`Select` 指定がないため `pets.*` 全カラムが SELECT される。

`pets.weight` カラム（`numeric` 型）にデータが空文字列 `""` で格納されているレコードがあり、Go の `float64` にスキャンできない。

GORM が JOIN 結果のカラムを `TrimmingRecord` 構造体にマッピングする際、`pets.weight` が `TrimmingRecord.BW`（エイリアス `bw`）にスキャンされてしまう。

## 修正方針

JOIN 時に `Select("trimming_records.*")` を明示し、pets テーブルのカラムがスキャン対象に含まれないようにする。

```go
// 修正前
if ownerID != nil {
    q = q.Joins("JOIN pets ON pets.id = trimming_records.pet_id").
        Where("pets.owner_id = ?", *ownerID)
}

// 修正後
if ownerID != nil {
    q = q.Joins("JOIN pets ON pets.id = trimming_records.pet_id").
        Select("trimming_records.*").
        Where("pets.owner_id = ?", *ownerID)
}
```

追加対策として、`pets.weight` カラムに空文字列が入っているデータを NULL に修正する seed/migration 修正も検討する。

## 対象ファイル

- `backend/internal/repository/trimming_repository.go` — FindAll メソッド（40行目付近）
- `backend/migrations/002_seed_master.sql` — pets の weight データ確認

## テスト

```bash
docker compose exec backend go test ./internal/repository/... -v -run Trimming
curl -s http://localhost:8080/api/v1/trimmings | jq .
```
