import { Check, Search, SearchX } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { EmptyState, ErrorFallback } from "@/components/shared/DataStates";
import { C, ICON } from "@/lib/design-tokens";
import type { Pet } from "@/types";

interface PatientSelectionTableBodyProps {
  pets: Pet[];
  isBusy: boolean;
  selectedPetIds: Set<string>;
  onSelect: (pet: Pet) => void;
}

export function PatientSelectionTableBody({
  pets,
  isBusy,
  selectedPetIds,
  onSelect,
}: PatientSelectionTableBodyProps) {
  return (
    <Table>
      <TableHeader className={`${C.bgPage} sticky top-0 z-10`}>
        <TableRow
          className={`border-b ${C.borderMedium} h-9 ${C.hoverBgPage}`}
        >
          <TableHead className={`min-w-[80px] ${C.text40} h-9`}>
            飼主No
          </TableHead>
          <TableHead className={`min-w-[120px] ${C.text40} h-9`}>
            飼主名
          </TableHead>
          <TableHead className={`min-w-[100px] ${C.text40} h-9`}>
            ペット名
          </TableHead>
          <TableHead className={`min-w-[60px] ${C.text40} h-9`}>
            種別
          </TableHead>
          <TableHead className={`min-w-[60px] ${C.text40} h-9`}>
            性別
          </TableHead>
          <TableHead className={`min-w-[80px] ${C.text40} h-9`}>
            生年月日
          </TableHead>
          <TableHead className={`min-w-[60px] ${C.text40} h-9`}>
            体重
          </TableHead>
          <TableHead className={`min-w-[60px] ${C.text40} h-9`}>
            操作
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {pets.map((pet) => (
          <PatientSelectionRow
            key={pet.id}
            pet={pet}
            isBusy={isBusy}
            isSelected={selectedPetIds.has(pet.id)}
            onSelect={onSelect}
          />
        ))}
      </TableBody>
    </Table>
  );
}

interface PatientSelectionRowProps {
  pet: Pet;
  isBusy: boolean;
  isSelected: boolean;
  onSelect: (pet: Pet) => void;
}

export function PatientSelectionRow({
  pet,
  isBusy,
  isSelected,
  onSelect,
}: PatientSelectionRowProps) {
  const isDeceased = pet.status === "死亡";
  const isAlive = pet.status === "生存";
  const isSelectable = isAlive && !isBusy;

  return (
    <TableRow
      className={`transition-colors ${C.hoverBgLight} h-12 ${
        isSelected ? C.bgPage : ""
      } ${!isAlive ? "opacity-50 grayscale-[0.5]" : ""}`}
    >
      <TableCell className={`text-sm font-mono ${C.text}`}>
        {pet.ownerId}
      </TableCell>
      <TableCell className={`text-sm font-medium ${C.text}`}>
        {pet.ownerName}
      </TableCell>
      <TableCell className={`text-sm ${C.text}`}>
        <span className="font-bold">{pet.name}</span>
      </TableCell>
      <TableCell className={`text-sm ${C.text}`}>
        {pet.species}
      </TableCell>
      <TableCell className={`text-sm ${C.text}`}>
        {pet.gender || "-"}
      </TableCell>
      <TableCell className={`text-sm font-mono ${C.text}`}>
        {pet.birthDate || "-"}
      </TableCell>
      <TableCell className={`text-sm font-mono ${C.text}`}>
        {pet.weight || "-"}
      </TableCell>
      <TableCell>
        <Button
          size="sm"
          disabled={!isSelectable}
          aria-label={
            isDeceased
              ? `死亡・選択不可: ${pet.name} (ID ${pet.id})`
              : isBusy
                ? `読み込み中・選択不可: ${pet.name} (ID ${pet.id})`
                : !isAlive
                  ? `状態不明・選択不可: ${pet.name} (ID ${pet.id})`
                  : isSelected
                    ? `選択中: ${pet.name} (ID ${pet.id})`
                    : `選択: ${pet.name} (ID ${pet.id})`
          }
          className={`h-11 min-w-11 gap-1 text-sm px-2 transition-colors ${
            !isSelectable
              ? `${C.bgPage} ${C.textStatusGray} border-transparent cursor-not-allowed`
              : isSelected
                ? `${C.bgBrand} ${C.textOnBrand} ${C.hoverBgBrand} ${C.hoverTextOnBrand}`
                : `bg-white border ${C.borderMediumLight} ${C.text} ${C.hoverBgSubtle}`
          }`}
          onClick={() => {
            if (isSelectable) onSelect(pet);
          }}
        >
          <Check
            className={`${ICON.xs} ${isSelected ? "" : "opacity-0"}`}
          />
          {isDeceased
            ? "死亡"
            : !isAlive
              ? "不明"
              : isSelected
                ? "選択中"
                : "選択"}
        </Button>
      </TableCell>
    </TableRow>
  );
}

interface PatientSelectionResultsProps {
  hasSearchConditions: boolean;
  isSearchPending: boolean;
  error: unknown;
  isBusy: boolean;
  pets: Pet[];
  selectedPetIds: Set<string>;
  onSelect: (pet: Pet) => void;
}

export function PatientSelectionResults({
  hasSearchConditions,
  isSearchPending,
  error,
  isBusy,
  pets,
  selectedPetIds,
  onSelect,
}: PatientSelectionResultsProps) {
  return (
    <div
      className={`flex-1 rounded-lg bg-white overflow-hidden border ${C.borderMedium} min-h-0`}
    >
      <div className="overflow-auto h-full flex flex-col">
        {!hasSearchConditions && !isSearchPending ? (
          <div className="flex flex-col items-center justify-center flex-1 text-center gap-3">
            <Search className={`${ICON.xl} ${C.text20}`} />
            <div className={`text-sm ${C.text40}`}>
              検索条件を入力してください
            </div>
          </div>
        ) : error ? (
          <div className="flex-1 flex items-center justify-center">
            <ErrorFallback message="患者一覧の取得に失敗しました" />
          </div>
        ) : isBusy && pets.length === 0 ? (
          <div className="flex flex-col items-center justify-center flex-1 text-center gap-3">
            <div className="animate-spin">
              <Search className={`${ICON.xl} ${C.text20}`} />
            </div>
            <div className={`text-sm ${C.text40}`}>検索中...</div>
          </div>
        ) : pets.length === 0 ? (
          <div className="flex-1 flex items-center justify-center">
            <EmptyState
              icon={<SearchX className={ICON.xl} />}
              message="該当する患者が見つかりませんでした"
            />
          </div>
        ) : (
          <PatientSelectionTableBody
            pets={pets}
            isBusy={isBusy}
            selectedPetIds={selectedPetIds}
            onSelect={onSelect}
          />
        )}
      </div>
    </div>
  );
}
