import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi } from "vitest";
import { ExaminationFormFields } from "./ExaminationFormFields";

type FieldsProps = Parameters<typeof ExaminationFormFields>[0];

describe("ExaminationFormFields", () => {
  function renderFields({
    formData = { date: "2026-07-21T00:00:00+09:00", status: "依頼中" },
    isConfirmed = false,
    isCompletedLocked = false,
    fieldErrors,
    onSetFormData = vi.fn(),
  }: Partial<
    Pick<FieldsProps, "formData" | "isConfirmed" | "isCompletedLocked" | "fieldErrors">
  > & {
    onSetFormData?: Parameters<typeof ExaminationFormFields>[0]["onSetFormData"];
  } = {}) {
    return render(
      <MemoryRouter>
        <ExaminationFormFields
          formData={formData}
          examTypes={[{ id: "5", name: "血液検査（院内）" }]}
          staffList={[{ id: "3", name: "林文明" }]}
          masterLoading={false}
          isConfirmed={isConfirmed}
          isCompletedLocked={isCompletedLocked}
          fieldErrors={fieldErrors}
          onSetFormData={onSetFormData}
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
    expect(screen.getByText("検査種別を選択してください")).toBeInTheDocument();
    expect(screen.getByText("担当医を選択してください")).toBeInTheDocument();
    expect(screen.getAllByRole("alert")).toHaveLength(2);

    expect(testType).toHaveAttribute("aria-invalid", "true");
    expect(doctor).toHaveAttribute("aria-invalid", "true");
    expect(testType).toHaveAttribute("aria-describedby", "testTypeId-error");
    expect(doctor).toHaveAttribute("aria-describedby", "doctorId-error");
    expect(document.getElementById("testTypeId-error")).toHaveTextContent(
      "検査種別を選択してください",
    );
    expect(document.getElementById("doctorId-error")).toHaveTextContent("担当医を選択してください");
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

    expect(onSetFormData).toHaveBeenCalledWith(expect.objectContaining({ testTypeId: "5" }));
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
    expect(screen.getByRole("combobox", { name: "ステータス" })).toHaveClass("h-11", "min-w-11");
    expect(screen.getByRole("textbox", { name: "備考・所見" })).toBeInTheDocument();
  });

  it("未確定でローカルステータスが確定でも入力とステータスを有効に保つ（A-S02-01）", () => {
    renderFields({ formData: { date: "2026-07-21T00:00:00+09:00", status: "確定" } });

    for (const name of ["検査種別", "担当医", "ステータス"]) {
      expect(screen.getByRole("combobox", { name })).toBeEnabled();
    }
    expect(screen.getByRole("textbox", { name: "検査日" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "カレンダーを開く" })).toBeEnabled();
    expect(screen.getByRole("textbox", { name: "備考・所見" })).toBeEnabled();
  });

  it("サーバ確定済みでは日付・カレンダー・結果・ステータスをロックする", () => {
    renderFields({
      isConfirmed: true,
      formData: { date: "2026-07-21T00:00:00+09:00", status: "確定" },
    });

    expect(screen.getByText("確定済みのため編集できません。")).toBeInTheDocument();
    for (const name of ["検査種別", "担当医", "ステータス"]) {
      expect(screen.getByRole("combobox", { name })).toBeDisabled();
    }
    expect(screen.getByRole("textbox", { name: "検査日" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "カレンダーを開く" })).toBeDisabled();
    expect(screen.getByRole("textbox", { name: "備考・所見" })).toBeDisabled();
  });

  it.each(["完了", "確定"] as const)(
    "BUG-033: 完了シールはローカルステータス%sでも結果をロックしステータスだけ変更可",
    (status) => {
      renderFields({
        isCompletedLocked: true,
        formData: { date: "2026-07-21T00:00:00+09:00", status },
      });

      expect(screen.getByText(/完了済みのため結果の編集・削除はできません/)).toBeInTheDocument();
      for (const name of ["検査種別", "担当医"]) {
        expect(screen.getByRole("combobox", { name })).toBeDisabled();
      }
      expect(screen.getByRole("textbox", { name: "検査日" })).toBeDisabled();
      expect(screen.getByRole("button", { name: "カレンダーを開く" })).toBeDisabled();
      expect(screen.getByRole("textbox", { name: "備考・所見" })).toBeDisabled();
      expect(screen.getByRole("combobox", { name: "ステータス" })).toBeEnabled();
    },
  );

  it("BUG-033: 完了シールでステータスを確定に変えたフォームはネイティブ送信できる（日付が妨げない）", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn((event: React.FormEvent<HTMLFormElement>) => event.preventDefault());
    const onSetFormData = vi.fn();
    render(
      <MemoryRouter>
        <form onSubmit={onSubmit}>
          <ExaminationFormFields
            formData={{ date: "2026-10-13T00:00:00+09:00", status: "確定" }}
            examTypes={[]}
            staffList={[]}
            masterLoading={false}
            isConfirmed={false}
            isCompletedLocked
            onSetFormData={onSetFormData}
          />
          <button type="submit">送信</button>
        </form>
      </MemoryRouter>,
    );

    const dateInput = screen.getByLabelText("検査日");
    expect(dateInput).toBeDisabled();
    expect(screen.getByRole("button", { name: "カレンダーを開く" })).toBeDisabled();

    await user.click(screen.getByRole("button", { name: "送信" }));
    expect(onSubmit).toHaveBeenCalledTimes(1);
    expect(onSetFormData).not.toHaveBeenCalled();
  });

  it("BUG-033: ロックなしの過去検査日ではネイティブ送信が通る基準値", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn((event: React.FormEvent<HTMLFormElement>) => event.preventDefault());
    render(
      <MemoryRouter>
        <form onSubmit={onSubmit}>
          <ExaminationFormFields
            formData={{ date: "2026-10-13T00:00:00+09:00", status: "依頼中" }}
            examTypes={[]}
            staffList={[]}
            masterLoading={false}
            isConfirmed={false}
            onSetFormData={vi.fn()}
          />
          <button type="submit">送信</button>
        </form>
      </MemoryRouter>,
    );

    await user.click(screen.getByRole("button", { name: "送信" }));
    expect(onSubmit).toHaveBeenCalledTimes(1);
  });
});

describe("EMR-415 E4: 必須印", () => {
  function renderFieldsPanel() {
    return render(
      <MemoryRouter>
        <ExaminationFormFields
          formData={{ date: "2026-07-21T00:00:00+09:00", status: "依頼中" }}
          examTypes={[{ id: "5", name: "血液検査（院内）" }]}
          staffList={[{ id: "3", name: "林文明" }]}
          masterLoading={false}
          isConfirmed={false}
          isCompletedLocked={false}
          onSetFormData={vi.fn()}
        />
      </MemoryRouter>,
    );
  }

  it("検査種別と担当医のラベルに必須印（*）を表示する", () => {
    renderFieldsPanel();

    const typeLabel = screen.getByText("検査種別").closest("label");
    expect(typeLabel).not.toBeNull();
    expect(typeLabel?.textContent).toContain("*");
    const doctorLabel = screen.getByText("担当医").closest("label");
    expect(doctorLabel).not.toBeNull();
    expect(doctorLabel?.textContent).toContain("*");
  });
});
