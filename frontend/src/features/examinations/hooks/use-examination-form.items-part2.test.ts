import { renderHook, act, waitFor } from "@testing-library/react";
import { startTransition } from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { useExaminationForm } from "./use-examination-form";
import { useSearchParams } from "react-router";
import { useGetPet } from "@/hooks/use-pet";
import { usePetSelection } from "@/hooks/use-pet-selection";
import { jstDateStartISOString, todayJSTISO } from "@/lib/jst-date";

// Mock dependencies
const mockNavigate = vi.fn();

vi.mock("react-router", () => ({
  useSearchParams: vi.fn(),
  useNavigate: vi.fn(() => mockNavigate),
}));

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn(), info: vi.fn() },
}));

vi.mock("@/hooks/use-pet-selection", () => ({
  usePetSelection: vi.fn(() => ({
    selectedPets: [],
    setSelectedPets: vi.fn(),
  })),
}));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: vi.fn(() => ({
    data: null,
    isLoading: false,
  })),
}));

vi.mock("../api/get-examination", () => ({
  useGetExamination: vi.fn(() => ({
    data: undefined,
    isLoading: false,
    isError: false,
    error: null,
    refetch: vi.fn(),
  })),
}));

vi.mock("../api/create-examination", () => ({
  useCreateExamination: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue({}),
  })),
}));

vi.mock("../api/update-examination", () => ({
  useUpdateExamination: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue({}),
  })),
}));

vi.mock("../api/delete-examination", () => ({
  useDeleteExamination: vi.fn(() => ({ mutate: vi.fn(), isPending: false })),
}));

vi.mock("../api/unconfirm-examination", () => ({
  useUnconfirmExamination: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue({}),
    isPending: false,
  })),
}));

vi.mock("../api/get-examination-items", () => ({
  useGetExaminationItems: vi.fn(() => ({
    data: undefined,
    isSuccess: false,
    isError: false,
  })),
}));

vi.mock("../api/update-examination-items", () => ({
  useUpdateExaminationItems: vi.fn(() => ({
    mutateAsync: vi.fn().mockResolvedValue([]),
    isPending: false,
  })),
}));

vi.mock("../api/get-exam-type-fields", () => ({
  useGetExamTypeFields: vi.fn(() => ({ data: undefined })),
}));

const ALLOWED_MUTATION_PERMISSIONS = {
  canCreate: true,
  canEdit: true,
  canDelete: true,
  canUnconfirm: true,
} as const;

function selectedPet(status: "生存" | "死亡" | "不明") {
  return {
    id: "42",
    name: "ポチ",
    ownerName: "田中",
    ownerId: "5",
    species: "犬",
    breed: "",
    birthday: "",
    gender: "男",
    weight: null,
    imageUrl: null,
    status,
    microchipNumber: null,
    insuranceNumber: null,
    insuranceExpiry: null,
    memo: null,
  };
}

function renderExaminationForm(id?: string) {
  return renderHook(() => useExaminationForm(id, undefined, ALLOWED_MUTATION_PERMISSIONS));
}

// FE-RC-045: use-examination-form.test.ts (2662行) をトピック別に分割した1ファイル。
// 各ファイルは vi.mock 済みモジュールを共有できないため、ヘッダ（import/mock/helper）を複製している。

