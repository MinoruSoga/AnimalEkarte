import { act, renderHook, waitFor } from "@testing-library/react";
import { startTransition } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useGetPet } from "@/hooks/use-pet";
import { useGetCheckupTypeFields } from "../api/get-checkup-type-fields";
import { useCheckupForm } from "./use-checkup-form";

const {
  createMedicalRecordMock,
  createCheckupMock,
  handleApiErrorMock,
  navigateMock,
  replaceCheckupFieldResultsMock,
  toastSuccessMock,
} = vi.hoisted(() => ({
  createMedicalRecordMock: vi.fn(),
  createCheckupMock: vi.fn(),
  handleApiErrorMock: vi.fn(),
  navigateMock: vi.fn(),
  replaceCheckupFieldResultsMock: vi.fn(),
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
  useGetPet: vi.fn(),
}));

vi.mock("../api/create-checkup-medical-record", () => ({
  createMedicalRecordForCheckup: createMedicalRecordMock,
  createCheckupOnMedicalRecord: createCheckupMock,
}));

vi.mock("../api/get-checkup-type-fields", () => ({
  useGetCheckupTypeFields: vi.fn(),
}));

vi.mock("../api/replace-checkup-field-results", () => ({
  replaceCheckupFieldResults: replaceCheckupFieldResultsMock,
}));

const ALLOWED_MUTATION_PERMISSIONS = {
  canCreate: true,
  canEdit: true,
} as const;

function runFormAction(action: (payload: FormData) => void) {
  act(() => {
    startTransition(() => {
      void action(new FormData());
    });
  });
}

function prepareValidForm(result: { current: ReturnType<typeof useCheckupForm> }) {
  act(() => {
    result.current.setCheckupTypeId("1");
    result.current.setDate("2026-07-22");
  });
}

