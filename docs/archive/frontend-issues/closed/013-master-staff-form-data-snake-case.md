# [master] StaffSettings: StaffFormData フィールドが snake_case（camelCase に統一せよ）

## 優先度
中

## 種別
コード品質 / 命名規則違反

## ステータス
status: closed
closed_at: 2026-03-16

## 対象ファイル
`frontend/src/features/master/routes/StaffSettings.tsx`

## 問題

### 1. フォームデータ型のフィールドが snake_case

`StaffFormData` 内のフィールドが `staff_role`・`license_number`・`is_active` と snake_case になっている。
フロントエンドのフォームデータ型は camelCase で統一すること（snake_case は API リクエスト送信時のみ使用）。

```ts
// 現状（違反）
interface StaffFormData {
  name: string;
  staff_role: string;     // ← snake_case（違反）
  license_number: string; // ← snake_case（違反）
  is_active: boolean;     // ← snake_case（違反）
  ...
}

// 期待
interface StaffFormData {
  name: string;
  staffRole: string;
  licenseNumber: string;
  isActive: boolean;
  ...
}
```

### 2. パスワードの長さバリデーションが未実装

UI のプレースホルダーに「8文字以上」と表示しているにも関わらず、
`handleSave` でのバリデーションは `!data.password`（空チェックのみ）。
8文字未満でもサブミット可能になっているため、セキュリティ上の問題がある。

```ts
// 現状
if (!data.password) { setError(...); return; }

// 修正後
if (!data.password || data.password.length < 8) {
  setError("パスワードは8文字以上で入力してください");
  return;
}
```

### 3. `startSaveTransition` の外でバリデーションを実行すべき

現状はバリデーションを `startSaveTransition` コールバック内で行っているが、
`useTransition` のセマンティクスとして、バリデーションは transition の**外**で行い、
通過した場合のみ transition を開始するのが正しいパターン。

## 修正方針

1. `StaffFormData` のフィールドを camelCase に変更
2. API リクエスト送信時に snake_case に変換するマッピングを追加（または `staffs.ts` の transform で処理）
3. パスワード長バリデーション（8文字以上）を追加
4. バリデーションを `startSaveTransition` の外に移動
