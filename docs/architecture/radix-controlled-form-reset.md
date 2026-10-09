# EMR-252 — Radix controlled コンポーネントの post-action form.reset() 巻き戻し: 修正方式と影響フォーム棚卸し

> **対象チケット**: EMR-252 / 作成: 2026-10-02 / 基準コミット: `fbc96c707` 派生 worktree
> **適用バージョン**: React 19.2.7（dev build の `requestFormReset` 経路）、@radix-ui/react-switch 1.3.7 / react-checkbox 1.3.11 / react-radio-group 1.4.7 / react-select 2.3.7

## 現象

React 19 で `<form action>`（`useActionState` の dispatch を含む）が完了すると、React は該当 `<form>` に対して `form.reset()` を同期的に発火する（`react-dom-client` 内 `startHostTransition → requestFormReset`）。Radix 各プリミティブは**マウント時点の値を初期ベースラインとして保持**しており、reset 時にそれへ復元する:

- `checkbox` / `switch`: `initialCheckedStateRef = useRef(checked)` + form `reset` リスナ → `setChecked(mount値)`。制御モードでは `onCheckedChange(mount値)` が発火して親 state を巻き戻す。
- `radio-group`: `initialValueRef` + reset リスナ → `setValue(mount値)`。
- `select`: `initialValueRef` + reset リスナ → `setValue(mount値)`。hidden `<select>` の `defaultValue` もマウント時値のため、`FormData` 経路も巻き戻る。

結果として「ユーザーが変更 → submit → action 完了」した直後にマウント時点の値へ戻る（例: 既存会計編集で mount 時 `hasInsurance=false` → 同期後 `true` にしたものが reset で `false` へ戻り OFF として送信）。

## 修正方式（共有ラッパで一元対応）

`frontend/src/components/ui/` の 4 ラッパの Root に**制御 prop 連動の `key`** を付けた。制御値が確定するたびに Radix root 以下が remount され、上記ベースライン（ref・hidden input の `default*`）が常に最新の確定値で再取得される。

- `switch.tsx`: `checked` → `key={String(checked)}`
- `checkbox.tsx`: `checked` → `key={String(checked)}`（`"indeterminate"` もそのまま文字列化）
- `radio-group.tsx`: `value` → `key={String(value)}`
- `select.tsx`: `value` → `key={String(value)}`（**`value` のみ**。`open` 等他 prop では key を付けない）

制御 prop が `undefined`（= 非制御 / `defaultChecked`・`defaultValue` のみ）の場合は key を付けず、Radix 標準の reset 復元（ネイティブセマンティクス）を維持する。

**新規コードのルール**（`frontend/CLAUDE.md` にも1行で記載）: 制御 Radix コンポーネントを `<form action>` 内で使う場合は共有 `ui/` ラッパ経由にする。やむを得ず生 Radix を直接使う場合は Radix root に `key={currentValue}` を付けること。

## 撤去済みの個別ワークアラウンド

- `features/accounting/components/InsuranceCard.tsx`: `Switch` の `key={useInsurance ? "on" : "off"}` と `Select` の `key={insuranceRatio}`（ラッパ側が担うため削除）。
- `features/trimming/components/TrimmingLeftColumn.tsx`: `optionToggleFromUserRef` による reset 由来 `onCheckedChange` 抑制（remount 後は reset が `onCheckedChange` を呼ばないため削除）。

## 残るリスクメモ: LstepSettingsForm のハザード

`LstepSettingsForm` では、reset 由来の `onCheckedChange(initial)` が `onDisableRequest()` を呼び「連携無効化」確認ダイアログを誤表示し得る経路があった。ラッパ修正により reset は `onCheckedChange` を発火しなくなり（setValue が最新値 === prop で早期 return）、このハザードは解消済み。今後 **生 Radix を form action 内に直置きする実装**では同型のハザードが再発し得るため、上記ルールの適用が必須。

## 影響フォーム棚卸し（18件）

`<form action>` / `useActionState` を使い、制御 Radix（Switch/Select/Checkbox/RadioGroup）を内包するフォーム。本修正によりラッパ経由で一律に解消。

| # | フォーム | 内包する制御 Radix |
|---|----------|-------------------|
| 1 | AccountingDetail（会計詳細） | InsuranceCard の Switch + Select、ItemListCard の Select 群 |
| 2 | RefundSection | Select |
| 3 | CreditCorrectionDialog | Select |
| 4 | CheckupForm | NextScheduleField 内 Select |
| 5 | ExaminationForm | status Select |
| 6 | EstimateForm | status Select |
| 7 | CarePlanTab EditRow | 行内 Select/Checkbox |
| 8 | CarePlanTab AddForm | Select/Checkbox |
| 9 | InventoryFormPanels | category Select |
| 10 | LineReservationSettingsForm | 制御 6 controls（Switch/Select 系） |
| 11 | MedicalRecordFormReadyPanels | VisitTypeSelect + RecommendationReasonSelect + NextScheduleField + 行 Checkbox |
| 12 | OwnerForm | OwnerBasicFields / Membership / AccountingHistory / LineIntegration の Switch 群 |
| 13 | LstepSettingsForm | Switch 群（上記ハザード対象） |
| 14 | TrimmingFormPanels | TrimmingLeftColumn の option Checkbox 群・initialStatus Select ほか |
| 15 | VaccinationFormPagePanels | NextScheduleField 内 Select ほか |
| 16 | DiagnosisNameSidePanel（MasterSidePanel 経由） | Select/Switch |
| 17 | TrimmingCourseSidePanel（MasterSidePanel 経由） | Select/Switch |
| 18 | ReservationTypeSidePanel（MasterSidePanel 経由） | Select/Switch |

## 検証

- 回帰テスト: `frontend/src/components/ui/{switch,checkbox,radio-group,select}.test.tsx` — 制御モードで「値変更 → action 完了 + reset → 値保持・`onXChange(mount値)` 不発」を、非制御で「reset 後 default 復元」をそれぞれ検証。Select は `FormData` 経路も検証。
- コマンド: `pnpm exec vitest run src/components/ui`（frontend コンテナ内）。
