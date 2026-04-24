# FE マスタ系 コード規約違反 完全スキャン

## 目的
FEマスタ関連コード（`features/master/`）が「API層・Routes層の責任分離・命名規則・パフォーマンス・デザイントークン準拠」の
規約に準拠しているかを体系的に検査するためのチェックリストだ。

下記チェックリスト × 対象ファイルリストの**全組み合わせ**を検査し、
PASS/FAIL を表で出力せよ。

**新パターンの発見・起票は禁止。チェックリストに定義された15パターンのみを報告する。**

---

## チェックリスト（固定・15パターン）

### ■ API 層 (FA: Frontend API)

#### FA1: transformXxx() 変換関数の存在（api/）
各 API ファイルに `transformXxx()` 関数が存在し、BackendModel（snake_case）→ フロントエンドドメイン型（camelCase）に変換しているか？
- 違反例: transform 関数なしで Backend 型をそのまま返す
- 正しい例: `function transformAnimalSpecies(data: ModelAnimalSpecies) { return { id: String(data.id ?? 0), name: data.name, isActive: data.is_active }; }`
- **注意**: `@/lib/transforms/xxx.ts` 等の外部ファイルに定義してインポートする形式も OK
- 対象: 上記「API」リストのファイル全件

#### FA2: ドメイン型の導出方法（api/）
ドメイン型が `ReturnType<typeof transformXxx>` で型推論されているか？（手書き interface 禁止）
- 違反例: `export interface AnimalSpecies { id: string; name: string; }`（手書き）
- 正しい例: `export type AnimalSpecies = ReturnType<typeof transformAnimalSpecies>;`
- **注意**: `@/lib/transforms/xxx.ts` 側で定義し `type XxxItem = ReturnType<typeof transformXxx>` でエクスポートする形式も OK
- 対象: 上記「API」リストのファイル全件

#### FA3: クエリキーの命名規則（api/）
`useQuery` のクエリキーが `["masters", "{resource-name}"]` パターンに従っているか？
- 違反例: `queryKey: ["animal-species"]`（`"masters"` プレフィックスなし）/ `queryKey: ["masterAnimalSpecies"]`
- 正しい例: `queryKey: ["masters", "animal-species"]`
- 対象: 上記「API」リストのファイル全件（`useQuery` 呼び出しのみ）

#### FA4: フック命名規則（api/）
Query/Mutation フックが以下の命名規則に従っているか？
- 一覧取得: `useGetXxx` または `useGetAllXxx`（`useXxx`・動詞省略は違反）
- 作成: `useCreateXxx`
- 更新: `useUpdateXxx`
- 削除: `useDeleteXxx`
- 並び替え: `useReorderXxx`
- 違反例: `useAnimalSpecies()`（動詞省略）/ `useFetchAnimalSpecies()`（`Fetch` 動詞）
- 対象: 上記「API」リストのファイル全件（export されるフック名）

#### FA5: onError での handleApiError（api/）
`useMutation` の `onError` コールバックで必ず `handleApiError(error, "コンテキスト")` が呼ばれているか？
- 違反例: `onError: (error) => console.error(error)` / `onError` なし
- 正しい例: `onError: (error) => handleApiError(error, "作成")`
- 対象: 上記「API」リストのファイル全件（全 `useMutation` の `onError`）

#### FA6: staleTime/gcTime 設定（api/）
一覧取得フック（`useGetXxx`/`useGetAllXxx`）に `staleTime: QUERY_STALE_TIMES.STATIC` と `gcTime: QUERY_GC_TIMES.LONG` が設定されているか？
- 違反例: `staleTime` 省略 / `staleTime: 1000 * 60 * 5`（ハードコード）
- 正しい例: `staleTime: QUERY_STALE_TIMES.STATIC, gcTime: QUERY_GC_TIMES.LONG`
- 対象: 上記「API」リストのファイル全件（`useQuery` を持つ一覧取得フックのみ）

#### FA7: リクエスト型の導出方法（api/）
`CreateXxxRequest` / `UpdateXxxRequest` が `models.ts` の型から `Omit`/`Partial` で導出されているか？手書き `interface` は禁止。
- 違反例: `export interface CreateAnimalSpeciesRequest { name: string; is_active: boolean; }`（手書き）
- 正しい例: `export type CreateAnimalSpeciesRequest = Omit<ModelAnimalSpecies, 'id' | 'created_at' | 'updated_at' | 'deleted_at'>`
- **注意**: 別ファイルで定義してインポートしている場合はその定義先も確認すること
- 対象: 上記「API」リストのファイル全件（`CreateXxxRequest`/`UpdateXxxRequest` が存在するファイルのみ）

---

### ■ Routes 層 (FR: Frontend Routes)

#### FR1: useMasterCRUD の使用（routes/）
`useMasterCRUD` フックを使用して CRUD 状態（editTarget, pendingDelete, filteredItems 等）を管理しているか？
独自の edit/delete state 管理（`useState<XxxItem | null>` を複数個並べる等）は違反。
- 違反例: `const [editItem, setEditItem] = useState<AnimalSpecies | null>(null); const [deleteItem, setDeleteItem] = useState<AnimalSpecies | null>(null);`
- 正しい例: `const crud = useMasterCRUD<AnimalSpecies>({ data, deleteMutation, entityLabel: "動物種類" });`
- **注意**: 独自 UI を持つページでも `useMasterCRUD` を使うのが正規パターン
- 対象: 上記「Routes」リストのファイル全件

