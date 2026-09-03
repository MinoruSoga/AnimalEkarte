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
 *   C18 — Table primitive / raw th・td 呼び出し側で typography / padding / 背景を非仕様値へ上書きすることを禁止。
 *   C19 — table row（DataTableRow / SortableDataTableRow / TableRow / tr）の行全体クリックを禁止。
 *   C20 — Tailwind 実行時合成（`hover:${` / `focus:${` / `text-[${`）禁止。完成形静的トークンのみ（FE-RC-013 / FE-RC-106）。
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
const C17_RE = /\bbox-shadow\s*:|\bfilter\s*:[^;]*\bdrop-shadow\s*\(/;
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
 * checkC17 は CSS から elevation token を迂回する直接 shadow 宣言を検出する。
 * `--shadow-*` custom property の定義は `box-shadow:` 宣言ではないため許可される。
 */
export function checkC17(text, relPath = "") {
  if (isFE11ExcludedPath(relPath)) return [];
  const violations = [];
  text.split("\n").forEach((line, i) => {
    if (C17_RE.test(line)) {
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
 * C1〜C20 違反を集計する純粋寄りの関数。
 * ファイル I/O のみ副作用を持ち、判定ロジック自体は checkC1〜checkC20 に委譲する。
 */
export async function collectViolations(cwd) {
  const result = { c1: [], c3: [], c5: [], c6: [], c7: [], c8: [], c9: [], c10: [], c11: [], c12: [], c13: [], c14: [], c15: [], c16: [], c17: [], c18: [], c19: [], c20: [] };

  for (const scanRoot of SCAN_ROOTS) {
    const root = path.join(cwd, scanRoot);
    const allFiles = await walk(root, SCAN_EXTENSIONS, EXCLUDE_DIR_NAMES);

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

  const total = result.c1.length + result.c3.length + result.c5.length + result.c6.length
    + result.c7.length + result.c8.length + result.c9.length + result.c10.length + result.c11.length
    + result.c12.length + result.c13.length + result.c14.length + result.c15.length + result.c16.length
    + result.c17.length + result.c18.length + result.c19.length + result.c20.length;
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
