import { lazy } from "react";

export const MasterSelectModal = lazy(() =>
  import("@/components/shared/MasterSelectModal").then((m) => ({ default: m.MasterSelectModal })),
);

export const ConfirmDialog = lazy(() =>
  import("@/components/shared/ConfirmDialog/ConfirmDialog").then((m) => ({
    default: m.ConfirmDialog,
  })),
);
