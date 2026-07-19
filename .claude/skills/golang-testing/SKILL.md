---
name: golang-testing
description: Go テストパターン。Table-driven tests、testify、モック、ベンチマーク、TDD サイクル。Go テスト作成・カバレッジ改善時に使用。
origin: ECC (adapted for AnimalEkarte)
---

# Go テストパターン

このプロジェクト（Go 1.25 / testify assert + 手書き fn-field モック）で使用するテストパターン。

## When to Activate

- 新規 Go 関数・メソッドのテスト作成
- 既存コードへのカバレッジ追加
- Service/Repository のモックテスト
- TDD ワークフロー

## テスト配置（このプロジェクト）

```
backend/internal/
├── service/
│   ├── owner_service.go
│   └── owner_service_test.go   ← 同パッケージ
├── repository/
│   ├── owner_repository.go
│   └── owner_repository_test.go
└── handler/
    ├── owner_handler.go
    └── owner_handler_test.go
```

## Table-Driven Tests（必須パターン）

モックは**手書き fn-field 差し替え型**（実例: `backend/internal/service/liff_service_mock_test.go`）。testify/mock の `m.On(...)` / `mock.Anything` / `AssertExpectations` は使わない。

```go
package service

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
)

// 手書き fn-field モック: 各テストケースで必要な fn だけ差し替える
type mockOwnerRepository struct {
    getByIDFn func(ctx context.Context, id uint) (*model.Owner, error)
}

func (m *mockOwnerRepository) GetByID(ctx context.Context, id uint) (*model.Owner, error) {
    if m.getByIDFn != nil {
        return m.getByIDFn(ctx, id)
    }
    return nil, apperrors.ErrNotFound
}

func TestOwnerService_GetOwner(t *testing.T) {
    tests := []struct {
        name    string
        id      uint
        repo    *mockOwnerRepository
        want    *model.Owner
        wantErr error
    }{
        {
            name: "正常: オーナー取得成功",
            id:   1,
            repo: &mockOwnerRepository{
                getByIDFn: func(_ context.Context, _ uint) (*model.Owner, error) {
                    return &model.Owner{ID: 1, Name: "佐藤太郎"}, nil
                },
            },
            want:    &model.Owner{ID: 1, Name: "佐藤太郎"},
            wantErr: nil,
        },
        {
            name: "エラー: 存在しないオーナー",
            id:   999,
            repo: &mockOwnerRepository{
                getByIDFn: func(_ context.Context, _ uint) (*model.Owner, error) {
                    return nil, apperrors.ErrNotFound
                },
            },
            want:    nil,
            wantErr: apperrors.ErrNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            svc := NewOwnerService(tt.repo)

            // Act
            got, err := svc.GetOwner(context.Background(), tt.id)

            // Assert
            if tt.wantErr != nil {
                assert.ErrorIs(t, err, tt.wantErr)
                assert.Nil(t, got)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

## Repository テスト（実DB使用）

```go
func TestOwnerRepository_GetByID(t *testing.T) {
    // テスト DB を使用（Docker の test DB）
    db := setupTestDB(t)  // 各 repository テスト共通ヘルパ（testutil パッケージは存在しない）

    tests := []struct {
        name      string
        setupFn   func()
        clinicID  uint64
        ownerID   uint
        want      *model.Owner
        wantErr   bool
    }{
        {
            name: "正常: clinic_id でスコープされたオーナーを取得",
            setupFn: func() {
                // テストデータ挿入
            },
            clinicID: 1,
            ownerID:  1,
            want:     &model.Owner{ID: 1, Name: "テスト太郎"},
            wantErr:  false,
        },
        {
            name: "エラー: 別の clinic_id のオーナーは取得不可（マルチテナント確認）",
            clinicID: 2, // 異なる clinic
            ownerID:  1,
            wantErr:  true, // ErrNotFound が返るべき
        },
    }
    // ...
}
```

## Handler テスト

```go
// service も同じ手書き fn-field モック形式
type mockOwnerService struct {
    getOwnerFn func(ctx context.Context, id uint) (*model.Owner, error)
}

func (m *mockOwnerService) GetOwner(ctx context.Context, id uint) (*model.Owner, error) {
    if m.getOwnerFn != nil {
        return m.getOwnerFn(ctx, id)
    }
    return nil, apperrors.ErrNotFound
}

func TestOwnerHandler_GetOwner(t *testing.T) {
    tests := []struct {
        name       string
        ownerID    string
        svc        *mockOwnerService
        wantStatus int
        wantBody   string
    }{
        {
            name:    "正常: 200 OK",
            ownerID: "1",
            svc: &mockOwnerService{
                getOwnerFn: func(_ context.Context, _ uint) (*model.Owner, error) {
                    return &model.Owner{ID: 1, Name: "佐藤太郎"}, nil
                },
            },
            wantStatus: http.StatusOK,
        },
        {
            name:    "エラー: 404 Not Found",
            ownerID: "999",
            svc: &mockOwnerService{
                getOwnerFn: func(_ context.Context, _ uint) (*model.Owner, error) {
                    return nil, apperrors.ErrNotFound
                },
            },
            wantStatus: http.StatusNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            h := NewOwnerHandler(tt.svc)

            w := httptest.NewRecorder()
            c, _ := gin.CreateTestContext(w)
            c.Params = gin.Params{{Key: "id", Value: tt.ownerID}}
            c.Request = httptest.NewRequest("GET", "/", nil)

            // Act
            h.GetOwner(c)

            // Assert
            assert.Equal(t, tt.wantStatus, w.Code)
        })
    }
}
```

## テスト実行コマンド

> ⚠️ `go test ./...` の全体実行は CLAUDE.md の自動実行禁止コマンド。スコープ限定版
> （例: `go test ./internal/service/...`）を使うか、ユーザーに手動実行を依頼する。

```bash
# 全テスト実行（⚠️ 自動実行禁止。ユーザー手動実行を依頼）
docker compose exec backend go test ./...

