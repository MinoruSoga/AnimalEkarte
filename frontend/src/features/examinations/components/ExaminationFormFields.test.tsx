import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";
import { ExaminationFormFields } from "./ExaminationFormFields";

describe("ExaminationFormFields", () => {
  function renderFields({
    isEdit = false,
    canCreate = true,
    canEdit = true,
    fieldErrors,
    onSetFormData = vi.fn(),
  }: {
    isEdit?: boolean;
    canCreate?: boolean;
    canEdit?: boolean;
    fieldErrors?: Record<string, string>;
    onSetFormData?: (next: Record<string, unknown>) => void;
  } = {}) {
    return render(
      <MemoryRouter>
        <ExaminationFormFields
          formData={{ date: "2026-07-21T00:00:00+09:00", status: "依頼中" }}
          examTypes={[{ id: "5", name: "血液検査（院内）" }]}
          staffList={[{ id: "3", name: "林文明" }]}
          masterLoading={false}
          isEdit={isEdit}
          isDeleting={false}
          isConfirmed={false}
          canEdit={canEdit}
          canCreate={canCreate}
          canDelete
          fieldErrors={fieldErrors}
          onSetFormData={onSetFormData}
          onBack={vi.fn()}
          onDeleteClick={vi.fn()}
        />
      </MemoryRouter>,
    );
  }

  it("BUG-017: 検査種別・担当医の fieldErrors を近傍表示し aria-invalid/describedby を付与する", () => {
    renderFields({
      fieldErrors: {
        testTypeId: "検査種別を選択してください",
        doctorId: "担当医を選択してください",
      },
    });

    const testType = screen.getByRole("combobox", { name: "検査種別" });
    const doctor = screen.getByRole("combobox", { name: "担当医" });
    expect(
      screen.getByText("検査種別を選択してください"),
    ).toBeInTheDocument();
    expect(screen.getByText("担当医を選択してください")).toBeInTheDocument();
    expect(screen.getAllByRole("alert")).toHaveLength(2);

    expect(testType).toHaveAttribute("aria-invalid", "true");
    expect(doctor).toHaveAttribute("aria-invalid", "true");
    expect(testType).toHaveAttribute("aria-describedby", "testTypeId-error");
    expect(doctor).toHaveAttribute("aria-describedby", "doctorId-error");
    expect(document.getElementById("testTypeId-error")).toHaveTextContent(
      "検査種別を選択してください",
    );
    expect(document.getElementById("doctorId-error")).toHaveTextContent(
      "担当医を選択してください",
    );
  });

  it("BUG-017: 値変更時に onSetFormData が呼ばれ sibling error 表示は親の fieldErrors に従う", async () => {
    const user = userEvent.setup();
    const onSetFormData = vi.fn();
    renderFields({
      fieldErrors: {
        testTypeId: "検査種別を選択してください",
        doctorId: "担当医を選択してください",
      },
      onSetFormData,
    });

    await user.click(screen.getByRole("combobox", { name: "検査種別" }));
    await user.click(screen.getByRole("option", { name: "血液検査（院内）" }));

    expect(onSetFormData).toHaveBeenCalledWith(
      expect.objectContaining({ testTypeId: "5" }),
    );
    // Component itself does not clear sibling; parent supplies remaining error
    expect(screen.getByText("担当医を選択してください")).toBeInTheDocument();
  });

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

  it("BUG-033: 完了シールでは結果ロック文言・保存/削除非表示、ステータスは変更可", () => {
    render(
      <MemoryRouter>
        <ExaminationFormFields
          formData={{ date: "2026-07-21T00:00:00+09:00", status: "完了" }}
          examTypes={[]}
          staffList={[]}
          masterLoading={false}
          isEdit
          isDeleting={false}
          isConfirmed={false}
          isCompletedLocked
          canEdit
          canCreate
          canDelete
          onSetFormData={vi.fn()}
          onBack={vi.fn()}
          onDeleteClick={vi.fn()}
        />
      </MemoryRouter>,
    );

    expect(
      screen.getByText(/完了済みのため結果の編集・削除はできません/),
    ).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "保存" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
    expect(screen.getByRole("combobox", { name: "ステータス" })).not.toBeDisabled();
  });

  it("BUG-033: 完了シールでステータスを確定に変えると保存buttonが再表示される", () => {
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
          isCompletedLocked
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
    expect(screen.queryByRole("button", { name: "削除" })).not.toBeInTheDocument();
  });
});
