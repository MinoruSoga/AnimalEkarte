# BUG-275: 複合 Medium 違反（第4回監査）

## 概要

個別チケットを起こすほどではない Medium レベルの違反をまとめたチケット。

## 1. swaggerignore タグ残存（~15ファイル）

swag は廃止済み（codegen は tygo に移行）だが `swaggerignore:"true"` タグが DeletedAt フィールド等に残存。

**該当ファイル**: account.go, checkup_record.go, checkup_type.go, estimate.go, examination_record.go, hospitalization.go, medical_record.go, owner.go, permission_group.go, pet.go, reservation.go, treatment.go, trimming.go, vaccination_record.go, accounting.go

**修正**: `swaggerignore:"true"` を全削除。`json:"-"` は残す。

## 2. liff_auth.go URL エンコード問題

`middleware/liff_auth.go:141-144` で `idToken` と `clientID` を文字列連結で POST body に構築。
`url.Values` を使って適切にエンコードすべき（パラメータ注入リスク）。

```go
// BEFORE
req, err := http.NewRequestWithContext(ctx, http.MethodPost, lineVerifyURL, strings.NewReader(
    "id_token="+idToken+"&client_id="+clientID,
))

// AFTER
body := url.Values{}
body.Set("id_token", idToken)
body.Set("client_id", clientID)
req, err := http.NewRequestWithContext(ctx, http.MethodPost, lineVerifyURL,
    strings.NewReader(body.Encode()))
```

## 3. audit_log.go OldValue/NewValue 型

`model/audit_log.go:14-15` で `OldValue []byte` / `NewValue []byte` が `json:"old_value"` / `json:"new_value"` でシリアライズ。
`[]byte` は JSON で base64 エンコードされるため、`json.RawMessage` に変更すべき。

## 4. liff_handler.go NOTE コメント欠落

`handler/liff_handler.go:199-205` の `c.JSON(http.StatusConflict, ...)` に規約要求の `// NOTE: Intentional direct response` コメントがない。

## 5. reservation_course_service.go FK 依存チェック欠落

`service/reservation_course_service.go:130` の Delete で予約実績の存在確認なし。
DB の FK 制約で 500 エラーになるが、プロジェクト規約では `apperrors.WrapConflict()` → 409 を返すべき。

## 6. staff_clinic_assignment_service.go Wrap に動的値

`service/staff_clinic_assignment_service.go:31,39` で `apperrors.Wrap` の message に `fmt.Sprintf` で動的値を埋め込み。
静的文字列にとどめ、動的情報は slog で記録するのが規約。

## 優先度

**Medium** — 個々は軽微だが、項目2（URLエンコード）はセキュリティ関連のため優先対応を推奨。

## 関連チケット

- BUG-270: 第4回監査 親チケット
