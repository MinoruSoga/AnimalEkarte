import { startTransition } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";

import { useMedicalRecordVaccinationForm } from "./use-medical-record-vaccination-form";
import type { VaccinationHistoryItem } from "../components/VaccinationHistory";

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

// EMR-431: カルテの予防接種タブの追加フォームも、開いた時点で「当日＋既定間隔」の
// 次回予定日を持つ。日付に触れずに保存しても next_date が送られる
// （EMR-427 の LINE 期限配信は next_date で判定するため、空のまま保存されると漏れる）。
// EMR-430: 既定の間隔を独立画面・健診タブと同じ 1年後（DEFAULT_NEXT_SCHEDULE_TYPE）にそろえる。
describe("useMedicalRecordVaccinationForm EMR-431 next date default", () => {
  it("初期表示で、当日と既定間隔から計算した次回予定日を持つ", () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    expect(result.current.date).toBe("2026-08-29");
    expect(result.current.nextScheduleType).toBe("1year");
    // 2026-08-29 + 1年 = 2027-08-29
    expect(result.current.nextDate).toBe("2027-08-29");
  });

  it("日付に触れずに保存すると next_date を送り、保存後も同じ初期値に戻る", async () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.setIsAdding(true);
    });
    act(() => {
      result.current.setVaccineName("7");
    });
    submit(result.current.formAction);

    // 既存テストと同じく、create 完了後の resetForm まで進んだことを成功トーストで待つ
    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith("接種記録を追加しました");
    });
    expect(mockCreateVaccination).toHaveBeenCalledWith(
      expect.objectContaining({
        date: "2026-08-29",
        next_date: "2027-08-29",
        next_schedule_type: "1year",
      }),
    );

    // 保存後の resetForm も同じ「当日＋既定間隔」に戻る（前レコードの値を残さない）
    expect(result.current.isAdding).toBe(false);
    expect(result.current.date).toBe("2026-08-29");
    expect(result.current.nextScheduleType).toBe("1year");
    expect(result.current.nextDate).toBe("2027-08-29");
  });
});

function historyItem(overrides: Partial<VaccinationHistoryItem> = {}): VaccinationHistoryItem {
  return {
    id: 1,
    name: "ワクチンA",
    date: "2026-08-29",
    next: "",
    vaccineId: 7,
    lot1: "",
    lot2: "",
    lot3: "",
    lot4: "",
    nextDate: "",
    remarks: "",
    ...overrides,
  };
}

// EMR-430: 複製は次回予定日だけを写すため、写した日付と食い違う間隔をそのまま見せない。
// 次回予定日のある記録は「以外（手動）」、無い記録は間隔を変えない。
describe("useMedicalRecordVaccinationForm EMR-430 duplicate rejudges interval", () => {
  it("次回予定日のある記録を複製すると「以外（手動）」になり、写した日付を表示する", () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.handleDuplicate(historyItem({ nextDate: "2026-12-24" }));
    });

    expect(result.current.isAdding).toBe(true);
    expect(result.current.nextDate).toBe("2026-12-24");
    expect(result.current.nextScheduleType).toBe("other");
  });

  it("次回予定日の無い記録を複製すると間隔は変わらない", () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.setNextScheduleType("4weeks");
    });
    act(() => {
      result.current.handleDuplicate(historyItem({ nextDate: "" }));
    });

    expect(result.current.nextDate).toBe("");
    expect(result.current.nextScheduleType).toBe("4weeks");
  });
});

// レビュー指摘（EMR-431）: 間隔を「以外（手動）」に変えたとき、既定で自動計算された
// 次回予定日が残り、手入力していない日付が「手動」の値として保存されていた。
describe("useMedicalRecordVaccinationForm レビュー指摘 — other 切替で次回予定日をクリア", () => {
  it("other に切り替えたら既定の次回予定日をクリアし、プリセットに戻すと再計算する", () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));
    expect(result.current.nextScheduleType).toBe("1year");
    expect(result.current.nextDate).toBe("2027-08-29");

    act(() => {
      result.current.setNextScheduleType("other");
    });
    expect(result.current.nextScheduleType).toBe("other");
    expect(result.current.nextDate).toBe("");

    act(() => {
      result.current.setNextScheduleType("1year");
    });
    expect(result.current.nextScheduleType).toBe("1year");
    expect(result.current.nextDate).toBe("2027-08-29");
  });
});
