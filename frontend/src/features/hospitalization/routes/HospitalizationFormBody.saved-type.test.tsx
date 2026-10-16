import type { ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { HospitalizationFormBody } from "./HospitalizationFormBody";
import type { HospitalizationFormData } from "../types";

// EMR-412 R4: 編集・削除確認の見出しは「保存済みの区分」からのみ作る。
// 未保存の formData.hospitalizationType を変更してもレコードの同一性（見出し）が変わらないこと。
vi.mock("@/components/shared/PageLayout/PageLayout", () => ({
  PageLayout: ({ title, children }: { title: string; children?: ReactNode }) => (
    <div>
      <h1>{title}</h1>
      {children}
    </div>
  ),
}));
vi.mock("@/components/shared/Form/FormHeaderActions", () => ({
  FormHeaderActions: () => null,
}));
vi.mock("@/components/shared/MasterSelectModal", () => ({
  MasterSelectModal: () => null,
}));
vi.mock("@/components/shared/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));
vi.mock("./HospitalizationFormHeaderExtra", () => ({
  HospitalizationFormHeaderExtra: () => null,
}));
vi.mock("./HospitalizationFormFields", () => ({
  HospitalizationFormFields: () => null,
}));

const formData: HospitalizationFormData = {
  hospitalizationType: "入院",
} as HospitalizationFormData;

function renderBody({
  savedHospitalizationType,
  editTypeInForm,
}: {
  savedHospitalizationType: "入院" | "ホテル";
  editTypeInForm: string;
}) {
  return render(
    <HospitalizationFormBody
      hospitalizationId="hosp-1"
      canSubmit
      canShowDelete={false}
      isDirty={false}
      isDeleteConfirmOpen
      isDeletePending={false}
      staffModalOpen={false}
      doctorStaffItems={[]}
      formData={{ ...formData, hospitalizationType: editTypeInForm }}
      savedHospitalizationType={savedHospitalizationType}
      formAction={vi.fn()}
      onBack={vi.fn()}
      onOpenDetail={vi.fn()}
      onOpenDeleteConfirm={vi.fn()}
      onCloseDeleteConfirm={vi.fn()}
      onConfirmDelete={vi.fn()}
      onStaffModalOpenChange={vi.fn()}
      onSelectDoctor={vi.fn()}
      fields={{} as never}
    />,
  );
}

describe("HospitalizationFormBody 保存済み区分の見出し (EMR-412 R4)", () => {
  it("ホテル保存済みレコードは formData を入院に変えてもホテル編集のまま", () => {
    renderBody({ savedHospitalizationType: "ホテル", editTypeInForm: "入院" });

    expect(screen.getByRole("heading", { name: "ホテル編集", hidden: true })).toBeInTheDocument();
    expect(screen.getByText("ホテルを削除しますか？")).toBeInTheDocument();
  });

  it("入院保存済みレコードは入院編集・入院を削除しますか？", () => {
    renderBody({ savedHospitalizationType: "入院", editTypeInForm: "ホテル" });

    expect(screen.getByRole("heading", { name: "入院編集", hidden: true })).toBeInTheDocument();
    expect(screen.getByText("入院を削除しますか？")).toBeInTheDocument();
  });
});
