import { memo } from "react";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { Check, Octagon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { formatDate } from "@/utils/format/date";
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

      <div className={`rounded-lg bg-white overflow-hidden shadow-sm border ${C.borderMedium}`}>
        <Table>
          <TableHeader>
            <TableRow
              className={`hover:bg-transparent ${C.bgPage} border-b ${C.borderMedium} h-12`}
            >
              <TableHead className={`min-w-[80px] text-sm ${C.text60} whitespace-nowrap h-12`}>飼主No</TableHead>
              <TableHead className={`min-w-[140px] text-sm ${C.text60} whitespace-nowrap h-12`}>飼主名</TableHead>
              <TableHead className={`min-w-[80px] text-sm ${C.text60} whitespace-nowrap h-12`}>ペット番号</TableHead>
              <TableHead className={`min-w-[100px] text-sm ${C.text60} whitespace-nowrap h-12`}>ペット名</TableHead>
              <TableHead className={`min-w-[50px] text-sm ${C.text60} whitespace-nowrap h-12`}>生死</TableHead>
              <TableHead className={`min-w-[50px] text-sm ${C.text60} whitespace-nowrap h-12`}>種</TableHead>
              <TableHead className={`min-w-[90px] text-sm ${C.text60} whitespace-nowrap h-12`}>生年月日</TableHead>
              <TableHead className={`min-w-[60px] text-sm ${C.text60} whitespace-nowrap h-12`}>体重</TableHead>
              <TableHead className={`min-w-[100px] text-sm ${C.text60} whitespace-nowrap h-12`}>環境</TableHead>
              <TableHead className={`min-w-[90px] text-sm ${C.text60} whitespace-nowrap h-12`}>前回来院</TableHead>
              <TableHead className={`min-w-[80px] text-sm ${C.text60} whitespace-nowrap h-12`}>操作</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {pets.map((pet, index) => {
              const isDeceased = pet.status === "死亡";
              return (
                <TableRow
                  key={pet.id}
                  className={`transition-colors ${C.hoverBgMedium} cursor-pointer h-12 ${
                    index < pets.length - 1 ? `border-b ${C.borderLight}` : "border-none"
                  } ${isDeceased ? "opacity-60 grayscale-[0.5]" : ""}`}
                  onClick={() => !isDeceased && onSelect(pet)}
                >
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap py-2`}>{pet.ownerId}</TableCell>
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap py-2`}>{pet.ownerName}</TableCell>
                  <TableCell className={`font-mono text-sm ${C.text} whitespace-nowrap py-2`}>{pet.petNumber || "-"}</TableCell>
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap py-2`}>{pet.name}</TableCell>
                  <TableCell className="whitespace-nowrap py-2">
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
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap py-2`}>{pet.species}</TableCell>
                  <TableCell className={`font-mono text-sm ${C.text} whitespace-nowrap py-2`}>{formatDate(pet.birthDate)}</TableCell>
                  <TableCell className={`font-mono text-sm ${C.text} whitespace-nowrap py-2`}>{pet.weight || "-"}</TableCell>
                  <TableCell className={`text-sm ${C.text} whitespace-nowrap py-2`}>{pet.environment || "-"}</TableCell>
                  <TableCell className={`font-mono text-sm ${C.text} whitespace-nowrap py-2`}>{formatDate(pet.lastVisit)}</TableCell>
                  <TableCell className="whitespace-nowrap py-2" onClick={(e) => e.stopPropagation()}>
                    <Button
                      size="sm"
                      variant={isDeceased ? "ghost" : "outline"}
                      disabled={isDeceased}
                      className={`h-11 gap-1 ${isDeceased ? C.textStatusGray : `${C.bgAccent} ${C.bgAccentHover} ${C.textWhite}`} text-sm px-4`}
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
