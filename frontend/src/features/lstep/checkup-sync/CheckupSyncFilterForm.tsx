import { useActionState } from "react";
import { C, STYLE } from "@/lib/design-tokens";
import { SubmitButton } from "@/components/shared/Form/SubmitButton";
import type { CheckupSyncParams, CheckupType } from "../api/get-checkup-sync-preview";

const CHECKUP_TYPE_OPTIONS: { value: CheckupType; label: string }[] = [
  { value: "annual", label: "定期健診" },
  { value: "dental", label: "デンタルケア" },
  { value: "blood", label: "血液検査" },
  { value: "skin", label: "皮膚チェック" },
  { value: "cancer", label: "ガン検診" },
  { value: "other", label: "その他" },
];

interface CheckupSyncFilterFormProps {
  onSearch: (params: CheckupSyncParams) => void;
  isLoading: boolean;
}

type FilterFormState = { error: string | null };

export function CheckupSyncFilterForm({
  onSearch,
  isLoading,
}: CheckupSyncFilterFormProps) {
  const [state, formAction] = useActionState(
    async (_prev: FilterFormState, formData: FormData): Promise<FilterFormState> => {
      const checkupType = formData.get("checkup_type") as string;
      if (!checkupType) {
        return { error: "検診種別を選択してください" };
      }

      const params: CheckupSyncParams = {
        checkup_type: checkupType as CheckupType,
      };

      const species = formData.get("species") as string;
      if (species.trim()) params.species = species.trim();

      const lastVisitBefore = formData.get("last_visit_before") as string;
      if (lastVisitBefore) params.last_visit_before = lastVisitBefore;

      const lastVisitAfter = formData.get("last_visit_after") as string;
      if (lastVisitAfter) params.last_visit_after = lastVisitAfter;

      onSearch(params);
      return { error: null };
    },
    { error: null }
  );

  return (
    <form action={formAction} className={`bg-white rounded-[4px] border ${C.borderLight} p-5 space-y-4`}>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* 検診種別 */}
        <div className="space-y-1.5">
          <label className={STYLE.formLabel}>
            検診種別
            <span className={`ml-1 ${C.textRequired}`}>*</span>
          </label>
          <select
            name="checkup_type"
            className={`w-full ${STYLE.formInput} border rounded-[4px] px-3`}
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
          <label className={STYLE.formLabel}>動物種（任意）</label>
          <input
            type="text"
            name="species"
            placeholder="例: dog, cat"
            className={`w-full ${STYLE.formInput} border rounded-[4px] px-3 ${C.textPlaceholder}`}
          />
        </div>

        {/* 最終来院日 from */}
        <div className="space-y-1.5">
          <label className={STYLE.formLabel}>最終来院日（以降・任意）</label>
          <input
            type="date"
            name="last_visit_after"
            className={`w-full ${STYLE.formInput} border rounded-[4px] px-3`}
          />
        </div>

        {/* 最終来院日 to */}
        <div className="space-y-1.5">
          <label className={STYLE.formLabel}>最終来院日（以前・任意）</label>
          <input
            type="date"
            name="last_visit_before"
            className={`w-full ${STYLE.formInput} border rounded-[4px] px-3`}
          />
        </div>
      </div>

      {state.error ? (
        <p className={`text-sm ${C.danger}`}>{state.error}</p>
      ) : null}

      <div className="flex justify-end">
        <SubmitButton
          loadingText="検索中..."
          disabled={isLoading}
          className="min-w-[120px]"
        >
          対象者を検索
        </SubmitButton>
      </div>
    </form>
  );
}
