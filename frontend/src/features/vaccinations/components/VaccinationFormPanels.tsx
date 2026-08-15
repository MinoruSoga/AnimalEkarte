import { HistoryFilterPanel } from "@/components/shared/HistoryFilterPanel";
import { MasterLink } from "@/components/shared/MasterLink";
import { FormFieldError } from "@/components/shared/FormFieldError/FormFieldError";
import { DatePicker } from "@/components/shared/DatePicker/DatePicker";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { C } from "@/lib/design-tokens";
import { toJSTWallDate } from "@/lib/jst-date";
import type { SortOrder } from "@/types";
import type { VaccinationRecord } from "../api/transforms";
import { VaccinationCard } from "./VaccinationCard";

// BUG-005: API/DB enum は "other"（model.NextScheduleTypeOther）。旧 "custom" は無効値で
// 保存失敗 or セレクタ非表示になり、手動相当 type を選べない/再表示できない。
const NEXT_SCHEDULE_ITEMS = (
  <>
    <SelectItem value="3weeks">3週後</SelectItem>
    <SelectItem value="4weeks">4週後</SelectItem>
    <SelectItem value="1year">1年後</SelectItem>
    <SelectItem value="other">以外（手動）</SelectItem>
  </>
);

interface VaccineOption {
  value: string;
  label: string;
}

interface VaccinationFieldsPanelProps {
  doctorName: string;
  date: string;
  vaccineId: string;
  vaccineOptions: VaccineOption[];
  supplemental: string;
  lot1: string;
  lot2: string;
  lot3: string;
  lot4: string;
  nextScheduleType: string;
  nextDate: string;
  remarks: string;
  fieldErrors: Record<string, string>;
  onDateChange: (value: string) => void;
  onVaccineIdChange: (value: string) => void;
  onSupplementalChange: (value: string) => void;
  onLot1Change: (value: string) => void;
  onLot2Change: (value: string) => void;
  onLot3Change: (value: string) => void;
  onLot4Change: (value: string) => void;
  onNextScheduleTypeChange: (value: string) => void;
  onNextDateChange: (value: string) => void;
  onRemarksChange: (value: string) => void;
  onMarkDirty: () => void;
}