# カバレッジ付き（⚠️ 自動実行禁止。ユーザー手動実行を依頼）
docker compose exec backend go test ./... -cover

# 特定パッケージ（✅ スコープ限定・自動実行可）
docker compose exec backend go test ./internal/service/... -v

# レースコンディション検出（⚠️ 全体実行は自動実行禁止。スコープ限定版を使う）
docker compose exec backend go test ./... -race
docker compose exec backend go test ./internal/service/... -race

# カバレッジレポート（⚠️ 全体実行は自動実行禁止。スコープ限定版を使う）
docker compose exec backend go test ./... -coverprofile=coverage.out
docker compose exec backend go test ./internal/service/... -coverprofile=coverage.out
docker compose exec backend go tool cover -html=coverage.out
```

## テストの命名規則

```go
// パターン: Test{対象}_{シナリオ}_{期待値}
func TestOwnerService_GetOwner_Success(t *testing.T)
func TestOwnerService_GetOwner_NotFound(t *testing.T)
func TestOwnerService_Create_DuplicateEmail(t *testing.T)
```

## カバレッジ要件

```
Project quality gate: 最低 80% カバレッジ（Go/Gin公式要件ではない）
HTTP boundary: 正常系 + binding/validation/authn/authz/主要errorをhttptestでカバー
Business/security invariant: 分岐と境界条件を重点的にカバー
Persistence/transaction: 実DB testで主要query・atomicity・tenant isolationをカバー
```

## テストアンチパターン（禁止）

```go
// ❌ テスト間で状態を共有
var globalCounter int

// ❌ time.Sleep でタイミング調整
time.Sleep(100 * time.Millisecond)

// ❌ エラーを無視
result, _ := svc.GetOwner(ctx, 1)

// ❌ コンテキストなし
svc.GetOwner(context.TODO(), 1) // context.Background() を使う

// ✅ テストは独立
func TestOwnerService_GetOwner(t *testing.T) {
    // 各テストで新しい mock を作成
    mockRepo := &mockOwnerRepository{}
}
```

## 判定ロジックの struct リテラルテストは全フィールド明示（暗黙ゼロ値禁止）

CPM/タグ/ステージ判定系のテストケースでは、判定に関与しうる**全フィールドを明示的にセット**する。「関係ない」フィールドも `AnnualVisitCount: 0` と明示する。暗黙ゼロ値依存テストは条件追加時に本筋と無関係な FAIL を起こす。ゼロを書く場合も「意味的にゼロが正しいか」で判断する（0 来院で LTV 5000 は矛盾）。

（出典: memory feedback_implicit_zero_test_fragility / be_second_lens_audit_20260630）

## 置換系（Replace/Upsert）テストは seed step を先置きする

`ReplaceForX` 型のテストで既存行 seed を忘れると `deleted=0` で暗黙 PASS/FAIL する罠がある。削除・置換を検証するテストは Arrange で必ず対象行を seed してから Act する。

（出典: memory issue211_audit_tx_atomicity_verify_20260630 / commit fe04b460）

## インベントリ lint テスト（実績パターン）

「あるべき集合」と「実コードの集合」を go/ast で双方向突合し、追加漏れを CI で fail させる。実例: preload_clinic_scope_lint_test.go / master_fk_write_inventory_lint_test.go / audit_taxonomy_exhaustiveness_test.go。
- 実装: go:embed *.go + go/parser + go/ast（trimpath 耐性）、_test.go は runtime skip
- **anti-vacuous 保証必須**: 検出件数の floor / occurrence-count pin / 例外リストが生きていることの検証。無いと対象ゼロでも GREEN になる空虚テスト
- **非空虚性は temp-revert で実証**: 述語を一時的に外して RED を確認 → 戻して GREEN。tracked ファイルは HEAD と byte-identical に復元
（出典: memory preload_clinic_scope_lint_p0_20260630 / issue211_audit_tx_inventory_lint_20260630、commit 8a51c2eb / d67469aa）

## CASCADE 挙動テストは実 DDL で（実績パターン）

GORM の AutoMigrate は ON DELETE 句を再現しない。CASCADE / SET NULL の挙動テストは migration から実 DDL を抽出してテーブルを再作成した上で delete を発火させる。実例: checkup_field_cascade_test.go（SET NULL→CASCADE の一時改変で三重 RED 実証済み）
（出典: memory issue211_checkup_package_spec_20260630）

## repository テスト3つの罠（実績由来）

1. **warm-DB が fresh-DB 失敗をマスク**: 前 run のテーブル残存で PASS しても CI の fresh DB で FAIL する。`DROP DATABASE ekarte_db_test` 後に1回走らせて確認
2. **setupTestDB の手書き ENUM リストは migration とドリフトする**: ENUM 追加 migration のたびに欠落し、カスケード失敗を起こす（migration009 の4 ENUM 欠落で25テスト失敗の実例）
3. **コネクション枯渇**: 接続を閉じないと full suite で too many clients (53300)。SetMaxOpenConns + t.Cleanup(Close)
（出典: memory ops_golangci_lint_cap_and_reconcile_20260630 / issue196_clinic_isolation_tests_complete_20260626）
