import { memo, type ReactNode } from "react";
import { Link } from "react-router";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { ChevronDown, User, Calendar, Activity } from "lucide-react";
import imgEllipse1 from "@/assets/231a870df600a37e011a0e1140e7608b1f4c3340.png";
import { ImageWithFallback } from "@/components/shared/Feedback";
import { CheckupAlertBadge } from "@/components/shared/CheckupAlertBadge/CheckupAlertBadge";
import { DangerBadge } from "@/components/shared/DangerBadge";
import { BirthdayBadge } from "@/components/shared/BirthdayBadge";
import { Button } from "@/components/ui/button";
import { Tooltip } from "@/components/ui/tooltip";
import { RequiredMark } from "@/components/shared/RequiredMark";

interface PatientInfoCardProps {
  ownerName: string;
  petName: string;
  petNumber: string;
  weight: string;
  status?: "alive" | "deceased";
  staffName?: string;
  staffLabel?: string;
  reservationType?: string;
  /** 診療種別ラベル（デフォルト: "診療種別"） */
  reservationTypeLabel?: string;
  className?: string;
  sticky?: boolean;
  hideStaff?: boolean;
  /** EMR-412 X13: 診療種別ブロックを出さない（カルテ編集ヘッダー用）。既定は従来どおり出す。 */
  hideReservationType?: boolean;
  petDetails?: string;
  /** Domain-specific patient attributes shown inside the shared card. */
  petAttributes?: ReactNode;
  insuranceName?: string;
  insuranceDetails?: string;
  nextVisitDate?: string;
  nextVisitContent?: string;
  /** EMR-412 X4/R2: 日付チップの見出しラベル。既定「次回予定」。入院の退院予定日は「退院予定」。 */
  nextVisitLabel?: string;
  /** EMR-412 X4: CheckupAlertBadge（期限間近）を出すか。既定 true。入院の退院予定日では出さない。 */
  nextVisitShowAlert?: boolean;
  visitCount?: number;
  /** スタッフ向け飼主特記マーク (EMR-173/231)。true なら飼主名横にアイコンのみ出す。 */
  ownerIsDangerous?: boolean;
  /** ペット特記レベル (表示値 "高"/"中"/"低" または wire 値)。高/中のみ Popover バッジを出す。 */
  petDangerLevel?: string;
  /** ペット特記メモ。未設定はバッジ Popover 内で「内容未登録」表示。 */
  petDangerReason?: string;
  /** EMR-301: 生年月日 (YYYY-MM-DD)。誕生月かつ生存ペットならペット名横に Cake バッジを出す。 */
  birthDate?: string;
  onStaffClick?: () => void;
  onVitalClick?: () => void;
  onOwnerClick?: () => void;
  onReservationTypeClick?: () => void;
  staffButtonId?: string;
  /** EMR-319: 担当者選択のエラー状態。true でボタンに aria-invalid + 赤枠を付与する */
  staffInvalid?: boolean;
  /** EMR-412 X14: 担当者ボタンに必須印を出す。既定 false（トリミング新規のみ渡す）。 */
  staffRequiredMark?: boolean;
  /** EMR-174: 飼主詳細へのリンク（onOwnerClick 併存時はボタン優先） */
  ownerDetailHref?: string;
  /** EMR-174: ペット詳細への deep link（飼主詳細 ?pet=） */
  petDetailHref?: string;
  /** EMR-412 R1: 既存 microchip 表示の継続。空は出さない。 */
  microchipNumber?: string;
}

function normalizeAlertDate(value: string): string | undefined {
  const match = /^(\d{4})([-/])(\d{2})\2(\d{2})$/.exec(value);
  if (!match) return undefined;

  const [, yearText, , monthText, dayText] = match;
  const year = Number(yearText);
  const month = Number(monthText);
  const day = Number(dayText);
  const candidate = new Date(Date.UTC(year, month - 1, day));
  if (
    candidate.getUTCFullYear() !== year ||
    candidate.getUTCMonth() !== month - 1 ||
    candidate.getUTCDate() !== day
  )
    return undefined;

  return `${yearText}-${monthText}-${dayText}`;
}

