/**
 * design-system-audit — docs/spec/ui-design-compliance.md §1 の機械判定。
 *
 * 判定内容:
 *   C1 — legacy 構造色（`C.accent` / 旧 accent `#2383E2`）禁止。brand/primary teal と臨床 semantic 色を許可。
 *   C3 — hex 直書き（文字列/テンプレートリテラル内）禁止。定義正本 `design-tokens.ts` / `shared-liff/brand-tokens.ts` は除外。
 *   C5 — CTA の `colorVariant` は semantic primary、brand、destructive、互換 default のみ。
 *   C6 — rgba()/rgb()/hsla()/hsl() の直値禁止（doc §1 の臨床安全 C6a とは ID 分離: 本スクリプトは C6b）。
 *   C7 — PageLayout maxWidth の生値禁止（`max-w-full` / `max-w-[Npx]`）。トークン経由のみ。
 *   C8 — `src/features/<feat>/routes/<file>.tsx` が PageLayout / Master*Page / allowlist のいずれか。
 *   C9 — `rounded(-[trbl]{1,2})?-[Npx]` 任意値禁止（トークン `rounded-xxs/xs/...` へ）。
 *   C10 — Tailwind 既定影と shadow 任意値禁止。DESIGN.md elevation トークンのみ許可。
 *   C11 — font-size 任意値禁止。
 *   C12 — DESIGN.md に存在しない font-size 段禁止。
 *   C13 — DESIGN.md ink 4段を迂回する黒アルファ禁止。
 *   C14 — typography role の letter-spacing を上書きする tracking utility 禁止（`*PrintArea*` は印刷透かしとして除外）。
 *   C15 — 本体 routes/pages で named white/black color の直接指定禁止。
 *   C16 — DESIGN.md spacing scale に存在しない 20px utility（`*-5`）禁止。
 *   C17 — CSS の直接 `box-shadow` / `filter: drop-shadow(...)` 禁止。
 *        値が `var(--*)` トークン参照のみの宣言（例: `box-shadow: var(--shadow-focus-primary);`）は
 *        elevation token 経由として許可。`--*` custom property 定義は宣言ではないため対象外。
 *   C18 — Table primitive / raw th・td 呼び出し側で typography / padding / 背景を非仕様値へ上書きすることを禁止。
 *   C19 — table row（DataTableRow / SortableDataTableRow / TableRow / tr）の行全体クリックを禁止。
 *   C20 — Tailwind 実行時合成（`hover:${` / `focus:${` / `text-[${`）禁止。完成形静的トークンのみ（FE-RC-013 / FE-RC-106）。
 *   C21 — `src/components/ui/*.tsx` に colocated `*.stories.tsx` を要求（design-token-layers.md §6）。
 *        Storybook カタログ網羅の担保。allowlist は現在空 — 新規 primitive は stories 必須。
 *   C24 — `src/components/shared/` 層にも stories を要求（ラチェット: 既存未カバー分は allowlist で棚卸し、削減方向のみ）。
 *   C22 — `PALETTE`（L1 raw値バケット）に Tailwind クラス文字列メンバを追加することを禁止
 *        （design-token-layers.md §3）。状態系は STATE、shadow 等は STYLE 側へ。
 *   C23 — `BADGE.*` コンボの text/bg コントラストが WCAG 4.5:1 以上であること
 *        （design-states.md §2.3）。`text-[#hex]`/`bg-[#hex]` 形式の C member を解決して計算。
 *   C25 — `truncate` / `text-ellipsis` / `line-clamp-N` を持つ opening tag に `title` / `aria-label` /
 *        `aria-labelledby` のいずれも無いことを禁止（design-patterns.md §3.1 の静的段。focus/タッチで
 *        開くツールチップの有無は静的に判定できないためレビュー項目）。既存分はファイル別の件数 baseline
 *        （C25_TRUNCATE_BASELINE）でラチェット: baseline を超えた追加だけを fail、削減時は baseline を下げる。
 *   C26 — `<nav>` ブロック内と `components/shared/Layout/Sidebar*.tsx` のナビ文言に `truncate` 禁止（§1.2）。
 *   C27 — `Z` / `Z_CLASS` トークン外の z-index 直書き（`z-N` / `focus:z-N` / `z-[N]` / `zIndex: N`）禁止（§5.1）。
 *        既存分はファイル別の件数 baseline（C27_Z_INDEX_BASELINE）でラチェット。
 *   C28 — `src/features/**` と `src/components/shared/**` で raw `<select>` と `<input type="date">` を
 *        業務・設定フォームに置くことを禁止（design-patterns.md §4.4）。検索・一覧フィルター・行内編集は
 *        C28 context allowlist で許容。shared 配下も見るのは、業務フォームが shared の部品経由で native 入力を
 *        描画するため（DynamicCheckupFields）。
 *   C29 — バリデーション理由の `disabled` / `submitDisabled`（`isValid` 系の名前、`.trim()`、`=== ""`、
 *        `.length` 比較）禁止（design-states.md §1.1・2026-10-20 決裁）。権限・送信中・ワークフロー状態
 *        （`canEdit` / `isPending` / `isFinalized`）は対象外。`FormHeaderActions` の `submitDisabled` prop は
 *        `SubmitButton disabled` に流れるため呼び出し側の式を検査する。対象は保存／送信ボタンの契約
 *        （submitDisabled・SubmitButton・PrimaryButton・type="submit"・「保存」を含む行）に限り、検索や
 *        一括操作の「選択 0 件で無効」は対象外。既存分は C29 allowlist でラチェット。
 *        注: 本リポジトリの `canSubmit` は RBAC 由来の命名のため C29 の対象にしない。
 *
 * C2/C4（PageLayout 使用）は C8 で routes 配下を機械化。新規リーフを allowlist に載せる場合は
 * C8_ALLOWLIST と docs/spec/ui-design-compliance.md §2 を同一コミットで更新する。
 *
 * strict fail（1件でも検出したら exit 1）。正当例外は allowlist 追記のみ。
 *
 * Exit: 0=OK / 1=違反 / 2=実行エラー
 * 実行: node scripts/design-system-audit.mjs [--cwd frontend]
 *       docker compose exec frontend pnpm design-audit
 */

import { readFileSync } from "node:fs";
import { readdir } from "node:fs/promises";
import path from "node:path";

const SCAN_ROOTS = ["src", path.join("liff", "src"), path.join("line-reserve", "src")];
const SCAN_EXTENSIONS = new Set([".ts", ".tsx"]);
const CSS_SCAN_EXTENSIONS = new Set([".css"]);
const EXCLUDE_DIR_NAMES = new Set(["node_modules", "dist", "generated"]);
const LEAF_DIR_NAMES = new Set(["routes", "pages"]);

// design-tokens.ts は本体色定数の定義正本のため C3/C6 の allowlist。
const DESIGN_TOKENS_REL_PATH = path.join("src", "lib", "design-tokens.ts");
// LIFF / LINE 予約の共有ブランドパレット定義正本（本体 design-tokens とは別系統）。
const BRAND_TOKENS_REL_PATH = path.join("src", "shared-liff", "brand-tokens.ts");
// 実行時に動的生成する rgba() のみ扱う（JSDoc で根拠を記載済み）ため C6 の allowlist。
const COLOR_MAP_REL_PATH = path.join("src", "hooks", "use-reservation-type-color-map.ts");

