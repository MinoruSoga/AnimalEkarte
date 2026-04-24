---
status: closed
closed_at: 2026-03-16
---

# [master] exam-types-master.ts と examination-types.ts で queryKey が不一致（mutation 後に UI が更新されないバグ）

## 優先度
高（バグ）

## 種別
バグ / API設計

## 対象ファイル
- `frontend/src/features/master/api/exam-types-master.ts`
- `frontend/src/features/master/api/examination-types.ts`
- `frontend/src/features/master/api/index.ts`

## 問題

同じエンドポイント `/v1/masters/examination-types` に対して 2 つの API ファイルが存在しており、
それぞれが異なる queryKey を使っているため、mutation 後にキャッシュが正しく無効化されない。

| ファイル | queryKey |
|---------|----------|
| `examination-types.ts` | `["examinationTypesMaster"]` |
| `exam-types-master.ts` | `["masterItems", "examination"]` |

`exam-types-master.ts` の `replaceExamTypeItems` の `onSuccess` で
`invalidateQueries({ queryKey: ["masterItems", "examination"] })` を呼んでも、
`examination-types.ts` の `["examinationTypesMaster"]` キャッシュは**無効化されない**。

## 結果

検査タイプの一括更新 (`replaceExamTypeItems`) を実行しても、一覧が古いデータを表示し続ける。

## 修正方針

1. `exam-types-master.ts` を `examination-types.ts` に統合する（items replace 操作をそちらへ移動）
2. queryKey を `["masters", "examination-types"]` に統一する
3. `index.ts` の `useGetMasterItemsByCategoryNew` alias export を削除する

## 付記

`exam-types-master.ts` の `replaceExamTypeItems` は raw `ExaminationType`（transform なし）を返しており、
統合時は `transformExaminationType` を経由するように修正すること。
