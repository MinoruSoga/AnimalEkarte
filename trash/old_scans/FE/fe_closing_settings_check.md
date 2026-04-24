# FE 締め時間設定 コード規約違反 完全スキャン

## 目的
`features/closing-settings/` が「API層の命名規則・エラーハンドリング・型安全性・フォームパターン」の規約に
準拠しているかを体系的に検査するためのチェックリストだ。

**本 feature は設定ページ（CRUD マスタではない）のため、`useMasterCRUD`/`useMasterSave`/`MasterCRUDPage` を
前提とする FR パターン（FR1〜FR5）は適用しない。**

下記チェックリスト × 対象ファイルリストの**全組み合わせ**を検査し、
PASS/FAIL を表で出力せよ。

**新パターンの発見・起票は禁止。チェックリストに定義された8パターンのみを報告する。**

---

## チェックリスト（固定・8パターン）

### ■ API 層 (FA: Frontend API)

#### FA4: フック命名規則（api/）
Query/Mutation フックが以下の命名規則に従っているか？
- 一覧取得: `useGetXxx` または `useGetAllXxx`（`useXxx`・動詞省略は違反）
- 更新: `useUpdateXxx`
- 違反例: `useClosingSettings()`（動詞省略）
- 正しい例: `useGetClosingSettings()`, `useUpdateClosingSettings()`
- 対象: 上記「API」リストのファイル全件（export されるフック名）

#### FA5: onError での handleApiError（api/）
`useMutation` の `onError` コールバックで必ず `handleApiError(error, "コンテキスト")` が呼ばれているか？
- 違反例: `onError` なし（mutation がエラーを返してもユーザーに通知されない）
- 正しい例: `onError: (error) => handleApiError(error, "締め時間設定の更新")`
- 対象: 上記「API」リストのファイル全件（全 `useMutation` の `onError`）

#### FA6: staleTime/gcTime 設定（api/）
一覧・単件取得フック（`useGetXxx`）に `staleTime: QUERY_STALE_TIMES.STATIC` と `gcTime: QUERY_GC_TIMES.LONG` が設定されているか？設定データは変更頻度が低いため STATIC が適切。
- 違反例: `staleTime` 省略 / ハードコード
- 正しい例: `staleTime: QUERY_STALE_TIMES.STATIC, gcTime: QUERY_GC_TIMES.LONG`
- 対象: 上記「API」リストのファイル全件（`useQuery` を持つフックのみ）

#### FA7: リクエスト型の導出方法（api/）
`CreateXxxRequest` / `UpdateXxxRequest` が `models.ts` の型から `Omit`/`Partial` で導出されているか？手書き `interface` は禁止。
また、レスポンス型（`XxxResponse`）が `models.ts` の型を直接使っているか確認すること。
- 違反例1: `export interface UpdateClosingSettingsRequest { closing_am_pm_boundary?: string; ... }`（手書き）
- 違反例2: `export interface ClosingHoliday { id: number; clinic_id: number; ... }`（models.ts に対応型があるのに手書き）
- 正しい例: `export type UpdateClosingSettingsRequest = Partial<Pick<ClinicSettings, 'closing_am_pm_boundary' | ...>>`
- **注意**: `models.ts` に対応する型が存在しない場合は `-`（該当なし）とみなす
- 対象: 上記「API」リストのファイル全件（リクエスト/レスポンス型が存在するファイルのみ）

---

### ■ 全体共通 (FG: Frontend General)

#### FG1: デザイントークン使用（routes/ + components/）
`C`, `STYLE`, `LAYOUT`, `ICON` 等のデザイントークン定数を使用し、Hex カラーや Tailwind のハードコードカラーを直接指定していないか？
- 違反例: `style={{ color: '#37352F' }}` / `className="text-gray-500"`
- 正しい例: `style={{ color: C.TEXT_MAIN }}` / `className={cn(STYLE.FLEX_CENTER)}`
- **注意**: shadcn/ui 由来の必須クラスやサードパーティ由来のクラスは除外。自前で書いたスタイリングのみ対象。
- 対象: 上記「Routes」「Components」リストの全件

#### FG2: 条件レンダーの三項演算子（routes/ + components/）
条件付きレンダリングが `condition ? (...) : null` を使っているか？（`&&` は禁止）
- 違反例: `{isLoading && <LoadingFallback />}` / `{data && <Section data={data} />}`
- 正しい例: `{isLoading ? <LoadingFallback /> : null}`
- 対象: 上記「Routes」「Components」リストの全件（全 JSX 中の条件レンダー）

#### FG3: any 型の不使用（api/ + routes/ + components/）
`any` 型を直接使用していないか？
- 違反例: `const data: any = response.data;`
- 正しい例: `unknown` + 型ガード または 適切な型推論
- 対象: 上記「API」「Routes」「Components」リスト全件

#### FG4: useActionState + SubmitButton パターン（components/）
フォーム送信を伴うコンポーネントが `useActionState` + `<form action={...}>` + `SubmitButton` を使っているか？
（`useState` + `onSubmit` + 独自ローディング管理は禁止）
- 違反例: `const [isLoading, setIsLoading] = useState(false); const handleSubmit = async () => { setIsLoading(true); ... }`
- 正しい例: `const [, formAction, isPending] = useActionState(async (_prev, formData) => { ... }); return <form action={formAction}><SubmitButton>保存</SubmitButton></form>;`
- 対象: 上記「Components」リストのファイル全件（フォーム送信を行うコンポーネントのみ）

---

## 対象ファイルリスト（全件）

### API（FA4, FA5, FA6, FA7, FG3 を検査）
- frontend/src/features/closing-settings/api/get-closing-settings.ts
- frontend/src/features/closing-settings/api/update-closing-settings.ts
- frontend/src/features/closing-settings/api/special-periods.ts
- frontend/src/features/closing-settings/api/holidays.ts

### Routes（FG1, FG2, FG3 を検査）
- frontend/src/features/closing-settings/routes/ClosingSettingsPage.tsx

### Components（FG1, FG2, FG3, FG4 を検査）
- frontend/src/features/closing-settings/components/StandardClosingTimeSection.tsx
- frontend/src/features/closing-settings/components/SpecialPeriodSection.tsx
- frontend/src/features/closing-settings/components/HolidaySection.tsx

---

## 実行方法

ファイル数が少ないため単一エージェントで実行してよい。
全ファイルを Read してから判定すること。推測での OK/FAIL 出力は禁止。

---

## 出力フォーマット（必須）

| ファイル | FA4 | FA5 | FA6 | FA7 | FG1 | FG2 | FG3 | FG4 | 違反詳細 |
|---------|-----|-----|-----|-----|-----|-----|-----|-----|---------|
| get-closing-settings.ts | OK | - | FAIL | FAIL | - | - | OK | - | FA6:staleTime/gcTime なし / FA7:ClosingSettingsResponse が手書き interface |
| StandardClosingTimeSection.tsx | - | - | - | - | OK | OK | OK | OK | |

凡例:
- `OK` = 問題なし
- `FAIL` = 違反あり（違反詳細列にファイル名:行番号と内容を必ず記載）
- `-` = 該当パターンなし（このファイルに対象メソッド/構造が存在しない）

---

## 禁止事項（遵守必須）

1. **新パターンの発見・起票禁止** — FA4〜FA7, FG1〜FG4 以外の問題を見つけても記録しない
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
