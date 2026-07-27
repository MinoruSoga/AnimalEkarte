import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ResourceCheckups, ResourceMasterMedical } from "@/types/generated/models";
import { TreatmentPlanMaster } from "./TreatmentPlanMaster";

const mocks = vi.hoisted(() => ({
  masterTabPageProps: [] as Array<{ resource?: string }>,
  sidePanelProps: [] as Array<{
    canEdit: boolean;
    canDelete: boolean;
    canCreate: boolean;
    examinationTypeId?: string;
  }>,
  tabContentProps: [] as Array<{
    entityLabel: string;
    canEdit: boolean;
    onReorder: (ids: number[]) => void;
  }>,
  checkupReorder: vi.fn(),
  medicalPermissions: {
    canCreate: true,
    canEdit: false,
    canDelete: true,
  },
  checkupPermissions: {
    canCreate: false,
    canEdit: true,
    canDelete: false,
  },
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: (resource: string) =>
    resource === ResourceCheckups
      ? mocks.checkupPermissions
      : mocks.medicalPermissions,
}));

vi.mock("@/hooks/use-side-peek-dirty", () => ({
  useSidePeekDirty: () => ({
    markDirty: vi.fn(),
    markClean: vi.fn(),
    confirmDiscard: vi.fn(() => true),
  }),
}));

vi.mock("../hooks/use-master-save", () => ({
  useMasterSave: () => ({ handleSave: vi.fn() }),
}));

vi.mock("../components/MasterTabPage", () => ({
  MasterTabPage: ({
    resource,
    sidePanel,
    children,
  }: {
    resource?: string;
    sidePanel: ReactNode;
    children: ReactNode;
  }) => {
    mocks.masterTabPageProps.push({ resource });
    return (
      <>
        {sidePanel}
        {children}
      </>
    );
  },
}));

vi.mock("../components/TreatmentPlanSidePanelHost", () => ({
  TreatmentPlanSidePanelHost: ({
    canEdit,
    canDelete,
    canCreate,
    examinationType,
  }: {
    canEdit: boolean;
    canDelete: boolean;
    canCreate: boolean;
    examinationType?: { id: string };
  }) => {
    mocks.sidePanelProps.push({
      canEdit,
      canDelete,
      canCreate,
      examinationTypeId: examinationType?.id,
    });
    return null;
  },
}));

vi.mock("../components/TreatmentPlanTabContent", () => ({
  TreatmentPlanTabContent: ({
    entityLabel,
    canEdit,
    onReorder,
    data,
    onEditTargetChange,
  }: {
    entityLabel: string;
    canEdit: boolean;
    onReorder: (ids: number[]) => void;
    data: Array<{ id: string; name: string }>;
    onEditTargetChange: (item: { id: string; name: string }) => void;
  }) => {
    mocks.tabContentProps.push({ entityLabel, canEdit, onReorder });
    return entityLabel === "検査" && data[0] ? (
      <button type="button" onClick={() => onEditTargetChange(data[0])}>
        検査を選択
      </button>
    ) : null;
  },
}));

vi.mock("../components/TreatmentPlanDeleteDialog", () => ({
  TreatmentPlanDeleteDialog: () => null,
}));

vi.mock("@/components/shared/UnifiedTabs", () => ({
  UnifiedTabs: ({ children }: { children: ReactNode }) => <>{children}</>,
  UnifiedTabsContent: ({ children }: { children: ReactNode }) => <>{children}</>,
}));

function mutationStub(mutate = vi.fn()) {
  return {
    mutate,
    mutateAsync: vi.fn(),
  };
}

vi.mock("../api/consultations", () => ({
  useGetAllConsultations: () => ({ data: [] }),
  useCreateConsultation: mutationStub,
  useUpdateConsultation: mutationStub,
  useDeleteConsultation: mutationStub,
  useReorderConsultations: mutationStub,
}));

