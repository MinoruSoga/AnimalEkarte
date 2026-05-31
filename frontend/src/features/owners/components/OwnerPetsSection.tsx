import { memo } from "react";
import { useNavigate } from "react-router";
import { Bed, Calendar, CreditCard, Edit, FileText, MoreHorizontal, PawPrint, Plus, Scissors, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import type { PetFormData } from "../types";

interface PetTableRowProps {
  pet: PetFormData;
  ownerId: string | undefined;
  canEdit: boolean;
  canCreate: boolean;
  canDelete: boolean;
  onEdit: (pet: PetFormData) => void;
  onDeleteRequest: (id: string, name: string) => void;
}

const PetTableRow = memo(function PetTableRow({
  pet,
  ownerId,
  canEdit,
  canCreate,
  canDelete,
  onEdit,
  onDeleteRequest,
}: PetTableRowProps) {
  const navigate = useNavigate();
  const backFrom = ownerId
    ? paths.owners.detail.getHref(ownerId)
    : paths.owners.getHref();

  return (
    <TableRow
      className={`transition-colors ${C.borderDivider} ${C.hoverBgPage} h-12 ${canEdit ? "cursor-pointer" : "cursor-default"}`}
      onClick={canEdit ? () => onEdit(pet) : undefined}
    >
      <TableCell className={STYLE.tableCell}>{pet.petNumber}</TableCell>
      <TableCell className={STYLE.tableCell}>{pet.petName}</TableCell>
      <TableCell className={STYLE.tableCell}>{pet.status}</TableCell>
      <TableCell className={STYLE.tableCell}>{pet.species}</TableCell>
      <TableCell className={STYLE.tableCell}>{pet.gender}</TableCell>
      <TableCell className={STYLE.tableCell}>
        {pet.birthDate ? pet.birthDate.slice(0, 10) : ""}
      </TableCell>
      <TableCell className={STYLE.tableCell}>{pet.color}</TableCell>
      <TableCell className={STYLE.tableCell}>
        {pet.weight ? `${pet.weight} kg` : ""}
      </TableCell>
      <TableCell className={STYLE.tableCell}>{pet.environment}</TableCell>
      <TableCell className={`${STYLE.tableCell} truncate max-w-[200px]`}>
        {pet.remarks}
      </TableCell>
      <TableCell className="py-2">
        <div className="flex gap-1 justify-end">
          <DropdownMenu>
            <DropdownMenuTrigger
              className={`inline-flex items-center justify-center rounded-[4px] cursor-pointer ${STYLE.tableActionBtn} ${C.hoverBgLight}`}
              aria-label="操作メニューを開く"
              onClick={(event) => event.stopPropagation()}
            >
              <MoreHorizontal className={ICON.page} />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>操作</DropdownMenuLabel>
              {canEdit ? (
                <DropdownMenuItem onClick={() => onEdit(pet)}>
                  <Edit className={`mr-2 ${ICON.action}`} />
                  詳細・編集
                </DropdownMenuItem>
              ) : null}
              {canCreate ? (
                <>
                  <DropdownMenuItem onClick={() => navigate(`${paths.reservations.getHref()}?petId=${pet.id}`)}>
                    <Calendar className={`mr-2 ${ICON.action}`} />
                    予約作成
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => navigate(`${paths.medicalRecords.new.getHref()}?petId=${pet.id}`, { state: { from: backFrom } })}
                  >
                    <FileText className={`mr-2 ${ICON.action}`} />
                    カルテ作成
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => navigate(`${paths.trimming.new.getHref()}?petId=${pet.id}`, { state: { from: backFrom } })}
                  >
                    <Scissors className={`mr-2 ${ICON.action}`} />
                    トリミング
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => navigate(`${paths.hospitalization.new.getHref()}?petId=${pet.id}`, { state: { from: backFrom } })}
                  >
                    <Bed className={`mr-2 ${ICON.action}`} />
                    入院登録
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => navigate(`${paths.accounting.new.getHref()}?petId=${pet.id}`, { state: { from: backFrom } })}
                  >
                    <CreditCard className={`mr-2 ${ICON.action}`} />
                    会計登録
                  </DropdownMenuItem>
                </>
              ) : null}
              {canDelete ? (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    onClick={() => onDeleteRequest(pet.id, pet.petName)}
                    className={`${C.danger} focus:${C.danger} ${C.focusBgLight}`}
                  >
                    <Trash2 className={`mr-2 ${ICON.action}`} />
                    削除
                  </DropdownMenuItem>
                </>
              ) : null}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </TableCell>
    </TableRow>
  );
});

const PET_TABLE_HEADER = (
  <TableHeader>
    <TableRow className={`hover:bg-transparent ${C.bgPage} border-b ${C.borderMedium} h-12`}>
      <TableHead className={STYLE.tableCellMuted}>ペット番号</TableHead>
      <TableHead className={STYLE.tableCellMuted}>ペット名</TableHead>
      <TableHead className={STYLE.tableCellMuted}>生死</TableHead>
      <TableHead className={STYLE.tableCellMuted}>種別</TableHead>
      <TableHead className={STYLE.tableCellMuted}>性別</TableHead>
      <TableHead className={STYLE.tableCellMuted}>生年月日</TableHead>
      <TableHead className={STYLE.tableCellMuted}>毛色</TableHead>
      <TableHead className={STYLE.tableCellMuted}>体重</TableHead>
      <TableHead className={STYLE.tableCellMuted}>環境</TableHead>
      <TableHead className={STYLE.tableCellMuted}>備考</TableHead>
      <TableHead className={STYLE.tableCellMuted}>操作</TableHead>
    </TableRow>
  </TableHeader>
);

interface OwnerPetsSectionProps {
  pets: PetFormData[];
  ownerId: string | undefined;
  canEdit: boolean;
  canCreate: boolean;
  canDelete: boolean;
  onAddPet: () => void;
  onEditPet: (pet: PetFormData) => void;
  onDeleteRequest: (id: string, name: string) => void;
}

export function OwnerPetsSection({
  pets,
  ownerId,
  canEdit,
  canCreate,
  canDelete,
  onAddPet,
  onEditPet,
  onDeleteRequest,
}: OwnerPetsSectionProps) {
  return (
    <div className="mb-4 space-y-3">
      <div className="flex items-center justify-between">
        <h2 className={`text-sm font-bold ${C.text} flex items-center gap-2`}>
          <PawPrint className={`${ICON.action} ${C.text60}`} />
          ペット情報
        </h2>
        {canEdit ? (
          <Button
            type="button"
            size="sm"
            onClick={onAddPet}
            className={`${STYLE.confirmPrimary} gap-1.5 text-sm px-4`}
          >
            <Plus className={ICON.action} />
            ペット追加
          </Button>
        ) : null}
      </div>

      <div className={`rounded-lg ${C.bgWhite} overflow-hidden border ${C.borderMedium}`}>
        <Table>
          {PET_TABLE_HEADER}
          <TableBody>
            {pets.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={11}
                  className={`text-center py-8 text-sm ${C.text60}`}
                >
                  ペット情報がありません。「ペット追加」ボタンから追加してください。
                </TableCell>
              </TableRow>
            ) : (
              pets.map((pet) => (
                <PetTableRow
                  key={pet.id}
                  pet={pet}
                  ownerId={ownerId}
                  canEdit={canEdit}
                  canCreate={canCreate}
                  canDelete={canDelete}
                  onEdit={onEditPet}
                  onDeleteRequest={onDeleteRequest}
                />
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
