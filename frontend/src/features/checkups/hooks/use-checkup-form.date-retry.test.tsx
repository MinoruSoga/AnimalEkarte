import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";

import { DatePicker } from "@/components/shared/DatePicker";
import { useCheckupForm } from "./use-checkup-form";
import { createMedicalRecordForCheckup } from "../api/create-checkup-medical-record";

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("@/hooks/use-pet", () => ({
  useGetPet: () => ({
    data: { id: "pet-1", ownerId: "owner-1", status: "生存" },
    isLoading: false,
  }),
}));
vi.mock("@/hooks/use-back-navigation", () => ({ useBackPath: () => "/checkups" }));
vi.mock("../api/get-checkup-type-fields", () => ({
  useGetCheckupTypeFields: () => ({ data: [] }),
}));
vi.mock("../api/create-checkup-medical-record", () => ({
  createMedicalRecordForCheckup: vi.fn().mockResolvedValue({ id: "record-1" }),
  createCheckupOnMedicalRecord: vi.fn().mockResolvedValue({ id: "checkup-1" }),
}));
vi.mock("../api/replace-checkup-field-results", () => ({
  replaceCheckupFieldResults: vi.fn(),
}));

function CheckupHarness() {
  const h = useCheckupForm({ canCreate: true, canEdit: true });
  return (
    <form action={h.formAction}>
      <DatePicker
        ariaLabel="実施日"
        value={h.form.date}
        onChange={h.setDate}
        ariaInvalid={Boolean(h.fieldErrors.date)}
      />
      <button type="button" onClick={() => h.setCheckupTypeId("1")}>
        種別設定
      </button>
      <button type="submit">保存</button>
    </form>
  );
}

function renderHarness() {
  render(
    <MemoryRouter initialEntries={["/checkups/new?petId=pet-1"]}>
      <CheckupHarness />
    </MemoryRouter>,
  );
  return {
    input: () => screen.getByRole<HTMLInputElement>("textbox", { name: "実施日" }),
    save: () => screen.getByRole("button", { name: "保存" }),
    typeBtn: () => screen.getByRole("button", { name: "種別設定" }),
  };
}

describe("useCheckupForm 実施日エラー後の再保存", () => {
  beforeEach(() => {
    vi.useFakeTimers({ toFake: ["Date"] });
    vi.setSystemTime(new Date("2026-10-14T03:00:00Z"));
    vi.mocked(createMedicalRecordForCheckup).mockClear();
  });
  afterEach(() => {
    cleanup();
    vi.useRealTimers();
  });

  it("最初から有効日なら保存で createMedicalRecordForCheckup を1回呼ぶ", async () => {
    const user = userEvent.setup();
    const { input, save, typeBtn } = renderHarness();
    await user.click(typeBtn());
    await user.click(input());
    await user.type(input(), "2026/10/13");
    await user.tab();
    await user.click(save());
    await waitFor(() => expect(vi.mocked(createMedicalRecordForCheckup)).toHaveBeenCalledTimes(1));
  });

  it("空日付のバリデーション失敗後に日付を直すと再 submit が API に届く", async () => {
    const user = userEvent.setup();
    const { input, save, typeBtn } = renderHarness();
    await user.click(typeBtn());
    await user.click(save());
    await waitFor(() => expect(input().getAttribute("aria-invalid")).toBe("true"));

    await user.click(input());
    await user.type(input(), "2026/10/13");
    await user.tab();
    expect(input().getAttribute("aria-invalid")).toBe("true");

    await user.click(save());
    await waitFor(() => {
      expect(vi.mocked(createMedicalRecordForCheckup)).toHaveBeenCalledTimes(1);
    });
  });
});
