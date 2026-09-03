import { renderHook, act, waitFor } from "@testing-library/react";
import { startTransition } from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { useExaminationForm } from "./use-examination-form";
import { useSearchParams } from "react-router";
import { useGetPet } from "@/hooks/use-pet";
import { usePetSelection } from "@/hooks/use-pet-selection";

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

  it("初期 formItems は空配列", () => {
    const { result } = renderExaminationForm();
    expect(result.current.formItems).toEqual([]);
  });

  it("編集モードで既存 items を formItems に反映する", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } = await import("../api/get-examination-items");
    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
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
          normalValue: "4.0-12.0",
          unit: "x10^3/μL",
          referenceValue: "4.0-12.0",
          refMin: 4,
          refMax: 12,
          isAssessed: true,
          isAbnormal: false,
          status: "normal" as const,
          sortOrder: 1,
        },
      ],
      isSuccess: true,
      isError: false,
    } as ReturnType<typeof useGetExaminationItems>);

    const { result } = renderExaminationForm("exam-001");
    expect(result.current.formItems).toHaveLength(1);
    expect(result.current.formItems[0].name).toBe("WBC");
    expect(result.current.formItems[0].inspectionValue).toBe("5.0");
    expect(result.current.formItems[0].status).toBe("normal");
    expect(result.current.formItems[0].isAssessed).toBe(true);
  });

  it("setInspectionValue で指定 row の inspectionValue を更新できる", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } = await import("../api/get-examination-items");
    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
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
          inspectionValue: "",
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

    const { result } = renderExaminationForm("exam-001");
    const key = result.current.formItems[0].key;
    act(() => {
      result.current.setInspectionValue(key, "7.5");
    });
    expect(result.current.formItems[0].inspectionValue).toBe("7.5");
  });

  it("手動行をimmutableに追加・改名・削除できる", () => {
    const { result } = renderExaminationForm();
    const before = result.current.formItems;

    act(() => {
      result.current.addManualItem();
    });
    const afterAdd = result.current.formItems;

    expect(afterAdd).not.toBe(before);
    expect(before).toEqual([]);
    expect(afterAdd).toHaveLength(1);
    expect(afterAdd[0].examTypeFieldId).toBeUndefined();
    expect(afterAdd[0]).toMatchObject({ name: "", inspectionValue: "" });

    const key = afterAdd[0].key;
    act(() => {
      result.current.setItemName(key, "手動項目");
    });
    expect(result.current.formItems).not.toBe(afterAdd);
    expect(afterAdd[0].name).toBe("");
    expect(result.current.formItems[0].name).toBe("手動項目");

    const afterRename = result.current.formItems;
    act(() => {
      result.current.removeItem(key);
    });
    expect(result.current.formItems).not.toBe(afterRename);
    expect(afterRename).toHaveLength(1);
    expect(result.current.formItems).toEqual([]);
  });

  it("追加した手動行の名前と結果値をPATCH itemsへ送る", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } = await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        petId: "41",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [],
      isSuccess: true,
      isError: false,
      isPending: false,
    } as ReturnType<typeof useGetExaminationItems>);
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");
    act(() => {
      result.current.addManualItem();
    });
    const key = result.current.formItems[0].key;
    act(() => {
      result.current.setItemName(key, "手動項目");
      result.current.setInspectionValue(key, "記録値");
    });

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-001",
        req: expect.objectContaining({
          items: expect.arrayContaining([
            expect.objectContaining({
              exam_type_field_id: null,
              name: "手動項目",
              inspection_value: "記録値",
            }),
          ]),
        }),
      }),
    );
  });

  it("結果値がある手動行の空名を拒否し、silent drop しない", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } = await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        petId: "41",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [],
      isSuccess: true,
      isError: false,
    } as ReturnType<typeof useGetExaminationItems>);
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");
    act(() => result.current.addManualItem());
    const key = result.current.formItems[0].key;
    act(() => result.current.setInspectionValue(key, "記録値"));

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    expect(updateMutate).not.toHaveBeenCalled();
    expect(result.current.formState.fieldErrors).toEqual({
      examItems: "結果値を入力した手動項目には項目名が必要です",
    });
  });

  it("編集モード保存時に表示値を保持し判定境界を含まない items を PATCH する", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } = await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");
    const { useUpdateExaminationItems } = await import("../api/update-examination-items");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
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
          normalValue: "4.0-12.0",
          unit: "x10^3/μL",
          referenceValue: "4.0-12.0",
          refMin: 4,
          refMax: 12,
          isAbnormal: false,
          status: "normal" as const,
          sortOrder: 1,
        },
      ],
      isSuccess: true,
      isError: false,
    } as ReturnType<typeof useGetExaminationItems>);

    const updateMutate = vi.fn().mockResolvedValue({});
    const updateItemsMutate = vi.fn().mockResolvedValue([]);
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);
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

    const { result } = renderExaminationForm("exam-001");

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-001",
        req: expect.objectContaining({
          items: [
            {
              exam_type_field_id: 1,
              name: "WBC",
              inspection_value: "5.0",
              normal_value: "4.0-12.0",
              unit: "x10^3/μL",
              reference_value: "4.0-12.0",
              sort_order: 1,
            },
          ],
        }),
      }),
    );
    expect(updateItemsMutate).not.toHaveBeenCalled();
  });

  it("編集モードで items が空でも空配列を PATCH に含める", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } = await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue({
      data: [],
      isSuccess: true,
      isError: false,
      isPending: false,
    } as ReturnType<typeof useGetExaminationItems>);

    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-001",
        req: expect.objectContaining({ items: [] }),
      }),
    );
  });

  it.each([
    [
      "取得中",
      {
        data: undefined,
        isSuccess: false,
        isError: false,
        isPending: true,
      },
    ],
    [
      "取得失敗",
      {
        data: [],
        isSuccess: false,
        isError: true,
        isPending: false,
        error: new Error("items unavailable"),
      },
    ],
  ])("編集モードで検査項目が%sなら保存 mutation を発行しない", async (_label, itemQueryResult) => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } = await import("../api/get-examination-items");
    const { useUpdateExamination } = await import("../api/update-examination");

    vi.mocked(useGetExamination).mockReturnValue({
      data: {
        id: "exam-001",
        testTypeId: "5",
        doctorId: "3",
        status: "検査中" as const,
        ownerName: "",
        petName: "",
        date: "",
      },
    } as ReturnType<typeof useGetExamination>);
    vi.mocked(useGetExaminationItems).mockReturnValue(
      itemQueryResult as ReturnType<typeof useGetExaminationItems>,
    );
    const updateMutate = vi.fn().mockResolvedValue({});
    vi.mocked(useUpdateExamination).mockReturnValue({
      mutateAsync: updateMutate,
    } as ReturnType<typeof useUpdateExamination>);

    const { result } = renderExaminationForm("exam-001");
    await waitFor(() => expect(result.current.formData.testTypeId).toBe("5"));

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });

    await waitFor(() =>
      expect(result.current.formState.fieldErrors).toEqual({
        examItems: "検査項目の読み込み完了後に保存してください",
      }),
    );
    expect(updateMutate).not.toHaveBeenCalled();
  });

  it("検査ID変更時は旧itemsを破棄し、新IDの取得完了後だけ新itemsをPATCHする", async () => {
    const { useGetExamination } = await import("../api/get-examination");
    const { useGetExaminationItems } = await import("../api/get-examination-items");
    const { useGetExamTypeFields } = await import("../api/get-exam-type-fields");
    const { useUpdateExamination } = await import("../api/update-examination");

    const examinationFor = (examID: string) => ({
      id: examID,
      testTypeId: examID === "exam-a" ? "5" : "6",
      doctorId: examID === "exam-a" ? "3" : "4",
      status: "検査中" as const,
      ownerName: "",
      petName: "",
      date: examID === "exam-a" ? "2026-08-01" : "2026-08-02",
      resultSummary: examID === "exam-a" ? "server summary A" : "server summary B",
      machine: examID === "exam-a" ? "server machine A" : "server machine B",
    });
    const itemFor = (id: string, name: string, inspectionValue: string) => ({
      id,
      examTypeFieldId: 1,
      name,
      result: "",
      inspectionValue,
      normalValue: "",
      unit: "",
      referenceValue: "",
      refMin: undefined,
      refMax: undefined,
      isAbnormal: false,
      status: "normal" as const,
      sortOrder: 1,
    });
    let examinationBItemsReady = false;
    vi.mocked(useGetExamination).mockImplementation(
      (examID) => ({ data: examinationFor(examID) }) as ReturnType<typeof useGetExamination>,
    );
    vi.mocked(useGetExaminationItems).mockImplementation((examID) => {
      if (examID === "exam-a") {
        return {
          data: [itemFor("item-a", "A-WBC", "1")],
          isSuccess: true,
          isError: false,
          isPending: false,
        } as ReturnType<typeof useGetExaminationItems>;
      }
      if (!examinationBItemsReady) {
        return {
          data: undefined,
          isSuccess: false,
          isError: false,
          isPending: true,
        } as ReturnType<typeof useGetExaminationItems>;
      }
      return {
        data: [itemFor("item-b", "B-WBC", "2")],
        isSuccess: true,
        isError: false,
        isPending: false,
      } as ReturnType<typeof useGetExaminationItems>;
    });
    vi.mocked(useGetExamTypeFields).mockImplementation(
      (examTypeID) =>
        ({
          data: [
            {
              id: Number(examTypeID) * 100,
              name: `TEMPLATE-${examTypeID}`,
              unit: "",
              normalValue: "",
              sortOrder: 1,
            },
          ],
        }) as ReturnType<typeof useGetExamTypeFields>,
    );
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

    act(() => {
      result.current.setFormData({
        testTypeId: "7",
        status: "確定",
        resultSummary: "unsaved summary A",
        machine: "unsaved machine A",
      });
    });
    await waitFor(() => expect(result.current.formData.testTypeId).toBe("7"));

    rerender({ examID: "exam-b" });
    expect(result.current.formData).toMatchObject({
      testTypeId: "6",
      doctorId: "4",
      status: "検査中",
      resultSummary: "server summary B",
      machine: "server machine B",
      date: "2026-08-02",
    });
    expect(result.current.formItems).toEqual([]);
    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });
    await waitFor(() =>
      expect(result.current.formState.fieldErrors?.examItems).toBe(
        "検査項目の読み込み完了後に保存してください",
      ),
    );
    expect(updateMutate).not.toHaveBeenCalled();

    examinationBItemsReady = true;
    rerender({ examID: "exam-b" });
    await waitFor(() => expect(result.current.formItems[0]?.name).toBe("B-WBC"));
    expect(result.current.formItems[0]?.name).not.toBe("TEMPLATE-6");

    await act(async () => {
      startTransition(() => result.current.formAction(new FormData()));
    });
    await waitFor(() => expect(updateMutate).toHaveBeenCalledOnce());
    expect(updateMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        id: "exam-b",
        req: expect.objectContaining({
          status: "in_progress",
          result_summary: "server summary B",
          machine: "server machine B",
          items: [expect.objectContaining({ name: "B-WBC" })],
        }),
      }),
    );
  });
});
