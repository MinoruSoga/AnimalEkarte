import { memo } from "react";
import { AnimatePresence, motion } from "motion/react";

import { LAYOUT } from "@/lib/design-tokens";
import type { Medicine } from "@/types";

import { MedicineSidePanelBody } from "./MedicineSidePanelBody";
import type { MedicineFormData } from "./medicine-side-panel-model";

export type { MedicineFormData } from "./medicine-side-panel-model";

interface MedicineSidePanelProps {
  isEditing: boolean;
  selectedMedicine: Medicine | null;
  isCategory: boolean;
  defaultParentId?: string;
  categoryMedicines: Medicine[];
  panelDuration: number;
  onCloseEdit: () => void;
  onSave: (data: MedicineFormData) => Promise<boolean> | boolean;
  onDeleteRequest: () => void;
  readOnly?: boolean;
  canDelete?: boolean;
  onDirtyChange?: (dirty: boolean) => void;
}

export const MedicineSidePanel = memo(function MedicineSidePanel({
  isEditing,
  selectedMedicine,
  isCategory,
  defaultParentId,
  categoryMedicines,
  panelDuration,
  onCloseEdit,
  onSave,
  onDeleteRequest,
  readOnly,
  canDelete,
  onDirtyChange,
}: MedicineSidePanelProps) {
  return (
    <AnimatePresence>
      {isEditing ? (
        <motion.div
          key="side-peek"
          initial={{ width: 0, opacity: 0 }}
          animate={{ width: LAYOUT.sidePeek.widthPx, opacity: 1 }}
          exit={{ width: 0, opacity: 0 }}
          transition={{ duration: panelDuration, ease: [0.25, 0.1, 0.25, 1] }}
          className="shrink-0 min-h-0 overflow-hidden"
        >
          <MedicineSidePanelBody
            key={selectedMedicine?.id ?? "new"}
            selectedMedicine={selectedMedicine}
            isCategory={isCategory}
            defaultParentId={defaultParentId}
            categoryMedicines={categoryMedicines}
            onCloseEdit={onCloseEdit}
            onSave={onSave}
            onDeleteRequest={onDeleteRequest}
            readOnly={readOnly}
            canDelete={canDelete}
            onDirtyChange={onDirtyChange}
          />
        </motion.div>
      ) : null}
    </AnimatePresence>
  );
});