vi.mock("../api/exam-types-master", () => ({
  useGetAllExaminationTypes: () => ({ data: [{
    id: "3",
    name: "血液検査",
    price: 1000,
    isActive: true,
    description: "",
    sortOrder: 1,
    isNonInsurance: false,
    createdAt: "",
    updatedAt: "",
    items: [],
  }] }),
  useCreateExaminationType: mutationStub,
  useUpdateExaminationType: mutationStub,
  useDeleteExaminationType: mutationStub,
  useReorderExaminationTypes: mutationStub,
}));

vi.mock("../api/procedures", () => ({
  useGetAllProcedures: () => ({ data: [] }),
  useCreateProcedure: mutationStub,
  useUpdateProcedure: mutationStub,
  useDeleteProcedure: mutationStub,
  useReorderProcedures: mutationStub,
}));

vi.mock("../api/vaccines-master", () => ({
  useGetAllVaccinesMaster: () => ({ data: [] }),
  useCreateVaccineMaster: mutationStub,
  useUpdateVaccineMaster: mutationStub,
  useDeleteVaccineMaster: mutationStub,
  useReorderVaccinesMaster: mutationStub,
}));

vi.mock("../api/checkup-types", () => ({
  useGetAllCheckupTypes: () => ({ data: [] }),
  useCreateCheckupType: mutationStub,
  useUpdateCheckupType: mutationStub,
  useDeleteCheckupType: mutationStub,
  useReorderCheckupTypes: () => mutationStub(mocks.checkupReorder),
}));

function renderRoute(path: string) {
  render(
    <MemoryRouter initialEntries={[path]}>
      <TreatmentPlanMaster />
    </MemoryRouter>,
  );
}

function latestTabContent(entityLabel: string) {
  return mocks.tabContentProps.findLast((props) => props.entityLabel === entityLabel);
}

describe("TreatmentPlanMaster resource-specific permissions", () => {
  beforeEach(() => {
    mocks.masterTabPageProps.length = 0;
    mocks.sidePanelProps.length = 0;
    mocks.tabContentProps.length = 0;
    mocks.checkupReorder.mockClear();
  });

  it("uses checkup resource and permissions for the active checkup tab", () => {
    renderRoute("/settings/treatment-items?tab=checkup");

    expect(mocks.masterTabPageProps.at(-1)?.resource).toBe(ResourceCheckups);
    expect(mocks.sidePanelProps.at(-1)).toEqual({
      canEdit: mocks.checkupPermissions.canEdit,
      canDelete: mocks.checkupPermissions.canDelete,
      canCreate: mocks.checkupPermissions.canCreate,
      examinationTypeId: undefined,
    });
    expect(latestTabContent("定期健診")?.canEdit).toBe(
      mocks.checkupPermissions.canEdit,
    );
    expect(latestTabContent("診察")?.canEdit).toBe(
      mocks.medicalPermissions.canEdit,
    );
    latestTabContent("定期健診")?.onReorder([2, 1]);
    expect(mocks.checkupReorder).toHaveBeenCalledWith({ ids: [2, 1] });
  });

  it("preserves medical resource and permissions for the consultation tab", () => {
    renderRoute("/settings/treatment-items?tab=consultation");

    expect(mocks.masterTabPageProps.at(-1)?.resource).toBe(ResourceMasterMedical);
    expect(mocks.sidePanelProps.at(-1)).toEqual({
      canEdit: mocks.medicalPermissions.canEdit,
      canDelete: mocks.medicalPermissions.canDelete,
      canCreate: mocks.medicalPermissions.canCreate,
      examinationTypeId: undefined,
    });
    expect(latestTabContent("診察")?.canEdit).toBe(
      mocks.medicalPermissions.canEdit,
    );
    expect(latestTabContent("定期健診")?.canEdit).toBe(
      mocks.checkupPermissions.canEdit,
    );
  });

  it("mounts the field editor data only for the active examination tab", async () => {
    const user = userEvent.setup();
    renderRoute("/settings/treatment-items?tab=examination");

    await user.click(screen.getByRole("button", { name: "検査を選択" }));

    expect(mocks.sidePanelProps.at(-1)?.examinationTypeId).toBe("3");
    expect(mocks.sidePanelProps.at(-1)?.canCreate).toBe(
      mocks.medicalPermissions.canCreate,
    );
  });
});