export function VaccinationFieldsPanel({
  doctorName,
  date,
  vaccineId,
  vaccineOptions,
  supplemental,
  lot1,
  lot2,
  lot3,
  lot4,
  nextScheduleType,
  nextDate,
  remarks,
  fieldErrors,
  onDateChange,
  onVaccineIdChange,
  onSupplementalChange,
  onLot1Change,
  onLot2Change,
  onLot3Change,
  onLot4Change,
  onNextScheduleTypeChange,
  onNextDateChange,
  onRemarksChange,
  onMarkDirty,
}: VaccinationFieldsPanelProps) {
  return (
    <div className="lg:col-span-3">
      <div className={`${C.bgWhite} p-6 rounded-lg border ${C.borderLight} space-y-6`}>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-2">
            <Label htmlFor="vaccination-date">
              接種日<span className={`${C.textRequired} ml-1`}>*</span>
            </Label>
            <DatePicker
              id="vaccination-date"
              value={date}
              onChange={(value) => {
                onMarkDirty();
                onDateChange(value);
              }}
              disabledDays={{ after: toJSTWallDate(new Date()) }}
            />
            <FormFieldError message={fieldErrors.date} />
          </div>
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="vaccine-select">
                ワクチン<span className={`${C.textRequired} ml-1`}>*</span>
              </Label>
              <MasterLink category="vaccine" label="編集" className="text-2xs" />
            </div>
            <Select
              value={vaccineId}
              onValueChange={(value) => {
                if (value === "") return;
                onMarkDirty();
                onVaccineIdChange(value);
              }}
            >
              <SelectTrigger id="vaccine-select">
                <SelectValue placeholder="選択してください" />
              </SelectTrigger>
              <SelectContent>
                {vaccineOptions.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <FormFieldError message={fieldErrors.vaccineId} />
          </div>
        </div>

        {doctorName ? (
          <div className="space-y-2">
            <span className={`block text-sm ${C.text60}`}>担当医</span>
            <p className={`min-h-11 flex items-center rounded-md border ${C.borderLight} ${C.bgPage} px-3 text-sm ${C.text}`}>
              {doctorName}
            </p>
          </div>
        ) : null}

        <div className="space-y-2">
          <Label htmlFor="vaccination-supplemental">補助説明</Label>
          <Input
            id="vaccination-supplemental"
            value={supplemental}
            onChange={(event) => {
              onMarkDirty();
              onSupplementalChange(event.target.value);
            }}
            placeholder="補助説明を入力"
          />
        </div>

        <div className="space-y-2">
          <Label>LOT番号</Label>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            <LotInput id="vaccination-lot-1" label="LOT 1" value={lot1} onChange={onLot1Change} onMarkDirty={onMarkDirty} />
            <LotInput id="vaccination-lot-2" label="LOT 2" value={lot2} onChange={onLot2Change} onMarkDirty={onMarkDirty} />
            <LotInput id="vaccination-lot-3" label="LOT 3" value={lot3} onChange={onLot3Change} onMarkDirty={onMarkDirty} />
            <LotInput id="vaccination-lot-4" label="LOT 4" value={lot4} onChange={onLot4Change} onMarkDirty={onMarkDirty} />
          </div>
        </div>

        <div className="space-y-2">
          <Label htmlFor="vaccination-next-schedule">次回の予定</Label>
          <div className="flex gap-3 items-center flex-wrap">
            <Select
              value={nextScheduleType}
              onValueChange={(value) => {
                onMarkDirty();
                onNextScheduleTypeChange(value);
              }}
            >
              <SelectTrigger id="vaccination-next-schedule" className="w-[130px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>{NEXT_SCHEDULE_ITEMS}</SelectContent>
            </Select>
            <Label htmlFor="vaccination-next-date" className="sr-only">次回接種予定日</Label>
            <DatePicker
              id="vaccination-next-date"
              value={nextDate}
              onChange={(value) => {
                onMarkDirty();
                onNextDateChange(value);
              }}
              className="flex-1 min-w-[160px]"
            />
          </div>
          <FormFieldError message={fieldErrors.nextDate} />
        </div>

        <div className="space-y-2">
          <Label htmlFor="vaccination-remarks">備考</Label>
          <Textarea
            id="vaccination-remarks"
            value={remarks}
            onChange={(event) => {
              onMarkDirty();
              onRemarksChange(event.target.value);
            }}
            placeholder="備考を入力"
            className="min-h-[100px]"
          />
        </div>
      </div>
    </div>
  );
}

interface LotInputProps {
  id: string;
  label: string;
  value: string;
  onChange: (value: string) => void;
  onMarkDirty: () => void;
}

function LotInput({ id, label, value, onChange, onMarkDirty }: LotInputProps) {
  return (
    <div className="space-y-1">
      <Label htmlFor={id} className={`text-xs ${C.text60}`}>{label}</Label>
      <Input
        id={id}
        value={value}
        onChange={(event) => {
          onMarkDirty();
          onChange(event.target.value);
        }}
        placeholder={`${label}番号`}
        maxLength={50}
      />
    </div>
  );
}

interface VaccinationHistoryPanelProps {
  petHistory: VaccinationRecord[];
  filterStartDate: string;
  filterEndDate: string;
  historySearchTerm: string;
  sortOrder: SortOrder;
  onFilterStartDateChange: (value: string) => void;
  onFilterEndDateChange: (value: string) => void;
  onHistorySearchTermChange: (value: string) => void;
  onSortOrderChange: (value: SortOrder) => void;
  onClear: () => void;
}

export function VaccinationHistoryPanel({
  petHistory,
  filterStartDate,
  filterEndDate,
  historySearchTerm,
  sortOrder,
  onFilterStartDateChange,
  onFilterEndDateChange,
  onHistorySearchTermChange,
  onSortOrderChange,
  onClear,
}: VaccinationHistoryPanelProps) {
  return (
    <div className="lg:col-span-2 flex flex-col gap-3">
      <h3 className={`text-base font-semibold ${C.text}`}>過去の接種履歴</h3>

      <HistoryFilterPanel
        showDateRange
        filterStartDate={filterStartDate}
        onFilterStartDateChange={onFilterStartDateChange}
        filterEndDate={filterEndDate}
        onFilterEndDateChange={onFilterEndDateChange}
        searchTerm={historySearchTerm}
        onSearchTermChange={onHistorySearchTermChange}
        searchPlaceholder="ワクチン名で検索..."
        sortOrder={sortOrder}
        onSortOrderChange={onSortOrderChange}
        onClear={onClear}
      />

      <div className="flex flex-col gap-2 overflow-y-auto max-h-[600px]">
        {petHistory.length === 0 ? (
          <p className={`text-sm ${C.text45} py-4 text-center`}>履歴がありません</p>
        ) : (
          petHistory.map((vaccination) => (
            <VaccinationCard key={vaccination.id} vaccination={vaccination} />
          ))
        )}
      </div>
    </div>
  );
}
