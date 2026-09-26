import { type ReactNode } from "react";
import { Link } from "react-router";
import { PawPrint, Weight } from "lucide-react";
import { C, ICON } from "@/lib/design-tokens";
import { calcAgePartsAt } from "@/lib/calc-age";
import imgEllipse1 from "@/assets/231a870df600a37e011a0e1140e7608b1f4c3340.png";
import { ImageWithFallback } from "@/components/shared/Feedback";
import { Tooltip } from "@/components/ui/tooltip";
import { DangerBadge } from "@/components/shared/DangerBadge";

// ──────────────────────────────────────────────────────────
// Age calculation (JST) — FE3-9: 計算部は共有ヘルパへ委譲。
// 表示フォーマット（1歳未満は年表記省略）と未来日ガード無しの現挙動はここに残す。
// ──────────────────────────────────────────────────────────

function calcAge(birthDateStr: string): string {
  const { years, months } = calcAgePartsAt(birthDateStr, new Date());
  if (years >= 1) return `${years}歳${months}ヶ月`;
  return `${months}ヶ月`;
}

function formatNeuteredStatus(
  gender: string | undefined,
  neuteredDate: string | undefined,
): string {
  if (!neuteredDate) return "—";
  if (gender === "雄") return "去勢済";
  if (gender === "雌") return "避妊済";
  return "避妊・去勢済";
}

// ──────────────────────────────────────────────────────────
// Props
// ──────────────────────────────────────────────────────────

export interface PatientContextHeaderProps {
  ownerName: string;
  petName: string;
  petNumber: string;
  weight?: string;
  status?: "alive" | "deceased";
  /** 後方互換用。内部では birthDate + species を使う */
  petDetails?: string;
  birthDate?: string;
  species?: string;
  gender?: string;
  neuteredDate?: string;
  breed?: string;
  insuranceName?: string;
  insuranceDetails?: string;
  visitCount?: number;
  /** 既存 pet.microchip_number。ヘッダーは表示専用。空は出さない。 */
  microchipNumber?: string;
  /** スタッフ向け飼主危険マーク (EMR-173)。true なら飼主名横に ⚠ 危険人物。 */
  ownerIsDangerous?: boolean;
  /** ペット危険度 (表示値 "高"/"中"/"低" または wire 値)。高/中のみ Popover バッジを出す。 */
  petDangerLevel?: string;
  /** ペット危険理由。未設定はバッジ Popover 内で「理由未登録」表示。 */
  petDangerReason?: string;
  /** 今回カルテの最新バイタル（表示専用。時刻は出さない）。 */
  vitalsSummary?: {
    temperature?: number;
    heartRate?: number;
    respirationRate?: number;
    weight?: number;
    weightUnit?: string;
  };
  onOwnerClick?: () => void;
  contextControls?: ReactNode;
  /** EMR-174: 飼主詳細へのリンク（onOwnerClick 併存時はボタン優先） */
  ownerDetailHref?: string;
  /** EMR-174: ペット詳細への deep link（飼主詳細 ?pet=） */
  petDetailHref?: string;
}

// ──────────────────────────────────────────────────────────
// Component
// ──────────────────────────────────────────────────────────

