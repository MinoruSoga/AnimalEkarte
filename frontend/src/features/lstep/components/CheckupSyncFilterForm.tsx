import { useActionState } from "react";
import { useAnimalSpecies } from "@/hooks/use-animal-species";
import { C, STYLE } from "@/lib/design-tokens";
import { getFormString } from "@/lib/form-data";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import type { CheckupSyncParams, CheckupType } from "../api/get-checkup-sync-preview";
import { CPM_STAGE_OPTIONS, type CPMStage } from "@/lib/cpm-stage";
import { RequiredMark } from "@/components/shared/RequiredMark";

const CHECKUP_TYPE_OPTIONS: { value: CheckupType; label: string }[] = [
  { value: "annual", label: "定期健診" },
  { value: "dental", label: "デンタルケア" },
  { value: "blood", label: "血液検査" },
  { value: "skin", label: "皮膚チェック" },
  { value: "cancer", label: "ガン検診" },
  { value: "other", label: "その他" },
];

// CPM ステージ選択肢は共有定義 @/lib/cpm-stage の CPM_STAGE_OPTIONS を使用（ISSUE-180）。

const CHRONIC_OPTIONS: { value: string; label: string }[] = [
  { value: "", label: "指定しない" },
  { value: "true", label: "あり" },
  { value: "false", label: "なし" },
];

interface CheckupSyncFilterFormProps {
  onSearch: (params: CheckupSyncParams) => void;
  isLoading: boolean;
}

type FilterFormState = { error: string | null };

// ISSUE-009: クエリ整数変換ヘルパー — 空文字は undefined、不正値はバリデーションエラーで握る。
function parseOptionalInt(
  raw: FormDataEntryValue | null,
): { ok: true; value: number | undefined } | { ok: false; error: string } {
  if (typeof raw !== "string" || raw.trim() === "") {
    return { ok: true, value: undefined };
  }
  const v = Number(raw);
  if (!Number.isFinite(v) || v < 0 || !Number.isInteger(v)) {
    return { ok: false, error: "0 以上の整数で指定してください" };
  }
  return { ok: true, value: v };
}

