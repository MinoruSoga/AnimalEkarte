import { act, renderHook, waitFor } from "@testing-library/react";
import { startTransition } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useCheckupForm } from "./use-checkup-form";

const {
  createMedicalRecordMock,
  createCheckupMock,
  handleApiErrorMock,
  navigateMock,
  toastSuccessMock,
} = vi.hoisted(() => ({
  createMedicalRecordMock: vi.fn(),
  createCheckupMock: vi.fn(),
  handleApiErrorMock: vi.fn(),
  navigateMock: vi.fn(),
  toastSuccessMock: vi.fn(),
}));

vi.mock("react-router", () => ({
  useNavigate: () => navigateMock,
  useSearchParams: () => [new URLSearchParams("petId=pet-1")],
}));

vi.mock("sonner", () => ({
  toast: { success: toastSuccessMock },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: handleApiErrorMock,
}));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: () => ({
    data: { id: "pet-1", ownerId: "owner-1" },
    isLoading: false,
  }),
}));

vi.mock("../api/create-checkup-medical-record", () => ({
  createMedicalRecordForCheckup: createMedicalRecordMock,
  createCheckupOnMedicalRecord: createCheckupMock,
}));

vi.mock("../api/get-checkup-type-fields", () => ({
  useGetCheckupTypeFields: () => ({ data: [] }),
}));

vi.mock("../api/replace-checkup-field-results", () => ({
  replaceCheckupFieldResults: vi.fn(),
}));

describe("useCheckupForm success navigation", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    createMedicalRecordMock.mockResolvedValue({ id: "record-1" });
    createCheckupMock.mockResolvedValue({ id: "checkup-1" });
  });

  it("全ての登録が成功したactionだけが一覧へ遷移する", async () => {
    const { result } = renderHook(() => useCheckupForm());

    act(() => {
      result.current.setCheckupTypeId("1");
      result.current.setDate("2026-07-22");
    });

    act(() => {
      startTransition(() => {
        void result.current.formAction(new FormData());
      });
    });

    await waitFor(() => expect(navigateMock).toHaveBeenCalledWith("/checkups"));
    expect(createMedicalRecordMock).toHaveBeenCalledOnce();
    expect(createCheckupMock).toHaveBeenCalledOnce();
    expect(toastSuccessMock).toHaveBeenCalledWith("定期健診を登録しました");
  });

  it("登録失敗時は一覧へ遷移しない", async () => {
    createMedicalRecordMock.mockRejectedValueOnce(new Error("create failed"));
    const { result } = renderHook(() => useCheckupForm());

    act(() => {
      result.current.setCheckupTypeId("1");
      result.current.setDate("2026-07-22");
    });

    act(() => {
      startTransition(() => {
        void result.current.formAction(new FormData());
      });
    });

    await waitFor(() =>
      expect(handleApiErrorMock).toHaveBeenCalledWith(expect.any(Error), "保存"),
    );
    expect(navigateMock).not.toHaveBeenCalledWith("/checkups");
  });
});
