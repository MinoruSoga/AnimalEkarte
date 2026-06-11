import { type ReactNode } from "react";
import { PawPrint, Weight } from "lucide-react";
import { C, ICON } from "@/lib/design-tokens";
import { toJSTWallDate } from "@/lib/jst-date";
import imgEllipse1 from "@/assets/231a870df600a37e011a0e1140e7608b1f4c3340.png";
import { ImageWithFallback } from "@/components/shared/Feedback";
import { Tooltip } from "@/components/ui/tooltip";

// ──────────────────────────────────────────────────────────
// Age calculation (JST)
// ──────────────────────────────────────────────────────────

function calcAge(birthDateStr: string): string {
  const today = toJSTWallDate(new Date());
  const birth = new Date(birthDateStr);
  const birthJST = toJSTWallDate(birth);
  let years = today.getFullYear() - birthJST.getFullYear();
  let months = today.getMonth() - birthJST.getMonth();
  if (months < 0) {
    years--;
    months += 12;
  }
  if (today.getDate() < birthJST.getDate()) {
    months--;
    if (months < 0) {
      years--;
      months += 12;
    }
  }
  if (years >= 1) return `${years}歳${months}ヶ月`;
  return `${months}ヶ月`;
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
  insuranceName?: string;
  insuranceDetails?: string;
  visitCount?: number;
  onOwnerClick?: () => void;
  contextControls?: ReactNode;
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
  insuranceName,
  insuranceDetails,
  visitCount,
  onOwnerClick,
  contextControls,
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
      <div className="flex flex-col gap-0.5 mr-3">
        <div className="flex items-baseline gap-2">
          {onOwnerClick ? (
            <button
              type="button"
              onClick={onOwnerClick}
              className={`text-base font-medium ${C.text} hover:underline decoration-dotted underline-offset-2 cursor-pointer`}
            >
              {ownerName}
            </button>
          ) : (
            <span className={`text-base font-medium ${C.text}`}>{ownerName}</span>
          )}
          <span className={`text-base font-medium ${isDeceased ? C.text60 : C.text}`}>
            {petName}
          </span>
          {isDeceased ? (
            <span
              className={`text-[11px] font-bold px-1.5 py-0.5 rounded ${C.bgDanger} ${C.textWhite} uppercase tracking-wider ml-1`}
            >
              【死亡】
            </span>
          ) : null}
        </div>
        <div className={`flex items-center gap-3 text-sm ${C.text60}`}>
          {petNumber ? (
            <span
              className={`font-mono text-[11px] px-1 py-0 rounded ${C.bgPage} border ${C.borderMediumLight} ${C.text40} leading-4 tracking-wide`}
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
          {weight ? (
            <Tooltip content="体重">
              <span className="flex items-center gap-1">
                <Weight className={ICON.xs} aria-hidden="true" />
                {weight}
              </span>
            </Tooltip>
          ) : null}
          {typeof visitCount === "number" && visitCount > 0 ? (
            <Tooltip content="このペットの通算来院回数です">
              <span className="flex items-center gap-1 cursor-default">
                来院 {visitCount} 回
              </span>
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
          <div className="flex items-center gap-2 overflow-x-auto no-scrollbar flex-1">
            {contextControls}
          </div>
        </>
      ) : null}
    </div>
  );
}
