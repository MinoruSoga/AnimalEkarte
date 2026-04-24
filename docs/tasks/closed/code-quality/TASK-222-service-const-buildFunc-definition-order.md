# TASK-222: service — const / buildFunc の定義順序が規約違反（2ファイル）

## 優先度
Medium

## 対象ファイル
- `backend/internal/service/permission_group_service.go`
- `backend/internal/service/reservation_type_group_service.go`

## 問題概要
プロジェクト規約の正しい定義順序：

```
1. Create{Entity}Input
2. Update{Entity}Input
3. const (colXxx = "xxx")    ← ここに列名定数
4. buildXxxUpdateFields()     ← ここにビルダー関数
5. type {Entity}Service interface
6. type {Entity}Service struct
7. func New{Entity}Service(...)
8. func methods...
```

以下のファイルでこの順序が守られていない。

### 1. `permission_group_service.go`（行253付近）
`const (colPermissionGroup...)` と `buildPermissionGroupUpdateFields` が
Interface・実装の**後ろ**に定義されている。ファイルの末尾に追記されたものと推測される。

### 2. `reservation_type_group_service.go`（行120付近）
`buildReservationTypeGroupUpdateFields` が `Update` メソッド（96行）よりも**後ろ**に定義されており、
呼び出し元が定義より前にある。他全 service では const/build 関数は Interface より前に置かれている。

## 比較（正しい実装例）
`animal_species_service.go`・`cage_service.go` 等の先頭部分で
const → buildFunc → Interface → struct の順序を確認してください。

## 修正方針
各ファイルで const と buildFunc を Interface 定義より前に移動する（ロジックの変更なし）。

## 完了条件
- [ ] `permission_group_service.go` の const と buildFunc を Interface より前に移動
- [ ] `reservation_type_group_service.go` の buildFunc を Interface より前に移動
- [ ] `go test ./backend/internal/...` がパス（ロジック変更なしのため確認のみ）
