import { useLayoutEffect, useRef, useState, type ReactNode } from "react";
import { act, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { CarePlanTab } from "../CarePlanTab/CarePlanTab";
import { DailyRecordsTab } from "./DailyRecordsTab";
import type {
  CarePlanItem,
  CreateCarePlanItemInput,
  UpdateCarePlanItemInput,
} from "../../api/care-plan-items";
import type {
  CreateCareLogRequest,
  CreateStaffNoteRequest,
  CreateVitalRecordRequest,
} from "../../api/daily-records-types";

const mocks = vi.hoisted(() => ({
  canCreate: true,
  canEdit: true,
  canDelete: true,
  createCarePlanItem: vi.fn(),
  updateCarePlanItem: vi.fn(),
  deleteCarePlanItem: vi.fn(),
  createDailyRecord: vi.fn(),
  createVital: vi.fn(),
  createCareLog: vi.fn(),
  createStaffNote: vi.fn(),
  dailyRecordIsError: false,
  carePlanCreateCallback: undefined as ((input: CreateCarePlanItemInput) => void) | undefined,
  carePlanUpdateCallback: undefined as ((input: UpdateCarePlanItemInput) => void) | undefined,
  carePlanDeleteCallback: undefined as ((itemId: string) => void) | undefined,
  carePlanEditCallback: undefined as ((itemId: string) => void) | undefined,
  vitalCallback: undefined as ((payload: CreateVitalRecordRequest) => Promise<void>) | undefined,
  careLogCallback: undefined as ((payload: CreateCareLogRequest) => void) | undefined,
  staffNoteCallback: undefined as ((payload: CreateStaffNoteRequest) => void) | undefined,
  dailyRecordCallback: undefined as (() => void) | undefined,
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ user: { id: "7" } }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canView: true,
    canCreate: mocks.canCreate,
    canEdit: mocks.canEdit,
    canDelete: mocks.canDelete,
  }),
}));

vi.mock("../../api/care-plan-items", () => ({
  useGetCarePlanItems: () => ({
    data: [
      {
        id: "item-1",
        hospitalization_id: "hospitalization-1",
        type: "instruction",
        name: "既存項目",
        description: "",
        timing: ["morning"],
        status: "active",
        notes: "",
        unit_price: 0,
        category: "",
        sort_order: 0,
        created_at: "",
        updated_at: "",
      } satisfies CarePlanItem,
    ],
    isLoading: false,
  }),
  useCreateCarePlanItem: () => ({ mutateAsync: mocks.createCarePlanItem, isPending: false }),
  useUpdateCarePlanItem: () => ({ mutateAsync: mocks.updateCarePlanItem, isPending: false }),
  useDeleteCarePlanItem: () => ({ mutate: mocks.deleteCarePlanItem, isPending: false }),
}));

vi.mock("../../api/daily-records", () => ({
  useGetDailyRecord: () => ({
    data: {
      id: "daily-1",
      hospitalization_id: "hospitalization-1",
      date: "2026-07-14T00:00:00+09:00",
      created_at: "",
      updated_at: "",
      vital_records: [],
      care_logs: [],
      staff_notes: [],
    },
    isLoading: false,
    isError: mocks.dailyRecordIsError,
  }),
  useCreateDailyRecord: () => ({ mutateAsync: mocks.createDailyRecord }),
  useCreateDailyVital: () => ({ mutateAsync: mocks.createVital }),
  useCreateCareLog: () => ({ mutateAsync: mocks.createCareLog }),
  useCreateStaffNote: () => ({ mutateAsync: mocks.createStaffNote }),
}));

vi.mock("../CarePlanTab/ItemRow", () => ({
  ItemRow: ({
    onEdit,
    onDelete,
  }: {
    onEdit?: (id: string) => void;
    onDelete?: (id: string) => void;
  }) => {
    mocks.carePlanEditCallback = onEdit;
    mocks.carePlanDeleteCallback = onDelete;
    return <button onClick={() => onEdit?.("item-1")}>編集</button>;
  },
}));

