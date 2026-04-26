# BUG-333: 会計精算フォームの「丁度」「千円単位」「一万単位」ボタンが確認なしに会計を即時確定する

**Status**: CLOSED  
**Priority**: High  
**Discovery**: 機能テスト Section 6 会計管理 (2026-04-12)

## 概要

`/accounting/:id` の会計精算フォームで「丁度」「千円単位」「一万単位」プリセットボタンをクリックすると、お預かり金額のセットではなく**会計の即時確定**が発生する。原因は `<form action={formAction}>` 内のボタンに `type="button"` が設定されておらず、HTML デフォルトの `type="submit"` として扱われるため。

## 再現手順

1. `/accounting` 一覧で「会計待ち」状態の会計を開く
2. 会計精算フォームが表示される（「会計を確定する」ボタンが青）
3. **「丁度」ボタンをクリック**
4. **結果**: 「会計を確定しました」トーストが表示され、ボタンが「精算完了済み」に変化する（会計が確定される）
5. **期待**: お預かり金額が請求金額と同額にセットされるだけで、会計確定はされない

## 現状コード

### `frontend/src/features/accounting/routes/AccountingDetail.tsx:547-575`
```tsx
<Button
  variant="outline"
  size="sm"
  onClick={() => onReceivedAmountChange(billingAmount.toString())}
  // type="button" がない → デフォルト type="submit" → form を送信
>
  丁度
</Button>
<Button
  variant="outline"
  size="sm"
  onClick={() =>
    onReceivedAmountChange(
      (Math.ceil(billingAmount / 1000) * 1000).toString(),
    )
  }
  // type="button" がない → デフォルト type="submit"
>
  千円単位
</Button>
<Button
  variant="outline"
  size="sm"
  onClick={() =>
    onReceivedAmountChange(
      (Math.ceil(billingAmount / 10000) * 10000).toString(),
    )
  }
  // type="button" がない → デフォルト type="submit"
>
  一万単位
</Button>
```

### 根本原因
```tsx
// frontend/src/features/accounting/routes/AccountingDetail.tsx:1083
<form action={formAction}>  // ← React 19 Action
  ...
  <Button ...>丁度</Button>  // type 未指定 → type="submit" → formAction が実行される
```

HTML 仕様: `<form>` 内の `<button>` は `type` 属性未指定の場合 `type="submit"` として動作する。

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// frontend/src/features/owners/routes/OwnerForm.tsx — type="button" を明示
<Button
  type="button"
  variant="outline"
  onClick={handleAddPet}
>
  ペットを追加
</Button>
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/accounting/routes/AccountingDetail.tsx:547` | 「丁度」ボタン | 要修正 |
| `frontend/src/features/accounting/routes/AccountingDetail.tsx:554` | 「千円単位」ボタン | 要修正 |
| `frontend/src/features/accounting/routes/AccountingDetail.tsx:565` | 「一万単位」ボタン | 要修正 |
| 他フォーム内ボタン全て | `type="button"` 欠如の可能性 | 要確認 |

## 修正方針

### AccountingDetail.tsx の3ボタンに `type="button"` を追加

```tsx
// 丁度
<Button
  type="button"   // ← 追加
  variant="outline"
  size="sm"
  onClick={() => onReceivedAmountChange(billingAmount.toString())}
>
  丁度
</Button>

// 千円単位
<Button
  type="button"   // ← 追加
  variant="outline"
  size="sm"
  onClick={() =>
    onReceivedAmountChange(
      (Math.ceil(billingAmount / 1000) * 1000).toString(),
    )
  }
>
  千円単位
</Button>

// 一万単位
<Button
  type="button"   // ← 追加
  variant="outline"
  size="sm"
  onClick={() =>
    onReceivedAmountChange(
      (Math.ceil(billingAmount / 10000) * 10000).toString(),
    )
  }
>
  一万単位
</Button>
```

### 追加確認事項

`AccountingDetail.tsx` 内の `<form action={formAction}>` 内に存在する他の全ボタンについても `type="button"` が設定されているか確認し、未設定のものは全て追加する（診療明細書・領収書・物販追加ボタン等）。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — React 19 Action パターン
> **React 19 Action**: 原則 `useActionState` と `<form action={formAction}>` を使用

`<form action={...}>` を使う場合、フォーム内の全 `<button>` に `type="button"` を明示しないと、全ボタンがデフォルトで form を submit してしまう。これは React 19 Action パターン使用時の必須注意事項。

### プロジェクト内参照実装
- `frontend/src/features/owners/routes/OwnerForm.tsx` — フォーム内の非送信ボタンに `type="button"` を明示

## 優先度

**High** — プリセットボタンをクリックするだけで会計が確定されてしまい、間違った金額・明細での確定ミスが発生する。取消し手段（返金登録）はあるが、操作ミスのリスクが高い。

## 関連チケット
なし

## 関連ファイル
- `frontend/src/features/accounting/routes/AccountingDetail.tsx:547-575` — 修正対象（3ボタン）
- `frontend/src/components/ui/button.tsx:17-35` — Button コンポーネント（type デフォルトなし）
