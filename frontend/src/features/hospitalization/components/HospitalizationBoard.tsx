// React/Framework
import { C, ICON } from "@/lib/design-tokens";
import { formatDate } from "@/lib/format/date";
import { memo, useCallback, useMemo } from "react";

// External
import {
  closestCorners,
  DndContext,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import { useDraggable, useDroppable } from "@dnd-kit/core";
import { Plus, GripVertical } from "lucide-react";
import { cageKeyboardCoordinateGetter } from "./cage-keyboard-coordinates";

// Internal
import { Card, CardHeader, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { getHospitalizationTypeColor } from "@/lib/status-helpers";

// Relative
import { H_STYLES } from "../styles";

// Types
import type { MasterItem, Hospitalization } from "@/types";

interface HospitalizationBoardProps {
  cages: MasterItem[];
  hospitalizations: Hospitalization[];
  onNavigateToForm: (id?: string) => void;
  onMovePet: (hospitalizationId: string, targetCageId: string) => void;
  canCreate: boolean;
  canEdit: boolean;
}

interface CageCardProps {
    cage: MasterItem;
    occupant?: Hospitalization;
    onNavigateToForm: (id?: string) => void;
    canCreate: boolean;
    canEdit: boolean;
}

const CageCard = memo(function CageCard({ cage, occupant, onNavigateToForm, canCreate, canEdit }: CageCardProps) {
    const isDeceased = occupant?.petIsDeceased ?? false;
    const cageContext = cage.category ? `${cage.category} ${cage.name}` : cage.name;
    const emptyCageActionLabel = `${cageContext}（ケージID: ${cage.id}）の空き枠に入院・ホテルを登録`;
    const canDrag = Boolean(occupant) && !isDeceased && canEdit;
    const canOpenCard = Boolean(occupant) && !isDeceased && canEdit;

    const {
        attributes,
        listeners,
        setNodeRef: setDragRef,
        isDragging,
    } = useDraggable({
        id: occupant?.id ?? `empty-${cage.id}`,
        data: { hospitalizationId: occupant?.id },
        disabled: !canDrag,
    });

    const { setNodeRef: setDropRef, isOver } = useDroppable({
        id: `cage-${cage.id}`,
        data: { cageId: cage.id },
        disabled: !canEdit,
    });

    return (
        <div ref={setDropRef} className="h-full">
            <Card
                ref={setDragRef}
                {...(canDrag ? attributes : {})}
                {...(canDrag ? listeners : {})}
                className={`relative flex flex-col h-40 transition-all border touch-none
                  ${occupant
                      ? isDeceased
                        ? `${C.bgPage} border-l-4 ${C.borderPrimary20} opacity-40`
                        : `${C.bgWhite} border-l-4 ${C.borderLMedicalBlue}`
                      : `${C.bgPage} border-dashed ${C.borderPrimary20}`
                  }
                  ${isDragging ? 'opacity-50 scale-95' : 'hover:shadow-level1'}
                  ${isOver ? `ring-2 ${C.ringMedicalBlue} ring-offset-2 ${C.bgMedicalBlue5}` : ''}
                  ${canOpenCard ? 'cursor-pointer' : 'cursor-default'}
                `}
                onClick={
                  canOpenCard && occupant ? () => onNavigateToForm(occupant.id) : undefined
                }
            >
                <CardHeader className={`${H_STYLES.padding.card} pb-0 flex flex-row items-center justify-between space-y-0`}>
                  <div className="flex items-center gap-1">
                      {canDrag ? (
                          <div className={`cursor-grab active:cursor-grabbing ${C.text20} ${C.hoverText60}`}>
                              <GripVertical className={ICON.action} />
                          </div>
                      ) : null}
                      <span className={`${H_STYLES.text.sm} font-mono ${C.text60} font-bold`}>{cage.name}</span>
                  </div>
                  {occupant ? (
                      <Badge variant="outline" className={`${getHospitalizationTypeColor(occupant.hospitalizationType)} ${H_STYLES.text.xs} px-1.5 py-0 h-5 border-none`}>
                          {occupant.hospitalizationType}
                      </Badge>
                  ) : null}
                </CardHeader>
                <CardContent className={`${H_STYLES.padding.card} flex-1 flex flex-col justify-center items-start text-left`}>
                  {occupant ? (
                    <>
                      {occupant.startDate ? (
                        <div className={`text-xs font-mono ${C.text60} w-full`}>
                          {formatDate(occupant.startDate)}
                        </div>
                      ) : null}
                      <div className={`font-bold ${C.text} ${H_STYLES.text.sm} truncate w-full`} title={occupant.ownerName}>
                          {occupant.ownerName}
                      </div>
                      <div className={`${H_STYLES.text.sm} ${C.text} truncate w-full`}>
                          {[occupant.species, occupant.petName].filter(Boolean).join(" ")}
                      </div>
                      {isDeceased ? (
                        <span className={`text-xs ${C.text40} font-medium`}>死亡</span>
                      ) : null}
                      {canOpenCard ? (
                        <button
                          type="button"
                          aria-label={`${occupant.petName}の詳細`}
                          className={`mt-2 flex min-h-11 items-center justify-center gap-1 text-2xs ${C.textBrand} ${C.bgBrandLight30} border ${C.borderBrandLight} rounded px-1.5 ${C.hoverBgBrandLight60} transition-colors`}
                          onClick={(event) => {
                            event.stopPropagation();
                            onNavigateToForm(occupant.id);
                          }}
                        >
                          詳細
                        </button>
                      ) : null}
                    </>
                  ) : (
                    <div className={`flex flex-col items-center justify-center h-full ${C.text20}`}>
                       <span className={H_STYLES.text.sm}>空き</span>
                       {canCreate ? (
                         <Button
                            variant="ghost"
                            size="icon"
                            aria-label={emptyCageActionLabel}
                            title={emptyCageActionLabel}
                            className={`h-10 w-10 mt-1 rounded-full ${C.hoverBgPrimary10} ${C.hoverText60}`}
                            onClick={(e) => {
                                e.stopPropagation();
                                onNavigateToForm();
                            }}
                         >
                           <Plus className={H_STYLES.button.icon} />
                         </Button>
                       ) : null}
                    </div>
                  )}
                </CardContent>
            </Card>
        </div>
    );
});

export const HospitalizationBoard = memo(function HospitalizationBoard({ cages, hospitalizations, onNavigateToForm, onMovePet, canCreate, canEdit }: HospitalizationBoardProps) {
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
    useSensor(KeyboardSensor, { coordinateGetter: cageKeyboardCoordinateGetter }),
  );

  const handleDragEnd = useCallback((event: DragEndEvent) => {
    if (!canEdit) return;
    const { active, over } = event;
    if (!over) return;

    const hospitalizationId = active.data.current?.hospitalizationId as string | undefined;
    if (!hospitalizationId) return;

    const overId = over.id as string;
    if (!overId.startsWith("cage-")) return;

    const targetCageId = overId.replace("cage-", "");
    onMovePet(hospitalizationId, targetCageId);
  }, [canEdit, onMovePet]);

  // Group cages by category (Area)
  const cagesByArea = cages.reduce((acc, cage) => {
    const area = cage.category || "その他";
    if (!acc[area]) acc[area] = [];
    acc[area].push(cage);
    return acc;
  }, {} as Record<string, MasterItem[]>);

  // js-index-maps: cageId → Hospitalization の Map を事前構築（O(n)）しレンダーループ内で O(1) 検索
  // 親コンポーネントがタブに応じて既にフィルタリング済みのデータを渡すため、ここでは再フィルタしない
  const occupantByCageId = useMemo(
    () => new Map(hospitalizations.map(h => [h.cageId, h])),
    [hospitalizations]
  );

  return (
    <DndContext sensors={sensors} collisionDetection={closestCorners} onDragEnd={handleDragEnd}>
      <div className="space-y-6 pb-4">
        {Object.entries(cagesByArea).map(([area, areaCages]) => (
          // No min-w-[800px]: reflow 1/2/3+ columns by container width (BUG-458).
          <div key={area} className="min-w-0">
            <h3 className={`${H_STYLES.text.lg} font-bold ${C.text} mb-3 border-b pb-1 ${C.borderPrimary10}`}>
              {area}
            </h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
              {areaCages.map(cage => {
                const occupant = occupantByCageId.get(String(cage.id));
                return (
                  <CageCard
                      key={cage.id}
                      cage={cage}
                      occupant={occupant}
                      onNavigateToForm={onNavigateToForm}
                      canCreate={canCreate}
                      canEdit={canEdit}
                  />
                );
              })}
            </div>
          </div>
        ))}
      </div>
    </DndContext>
  );
});