#### FR2: useMasterSave の使用（routes/）
`useMasterSave` フックを使用して保存ロジック（create/update 振り分け・validation・close）を管理しているか？
独自の handleSave 関数を実装している場合は違反。
- 違反例: `const handleSave = async (data: FormData) => { if (editTarget) { await updateMutation.mutateAsync(...) } else { ... } }`
- 正しい例: `const { handleSave } = useMasterSave<AnimalSpecies, FormData, CreateReq, UpdateReq>({ crud, createMutation, updateMutation, validate, toCreateRequest, toUpdateRequest });`
- **注意**: 独自で `useTransition` + mutate を組み合わせている場合も違反
- 対象: 上記「Routes」リストのファイル全件

#### FR3: usePermission による権限チェック（routes/）
`usePermission(ResourceXxx)` を呼び出し、`canEdit` / `canDelete` などを UI に反映しているか？
- 違反例: `usePermission` なし（権限チェック未実装）
- 正しい例: `const { canEdit } = usePermission(ResourceMasterAnimalSpecies);`
- 対象: 上記「Routes」リストのファイル全件

#### FR4: SidePanel の memo() 適用（routes/）
routes 内でローカル定義している `SidePanel` サブコンポーネントが `memo()` で包まれているか？
- 違反例: `function SidePanel({ ... }) { return <MasterSidePanel ...>; }`（memo なし）
- 正しい例: `const SidePanel = memo(function SidePanel({ ... }) { return <MasterSidePanel ...>; });`
- **注意**: SidePanel コンポーネントが存在しない場合は `-`
- 対象: 上記「Routes」リストのファイル全件（ローカル SidePanel が存在するファイルのみ）

#### FR5: SidePanel の useState lazy initializer（routes/）
`SidePanel` 内の `useState` が `() =>` lazy initializer 形式を使っているか？
- 違反例: `const [formData, setFormData] = useState<FormData>({ name: item?.name ?? "" });`（直接値）
- 正しい例: `const [formData, setFormData] = useState<FormData>(() => ({ name: item?.name ?? "" }));`
- **注意**: SidePanel が存在しない場合は `-`
- 対象: 上記「Routes」リストのファイル全件（ローカル SidePanel が存在するファイルのみ）

---

### ■ 全体共通 (FG: Frontend General)

#### FG1: デザイントークン使用（routes/ + components/）
`C`, `STYLE`, `LAYOUT`, `ICON` 等のデザイントークン定数を使用し、Hex カラーや Tailwind のハードコードカラーを直接指定していないか？
- 違反例: `style={{ color: '#37352F' }}` / `className="text-gray-500 border-gray-200"`
- 正しい例: `style={{ color: C.TEXT_MAIN }}` / `className={cn(STYLE.FLEX_CENTER, C.text)}`
- **注意**: shadcn/ui コンポーネントの内部 className やサードパーティ由来の必須クラスは除外
- 対象: 上記「Routes」リストと「Components」リストの全件

#### FG2: 条件レンダーの三項演算子（routes/ + components/）
条件付きレンダリングが `condition ? (...) : null` を使っているか？（`&&` は禁止）
- 違反例: `{canEdit && <RowActionButton ... />}` / `{items.length && <List />}`
- 正しい例: `{canEdit ? <RowActionButton ... /> : null}`
- 対象: 上記「Routes」リストと「Components」リストの全件（全 JSX 中の条件レンダー）

#### FG3: any 型の不使用（api/ + routes/ + components/）
`any` 型を直接使用していないか？
- 違反例: `const data: any = response.data;` / `(e: any) => {}`
- 正しい例: `unknown` + 型ガード または 適切な型推論
- 対象: 上記「API」「Routes」「Components」リスト全件

---

## 対象ファイルリスト（全件）

### API（FA1〜FA7 を検査）
- frontend/src/features/master/api/animal-species.ts
- frontend/src/features/master/api/cages.ts
- frontend/src/features/master/api/checkup-types.ts
- frontend/src/features/master/api/chief-complaint-types.ts
- frontend/src/features/master/api/consultations.ts
- frontend/src/features/master/api/diagnosis.ts
- frontend/src/features/master/api/exam-types-master.ts
- frontend/src/features/master/api/hospitalization-plans.ts
- frontend/src/features/master/api/inquiry-templates.ts
- frontend/src/features/master/api/insurances.ts
- frontend/src/features/master/api/medicines.ts
- frontend/src/features/master/api/merchandise-items.ts
- frontend/src/features/master/api/occupations.ts
- frontend/src/features/master/api/permission-groups.ts
- frontend/src/features/master/api/procedures.ts
- frontend/src/features/master/api/reservation-type-groups.ts
- frontend/src/features/master/api/reservation-type-occupations.ts
- frontend/src/features/master/api/reservation-type-unavailable-times.ts
- frontend/src/features/master/api/reservation-types.ts
- frontend/src/features/master/api/staffs.ts
- frontend/src/features/master/api/trimming.ts
- frontend/src/features/master/api/vaccines-master.ts
- frontend/src/features/master/api/company.ts
- frontend/src/features/master/api/payment-method-master.ts
- frontend/src/features/master/api/create-master-item.ts
- frontend/src/features/master/api/delete-master-item.ts
- frontend/src/features/master/api/update-master-item.ts

