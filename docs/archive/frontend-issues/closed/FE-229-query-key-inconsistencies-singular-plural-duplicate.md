# FE-229: クエリキーの不一致（単数/複数形・重複ファイル）

## 概要

React Query のクエリキーに、同一リソースを指すが文字列表記が異なるケースが存在する。
キャッシュの invalidate が正しく機能しない恐れがある。

FE-212（`chief-complaints` vs `chief-complaint-categories`）と同種の問題。

## 問題1: `medical-record` vs `medical-records`（単数/複数形の不一致）

同じ medical-records feature 内で単数形と複数形が混在している。

| ファイル | 行 | クエリキー |
|---------|---|-----------|
| `frontend/src/features/medical-records/api/estimates.ts` | 18付近 | `["medical-record", ...]`（単数形） |
| `frontend/src/features/medical-records/api/inquiries.ts` | 16付近 | `["medical-record", ...]`（単数形） |
| `frontend/src/features/medical-records/api/` 他ファイル | — | `["medical-records", ...]`（複数形） |

**影響**: `estimates` や `inquiries` を invalidate しても、
`medical-records` キーを使う別 API のキャッシュに影響しない（または逆も然り）。

## 問題2: `examination-types` が2つのファイルに重複定義

| ファイル | 行 | 役割 |
|---------|---|------|
| `frontend/src/features/master/api/exam-types-master.ts` | 22 | `["masters", "examination-types"]` でクエリ定義 |
| `frontend/src/features/master/api/examination-types.ts` | 35 | `["masters", "examination-types"]` で invalidate のみ |

同一リソースが2つのファイルに分割されており、どちらが正規の定義か不明確。
片方が古い（使われなくなった）ファイルの可能性がある。

## 修正方針

### 問題1
`estimates.ts` と `inquiries.ts` のクエリキーを `medical-records`（複数形）に統一する。
または、`queryKeys` 定数オブジェクトを `src/lib/query-keys.ts` 等で一元管理する（FE-212 推奨事項と同様）。

### 問題2
`examination-types.ts` が不要なファイルであれば削除。
`exam-types-master.ts` に統合し、`queryKey` は1箇所で管理する。

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — 型安全性最優先
> 重複コードの排除。同一リソースのクエリキーは一元管理すること。

## 優先度
**Medium** — 単数/複数形の不一致はキャッシュ invalidate の漏れを引き起こす。
`examination-types` の重複は混乱の原因となりコード保守性を下げる。

## 関連ファイル
- `frontend/src/features/medical-records/api/estimates.ts`
- `frontend/src/features/medical-records/api/inquiries.ts`
- `frontend/src/features/master/api/exam-types-master.ts`
- `frontend/src/features/master/api/examination-types.ts`
- 関連: FE-212（`chief-complaints` vs `chief-complaint-categories` の同種問題）
