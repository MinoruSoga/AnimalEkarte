// React/Framework
import { useId, useMemo } from "react";

// Internal
import { SearchableSelect, type SearchableSelectOption } from "@/components/ui/searchable-select";
import { MasterLink } from "@/components/shared/MasterLink";
import { C } from "@/lib/design-tokens";
import {
  useGetAllMedicinesMaster,
  useGetAllProcedures,
  useGetAllHospitalizationPlansMaster,
} from "@/hooks/use-treatment-master";
import { usePermission } from "@/hooks/use-permission";

// Types
import type { CarePlanItemType } from "../../api/care-plan-items";

interface CarePlanRefSelectProps {
  /** ケアプラン項目の type。medicine/treatment/item 以外は何も描画しない。 */
  type: CarePlanItemType;
  /** 選択中の参照先マスタ ID(未選択時は null)。 */
  value: string | null;
  onChange: (value: string | null) => void;
  /**
   * 参照必須 type の選択変更時に選択マスタの price を伝播する。
   * クリア・マスタ未検出時は null。0 は有限値としてそのまま渡す。
   */
  onUnitPriceChange?: (price: number | null) => void;
  /** EMR-319: 参照必須エラー時に内部 SearchableSelect へ aria-invalid を伝播する */
  ariaInvalid?: boolean;
}

/**
 * ケアプラン項目の type に応じて必須マスタ参照(投薬=薬剤/処置・検査=処置/持ち物=入院プラン)を
 * 選択させる SearchableSelect。DDL の chk_care_plan_item_ref 制約に対応する UI 側の入力欄。
 *
 * 既存パターン(VaccinationForm.tsx の SearchableSelect + マスタ取得 hook)を再利用し、
 * 新しい選択 UI は発明しない。
 */
export function CarePlanRefSelect({
  type,
  value,
  onChange,
  onUnitPriceChange,
  ariaInvalid = false,
}: CarePlanRefSelectProps) {
  // 閲覧/登録権限は選択 type のマスタで判定する。item は入院マスタ、
  // medicine/treatment は医療マスタ。参照不要 type では結果を使わない。
  const permission = usePermission(type === "item" ? "master-hospitalization" : "master-medical");
  const medicinesQuery = useGetAllMedicinesMaster({
    enabled: type === "medicine" && permission.canView,
  });
  const proceduresQuery = useGetAllProcedures({
    enabled: type === "treatment" && permission.canView,
  });
  const plansQuery = useGetAllHospitalizationPlansMaster({
    enabled: type === "item" && permission.canView,
  });
  const stateMessageId = useId();

  const medicines = medicinesQuery.data;
  const procedures = proceduresQuery.data;
  const plans = plansQuery.data;

  const medicineOptions = useMemo<SearchableSelectOption[]>(
    () => (medicines ?? []).map((m) => ({ value: m.id, label: m.name })),
    [medicines],
  );
  const procedureOptions = useMemo<SearchableSelectOption[]>(
    () => (procedures ?? []).map((p) => ({ value: p.id, label: p.name })),
    [procedures],
  );
  const planOptions = useMemo<SearchableSelectOption[]>(
    () => (plans ?? []).map((p) => ({ value: p.id, label: p.name })),
    [plans],
  );

  // 選択したマスタの price を unit_price 転記用に伝播する。
  // クリア・マスタ未検出時は null。0 は有限値としてそのまま渡す。
  const emitSelection = (masters: { id: string; price: number }[] | undefined, next: string) => {
    const id = next || null;
    onChange(id);
    onUnitPriceChange?.(masters?.find((m) => m.id === id)?.price ?? null);
  };

  const ref = (() => {
    switch (type) {
      case "medicine":
        return {
          masters: medicines,
          options: medicineOptions,
          label: "薬剤",
          selectLabel: "薬剤を選択",
          searchPlaceholder: "薬剤を検索...",
          linkCategory: "medicine" as const,
        };
      case "treatment":
        return {
          masters: procedures,
          options: procedureOptions,
          label: "処置・検査",
          selectLabel: "処置・検査を選択",
          searchPlaceholder: "処置・検査を検索...",
          linkCategory: "procedure" as const,
        };
      case "item":
        return {
          masters: plans,
          options: planOptions,
          label: "入院プラン",
          selectLabel: "入院プランを選択",
          searchPlaceholder: "入院プランを検索...",
          linkCategory: "hospitalization" as const,
        };
      default:
        return null;
    }
  })();

  if (!ref) return null;

  // EMR-355: 選択中 type のクエリ状態だけを追跡する。
  // 権限不足 > error(isSuccess=false なら data=[] のキャッシュが残っていても失敗) >
  // loading > 成功0件 > 成功あり の順で判定し、error を「未登録」と混同しない。
  const query =
    type === "medicine" ? medicinesQuery : type === "treatment" ? proceduresQuery : plansQuery;
  const cannotView = !permission.canView;
  const isLoading = !cannotView && (query.isLoading || query.isPending);
  const isEmpty = query.isSuccess && !query.isFetching && ref.options.length === 0;

  return (
    <div className="flex min-w-0 flex-1 flex-col gap-1">
      <SearchableSelect
        value={value ?? ""}
        onValueChange={(next) => emitSelection(ref.masters, next)}
        options={cannotView ? [] : ref.options}
        disabled={cannotView || isLoading}
        placeholder={isLoading ? "読み込み中..." : ref.selectLabel}
        searchPlaceholder={ref.searchPlaceholder}
        ariaLabel={ref.selectLabel}
        ariaDescribedBy={cannotView || query.isError || isEmpty ? stateMessageId : undefined}
        ariaInvalid={ariaInvalid}
      />
      {cannotView ? (
        <div id={stateMessageId} role="status" className="flex items-center gap-2">
          <span className={`text-xs ${C.text60}`}>
            {ref.label}の候補を閲覧する権限がありません。権限のあるスタッフに選択を依頼してください
          </span>
        </div>
      ) : query.isError ? (
        // 生の error/HTTP 情報は出さず汎用文言に限定し、選択 type のクエリのみ再取得する。
        <div id={stateMessageId} role="alert" className="flex items-center gap-2">
          <span className={`text-xs ${C.danger}`}>{ref.label}の候補を取得できませんでした</span>
          <button
            type="button"
            disabled={query.isFetching}
            onClick={() => {
              void query.refetch();
            }}
            className={`inline-flex min-h-11 min-w-11 shrink-0 items-center justify-center rounded-xxs px-1 text-xs underline outline-none focus-visible:ring-2 ${C.focusRingAccent40} disabled:cursor-not-allowed disabled:opacity-50 ${C.text60}`}
          >
            再試行
          </button>
        </div>
      ) : isEmpty ? (
        <div id={stateMessageId} role="status" className="flex items-center gap-2">
          <span className={`text-xs ${C.text60}`}>
            登録されている{ref.label}がありません。
            {permission.canView && permission.canCreate
              ? "設定画面から登録してください"
              : "権限のあるスタッフに登録を依頼してください"}
          </span>
          {permission.canView ? <MasterLink category={ref.linkCategory} /> : null}
        </div>
      ) : null}
    </div>
  );
}
