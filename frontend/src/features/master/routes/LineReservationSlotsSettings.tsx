import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router";
import { CalendarDays, Info } from "lucide-react";
import { FormHeader } from "@/components/shared/Form/FormHeader";
import { PermissionBadges } from "@/components/shared/PermissionBadges/PermissionBadges";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { C, ICON } from "@/lib/design-tokens";
import { ResourceMasterReservationType } from "@/types/generated/models";
import { useGetReservationTypes } from "../api/reservation-types";
import { ReservationTypeAvailableSlotsCalendar } from "../components/ReservationTypeAvailableSlotsCalendar";

export function LineReservationSlotsSettings() {
  const { data: types = [], isLoading } = useGetReservationTypes();
  const activeTypes = useMemo(() => types.filter((t) => t.isActive), [types]);

  const [searchParams, setSearchParams] = useSearchParams();
  const typeIdParam = searchParams.get("typeId");
  // typeId 明示指定は無効区分も許可する（予約区分マスタの無効区分パネルから遷移できるため）。
  // 未指定時のデフォルトは有効区分を優先する。
  const selectedType = types.find((t) => t.id === typeIdParam) ?? activeTypes[0] ?? types[0];

  const handleTypeChange = useCallback(
    (value: string) => {
      setSearchParams(
        (prev) => {
          const params = new URLSearchParams(prev);
          params.set("typeId", value);
          return params;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );

  return (
    <div className={`flex-1 flex flex-col h-full ${C.bgPage} min-w-0 w-full`}>
      <FormHeader
        title="LINE予約枠"
        description="LINE予約で選択できる開始時刻を日別に設定します"
        icon={<CalendarDays className={`${ICON.page} ${C.text}`} />}
        action={<PermissionBadges resource={ResourceMasterReservationType} />}
      />

      <div className="flex-1 min-h-0 flex flex-col p-4 gap-3">
        <div className="flex flex-wrap items-center gap-3">
          <Select value={selectedType?.id ?? ""} onValueChange={handleTypeChange}>
            <SelectTrigger className={`w-[240px] ${C.bgWhite} ${C.borderMedium} h-10 text-base`}>
              <SelectValue placeholder="予約区分を選択" />
            </SelectTrigger>
            <SelectContent>
              {types.map((t) => (
                <SelectItem key={t.id} value={t.id}>
                  {t.isActive ? t.name : `${t.name}（無効）`}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <div className={`flex items-center gap-1.5 text-sm ${C.text50}`}>
            <Info className={`${ICON.smXs} shrink-0`} />
            枠が1件でも登録されている場合、登録された開始時刻のみ予約可能になります（枠のない日は予約不可）。未登録の場合は営業時間から自動生成されます。
          </div>
        </div>

        {isLoading ? (
          <p className={`text-sm ${C.text40} py-4`}>読み込み中...</p>
        ) : selectedType ? (
          <ReservationTypeAvailableSlotsCalendar
            key={selectedType.id}
            clinicId={selectedType.clinicId}
            reservationTypeId={selectedType.id}
          />
        ) : (
          <p className={`text-sm ${C.text40} py-4`}>予約区分がありません</p>
        )}
      </div>
    </div>
  );
}
