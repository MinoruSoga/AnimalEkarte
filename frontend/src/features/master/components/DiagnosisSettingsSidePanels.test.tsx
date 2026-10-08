import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { createTestWrapper } from "@/testing/TestUtils";
import { DiagnosisSettingsSidePanels } from "./DiagnosisSettingsSidePanels";
import type { DiagnosisName, DiagnosisType } from "../api/diagnosis";

const typeFixture: DiagnosisType = {
  id: "1",
  clinicId: 1,
  name: "テストカテゴリ",
  isActive: true,
  description: "",
  sortOrder: 1,
  createdAt: "2026-09-28T00:00:00Z",
  updatedAt: "2026-09-28T00:00:00Z",
};

const nameFixture: DiagnosisName = {
  id: "7",
  clinicId: 1,
  name: "テスト病名",
  isActive: true,
  description: "",
  diagnosisTypeId: "1",
  sortOrder: 1,
  createdAt: "2026-09-28T00:00:00Z",
  updatedAt: "2026-09-28T00:00:00Z",
};

type PanelsProps = React.ComponentProps<typeof DiagnosisSettingsSidePanels>;

const noop = vi.fn();

function baseProps(): PanelsProps {
  return {
    activeTab: "diagnosis_type",
    typeEditTarget: typeFixture,
    typePanelItem: typeFixture,
    nameEditTarget: null,
    namePanelItem: null,
    categories: [typeFixture],
    canDelete: false,
    canEdit: false,
    onTypeClose: noop,
    onTypeSave: vi.fn().mockResolvedValue(true),
    onTypeDeleteRequest: noop,
    onNameClose: noop,
    onNameSave: vi.fn().mockResolvedValue(true),
    onNameDeleteRequest: noop,
    onDirtyChange: noop,
  };
}

function renderPanels(props: PanelsProps) {
  return render(<DiagnosisSettingsSidePanels {...props} />, {
    wrapper: createTestWrapper({ router: true }),
  });
}

describe("DiagnosisSettingsSidePanels (view のみ権限)", () => {
  it("診断カテゴリ: readOnly で開き、保存・削除の導線を出さない", () => {
    renderPanels(baseProps());

    expect(screen.getByText("詳細")).toBeInTheDocument();
    expect(screen.getByDisplayValue("テストカテゴリ")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "保存" })).toBeNull();
    expect(screen.queryByLabelText("削除")).toBeNull();
    expect(screen.getAllByRole("button", { name: "閉じる" }).length).toBeGreaterThan(0);
  });

  it("診断病名: readOnly で開き、保存・削除の導線を出さない", () => {
    renderPanels({
      ...baseProps(),
      activeTab: "diagnosis_name",
      nameEditTarget: nameFixture,
      namePanelItem: nameFixture,
    });

    expect(screen.getByText("詳細")).toBeInTheDocument();
    expect(screen.getByDisplayValue("テスト病名")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "保存" })).toBeNull();
    expect(screen.queryByLabelText("削除")).toBeNull();
    expect(screen.getAllByRole("button", { name: "閉じる" }).length).toBeGreaterThan(0);
  });
});

describe("DiagnosisSettingsSidePanels (編集権限あり)", () => {
  it("診断カテゴリ: 編集モードで保存・削除の導線を従来どおり表示する", () => {
    renderPanels({ ...baseProps(), canEdit: true, canDelete: true });

    expect(screen.getByText("編集")).toBeInTheDocument();
    expect(screen.getByDisplayValue("テストカテゴリ")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
    expect(screen.getByLabelText("削除")).toBeInTheDocument();
  });

  it("診断病名: 編集モードで保存・削除の導線を従来どおり表示する", () => {
    renderPanels({
      ...baseProps(),
      activeTab: "diagnosis_name",
      nameEditTarget: nameFixture,
      namePanelItem: nameFixture,
      canEdit: true,
      canDelete: true,
    });

    expect(screen.getByText("編集")).toBeInTheDocument();
    expect(screen.getByDisplayValue("テスト病名")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
    expect(screen.getByLabelText("削除")).toBeInTheDocument();
  });

  // EMR-319: 必須のカテゴリが未選択で保存すると aria-invalid+エラー文が出る
  it("診断病名: 名称あり・カテゴリ未選択で保存するとカテゴリ欄が aria-invalid+エラー文になる (EMR-319)", async () => {
    const user = userEvent.setup();
    const onNameSave = vi.fn().mockResolvedValue(true);
    const noCategory: DiagnosisName = { ...nameFixture, diagnosisTypeId: "" };
    renderPanels({
      ...baseProps(),
      activeTab: "diagnosis_name",
      nameEditTarget: noCategory,
      namePanelItem: noCategory,
      canEdit: true,
      onNameSave,
    });

    await user.click(screen.getByRole("button", { name: "保存" }));

    await waitFor(() => {
      expect(document.getElementById("diagnosis-category-error")).toHaveTextContent(
        "カテゴリを選択してください",
      );
    });
    const categoryTrigger = screen.getByRole("combobox");
    expect(categoryTrigger).toHaveAttribute("aria-invalid", "true");
    expect(categoryTrigger).toHaveAttribute("aria-describedby", "diagnosis-category-error");
    expect(onNameSave).not.toHaveBeenCalled();
  });
});