### Routes（FR1〜FR5, FG1, FG2, FG3 を検査）
- frontend/src/features/master/routes/AnimalSpeciesSettings.tsx
- frontend/src/features/master/routes/CageSettings.tsx
- frontend/src/features/master/routes/ChiefComplaintSettings.tsx
- frontend/src/features/master/routes/DiagnosisSettings.tsx
- frontend/src/features/master/routes/HospitalizationSettings.tsx
- frontend/src/features/master/routes/InsuranceSettings.tsx
- frontend/src/features/master/routes/InterviewTemplateSettings.tsx
- frontend/src/features/master/routes/MasterSettingsIndex.tsx
- frontend/src/features/master/routes/MedicineSettings.tsx
- frontend/src/features/master/routes/MerchandiseItemSettings.tsx
- frontend/src/features/master/routes/OccupationSettings.tsx
- frontend/src/features/master/routes/PaymentMethodSettings.tsx
- frontend/src/features/master/routes/PermissionGroupSettings.tsx
- frontend/src/features/master/routes/ReservationTypeGroupSidePanel.tsx
- frontend/src/features/master/routes/ReservationTypeSettings.tsx
- frontend/src/features/master/routes/ReservationTypeSidePanel.tsx
- frontend/src/features/master/routes/StaffSettings.tsx
- frontend/src/features/master/routes/TreatmentPlanMaster.tsx
- frontend/src/features/master/routes/TrimmingSettings.tsx

### Components（FG1, FG2, FG3 を検査）
- frontend/src/features/master/components/MasterCRUDPage.tsx
- frontend/src/features/master/components/MasterListPage.tsx
- frontend/src/features/master/components/PermissionRuleTable.tsx
- frontend/src/features/master/components/ReservationTypeOccupationsSection.tsx
- frontend/src/features/master/components/ReservationTypeUnavailableTimesSection.tsx

---

## 実行方法（AgentTeam 推奨）

以下の3チームで並列実行せよ。各チームは担当ファイルのみを読む。

| チーム | 担当パターン | 担当ファイル |
|--------|------------|------------|
| Team-API | FA1, FA2, FA3, FA4, FA5, FA6, FA7 | 上記「API」リスト |
| Team-Routes | FR1, FR2, FR3, FR4, FR5, FG1, FG2, FG3 | 上記「Routes」リスト |
| Team-Components | FG1, FG2, FG3 | 上記「Components」リスト |

---

## 出力フォーマット（必須）

| ファイル | FA1 | FA2 | FA3 | FA4 | FA5 | FA6 | FA7 | FR1 | FR2 | FR3 | FR4 | FR5 | FG1 | FG2 | FG3 | 違反詳細 |
|---------|-----|-----|-----|-----|-----|-----|-----|-----|-----|-----|-----|-----|-----|-----|-----|---------|
| animal-species.ts | OK | OK | OK | OK | OK | OK | FAIL | - | - | - | - | - | - | - | OK | FA7:行28 CreateAnimalSpeciesRequest が手書き interface |
| AnimalSpeciesSettings.tsx | - | - | - | - | - | - | - | OK | OK | OK | OK | OK | OK | FAIL | OK | FG2:行133 `{canEdit &&` で && 使用 |

凡例:
- `OK` = 問題なし
- `FAIL` = 違反あり（違反詳細列にファイル名:行番号と内容を必ず記載）
- `-` = 該当パターンなし（このファイルに対象メソッド/構造が存在しない）

---

## 禁止事項（遵守必須）

1. **新パターンの発見・起票禁止** — FA1〜FA7, FR1〜FR5, FG1〜FG3 以外の問題を見つけても記録しない
2. **推測判定禁止** — 必ずファイルを Read してから判定する。コードを読まずに OK/FAIL を出力しない
3. **曖昧出力禁止** — 「〜かもしれない」「要確認」は使わない。`OK` か `FAIL` かのみ
4. **ファイル追加禁止** — 上記リスト外のファイルをスキャンしない
5. **スキャン中の即時起票禁止** — 全ファイルスキャン完了後に PASS/FAIL 表と違反サマリを出力してから起票する
6. **スキップ禁止** — ファイルリストの全件を読むこと

---

## 完了条件

1. 上記全ファイル × 全パターンの PASS/FAIL 表が出力される
2. FAIL セルの一覧をまとめた「違反サマリ」を出力する
3. `docs/tasks/open/code-quality/` と `docs/tasks/closed/code-quality/` の既存タスクタイトルと照合し、**未起票の違反のみ**を新規タスクとして `docs/tasks/open/code-quality/` に起票する（タスク番号は既存の最大番号+1から採番）
