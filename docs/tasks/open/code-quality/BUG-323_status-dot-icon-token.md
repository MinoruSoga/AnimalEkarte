# BUG-323: Status Dot Icon Token の一部コンポーネント非使用

## 優先度
LOW

## 問題の説明

`src/lib/design-tokens.ts` で定義された `ICON.dotMd` (Status Dot icon token) が、一部のコンポーネント内で直接値指定されており、トークンが利用されていない。

## 期待動作

Status Dot アイコン表示では、常に `ICON.dotMd` トークンを使用：

```typescript
// ✅ 正しい
<span className={ICON.dotMd} style={{ color: C.SUCCESS }} />

// ❌ 間違い
<span className="w-2 h-2 rounded-full" style={{ color: C.SUCCESS }} />
```

## 修正方法

1. `design-tokens.ts` で定義済みトークンを確認：
```typescript
export const ICON = {
  dotMd: "w-2 h-2 rounded-full",
  // ...
};
```

2. 以下のパターンで直接値指定を検索：
```bash
grep -r "w-2 h-2" frontend/src --include="*.tsx" --include="*.ts" | grep "rounded-full"
grep -r "w-1.5 h-1.5" frontend/src --include="*.tsx" --include="*.ts" | grep "rounded-full"
```

3. 該当箇所を `ICON.dotMd` または新規トークン `ICON.dotSm` に置換

## テスト確認項目
- [ ] Status Dot が表示されるすべての箇所で `ICON.dotMd` が使用されている
- [ ] 直接値指定がない（grep で確認）
- [ ] UI 表示に変化なし

## 修正優先度
- LOW（機能的には正常動作、コンシステンシー向上のため）
- BUG-322 と同時実装推奨（デザインシステム整理タイミング）
