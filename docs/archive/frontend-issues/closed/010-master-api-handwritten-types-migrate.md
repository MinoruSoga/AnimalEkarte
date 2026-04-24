# [master] staffs.ts / cages.ts の手書き型を models.ts 経由に移行せよ

## 優先度
中

## 種別
コード品質 / 規約違反

## 対象ファイル
- `frontend/src/features/master/api/staffs.ts`
- `frontend/src/features/master/api/cages.ts`
- `frontend/src/features/master/api/hospitalization-plans.ts`
- `frontend/src/features/master/api/trimming.ts`
- `frontend/src/features/master/api/company.ts`

## 問題

プロジェクト規約では「手書き interface 禁止、`models.ts` からの `Omit/Partial` 導出」が必須だが、
上記ファイルで `BackendXxx` 型と `CreateXxxRequest`・`UpdateXxxRequest` が手書きで定義されている。

Go モデルが変更された際に `make codegen` で `models.ts` は更新されるが、手書き型は更新されず乖離する。

### 具体的な違反箇所

**staffs.ts:**
```ts
// 手書き（削除すべき）
interface BackendStaff { ... }
interface Staff { ... }
interface CreateStaffRequest { ... }
interface UpdateStaffRequest { ... }
```

**cages.ts:**
```ts
// 手書き（削除すべき）
interface BackendCage { ... }
type Cage = { ... }
interface CreateCageRequest { ... }
interface UpdateCageRequest { ... }
```

**trimming.ts:**
- `BackendTrimmingCourse`、`BackendTrimmingOption` が手書き
- `TargetSize` 型が `types/index.ts` と二重定義

**hospitalization-plans.ts:**
- `HospitalizationPlan` が手書き interface（`ReturnType<typeof transformHospitalizationPlan>` で導出すべき）

## 修正方針

1. `make codegen` を実行し、最新の `models.ts` を確認する
2. `BackendXxx` 型を `models.ts` の対応する型に置き換える
3. `XxxRequest` 型を `models.ts` の型から `Omit/Partial/Pick` で導出する
4. `as` 型アサーション（`data.cage_type as Cage["cageType"]` 等）は `models.ts` 由来であれば不要になるため削除する

## 確認事項

`models.ts` に `Staff`・`Cage`・`TrimmingCourse`・`TrimmingOption`・`HospitalizationPlan` が生成されているか
`make codegen` 実行後に確認すること。