describe("useCheckupForm mutation boundary", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGetPet).mockReturnValue({
      data: { id: "pet-1", ownerId: "owner-1" },
      isLoading: false,
    } as ReturnType<typeof useGetPet>);
    vi.mocked(useGetCheckupTypeFields).mockReturnValue({
      data: [],
    } as ReturnType<typeof useGetCheckupTypeFields>);
    createMedicalRecordMock.mockResolvedValue({ id: "record-1" });
    createCheckupMock.mockResolvedValue({ id: "checkup-1" });
    replaceCheckupFieldResultsMock.mockResolvedValue(undefined);
  });

  it("全ての登録が成功したactionだけが一覧へ遷移する", async () => {
    const { result } = renderHook(() => useCheckupForm(ALLOWED_MUTATION_PERMISSIONS));

    prepareValidForm(result);
    runFormAction(result.current.formAction);

    await waitFor(() => expect(navigateMock).toHaveBeenCalledWith("/checkups"));
    expect(createMedicalRecordMock).toHaveBeenCalledOnce();
    expect(createCheckupMock).toHaveBeenCalledOnce();
    expect(toastSuccessMock).toHaveBeenCalledWith("定期健診を登録しました");
  });

  it("登録失敗時は一覧へ遷移しない", async () => {
    createMedicalRecordMock.mockRejectedValueOnce(new Error("create failed"));
    const { result } = renderHook(() => useCheckupForm(ALLOWED_MUTATION_PERMISSIONS));

    prepareValidForm(result);
    runFormAction(result.current.formAction);

    await waitFor(() => expect(handleApiErrorMock).toHaveBeenCalledWith(expect.any(Error), "保存"));
    expect(navigateMock).not.toHaveBeenCalledWith("/checkups");
  });

  it.each([
    { canCreate: false, canEdit: true },
    { canCreate: true, canEdit: false },
    { canCreate: false, canEdit: false },
  ])(
    "create/editのどちらかがstrict trueでなければ最初のmutationを実行しない ($canCreate/$canEdit)",
    async (permissions) => {
      const { result } = renderHook(() => useCheckupForm(permissions));

      prepareValidForm(result);
      runFormAction(result.current.formAction);

      await waitFor(() => expect(result.current.formState.timestamp).not.toBe(0));
      expect(createMedicalRecordMock).not.toHaveBeenCalled();
      expect(createCheckupMock).not.toHaveBeenCalled();
      expect(replaceCheckupFieldResultsMock).not.toHaveBeenCalled();
      expect(toastSuccessMock).not.toHaveBeenCalled();
    },
  );

  it("明示的に死亡しているpetId直指定では権限があってもmutationを実行しない", async () => {
    vi.mocked(useGetPet).mockReturnValue({
      data: {
        id: "pet-1",
        ownerId: "owner-1",
        status: "死亡",
      },
      isLoading: false,
    } as ReturnType<typeof useGetPet>);
    const { result } = renderHook(() => useCheckupForm(ALLOWED_MUTATION_PERMISSIONS));

    prepareValidForm(result);
    runFormAction(result.current.formAction);

    await waitFor(() => expect(result.current.formState.timestamp).not.toBe(0));
    expect(createMedicalRecordMock).not.toHaveBeenCalled();
    expect(createCheckupMock).not.toHaveBeenCalled();
    expect(replaceCheckupFieldResultsMock).not.toHaveBeenCalled();
    expect(toastSuccessMock).not.toHaveBeenCalled();
  });

  it("カルテ作成待機中にcreate権限を剥奪すると後続の健診createを実行しない", async () => {
    let resolveMedicalRecord!: (value: { id: string }) => void;
    createMedicalRecordMock.mockReturnValue(
      new Promise((resolve) => {
        resolveMedicalRecord = resolve;
      }),
    );
    const { result, rerender } = renderHook(
      ({ canCreate }: { canCreate: boolean }) => useCheckupForm({ canCreate, canEdit: true }),
      { initialProps: { canCreate: true } },
    );

    prepareValidForm(result);
    runFormAction(result.current.formAction);
    await waitFor(() => expect(createMedicalRecordMock).toHaveBeenCalledOnce());

    rerender({ canCreate: false });
    await act(async () => {
      resolveMedicalRecord({ id: "record-1" });
      await Promise.resolve();
    });

    expect(createCheckupMock).not.toHaveBeenCalled();
    expect(replaceCheckupFieldResultsMock).not.toHaveBeenCalled();
    expect(toastSuccessMock).not.toHaveBeenCalled();
  });

  it("カルテ作成待機中にedit権限を剥奪すると後続の健診createを実行しない", async () => {
    let resolveMedicalRecord!: (value: { id: string }) => void;
    createMedicalRecordMock.mockReturnValue(
      new Promise((resolve) => {
        resolveMedicalRecord = resolve;
      }),
    );
    const { result, rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) => useCheckupForm({ canCreate: true, canEdit }),
      { initialProps: { canEdit: true } },
    );

    prepareValidForm(result);
    runFormAction(result.current.formAction);
    await waitFor(() => expect(createMedicalRecordMock).toHaveBeenCalledOnce());

    rerender({ canEdit: false });
    await act(async () => {
      resolveMedicalRecord({ id: "record-1" });
      await Promise.resolve();
    });

    expect(createCheckupMock).not.toHaveBeenCalled();
    expect(replaceCheckupFieldResultsMock).not.toHaveBeenCalled();
    expect(toastSuccessMock).not.toHaveBeenCalled();
  });

  it("カルテ作成待機中にpetが死亡へ変わると後続の健診createを実行しない", async () => {
    let resolveMedicalRecord!: (value: { id: string }) => void;
    createMedicalRecordMock.mockReturnValue(
      new Promise((resolve) => {
        resolveMedicalRecord = resolve;
      }),
    );
    const { result, rerender } = renderHook(() => useCheckupForm(ALLOWED_MUTATION_PERMISSIONS));

    prepareValidForm(result);
    runFormAction(result.current.formAction);
    await waitFor(() => expect(createMedicalRecordMock).toHaveBeenCalledOnce());

    vi.mocked(useGetPet).mockReturnValue({
      data: { id: "pet-1", ownerId: "owner-1", status: "死亡" },
      isLoading: false,
    } as ReturnType<typeof useGetPet>);
    rerender();
    await act(async () => {
      resolveMedicalRecord({ id: "record-1" });
      await Promise.resolve();
    });

    expect(createCheckupMock).not.toHaveBeenCalled();
    expect(replaceCheckupFieldResultsMock).not.toHaveBeenCalled();
    expect(toastSuccessMock).not.toHaveBeenCalled();
  });

  it("健診作成待機中にedit権限を剥奪するとfield resultsを置換しない", async () => {
    let resolveCheckup!: (value: { id: string }) => void;
    createCheckupMock.mockReturnValue(
      new Promise((resolve) => {
        resolveCheckup = resolve;
      }),
    );
    vi.mocked(useGetCheckupTypeFields).mockReturnValue({
      data: [
        {
          id: 1,
          checkupTypeId: 1,
          name: "所見",
          fieldType: "text",
          unit: "",
          options: [],
          isProvisional: false,
          sortOrder: 1,
        },
      ],
    } as ReturnType<typeof useGetCheckupTypeFields>);
    const { result, rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) => useCheckupForm({ canCreate: true, canEdit }),
      { initialProps: { canEdit: true } },
    );

    prepareValidForm(result);
    act(() => {
      result.current.setFieldValue(1, { text: "異常なし" });
    });
    runFormAction(result.current.formAction);
    await waitFor(() => expect(createCheckupMock).toHaveBeenCalledOnce());

    rerender({ canEdit: false });
    await act(async () => {
      resolveCheckup({ id: "checkup-1" });
      await Promise.resolve();
    });

    expect(replaceCheckupFieldResultsMock).not.toHaveBeenCalled();
    expect(toastSuccessMock).not.toHaveBeenCalled();
  });

  it("健診作成待機中にcreate権限を剥奪するとfield resultsを置換しない", async () => {
    let resolveCheckup!: (value: { id: string }) => void;
    createCheckupMock.mockReturnValue(
      new Promise((resolve) => {
        resolveCheckup = resolve;
      }),
    );
    vi.mocked(useGetCheckupTypeFields).mockReturnValue({
      data: [
        {
          id: 1,
          checkupTypeId: 1,
          name: "所見",
          fieldType: "text",
          unit: "",
          options: [],
          isProvisional: false,
          sortOrder: 1,
        },
      ],
    } as ReturnType<typeof useGetCheckupTypeFields>);
    const { result, rerender } = renderHook(
      ({ canCreate }: { canCreate: boolean }) => useCheckupForm({ canCreate, canEdit: true }),
      { initialProps: { canCreate: true } },
    );

    prepareValidForm(result);
    act(() => {
      result.current.setFieldValue(1, { text: "異常なし" });
    });
    runFormAction(result.current.formAction);
    await waitFor(() => expect(createCheckupMock).toHaveBeenCalledOnce());

    rerender({ canCreate: false });
    await act(async () => {
      resolveCheckup({ id: "checkup-1" });
      await Promise.resolve();
    });

    expect(replaceCheckupFieldResultsMock).not.toHaveBeenCalled();
    expect(toastSuccessMock).not.toHaveBeenCalled();
  });

  it("健診作成待機中にpetが死亡へ変わるとfield resultsを置換しない", async () => {
    let resolveCheckup!: (value: { id: string }) => void;
    createCheckupMock.mockReturnValue(
      new Promise((resolve) => {
        resolveCheckup = resolve;
      }),
    );
    vi.mocked(useGetCheckupTypeFields).mockReturnValue({
      data: [
        {
          id: 1,
          checkupTypeId: 1,
          name: "所見",
          fieldType: "text",
          unit: "",
          options: [],
          isProvisional: false,
          sortOrder: 1,
        },
      ],
    } as ReturnType<typeof useGetCheckupTypeFields>);
    const { result, rerender } = renderHook(() => useCheckupForm(ALLOWED_MUTATION_PERMISSIONS));

    prepareValidForm(result);
    act(() => {
      result.current.setFieldValue(1, { text: "異常なし" });
    });
    runFormAction(result.current.formAction);
    await waitFor(() => expect(createCheckupMock).toHaveBeenCalledOnce());

    vi.mocked(useGetPet).mockReturnValue({
      data: { id: "pet-1", ownerId: "owner-1", status: "死亡" },
      isLoading: false,
    } as ReturnType<typeof useGetPet>);
    rerender();
    await act(async () => {
      resolveCheckup({ id: "checkup-1" });
      await Promise.resolve();
    });

    expect(replaceCheckupFieldResultsMock).not.toHaveBeenCalled();
    expect(toastSuccessMock).not.toHaveBeenCalled();
  });
});