export function CheckupSyncFilterForm({ onSearch, isLoading }: CheckupSyncFilterFormProps) {
  const {
    activeSpecies,
    isLoading: isSpeciesLoading,
    isError: isSpeciesError,
  } = useAnimalSpecies();
  const speciesOptions = activeSpecies.filter((species) => species.is_active === true);
  const speciesUnavailable = isSpeciesError || speciesOptions.length === 0;
  const [state, formAction] = useActionState(
    async (_prev: FilterFormState, formData: FormData): Promise<FilterFormState> => {
      const checkupType = getFormString(formData, "checkup_type");
      if (!checkupType) {
        return { error: "検診種別を選択してください" };
      }

      const params: CheckupSyncParams = {
        checkup_type: checkupType as CheckupType,
      };

      // PR665: 貼り付けた前後空白を正規化し、空白のみは未指定として扱う。
      const species = getFormString(formData, "species").trim();
      if (species) params.species = species;

      const lastVisitBefore = getFormString(formData, "last_visit_before");
      if (lastVisitBefore) params.last_visit_before = lastVisitBefore;

      const lastVisitAfter = getFormString(formData, "last_visit_after");
      if (lastVisitAfter) params.last_visit_after = lastVisitAfter;

      // ISSUE-009: 追加フィルタ
      const minAge = parseOptionalInt(formData.get("min_age_years"));
      if (!minAge.ok) return { error: `最小年齢は${minAge.error}` };
      if (minAge.value !== undefined) params.min_age_years = minAge.value;

      const maxAge = parseOptionalInt(formData.get("max_age_years"));
      if (!maxAge.ok) return { error: `最大年齢は${maxAge.error}` };
      if (maxAge.value !== undefined) params.max_age_years = maxAge.value;

      if (minAge.value !== undefined && maxAge.value !== undefined && minAge.value > maxAge.value) {
        return { error: "最小年齢は最大年齢以下で指定してください" };
      }

      const chronic = getFormString(formData, "has_chronic_condition");
      if (chronic === "true") params.has_chronic_condition = true;
      else if (chronic === "false") params.has_chronic_condition = false;

      const cpmStage = getFormString(formData, "cpm_stage");
      if (cpmStage) params.cpm_stage = cpmStage as CPMStage;

      const minTotalAmount = parseOptionalInt(formData.get("min_total_amount"));
      if (!minTotalAmount.ok) {
        return { error: `累計診療費（最小）は${minTotalAmount.error}` };
      }
      if (minTotalAmount.value !== undefined) {
        params.min_total_amount = minTotalAmount.value;
      }

      const minAnnualVisitCount = parseOptionalInt(formData.get("min_annual_visit_count"));
      if (!minAnnualVisitCount.ok) {
        return { error: `年間来院回数（最小）は${minAnnualVisitCount.error}` };
      }
      if (minAnnualVisitCount.value !== undefined) {
        params.min_annual_visit_count = minAnnualVisitCount.value;
      }

      const lastCheckupBefore = getFormString(formData, "last_checkup_before");
      if (lastCheckupBefore) params.last_checkup_before = lastCheckupBefore;

      const lastCheckupAfter = getFormString(formData, "last_checkup_after");
      if (lastCheckupAfter) params.last_checkup_after = lastCheckupAfter;

      onSearch(params);
      return { error: null };
    },
    { error: null },
  );

  return (
    <form
      action={formAction}
      className={`${C.bgWhite} rounded-xs border ${C.borderLight} p-6 space-y-4`}
    >
      {/* EMR-385: lg 4列だとラベル（最終健診実施日（以降・任意）等）が1文字単位で
          折り返されるため 3列に広げ、十分な幅で xl から 4列にする */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
        {/* 検診種別 */}
        <div className="space-y-1.5">
          <label
            htmlFor="checkup_type"
            className={`${STYLE.formLabel} inline-flex items-center gap-2`}
          >
            検診種別
            <RequiredMark />
          </label>
          <select
            id="checkup_type"
            name="checkup_type"
            className={`w-full ${STYLE.formInput} border rounded-xs px-3`}
          >
            <option value="">選択してください</option>
            {CHECKUP_TYPE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        {/* 動物種 */}
        <div className="space-y-1.5">
          {/* EMR-385: 括弧内を含むラベルはまとまり単位（主語部・括弧部）でのみ折返す */}
          <label htmlFor="species" className={STYLE.formLabel}>
            <span className="whitespace-nowrap">動物種</span>
            <span className="whitespace-nowrap">（任意）</span>
          </label>
          {isSpeciesLoading ? (
            <select
              id="species"
              name="species"
              disabled
              className={`w-full ${STYLE.formInput} border rounded-xs px-3`}
            >
              <option value="">指定しない</option>
            </select>
          ) : speciesUnavailable ? (
            <>
              {/* PR665: マスタ権限がない owners:view ユーザーでも preview を使えるよう、
                  権限拒否・エラー・空マスタ時は自由入力へフォールバックする（サーバ権限は広げない）。 */}
              <input
                id="species"
                name="species"
                type="text"
                placeholder="動物種を直接入力（任意）"
                aria-describedby="species-lookup-guidance"
                className={`w-full ${STYLE.formInput} border rounded-xs px-3`}
              />
              <p id="species-lookup-guidance" className={`text-sm ${C.text65}`}>
                動物種マスタを取得できないため、直接入力できます
              </p>
            </>
          ) : (
            <select
              id="species"
              name="species"
              className={`w-full ${STYLE.formInput} border rounded-xs px-3`}
            >
              <option value="">指定しない</option>
              {speciesOptions.map((species) => (
                <option key={species.id} value={species.name}>
                  {species.name}
                </option>
              ))}
            </select>
          )}
        </div>

        {/* 最終来院日 from */}
        <div className="space-y-1.5">
          <label htmlFor="last_visit_after" className={STYLE.formLabel}>
            <span className="whitespace-nowrap">最終来院日</span>
            <span className="whitespace-nowrap">（以降・任意）</span>
          </label>
          <input
            id="last_visit_after"
            type="date"
            name="last_visit_after"
            className={`w-full ${STYLE.formInput} border rounded-xs px-3`}
          />
        </div>

        {/* 最終来院日 to */}
        <div className="space-y-1.5">
          <label htmlFor="last_visit_before" className={STYLE.formLabel}>
            <span className="whitespace-nowrap">最終来院日</span>
            <span className="whitespace-nowrap">（以前・任意）</span>
          </label>
          <input
            id="last_visit_before"
            type="date"
            name="last_visit_before"
            className={`w-full ${STYLE.formInput} border rounded-xs px-3`}
          />
        </div>

        {/* ISSUE-009: 年齢（最小） */}
        <div className="space-y-1.5">
          <label htmlFor="min_age_years" className={STYLE.formLabel}>
            <span className="whitespace-nowrap">年齢 最小</span>
            <span className="whitespace-nowrap">（任意・歳）</span>
          </label>
          <input
            id="min_age_years"
            type="number"
            name="min_age_years"
            min={0}
            step={1}
            placeholder="例: 7"
            className={`w-full ${STYLE.formInput} border rounded-xs px-3`}
          />
        </div>

        {/* ISSUE-009: 年齢（最大） */}
        <div className="space-y-1.5">
          <label htmlFor="max_age_years" className={STYLE.formLabel}>
            <span className="whitespace-nowrap">年齢 最大</span>
            <span className="whitespace-nowrap">（任意・歳）</span>
          </label>
          <input
            id="max_age_years"
            type="number"
            name="max_age_years"
            min={0}
            step={1}
            placeholder="例: 12"
            className={`w-full ${STYLE.formInput} border rounded-xs px-3`}
          />
        </div>

        {/* ISSUE-009: 慢性疾患 */}
        <div className="space-y-1.5">
          <label htmlFor="has_chronic_condition" className={STYLE.formLabel}>
            <span className="whitespace-nowrap">慢性疾患</span>
            <span className="whitespace-nowrap">（任意）</span>
          </label>
          <select
            id="has_chronic_condition"
            name="has_chronic_condition"
            defaultValue=""
            className={`w-full ${STYLE.formInput} border rounded-xs px-3`}
          >
            {CHRONIC_OPTIONS.map((opt) => (
              <option key={opt.value || "any"} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        {/* ISSUE-009: CPM ステージ */}
        <div className="space-y-1.5">
          <label htmlFor="cpm_stage" className={STYLE.formLabel}>
            <span className="whitespace-nowrap">CPMステージ</span>
            <span className="whitespace-nowrap">（任意）</span>
          </label>
          <select
            id="cpm_stage"
            name="cpm_stage"
            defaultValue=""
            className={`w-full ${STYLE.formInput} border rounded-xs px-3`}
          >
            <option value="">指定しない</option>
            {CPM_STAGE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        {/* ISSUE-009: 累計診療費（最小） */}
        <div className="space-y-1.5">
          <label htmlFor="min_total_amount" className={STYLE.formLabel}>
            <span className="whitespace-nowrap">累計診療費</span>
            <span className="whitespace-nowrap">（最小・円）</span>
          </label>
          <input
            id="min_total_amount"
            type="number"
            name="min_total_amount"
            min={0}
            step={1000}
            placeholder="例: 50000"
            className={`w-full ${STYLE.formInput} border rounded-xs px-3`}
          />
        </div>

        {/* ISSUE-009: 年間来院回数（最小） */}
        <div className="space-y-1.5">
          <label htmlFor="min_annual_visit_count" className={STYLE.formLabel}>
            <span className="whitespace-nowrap">年間来院回数</span>
            <span className="whitespace-nowrap">（最小・回）</span>
          </label>
          <input
            id="min_annual_visit_count"
            type="number"
            name="min_annual_visit_count"
            min={0}
            step={1}
            placeholder="例: 2"
            className={`w-full ${STYLE.formInput} border rounded-xs px-3`}
          />
        </div>

        {/* ISSUE-009: 最終健診実施日 from */}
        <div className="space-y-1.5">
          <label htmlFor="last_checkup_after" className={STYLE.formLabel}>
            <span className="whitespace-nowrap">最終健診実施日</span>
            <span className="whitespace-nowrap">（以降・任意）</span>
          </label>
          <input
            id="last_checkup_after"
            type="date"
            name="last_checkup_after"
            className={`w-full ${STYLE.formInput} border rounded-xs px-3`}
          />
        </div>

        {/* ISSUE-009: 最終健診実施日 to */}
        <div className="space-y-1.5">
          <label htmlFor="last_checkup_before" className={STYLE.formLabel}>
            <span className="whitespace-nowrap">最終健診実施日</span>
            <span className="whitespace-nowrap">（以前・任意）</span>
          </label>
          <input
            id="last_checkup_before"
            type="date"
            name="last_checkup_before"
            className={`w-full ${STYLE.formInput} border rounded-xs px-3`}
          />
        </div>
      </div>

      {state.error ? <p className={`text-sm ${C.danger}`}>{state.error}</p> : null}

      <div className="flex justify-end">
        <SubmitButton loadingText="検索中..." disabled={isLoading} className="min-w-[120px]">
          対象者を検索
        </SubmitButton>
      </div>
    </form>
  );
}
