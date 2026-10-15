// External
import { memo } from "react";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { Building2, Calendar, ShieldCheck } from "lucide-react";

// Internal
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { MasterLink } from "@/components/shared/MasterLink";
import { Textarea } from "@/components/ui/textarea";
import { DatePicker } from "@/components/shared/DatePicker/DatePicker";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { SearchableSelect } from "@/components/ui/searchable-select";
import {
  NextScheduleField,
  calculateNextDate,
  resolveScheduleTypeAfterManualDate,
} from "@/components/shared/NextScheduleField";
import { cageMembersByRepresentative } from "@/lib/cage-merge";

// Relative
import { H_STYLES } from "../lib/styles";

// Types
import type { MasterItem } from "@/types";
import type { HospitalizationFormData } from "../types";

interface HospitalizationBasicInfoProps {
  formData: HospitalizationFormData;
  onChange: (updates: Partial<HospitalizationFormData>) => void;
  cageItems: MasterItem[];
}

interface CageSelectOption {
  value: string;
  label: string;
}

function formatCageOptionLabel(cage: MasterItem): string {
  return cage.description ? `${cage.name}（${cage.description}）` : cage.name;
}

/**
 * EMR-306 ケージ一時連結: 連結メンバーケージ（mergedIntoCageId 設定済み）は
 * 新規選択肢から除外する（選択しても backend が 422 で拒否するため UX 側で防ぐ）。
 * 代表ケージのラベルは「代表名 + メンバー名…」（' + ' 連結・cageItems 順）で
 * メンバー構成を表示し、（description）suffix は非連結ケージのみに付ける。
 * フォーム現在値がメンバーを指す場合（登録後に連結された）は、保存値の表示を
 * 維持するため通常ラベルで候補に残す。
 */
function buildCageSelectOptions(
  cageItems: MasterItem[],
  currentCageId: string,
): CageSelectOption[] {
  // 代表 cageId → メンバー一覧（cageItems 順）。索引構築は shared predicate
  // （src/lib/cage-merge.ts）に委譲する。
  const membersByRepId = cageMembersByRepresentative(cageItems);

  return cageItems
    .filter((cage) => cage.mergedIntoCageId == null || String(cage.id) === currentCageId)
    .map((cage) => {
      const members =
        cage.mergedIntoCageId == null ? membersByRepId.get(String(cage.id)) : undefined;
      const label =
        members == null
          ? formatCageOptionLabel(cage)
          : [cage.name, ...members.map((member) => member.name)].join(" + ");
      return { value: String(cage.id), label };
    });
}

