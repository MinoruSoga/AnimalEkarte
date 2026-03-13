import { User, Calendar, Activity } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { BADGE, C } from "@/lib/design-tokens";
import type { Pet } from "@/types";

interface PetCardProps {
  pet: Pet;
  onClick?: () => void;
  className?: string;
}

function calculateAge(birthDate: string): string {
  const birth = new Date(birthDate);
  const now = new Date();

  if (isNaN(birth.getTime())) {
    return "";
  }

  const totalMonths =
    (now.getFullYear() - birth.getFullYear()) * 12 +
    (now.getMonth() - birth.getMonth());

  const years = Math.floor(totalMonths / 12);
  const months = totalMonths % 12;

  if (years === 0) {
    return `${months}ヶ月`;
  }
  if (months === 0) {
    return `${years}才`;
  }
  return `${years}才${months}ヶ月`;
}

function formatGender(gender: string | undefined): string {
  if (!gender) return "";
  const map: Record<string, string> = {
    male: "オス",
    female: "メス",
    unknown: "不明",
  };
  return map[gender] ?? gender;
}

export function PetCard({ pet, onClick, className }: PetCardProps) {
  const age = pet.birthDate ? calculateAge(pet.birthDate) : null;
  const gender = formatGender(pet.gender);
  const isAlive = pet.status !== "死亡";

  const detailParts = [age, gender, pet.breed].filter(Boolean);

  return (
    <Card
      className={`bg-white border ${C.borderLight} shadow-none rounded-[4px] hover:bg-[#F7F6F3] transition-colors ${onClick ? "cursor-pointer" : ""} ${className ?? ""}`}
      onClick={onClick}
    >
      <CardContent className="px-4 py-3 flex items-start gap-3">
        {/* Avatar placeholder */}
        <div className="shrink-0 size-9 rounded-full bg-[#F7F6F3] border border-[rgba(55,53,47,0.09)] flex items-center justify-center">
          <User className={`size-4 ${C.text40}`} />
        </div>

        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-baseline gap-2 flex-wrap">
            <span className={`text-base font-medium ${C.text} truncate`}>
              {pet.name}
            </span>
            {pet.species && (
              <span className={`text-sm ${C.text60}`}>{pet.species}</span>
            )}
            <Badge
              variant="outline"
              className={`text-xs px-1.5 h-5 font-normal border ${
                isAlive ? BADGE.green : BADGE.gray
              }`}
            >
              {pet.status ?? "生存"}
            </Badge>
          </div>

          {/* Details */}
          <div className={`flex items-center gap-3 mt-0.5 text-sm ${C.text60} flex-wrap`}>
            {detailParts.length > 0 && (
              <span className="flex items-center gap-1">
                <User className="size-3 shrink-0" />
                {detailParts.join(" / ")}
              </span>
            )}
            {pet.birthDate && (
              <span className="flex items-center gap-1">
                <Calendar className="size-3 shrink-0" />
                {pet.birthDate}
              </span>
            )}
            {pet.weight && (
              <span className="flex items-center gap-1">
                <Activity className="size-3 shrink-0" />
                {pet.weight}
              </span>
            )}
          </div>

          {/* Pet number */}
          {pet.petNumber && (
            <span
              className={`inline-block mt-1 font-mono text-[11px] px-1 py-0 rounded bg-[#F7F6F3] border border-[rgba(55,53,47,0.12)] ${C.text40} leading-4 tracking-wide`}
            >
              #{pet.petNumber}
            </span>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
