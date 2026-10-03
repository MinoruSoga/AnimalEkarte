import { startTransition } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";

import { useMedicalRecordVaccinationForm } from "./use-medical-record-vaccination-form";

// useActionState の formAction は <form action> 経由の呼び出しを前提とするため、
// フックのみを直接テストする場合は startTransition でラップしないと
// 「called outside of a transition」警告と状態更新の取りこぼしが発生する。
function submit(formAction: () => void) {
  act(() => {
    startTransition(() => {
      formAction();
    });
  });
}

const { mockCreateVaccination } = vi.hoisted(() => ({
  mockCreateVaccination: vi.fn(),
}));

vi.mock("@/hooks/use-treatment-master", () => ({
  useGetAllVaccinesMaster: () => ({ data: [] }),
}));

vi.mock("@/hooks/use-create-vaccination", () => ({
  useCreateVaccination: () => ({ mutateAsync: mockCreateVaccination }),
}));

vi.mock("../api/get-pet-vaccinations", () => ({
  useGetPetVaccinations: () => ({ data: [], isLoading: false }),
}));

vi.mock("@/lib/jst-date", () => ({
  todayJSTISO: () => "2026-08-29",
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

beforeEach(() => {
  mockCreateVaccination.mockReset();
  mockCreateVaccination.mockResolvedValue({});
  vi.mocked(toast.success).mockClear();
});

describe("useMedicalRecordVaccinationForm BUG-501 default date", () => {
  it("接種日は JST 当日で初期化される", () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));
    expect(result.current.date).toBe("2026-08-29");
  });
});

describe("useMedicalRecordVaccinationForm BUG-015 required validation", () => {
  it("isAdding が false の間に formAction が呼ばれても create を呼ばない（履歴検索 Enter 誤送信ガード）", async () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    submit(result.current.formAction);

    await waitFor(() => {
      expect(result.current.isSaving).toBe(false);
    });
    expect(mockCreateVaccination).not.toHaveBeenCalled();
  });

  it("ワクチン未選択のまま保存すると明示エラーを出し create を呼ばない", async () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.setIsAdding(true);
    });
    act(() => {
      result.current.setDate("2026-07-20");
    });
    submit(result.current.formAction);

    await waitFor(() => {
      expect(result.current.fieldErrors.vaccineId).toBe("ワクチン種別を選択してください");
    });
    expect(mockCreateVaccination).not.toHaveBeenCalled();
  });

  it("接種日を明示クリアしたまま保存すると明示エラーを出し create を呼ばない", async () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.setIsAdding(true);
    });
    act(() => {
      result.current.setVaccineName("7");
    });
    act(() => {
      result.current.setDate("");
    });
    submit(result.current.formAction);

    await waitFor(() => {
      expect(result.current.fieldErrors.date).toBe("接種日を入力してください");
    });
    expect(mockCreateVaccination).not.toHaveBeenCalled();
  });
});

describe("useMedicalRecordVaccinationForm save payload", () => {
  it("supplemental / next_schedule_type を含めて create を呼ぶ（サイレント消失防止）", async () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.setIsAdding(true);
    });
    act(() => {
      result.current.setVaccineName("7");
    });
    act(() => {
      result.current.setDate("2026-07-20");
    });
    act(() => {
      result.current.setSupplemental("補助説明テキスト");
    });
    act(() => {
      result.current.setNextScheduleType("3weeks");
    });
    submit(result.current.formAction);

    await waitFor(() => {
      expect(mockCreateVaccination).toHaveBeenCalledWith(
        expect.objectContaining({
          pet_id: 1,
          medical_record_id: 99,
          vaccine_id: 7,
          date: "2026-07-20",
          supplemental: "補助説明テキスト",
          next_schedule_type: "3weeks",
        }),
      );
    });
  });

  it("保存に成功すると成功トーストを出しフォームをリセットする", async () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.setIsAdding(true);
    });
    act(() => {
      result.current.setVaccineName("7");
    });
    act(() => {
      result.current.setDate("2026-07-20");
    });
    submit(result.current.formAction);

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith("接種記録を追加しました");
    });
    expect(result.current.isAdding).toBe(false);
    expect(result.current.vaccineName).toBe("");
    expect(result.current.date).toBe("2026-08-29");
  });

  it("保存に失敗すると成功トーストを出さずフォームを残す", async () => {
    mockCreateVaccination.mockRejectedValue(new Error("create failed"));
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.setIsAdding(true);
    });
    act(() => {
      result.current.setVaccineName("7");
    });
    act(() => {
      result.current.setDate("2026-07-20");
    });
    submit(result.current.formAction);

    await waitFor(() => {
      expect(mockCreateVaccination).toHaveBeenCalled();
    });
    expect(toast.success).not.toHaveBeenCalled();
    expect(result.current.isAdding).toBe(true);
    expect(result.current.vaccineName).toBe("7");
  });
});

describe("useMedicalRecordVaccinationForm EMR-105 same-day sequential registrations", () => {
  it("同日2件の順次登録で各 create payload が独立する（lot/next_date が混ざらない）", async () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    // 1件目: vaccine 7 / LOT-A。接種日は JST 当日 default（resetForm 後も当日に戻る）。
    act(() => {
      result.current.setIsAdding(true);
    });
    act(() => {
      result.current.setVaccineName("7");
    });
    act(() => {
      result.current.setLot1("LOT-A");
    });
    act(() => {
      result.current.setNextDate("2026-09-26");
    });
    act(() => {
      result.current.setRemarks("1回目メモ");
    });
    submit(result.current.formAction);

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledTimes(1);
    });
    // 成功後はリセット済み — 2件目は新規入力として開き直す（接種日は当日に戻る＝同日）。
    expect(result.current.isAdding).toBe(false);
    expect(result.current.lot1).toBe("");

    // 2件目: vaccine 8 / LOT-B — 同日（当日 default）のまま登録。
    act(() => {
      result.current.setIsAdding(true);
    });
    act(() => {
      result.current.setVaccineName("8");
    });
    act(() => {
      result.current.setLot1("LOT-B");
    });
    act(() => {
      result.current.setNextDate("2026-10-10");
    });
    act(() => {
      result.current.setRemarks("2回目メモ");
    });
    submit(result.current.formAction);

    await waitFor(() => {
      expect(mockCreateVaccination).toHaveBeenCalledTimes(2);
    });

    expect(mockCreateVaccination).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({
        pet_id: 1,
        medical_record_id: 99,
        vaccine_id: 7,
        date: "2026-08-29",
        lot1: "LOT-A",
        next_date: "2026-09-26",
        remarks: "1回目メモ",
      }),
    );
    expect(mockCreateVaccination).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({
        pet_id: 1,
        medical_record_id: 99,
        vaccine_id: 8,
        date: "2026-08-29",
        lot1: "LOT-B",
        next_date: "2026-10-10",
        remarks: "2回目メモ",
      }),
    );
  });
});