describe("useExaminationForm — 検査項目テーブル（FE-EXAM-001）", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useSearchParams).mockReturnValue([new URLSearchParams(), vi.fn()]);
    vi.mocked(useGetPet).mockReturnValue({
      data: null,
      isLoading: false,
      isError: false,
    } as ReturnType<typeof useGetPet>);
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);
  });

  it("A取得済み→B取得中→A再表示でもA itemsを復元してAだけをPATCHする", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } = await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    const itemA = {
      id: "item-a",
      examTypeFieldId: 1,
      name: "A-WBC",
      result: "",
      inspectionValue: "1",
      normalValue: "",
      unit: "",
      referenceValue: "",
      refMin: undefined,
      refMax: undefined,
      isAbnormal: false,
      status: "normal" as const,
      sortOrder: 1,
    };
    vi.mocked(useGetExamination).mockImplementation(
      (examID) =>
        ({
          data: {
            id: examID,
            testTypeId: "5",
            doctorId: "3",
            status: "検査中" as const,
            ownerName: "",
            petName: "",
            date: "2026-08-01",
          },
        }) as ReturnType<typeof useGetExamination>,
    );
    vi.mocked(useGetExaminationItems).mockImplementation((examID) => {
      if (examID === "exam-a") {
        return {
          data: [itemA],
          isSuccess: true,
          isError: false,
          isPending: false,
        } as ReturnType<typeof useGetExaminationItems>;
      }
      return {
        data: undefined,
        isSuccess: false,
        isError: false,
        isPending: true,
      } as ReturnType<typeof useGetExaminationItems>;
    });
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result, rerender } = renderHook(
      ({ examID }: { examID: string }) =>
        useExaminationForm(examID, undefined, ALLOWED_MUTATION_PERMISSIONS),
      { initialProps: { examID: "exam-a" } },
    );
    await waitFor(() => expect(result.current.formItems[0]?.name).toBe("A-WBC"));

    rerender({ examID: "exam-b" });
    await waitFor(() => expect(result.current.formItems).toEqual([]));

    rerender({ examID: "exam-a" });
    await waitFor(() => expect(result.current.formItems[0]?.name).toBe("A-WBC"));

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });
    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-a",
        req: expect.objectContaining({
          items: [expect.objectContaining({ name: "A-WBC" })],
        }),
      }),
    );
  });

  it("同じ検査種別のA→B遷移後もBの種別変更で新テンプレへ再構築する", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } = await import("../api/get-examination-items");
    const { useGetExamTypeFields } = await import("../api/get-exam-type-fields");

    const templates = {
      "5": [
        {
          id: 500,
          name: "TEMPLATE-5",
          unit: "",
          normalValue: "",
          sortOrder: 1,
        },
      ],
      "6": [
        {
          id: 600,
          name: "TEMPLATE-6",
          unit: "",
          normalValue: "",
          sortOrder: 1,
        },
      ],
    };
    vi.mocked(useGetExamination).mockImplementation(
      (examID) =>
        ({
          data: {
            id: examID,
            testTypeId: "5",
            doctorId: "3",
            status: "検査中" as const,
            ownerName: "",
            petName: "",
            date: "2026-08-01",
          },
        }) as ReturnType<typeof useGetExamination>,
    );
    vi.mocked(useGetExaminationItems).mockImplementation(
      (examID) =>
        ({
          data: [
            {
              id: `item-${examID}`,
              examTypeFieldId: 1,
              name: `${examID}-WBC`,
              result: "",
              inspectionValue: "1",
              normalValue: "",
              unit: "",
              referenceValue: "",
              refMin: undefined,
              refMax: undefined,
              isAbnormal: false,
              status: "normal" as const,
              sortOrder: 1,
            },
          ],
          isSuccess: true,
          isError: false,
          isPending: false,
        }) as ReturnType<typeof useGetExaminationItems>,
    );
    vi.mocked(useGetExamTypeFields).mockImplementation(
      (examTypeID) =>
        ({ data: templates[examTypeID as keyof typeof templates] }) as ReturnType<
          typeof useGetExamTypeFields
        >,
    );

    const { result, rerender } = renderHook(
      ({ examID }: { examID: string }) =>
        useExaminationForm(examID, undefined, ALLOWED_MUTATION_PERMISSIONS),
      { initialProps: { examID: "exam-a" } },
    );
    await waitFor(() => expect(result.current.formItems[0]?.name).toBe("exam-a-WBC"));

    rerender({ examID: "exam-b" });
    await waitFor(() => expect(result.current.formItems[0]?.name).toBe("exam-b-WBC"));

    act(() => result.current.setFormData({ testTypeId: "6" }));
    await waitFor(() => expect(result.current.formItems[0]?.name).toBe("TEMPLATE-6"));
  });

  it("確定済み (status=確定) では PATCH を発行しない", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } = await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "確定" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [
        {
          id: "101",
          examTypeFieldId: 1,
          name: "WBC",
          result: "",
          inspectionValue: "5.0",
          normalValue: "",
          unit: "",
          referenceValue: "",
          refMin: undefined,
          refMax: undefined,
          isAbnormal: false,
          status: "normal" as const,
          sortOrder: 1,
        },
      ],
      isSuccess: true,
      isError: false,
    } as ReturnType<typeof useGetExaminationItems>);

    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(result.current.isPersistedConfirmed).toBe(true));
    expect(result.current.isPersistedResultsLocked).toBe(true);
    expect(updateMutate).not.toHaveBeenCalled();
  });

  it("BUG-033: 完了シールでは PATCH から items を省略し status 遷移を送る", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } = await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "完了" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [
        {
          id: "101",
          examTypeFieldId: 1,
          name: "WBC",
          result: "",
          inspectionValue: "5.0",
          normalValue: "",
          unit: "",
          referenceValue: "",
          refMin: undefined,
          refMax: undefined,
          isAbnormal: false,
          status: "normal" as const,
          sortOrder: 1,
        },
      ],
      isSuccess: true,
      isError: false,
    } as ReturnType<typeof useGetExaminationItems>);

    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");

    await waitFor(() => expect(result.current.isPersistedCompletedLocked).toBe(true));
    expect(result.current.isPersistedResultsLocked).toBe(true);

    await act(async () => {
      result.current.setFormData({ status: "確定" });
    });

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-001",
        req: expect.objectContaining({
          status: "confirmed",
        }),
      }),
    );
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        req: expect.not.objectContaining({ items: expect.anything() }),
      }),
    );
  });

  it("未確定検査でステータスを確定に変えても items を PATCH に含め保存できる（A-S02-01）", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } = await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "結果入力済み" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [
        {
          id: "101",
          examTypeFieldId: 1,
          name: "WBC",
          result: "",
          inspectionValue: "5.0",
          normalValue: "",
          unit: "",
          referenceValue: "",
          refMin: undefined,
          refMax: undefined,
          isAbnormal: false,
          status: "normal" as const,
          sortOrder: 1,
        },
      ],
      isSuccess: true,
      isError: false,
    } as ReturnType<typeof useGetExaminationItems>);

    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");

    await waitFor(() => expect(result.current.isPersistedConfirmed).toBe(false));

    await act(async () => {
      result.current.setFormData({ status: "確定" });
    });

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-001",
        req: expect.objectContaining({
          status: "confirmed",
          items: expect.arrayContaining([
            expect.objectContaining({ name: "WBC", inspection_value: "5.0" }),
          ]),
        }),
      }),
    );
    // ドラフト選択だけでは isPersistedConfirmed は true にならない
    expect(result.current.isPersistedConfirmed).toBe(false);
  });

  it("新規保存時に表示値を保持し判定境界を含まない items を POST する", async () => {
    const { useCreateExamination } = await import("../api/create-examination");
    const { useUpdateExaminationItems } = await import("../api/update-examination-items");
    const { useGetExamTypeFields } = await import("../api/get-exam-type-fields");

    // テンプレ展開で formItems が初期化される
    vi.mocked(useGetExamTypeFields).mockReturnValue({
      data: [
        {
          id: 1,
          name: "WBC",
          unit: "x10^3/μL",
          normalValue: "4.0-12.0",
          sortOrder: 1,
        },
      ],
    } as ReturnType<typeof useGetExamTypeFields>);

    const createMutate = vi.fn().mockResolvedValue({ id: "new-99" });
    const updateItemsMutate = vi.fn().mockResolvedValue([]);
    vi.mocked(useCreateExamination).mockReturnValue({
      mutateAsync: createMutate,
    } as ReturnType<typeof useCreateExamination>);
    vi.mocked(useUpdateExaminationItems).mockReturnValue({
      mutateAsync: updateItemsMutate,
      isPending: false,
    } as ReturnType<typeof useUpdateExaminationItems>);

    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [
        {
          id: "42",
          name: "ポチ",
          ownerName: "田中",
          ownerId: "5",
          species: "犬",
          breed: "",
          birthday: "",
          gender: "男",
          weight: null,
          imageUrl: null,
          status: "生存" as const,
          microchipNumber: null,
          insuranceNumber: null,
          insuranceExpiry: null,
          memo: null,
        },
      ],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);

    const { result } = renderExaminationForm();

    act(() => {
      result.current.setFormData({ testTypeId: "5", doctorId: "3" });
    });

    // テンプレが反映されるのを待つ（useEffect で自動展開）
    await act(async () => {
      await Promise.resolve();
    });

    // 値を入力
    if (result.current.formItems.length > 0) {
      act(() => {
        result.current.setInspectionValue(result.current.formItems[0].key, "7.5");
      });
    }

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(createMutate).toHaveBeenCalledOnce());
    expect(createMutate).toHaveBeenCalledWith({
      medical_record_id: null,
      pet_id: 42,
      exam_type_id: 5,
      doctor_id: 3,
      date: jstDateStartISOString(todayJSTISO()),
      result_summary: undefined,
      machine: undefined,
      items: [
        {
          exam_type_field_id: 1,
          name: "WBC",
          inspection_value: "7.5",
          normal_value: "4.0-12.0",
          unit: "x10^3/μL",
          reference_value: "4.0-12.0",
          sort_order: 1,
        },
      ],
    });
    expect(updateItemsMutate).not.toHaveBeenCalled();
  });

  it("新規保存時に items が空でも空配列を POST に含める", async () => {
    const { useCreateExamination } = await import("../api/create-examination");
    const { useGetExamTypeFields } = await import("../api/get-exam-type-fields");

    const createMutate = vi.fn().mockResolvedValue({ id: "new-99" });
    vi.mocked(useCreateExamination).mockReturnValue({
      mutateAsync: createMutate,
    } as ReturnType<typeof useCreateExamination>);
    vi.mocked(useGetExamTypeFields).mockReturnValue({
      data: undefined,
    } as ReturnType<typeof useGetExamTypeFields>);
    vi.mocked(usePetSelection).mockReturnValue({
      selectedPets: [selectedPet("生存")],
      setSelectedPets: vi.fn(),
    } as ReturnType<typeof usePetSelection>);

    const { result } = renderExaminationForm();
    act(() => {
      result.current.setFormData({ testTypeId: "5", doctorId: "3" });
    });

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(createMutate).toHaveBeenCalledOnce());
    expect(createMutate).toHaveBeenCalledWith(expect.objectContaining({ items: [] }));
  });
});
