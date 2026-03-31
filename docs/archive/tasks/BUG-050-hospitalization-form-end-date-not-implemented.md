# BUG-050: 入院フォームの終了日フィールドが未実装

## 概要
入院登録フォーム（`/hospitalization/new`）の「期間」セクションで、終了日（終了日）フィールドが UI 上は表示されているが、値の設定・保存が一切できない。

## 再現手順
1. `/hospitalization/new?petId=XX` を開く（ペット選択後）
2. 「期間」セクションの終了日カレンダーボタンをクリック
3. カレンダーから任意の日付を選択する
4. 終了日フィールドを確認する

## 期待動作
- 選択した日付が「終了日」フィールドに表示される
- 登録時に終了日がDBに保存される

## 実際の動作
- カレンダーを開いて日付を選択しても「終了日」フィールドは空のまま
- 値が設定されない（選択しても反映されない）

## 原因（コード）
`frontend/src/features/hospitalization/components/HospitalizationBasicInfo.tsx` の76〜81行目：

```tsx
<NotionDatePicker
  value=""        // ← 空文字ハードコード
  onChange={() => {}}  // ← no-op（何もしない）
  placeholder="終了日"
  className="flex-1"
/>
```

`value` と `onChange` が未実装のプレースホルダー状態。`formData` に `endDate` フィールドを追加し、正しく接続する必要がある。

## 影響範囲
- 入院登録フォームの終了日設定
- 入院期間の計算（日数・料金）にも影響する可能性

## 優先度
高（入院期間の終了日が保存できない）

## 発見日
2026-03-29

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| FE-145 | Frontend | `HospitalizationBasicInfo.tsx` の終了日 `value=""` / `onChange={() => {}}` をハードコードから実装に修正 |