vi.mock("../CarePlanTab/EditRow", () => ({
  EditRow: ({ onSave }: { onSave: (input: UpdateCarePlanItemInput) => void }) => {
    mocks.carePlanUpdateCallback = onSave;
    return null;
  },
}));

vi.mock("../CarePlanTab/AddForm", () => ({
  AddForm: ({ onSubmit }: { onSubmit: (input: CreateCarePlanItemInput) => void }) => {
    mocks.carePlanCreateCallback = onSubmit;
    return null;
  },
}));

vi.mock("./DailyDateNav", () => ({ DailyDateNav: () => null }));
vi.mock("@/components/ui/separator", () => ({ Separator: () => null }));
vi.mock("@/components/shared/DataStates", () => ({
  EmptyState: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));
vi.mock("@/components/ui/button", () => ({
  Button: ({ children, onClick }: { children: ReactNode; onClick?: () => void }) => {
    mocks.dailyRecordCallback = onClick;
    return <button onClick={onClick}>{children}</button>;
  },
}));
vi.mock("./DailyVitalsSection", () => ({
  DailyVitalsSection: ({
    onAddVital,
  }: {
    onAddVital: (payload: CreateVitalRecordRequest) => Promise<void>;
  }) => {
    mocks.vitalCallback = onAddVital;
    return null;
  },
}));
vi.mock("./DailyCareLogsSection", () => ({
  DailyCareLogsSection: ({
    onAddCareLog,
  }: {
    onAddCareLog: (payload: CreateCareLogRequest) => void;
  }) => {
    mocks.careLogCallback = onAddCareLog;
    return null;
  },
}));
vi.mock("./DailyStaffNotesSection", () => ({
  DailyStaffNotesSection: ({
    onAddStaffNote,
  }: {
    onAddStaffNote: (payload: CreateStaffNoteRequest) => void;
  }) => {
    mocks.staffNoteCallback = onAddStaffNote;
    return null;
  },
}));

function renderChildMutationBoundaries() {
  return render(
    <>
      <CarePlanTab hospitalizationId="hospitalization-1" petIsDeceased={false} />
      <DailyRecordsTab
        hospitalizationId="hospitalization-1"
        admissionDate="2026-07-01"
        dischargeDate="2026-07-14"
        petIsDeceased={false}
      />
    </>,
  );
}

function SameCommitRevocationHarness() {
  const [revoked, setRevoked] = useState(false);
  const capturedCreateRef = useRef<((input: CreateCarePlanItemInput) => void) | undefined>(
    undefined,
  );
  const capturedVitalRef = useRef<((payload: CreateVitalRecordRequest) => void) | undefined>(
    undefined,
  );

  useLayoutEffect(() => {
    if (revoked) {
      capturedCreateRef.current?.({
        type: "instruction",
        name: "追加",
        timing: ["morning"],
      });
      capturedVitalRef.current?.({
        time: "10:00:00",
        temperature: 38.5,
      });
      return;
    }

    capturedCreateRef.current = mocks.carePlanCreateCallback;
    capturedVitalRef.current = mocks.vitalCallback;
  }, [revoked]);

  return (
    <>
      <button
        type="button"
        onClick={() => {
          mocks.canCreate = false;
          setRevoked(true);
        }}
      >
        作成権限を失効
      </button>
      <CarePlanTab
        hospitalizationId={revoked ? "hospitalization-2" : "hospitalization-1"}
        petIsDeceased={false}
      />
      <DailyRecordsTab
        hospitalizationId={revoked ? "hospitalization-2" : "hospitalization-1"}
        admissionDate="2026-07-01"
        dischargeDate="2026-07-14"
        petIsDeceased={false}
      />
    </>
  );
}

beforeEach(() => {
  mocks.canCreate = true;
  mocks.canEdit = true;
  mocks.canDelete = true;
  mocks.dailyRecordIsError = false;
  mocks.createCarePlanItem.mockReset();
  mocks.updateCarePlanItem.mockReset();
  mocks.deleteCarePlanItem.mockReset();
  mocks.createDailyRecord.mockReset();
  mocks.createVital.mockReset();
  mocks.createCareLog.mockReset();
  mocks.createStaffNote.mockReset();
  mocks.carePlanCreateCallback = undefined;
  mocks.carePlanUpdateCallback = undefined;
  mocks.carePlanDeleteCallback = undefined;
  mocks.carePlanEditCallback = undefined;
  mocks.vitalCallback = undefined;
  mocks.careLogCallback = undefined;
  mocks.staffNoteCallback = undefined;
  mocks.dailyRecordCallback = undefined;
});

describe("hospitalization child mutation permission boundaries", () => {
  it("petが死亡している場合はcare-planとdaily-recordの全mutationを拒否する", () => {
    mocks.dailyRecordIsError = true;
    render(
      <>
        <CarePlanTab hospitalizationId="hospitalization-1" petIsDeceased />
        <DailyRecordsTab
          hospitalizationId="hospitalization-1"
          admissionDate="2026-07-01"
          dischargeDate="2026-07-14"
          petIsDeceased
        />
      </>,
    );

    act(() => {
      mocks.carePlanCreateCallback?.({
        type: "instruction",
        name: "追加",
        timing: ["morning"],
      });
      mocks.carePlanEditCallback?.("item-1");
      mocks.carePlanDeleteCallback?.("item-1");
      mocks.vitalCallback?.({ time: "10:00:00", temperature: 38.5 });
      mocks.careLogCallback?.({ time: "10:01:00", type: "food" });
      mocks.staffNoteCallback?.({ time: "10:02:00", content: "メモ" });
      mocks.dailyRecordCallback?.();
    });
    act(() => {
      mocks.carePlanUpdateCallback?.({
        type: "instruction",
        name: "更新",
        timing: ["morning"],
      });
    });

    expect(mocks.createCarePlanItem).not.toHaveBeenCalled();
    expect(mocks.updateCarePlanItem).not.toHaveBeenCalled();
    expect(mocks.deleteCarePlanItem).not.toHaveBeenCalled();
    expect(mocks.createDailyRecord).not.toHaveBeenCalled();
    expect(mocks.createVital).not.toHaveBeenCalled();
    expect(mocks.createCareLog).not.toHaveBeenCalled();
    expect(mocks.createStaffNote).not.toHaveBeenCalled();
  });

  it("permission revocation blocks captured care-plan callbacks", () => {
    const view = renderChildMutationBoundaries();
    const createCallback = mocks.carePlanCreateCallback;
    const editCallback = mocks.carePlanEditCallback;
    const deleteCallback = mocks.carePlanDeleteCallback;

    act(() => {
      editCallback?.("item-1");
    });
    const updateCallback = mocks.carePlanUpdateCallback;

    mocks.canCreate = false;
    mocks.canEdit = false;
    mocks.canDelete = false;
    view.rerender(
      <>
        <CarePlanTab hospitalizationId="hospitalization-2" petIsDeceased={false} />
        <DailyRecordsTab
          hospitalizationId="hospitalization-2"
          admissionDate="2026-07-01"
          dischargeDate="2026-07-14"
          petIsDeceased={false}
        />
      </>,
    );

    act(() => {
      createCallback?.({ type: "instruction", name: "追加", timing: ["morning"] });
      updateCallback?.({ type: "instruction", name: "更新", timing: ["morning"] });
      deleteCallback?.("item-1");
      mocks.vitalCallback?.({ time: "10:00:00", temperature: 38.5 });
      mocks.careLogCallback?.({ time: "10:01:00", type: "food" });
      mocks.staffNoteCallback?.({ time: "10:02:00", content: "メモ" });
    });

    expect(mocks.createCarePlanItem).not.toHaveBeenCalled();
    expect(mocks.updateCarePlanItem).not.toHaveBeenCalled();
    expect(mocks.deleteCarePlanItem).not.toHaveBeenCalled();
    expect(mocks.createVital).not.toHaveBeenCalled();
    expect(mocks.createCareLog).not.toHaveBeenCalled();
    expect(mocks.createStaffNote).not.toHaveBeenCalled();
  });

  it("permission granted preserves care-plan and daily mutation payloads", async () => {
    renderChildMutationBoundaries();

    act(() => {
      mocks.carePlanCreateCallback?.({ type: "instruction", name: "追加", timing: ["morning"] });
      mocks.carePlanEditCallback?.("item-1");
      mocks.carePlanDeleteCallback?.("item-1");
      mocks.vitalCallback?.({ time: "10:00:00", temperature: 38.5 });
      mocks.careLogCallback?.({ time: "10:01:00", type: "food" });
      mocks.staffNoteCallback?.({ time: "10:02:00", content: "メモ" });
    });

    const updateCallback = mocks.carePlanUpdateCallback;
    act(() => {
      updateCallback?.({ type: "instruction", name: "更新", timing: ["morning"] });
    });

    await waitFor(() => {
      expect(mocks.createVital).toHaveBeenCalledWith({
        time: "10:00:00",
        temperature: 38.5,
        staff_id: 7,
      });
      expect(mocks.createCareLog).toHaveBeenCalledWith({
        time: "10:01:00",
        type: "food",
        staff_id: 7,
      });
      expect(mocks.createStaffNote).toHaveBeenCalledWith({
        time: "10:02:00",
        content: "メモ",
        staff_id: 7,
      });
    });
    expect(mocks.createCarePlanItem).toHaveBeenCalledWith({
      type: "instruction",
      name: "追加",
      timing: ["morning"],
    });
    expect(mocks.updateCarePlanItem).toHaveBeenCalledWith({
      itemId: "item-1",
      input: { type: "instruction", name: "更新", timing: ["morning"] },
    });
    expect(mocks.deleteCarePlanItem).toHaveBeenCalledWith("item-1", expect.any(Object));
  });

  it("permission granted allows daily-record creation with the existing date payload", async () => {
    mocks.dailyRecordIsError = true;
    render(
      <DailyRecordsTab
        hospitalizationId="hospitalization-1"
        admissionDate="2026-07-01"
        dischargeDate="2026-07-14"
        petIsDeceased={false}
      />,
    );

    const createButton = screen.getByRole("button", { name: "この日の記録を作成" });
    act(() => {
      createButton.click();
    });

    await waitFor(() => expect(mocks.createDailyRecord).toHaveBeenCalledWith("2026-07-14"));
  });

  it("permission revoked after CTA render blocks the captured daily-record callback", () => {
    mocks.dailyRecordIsError = true;
    const view = render(
      <DailyRecordsTab
        hospitalizationId="hospitalization-1"
        admissionDate="2026-07-01"
        dischargeDate="2026-07-14"
        petIsDeceased={false}
      />,
    );
    const dailyRecordCallback = mocks.dailyRecordCallback;

    mocks.canCreate = false;
    view.rerender(
      <DailyRecordsTab
        hospitalizationId="hospitalization-2"
        admissionDate="2026-07-01"
        dischargeDate="2026-07-14"
        petIsDeceased={false}
      />,
    );
    act(() => dailyRecordCallback?.());

    expect(mocks.createDailyRecord).not.toHaveBeenCalled();
  });

  it("same-commit permission revocation blocks captured care-plan and daily callbacks", () => {
    const view = render(<SameCommitRevocationHarness />);

    act(() => {
      screen.getByRole("button", { name: "作成権限を失効" }).click();
    });

    expect(mocks.createCarePlanItem).not.toHaveBeenCalled();
    expect(mocks.createVital).not.toHaveBeenCalled();
    view.unmount();
  });
});
