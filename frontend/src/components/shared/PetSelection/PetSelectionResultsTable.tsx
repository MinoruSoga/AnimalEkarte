import { memo } from "react";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { Check, Octagon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { formatDate } from "@/lib/format/date";
import type { Pet } from "@/types";

interface PetSelectionResultsTableProps {
  pets: Pet[];
  onSelect: (pet: Pet) => void;
}

export const PetSelectionResultsTable = memo(function PetSelectionResultsTable({ pets, onSelect }: PetSelectionResultsTableProps) {
  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <h2 className={`text-sm font-medium ${C.text}`}>検索結果</h2>
        <span className={`text-sm ${C.text60}`}>
          {pets.length}件
        </span>
      </div>

      <div className={`rounded-lg bg-white overflow-hidden border ${C.borderMedium}`}>
        <Table>
          <TableHeader>
            <TableRow
              className={`hover:bg-transparent ${C.bgPage} border-b ${C.borderMedium} h-12`}
            >
              <TableHead className={`min-w-[80px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>飼主No</TableHead>
              <TableHead className={`min-w-[140px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>飼主名</TableHead>
              <TableHead className={`min-w-[80px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>ペット番号</TableHead>
              <TableHead className={`min-w-[100px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>ペット名</TableHead>
              <TableHead className={`min-w-[50px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>生死</TableHead>
              <TableHead className={`min-w-[50px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>種</TableHead>
              <TableHead className={`min-w-[90px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>生年月日</TableHead>
              <TableHead className={`min-w-[60px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>体重</TableHead>
              <TableHead className={`min-w-[100px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>環境</TableHead>
              <TableHead className={`min-w-[90px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>前回来院</TableHead>
              <TableHead className={`min-w-[80px] ${STYLE.sectionLabel} whitespace-nowrap h-12`}>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {pets.map((pet, index) => {
              const isDeceased = pet.status === "死亡";
              return (
                <TableRow
                  key={pet.id}
                  className={`transition-colors ${C.hoverBgMedium} h-12 ${
                    index < pets.length - 1 ? `border-b ${C.borderLight}` : "border-none"
                  } ${isDeceased ? "opacity-60 grayscale-[0.5]" : ""}`}
                >
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap`}>{pet.ownerId}</TableCell>
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap`}>{pet.ownerName}</TableCell>
                  <TableCell className={`font-mono text-sm ${C.text} whitespace-nowrap`}>{pet.petNumber || "-"}</TableCell>
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap`}>
                    <span className="flex items-center gap-1.5">
                      <span>{pet.name}</span>
                      {pet.dangerLevel === "高" ? (
                        <Popover>
                          <PopoverTrigger asChild>
                            <button
                              type="button"
                              aria-label={`${pet.name}の危険理由を表示`}
                              className={`inline-flex items-center rounded px-1.5 py-0.5 text-xs font-semibold ${C.bgDanger10} ${C.danger} ${C.borderDanger20} outline-none focus-visible:ring-2 ${C.focusRingAccent40}`}
                            >
                              ⚠ 危険
                            </button>
                          </PopoverTrigger>
                          <PopoverContent
                            align="start"
                            aria-label={`${pet.name}の危険理由`}
                            onOpenAutoFocus={(event) => event.preventDefault()}
                            className="w-64"
                          >
                            <p className={`text-sm font-semibold ${C.danger}`}>危険理由</p>
                            <p className={`mt-1 whitespace-pre-wrap break-words text-sm ${C.textInkSecondary}`}>
                              {pet.dangerReason?.trim() || "理由未登録"}
                            </p>
                          </PopoverContent>
                        </Popover>
                      ) : null}
                    </span>
                  </TableCell>
                  <TableCell className="whitespace-nowrap">
                    {pet.status ? (
                      <Badge
                        variant="secondary"
                        className={
                          pet.status === "生存"
                            ? `${C.bgStatusGreen} ${C.textStatusGreen} ${C.borderStatusGreen} ${C.hoverBgStatusGreen} text-xs px-2 py-0 h-7`
                            : `${C.bgStatusGray} ${C.textStatusGray} ${C.borderStatusGray} ${C.hoverBgStatusGray} text-xs px-2 py-0 h-7`
                        }
                      >
                        {pet.status}
                      </Badge>
                    ) : null}
                  </TableCell>
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap`}>{pet.species}</TableCell>
                  <TableCell className={`font-mono text-sm ${C.text} whitespace-nowrap`}>{formatDate(pet.birthDate)}</TableCell>
                  <TableCell className={`font-mono text-sm ${C.text} whitespace-nowrap`}>{pet.weight || "-"}</TableCell>
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap`}>{pet.environment || "-"}</TableCell>
                  <TableCell className={`font-mono text-sm ${C.text} whitespace-nowrap`}>{formatDate(pet.lastVisit)}</TableCell>
                  <TableCell className="whitespace-nowrap">
                    <Button
                      size="sm"
                      variant={isDeceased ? "ghost" : "outline"}
                      disabled={isDeceased}
                      aria-label={
                        isDeceased
                          ? `死亡・選択不可: ${pet.name} (ID ${pet.id})`
                          : `選択: ${pet.name} (ID ${pet.id})`
                      }
                      // docs/spec/design-system.md button-primary: brand と同じ primary teal + pill
                      className={`h-11 min-w-11 gap-1 ${isDeceased ? C.textStatusGray : `${C.bgActionPrimary} ${C.textOnActionPrimary} ${C.hoverBgActionPrimary} ${C.hoverTextOnActionPrimary} rounded-full`} text-sm px-4`}
                      onClick={() => onSelect(pet)}
                    >
                      {isDeceased ? (
                        <>
                          <Octagon className={ICON.action} />
                          選択不可
                        </>
                      ) : (
                        <>
                          <Check className={ICON.action} />
                          選択
                        </>
                      )}
                    </Button>
                  </TableCell>
                </TableRow>
              );
            })}
            {pets.length === 0 ? (
              <TableRow>
                <TableCell colSpan={11} className={STYLE.tableEmpty}>
                  該当するペットが見つかりません
                </TableCell>
              </TableRow>
            ) : null}
          </TableBody>
        </Table>
      </div>
    </div>
  );
});
