# BUG-387: 2回目以降のLINE予約でお客様情報が自動復元されない

**作成日**: 2026-04-15
**Status**: CLOSED
**Priority**: **MEDIUM** (仕様上は自動復元されるべき情報が毎回手入力になり、再予約UXが破綻している)
**Affects**: `backend/internal/service/liff_service.go`, `backend/internal/handler/liff_handler.go`, `frontend/line-reserve/src/api/liff-api.ts`, `frontend/line-reserve/src/pages/CustomerInfoPage.tsx`

---

## 概要

同じ LINE アカウントで 2回目以降の予約を行っても、`飼い主名` `電話番号` `ペット情報` が前回入力値から復元されず、毎回入力し直しになる。仕様では 2回目以降は前回入力値を表示する想定であり、現状は実機再現済みの不具合である。

## 脆弱性分類（セキュリティ系の場合）

- 該当なし（本 issue は LIFFプロフィール復元の機能不具合）

## 再現手順

1. 同一の LINE アカウントで LIFF 予約画面を開く
2. STEP 1 で `電話番号` `飼い主名` `ペット情報` を入力して予約を確定する
3. 予約完了後、同じ LINE アカウントのまま再度 LIFF 予約画面を開く
4. STEP 1 の `お客様情報` 画面へ進む
5. **結果**: 前回入力した `電話番号` `飼い主名` `ペット情報` が自動表示されず、再入力が必要になる

## 期待する動作

- 同一 `line_user_id` の再予約では前回入力値を `GET /api/liff/:clinicId/profile` から取得できること
- `電話番号` `飼い主名` `ペット情報` が STEP 1 初期値へ反映されること
- 既存オーナー紐付け済みの場合は owner 情報を優先し、未紐付け時は `additional_fields` をフォールバックすること

## 現状コード

### `backend/internal/service/liff_service.go:268`
```go
// 顧客の追加フィールドを更新（プロフィール自動保存）
if len(input.CustomerFields) > 0 && string(input.CustomerFields) != "{}" {
	if err := s.customerRepo.UpdateAdditionalFields(ctx, clinicID, customerID, input.CustomerFields); err != nil {
		slog.WarnContext(ctx, "failed to update customer additional fields (best-effort)", "error", err)
	}
}
```

### `frontend/line-reserve/src/api/liff-api.ts:29`
```ts
getProfile: async (clinicId: string, idToken: string): Promise<LiffProfile> => {
  const res = await httpClient.get<LiffProfile>(`/api/liff/${clinicId}/profile`, {
    headers: authHeaders(idToken),
  });
  return res.data;
},
```

### `frontend/line-reserve/src/pages/CustomerInfoPage.tsx:45`
```tsx
const [name, setName] = useState(() => initialInfo.name || owner?.owner_name || f?.name || '');
const [phone, setPhone] = useState(() => initialInfo.phone || owner?.phone || f?.phone || '');
const [ownerName, setOwnerName] = useState(
  () => initialInfo.ownerName || owner?.owner_name || f?.owner_name || '',
);

const [newPets, setNewPets] = useState<PetSelection[]>(() => {
  const restored = restorePetsFromProfile(profile, initialInfo.pets);
  return restored.filter(p => p.isNew === true);
});
```

### `backend/internal/repository/line_customer_repository.go:55`
```go
func (r *lineCustomerRepository) FindOrCreateByLineUserID(ctx context.Context, clinicID uint64, lineUserID, displayName string) (*model.LineCustomer, error) {
	var c model.LineCustomer
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("line_user_id = ?", lineUserID).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c = model.LineCustomer{
			ClinicID:    clinicID,
			LineUserID:  lineUserID,
			DisplayName: displayName,
		}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```md
// docs/line/reservation-spec.md:1545
| お名前 | * | LINE表示名をデフォルト表示 | 前回入力値を表示 |
| 電話番号 | * | 空欄 | 前回入力値を表示 |
| 飼い主名 | * | 空欄 | 前回入力値を表示 |
| ペットの名前と種類 | * | 空欄 | 前回入力値を表示 |
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `backend/internal/service/liff_service.go` | 予約確定時の `additional_fields` 保存 | 影響あり |
| `backend/internal/handler/liff_handler.go` | `GET /api/liff/:clinicId/profile` のプロフィール返却 | 影響あり |
| `frontend/line-reserve/src/api/liff-api.ts` | LIFFプロフィール取得 | 影響あり |
| `frontend/line-reserve/src/pages/CustomerInfoPage.tsx` | STEP 1 の初期値復元 | 影響あり |
| `backend/internal/repository/line_customer_repository.go` | 同一 `line_user_id` への顧客レコード解決 | 影響あり |
| `docs/line/reservation-spec.md:1543` | 2回目以降に前回入力値を表示する仕様 | 仕様不一致 |

## 修正方針

