import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { MemoryRouter, Route, Routes } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LAYOUT } from "@/lib/design-tokens";
import { DiagnosisSettings } from "./DiagnosisSettings";

const pageLayoutPropsSpy = vi.fn();

vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: (props: {
    title: string;
    resource?: string;
    maxWidth?: string;
    children: ReactNode;
  }) => {
    pageLayoutPropsSpy(props);
    return (
      <div data-testid="page-layout-mock" data-max-width={props.maxWidth}>
        <h1>{props.title}</h1>
        {props.children}
      </div>
    );
  },
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canCreate: true,
    canEdit: true,
    canDelete: true,
  }),
}));

vi.mock("@/hooks/use-side-peek-dirty", () => ({
  useSidePeekDirty: () => ({
    markDirty: vi.fn(),
    markClean: vi.fn(),
    confirmDiscard: vi.fn(() => true),
    runWithDiscardCheck: (fn: () => void) => fn(),
    discardDialog: null,
  }),
}));

function mutationStub() {
  return {
    mutate: vi.fn(),
    mutateAsync: vi.fn(),
  };
}

vi.mock("../api/diagnosis", () => ({
  useGetDiagnosisTypes: () => ({
    data: [{ id: "1", name: "一般診断", sortOrder: 1, isActive: true }],
    isPending: false,
    isError: false,
  }),
  useGetDiagnosisNames: () => ({
    data: [{ id: "10", diagnosisTypeId: "1", name: "皮膚炎", sortOrder: 1, isActive: true }],
    isPending: false,
    isError: false,
  }),
  useCreateDiagnosisType: mutationStub,
  useUpdateDiagnosisType: mutationStub,
  useDeleteDiagnosisType: mutationStub,
  useCreateDiagnosisName: mutationStub,
  useUpdateDiagnosisName: mutationStub,
  useDeleteDiagnosisName: mutationStub,
}));

vi.mock("../components/DiagnosisSettingsSidePanels", () => ({
  DiagnosisSettingsSidePanels: () => null,
}));

vi.mock("../components/DiagnosisDeleteDialogs", () => ({
  DiagnosisDeleteDialogs: () => null,
}));

vi.mock("../components/DiagnosisTabs", () => ({
  DiagnosisTypeTab: () => (
    <div data-testid="diagnosis-type-tab-content">カテゴリタブコンテンツ</div>
  ),
  DiagnosisNameTab: () => (
    <div data-testid="diagnosis-name-tab-content">診断病名タブコンテンツ</div>
  ),
}));

function renderDiagnosisSettings(initialEntry = "/settings/diagnosis") {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route path="/settings/diagnosis" element={<DiagnosisSettings />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe("DiagnosisSettings 少列マスタ幅制限（EMR-333）", () => {
  beforeEach(() => {
    pageLayoutPropsSpy.mockClear();
  });

  it("デフォルト（カテゴリタブ）表示時に tableNarrow (max-w-5xl) が PageLayout に渡される", () => {
    renderDiagnosisSettings("/settings/diagnosis");

    expect(pageLayoutPropsSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        title: "診断マスタ",
        resource: "master-medical",
        maxWidth: LAYOUT.pageContentMaxWidth.tableNarrow,
      }),
    );
    expect(screen.getByTestId("page-layout-mock")).toHaveAttribute(
      "data-max-width",
      LAYOUT.pageContentMaxWidth.tableNarrow,
    );
    expect(screen.getByTestId("diagnosis-type-tab-content")).toBeInTheDocument();
  });

  it("病名タブ（?tab=diagnosis_name）表示時にも tableNarrow (max-w-5xl) が PageLayout に渡される", () => {
    renderDiagnosisSettings("/settings/diagnosis?tab=diagnosis_name");

    expect(pageLayoutPropsSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        title: "診断マスタ",
        resource: "master-medical",
        maxWidth: LAYOUT.pageContentMaxWidth.tableNarrow,
      }),
    );
    expect(screen.getByTestId("page-layout-mock")).toHaveAttribute(
      "data-max-width",
      LAYOUT.pageContentMaxWidth.tableNarrow,
    );
    expect(screen.getByTestId("diagnosis-name-tab-content")).toBeInTheDocument();
  });
});