const C1_RE = /C\.accent\b|#2383E2/i;
const C3_RE = /['"`]#[0-9A-Fa-f]{3,8}['"`]|(?:(?:[a-z-]+):)*(?:bg|text|border|ring|outline|fill|stroke|decoration)-\[#[0-9A-Fa-f]{3,8}\](?:\/[0-9]+)?/;
const C5_RE = /colorVariant="([a-zA-Z]+)"/g;
const C5_ALLOWED_VALUES = new Set(["default", "primary", "brand", "destructive"]);
const C6_RE = /rgba?\(|hsla?\(/;
const C7_RE = /maxWidth=["']max-w-(full|\[[0-9]+px\])["']/;
const C9_RE = /rounded(?:-[trbl]{1,2})?-\[[0-9]+px\]/
const C10_RE = /(?:^|[^-\w])(?:shadow-(?:2xs|xs|sm|md|lg|xl|2xl)\b|shadow-\[|drop-shadow-(?:2xs|xs|sm|md|lg|xl|2xl)\b|drop-shadow-\[)/;
const C11_RE = /text-\[[0-9.]+(?:px|rem)\]/;
/** C12: DESIGN.md typography ロールに存在しないサイズ段（Tailwind 既定 18/24/30/36px 以上）の使用禁止（FE11-F4）。
 *  許可されるのは text-2xs/xs/sm/base/xl と text-heading-1/2/3 のみ。 */
const C12_RE = /(?:^|[^-\w])text-(?:lg|2xl|3xl|4xl|5xl|6xl|7xl|8xl|9xl)\b/;
/** C13: ink は DESIGN.md 4段のみ。黒アルファ（text-[#000000]/NN）の再導入禁止（FE11-F2）。 */
const C13_RE = /text-\[#000000\]\/[0-9]+|placeholder:text-\[rgba\(0,0,0/;
/** C14: role の letter-spacing を Tailwind built-in / arbitrary / shorthand で上書きしない。 */
const C14_RE = /(?:^|[^-\w])(?:-?tracking-(?:tighter|tight|normal|wide|wider|widest)\b|-?tracking-\[[^\]]+\]|-?tracking-\([^)]+\))/;
const C15_RE = /(?:^|[^-\w])(?:(?:[a-z-]+):)*(?:bg|text|border|ring|outline|fill|stroke|decoration|divide)-(?:white|black)(?:\/[0-9]+)?\b/;
const C16_RE = /(?:^|[^-\w])-?(?:[mp][trblxy]?|gap|space-[xy])-(?:5\b|\[(?:20px|1\.25rem)\])/;
/** C17: `box-shadow:` 宣言とその値（`;` / `}` まで）を取り出す。`var(--*)` のみの値は許可。 */
const C17_BOX_SHADOW_DECL_RE = /\bbox-shadow\s*:\s*([^;}]*)/g;
/** C17: `filter:` 宣言とその値（`;` / `}` まで）を取り出し、値内の drop-shadow() を検査する。 */
const C17_FILTER_DECL_RE = /\bfilter\s*:\s*([^;}]*)/g;
const C17_DROP_SHADOW_CALL_RE = /\bdrop-shadow\s*\(/g;
/** C17: トークン参照 `var(--ident)`（fallback は nested var / none / CSS-wide keyword のみ許容）。 */
const C17_VAR_TOKEN_RE = /var\(\s*--[\w-]+(?:\s*,\s*(?:var\(\s*--[\w-]+\s*\)|none|initial|inherit|unset|revert(?:-layer)?))?\s*\)/g;
/** C17: shadow 値を新設しない修飾子（inset / none / CSS-wide keyword / !important）。 */
const C17_SHADOW_KEYWORD_RE = /\b(?:inset|none|initial|inherit|unset|revert|revert-layer)\b|!\s*important\b/g;
const C18_TABLE_OPENING_TAG_START_RE = /<(?:Table(Cell|Head)|(td|th))\b/g;
const C18_TABLE_CELL_RE = /\b(?:text-(?:base|xs|2xs)|p-0)\b/;
const C18_TABLE_HEAD_RE = /\b(?:text-(?:sm|base|xs)|font-medium|p-0)\b/;
const C18_TABLE_CELL_STYLE_RE = /\bSTYLE\.tableHeaderCell\b/;
const C18_TABLE_HEAD_STYLE_RE = /\bSTYLE\.tableCell(?:Mono|Muted)?\b/;
const C18_TABLE_TYPE_SIZE_RE = /\btext-(?:2xs|xs|sm|base|lg|xl|2xl|3xl|4xl|5xl|6xl|7xl|8xl|9xl|heading-[123]|\[[^\]]+\])\b/g;
const C18_TABLE_FONT_WEIGHT_RE = /\bfont-(?:thin|extralight|light|normal|medium|semibold|bold|extrabold|black)\b/g;
const C18_TABLE_VERTICAL_PADDING_RE = /\b(?:p|py|pt|pb)-(?:px|[0-9]+(?:\.5)?|\[[^\]]+\]|\([^)]+\))/g;
const C18_TABLE_CELL_ALLOWED_WEIGHTS = new Set(["font-normal", "font-medium", "font-semibold"]);
/** C18(raw): 横paddingは DESIGN.md ex-data-table-cell の px-4（16px）のみ。primitive は基底styleがpx-4を保証するため raw 限定。 */
const C18_TABLE_HORIZONTAL_PADDING_RE = /\b(?:px|pl|pr|ps|pe)-(?:px|[0-9]+(?:\.5)?|\[[^\]]+\]|\([^)]+\))/g;
/** C18(raw): 背景は行側（STYLE.tableHeaderRow の bgPage 帯）が正本。セル単位の bg 再指定を禁止。 */
const C18_TABLE_RAW_CELL_BG_RE = /(?:^|[^-\w])bg-|(?:^|[^\w$])C\.bg[A-Z]/;
const C19_TABLE_ROW_OPENING_TAG_START_RE = /<(?:DataTableRow|SortableDataTableRow|TableRow|tr)\b/g;
/** C20: Tailwind v4 は静的走査のみ。`hover:${C.x}` 等の実行時合成は CSS に出ない（FE-RC-013 / FE-RC-106）。 */
const C20_RUNTIME_SYNTHESIS_RE = /hover:\$\{|focus:\$\{|text-\[\$\{/;

/**
 * C21: `src/components/ui/*.tsx` に colocated `*.stories.tsx` を要求する。
 * 2026-09-30 時点で全 ui primitive が stories を持つため許可リストは空。
 * 新規 ui primitive を追加したら同時に stories も書くこと。正当例外のみ追記可。
 */
export const C21_STORIES_ALLOWLIST = new Set();
const C21_UI_COMPONENT_RE = /^src[/\\]components[/\\]ui[/\\][a-z][a-z-]*\.tsx$/;

/**
 * C24: `src/components/shared/` 層の stories カバレッジ・ラチェット。
 * - `<Dir>/` 直下に `*.stories.tsx` が1件もないディレクトリ、または
 *   shared/ 直下のトップレベル `<name>.tsx` で sibling stories がないものを検出。
 * - 既存未カバー分はベースライン allowlist で棚卸し済み — 新規 shared 部品は
 *   stories 必須。allowlist は stories 追加時に該当エントリを削ること（増やさない）。
 */
export const C24_SHARED_STORIES_ALLOWLIST = new Set([
]);

/**
 * C22: PALETTE に残る Tailwind クラス文字列メンバ（STATE/STYLE への移行対象）。
 * L1 PALETTE は raw 値のみ許可 — 新規のクラス文字列メンバは hard fail。
 */
export const C22_PALETTE_CLASS_MEMBER_ALLOWLIST = new Set([
  "dragOverlayShadow",
  "brandGlow",
  "primaryGlow",
  "tableRowHover",
]);
const C22_PALETTE_BLOCK_RE = /export const PALETTE = \{([\s\S]*?)\} as const;/;
const C22_MEMBER_RE = /^\s*(\w+)\s*:\s*"((?:[^"\\]|\\.)*)"/;
const C22_VARIANT_PREFIX_RE = /^(?:hover|focus|focus-visible|focus-within|active|disabled|enabled|data-\[|aria-|group-|peer-|dark|sm|md|lg|xl|2xl|before|after|placeholder|selection|marker|file|backdrop|first|last|odd|even|visited|checked|empty):/;
const C22_UTILITY_PREFIX_RE = /^[a-z][a-z]*-[a-z0-9]/;

/** C25: 省略を起こす class。opening tag 単位で title / aria-label / aria-labelledby の有無を見る。 */
const C25_TRUNCATE_RE = /(?:^|[\s"'`{(])(?:truncate|text-ellipsis|line-clamp-[0-9]+)(?=$|[\s"'`}\)])/;
const C25_AFFORDANCE_RE = /(?:^|\s)(?:title|aria-label|aria-labelledby)\s*=/;
const C25_OPENING_TAG_START_RE = /<[A-Za-z][\w.]*\b/g;
/**
 * C25: 既存の truncate をファイル別件数で棚卸しした baseline（2026-10-20 時点 57 ファイル / 96 箇所）。
 * ラチェット運用 — baseline を超える追加は fail。全文到達手段を付けて減らしたら件数を下げる（増やさない）。
 * baseline に無いファイルは 1 件目から fail。
 */
export const C25_TRUNCATE_BASELINE = new Map([
  ["line-reserve/src/components/ListItem.tsx", 1],
  ["line-reserve/src/pages/TrimmingOptionSelectPage.tsx", 1],
  ["src/components/shared/DatePicker/DatePickerRange.tsx", 1],
  ["src/components/shared/Layout/SidebarChrome.tsx", 2],
  ["src/components/shared/Layout/SidebarItems.tsx", 2],
  ["src/components/shared/MasterSelectModal/MasterSelectTrigger.tsx", 1],
  ["src/components/shared/OwnerSearchModal/OwnerSearchModal.tsx", 1],
  ["src/components/shared/PastRecordHistoryPanel/PastRecordHistoryPanel.tsx", 2],
  ["src/components/shared/PatientContextHeader/PatientContextHeader.tsx", 3],
  ["src/components/shared/PatientInfoCard/PatientInfoCard.tsx", 6],
  ["src/components/shared/PropertyFilter/FilterRuleRow.tsx", 1],
  ["src/components/shared/PropertyFilter/InlineSelector.tsx", 2],
  ["src/components/shared/PropertyFilter/SortPill.tsx", 1],
  ["src/components/shared/PropertyFilter/SortPopover.tsx", 2],
  ["src/components/shared/ReservationFormModal/ReservationTypeAndStaffFields.tsx", 1],
  ["src/components/shared/SidePeek/PropertyRow.tsx", 1],
  ["src/components/ui/searchable-select.tsx", 1],
  ["src/features/accounting/components/RefundSection.tsx", 1],
  ["src/features/auth/components/LoginForm.tsx", 1],
  ["src/features/checkups/routes/CheckupsListPanels.tsx", 1],
  ["src/features/clinic-settings/components/ClinicMasterSidePanelProperties.tsx", 1],
  ["src/features/examinations/components/ExaminationCard.tsx", 2],
  ["src/features/examinations/routes/ExaminationsListPanels.tsx", 1],
  ["src/features/hospitalization/components/CarePlanTab/ItemRow.tsx", 2],
  ["src/features/hospitalization/components/HospitalizationBoard.tsx", 1],
  ["src/features/hospitalization/components/HospitalizationExpandedView.tsx", 1],
  ["src/features/hospitalization/components/HospitalizationUnassignedSection.tsx", 1],
  ["src/features/line-reservation/components/LinkedLineCustomers.tsx", 2],
  ["src/features/lstep/components/LstepCsvImportSection.tsx", 1],
  ["src/features/lstep/components/TagOwnerListDrawer.tsx", 2],
  ["src/features/manual/components/ManualContent.tsx", 2],
  ["src/features/manual/components/ManualEditor.tsx", 1],
  ["src/features/master/components/DiagnosisTabRows.tsx", 1],
  ["src/features/master/components/ReservationTypeGroupedTableRows.tsx", 1],
  ["src/features/master/components/ReservationTypeTree.tsx", 3],
  ["src/features/master/routes/MasterSettingsIndex.tsx", 1],
  ["src/features/medical-records/components/ExaminationImportDialog.tsx", 1],
  ["src/features/medical-records/components/ImageGalleryGroup.tsx", 2],
  ["src/features/medical-records/components/InterviewHistory.tsx", 2],
  ["src/features/medical-records/components/MedicalRecordStickyHeader.tsx", 3],
  ["src/features/medical-records/components/VaccinationHistory.tsx", 1],
  ["src/features/owner-report/components/ClinicalBriefingPanel.tsx", 1],
  ["src/features/owner-report/components/ClinicalHistoryMatrix.tsx", 1],
  ["src/features/owner-report/components/OwnerReportPanel.tsx", 2],
  ["src/features/owner-report/components/ReportPanel.tsx", 1],
  ["src/features/owners/components/LineSendHistory.tsx", 1],
  ["src/features/owners/components/OwnerPetsSection.tsx", 2],
  ["src/features/owners/components/PetSubOwnersSection.tsx", 2],
  ["src/features/reception/components/AppointmentCard.tsx", 5],
  ["src/features/reservations/components/MonthView.tsx", 3],
  ["src/features/reservations/components/WeekViewAppointmentCard.tsx", 8],
  ["src/features/reservations/components/WeekViewParts.tsx", 1],
  ["src/features/shifts/components/ShiftTemplateSidePanelFields.tsx", 1],
  ["src/features/support/components/ManualHelpTab.tsx", 1],
  ["src/features/support/routes/ChatHistoryPage.tsx", 1],
  ["src/features/trimming/components/TrimmingListTable.tsx", 1],
  ["src/features/vaccinations/components/VaccinationCard.tsx", 1],
]);
/** C26: nav ブロック。SidebarItems は別ファイルで `<nav>` を持たないため path でも判定する。 */
const C26_NAV_BLOCK_RE = /<nav\b[\s\S]*?<\/nav>/g;
const C26_SIDEBAR_FILE_RE = /^src[/\\]components[/\\]shared[/\\]Layout[/\\]Sidebar[^/\\]*\.tsx$/;
/** C26: 既存のナビ省略（STG 監査 H-2）。文言を短くするか省略を外したらエントリを削る。 */
export const C26_NAV_TRUNCATE_ALLOWLIST = new Set([
  // 2026-10-20 STG 監査 H-2（サイドバー項目名の末尾切れ）と manual の文書 TOC。
  "src/components/shared/Layout/SidebarChrome.tsx",
  "src/components/shared/Layout/SidebarItems.tsx",
  "src/features/manual/components/ManualContent.tsx",
]);
/** C27: Tailwind の z-N（variant 付き含む）・任意値 z-[N]・インライン style の数値 z-index。`Z.*` / `Z_CLASS.*` 参照は対象外。 */
const C27_Z_INDEX_RE = /(?:^|[\s"'`{(])(?:[a-z-]+:)*-?z-(?:[0-9]+|\[[0-9]+\])(?=$|[\s"'`}\)])|\bzIndex\s*:\s*[0-9]+/;
/**
 * C27: 既存の直書きをファイル別件数で棚卸しした baseline（2026-10-20 時点）。`Z` に段を追加して置換したら件数を下げる。
 * baseline に無いファイルは 1 件目から fail。
 */
export const C27_Z_INDEX_BASELINE = new Map([
  ["line-reserve/src/components/Calendar.tsx", 1],
  ["src/components/shared/DataTable/DataTable.tsx", 1],
  ["src/components/shared/PatientInfoCard/PatientInfoCard.tsx", 1],
  ["src/components/shared/PetSelection/PetSelectionResultsTable.tsx", 2],
  ["src/components/shared/ReservationFormModal/PatientSelectionResults.tsx", 1],
  ["src/components/ui/alert-dialog.tsx", 2],
  ["src/components/ui/calendar.tsx", 1],
  ["src/components/ui/dialog.tsx", 2],
  ["src/components/ui/dropdown-menu.tsx", 1],
  ["src/components/ui/popover.tsx", 1],
  ["src/components/ui/select.tsx", 1],
  ["src/components/ui/sheet.tsx", 2],
  ["src/components/ui/switch.tsx", 1],
  ["src/components/ui/toggle-group.tsx", 1],
  ["src/constants/medical-record-fixed-action-layout.ts", 1],
  ["src/features/accounting/components/AccountingItemRow.tsx", 2],
  ["src/features/accounting/components/ItemListCard.tsx", 3],
  ["src/features/auth/components/ClinicSelectionBlockedScreen.tsx", 1],
  ["src/features/hospitalization/components/HospitalizationExpandedView.tsx", 1],
  ["src/features/manual/components/ManualPageChrome.tsx", 3],
  ["src/features/master/components/CageSortableTable.tsx", 1],
  ["src/features/master/components/MedicineTable.tsx", 1],
  ["src/features/master/components/MerchandiseSortableTable.tsx", 1],
  ["src/features/medical-records/components/MedicalRecordImage.tsx", 1],
  ["src/features/medical-records/components/MedicalRecordStickyHeader.tsx", 1],
  ["src/features/owner-report/components/ClinicalHistoryMatrix.tsx", 3],
  ["src/features/owner-report/components/HistoryTable.tsx", 1],
  ["src/features/owner-report/routes/OwnerReport.tsx", 1],
  ["src/features/reservations/components/WeekViewAppointmentCard.tsx", 5],
  ["src/features/reservations/components/WeekViewDayColumn.tsx", 2],
  ["src/features/reservations/components/WeekViewParts.tsx", 4],
  ["src/features/reservations/lib/week-view-grid-constants.ts", 2],
  ["src/features/shifts/components/ShiftCalendar/ShiftCalendar.tsx", 3],
]);
/** C28: raw `<select` と `<input ... type="date"`（`<Select` 共有部品は対象外）。 */
const C28_NATIVE_OPENING_TAG_START_RE = /<(?:select|input)\b/g;
const C28_INPUT_DATE_RE = /\btype\s*=\s*["']date["']/;
/**
 * C28: design-patterns.md §4.4 が許容する文脈（検索・一覧フィルター・表の行内編集・時刻入力）の
 * ファイル。許容理由が §3.4 の列挙に無いものは入れない。
 */
export const C28_NATIVE_INPUT_CONTEXT_ALLOWLIST = new Set([
  // 検索・一覧フィルター（集計レポートの期間、レジ締め履歴、Lステップ連携・配信モニターの絞り込み）
  "src/features/accounting-reports/routes/AccountingReportsPagePanels.tsx",
  "src/features/cash-register/components/CashRegisterHistoryFilters.tsx",
  "src/features/lstep/components/CheckupSyncFilterForm.tsx",
  "src/features/lstep/components/LstepDeliveryStatsSection.tsx",
  // 単独ウィンドウ（飼主レポート）の表示対象ペット切替 — 一覧フィルターと同じ役割
  "src/features/owner-report/components/PetSwitcher.tsx",
  // 一覧・集計の期間フィルター部品（shared）
  "src/components/shared/DateRangeInputs/DateRangeInputs.tsx",
]);
/**
 * C28: 業務・設定フォームに残る既存の native 入力のファイル別件数（ラチェット）。
 * baseline を超える追加は既存ファイルでも違反。共有 `ui/` へ置換したら件数を下げる／エントリを削る。
 */
export const C28_NATIVE_INPUT_BASELINE = new Map([
  // 2026-10-20 STG 監査: 業務・設定フォームに残る native 入力。共有 DatePicker / Select へ置換したら削る。
  ["src/features/cash-register/components/CashRegisterClosePanels.tsx", 1],
  ["src/features/settings/components/LstepSettingsFields.tsx", 1],
  // 健診フォーム（features/checkups/CheckupFormPanels）が描画する shared の動的項目
  ["src/components/shared/DynamicCheckupFields/DynamicCheckupFields.tsx", 1],
]);
/**
 * C29: バリデーション結果を disabled / submitDisabled に流す式。妥当性の名前（isValid 系）、未入力判定
 * （`.trim()`、`=== ""` / `!== ""`、`.length` 比較、`errors`）を検出する。権限（can*）・送信中
 * （isPending / isSubmitting / isSaving）・ワークフロー状態（isFinalized 等）は含めない。
 */
const C29_SUBMIT_CONTEXT_RE = /\bsubmitDisabled\s*=|\bSubmitButton\b|\bPrimaryButton\b|type\s*=\s*["']submit["']|保存/;
const C29_VALIDATION_DISABLED_RE = /\b(?:disabled|submitDisabled)\s*=\s*\{[^}]*(?:\bis(?:Form)?Valid\b|\bformValid\b|\bisComplete\b|\bhasErrors\b|\berrors\s*\.\s*length\b|Object\s*\.\s*keys\s*\(\s*errors\s*\)|\.trim\s*\(\s*\)|[!=]==?\s*(?:""|''|``)|\.length\s*(?:[<>]=?|[!=]==?)\s*[0-9])/;
/**
 * C29: 既存のバリデーション disabled のファイル別件数（STG 監査で観測。ラチェット）。
 * baseline を超える追加は既存ファイルでも違反。押下→エラー表示へ直したら件数を下げる／エントリを削る。
 */
export const C29_VALIDATION_DISABLED_BASELINE = new Map([
  // 2026-10-20 STG 監査: 見積の保存が未入力理由で disabled（submitDisabled={!form.title.trim()}）
  ["src/features/estimates/routes/EstimateForm.tsx", 1],
  // レジ締めの確定が未入力理由で disabled（submitDisabled={actualCash === "" ...}）
  ["src/features/cash-register/routes/CashRegisterClosePage.tsx", 1],
  // シフトテンプレートの保存が名前未入力理由で disabled（disabled={isSaving || !formData.name.trim()}。複数行 prop）
  ["src/features/shifts/components/ShiftTemplateSettingsParts.tsx", 1],
  // ヘルプチャットの送信（type="submit"）が空メッセージ理由で disabled。保存フォームではないため §1.1 の適用可否は決裁待ち
  ["src/features/support/components/HelpChat.tsx", 1],
]);
/** C29: 走査する opening tag（JSX 要素すべて。`<` 直後の要素名）。 */
const C29_OPENING_TAG_START_RE = /<([A-Za-z][\w.]*)\b/g;

/** C23: BADGE コンボの WCAG コントラスト検査用パターン（design-tokens.ts 限定で呼ぶ）。 */
const C23_C_BLOCK_RE = /export const C = \{([\s\S]*?)\} as const;/;
const C23_BADGE_BLOCK_RE = /export const BADGE = \{([\s\S]*?)\} as const;/;
const C23_MEMBER_RE = /^\s*(\w+)\s*:\s*("(?:[^"\\]|\\.)*"|`(?:[^`\\]|\\.)*`)/;
const C23_HEX_RE = /#([0-9A-Fa-f]{6})/;
const C23_C_REF_RE = /\$\{C\.(\w+)\}/g;
const C23_MIN_RATIO = 4.5;

/** C8: 独自 shell を持つ正当な route page（相対パス完全一致）。 */
export const C8_PAGE_ALLOWLIST = new Set([
  path.join("src", "features", "auth", "routes", "Login.tsx"),
  path.join("src", "features", "auth", "routes", "ForgotPasswordPage.tsx"),
  path.join("src", "features", "auth", "routes", "ResetPasswordPage.tsx"),
  path.join("src", "features", "owner-report", "routes", "OwnerReport.tsx"),
  path.join("src", "features", "reception", "routes", "Reception.tsx"),
  path.join("src", "features", "manual", "routes", "ManualPage.tsx"),
  path.join("src", "features", "reservations", "routes", "ReservationManagement.tsx"),
  path.join("src", "features", "clinic-settings", "routes", "ClinicMasterSettings.tsx"),
  path.join("src", "features", "master", "routes", "LineReservationSlotsSettings.tsx"),
]);

/** C8: routes/ に置かれているが route page ではない helper、または 150 行分割で shell を sibling へ移した薄い route（相対パス完全一致）。 */
export const C8_ROUTE_HELPER_ALLOWLIST = new Set([
  path.join("src", "features", "trimming", "routes", "TrimmingLazyModals.tsx"),
  path.join("src", "features", "reception", "routes", "ReceptionLazyModals.tsx"),
  path.join("src", "features", "lstep", "routes", "LstepDeliveryMonitorPageParts.tsx"),
  path.join("src", "features", "lstep", "routes", "LstepDeliveryMonitorLogsTable.tsx"),
  path.join("src", "features", "accounting", "routes", "AccountingListPanels.tsx"),
  path.join("src", "features", "accounting-reports", "routes", "AccountingReportsPagePanels.tsx"),
  path.join("src", "features", "aggregation", "routes", "AggregationDashboardPage.tsx"),
  path.join("src", "features", "auth", "routes", "ResetPasswordPageSections.tsx"),
  path.join("src", "features", "checkups", "routes", "CheckupsListPanels.tsx"),
  path.join("src", "features", "checkups", "routes", "CheckupFormPanels.tsx"),
  path.join("src", "features", "estimates", "routes", "EstimateDetailPanels.tsx"),
  path.join("src", "features", "estimates", "routes", "EstimateListPanels.tsx"),
  path.join("src", "features", "examinations", "routes", "ExaminationsListPanels.tsx"),
  path.join("src", "features", "hospitalization", "routes", "HospitalizationForm.tsx"),
  path.join("src", "features", "hospitalization", "routes", "HospitalizationList.tsx"),
  path.join("src", "features", "inventory", "routes", "InventoryForm.tsx"),
  path.join("src", "features", "inventory", "routes", "InventoryListPanels.tsx"),
  path.join("src", "features", "lab-device", "routes", "LabDeviceBoardPanels.tsx"),
  path.join("src", "features", "master", "routes", "TreatmentPlanMaster.tsx"),
  path.join("src", "features", "medical-records", "routes", "MedicalRecordForm.tsx"),
  path.join("src", "features", "medical-records", "routes", "MedicalRecords.tsx"),
  path.join("src", "features", "reception", "routes", "ReceptionPagePanels.tsx"),
  path.join("src", "features", "reception", "routes", "use-reception-column-view.tsx"),
  path.join("src", "features", "shifts", "routes", "ShiftTemplateSettings.tsx"),
  path.join("src", "features", "trimming", "routes", "TrimmingForm.tsx"),
  path.join("src", "features", "trimming", "routes", "TrimmingList.tsx"),
  path.join("src", "features", "trimming", "routes", "TrimmingFormColumns.tsx"),
  path.join("src", "features", "hospitalization", "routes", "HospitalizationFormBody.tsx"),
  path.join("src", "features", "hospitalization", "routes", "HospitalizationFormFields.tsx"),
  path.join("src", "features", "hospitalization", "routes", "HospitalizationFormStatusView.tsx"),
  path.join("src", "features", "hospitalization", "routes", "HospitalizationFormHeaderExtra.tsx"),
  path.join("src", "features", "trimming", "routes", "TrimmingFormPanels.tsx"),
  path.join("src", "features", "vaccinations", "routes", "VaccinationForm.tsx"),
  path.join("src", "features", "vaccinations", "routes", "VaccinationListPanels.tsx"),
]);
export const C8_ALLOWLIST = new Set([
  ...C8_PAGE_ALLOWLIST,
  ...C8_ROUTE_HELPER_ALLOWLIST,
]);
const C8_SHELL_RE = /<PageLayout|Master(?:CRUD|Tab|List)Page/;

/**
 * C18: raw `<th>` / `<td>` 検査の対象外ファイル（相対パス完全一致・根拠必須）。
 * Table primitive（TableHead/TableCell）の検査はこの allowlist の影響を受けない。
 */
export const C18_RAW_CELL_ALLOWLIST = new Set([
  // Table primitive の実装正本 — TableHead/TableCell の中身が raw th/td。
  path.join("src", "components", "ui", "table.tsx"),
  // print 帳票 — 画面用 DESIGN.md ex-data-table-cell の適用対象外（*PrintArea*.tsx は basename パターンで除外）。
  path.join("src", "features", "accounting", "components", "AccountingDocument.tsx"),
  // print 帳票部品 — raw td は DailyAccountingPrintArea.tsx 専用の CatCell のみ。
  path.join("src", "features", "accounting", "components", "DailyAccountingTabParts.tsx"),
  // 操作マニュアル本文のドキュメント表 — 業務 data-table ではない。
  path.join("src", "features", "manual", "components", "ManualContent.tsx"),
  // owner-report 密度レイアウト群 — py-1.5 pr-3 はレポート密度目的の意図的判断（FE11 ⚠領域・最終裁定は統括判断）。
  path.join("src", "features", "owner-report", "components", "HistoryTable.tsx"),
  path.join("src", "features", "owner-report", "components", "TreatmentHistorySection.tsx"),
  path.join("src", "features", "owner-report", "components", "TrimmingHistorySection.tsx"),
  path.join("src", "features", "owner-report", "components", "VaccinationHistorySection.tsx"),
  // sticky 履歴マトリクス（縦=種類×横=日付）— 標準 data-table ではない特殊グリッド。
  path.join("src", "features", "owner-report", "components", "ClinicalHistoryMatrix.tsx"),
  // グループ行への nested DataTable 埋め込み構造セル（px-2 py-0）と空行 — 標準 cellPadding を当てると内側テーブルが崩れる。
  path.join("src", "features", "master", "components", "ReservationTypeGroupedTableRows.tsx"),
]);
/** C18: print 帳票（MonthlyReportPrintArea / DailyAccountingPrintArea / ClosePrintArea 等）の basename パターン。 */
const C18_RAW_CELL_PRINT_BASENAME_RE = /PrintArea/;

/**
 * isPrintSurfaceFile は印刷帳票コンポーネント（basename に PrintArea）かを判定する。
 * 画面用 typography / cell 規則（C14 / C18）の対象外。
 */
export function isPrintSurfaceFile(relPath) {
  return C18_RAW_CELL_PRINT_BASENAME_RE.test(path.basename(relPath));
}

/**
 * isC18RawCellExemptFile は raw th/td 検査の対象外ファイルか判定する。
 * allowlist・print 帳票・別アプリ（liff / line-reserve / shared-liff / MedicalRecordPrintView = FE11 除外）を除外する。
 */
export function isC18RawCellExemptFile(relPath) {
  if (C18_RAW_CELL_ALLOWLIST.has(relPath)) return true;
  if (isPrintSurfaceFile(relPath)) return true;
  return isFE11ExcludedPath(relPath);
}

function parseArgs(argv) {
  const args = { cwd: process.cwd() };
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "--cwd") args.cwd = argv[++i];
  }
  return args;
}

/**
 * isTestFile はファイル名が `*.test.*` パターンか判定する（C3/C6 除外用）。
 */
export function isTestFile(fileName) {
  return /\.test\./.test(fileName);
}

/**
 * isLeafRouteFile は相対パスが routes/ または pages/ ディレクトリ配下かを判定する
 * （`--glob '**\/routes/**' --glob '**\/pages/**'` 相当）。
 * FE3-2 以降 collectViolations 自体はこの判定で走査対象を絞らないが、
 * ユーティリティとして引き続き提供する。
 */
export function isLeafRouteFile(relPath) {
  const segments = relPath.split(path.sep);
  return segments.some((seg) => LEAF_DIR_NAMES.has(seg));
}

/**
 * isDesignTokensFile は色定数カタログ本体（C3/C6 allowlist）かを判定する。
 */
export function isDesignTokensFile(relPath) {
  return relPath === DESIGN_TOKENS_REL_PATH;
}

/**
 * isBrandTokensFile は LIFF/LINE 予約ブランドパレット本体（C3 allowlist）かを判定する。
 */
export function isBrandTokensFile(relPath) {
  return relPath === BRAND_TOKENS_REL_PATH;
}

/**
 * isColorMapFile は実行時動的生成 rgba() の唯一の許容ファイル（C6 allowlist）かを判定する。
 */
export function isColorMapFile(relPath) {
  return relPath === COLOR_MAP_REL_PATH;
}

/**
 * checkC1 はテキストから legacy accent 参照を行単位で検出する。
 */
export function checkC1(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C1_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * checkC3 はテキストから引用符で囲まれた hex 直書きリテラルを行単位で検出する。
 * 引用符無しの `#158` のような issue 番号コメントは対象外（誤検知しない）。
 */
export function checkC3(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C3_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * checkC5 はテキストから許可されていない `colorVariant="..."` を検出する。
 */
export function checkC5(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    const re = new RegExp(C5_RE.source, "g");
    let match;
    while ((match = re.exec(line)) !== null) {
      if (!C5_ALLOWED_VALUES.has(match[1])) {
        violations.push({ lineNumber: i + 1, text: line.trim(), value: match[1] });
      }
    }
  });
  return violations;
}

/**
 * checkC6 はテキストから rgba()/rgb()/hsla()/hsl() の直値を行単位で検出する（FE3-2 新設）。
 */
export function checkC6(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C6_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}


/**
 * checkC7 は maxWidth="max-w-full" / maxWidth="max-w-[Npx]" の生値を検出する（FE8-5）。
 */
export function checkC7(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C7_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * checkC8 は routes ファイル本文が PageLayout / Master*Page / allowlist のいずれかを満たすか判定する。
 * 違反時は 1 件の擬似 violation を返す（ファイル単位）。
 */
export function checkC8(relPath, text) {
  if (C8_ALLOWLIST.has(relPath)) return [];
  if (C8_SHELL_RE.test(text)) return [];
  return [{ lineNumber: 1, text: "PageLayout / Master*Page / C8 allowlist のいずれも無し" }];
}

/**
 * checkC9 は rounded-[Npx] / rounded-t-[Npx] 等の任意値角丸を検出する（FE8-5）。
 */
export function checkC9(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C9_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * checkC10 は Tailwind 既定影（shadow-2xs〜2xl）と shadow-[...] 任意値を検出する（FE9-2）。
 * 許可は design-system.md §5.1 のトークン（shadow-level1/level2/btn/panel/focus-brand）と shadow-none のみ。
 */
export function checkC10(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C10_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * checkC11 は text-[Npx|Nrem] の font-size 任意値を検出する（FE9-2）。
 * サイズは design-system.md §3.4 のロール（text-2xs〜text-2xl）経由のみ。
 */
export function checkC11(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C11_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * checkC12 は DESIGN.md typography に存在しないサイズ段の使用を検出する（FE11-F4）。
 * DESIGN.md のスケールは 12/14/15/16/20/22/26/40/54/64px。Tailwind 既定の
 * text-lg(18) / text-2xl(24) / text-3xl(30) / text-4xl(36) 等は非仕様値なので、
 * text-heading-1/2/3 または text-xl 以下のロールへ写像すること。
 */
export function checkC12(text, relPath = "") {
  // LINE ミニアプリ（liff / line-reserve / 共有 shared-liff）は FE10/FE11 の DESIGN.md スイープ
  // 対象外として明示宣言された別アプリ（対象=本体84ルート）。装飾絵文字に text-6xl 等を使うため除外する。
  // ※ 黙って握り潰さないための明示 allowlist。本体へ適用範囲を広げる場合はここを外すこと。
  if (isFE11ExcludedPath(relPath)) return [];
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C12_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * checkC13 は ink ランプの黒アルファ再導入を検出する（FE11-F2）。
 * DESIGN.md ink は #000000 / #31302E / #615D59 / #A39E98 の4段のみ。
 */
export function checkC13(text) {
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C13_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * checkC14 は typography role 固有の letter-spacing を上書きする tracking utility を検出する。
 */
export function checkC14(text, relPath = "") {
  // FE11 は本体84ルートが対象。別アプリは C12 と同じ明示 allowlist とする。
  // 印刷帳票の透かし letter-spacing は画面用 typography role の対象外（C18 と同じ *PrintArea*）。
  if (isFE11ExcludedPath(relPath) || isPrintSurfaceFile(relPath)) return [];
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C14_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * checkC15 は本体アプリの route/page surface に残る named white/black color を検出する。
 * 色の用途と出典を追跡できるよう、C / STYLE の semantic token 経由に統一する。
 */
export function checkC15(text, relPath = "") {
  if (isFE11ExcludedPath(relPath) || !isLeafRouteFile(relPath)) return [];
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C15_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * checkC16 は DESIGN.md spacing scale にない 20px (`*-5`) の spacing utility を検出する。
 * アイコン寸法などの size/w/h-5 は spacing ではないため対象外。
 */
export function checkC16(text, relPath = "") {
  if (isFE11ExcludedPath(relPath)) return [];
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C16_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * isC17CustomPropertyName は `box-shadow` / `filter` のマッチ位置が
 * custom property 名（`--box-shadow:` / `--my-filter:` 等の `--*` 定義）の一部かを判定する。
 * token 定義は宣言ではなく、値に生の shadow を保持するのが役目のため C17 の対象外。
 * vendor prefix（`-webkit-box-shadow:` 等）は `--` 始まりでないため宣言として検査される。
 */
function isC17CustomPropertyName(line, matchIndex) {
  let start = matchIndex;
  while (start > 0 && /[-\w]/.test(line[start - 1])) start -= 1;
  return line.startsWith("--", start);
}

/**
 * isC17TokenOnlyShadowValue は shadow 値が `var(--*)` トークン参照（カンマ区切りの複数レイヤを含む）と
 * `inset` / `none` / CSS-wide keyword / `!important` のみで構成されるかを判定する。
 * 生の長さ・色値を含む場合は false（= C17 違反）。
 */
function isC17TokenOnlyShadowValue(value) {
  const stripped = value
    .replace(C17_VAR_TOKEN_RE, "")
    .replace(C17_SHADOW_KEYWORD_RE, "");
  return /^[\s,]*$/.test(stripped);
}

/**
 * extractC17ParenArgs は openParenIndex の `(` から対応する `)` までの引数を切り出す。
 * 括弧が不整合なら null を返す（呼び出し側は違反として扱う）。
 */
function extractC17ParenArgs(text, openParenIndex) {
  let depth = 0;
  for (let i = openParenIndex; i < text.length; i++) {
    if (text[i] === "(") depth += 1;
    else if (text[i] === ")") {
      depth -= 1;
      if (depth === 0) return text.slice(openParenIndex + 1, i);
    }
  }
  return null;
}

/**
 * checkC17 は CSS から elevation token を迂回する直接 shadow 宣言を検出する。
 * `--shadow-*` 等の custom property 定義は宣言ではないため許可される。
 * `box-shadow` / `drop-shadow()` の値が `var(--*)` トークン参照のみで構成される宣言
 * （例: `box-shadow: var(--shadow-focus-primary);`）は elevation token 経由として許可し、
 * 生の長さ・色値を含む宣言のみを違反とする。
 */
export function checkC17(text, relPath = "") {
  if (isFE11ExcludedPath(relPath)) return [];
  const violations = [];
  text.split("\n").forEach((line, i) => {
    let flagged = false;
    const boxShadowRe = new RegExp(C17_BOX_SHADOW_DECL_RE.source, "g");
    let match;
    while (!flagged && (match = boxShadowRe.exec(line)) !== null) {
      if (isC17CustomPropertyName(line, match.index)) continue;
      if (!isC17TokenOnlyShadowValue(match[1])) flagged = true;
    }
    if (!flagged) {
      const filterRe = new RegExp(C17_FILTER_DECL_RE.source, "g");
      while (!flagged && (match = filterRe.exec(line)) !== null) {
        if (isC17CustomPropertyName(line, match.index)) continue;
        const dropShadowRe = new RegExp(C17_DROP_SHADOW_CALL_RE.source, "g");
        let dropMatch;
        while (!flagged && (dropMatch = dropShadowRe.exec(match[1])) !== null) {
          const args = extractC17ParenArgs(match[1], dropShadowRe.lastIndex - 1);
          if (args === null || !isC17TokenOnlyShadowValue(args)) flagged = true;
        }
      }
    }
    if (flagged) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

function extractC18ClassNameSegments(openingTag) {
  const segments = [];
  let braceDepth = 0;
  let quote = null;
  let escaped = false;
  for (let cursor = 0; cursor < openingTag.length; cursor += 1) {
    const char = openingTag[cursor];
    if (quote !== null) {
      if (escaped) escaped = false;
      else if (char === "\\") escaped = true;
      else if (char === quote) quote = null;
      continue;
    }
    if (char === '"' || char === "'" || char === "`") {
      quote = char;
      continue;
    }
    if (char === "{") {
      braceDepth += 1;
      continue;
    }
    if (char === "}") {
      braceDepth = Math.max(0, braceDepth - 1);
      continue;
    }
    if (
      braceDepth !== 0
      || !openingTag.startsWith("className", cursor)
      || /[\w$]/.test(openingTag[cursor - 1] ?? "")
    ) continue;

    let assignment = cursor + "className".length;
    while (/\s/.test(openingTag[assignment] ?? "")) assignment += 1;
    if (openingTag[assignment] !== "=") continue;

    let start = assignment + 1;
    while (/\s/.test(openingTag[start] ?? "")) start += 1;

    let end = start;
    const delimiter = openingTag[start];
    if (delimiter === '"' || delimiter === "'" || delimiter === "`") {
      let escaped = false;
      end += 1;
      for (; end < openingTag.length; end += 1) {
        const char = openingTag[end];
        if (escaped) escaped = false;
        else if (char === "\\") escaped = true;
        else if (char === delimiter) {
          end += 1;
          break;
        }
      }
    } else if (delimiter === "{") {
      let braceDepth = 0;
      let quote = null;
      let escaped = false;
      for (; end < openingTag.length; end += 1) {
        const char = openingTag[end];
        if (quote !== null) {
          if (escaped) escaped = false;
          else if (char === "\\") escaped = true;
          else if (char === quote) quote = null;
          continue;
        }
        if (char === '"' || char === "'" || char === "`") quote = char;
        else if (char === "{") braceDepth += 1;
        else if (char === "}" && --braceDepth === 0) {
          end += 1;
          break;
        }
      }
    } else {
      while (end < openingTag.length && !/[\s>]/.test(openingTag[end])) end += 1;
    }

    segments.push({
      text: openingTag.slice(start, end),
      lineOffset: openingTag.slice(0, start).split("\n").length - 1,
    });
    cursor = Math.max(end - 1, cursor);
  }
  return segments;
}

function hasC18UnprefixedClassToken(source, token) {
  const escaped = token.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  return new RegExp(
    "(?:^|[\\s\"'`{(])" + escaped + "(?=$|[\\s\"'`}\\)])",
  ).test(source);
}

function isC18IgnoredOpeningTag(text, index) {
  let mode = "code";
  let escaped = false;
  for (let cursor = 0; cursor < index; cursor += 1) {
    const char = text[cursor];
    const next = text[cursor + 1];

    if (mode === "line-comment") {
      if (char === "\n") mode = "code";
      continue;
    }
    if (mode === "block-comment") {
      if (char === "*" && next === "/") {
        mode = "code";
        cursor += 1;
      }
      continue;
    }
    if (mode !== "code") {
      if (escaped) escaped = false;
      else if (char === "\\") escaped = true;
      else if (
        (mode === "single-quote" && char === "'")
        || (mode === "double-quote" && char === '"')
        || (mode === "template" && char === "`")
      ) mode = "code";
      else if (char === "\n" && mode !== "template") mode = "code";
      continue;
    }

    if (char === "/" && next === "/") {
      mode = "line-comment";
      cursor += 1;
    } else if (char === "/" && next === "*") {
      mode = "block-comment";
      cursor += 1;
    } else if (char === "'") mode = "single-quote";
    else if (char === '"') mode = "double-quote";
    else if (char === "`") mode = "template";
  }
  return mode !== "code";
}

/**
 * checkC18 は Table primitive（TableHead/TableCell）と raw `<th>` / `<td>` の opening tag を調べ、
 * DESIGN.md ex-data-table-cell と `components/ui/table.tsx` の標準 typography / padding を
 * 上書きする class を検出する（th→Head 規則・td→Cell 規則）。raw cell は基底 style を継承しない
 * ため、STYLE.tableHeaderCell / STYLE.tableCell(Mono|Muted) 経由か仕様値 literal を正の準拠経路とし、
 * 加えて横padding（px-4 以外）と背景のセル単位再指定も禁止する。
 * `data-empty-state` + `colSpan` + `text-center` の空 state 追加余白、明示的な
 * `data-c18-structural-cell` を持つself-closing構造セル、cell 内の子要素、別 component、
 * C18_RAW_CELL_ALLOWLIST（raw のみ）、test / 非 TSX は対象外。
 */
export function checkC18(text, relPath = "") {
  if (path.extname(relPath) !== ".tsx" || isTestFile(path.basename(relPath))) return [];

  const rawCellExempt = isC18RawCellExemptFile(relPath);
  const violations = [];
  const openingTagRe = new RegExp(C18_TABLE_OPENING_TAG_START_RE.source, "g");
  let match;
  while ((match = openingTagRe.exec(text)) !== null) {
    if (isC18IgnoredOpeningTag(text, match.index)) continue;
    const isRawCell = match[1] === undefined;
    const kind = match[1] ?? (match[2] === "th" ? "Head" : "Cell");
    if (isRawCell && rawCellExempt) continue;
    let braceDepth = 0;
    let quote = null;
    let escaped = false;
    let endIndex = text.length;

    for (let index = openingTagRe.lastIndex; index < text.length; index += 1) {
      const char = text[index];
      if (quote !== null) {
        if (escaped) {
          escaped = false;
        } else if (char === "\\") {
          escaped = true;
        } else if (char === quote) {
          quote = null;
        }
        continue;
      }
      if (char === '"' || char === "'" || char === "`") {
        quote = char;
      } else if (char === "{") {
        braceDepth += 1;
      } else if (char === "}") {
        braceDepth = Math.max(0, braceDepth - 1);
      } else if (char === ">" && braceDepth === 0) {
        endIndex = index + 1;
        break;
      }
    }

    const openingTag = text.slice(match.index, endIndex);
    // 内容を持たない構造セルは明示marker付きself-closing要素だけ基底style必須から除外する。
    // markerがあっても背景・typography・padding overrideの検査は継続する。
    const isStructuralCell = (
      isRawCell
      && /\/\s*>$/.test(openingTag)
      && /\bdata-c18-structural-cell\b/.test(openingTag)
    );
    const classNameSegments = extractC18ClassNameSegments(openingTag);
    const classNameSource = classNameSegments.map((segment) => segment.text).join(" ");
    const forbiddenRe = kind === "Head" ? C18_TABLE_HEAD_RE : C18_TABLE_CELL_RE;
    const forbiddenStyleRe = kind === "Head" ? C18_TABLE_HEAD_STYLE_RE : C18_TABLE_CELL_STYLE_RE;
    const isEmptyState = kind === "Cell"
      && /\bdata-empty-state\b/.test(openingTag)
      && /\bcolSpan\s*=/.test(openingTag)
      && /\btext-center\b/.test(classNameSource);
    const firstLineNumber = text.slice(0, match.index).split("\n").length;
    const hasRawStyleToken = kind === "Head"
      ? /\bSTYLE\.tableHeaderCell\b/.test(classNameSource)
      : /\bSTYLE\.tableCell(?:Mono|Muted)?\b/.test(classNameSource);
    const hasRawLiteralContract = kind === "Head"
      ? ["text-2xs", "font-semibold", "px-4", "py-3"]
        .every((token) => hasC18UnprefixedClassToken(classNameSource, token))
      : ["text-sm", "font-normal", "px-4", "py-3"]
        .every((token) => hasC18UnprefixedClassToken(classNameSource, token));
    const hasRawEmptyStyle = kind === "Cell" && /\bSTYLE\.tableEmpty(?:Sm)?\b/.test(classNameSource);
    if (
      isRawCell
      && !isEmptyState
      && !isStructuralCell
      && !hasRawEmptyStyle
      && !hasRawStyleToken
      && !hasRawLiteralContract
    ) {
      violations.push({
        lineNumber: firstLineNumber,
        text: openingTag.split("\n")[0].trim(),
      });
      openingTagRe.lastIndex = endIndex;
      continue;
    }
    classNameSegments.forEach((segment) => {
      segment.text.split("\n").forEach((line, index) => {
        const expectedTypeSize = kind === "Head" ? "text-2xs" : "text-sm";
        const hasWrongTypeSize = [...line.matchAll(C18_TABLE_TYPE_SIZE_RE)]
          .some(([token]) => token !== expectedTypeSize);
        const hasWrongFontWeight = [...line.matchAll(C18_TABLE_FONT_WEIGHT_RE)].some(([token]) => (
          kind === "Head" ? token !== "font-semibold" : !C18_TABLE_CELL_ALLOWED_WEIGHTS.has(token)
        ));
        const hasNonstandardVerticalPadding = [...line.matchAll(C18_TABLE_VERTICAL_PADDING_RE)]
          .some(([token]) => token !== "py-3" && !(isEmptyState && token.startsWith("py-")));
        const hasNonstandardHorizontalPadding = isRawCell
          && [...line.matchAll(C18_TABLE_HORIZONTAL_PADDING_RE)].some(([token]) => (
            token !== "px-4" && !(isStructuralCell && token === "px-0")
          ));
        const hasCellBackground = isRawCell && C18_TABLE_RAW_CELL_BG_RE.test(line);
        if (
          forbiddenRe.test(line)
          || forbiddenStyleRe.test(line)
          || hasWrongTypeSize
          || hasWrongFontWeight
          || hasNonstandardVerticalPadding
          || hasNonstandardHorizontalPadding
          || hasCellBackground
        ) {
          violations.push({
            lineNumber: firstLineNumber + segment.lineOffset + index,
            text: line.trim(),
          });
        }
      });
    });
    openingTagRe.lastIndex = endIndex;
  }
  return violations;
}

/**
 * checkC19 は table row の opening tag を調べ、行全体へ onClick を付ける実装を検出する。
 * 詳細遷移は cell 内の native link、編集・表示は native button を使う。test / 非 TSX は対象外。
 */
export function checkC19(text, relPath = "") {
  if (path.extname(relPath) !== ".tsx" || isTestFile(path.basename(relPath))) return [];

  const violations = [];
  const openingTagRe = new RegExp(C19_TABLE_ROW_OPENING_TAG_START_RE.source, "g");
  let match;
  while ((match = openingTagRe.exec(text)) !== null) {
    if (isC18IgnoredOpeningTag(text, match.index)) continue;

    let braceDepth = 0;
    let quote = null;
    let escaped = false;
    let endIndex = text.length;
    let onClickIndex = -1;

    for (let index = openingTagRe.lastIndex; index < text.length; index += 1) {
      const char = text[index];
      if (quote !== null) {
        if (escaped) escaped = false;
        else if (char === "\\") escaped = true;
        else if (char === quote) quote = null;
        continue;
      }
      if (char === '"' || char === "'" || char === "`") {
        quote = char;
      } else if (char === "{") {
        braceDepth += 1;
      } else if (char === "}") {
        braceDepth = Math.max(0, braceDepth - 1);
      } else if (char === ">" && braceDepth === 0) {
        endIndex = index + 1;
        break;
      } else if (
        braceDepth === 0
        && text.startsWith("onClick", index)
        && /\s/.test(text[index - 1] ?? "")
        && !/[\w$]/.test(text[index + "onClick".length] ?? "")
      ) {
        let assignment = index + "onClick".length;
        while (/\s/.test(text[assignment] ?? "")) assignment += 1;
        if (text[assignment] === "=") onClickIndex = index;
      }
    }

    if (onClickIndex >= 0) {
      const lineStart = text.lastIndexOf("\n", onClickIndex - 1) + 1;
      const nextLine = text.indexOf("\n", onClickIndex);
      const lineEnd = nextLine === -1 ? text.length : nextLine;
      violations.push({
        lineNumber: text.slice(0, onClickIndex).split("\n").length,
        text: text.slice(lineStart, lineEnd).trim(),
      });
    }
    openingTagRe.lastIndex = endIndex;
  }
  return violations;
}

/**
 * readOpeningTag は `<Tag` の直後から、brace / quote を考慮して opening tag の終端 `>` を探し、
 * tag 全文と終端 index を返す（C19 と同じ走査規則）。
 */
function readOpeningTag(text, startIndex, afterNameIndex) {
  let braceDepth = 0;
  let quote = null;
  let escaped = false;
  let endIndex = text.length;
  for (let index = afterNameIndex; index < text.length; index += 1) {
    const char = text[index];
    if (quote !== null) {
      if (escaped) escaped = false;
      else if (char === "\\") escaped = true;
      else if (char === quote) quote = null;
      continue;
    }
    if (char === '"' || char === "'" || char === "`") quote = char;
    else if (char === "{") braceDepth += 1;
    else if (char === "}") braceDepth = Math.max(0, braceDepth - 1);
    else if (char === ">" && braceDepth === 0) {
      endIndex = index + 1;
      break;
    }
  }
  return { tag: text.slice(startIndex, endIndex), endIndex };
}

function lineNumberAt(text, index) {
  return text.slice(0, index).split("\n").length;
}

function firstLineOf(tag) {
  return tag.split("\n")[0].trim();
}

/**
 * checkC25 は省略（truncate / text-ellipsis / line-clamp-N）を持つ opening tag を調べ、
 * title / aria-label / aria-labelledby のいずれも無いものを検出する（design-patterns.md §3.1）。
 * ツールチップ／展開の有無は静的に判定できないため、本検査は「全文がどこにも無い」最低線だけを機械化する。
 */
export function checkC25(text, relPath = "") {
  if (path.extname(relPath) !== ".tsx" || isTestFile(path.basename(relPath))) return [];

  const violations = [];
  const openingTagRe = new RegExp(C25_OPENING_TAG_START_RE.source, "g");
  let match;
  while ((match = openingTagRe.exec(text)) !== null) {
    if (isC18IgnoredOpeningTag(text, match.index)) continue;
    const { tag, endIndex } = readOpeningTag(text, match.index, openingTagRe.lastIndex);
    if (C25_TRUNCATE_RE.test(tag) && !C25_AFFORDANCE_RE.test(tag)) {
      violations.push({ lineNumber: lineNumberAt(text, match.index), text: firstLineOf(tag) });
    }
    openingTagRe.lastIndex = Math.max(endIndex, openingTagRe.lastIndex);
  }
  return applyCountBaseline(violations, C25_TRUNCATE_BASELINE.get(relPath), "C25_TRUNCATE_BASELINE");
}

/**
 * applyCountBaseline は件数ラチェット: baseline 件数以内なら 0 件、超過分だけを「baseline 超過」として返す。
 * baseline が無いファイルは全件返す。
 */
function applyCountBaseline(violations, baseline, baselineName) {
  if (baseline === undefined) return violations;
  if (violations.length <= baseline) return [];
  return violations.slice(baseline).map((v) => ({
    ...v,
    text: `${v.text}（${baselineName} ${baseline} 件を超過: ${violations.length} 件）`,
  }));
}

/**
 * checkC26 はナビ文言の省略を検出する（design-patterns.md §1.2）。`<nav>…</nav>` ブロック内、
 * および `components/shared/Layout/Sidebar*.tsx` 全体を対象にする。
 */
export function checkC26(text, relPath = "") {
  if (path.extname(relPath) !== ".tsx" || isTestFile(path.basename(relPath))) return [];
  if (C26_NAV_TRUNCATE_ALLOWLIST.has(relPath)) return [];

  const code = stripTsCommentsPreservingLines(text);
  const ranges = [];
  if (C26_SIDEBAR_FILE_RE.test(relPath)) {
    ranges.push([0, code.length]);
  } else {
    const navRe = new RegExp(C26_NAV_BLOCK_RE.source, "g");
    let match;
    while ((match = navRe.exec(code)) !== null) ranges.push([match.index, match.index + match[0].length]);
  }
  if (ranges.length === 0) return [];

  const violations = [];
  let offset = 0;
  code.split("\n").forEach((line, i) => {
    // 行の区間 [offset, lineEnd) が nav 区間と交差するか（行の途中から始まる <nav> も含める）。
    const lineEnd = offset + line.length;
    const inRange = ranges.some(([start, end]) => offset < end && lineEnd > start);
    if (inRange && C25_TRUNCATE_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
    offset += line.length + 1;
  });
  return violations;
}

/** checkC27 は `Z` / `Z_CLASS` を介さない z-index 直書きを検出する（design-patterns.md §5.1）。 */
export function checkC27(text, relPath = "") {
  if (isDesignTokensFile(relPath) || isTestFile(path.basename(relPath))) return [];
  const violations = [];
  stripTsCommentsPreservingLines(text).split("\n").forEach((line, i) => {
    if (C27_Z_INDEX_RE.test(line)) violations.push({ lineNumber: i + 1, text: line.trim() });
  });
  return applyCountBaseline(violations, C27_Z_INDEX_BASELINE.get(relPath), "C27_Z_INDEX_BASELINE");
}

/**
 * checkC28 は `src/features/**` の raw `<select>` と `<input type="date">` を検出する
 * （design-patterns.md §4.4: 業務・設定フォームは共有 `DatePicker` / `Select`）。
 * §3.4 の許容文脈は C28_NATIVE_INPUT_CONTEXT_ALLOWLIST、既存の違反は RATCHET allowlist。
 */
export function checkC28(text, relPath = "") {
  if (path.extname(relPath) !== ".tsx" || isTestFile(path.basename(relPath))) return [];
  const parts = relPath.split(path.sep);
  const inScope =
    parts[0] === "src" && (parts[1] === "features" || (parts[1] === "components" && parts[2] === "shared"));
  if (!inScope) return [];
  if (C28_NATIVE_INPUT_CONTEXT_ALLOWLIST.has(relPath)) return [];

  const violations = [];
  const openingTagRe = new RegExp(C28_NATIVE_OPENING_TAG_START_RE.source, "g");
  let match;
  while ((match = openingTagRe.exec(text)) !== null) {
    if (isC18IgnoredOpeningTag(text, match.index)) continue;
    const { tag, endIndex } = readOpeningTag(text, match.index, openingTagRe.lastIndex);
    const isSelect = tag.startsWith("<select");
    if (isSelect || C28_INPUT_DATE_RE.test(tag)) {
      violations.push({ lineNumber: lineNumberAt(text, match.index), text: firstLineOf(tag) });
    }
    openingTagRe.lastIndex = Math.max(endIndex, openingTagRe.lastIndex);
  }
  return applyCountBaseline(violations, C28_NATIVE_INPUT_BASELINE.get(relPath), "C28_NATIVE_INPUT_BASELINE");
}

/**
 * checkC29 はバリデーション結果を直接 `disabled` / `submitDisabled` に流す式を検出する（design-states.md §1.1）。
 * opening tag 全文（複数行の prop を含む）と直後のテキスト子（「保存」ラベル）を 1 単位として判定する。
 * 保存／送信ボタンの契約（submitDisabled prop・SubmitButton・PrimaryButton・type="submit"・「保存」）に限り、
 * 検索・一括操作ボタンの「選択 0 件で無効」は対象外（design-states §1.1 は保存ボタンの規則）。
 */
export function checkC29(text, relPath = "") {
  if (path.extname(relPath) !== ".tsx" || isTestFile(path.basename(relPath))) return [];
  const code = stripTsCommentsPreservingLines(text);
  // prop 内に入れ子の JSX（`headerAction={<FormHeaderActions … />}`）があると外側の tag 全文にも式が含まれるため、
  // 式の位置をキーにして最も内側の tag（後から見つかる方）を採用する。
  const innermostByExpression = new Map();
  const openingTagRe = new RegExp(C29_OPENING_TAG_START_RE.source, "g");
  let match;
  while ((match = openingTagRe.exec(code)) !== null) {
    if (isC18IgnoredOpeningTag(code, match.index)) continue;
    const { tag, endIndex } = readOpeningTag(code, match.index, openingTagRe.lastIndex);
    const expressionOffset = tag.search(C29_VALIDATION_DISABLED_RE);
    if (expressionOffset === -1) continue;
    innermostByExpression.set(match.index + expressionOffset, { index: match.index, tag, endIndex });
  }
  const violations = [];
  for (const { index, tag, endIndex } of [...innermostByExpression.values()].sort((a, b) => a.index - b.index)) {
    const nextTagIndex = code.indexOf("<", endIndex);
    const textChild = code.slice(endIndex, nextTagIndex === -1 ? code.length : nextTagIndex);
    if (C29_SUBMIT_CONTEXT_RE.test(tag) || C29_SUBMIT_CONTEXT_RE.test(textChild)) {
      violations.push({ lineNumber: lineNumberAt(code, index), text: firstLineOf(tag) });
    }
  }
  return applyCountBaseline(
    violations,
    C29_VALIDATION_DISABLED_BASELINE.get(relPath),
    "C29_VALIDATION_DISABLED_BASELINE",
  );
}

/**
 * stripTsCommentsPreservingLines は行コメントとブロックコメントを空白化し、行番号を保つ。
 * 文字列・テンプレートリテラル内はそのまま残す（C20 の検出対象は template 内の合成）。
 */
export function stripTsCommentsPreservingLines(text) {
  let result = "";
  let mode = "code";
  let escaped = false;
  for (let i = 0; i < text.length; i += 1) {
    const char = text[i];
    const next = text[i + 1];

    if (mode === "line-comment") {
      if (char === "\n") {
        mode = "code";
        result += "\n";
      } else {
        result += " ";
      }
      continue;
    }
    if (mode === "block-comment") {
      if (char === "*" && next === "/") {
        mode = "code";
        result += "  ";
        i += 1;
      } else {
        result += char === "\n" ? "\n" : " ";
      }
      continue;
    }
    if (mode === "single" || mode === "double" || mode === "template") {
      result += char;
      if (escaped) escaped = false;
      else if (char === "\\") escaped = true;
      else if (
        (mode === "single" && char === "'")
        || (mode === "double" && char === '"')
        || (mode === "template" && char === "`")
      ) {
        mode = "code";
      }
      continue;
    }

    if (char === "/" && next === "/") {
      mode = "line-comment";
      result += "  ";
      i += 1;
    } else if (char === "/" && next === "*") {
      mode = "block-comment";
      result += "  ";
      i += 1;
    } else if (char === "'") {
      mode = "single";
      result += char;
    } else if (char === '"') {
      mode = "double";
      result += char;
    } else if (char === "`") {
      mode = "template";
      result += char;
    } else {
      result += char;
    }
  }
  return result;
}

/**
 * checkC20 は Tailwind v4 が拾えない実行時クラス合成を検出する（FE-RC-013 / FE-RC-106）。
 * コメントは除外する（design-tokens の FE-RC-013 説明は違反にしない）。
 */
export function checkC20(text) {
  const violations = [];
  const code = stripTsCommentsPreservingLines(text);
  code.split("\n").forEach((line, i) => {
    if (C20_RUNTIME_SYNTHESIS_RE.test(line)) {
      violations.push({ lineNumber: i + 1, text: line.trim() });
    }
  });
  return violations;
}

/**
 * checkC21 は `src/components/ui/<name>.tsx` に sibling `<name>.stories.tsx` が
 * 存在するか判定する（design-token-layers.md §6）。
 */
export function checkC21(relPath, allRelPaths) {
  const violations = [];
  const parts = relPath.split(path.sep);
  if (
    parts.length === 4 &&
    parts[0] === "src" &&
    parts[1] === "components" &&
    parts[2] === "ui" &&
    C21_UI_COMPONENT_RE.test(relPath) &&
    !C21_STORIES_ALLOWLIST.has(relPath)
  ) {
    const storiesPath = path.join(
      "src",
      "components",
      "ui",
      parts[3].replace(/\.tsx$/, ".stories.tsx"),
    );
    if (!allRelPaths.has(storiesPath)) {
      violations.push({ lineNumber: 1, text: "colocated *.stories.tsx なし（C21_STORIES_ALLOWLIST または stories 追加）" });
    }
  }
  return violations;
}

/**
 * checkC24 は `src/components/shared/` 配下のコンポーネントに stories があるか
 * 判定する。ディレクトリ単位（Dir/ 直下に *.stories.tsx が1件以上）と
 * shared/ 直下ファイル単位（sibling stories）の2形態を扱う。
 * seenDirs は collectViolations 側で使い回し、ディレクトリの重複報告を防ぐ。
 */
export function checkC24(relPath, allRelPaths, seenDirs) {
  const violations = [];
  const parts = relPath.split(path.sep);
  if (parts[0] !== "src" || parts[1] !== "components" || parts[2] !== "shared") return violations;
  const file = parts.at(-1);
  if (!file.endsWith(".tsx")) return violations;
  if (file.endsWith(".test.tsx") || file.endsWith(".stories.tsx") || file === "index.tsx") return violations;

  if (parts.length === 4) {
    if (C24_SHARED_STORIES_ALLOWLIST.has(relPath)) return violations;
    if (!allRelPaths.has(relPath.replace(/\.tsx$/, ".stories.tsx"))) {
      violations.push({ lineNumber: 1, text: "sibling *.stories.tsx なし（C24 allowlist 削減または stories 追加）" });
    }
    return violations;
  }

  const dir = parts.slice(0, 4).join(path.sep);
  if (seenDirs.has(dir) || C24_SHARED_STORIES_ALLOWLIST.has(dir)) return violations;
  seenDirs.add(dir);
  const prefix = dir + path.sep;
  const hasStories = [...allRelPaths].some(
    (p) => p.startsWith(prefix) && p.endsWith(".stories.tsx"),
  );
  if (!hasStories) {
    violations.push({ lineNumber: 1, text: "dir 内に *.stories.tsx なし（C24 allowlist 削減または stories 追加）" });
  }
  return violations;
}

/**
 * checkC22 は design-tokens.ts の PALETTE ブロック内で Tailwind クラス文字列を
 * 値に持つ新規メンバを検出する（design-token-layers.md §3 L1 純粋性ガード）。
 */
export function checkC22(text) {
  const violations = [];
  const block = text.match(C22_PALETTE_BLOCK_RE);
  if (!block) return violations;
  const blockStartLine = text.slice(0, block.index).split("\n").length;
  block[1].split("\n").forEach((line, i) => {
    const m = line.match(C22_MEMBER_RE);
    if (!m) return;
    const [, name, value] = m;
    const looksLikeClass =
      C22_VARIANT_PREFIX_RE.test(value) ||
      (C22_UTILITY_PREFIX_RE.test(value) && !value.includes("(") && !value.includes(" "));
    if (looksLikeClass && !C22_PALETTE_CLASS_MEMBER_ALLOWLIST.has(name)) {
      violations.push({
        lineNumber: blockStartLine + i,
        text: `${name}: "${value}" — PALETTE は raw 値のみ。クラス文字列は STATE / STYLE へ`,
      });
    }
  });
  return violations;
}

/** C23: sRGB hex → WCAG 相対輝度。 */
function c23Luminance(hex) {
  const n = parseInt(hex, 16);
  const channel = (v) => {
    const s = v / 255;
    return s <= 0.04045 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
  };
  return 0.2126 * channel((n >> 16) & 0xff) + 0.7152 * channel((n >> 8) & 0xff) + 0.0722 * channel(n & 0xff);
}

/**
 * checkC23 は design-tokens.ts の `BADGE.*` 各メンバについて、テンプレート内の
 * `${C.<name>}` 参照から最初の `bg-[#hex]` と `text-[#hex]` を解決し、
 * WCAG 通常文字コントラスト 4.5:1 未満のコンボを検出する（design-states.md §2）。
 * hex を持たない参照（rgba 等）はスキップ — 計算不能は違反にしない。
 */
export function checkC23(text) {
  const violations = [];
  const cBlock = text.match(C23_C_BLOCK_RE);
  const badgeBlock = text.match(C23_BADGE_BLOCK_RE);
  if (!cBlock || !badgeBlock) return violations;

  const cHexByName = new Map();
  cBlock[1].split("\n").forEach((line) => {
    const m = line.match(C23_MEMBER_RE);
    if (!m) return;
    const hexMatch = m[2].match(C23_HEX_RE);
    if (hexMatch) cHexByName.set(m[1], { hex: hexMatch[1], value: m[2] });
  });

  const badgeStartLine = text.slice(0, badgeBlock.index).split("\n").length;
  badgeBlock[1].split("\n").forEach((line, i) => {
    const m = line.match(C23_MEMBER_RE);
    if (!m) return;
    const [, name, rawValue] = m;
    let bgHex = null;
    let textHex = null;
    for (const ref of rawValue.matchAll(C23_C_REF_RE)) {
      const entry = cHexByName.get(ref[1]);
      if (!entry) continue;
      if (entry.value.includes('"bg-') && bgHex === null) bgHex = entry.hex;
      if (entry.value.includes('"text-') && textHex === null) textHex = entry.hex;
    }
    if (bgHex === null || textHex === null) return;
    const normalized =
      (Math.max(c23Luminance(bgHex), c23Luminance(textHex)) + 0.05) /
      (Math.min(c23Luminance(bgHex), c23Luminance(textHex)) + 0.05);
    if (normalized < C23_MIN_RATIO) {
      violations.push({
        lineNumber: badgeStartLine + i,
        text: `BADGE.${name} コントラスト ${normalized.toFixed(2)}:1 < ${C23_MIN_RATIO}:1（text #${textHex} / bg #${bgHex}）`,
      });
    }
  });
  return violations;
}

function isFE11ExcludedPath(relPath) {
  const segments = relPath.split(/[\\/]/);
  const normalizedPath = segments.join("/");
  return (
    segments[0] === "liff" ||
    segments[0] === "line-reserve" ||
    (segments[0] === "src" && segments[1] === "shared-liff") ||
    normalizedPath === "src/features/medical-records/components/MedicalRecordPrintView.tsx"
  );
}

async function walk(dir, exts, excludeNames) {
  let entries;
  try {
    entries = await readdir(dir, { withFileTypes: true });
  } catch {
    return [];
  }
  const files = [];
  for (const entry of entries) {
    if (excludeNames.has(entry.name)) continue;
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await walk(full, exts, excludeNames)));
    } else if (exts.has(path.extname(entry.name))) {
      files.push(full);
    }
  }
  return files;
}

/**
 * collectViolations は cwd 配下の SCAN_ROOTS（src・liff/src・line-reserve/src）を走査し、
 * C1〜C29 違反を集計する純粋寄りの関数。
 * ファイル I/O のみ副作用を持ち、判定ロジック自体は checkC1〜checkC29 に委譲する。
 */
export async function collectViolations(cwd) {
  const result = { c1: [], c3: [], c5: [], c6: [], c7: [], c8: [], c9: [], c10: [], c11: [], c12: [], c13: [], c14: [], c15: [], c16: [], c17: [], c18: [], c19: [], c20: [], c21: [], c22: [], c23: [], c24: [], c25: [], c26: [], c27: [], c28: [], c29: [] };
  const allRelPaths = new Set();
  const c24SeenDirs = new Set();

  for (const scanRoot of SCAN_ROOTS) {
    const root = path.join(cwd, scanRoot);
    const allFiles = await walk(root, SCAN_EXTENSIONS, EXCLUDE_DIR_NAMES);
    for (const file of allFiles) {
      allRelPaths.add(path.relative(cwd, file));
    }

    for (const file of allFiles) {
      const relPath = path.relative(cwd, file);
      const isTest = isTestFile(path.basename(file));
      const text = readFileSync(file, "utf-8");

      for (const v of checkC1(text)) {
        result.c1.push({ file: relPath, ...v });
      }
      if (!isTest && !isDesignTokensFile(relPath) && !isBrandTokensFile(relPath)) {
        for (const v of checkC3(text)) {
          result.c3.push({ file: relPath, ...v });
        }
      }
      if (!isTest) {
        for (const v of checkC5(text)) {
          result.c5.push({ file: relPath, ...v });
        }
      }
      if (!isTest && !isDesignTokensFile(relPath) && !isColorMapFile(relPath)) {
        for (const v of checkC6(text)) {
          result.c6.push({ file: relPath, ...v });
        }
      }
      if (!isTest) {
        for (const v of checkC7(text)) {
          result.c7.push({ file: relPath, ...v });
        }
        for (const v of checkC9(text)) {
          result.c9.push({ file: relPath, ...v });
        }
        for (const v of checkC10(text)) {
          result.c10.push({ file: relPath, ...v });
        }
        for (const v of checkC11(text)) {
          result.c11.push({ file: relPath, ...v });
        }
        for (const v of checkC12(text, relPath)) {
          result.c12.push({ file: relPath, ...v });
        }
        for (const v of checkC13(text)) {
          result.c13.push({ file: relPath, ...v });
        }
        for (const v of checkC14(text, relPath)) {
          result.c14.push({ file: relPath, ...v });
        }
        for (const v of checkC15(text, relPath)) {
          result.c15.push({ file: relPath, ...v });
        }
        for (const v of checkC16(text, relPath)) {
          result.c16.push({ file: relPath, ...v });
        }
        for (const v of checkC18(text, relPath)) {
          result.c18.push({ file: relPath, ...v });
        }
        for (const v of checkC19(text, relPath)) {
          result.c19.push({ file: relPath, ...v });
        }
        for (const v of checkC20(text)) {
          result.c20.push({ file: relPath, ...v });
        }
        for (const v of checkC25(text, relPath)) {
          result.c25.push({ file: relPath, ...v });
        }
        for (const v of checkC26(text, relPath)) {
          result.c26.push({ file: relPath, ...v });
        }
        for (const v of checkC27(text, relPath)) {
          result.c27.push({ file: relPath, ...v });
        }
        for (const v of checkC28(text, relPath)) {
          result.c28.push({ file: relPath, ...v });
        }
        for (const v of checkC29(text, relPath)) {
          result.c29.push({ file: relPath, ...v });
        }
      }
      if (isDesignTokensFile(relPath)) {
        for (const v of checkC22(text)) {
          result.c22.push({ file: relPath, ...v });
        }
        for (const v of checkC23(text)) {
          result.c23.push({ file: relPath, ...v });
        }
      }
      for (const v of checkC21(relPath, allRelPaths)) {
        result.c21.push({ file: relPath, ...v });
      }
      for (const v of checkC24(relPath, allRelPaths, c24SeenDirs)) {
        result.c24.push({ file: relPath, ...v });
      }

      // C8: src/features/<feat>/routes/<file>.tsx のみ（ネスト無し・.test 除外）
      const parts = relPath.split(path.sep);
      if (
        parts[0] === "src" &&
        parts[1] === "features" &&
        parts[3] === "routes" &&
        parts.length === 5 &&
        path.extname(parts[4]) === ".tsx" &&
        !isTest
      ) {
        for (const v of checkC8(relPath, text)) {
          result.c8.push({ file: relPath, ...v });
        }
      }
    }

    const cssFiles = await walk(root, CSS_SCAN_EXTENSIONS, EXCLUDE_DIR_NAMES);
    for (const file of cssFiles) {
      const relPath = path.relative(cwd, file);
      const text = readFileSync(file, "utf-8");
      for (const v of checkC1(text)) {
        result.c1.push({ file: relPath, ...v });
      }
      for (const v of checkC17(text, relPath)) {
        result.c17.push({ file: relPath, ...v });
      }
    }
  }
  return result;
}

function printGroup(label, violations) {
  console.log(`design-system-audit: ${label} — ${violations.length} 件`);
  for (const v of violations) {
    console.log(`  ${v.file}:${v.lineNumber}: ${v.text}`);
  }
}

async function main() {
  const args = parseArgs(process.argv.slice(2));

  let result;
  try {
    result = await collectViolations(args.cwd);
  } catch (err) {
    console.error(`design-system-audit: 実行エラー — ${err instanceof Error ? err.message : String(err)}`);
    process.exit(2);
    return;
  }

  printGroup("C1 legacy accent", result.c1);
  printGroup("C3 hex 直書き", result.c3);
  printGroup("C5 不正な colorVariant", result.c5);
  printGroup("C6 rgba/hsla 直値", result.c6);
  printGroup("C7 maxWidth 生値", result.c7);
  printGroup("C8 PageLayout 未使用 routes", result.c8);
  printGroup("C9 rounded 任意値", result.c9);
  printGroup("C10 shadow 既定/任意値", result.c10);
  printGroup("C11 font-size 任意値", result.c11);
  printGroup("C12 非仕様サイズ段(text-lg/2xl+)", result.c12);
  printGroup("C13 ink 黒アルファ", result.c13);
  printGroup("C14 tracking 直書き", result.c14);
  printGroup("C15 route named color", result.c15);
  printGroup("C16 非仕様 spacing(*-5)", result.c16);
  printGroup("C17 CSS shadow 直書き", result.c17);
  printGroup("C18 table cell override", result.c18);
  printGroup("C19 table row onClick", result.c19);
  printGroup("C20 runtime Tailwind synthesis", result.c20);
  printGroup("C21 ui stories 不在", result.c21);
  printGroup("C22 PALETTE クラス文字列", result.c22);
  printGroup("C23 BADGE コントラスト", result.c23);
  printGroup("C24 shared stories 不在", result.c24);
  printGroup("C25 truncate 全文到達手段なし", result.c25);
  printGroup("C26 nav 文言の truncate", result.c26);
  printGroup("C27 z-index 直書き", result.c27);
  printGroup("C28 features の native select/date", result.c28);
  printGroup("C29 バリデーション理由の disabled", result.c29);

  const total = Object.values(result).reduce((sum, group) => sum + group.length, 0);
  if (total > 0) {
    console.log(`design-system-audit: FAIL — ${total} 件の違反`);
    process.exit(1);
  }
  console.log("design-system-audit: PASS — 違反 0 件");
}

const isMain = process.argv[1] && import.meta.url === `file://${process.argv[1]}`;
if (isMain) {
  main();
}
