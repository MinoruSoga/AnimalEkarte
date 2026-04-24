# TASK-220: company_response.go — ID フィールドが string 型（他全 Response は uint64）

## 優先度
Medium

## 対象ファイル
`backend/internal/handler/company_response.go`

## 問題概要
`companyResponse.ID` が `string` 型で `strconv.FormatUint(c.ID, 10)` で変換されている。
他の全 Response struct（cage, clinic, exam_type, insurance, medicine 等すべて）は `ID uint64` で統一されている。

フロントエンドの `models.ts` で `company.id` だけ `string` 型になり、他エンティティの `id: number` と型が乖離する。意図的な設計であればコメントで理由を明示すること。意図がなければ `uint64` に統一すべき。

## 現状コード（行11付近）

```go
type companyResponse struct {
    ID   string `json:"id"`  // NG: 他全 Response は uint64
    Name string `json:"name"`
    ...
}

func toCompanyResponse(c model.Company) companyResponse {
    return companyResponse{
        ID:   strconv.FormatUint(c.ID, 10),  // uint64 → string 変換
        ...
    }
}
```

## あるべき姿

```go
type companyResponse struct {
    ID   uint64 `json:"id"`  // 他 Response と統一
    Name string `json:"name"`
    ...
}

func toCompanyResponse(c model.Company) companyResponse {
    return companyResponse{
        ID:   c.ID,  // 変換不要
        ...
    }
}
```

## 注意
フロントエンドで `company.id` を `string` として扱っている箇所がある場合、型変更が必要になる。
`frontend/src/` で `company` の `id` 利用箇所を事前に確認してから修正すること。

## 完了条件
- [ ] `companyResponse.ID` を `uint64` に変更
- [ ] `toCompanyResponse` から `strconv.FormatUint` を削除
- [ ] フロントエンド側の型整合を確認
- [ ] `go test ./backend/internal/...` がパス
