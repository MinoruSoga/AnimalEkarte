import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";
import { ExaminationFormFields } from "./ExaminationFormFields";

describe("ExaminationFormFields", () => {
  function renderFields({
    isEdit = false,
    canCreate = true,
    canEdit = true,
  }: {
    isEdit?: boolean;
    canCreate?: boolean;
    canEdit?: boolean;
  } = {}) {
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
          canEdit={canEdit}
          canCreate={canCreate}
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

  it("新規作成は作成権限があっても編集権限なしなら保存buttonを表示しない", () => {
    renderFields({ canCreate: true, canEdit: false });

    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
  });

  it("編集は作成権限なしでも編集権限があれば保存buttonを表示する", () => {
    renderFields({ isEdit: true, canCreate: false, canEdit: true });

    expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
  });

  it("未確定（isConfirmed=false）でステータスが確定でも保存buttonを表示する（A-S02-01）", () => {
    render(
      <MemoryRouter>
        <ExaminationFormFields
          formData={{ date: "2026-07-21T00:00:00+09:00", status: "確定" }}
          examTypes={[]}
          staffList={[]}
          masterLoading={false}
          isEdit
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

    expect(screen.getByRole("button", { name: "保存" })).toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "ステータス" })).not.toBeDisabled();
  });

  it("サーバ確定済み（isConfirmed=true）では保存buttonを消しステータスを無効化する", () => {
    render(
      <MemoryRouter>
        <ExaminationFormFields
          formData={{ date: "2026-07-21T00:00:00+09:00", status: "確定" }}
          examTypes={[]}
          staffList={[]}
          masterLoading={false}
          isEdit
          isDeleting={false}
          isConfirmed
          canEdit
          canCreate
          canDelete
          onSetFormData={vi.fn()}
          onBack={vi.fn()}
          onDeleteClick={vi.fn()}
        />
      </MemoryRouter>,
    );

    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "ステータス" })).toBeDisabled();
  });
});
