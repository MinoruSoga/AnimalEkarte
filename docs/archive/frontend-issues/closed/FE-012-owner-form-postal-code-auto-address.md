# FE-012: 飼主フォーム — 郵便番号から住所を自動入力

**Status**: Open
**Priority**: Medium
**Affects**: owners feature — 飼主登録・編集フォーム
**Date Created**: 2026-03-17
**Related**: TASK-001

## Summary

飼主登録・編集フォームの郵便番号入力欄に、郵便番号から住所を自動入力する機能を追加する。現在は全て手入力。郵便番号API（zipcloud 等）を利用して、7桁入力時に都道府県・市区町村・町域を address1 に自動セットする。

## 現状のコード

### OwnerForm — 郵便番号入力欄

```typescript
// frontend/src/features/owners/routes/OwnerForm.tsx:246-254
<Label htmlFor="postalCode">郵便番号</Label>
<Input
  id="postalCode"
  placeholder="123-4567"
  value={ownerData.postalCode}
  onChange={(e) => handleInputChange("postalCode", e.target.value)}
/>

// 同:289-297（自宅郵便番号）
<Label>郵便番号(自宅)</Label>
<Input
  id="homePostalCode"
  placeholder="123-4567"
  value={ownerData.homePostalCode || ""}
  onChange={(e) => handleInputChange("homePostalCode", e.target.value)}
/>
```

### OwnerForm — 住所入力欄

```typescript
// frontend/src/features/owners/routes/OwnerForm.tsx:285-286（住所1）
// frontend/src/features/owners/routes/OwnerForm.tsx:336-337（住所2）
// frontend/src/features/owners/routes/OwnerForm.tsx:340-341（自宅住所1）
// frontend/src/features/owners/routes/OwnerForm.tsx:354-355（自宅住所2）
```

### useOwnerForm — フォーム状態

```typescript
// frontend/src/features/owners/hooks/useOwnerForm.ts:30,34,36-38
postalCode: "",
address1: "",
address2: "",
homeAddress1: "",
homeAddress2: "",
```

### package.json — 郵便番号ライブラリなし

```json
// frontend/package.json — 関連ライブラリ未導入
// zipcloud, ken-all, jp-zipcode 等なし
```

## 必要な変更

### 1. 郵便番号→住所変換の方針

外部APIを利用する。選択肢:

**A: zipcloud API（推奨 — 依存ライブラリ不要）**
```
GET https://zipcloud.ibsnet.co.jp/api/search?zipcode=1000001
```
- 無料、登録不要
- CORS 対応済み
- レスポンス: `{ results: [{ address1: "東京都", address2: "千代田区", address3: "千代田" }] }`

**B: npm ライブラリ（ken-all 等）**
- バンドルサイズ増加のため非推奨

### 2. 郵便番号検索 hook 作成

```typescript
// frontend/src/hooks/use-postal-code-lookup.ts（新規作成）

interface PostalCodeResult {
  prefecture: string;  // 都道府県
  city: string;        // 市区町村
  town: string;        // 町域
}

export function usePostalCodeLookup() {
  // 7桁の郵便番号を受け取り、住所を返す
  // debounce 付き（300ms）
  // エラー時は null を返す（トースト表示なし — 静かに失敗）
  const lookup = useCallback(async (postalCode: string): Promise<PostalCodeResult | null> => {
    const cleaned = postalCode.replace(/[-−ー]/g, "");
    if (cleaned.length !== 7 || !/^\d{7}$/.test(cleaned)) return null;

    const response = await fetch(
      `https://zipcloud.ibsnet.co.jp/api/search?zipcode=${cleaned}`
    );
    const data = await response.json();
    if (!data.results?.[0]) return null;

    const r = data.results[0];
    return {
      prefecture: r.address1,
      city: r.address2,
      town: r.address3,
    };
  }, []);

  return { lookup };
}
```

### 3. OwnerForm — 郵便番号入力欄の変更

```typescript
// frontend/src/features/owners/routes/OwnerForm.tsx
// 郵便番号入力欄の横に「検索」ボタン追加
// または 7桁入力完了時に自動で住所検索

// パターン A: 検索ボタン付き
<div className="flex gap-2">
  <Input
    id="postalCode"
    placeholder="123-4567"
    value={ownerData.postalCode}
    onChange={(e) => handleInputChange("postalCode", e.target.value)}
  />
  <Button
    type="button"
    variant="outline"
    size="sm"
    onClick={() => handlePostalCodeLookup("postalCode", "address1")}
  >
    検索
  </Button>
</div>

// パターン B: 自動入力（7桁入力で自動検索）
// useEffect で postalCode の変更を監視し、7桁になったら自動検索
// UI仕様は別途確認
```

### 4. useOwnerForm — 住所自動入力ロジック

```typescript
// frontend/src/features/owners/hooks/useOwnerForm.ts に追加:
// handlePostalCodeLookup(postalCodeField, addressField) を実装
// lookup 成功時に address1 に「都道府県 + 市区町村 + 町域」をセット
// 既に address1 に値がある場合は上書き確認不要（上書きする）
```

## UI 操作フロー

1. ユーザーが飼主登録/編集画面を開く
2. 郵便番号欄に「100-0001」と入力
3. 「検索」ボタンをクリック（or 自動検索）
4. address1 に「東京都千代田区千代田」が自動セットされる
5. ユーザーは address2 に番地・建物名を手入力
6. 自宅郵便番号でも同様に動作

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（`useState(false)` + `setIsPending` 禁止）
- [ ] 型は `models.ts` から導出（手書き interface 禁止）
- [ ] `useCallback` でハンドラ安定化（memo の前提条件）
- [ ] fetch は axios ではなく素の fetch でOK（外部API）

## 依存関係

- Backend 変更なし（postal_code, address1, address2 は既に API 対応済み）
- 外部API依存: zipcloud（https://zipcloud.ibsnet.co.jp/）
- UI仕様: ボタンか自動入力かは Figma デザインがない場合は「検索ボタン」で実装

## 完了条件

- [ ] `use-postal-code-lookup.ts` hook 新規作成
- [ ] 郵便番号（会社）→ address1 自動入力が動作
- [ ] 郵便番号（自宅）→ homeAddress1 自動入力が動作
- [ ] 無効な郵便番号で静かに失敗（エラー表示なし）
- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