### 1. LIFFプロフィール復元の回帰テストを追加する — `backend/internal/handler/liff_handler_test.go`, `frontend/line-reserve/src/pages/CustomerInfoPage.test.tsx`
```go
func TestGetLiffProfile_ReturnsAdditionalFieldsForRepeatCustomer(t *testing.T) {
	// Arrange: line_customer.additional_fields に phone / owner_name / pets を保存
	// Act: GET /api/liff/:clinicId/profile
	// Assert: レスポンスに追加フィールドが含まれる
}
```

```tsx
it("should prefill phone ownerName and pets from profile.additional_fields", () => {
  render(
    <CustomerInfoPage
      profile={{
        line_user_id: "line-1",
        display_name: "テスト",
        additional_fields: {
          phone: "090-1234-5678",
          owner_name: "田中太郎",
          pets: [{ name: "ポチ", type: "柴犬", is_new: true }],
        },
      }}
      initialInfo={{ name: "", phone: "", ownerName: "", pets: [] }}
      onNext={vi.fn()}
      onBack={vi.fn()}
    />
  );

  expect(screen.getByDisplayValue("090-1234-5678")).toBeInTheDocument();
  expect(screen.getByDisplayValue("田中太郎")).toBeInTheDocument();
  expect(screen.getByText("ポチ")).toBeInTheDocument();
});
```

### 2. `GET /profile` のレスポンス整形を固定し、返却契約を明示する — `backend/internal/handler/liff_handler.go:32`
```go
func (h *Handler) GetLiffProfile(c *gin.Context) {
	clinicID, ok := extractClinicIDFromParam(c)
	if !ok {
		return
	}
	customerID, ok := middleware.ExtractLiffCustomerID(c)
	if !ok {
		RespondError(c, apperrors.WrapUnauthorized("missing customer id"))
		return
	}
	profile, err := h.svc.Liff.GetProfile(c.Request.Context(), clinicID, customerID)
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"line_user_id":       profile.LineUserID,
		"display_name":       profile.DisplayName,
		"additional_fields":  json.RawMessage(profile.AdditionalFields),
		"owner_id":           profile.OwnerID,
		"owner":              profile.Owner,
	})
}
```

### 3. 初期値優先順を `profile` 優先に統一する — `frontend/line-reserve/src/pages/CustomerInfoPage.tsx:45`
```tsx
const [name, setName] = useState(() => initialInfo.name || f?.name || owner?.owner_name || '');
const [phone, setPhone] = useState(() => initialInfo.phone || f?.phone || owner?.phone || '');
const [ownerName, setOwnerName] = useState(
  () => initialInfo.ownerName || f?.owner_name || owner?.owner_name || '',
);
```

`initialInfo` が空文字のまま渡るケースでも、APIから返った `additional_fields` を確実に利用できるように順序を明示する。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/api.md` — API Development Rules
> Return consistent error response format

プロフィール復元フローは API 契約が曖昧だとフロントとバックエンドで解釈がずれるため、`GET /profile` のレスポンス構造を固定しテストで保証すること。

### `.claude/CLAUDE.md` — Frontend ベストプラクティス参照実装
> **型安全性最優先**: Go/TypeScript 共に `any` を禁止し、厳格な型定義を行う。

`LiffProfile` と `additional_fields` の契約を実装/テストで揃え、暗黙の JSON 依存を避けること。

### `.claude/rules/typescript-react.md` — TypeScript / React 19 Rules
> **❌ 禁止: any**

LIFFプロフィール復元でも `additional_fields` の型を明示し、`unknown` または明示 interface で扱うこと。

### `.claude/rules/testing.md` — Testing Rules
> **Bug fixes: Add regression test**

今回の不具合は save → getProfile → hydrate の多段フローなので、BE/FE の両方に回帰テストが必要。

### プロジェクト内参照実装

- `backend/internal/repository/line_customer_repository.go:55` — 同一 `line_user_id` を同じ顧客へ解決する実装
- `frontend/line-reserve/src/pages/CustomerInfoPage.tsx:53` — `pets` 復元ロジックの既存実装

## 優先度

**Medium** — データ欠損ではないが、リピーター予約の主要UXが壊れており、再予約時の離脱や二重入力を招くため。

## 関連チケット

- BUG-386: LINE予約フォームの飼い主名・ペット名・犬種が受付/カルテ予約詳細に反映されない
  - 優先度は BUG-386 が先。受付/カルテ連携の欠落が運用影響としてより大きい

## 関連ファイル

- `backend/internal/service/liff_service.go:268` — 追加プロフィール保存
- `backend/internal/handler/liff_handler.go:32` — LIFFプロフィール API
- `frontend/line-reserve/src/api/liff-api.ts:29` — プロフィール取得クライアント
- `frontend/line-reserve/src/pages/CustomerInfoPage.tsx:45` — 初期値復元
- `backend/internal/repository/line_customer_repository.go:55` — 顧客解決
- `docs/line/reservation-spec.md:1543` — 2回目以降の復元仕様
