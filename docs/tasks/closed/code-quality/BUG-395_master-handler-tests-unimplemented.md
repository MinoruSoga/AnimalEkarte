# BUG-395: マスタハンドラのテストが実装されていない（コンパイル確認のみ）

## 概要
マスタ関連ハンドラ 17 ファイルすべての `*_handler_test.go` が「コンパイル成功確認」のみのダミーテストになっている。実際のハンドラロジック（バリデーション・認可・レスポンス構造・エラーケース）を検証するテストが一切存在しない。各テストファイルには詳細なテスト仕様コメントが記述されているが、実装は空である。

## 再現手順
```bash
docker compose exec backend go test ./internal/handler/... -v -run "TestCheckupType"
```
**結果**:
```
--- PASS: TestCheckupTypeHandlerCompiles (0.00s)
PASS
```

コンパイルが通ることの確認しかしていない。実際のハンドラ動作は検証されていない。

## 期待する動作
- 各エンドポイントのハッピーパスと主要エラーケースをカバーするテストが実装されている
- `httptest.NewRecorder()` を使用したハンドラ単体テストが存在する
- 少なくとも 80% のカバレッジ基準を満たしている

## 現状コード（代表例）

### `backend/internal/handler/checkup_type_handler_test.go`
```go
package handler

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

// TestCheckupTypeHandlerCompiles verifies checkup_type_handler.go compiles
func TestCheckupTypeHandlerCompiles(t *testing.T) {
    assert.True(t, true, "checkup_type_handler.go compiled successfully")
    // ← 実際のテストロジックなし
}

// ---- Comprehensive Test Coverage Documentation ----
// (テスト仕様コメント 100行以上が続くが、実装なし)
```

### 未実装のマスタハンドラテストファイル（17ファイル）
```
animal_species_handler_test.go
cage_handler_test.go
checkup_type_handler_test.go
chief_complaint_handler_test.go
diagnosis_handler_test.go
exam_type_handler_test.go
insurance_handler_test.go
medicine_handler_test.go
merchandise_item_handler_test.go
occupation_handler_test.go
procedure_handler_test.go
reservation_type_group_handler_test.go
reservation_type_handler_test.go
reservation_type_liff_handler_test.go
trimming_handler_test.go
trimming_master_handler_test.go
vaccine_handler_test.go
```

## 影響範囲

バグ・回帰を検出できない主要エンドポイント（各ファイルにつき 5-6 エンドポイント × 17 ファイル = 約 85 エンドポイント）:
- `List`, `Get`, `Create`, `Update`, `Delete`, `Reorder`（ハンドラにより異なる）

## 修正方針

各ハンドラテストファイルに以下の形式でテストを実装する。モックを使いサービス層への依存を注入する。

### 実装例（`cage_handler_test.go` の Create エンドポイント）
```go
func TestCageHandler_CreateCage(t *testing.T) {
    tests := []struct {
        name       string
        body       string
        mockSetup  func(*MockCageService)
        wantStatus int
    }{
        {
            name: "正常系: 200 OK",
            body: `{"name":"A棟 1号室","cage_type":"standard","capacity":1}`,
            mockSetup: func(m *MockCageService) {
                m.On("Create", mock.Anything, uint64(1), mock.Anything).
                    Return(&model.Cage{ID: 1, Name: "A棟 1号室"}, nil)
            },
            wantStatus: http.StatusCreated,
        },
        {
            name:       "異常系: name 欠落 → 400",
            body:       `{"cage_type":"standard"}`,
            mockSetup:  func(m *MockCageService) {},
            wantStatus: http.StatusBadRequest,
        },
        {
            name: "異常系: 使用中 → 409",
            body: `{"name":"A棟 1号室","cage_type":"standard","capacity":1}`,
            mockSetup: func(m *MockCageService) {
                m.On("Create", mock.Anything, uint64(1), mock.Anything).
                    Return(nil, apperrors.WrapConflict("このケージ名は既に使用中です"))
            },
            wantStatus: http.StatusConflict,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockSvc := NewMockCageService()
            tt.mockSetup(mockSvc)
            handler := &Handler{...}

            w := httptest.NewRecorder()
            req := httptest.NewRequest("POST", "/cages", strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")

            ginCtx := setupGinContext(w, req, clinicID)
            handler.CreateCage(ginCtx)

            assert.Equal(t, tt.wantStatus, w.Code)
            mockSvc.AssertExpectations(t)
        })
    }
}
```

### 優先度の高いハンドラ（先に実装するもの）
1. `medicine_handler_test.go` — 在庫連携・階層構造・複雑なビジネスロジック
2. `reservation_type_handler_test.go` — 予約不可時間・LIFF 連携
3. `exam_type_handler_test.go` — `ExamTypeField` を含む子リソース

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/testing.md` — カバレッジ基準
> New features: Minimum 80% coverage

### `.claude/rules/testing.md` — テスト構造
> Use table-driven tests for multiple cases. Follow AAA pattern: Arrange, Act, Assert.

## 優先度
**High** — 17 個のエンドポイントグループが無テスト状態。リグレッションを検出できないリスクが高い。コード品質・デプロイ安全性に直結する。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/handler/*_handler_test.go`（上記17ファイル）— すべて修正対象
- `backend/internal/handler/*_handler.go`（対応するハンドラ実装）
