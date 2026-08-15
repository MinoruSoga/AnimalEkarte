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
      <button type="button" onClick={() => onDirtyChange(true)}>親変更</button>
      <button type="button" onClick={() => onDirtyChange(false)}>親保存</button>
      {details}
    </>
  ),
}));

vi.mock("./ExamTypeFieldsEditor", () => ({
  ExamTypeFieldsEditor: ({
    onDirtyChange,
  }: {
    onDirtyChange: (dirty: boolean) => void;
  }) => (
    <>
      <button type="button" onClick={() => onDirtyChange(true)}>項目変更</button>
      <button type="button" onClick={() => onDirtyChange(false)}>項目保存</button>
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
});