export const HospitalizationBasicInfo = memo(function HospitalizationBasicInfo({
  formData,
  onChange,
  cageItems,
}: HospitalizationBasicInfoProps) {
  const cageOptions = buildCageSelectOptions(cageItems, formData.cageId);
  // EMR-353/R353-01: 明示選択があればそれを優先し、なければ「以外（手動）」。
  // 未選択時に「4週後」を既定表示すると、同項目の再選択が Radix の no-op になり
  // 日付が生成されない（新規フォームで実害あり）。
  const nextScheduleType = formData.nextScheduleType ?? "other";
  return (
    <div className={STYLE.formCard}>
      <h2 className={`${H_STYLES.form.cardHeading} mb-3 flex items-center gap-2`}>
        <Building2 className={`${ICON.action} ${C.text60}`} />
        基本情報
      </h2>

      {/* 入院タイプ */}
      <div className="mb-3">
        <Label
          id="hospitalization-type-label"
          className={`${H_STYLES.text.base} ${C.text60} mb-1.5 block`}
        >
          入院タイプ
        </Label>
        <RadioGroup
          value={formData.hospitalizationType}
          onValueChange={(val) => onChange({ hospitalizationType: val })}
          className="flex gap-4"
          id="hospitalization_type"
          aria-labelledby="hospitalization-type-label"
        >
          <div className="flex items-center gap-2 cursor-pointer">
            <RadioGroupItem
              value="入院"
              id="type-hospitalization"
              aria-label="入院タイプ: 入院"
              className={`${C.text}`}
            />
            <Label
              htmlFor="type-hospitalization"
              className={`inline-flex min-h-11 min-w-11 items-center ${H_STYLES.text.base} ${C.text} cursor-pointer`}
            >
              入院
            </Label>
          </div>
          <div className="flex items-center gap-2 cursor-pointer">
            <RadioGroupItem
              value="ホテル"
              id="type-hotel"
              aria-label="入院タイプ: ホテル"
              className={`${C.text}`}
            />
            <Label
              htmlFor="type-hotel"
              className={`inline-flex min-h-11 min-w-11 items-center ${H_STYLES.text.base} ${C.text} cursor-pointer`}
            >
              ホテル
            </Label>
          </div>
        </RadioGroup>
      </div>

      {/* 期間 */}
      <div className="mb-3">
        <p className={`${H_STYLES.text.base} ${C.text60} mb-1.5 block flex items-center gap-2`}>
          <Calendar className={ICON.action} />
          期間
        </p>
        <div className={`flex flex-col items-stretch ${H_STYLES.gap.default}`}>
          <Label htmlFor="start_date" className={`${H_STYLES.text.base} ${C.text60}`}>
            開始日
          </Label>
          <DatePicker
            id="start_date"
            value={formData.displayDate}
            onChange={(v) =>
              onChange({
                displayDate: v,
                // EMR-353: 明示選択済みプリセットのみ基準日変更へ追従させる。
                // 手動（other）や未選択時の既存日付は書き換えない。
                ...(formData.nextScheduleType && formData.nextScheduleType !== "other"
                  ? { nextVisit: calculateNextDate(v, formData.nextScheduleType) }
                  : {}),
              })
            }
            placeholder="開始日"
            className="flex-1"
          />
          <span className={`text-center text-sm ${C.text60}`} aria-hidden="true">
            〜
          </span>
          <Label htmlFor="end_date" className={`${H_STYLES.text.base} ${C.text60}`}>
            終了日
          </Label>
          <DatePicker
            id="end_date"
            value={formData.endDate ?? ""}
            onChange={(v) => onChange({ endDate: v })}
            placeholder="終了日"
            className="flex-1"
          />
        </div>
      </div>

      {/* ケージ/個室（EMR-316: 任意。BUG-037 の必須化を撤回し「選択なし」で登録可能。
          ケージ未設定の入院はボードの「ケージ未割当」欄に表示される） */}
      <div className="mb-3">
        <div className="flex items-center justify-between mb-1.5">
          <Label htmlFor="cage_id" className={`${H_STYLES.text.base} ${C.text60}`}>
            ケージ・個室
          </Label>
          <MasterLink category="cage" label="編集" className="text-2xs" />
        </div>
        <SearchableSelect
          id="cage_id"
          value={formData.cageId}
          onValueChange={(val) => onChange({ cageId: val })}
          options={cageOptions}
          placeholder="選択なし"
          searchPlaceholder="ケージを検索..."
          clearable
          clearLabel="選択なし"
        />
      </div>

      {/* 保険 */}
      <div className="mb-3">
        <div className="flex items-center gap-2 mb-1.5">
          <Checkbox
            id="is_insurance"
            aria-label="保険適用"
            touchTarget
            checked={formData.isInsurance}
            onCheckedChange={(checked) =>
              onChange({
                isInsurance: checked === true,
                insuranceCompanyName: checked === true ? formData.insuranceCompanyName : "",
                insuranceNumber: checked === true ? formData.insuranceNumber : "",
              })
            }
          />
          <Label
            htmlFor="is_insurance"
            className={`${H_STYLES.text.sm} ${C.text} cursor-pointer flex min-h-11 items-center gap-1`}
          >
            <ShieldCheck className={`${ICON.action} ${C.text60}`} />
            保険適用
          </Label>
        </div>
        {formData.isInsurance ? (
          <div className={`flex flex-col ${H_STYLES.gap.default} pl-6`}>
            <div>
              <Label
                htmlFor="insurance_company_name"
                className={`${H_STYLES.text.xs} ${C.text60} mb-1 block`}
              >
                保険会社名
              </Label>
              <Input
                id="insurance_company_name"
                value={formData.insuranceCompanyName}
                onChange={(e) => onChange({ insuranceCompanyName: e.target.value })}
                placeholder="保険会社名を入力..."
                className={`${H_STYLES.text.base} ${C.bgWhite} ${C.borderMedium}`}
                maxLength={100}
              />
            </div>
            <div>
              <Label
                htmlFor="insurance_number"
                className={`${H_STYLES.text.xs} ${C.text60} mb-1 block`}
              >
                保険番号
              </Label>
              <Input
                id="insurance_number"
                value={formData.insuranceNumber}
                onChange={(e) => onChange({ insuranceNumber: e.target.value })}
                placeholder="保険番号を入力..."
                className={`${H_STYLES.text.base} ${C.bgWhite} ${C.borderMedium}`}
                maxLength={50}
              />
            </div>
          </div>
        ) : null}
      </div>

      <NextScheduleField
        className="mb-3"
        typeId="hospitalization-next-schedule"
        dateId="hospitalization-next-date"
        scheduleType={nextScheduleType}
        nextDate={formData.nextVisit}
        onScheduleTypeChange={(value) =>
          onChange({
            nextScheduleType: value,
            nextVisit:
              value === "other"
                ? formData.nextVisit
                : calculateNextDate(formData.displayDate, value),
          })
        }
        onNextDateChange={(value) =>
          onChange({
            nextVisit: value,
            nextScheduleType: resolveScheduleTypeAfterManualDate(
              formData.displayDate,
              nextScheduleType,
              value,
            ),
          })
        }
      />

      {/* メモ */}
      <div>
        <Label htmlFor="memo" className={`${H_STYLES.text.base} ${C.text60} mb-1.5 block`}>
          メモ
        </Label>
        <Textarea
          id="memo"
          value={formData.memo}
          onChange={(e) => onChange({ memo: e.target.value })}
          placeholder="メモを入力..."
          className={`min-h-20 resize-none ${C.bgWhite} ${C.borderMedium} ${C.focusVisibleRingActionPrimary}`}
        />
      </div>
    </div>
  );
});