export const PatientInfoCard = memo(function PatientInfoCard({
  ownerName,
  petName,
  petNumber,
  weight,
  status = "alive",
  staffName = "",
  staffLabel = "担当医",
  reservationType = "診療",
  reservationTypeLabel = "診療種別",
  // BUG-006: 臨床画面に固定ダミー属性を出さない。未指定は不明。
  petDetails = "不明",
  petAttributes,
  insuranceName = "",
  insuranceDetails = "",
  nextVisitDate,
  nextVisitContent,
  nextVisitLabel = "次回予定",
  nextVisitShowAlert = true,
  visitCount,
  ownerIsDangerous,
  petDangerLevel,
  petDangerReason,
  birthDate,
  sticky = true,
  hideStaff = false,
  hideReservationType = false,
  onStaffClick,
  onVitalClick,
  onOwnerClick,
  onReservationTypeClick,
  staffButtonId,
  staffInvalid = false,
  staffRequiredMark = false,
  ownerDetailHref,
  petDetailHref,
  microchipNumber,
}: PatientInfoCardProps) {
  const isDeceased = status === "deceased";
  const nextVisitAlertDate = nextVisitDate ? normalizeAlertDate(nextVisitDate) : undefined;
  // EMR-412 X14/R3: 既定のアクセシブル名を維持しつつ、必須時のみ （必須） を末尾へ付与する。
  const staffAriaLabel = staffLabel ? `${staffLabel}${staffName}` : `担当者を選択: ${staffName}`;

  return (
    <div
      className={`${sticky ? "sticky top-0 z-10" : ""} ${isDeceased ? C.bgPage60 : "bg-white"} px-4 py-2.5 border-b ${C.borderMedium} transition-colors`}
    >
      <div className="flex flex-wrap items-center gap-3">
        {/* Avatar */}
        <div className="shrink-0 size-9">
          <ImageWithFallback
            src={imgEllipse1}
            alt=""
            className={`size-full rounded-full object-cover border ${C.borderLight} ${isDeceased ? "grayscale opacity-60" : ""}`}
          />
        </div>

        {/* Basic Info */}
        <div className="flex min-w-0 flex-col gap-0.5 mr-3">
          <div className="flex flex-wrap items-baseline gap-2">
            <Tooltip content={ownerName} className="min-w-0 max-w-full">
              {onOwnerClick ? (
                <button
                  type="button"
                  onClick={onOwnerClick}
                  className={`inline-flex min-h-11 min-w-11 max-w-full items-center px-2 -mx-2 text-base font-medium whitespace-nowrap ${C.text} hover:underline decoration-dotted underline-offset-2 cursor-pointer`}
                >
                  <span className="truncate">{ownerName}</span>
                </button>
              ) : ownerDetailHref ? (
                // EMR-174: onOwnerClick 未指定時は飼主詳細へ link
                <Link
                  to={ownerDetailHref}
                  aria-label="飼主詳細を開く"
                  className={`inline-flex min-h-11 min-w-11 max-w-full items-center px-2 -mx-2 text-base font-medium whitespace-nowrap ${C.text} hover:underline decoration-dotted underline-offset-2`}
                >
                  <span className="truncate">{ownerName}</span>
                </Link>
              ) : (
                <span
                  className={`text-base font-medium whitespace-nowrap ${C.text} min-w-0 max-w-full`}
                >
                  <span className="block truncate">{ownerName}</span>
                </span>
              )}
            </Tooltip>
            {ownerIsDangerous ? <DangerBadge variant="owner" /> : null}
            <Tooltip content={petName} className="min-w-0 max-w-full">
              {petDetailHref ? (
                // EMR-174: ペット詳細（飼主詳細 ?pet=）へ link
                <Link
                  to={petDetailHref}
                  aria-label="ペット詳細を開く"
                  className={`inline-flex min-h-11 min-w-11 max-w-full items-center text-base font-medium whitespace-nowrap ${isDeceased ? C.text60 : C.text} hover:underline decoration-dotted underline-offset-2`}
                >
                  <span className="truncate">{petName}</span>
                </Link>
              ) : (
                <span
                  className={`text-base font-medium whitespace-nowrap ${isDeceased ? C.text60 : C.text} min-w-0 max-w-full`}
                >
                  <span className="block truncate">{petName}</span>
                </span>
              )}
            </Tooltip>
            {microchipNumber ? (
              <Tooltip content={microchipNumber} className="min-w-0 max-w-[200px]">
                <span
                  className={`font-mono text-2xs px-1 py-0 rounded ${C.bgPage} border ${C.borderMediumLight} ${C.text65} leading-4 truncate max-w-[12rem]`}
                  aria-label={`マイクロチップ番号 ${microchipNumber}`}
                >
                  {microchipNumber}
                </span>
              </Tooltip>
            ) : null}
            <DangerBadge
              variant="pet"
              level={petDangerLevel}
              subjectName={petName}
              reason={petDangerReason}
            />
            {/* EMR-301: 誕生月は生存ペットのみ。死亡は呼び出し側 status で suppress。 */}
            {!isDeceased ? <BirthdayBadge birthDate={birthDate} subjectName={petName} /> : null}
            {isDeceased ? (
              <span
                className={`text-2xs font-semibold px-1.5 py-0.5 rounded ${C.bgDanger} ${C.textWhite} uppercase ml-1`}
              >
                【死亡】
              </span>
            ) : null}
          </div>
          <div className={`flex flex-wrap items-center gap-x-3 gap-y-1 text-sm ${C.text60}`}>
            {petNumber ? (
              <span
                className={`font-mono text-2xs px-1 py-0 rounded ${C.bgPage} border ${C.borderMediumLight} ${C.text65} leading-4`}
              >
                #{petNumber}
              </span>
            ) : null}
            <span className="flex items-center gap-1">
              <User className={ICON.xs} /> {petDetails}
            </span>
            <span className="flex items-center gap-1">
              <Activity className={ICON.xs} /> {weight}
            </span>
            {petAttributes}
            {typeof visitCount === "number" ? (
              <span className="flex items-center gap-1">来院 {visitCount.toLocaleString()} 回</span>
            ) : null}
          </div>
        </div>

        {/* Service / Insurance / Next Visit — wrap at narrow widths; no rigid min-width cards (BUG-458) */}
        <div className="flex min-w-0 lg:flex-1 basis-full lg:basis-0 flex-wrap items-center gap-3">
          {/* Service Type (EMR-412 X13: hideReservationType でブロックごと非表示) */}
          {!hideReservationType ? (
            <div className="flex min-w-0 shrink-0 flex-col gap-0 whitespace-nowrap">
              <span className={`text-sm ${C.text60}`}>{reservationTypeLabel}</span>
              {onReservationTypeClick ? (
                <button
                  type="button"
                  onClick={onReservationTypeClick}
                  className={`flex items-center gap-1 cursor-pointer ${C.hoverBgPage} rounded px-1 -ml-1 transition-colors`}
                >
                  <span className={`text-sm font-medium ${C.text}`}>{reservationType}</span>
                  <ChevronDown className={`${ICON.xs} ${C.text40}`} />
                </button>
              ) : (
                <span className={`text-sm font-medium ${C.text} px-1 -ml-1`}>
                  {reservationType}
                </span>
              )}
            </div>
          ) : null}

          {/* Insurance */}
          {/* EMR-412 X5/R3: 狭幅（<lg, 820px 含む）では属性行の下へ全幅折返し・省略なし。
              lg 以上は従来の 140px カード内配置 + truncate。 */}
          <div
            className={`flex min-w-0 flex-col gap-0.5 px-3 py-1.5 rounded min-h-[38px] justify-center ${C.bgPage} border ${C.borderLight} basis-full lg:basis-[140px] lg:grow`}
          >
            <span className={`text-sm font-medium ${C.text} lg:truncate`}>
              {insuranceName || "保険情報未登録"}
            </span>
            {insuranceDetails ? (
              <span className={`text-sm ${C.text60} lg:truncate`}>{insuranceDetails}</span>
            ) : null}
          </div>

          {nextVisitDate && nextVisitDate !== "-" ? (
            <div
              className={`flex min-w-0 flex-col gap-0.5 px-3 py-1.5 rounded min-h-[38px] justify-center ${C.bgPage} border ${C.borderLight} basis-[140px] grow`}
            >
              <div className="flex items-center gap-1">
                <Calendar className={`${ICON.xs} ${C.text60}`} />
                <span className={`text-sm ${C.text} whitespace-nowrap`}>
                  {nextVisitLabel} {nextVisitDate}
                </span>
              </div>
              {nextVisitShowAlert ? <CheckupAlertBadge nextDate={nextVisitAlertDate} /> : null}
              {nextVisitContent && nextVisitContent !== "-" ? (
                <span className={`text-sm ${C.text60} truncate`}>{nextVisitContent}</span>
              ) : null}
            </div>
          ) : null}
        </div>

        {/* Staff & Actions */}
        <div className="flex w-full items-center justify-end gap-2 shrink-0 sm:w-auto sm:ml-auto">
          {onVitalClick ? (
            <Button
              type="button"
              variant="outline"
              size="sm"
              className={`h-9 text-sm font-normal bg-white ${C.hoverBgPage} px-3 ${C.borderMedium} ${C.text}`}
              onClick={onVitalClick}
            >
              +生体情報
            </Button>
          ) : null}
          {!hideStaff ? (
            onStaffClick ? (
              <Button
                id={staffButtonId}
                type="button"
                aria-label={`${staffAriaLabel}${staffRequiredMark ? "（必須）" : ""}`}
                aria-invalid={staffInvalid || undefined}
                variant="ghost"
                size="sm"
                className={`h-11 text-sm gap-1 px-3 ${C.bgPage} ${C.hoverBgMedium} ${C.text} border ${staffInvalid ? STYLE.formInputError : C.borderMedium}`}
                onClick={onStaffClick}
              >
                {staffLabel ? `${staffLabel}${staffName}` : staffName}
                {staffRequiredMark ? <RequiredMark /> : null}
                <ChevronDown className={`${ICON.xs} ${C.text40}`} />
              </Button>
            ) : (
              <div
                className={`h-9 text-sm flex items-center gap-1 px-3 rounded-md ${C.bgPage} ${C.text} border ${C.borderMedium}`}
              >
                {staffLabel ? `${staffLabel}${staffName}` : staffName}
              </div>
            )
          ) : null}
        </div>
      </div>
    </div>
  );
});
