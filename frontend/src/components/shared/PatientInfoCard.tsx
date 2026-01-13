import { ChevronDown, User, Calendar, Activity } from "lucide-react";
import imgEllipse1 from "@/assets/231a870df600a37e011a0e1140e7608b1f4c3340.png";
import { ImageWithFallback } from "../figma/ImageWithFallback";
import { Button } from "../ui/button";

interface PatientInfoCardProps {
  ownerName: string;
  petName: string;
  petNumber: string;
  weight: string;
  staffName?: string;
  staffLabel?: string;
  serviceType?: string;
  className?: string;

  // New optional props for dynamic data
  petDetails?: string; // e.g. "9才5ヶ月 / メス / 避妊済"
  insuranceName?: string; // e.g. "ペット保険Aプラン"
  insuranceDetails?: string; // e.g. "普通or危険"
  nextVisitDate?: string; // e.g. "2025/10/10"
  nextVisitContent?: string; // e.g. "ノミ予防"
}

export function PatientInfoCard({
  ownerName,
  petName,
  petNumber,
  weight,
  staffName = "医師A",
  staffLabel = "",
  serviceType = "診療",
  className: _className,
  petDetails = "9才5ヶ月 / メス / 避妊済",
  insuranceName = "ペット保険Aプラン",
  insuranceDetails = "普通or危険",
  nextVisitDate = "2025/10/10",
  nextVisitContent = "ノミ予防",
}: PatientInfoCardProps) {
  return (
    <div className="sticky top-0 z-10 bg-white px-3 py-1.5 border-b border-[rgba(55,53,47,0.16)]">
      <div className="flex flex-wrap items-center gap-2">
        {/* Avatar */}
        <div className="shrink-0 size-8">
          <ImageWithFallback
            src={imgEllipse1}
            alt="Pet"
            className="size-full rounded-full object-cover border border-[rgba(55,53,47,0.09)]"
          />
        </div>

        {/* Basic Info - Compact Row */}
        <div className="flex flex-col gap-0.5 mr-3">
             <div className="flex items-baseline gap-2">
                 <span className="text-base font-medium text-[#37352F]">{ownerName}</span>
                 <span className="text-base font-medium text-[#37352F]">{petName}</span>
                 <span className="text-sm text-[#37352F]/60">{petNumber}</span>
             </div>
             <div className="flex items-center gap-3 text-sm text-[#37352F]/60">
                 <span className="flex items-center gap-1"><User className="size-3" /> {petDetails}</span>
                 <span className="flex items-center gap-1"><Activity className="size-3" /> {weight}</span>
             </div>
        </div>

        {/* Status / Discount / Service */}
        <div className="flex items-center gap-3 flex-1 overflow-x-auto no-scrollbar">
             {/* Discount/Service Group */}
             <div className="flex flex-col gap-0 min-w-[60px]">
                 <span className="text-sm text-[#37352F]/60">値引率(100%)</span>
                 <div className="flex items-center gap-1 cursor-pointer hover:bg-gray-50 rounded px-1 -ml-1 transition-colors">
                     <span className="text-sm font-medium text-[#37352F]">{serviceType}</span>
                     <ChevronDown className="size-3.5 text-[#37352F]/40" />
                 </div>
             </div>

             {/* Insurance */}
             <div className="flex flex-col gap-0 px-2 py-0.5 rounded bg-[#F7F6F3] border border-[#37352F]/10 min-w-[100px]">
                 <span className="text-sm font-medium text-[#37352F] truncate">{insuranceName}</span>
                 <span className="text-sm text-[#37352F]/60 truncate">{insuranceDetails}</span>
             </div>

             {/* Next Visit */}
             <div className="flex flex-col gap-0 px-2 py-0.5 rounded bg-[#F7F6F3] border border-[#37352F]/10 min-w-[100px]">
                  <div className="flex items-center gap-1">
                     <Calendar className="size-3 text-[#37352F]/60" />
                     <span className="text-sm text-[#37352F]">次回 {nextVisitDate}</span>
                  </div>
                  <span className="text-sm text-[#37352F]/60 truncate">{nextVisitContent}</span>
             </div>
        </div>

        {/* Staff & Actions */}
        <div className="flex items-center gap-2 ml-auto shrink-0">
            <Button variant="outline" size="sm" className="h-10 text-sm font-normal bg-white hover:bg-gray-50 px-2">
                +生体情報
            </Button>
            <Button size="sm" className="h-10 bg-[#37352F] hover:bg-[#37352F]/90 text-white text-sm gap-1 px-2">
                {staffLabel ? `${staffLabel}${staffName}` : staffName}
            </Button>
        </div>
      </div>
    </div>
  );
}
