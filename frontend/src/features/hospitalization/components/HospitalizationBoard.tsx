// External
import { Plus, GripVertical } from "lucide-react";
import { useDrag, useDrop } from "react-dnd";

// Internal
import { Card, CardHeader, CardContent } from "../../../components/ui/card";
import { Badge } from "../../../components/ui/badge";
import { Button } from "../../../components/ui/button";
import { getHospitalizationTypeColor } from "../../../lib/status-helpers";

// Relative
import { H_STYLES } from "../styles";

// Types
import type { MasterItem, Hospitalization } from "../../../types";

interface HospitalizationBoardProps {
  cages: MasterItem[];
  hospitalizations: Hospitalization[];
  onNavigateToForm: (id?: string) => void;
  onMovePet: (hospitalizationId: string, targetCageId: string) => void;
}

const ItemTypes = {
  PET: 'pet',
};

interface CageCardProps {
    cage: MasterItem;
    occupant?: Hospitalization;
    onNavigateToForm: (id?: string) => void;
    onMovePet: (hospitalizationId: string, targetCageId: string) => void;
}

const CageCard = ({ cage, occupant, onNavigateToForm, onMovePet }: CageCardProps) => {
    const [{ isDragging }, drag] = useDrag(() => ({
        type: ItemTypes.PET,
        item: { id: occupant?.id },
        canDrag: !!occupant,
        collect: (monitor) => ({
            isDragging: !!monitor.isDragging(),
        }),
    }), [occupant]);

    const [{ isOver, canDrop }, drop] = useDrop(() => ({
        accept: ItemTypes.PET,
        drop: (item: { id: string }) => {
            if (occupant && item.id === occupant.id) return; // Dropped on itself
            onMovePet(item.id, cage.id);
        },
        collect: (monitor) => ({
            isOver: !!monitor.isOver(),
            canDrop: !!monitor.canDrop(),
        }),
    }), [cage.id, occupant, onMovePet]);

    return (
        <div ref={drop} className="h-full">
            <Card 
                ref={drag}
                className={`relative flex flex-col h-40 transition-all border
                  ${occupant 
                      ? "bg-white border-l-4 border-l-[#2EAADC]" 
                      : "bg-[#F7F6F3] border-dashed border-[#37352F]/20"
                  }
                  ${isDragging ? 'opacity-50 scale-95' : 'hover:shadow-md'}
                  ${isOver && canDrop ? 'ring-2 ring-[#2EAADC] ring-offset-2 bg-[#2EAADC]/5' : ''}
                  cursor-pointer
                `}
                onClick={() => onNavigateToForm(occupant?.id)}
            >
                <CardHeader className={`${H_STYLES.padding.card} pb-0 flex flex-row items-center justify-between space-y-0`}>
                  <div className="flex items-center gap-1">
                      {occupant && (
                          <div className="cursor-grab active:cursor-grabbing text-[#37352F]/20 hover:text-[#37352F]/60">
                              <GripVertical className="h-3 w-3" />
                          </div>
                      )}
                      <span className={`${H_STYLES.text.sm} font-mono text-[#37352F]/60 font-bold`}>{cage.name}</span>
                  </div>
                  {occupant && (
                      <Badge variant="outline" className={`${getHospitalizationTypeColor(occupant.hospitalizationType)} ${H_STYLES.text.xs} px-1.5 py-0 h-5 border-none`}>
                          {occupant.hospitalizationType}
                      </Badge>
                  )}
                </CardHeader>
                <CardContent className={`${H_STYLES.padding.card} flex-1 flex flex-col justify-center items-center text-center`}>
                  {occupant ? (
                    <>
                      <div className={`font-bold text-[#37352F] ${H_STYLES.text.lg} truncate w-full`} title={occupant.petName}>
                          {occupant.petName}
                      </div>
                      <div className={`${H_STYLES.text.sm} text-[#37352F]/60 truncate w-full mb-2`}>
                          {occupant.ownerName}
                      </div>
                      <div className={`${H_STYLES.text.xs} text-[#37352F]/40 font-mono`}>
                         {occupant.species}
                      </div>
                    </>
                  ) : (
                    <div className="flex flex-col items-center justify-center h-full text-[#37352F]/20">
                       <span className={H_STYLES.text.sm}>空き</span>
                       <Button 
                          variant="ghost" 
                          size="icon" 
                          className="h-10 w-10 mt-1 rounded-full hover:bg-[#37352F]/10 hover:text-[#37352F]/60"
                          onClick={(e) => {
                              e.stopPropagation();
                              onNavigateToForm();
                          }}
                       >
                          <Plus className={H_STYLES.button.icon} />
                       </Button>
                    </div>
                  )}
                </CardContent>
            </Card>
        </div>
    );
};

export const HospitalizationBoard = ({ cages, hospitalizations, onNavigateToForm, onMovePet }: HospitalizationBoardProps) => {
  // Group cages by category (Area)
  const cagesByArea = cages.reduce((acc, cage) => {
    const area = cage.category || "その他";
    if (!acc[area]) acc[area] = [];
    acc[area].push(cage);
    return acc;
  }, {} as Record<string, MasterItem[]>);

  // Helper to find hospitalization for a cage
  const getOccupant = (cageId: string) => {
    return hospitalizations.find(h => h.cageId === cageId && h.status === "入院中");
  };

  return (
    <div className="space-y-6 overflow-x-auto pb-4">
      {Object.entries(cagesByArea).map(([area, areaCages]) => (
        <div key={area} className="min-w-[800px]">
          <h3 className={`${H_STYLES.text.lg} font-bold text-[#37352F] mb-3 border-b pb-1 border-[#37352F]/10`}>
            {area}
          </h3>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
            {areaCages.map(cage => {
              const occupant = getOccupant(cage.id);
              return (
                <CageCard
                    key={cage.id}
                    cage={cage}
                    occupant={occupant}
                    onNavigateToForm={onNavigateToForm}
                    onMovePet={onMovePet}
                />
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
};
