import { lazy } from "react";

export const StaffSelectionModal = lazy(() =>
  import("./StaffSelectionModal").then((m) => ({ default: m.StaffSelectionModal })),
);

export const VitalsModal = lazy(() =>
  import("./VitalsModal").then((m) => ({ default: m.VitalsModal })),
);

export const OwnerSearchModal = lazy(() =>
  import("@/components/shared/OwnerSearchModal/OwnerSearchModal").then((m) => ({
    default: m.OwnerSearchModal,
  })),
);

export const ExaminationImportDialog = lazy(() =>
  import("./ExaminationImportDialog").then((m) => ({ default: m.ExaminationImportDialog })),
);
