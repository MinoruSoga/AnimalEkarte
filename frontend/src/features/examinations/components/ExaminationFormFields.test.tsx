import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";
import { ExaminationFormFields } from "./ExaminationFormFields";

describe("ExaminationFormFields", () => {
  function renderFields({ isEdit = false }: { isEdit?: boolean } = {}) {
    return render(
      <MemoryRouter>
        <ExaminationFormFields
          formData={{ date: "2026-07-21T00:00:00+09:00", status: "依頼中" }}
          examTypes={[]}
          staffList={[]}
          masterLoading={false}
          isEdit={isEdit}
          isDeleting={false}
          isConfirmed={false}
          canEdit
          canCreate
          canDelete
          onSetFormData={vi.fn()}
          onBack={vi.fn()}
          onDeleteClick={vi.fn()}
        />
      </MemoryRouter>,
    );
  }

  it("検査日の実inputをラベル付けし44px以上の操作領域にする", () => {
    renderFields();

    expect(screen.getByRole("textbox", { name: "検査日" })).toHaveClass("min-h-11", "-my-px");
  });

  it("検査種別・担当医・ステータス・備考のvisible labelをcontrolへ接続する", () => {
    renderFields();

    expect(screen.getByRole("combobox", { name: "検査種別" })).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "担当医" })).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "ステータス" })).toHaveClass(
      "h-11",
      "min-w-11",
    );
    expect(screen.getByRole("textbox", { name: "備考・所見" })).toBeInTheDocument();
  });

  it("削除・キャンセル・保存buttonでh-10 overrideを再導入しない", () => {
    renderFields({ isEdit: true });

    for (const name of ["削除", "キャンセル", "保存"]) {
      const button = screen.getByRole("button", { name });
      expect(button).toHaveClass("h-11", "min-h-11");
      expect(button).not.toHaveClass("h-10");
    }
  });
});
