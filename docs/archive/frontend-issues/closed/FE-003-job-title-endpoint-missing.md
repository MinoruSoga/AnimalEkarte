# FE-003: 職種マスタエンドポイント未登録

## 重大度
**High** — 職種マスタの登録・編集・削除が機能しない（API リクエスト送信されない）

## 症状

- 職種マスタの新規登録フォームで「保存」をクリックしても、API リクエストが送信されない
- 編集・削除も同様にリクエスト送信されない
- ブラウザネットワークタブに POST/PATCH/DELETE リクエストが記録されない

## 根本原因

`frontend/src/features/master/api/get-master-items.ts` の `MASTER_CATEGORY_ENDPOINT` マッピングに `job_title` が登録されていない。

```ts
// ❌ 問題: job_title が存在しない
export const MASTER_CATEGORY_ENDPOINT: Record<string, string> = {
  examination: "/v1/masters/examination-types",
  // ... 他のカテゴリ
  checkup: "/v1/masters/checkup-types",
  // ← job_title がない
};
```

結果：
- `useGetMasterItemsByCategory("job_title")` が `endpoint = undefined` になる
- `useQuery` で `enabled: !!endpoint` が false になる
- API リクエストが送信されない

## 修正

```ts
// ✅ 修正: job_title を追加
export const MASTER_CATEGORY_ENDPOINT: Record<string, string> = {
  // ... 既存カテゴリ
  job_title: "/v1/masters/job-titles",  // ← 追加
};
```

## 確認

修正後、職種マスタの新規登録・編集が正常に機能することを確認：
- POST /api/v1/masters/job-titles → 201 Created ✅
- PATCH /api/v1/masters/job-titles/6 → 200 OK ✅
- DB に正しくデータが保存・更新される ✅

## ステータス

**CLOSED** — 修正完了（2026-03-16）

- `frontend/src/features/master/api/get-master-items.ts` に `job_title: "/v1/masters/job-titles"` を追加
- テスト完了: 職種マスタの登録・編集が正常に機能

## 関連

- チケット発見: 全マスタ設定ページテスト中（2026-03-16）
