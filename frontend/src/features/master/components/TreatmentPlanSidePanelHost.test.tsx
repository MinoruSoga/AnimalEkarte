import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { ExaminationTypeMaster } from "../api/exam-types-master";
import { TreatmentPlanSidePanelHost } from "./TreatmentPlanSidePanelHost";

vi.mock("./TreatmentItemSidePanel", () => ({
  TreatmentItemSidePanel: ({
    onDirtyChange,
    details,
  }: {
    onDirtyChange: (dirty: boolean) => void;
    details: React.ReactNode;
  }) => (
    <>
      <button type="button" onClick={() => onDirtyChange(true)}>
        親変更
      </button>
      <button type="button" onClick={() => onDirtyChange(false)}>
        親保存
      </button>
      {details}
    </>
  ),
}));

vi.mock("./ExamTypeFieldsEditor", () => ({
  ExamTypeFieldsEditor: ({ onDirtyChange }: { onDirtyChange: (dirty: boolean) => void }) => (
    <>
      <button type="button" onClick={() => onDirtyChange(true)}>
        項目変更
      </button>
      <button type="button" onClick={() => onDirtyChange(false)}>
        項目保存
      </button>
    </>
  ),
}));

vi.mock("./CheckupTypeFieldsEditor", () => ({
  CheckupTypeFieldsEditor: ({
    checkupTypeId,
    onDirtyChange,
  }: {
    checkupTypeId: string;
    onDirtyChange: (dirty: boolean) => void;
  }) => (
    <>
      <span data-testid="checkup-fields-editor">{checkupTypeId}</span>
      <button type="button" onClick={() => onDirtyChange(true)}>
        健診項目変更
      </button>
      <button type="button" onClick={() => onDirtyChange(false)}>
        健診項目保存
      </button>
    </>
  ),
}));

const examinationType: ExaminationTypeMaster = {
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
};

describe("TreatmentPlanSidePanelHost dirty aggregation", () => {
  it("stays dirty until both parent and field drafts are clean", async () => {
    const user = userEvent.setup();
    const onDirtyChange = vi.fn();
    render(
      <TreatmentPlanSidePanelHost
        editTarget={examinationType}
        selectedItem={examinationType}
        parentCandidates={[]}
        hasChildren={false}
        canCreate
        canEdit
        canDelete
        examinationType={examinationType}
        onClose={vi.fn()}
        onSave={vi.fn()}
        onDeleteRequest={vi.fn()}
        onDirtyChange={onDirtyChange}
      />,
    );

    await user.click(screen.getByRole("button", { name: "親変更" }));
    expect(onDirtyChange).toHaveBeenLastCalledWith(true);
    await user.click(screen.getByRole("button", { name: "項目変更" }));
    await user.click(screen.getByRole("button", { name: "親保存" }));
    expect(onDirtyChange).toHaveBeenLastCalledWith(true);
    await user.click(screen.getByRole("button", { name: "項目保存" }));
    expect(onDirtyChange).toHaveBeenLastCalledWith(false);
  });

  it("mounts the checkup field editor for checkup items and aggregates its dirty state", async () => {
    const user = userEvent.setup();
    const onDirtyChange = vi.fn();
    const checkupItem = {
      id: "9",
      name: "年次健診",
      price: 12000,
      isActive: true,
      description: "",
      sortOrder: 1,
      masterKind: "checkup" as const,
    };
    render(
      <TreatmentPlanSidePanelHost
        editTarget={checkupItem}
        selectedItem={checkupItem}
        parentCandidates={[]}
        hasChildren={false}
        canCreate
        canEdit
        canDelete
        onClose={vi.fn()}
        onSave={vi.fn()}
        onDeleteRequest={vi.fn()}
        onDirtyChange={onDirtyChange}
      />,
    );

    // EMR-225: masterKind="checkup" の行は健診フィールド定義エディタをぶら下げる。
    expect(screen.getByTestId("checkup-fields-editor")).toHaveTextContent("9");
    await user.click(screen.getByRole("button", { name: "健診項目変更" }));
    expect(onDirtyChange).toHaveBeenLastCalledWith(true);
    await user.click(screen.getByRole("button", { name: "親保存" }));
    expect(onDirtyChange).toHaveBeenLastCalledWith(true);
    await user.click(screen.getByRole("button", { name: "健診項目保存" }));
    expect(onDirtyChange).toHaveBeenLastCalledWith(false);
  });

  it("does not mount the checkup field editor for non-checkup items", () => {
    const plainItem = {
      id: "1",
      name: "初診",
      price: 1000,
      isActive: true,
      description: "",
      sortOrder: 1,
    };
    render(
      <TreatmentPlanSidePanelHost
        editTarget={plainItem}
        selectedItem={plainItem}
        parentCandidates={[]}
        hasChildren={false}
        canCreate
        canEdit
        canDelete
        onClose={vi.fn()}
        onSave={vi.fn()}
        onDeleteRequest={vi.fn()}
        onDirtyChange={vi.fn()}
      />,
    );
    expect(screen.queryByTestId("checkup-fields-editor")).not.toBeInTheDocument();
  });
});
