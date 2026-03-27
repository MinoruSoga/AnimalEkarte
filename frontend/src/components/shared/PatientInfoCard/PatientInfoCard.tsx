import { ICON } from "@/lib/design-tokens";
import { ChevronDown, User, Calendar, Activity } from "lucide-react";
import imgEllipse1 from "@/assets/231a870df600a37e011a0e1140e7608b1f4c3340.png";
import { ImageWithFallback } from "@/components/shared/Feedback";
import { Button } from "@/components/ui/button";

interface PatientInfoCardProps {
  ownerName: string;
  petName: string;
  petNumber: string;
  weight: string;
  staffName?: string;
  staffLabel?: string;
  serviceType?: string;
  /** 診療種別ラベル（デフォルト: "診療種別"） */
  serviceTypeLabel?: string;
  className?: string;
  sticky?: boolean;
  hideStaff?: boolean;
  petDetails?: string;
  insuranceName?: string;
  insuranceDetails?: string;
  nextVisitDate?: string;
  nextVisitContent?: string;
  onStaffClick?: () => void;
  onVitalClick?: () => void;
  onOwnerClick?: () => void;
  onServiceTypeClick?: () => void;
}

export function PatientInfoCard({
  ownerName,
  petName,
  petNumber,
  weight,
  staffName = "医師A",
  staffLabel = "",
  serviceType = "診療",
  serviceTypeLabel = "診療種別",
  petDetails = "9才5ヶ月 / メス / 避妊済",
  insuranceName = "ペット保険Aプラン",
  insuranceDetails = "普通or危険",
  nextVisitDate = "2025/10/10",
  nextVisitContent = "ノミ予防",
  sticky = true,
  hideStaff = false,
  onStaffClick,
  onVitalClick,
  onOwnerClick,
  onServiceTypeClick,
}: PatientInfoCardProps) {
  return (
    <div
      className={`${sticky ? "sticky top-0 z-10" : ""} bg-white px-4 py-2.5 border-b border-[rgba(55,53,47,0.16)]`}
    >
      <div className="flex flex-wrap items-center gap-3">
        {/* Avatar */}
        <div className="shrink-0 size-9">
          <ImageWithFallback
            src={imgEllipse1}
            alt="Pet"
            className="size-full rounded-full object-cover border border-[rgba(55,53,47,0.09)]"
          />
        </div>

        {/* Basic Info */}
        <div className="flex flex-col gap-0.5 mr-3">
          <div className="flex items-baseline gap-2">
            {onOwnerClick ? (
              <button
                type="button"
                onClick={onOwnerClick}
                className="text-base font-medium text-[#37352F] hover:underline decoration-dotted underline-offset-2 cursor-pointer"
              >
                {ownerName}
              </button>
            ) : (
              <span className="text-base font-medium text-[#37352F]">{ownerName}</span>
            )}
            <span className="text-base font-medium text-[#37352F]">{petName}</span>
          </div>
          <div className="flex items-center gap-3 text-sm text-[#37352F]/60">
            {petNumber ? (
              <span className="font-mono text-[11px] px-1 py-0 rounded bg-[#F7F6F3] border border-[rgba(55,53,47,0.12)] text-[#37352F]/40 leading-4 tracking-wide">
                #{petNumber}
              </span>
            ) : null}
            <span className="flex items-center gap-1">
              <User className={ICON.xs} /> {petDetails}
            </span>
            <span className="flex items-center gap-1">
              <Activity className={ICON.xs} /> {weight}
            </span>
          </div>
        </div>

        {/* Service / Insurance / Next Visit */}
        <div className="flex items-center gap-3 flex-1 overflow-x-auto no-scrollbar">
          {/* Service Type */}
          <div className="flex flex-col gap-0 min-w-[60px]">
            <span className="text-sm text-[#37352F]/60">{serviceTypeLabel}</span>
            {onServiceTypeClick ? (
              <div
                className="flex items-center gap-1 cursor-pointer hover:bg-[#F7F6F3] rounded px-1 -ml-1 transition-colors"
                onClick={onServiceTypeClick}
              >
                <span className="text-sm font-medium text-[#37352F]">{serviceType}</span>
                <ChevronDown className={`${ICON.xs}.5 text-[#37352F]/40`} />
              </div>
            ) : (
              <span className="text-sm font-medium text-[#37352F] px-1 -ml-1">{serviceType}</span>
            )}
          </div>

          {/* Insurance */}
          <div className="flex flex-col gap-0.5 px-3 py-1.5 rounded min-h-[38px] justify-center bg-[#F7F6F3] border border-[rgba(55,53,47,0.1)] min-w-[120px]">
            <span className="text-sm font-medium text-[#37352F] truncate">{insuranceName}</span>
            <span className="text-sm text-[#37352F]/60 truncate">{insuranceDetails}</span>
          </div>

          {/* Next Visit */}
          <div className="flex flex-col gap-0.5 px-3 py-1.5 rounded min-h-[38px] justify-center bg-[#F7F6F3] border border-[rgba(55,53,47,0.1)] min-w-[120px]">
            <div className="flex items-center gap-1">
              <Calendar className={`${ICON.xs} text-[#37352F]/60`} />
              <span className="text-sm text-[#37352F]">次回 {nextVisitDate}</span>
            </div>
            <span className="text-sm text-[#37352F]/60 truncate">{nextVisitContent}</span>
          </div>
        </div>

        {/* Staff & Actions */}
        <div className="flex items-center gap-2 ml-auto shrink-0">
          {onVitalClick ? (
            <Button
              variant="outline"
              size="sm"
              className="h-9 text-sm font-normal bg-white hover:bg-[#F7F6F3] px-3 border-[rgba(55,53,47,0.16)] text-[#37352F]"
              onClick={onVitalClick}
            >
              +生体情報
            </Button>
          ) : null}
          {!hideStaff ? (
            onStaffClick ? (
              <Button
                variant="ghost"
                size="sm"
                className="h-9 text-sm gap-1 px-3 bg-[#F7F6F3] hover:bg-[rgba(55,53,47,0.1)] text-[#37352F] border border-[rgba(55,53,47,0.16)]"
                onClick={onStaffClick}
              >
                {staffLabel ? `${staffLabel}${staffName}` : staffName}
                <ChevronDown className={`${ICON.xs}.5 text-[#37352F]/40`} />
              </Button>
            ) : (
              <div className="h-9 text-sm flex items-center gap-1 px-3 rounded-md bg-[#F7F6F3] text-[#37352F] border border-[rgba(55,53,47,0.16)]">
                {staffLabel ? `${staffLabel}${staffName}` : staffName}
              </div>
            )
          ) : null}
        </div>
      </div>
    </div>
  );
}