export function PatientContextHeader({
  ownerName,
  petName,
  petNumber,
  weight,
  status = "alive",
  petDetails: _petDetails,
  birthDate,
  species,
  gender,
  neuteredDate,
  breed,
  insuranceName,
  insuranceDetails,
  visitCount,
  microchipNumber,
  ownerIsDangerous,
  petDangerLevel,
  petDangerReason,
  vitalsSummary,
  onOwnerClick,
  contextControls,
  ownerDetailHref,
  petDetailHref,
}: PatientContextHeaderProps) {
  const isDeceased = status === "deceased";

  // Build pet info line: "2023-01-07生（2歳5ヶ月）/ 猫" or "/ 猫"
  const petInfoText = (() => {
    const parts: string[] = [];
    if (birthDate) {
      const age = calcAge(birthDate);
      parts.push(`${birthDate}生（${age}）`);
    }
    if (species) {
      parts.push(species);
    }
    return parts.join(" / ");
  })();

  const hasPetAttributes =
    gender !== undefined || neuteredDate !== undefined || breed !== undefined;
  const genderText = gender || "—";
  const neuteredStatus = formatNeuteredStatus(gender, neuteredDate);
  const breedText = breed || "—";
  const hasInsurance = !!(insuranceName || insuranceDetails);
  const insuranceTooltip = hasInsurance
    ? "ペット情報に登録された保険情報です"
    : "保険情報は未設定です";

  return (
    <div
      className={`flex flex-wrap items-center gap-3 px-4 py-2.5 border-b ${C.borderMedium} ${isDeceased ? C.bgPage60 : "bg-white"}`}
    >
      {/* Avatar */}
      <div className="shrink-0 size-9">
        <ImageWithFallback
          src={imgEllipse1}
          alt="Pet"
          className={`size-full rounded-full object-cover border ${C.borderLight} ${isDeceased ? "grayscale opacity-60" : ""}`}
        />
      </div>

      {/* Patient Info */}
      <div className="flex flex-col gap-0.5 mr-3 min-w-0">
        <div className="flex items-baseline gap-2 min-w-0">
          {onOwnerClick ? (
            <Tooltip content={ownerName} className="min-w-0 max-w-[200px]">
              <button
                type="button"
                onClick={onOwnerClick}
                className={`min-h-11 min-w-11 inline-flex items-center px-2 -mx-2 text-base font-medium ${C.text} hover:underline decoration-dotted underline-offset-2 cursor-pointer truncate w-full`}
              >
                {ownerName}
              </button>
            </Tooltip>
          ) : ownerDetailHref ? (
            // EMR-174: onOwnerClick 未指定時は飼主詳細へ link
            <Tooltip content={ownerName} className="min-w-0 max-w-[200px]">
              <Link
                to={ownerDetailHref}
                aria-label="飼主詳細を開く"
                className={`min-h-11 min-w-11 inline-flex items-center px-2 -mx-2 text-base font-medium ${C.text} hover:underline decoration-dotted underline-offset-2 truncate w-full`}
              >
                {ownerName}
              </Link>
            </Tooltip>
          ) : (
            <Tooltip content={ownerName} className="min-w-0 max-w-[200px]">
              <span className={`text-base font-medium ${C.text} truncate w-full`}>{ownerName}</span>
            </Tooltip>
          )}
          {ownerIsDangerous ? <DangerBadge variant="owner" /> : null}
          <Tooltip content={petName} className="min-w-0 max-w-[160px]">
            {petDetailHref ? (
              // EMR-174: ペット詳細（飼主詳細 ?pet=）へ link
              <Link
                to={petDetailHref}
                aria-label="ペット詳細を開く"
                className={`inline-flex min-h-11 items-center text-base font-medium ${isDeceased ? C.text60 : C.text} hover:underline decoration-dotted underline-offset-2 truncate w-full`}
              >
                {petName}
              </Link>
            ) : (
              <span
                className={`text-base font-medium ${isDeceased ? C.text60 : C.text} truncate w-full`}
              >
                {petName}
              </span>
            )}
          </Tooltip>
          <DangerBadge
            variant="pet"
            level={petDangerLevel}
            subjectName={petName}
            reason={petDangerReason}
          />
          {microchipNumber ? (
            <Tooltip content={microchipNumber} className="min-w-0 max-w-[200px]">
              <span
                className={`font-mono text-2xs px-1 py-0 rounded ${C.bgPage} border ${C.borderMediumLight} ${C.text40} leading-4 truncate max-w-[12rem]`}
                aria-label={`マイクロチップ番号 ${microchipNumber}`}
              >
                {microchipNumber}
              </span>
            </Tooltip>
          ) : null}
          {isDeceased ? (
            <span
              className={`text-2xs font-semibold px-1.5 py-0.5 rounded ${C.bgDanger} ${C.textWhite} uppercase ml-1`}
            >
              【死亡】
            </span>
          ) : null}
        </div>
        <div className={`flex items-center flex-wrap gap-x-3 gap-y-0.5 text-sm ${C.text60}`}>
          {petNumber ? (
            <span
              className={`font-mono text-2xs px-1 py-0 rounded ${C.bgPage} border ${C.borderMediumLight} ${C.text40} leading-4`}
            >
              #{petNumber}
            </span>
          ) : null}
          {petInfoText ? (
            <Tooltip content={petInfoText}>
              <span className={`flex items-center gap-1 max-w-[200px] truncate`}>
                <PawPrint className={ICON.xs} aria-hidden="true" />
                <span className="truncate">{petInfoText}</span>
              </span>
            </Tooltip>
          ) : null}
          {hasPetAttributes ? (
            <>
              <span className="flex items-center gap-1 whitespace-nowrap">
                <span className={`text-2xs ${C.text40}`}>性別</span>
                <span className={`font-medium ${C.text}`}>{genderText}</span>
              </span>
              <span className="flex items-center gap-1 whitespace-nowrap">
                <span className={`text-2xs ${C.text40}`}>避妊去勢</span>
                <span className={`font-medium ${C.text}`}>{neuteredStatus}</span>
              </span>
              <Tooltip content={`品種: ${breedText}`} className="min-w-0 max-w-[180px]">
                <span className="flex items-center gap-1 min-w-0">
                  <span className={`text-2xs ${C.text40} shrink-0`}>品種</span>
                  <span className={`font-medium ${C.text} truncate`}>{breedText}</span>
                </span>
              </Tooltip>
            </>
          ) : null}
          {weight ? (
            <Tooltip content="体重">
              <span className="flex items-center gap-1">
                <Weight className={ICON.xs} aria-hidden="true" />
                {weight}
              </span>
            </Tooltip>
          ) : null}
          {vitalsSummary ? (
            <span
              className="flex items-center flex-wrap gap-x-2 gap-y-0.5"
              aria-label="今回のバイタル"
            >
              {vitalsSummary.temperature != null ? (
                <span className="whitespace-nowrap">T {vitalsSummary.temperature}</span>
              ) : null}
              {vitalsSummary.heartRate != null ? (
                <span className="whitespace-nowrap">HR {vitalsSummary.heartRate}</span>
              ) : null}
              {vitalsSummary.respirationRate != null ? (
                <span className="whitespace-nowrap">RR {vitalsSummary.respirationRate}</span>
              ) : null}
              {vitalsSummary.weight != null ? (
                <span className="whitespace-nowrap">
                  測定体重 {vitalsSummary.weight}
                  {vitalsSummary.weightUnit ?? ""}
                </span>
              ) : null}
            </span>
          ) : null}
          {typeof visitCount === "number" && visitCount > 0 ? (
            <Tooltip content="このペットの通算来院回数です">
              <span className="flex items-center gap-1 cursor-default">来院 {visitCount} 回</span>
            </Tooltip>
          ) : null}
          {/* Insurance */}
          <Tooltip content={insuranceTooltip}>
            <div
              className={`flex flex-col gap-0.5 px-3 py-1.5 rounded min-h-[32px] justify-center ${C.bgPage} border ${C.borderLight} cursor-default`}
              aria-label={insuranceTooltip}
            >
              <span className={`text-xs font-medium ${C.text} truncate max-w-[140px]`}>
                {insuranceName || "保険情報未登録"}
              </span>
              {insuranceDetails ? (
                <span className={`text-xs ${C.text60} truncate max-w-[140px]`}>
                  {insuranceDetails}
                </span>
              ) : null}
            </div>
          </Tooltip>
        </div>
      </div>

      {/* Context Controls */}
      {contextControls ? (
        <>
          <div className={`self-stretch w-px ${C.bgLight} mx-1`} aria-hidden="true" />
          <div className="flex items-center gap-2 min-w-0 overflow-x-auto flex-1 pb-1">
            {contextControls}
          </div>
        </>
      ) : null}
    </div>
  );
}
