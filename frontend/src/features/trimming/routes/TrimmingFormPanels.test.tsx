import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { TrimmingFormBody } from "./TrimmingFormPanels";
import type { TrimmingFormData } from "@/types/trimming";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  }),
}));

vi.mock("@/components/shared/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

vi.mock("./TrimmingFormColumns", () => ({
  TrimmingFormColumns: () => <div data-testid="trimming-columns" />,
}));

vi.mock("./TrimmingLazyModals", () => ({
  MasterSelectModal: () => null,
  ConfirmDialog: () => null,
}));

const formData: TrimmingFormData = {
  reservationTypeId: "",
  startTime: "",
  endTime: "",
  styleRequest: "",
  styleImage: null,
  bw: "",
  bwUnit: "Kg",
  bt: "",
  usedShampoo: "",
  usedRibbon: "",
  remarks: "",
  completedImage: null,
  courseId: "",
  optionIds: [],
  staffId: "",
  staffName: "",
  initialStatus: "in_consultation",
  nextScheduleType: "",
  nextDate: "",
};

function renderBody(overrides: Partial<Parameters<typeof TrimmingFormBody>[0]> = {}) {
  const formAction = vi.fn();
  render(
    <TrimmingFormBody
      mode="edit"
      canSubmit={true}
      canDelete={false}
      isSaving={false}
      isImageUploading={false}
      isDeleting={false}
      isDirty={false}
      hasExistingAppointment={false}
      selectedPet={{
        ownerId: "owner-1",
        ownerName: "テスト飼主",
        name: "ポチ",
        status: "生存",
      }}
      formData={formData}
      fieldErrors={{}}
      courses={[]}
      options={[]}
      styleImagePreview={null}
      completedImagePreview={null}
      sortedHistory={[]}
      isHistoryLoading={false}
      historySearchTerm=""
      historySortOrder="desc"
      historyDateRange={{ from: "", to: "" }}
      courseModalOpen={false}
      staffModalOpen={false}
      deleteConfirmOpen={false}
      activeStaffItems={[]}
      formAction={formAction}
      onBack={vi.fn()}
      onFormChange={vi.fn()}
      onOpenCourseModal={vi.fn()}
      onOpenStaffModal={vi.fn()}
      onOpenDeleteConfirm={vi.fn()}
      onCourseModalOpenChange={vi.fn()}
      onStaffModalOpenChange={vi.fn()}
      onCloseDeleteConfirm={vi.fn()}
      onConfirmDelete={vi.fn()}
      onStyleImageChange={vi.fn()}
      onCompletedImageChange={vi.fn()}
      onRemoveStyleImage={vi.fn()}
      onRemoveCompletedImage={vi.fn()}
      onHistorySearchTermChange={vi.fn()}
      onHistorySortOrderChange={vi.fn()}
      onHistoryClear={vi.fn()}
      onHistoryStartDateChange={vi.fn()}
      onHistoryEndDateChange={vi.fn()}
      onHistoryClick={vi.fn()}
      {...overrides}
    />,
  );
  return { formAction };
}

describe("TrimmingFormBody — 既存レコード見出し (PR666)", () => {
  it("新規モードでは見出しがトリミング登録", () => {
    renderBody({ mode: "new" });
    expect(screen.getByRole("heading", { name: "トリミング登録" })).toBeInTheDocument();
  });

  it("既存レコードで canSubmit なら見出しがトリミング編集", () => {
    renderBody({ mode: "edit", canSubmit: true });
    expect(screen.getByRole("heading", { name: "トリミング編集" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
  });

  it("既存レコードで canSubmit=false なら見出しがトリミング詳細で保存ボタンなし", () => {
    const { formAction } = renderBody({ mode: "edit", canSubmit: false });
    expect(screen.getByRole("heading", { name: "トリミング詳細" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(formAction).not.toHaveBeenCalled();
  });

  it("死亡ペットの既存レコードは canSubmit=false でトリミング詳細・保存なし", () => {
    const { formAction } = renderBody({
      mode: "edit",
      canSubmit: false,
      selectedPet: {
        ownerId: "owner-1",
        ownerName: "テスト飼主",
        name: "ポチ",
        status: "死亡",
        deceasedAt: "2026-01-01T00:00:00+09:00",
      },
    });
    expect(screen.getByRole("heading", { name: "トリミング詳細" })).toBeInTheDocument();
    expect(screen.getByRole("status", { name: "死亡ペットのため保存不可" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(formAction).not.toHaveBeenCalled();
  });
});
