# FE-194: fieldset disabled + className="contents" でフィールド disabled が伝播しない（FE-166 修正実装の欠陥）

## 概要

FE-166 の「修正」として各フォームに `<fieldset disabled={!canSubmit} className="contents">` が実装されたが、
Tailwind の `contents` クラス（`display: contents`）と `disabled` 属性の組み合わせで
**フィールドへの disabled 伝播が機能していない**。

DOM レベルで `fieldset.disabled = true` は確認できるが、
内部の各 `input.disabled` は `false` のままになり、
`canEdit=false` / `canCreate=false` ユーザーが全フォームフィールドを自由に操作できる。

## 再現手順

1. `canEdit=false` ユーザーでログイン
2. `/inventory/1` など編集フォームページに直接アクセス
3. Chrome DevTools Console で確認:
   ```js
   document.querySelector('fieldset').disabled  // → true
   document.querySelector('input').disabled     // → false  ← 伝播していない
   ```

## 実測値

```json
{ "fieldsets": [{"disabled": true, "className": "contents"}],
  "firstInputDisabled": false }
```

`fieldset.disabled = true` にもかかわらず `input.disabled = false` → フィールド操作可能。

## 影響ファイル（9 ファイル）

| ファイル | フォーム |
|---------|---------|
| `features/inventory/routes/InventoryForm.tsx:350` | 在庫編集 |
| `features/vaccinations/routes/VaccinationForm.tsx:178` | 予防接種フォーム |
| `features/hospitalization/routes/HospitalizationForm.tsx:182` | 入院フォーム |
| `features/trimming/routes/TrimmingForm.tsx:570` | トリミングフォーム |
| `features/examinations/routes/ExaminationForm.tsx:351` | 定期健診フォーム |
| `features/accounting/routes/AccountingDetail.tsx:1091` | 会計詳細 |
| `features/owners/routes/OwnerForm.tsx:589` | 飼主フォーム |
| `features/owners/components/PetEditModal.tsx:276` | ペット編集モーダル |
| `features/hospital-settings/routes/ClinicMasterSettings.tsx:383` | 病院設定 |

## 根本原因

HTML 仕様では `<fieldset disabled>` は子要素に `disabled` を伝播させる。
しかし Tailwind `contents` クラス (`display: contents`) を CSS で適用すると、
Chromium では fieldset の Box が消滅するため `disabled` 状態の子要素への伝播が
期待通りに動作しない可能性がある。

```tsx
// 全フォームでこのパターン（機能していない）
<fieldset disabled={!canSubmit} className="contents">
  <Input name="name" />  {/* disabled が伝播しない ❌ */}
</fieldset>
```

## 問題の深刻度

- 保存ボタンは `{canSubmit ? <SubmitButton> : null}` で別途ガードされており非表示 ✅
- しかし `<form action={formAction}>` を使用しているため **Enter キーでのフォーム送信が可能** ❌
  - formAction 内に canEdit/canCreate チェックなし
  - API 呼び出しが実行される → バックエンドで 403 返却
- フィールドが見た目上操作可能 → 誤操作の UX 問題 ❌

## 修正方針

### Option A: `className="contents"` を削除（最もシンプル）

```tsx
// Before（機能していない）
<fieldset disabled={!canSubmit} className="contents">

// After（HTMLの標準動作を使う）
<fieldset disabled={!canSubmit}>
```

ただし `display: contents` を削除すると fieldset のデフォルトスタイル（border 等）が表示されるため、
以下のスタイルリセットが必要:

```tsx
<fieldset
  disabled={!canSubmit}
  className="border-0 p-0 m-0 min-w-0"
>
```

### Option B: formAction 内に権限チェックを追加（防御的プログラミング）

```tsx
const [formState, formAction] = useActionState(
  async (_prevState, formData) => {
    if (!canSubmit) return { success: false };  // 追加 ← Enter キー送信対策
    // ... 既存ロジック
  }
);
```

### Option C: 各フィールドに個別 disabled を追加

```tsx
<Input disabled={!canSubmit} name="name" />
```

**推奨: Option A + Option B の組み合わせ**
- `className="contents"` 削除 → 視覚的に disabled が反映される
- formAction に権限チェック追加 → Enter キー送信の防御

## 優先度

**CRITICAL** — 9 フォームで閲覧のみユーザーがフィールドを操作でき、
Enter キーで API 呼び出し（→ 403）が発生する。
FE-166 の修正が実際には機能していない。

## 発見日

2026-04-08（RBAC Phase 3 テスト中）

## 関連

- FE-166（この問題の修正として fieldset disabled を実装したが機能せず — closed 取り消しが必要）
