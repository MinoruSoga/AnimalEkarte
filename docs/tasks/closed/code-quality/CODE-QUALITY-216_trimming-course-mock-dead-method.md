# CODE-QUALITY-216: trimming_course_service_test.go モックに存在しないメソッドが定義されている

## 概要

`backend/internal/service/trimming_course_service_test.go` のモック構造体に
`CountRecordsByTypeID(ctx, clinicID)` メソッドが定義されているが、
`TrimmingCourseRepository` インターフェースには存在しない。

これはリファクタリング時に生じたデッドコードで、インターフェースとモックの乖離を示す。
テスト自体はコンパイルが通っているが（Go は構造体が余分なメソッドを持っても問題ない）、
モックが「実際に使われていない古いシグネチャ」を保持し続けることはメンテナンス負債になる。

## 該当箇所

### モック定義（`trimming_course_service_test.go`）

```go
type mockTrimmingCourseRepository struct {
    // ...
    countRecordsByTypeIDFn func(ctx context.Context, clinicID uint64) (int64, error)
}

func (m *mockTrimmingCourseRepository) CountRecordsByTypeID(
    ctx context.Context, clinicID uint64,
) (int64, error) {
    return m.countRecordsByTypeIDFn(ctx, clinicID)
}
```

シグネチャ: `CountRecordsByTypeID(ctx, clinicID uint64) (int64, error)` — 引数2つ

### 実際のインターフェース（`trimming_course_repository.go`）

`TrimmingCourseRepository` インターフェースには `CountRecordsByTypeID` は**存在しない**。
存在するのは:
```go
CountRecordsByCourseID(ctx context.Context, clinicID, courseID uint64) (int64, error)
```
シグネチャ: `CountRecordsByCourseID(ctx, clinicID, courseID uint64) (int64, error)` — 引数3つ

## 乖離の背景

リファクタリング時に:
1. メソッド名が `CountRecordsByTypeID` → `CountRecordsByCourseID` に変更された
2. 引数が `(clinicID)` → `(clinicID, courseID)` に増えた

しかし `service_test.go` のモックが更新されなかった。

## 修正方針

### 1. デッドメソッドをモックから削除

`trimming_course_service_test.go` のモック構造体から以下を削除:
- `countRecordsByTypeIDFn` フィールド
- `CountRecordsByTypeID` メソッド実装

### 2. `CountRecordsByCourseID` をモックに追加（未実装なら）

現在 `Delete` サービスが `CountRecordsByCourseID` を呼ぶのに、
モックに対応メソッドがなければテストが不完全。
確認して欠落していれば追加する:

```go
countRecordsByCourseIDFn func(ctx context.Context, clinicID, courseID uint64) (int64, error)

func (m *mockTrimmingCourseRepository) CountRecordsByCourseID(
    ctx context.Context, clinicID, courseID uint64,
) (int64, error) {
    return m.countRecordsByCourseIDFn(ctx, clinicID, courseID)
}
```

### 3. Delete テストに依存チェック検証を追加

`Delete` が `CountRecordsByCourseID > 0` のとき 409 を返すケースの
テストが欠落していれば追加する。

## 優先度

MEDIUM — コンパイルは通るが、モックが古いシグネチャを保持し続けることで
「テストがインターフェースを正しく反映している」という前提が崩れる。

## 参照

- `backend/internal/repository/trimming_course_repository.go` — 実際のインターフェース
- `backend/internal/service/trimming_course_service_test.go` — 修正対象
