import { memo, useLayoutEffect, useRef } from "react";
import { useNavigate } from "react-router";
import { Bed, Calendar, CreditCard, Edit, FileText, MoreHorizontal, PawPrint, Plus, Scissors, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { DataTableRowButton } from "@/components/shared/DataTable/DataTableRowButton";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { C, ICON, STYLE } from "@/lib/design-tokens";
import { paths } from "@/config/paths";
import { mapPetStatusLabel, PET_GENDER_MAP } from "@/lib/transforms/pet";
import {
  useGetOwnerSharedPets,
  type OwnerSharedPetApiResponse,
} from "../api/get-owner-shared-pets";
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
  const actionBoundaryRef = useRef({
    status: pet.status,
    canEdit,
    canCreate,
    canDelete,
  });
  useLayoutEffect(() => {
    actionBoundaryRef.current = {
      status: pet.status,
      canEdit,
      canCreate,
      canDelete,
    };
  }, [canCreate, canDelete, canEdit, pet.status]);
  const backFrom = ownerId
    ? paths.owners.detail.getHref(ownerId)
    : paths.owners.getHref();

  return (
    <TableRow className={`transition-colors ${C.borderDivider} ${C.hoverBgPage} h-12`}>
      <TableCell className={STYLE.tableCell}>{pet.petNumber}</TableCell>
      <TableCell className={STYLE.tableCell}>
        {canEdit ? (
          <DataTableRowButton
            aria-label={`詳細・編集: ペット ${pet.petName} (ID ${pet.id})`}
            onClick={() => {
              if (actionBoundaryRef.current.canEdit !== true) return;
              onEdit(pet);
            }}
          >
            {pet.petName}
          </DataTableRowButton>
        ) : pet.petName}
      </TableCell>
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
      <TableCell>
        <div className="flex gap-1 justify-end">
          <DropdownMenu>
            <DropdownMenuTrigger
              className={`inline-flex items-center justify-center rounded-xs cursor-pointer ${STYLE.tableActionBtn} ${C.hoverBgLight}`}
              aria-label={`操作メニュー: ペット ${pet.petName} (ID ${pet.id})`}
            >
              <MoreHorizontal className={ICON.page} />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>操作</DropdownMenuLabel>
              {canEdit ? (
                <DropdownMenuItem
                  onClick={() => {
                    if (actionBoundaryRef.current.canEdit !== true) return;
                    onEdit(pet);
                  }}
                >
                  <Edit className={`mr-2 ${ICON.action}`} />
                  詳細・編集
                </DropdownMenuItem>
              ) : null}
              {pet.status === "生存" && canCreate ? (
                <>
                  <DropdownMenuItem
                    onClick={() => {
                      const current = actionBoundaryRef.current;
                      if (current.status !== "生存" || current.canCreate !== true) return;
                      navigate(`${paths.reservations.getHref()}?petId=${pet.id}`);
                    }}
                  >
                    <Calendar className={`mr-2 ${ICON.action}`} />
                    予約作成
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => {
                      const current = actionBoundaryRef.current;
                      if (current.status !== "生存" || current.canCreate !== true) return;
                      navigate(`${paths.medicalRecords.new.getHref()}?petId=${pet.id}`, { state: { from: backFrom } });
                    }}
                  >
                    <FileText className={`mr-2 ${ICON.action}`} />
                    カルテ作成
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => {
                      const current = actionBoundaryRef.current;
                      if (current.status !== "生存" || current.canCreate !== true) return;
                      navigate(`${paths.trimming.new.getHref()}?petId=${pet.id}`, { state: { from: backFrom } });
                    }}
                  >
                    <Scissors className={`mr-2 ${ICON.action}`} />
                    トリミング
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => {
                      const current = actionBoundaryRef.current;
                      if (current.status !== "生存" || current.canCreate !== true) return;
                      navigate(`${paths.hospitalization.new.getHref()}?petId=${pet.id}`, { state: { from: backFrom } });
                    }}
                  >
                    <Bed className={`mr-2 ${ICON.action}`} />
                    入院登録
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => {
                      const current = actionBoundaryRef.current;
                      if (current.status !== "生存" || current.canCreate !== true) return;
                      navigate(`${paths.accounting.new.getHref()}?petId=${pet.id}`, { state: { from: backFrom } });
                    }}
                  >
                    <CreditCard className={`mr-2 ${ICON.action}`} />
                    会計登録
                  </DropdownMenuItem>
                </>
              ) : null}
              {pet.status === "生存" && canDelete ? (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem
                    onClick={() => {
                      const current = actionBoundaryRef.current;
                      if (current.status !== "生存" || current.canDelete !== true) return;
                      onDeleteRequest(pet.id, pet.petName);
                    }}
                    className={`${C.danger} ${C.focusDanger} ${C.focusBgLight}`}
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

const SharedPetTableRow = memo(function SharedPetTableRow({
  pet,
}: {
  pet: OwnerSharedPetApiResponse;
}) {
  return (
    <TableRow className={`transition-colors ${C.borderDivider} ${C.hoverBgPage} h-12`}>
      <TableCell className={STYLE.tableCell}>{pet.pet_number}</TableCell>
      <TableCell className={STYLE.tableCell}>
        <div className="flex items-center gap-2">
          <span>{pet.name}</span>
          <Badge variant="outline">副飼主</Badge>
          {pet.relationship !== "" ? <span>{pet.relationship}</span> : null}
        </div>
      </TableCell>
      <TableCell className={STYLE.tableCell}>
        {mapPetStatusLabel(pet.status)}
      </TableCell>
      <TableCell className={STYLE.tableCell}>{pet.animal_species.name}</TableCell>
      <TableCell className={STYLE.tableCell}>
        {PET_GENDER_MAP[pet.gender] ?? pet.gender}
      </TableCell>
      <TableCell className={STYLE.tableCell}>
        {pet.birth_date ? pet.birth_date.slice(0, 10) : ""}
      </TableCell>
      <TableCell className={STYLE.tableCell}>{pet.color}</TableCell>
      <TableCell className={STYLE.tableCell}>
        {pet.weight !== null ? `${pet.weight} kg` : ""}
      </TableCell>
      <TableCell className={STYLE.tableCell}>{pet.environment}</TableCell>
      <TableCell className={`${STYLE.tableCell} truncate max-w-[200px]`}>
        {pet.remarks}
      </TableCell>
      <TableCell />
    </TableRow>
  );
});

// DESIGN.md ex-data-table-cell: header は canvas-soft 背景（既存の C.bgPage）+ eyebrow 相当タイポグラフィ。
// STYLE.tableCellMuted（body 用）ではなく既存の STYLE.sectionLabel（eyebrow role）を再利用する。
const PET_TABLE_HEADER_CELL = `${STYLE.sectionLabel} whitespace-nowrap`;

const PET_TABLE_HEADER = (
  <TableHeader>
    <TableRow className={`hover:bg-transparent ${C.bgPage} border-b ${C.borderMedium} h-12`}>
      <TableHead className={PET_TABLE_HEADER_CELL}>ペット番号</TableHead>
      <TableHead className={PET_TABLE_HEADER_CELL}>ペット名</TableHead>
      <TableHead className={PET_TABLE_HEADER_CELL}>生死</TableHead>
      <TableHead className={PET_TABLE_HEADER_CELL}>種別</TableHead>
      <TableHead className={PET_TABLE_HEADER_CELL}>性別</TableHead>
      <TableHead className={PET_TABLE_HEADER_CELL}>生年月日</TableHead>
      <TableHead className={PET_TABLE_HEADER_CELL}>毛色</TableHead>
      <TableHead className={PET_TABLE_HEADER_CELL}>体重</TableHead>
      <TableHead className={PET_TABLE_HEADER_CELL}>環境</TableHead>
      <TableHead className={PET_TABLE_HEADER_CELL}>備考</TableHead>
      <TableHead className={PET_TABLE_HEADER_CELL}>操作</TableHead>
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
  const {
    data: sharedPetsResponse,
    isError: isSharedPetsError,
    isLoading: isSharedPetsLoading,
  } = useGetOwnerSharedPets(ownerId);
  const sharedPets = sharedPetsResponse?.shared_pets ?? [];
  const showEmptyState =
    pets.length === 0 &&
    sharedPets.length === 0 &&
    !isSharedPetsLoading &&
    !isSharedPetsError;

  return (
    <div className="mb-4 space-y-3">
      <div className="flex items-center justify-between">
        <h2 className={`text-sm font-bold ${C.text} flex items-center gap-2`}>
          <PawPrint className={`${ICON.action} ${C.text60}`} />
          ペット情報
        </h2>
        {canCreate ? (
          <Button
            type="button"
            size="sm"
            onClick={onAddPet}
            // docs/spec/design-system.md button-primary: brand と同じ primary teal + pill
            className={`${C.bgActionPrimary} ${C.textOnActionPrimary} ${C.hoverBgActionPrimary} ${C.hoverTextOnActionPrimary} gap-1.5 text-sm px-4 rounded-full transition-colors shadow-none border-transparent`}
          >
            <Plus className={ICON.action} />
            ペット追加
          </Button>
        ) : null}
      </div>

      {isSharedPetsError ? (
        <p role="alert" className={`text-sm ${C.danger}`}>
          共有ペット情報の取得に失敗しました。
        </p>
      ) : null}

      <div className={`rounded-lg ${C.bgWhite} overflow-hidden border ${C.borderMedium}`}>
        <Table>
          {PET_TABLE_HEADER}
          <TableBody>
            {showEmptyState ? (
              <TableRow>
                <TableCell
                  data-empty-state
                  colSpan={11}
                  className={`text-center py-8 text-sm ${C.text60}`}
                >
                  ペット情報がありません。「ペット追加」ボタンから追加してください。
                </TableCell>
              </TableRow>
            ) : (
              <>
                {pets.map((pet) => (
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
                ))}
                {sharedPets.map((pet) => (
                  <SharedPetTableRow key={`shared-${pet.id}`} pet={pet} />
                ))}
              </>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
