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

const { mockCreateVaccination, vaccinesMasterState } = vi.hoisted(() => ({
  mockCreateVaccination: vi.fn(),
  // EMR-430: 種別マスタの interval はテストごとに差し替える（setVaccinesMasterMock）。
  vaccinesMasterState: {
    data: [] as { id: string; name: string; isActive: boolean; interval: string }[],
    isError: false,
  },
}));

vi.mock("@/hooks/use-treatment-master", () => ({
  useGetAllVaccinesMaster: () => ({
    data: vaccinesMasterState.data,
    isError: vaccinesMasterState.isError,
    refetch: vi.fn(),
  }),
}));

function setVaccinesMasterMock(
  data: { id: string; name: string; isActive: boolean; interval: string }[],
  isError = false,
) {
  vaccinesMasterState.data = data;
  vaccinesMasterState.isError = isError;
}

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
  setVaccinesMasterMock([]);
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

// EMR-430（2026-10-20 本人回答「ワクチン種類ごとの設定を使う。設定がなければ空欄。」）:
// カルテの予防接種タブの追加フォームは種別未選択では次回予定を空欄で開く。
// 旧 EMR-431 の「当日＋既定間隔」自動入力と、旧 EMR-430 の一律 1year は撤回。
// 種別選択時にマスタ interval（"1年"/"1ヶ月"）へ従属し、設定なし・取得不定では
// 既存値を保持する。
describe("useMedicalRecordVaccinationForm EMR-430 next schedule from vaccine master interval", () => {
  it("初期表示では type/date とも空欄（一律間隔を補わない）", () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    expect(result.current.date).toBe("2026-08-29");
    expect(result.current.nextScheduleType).toBe("");
    expect(result.current.nextDate).toBe("");
  });

  it("interval 1年 の種別を選ぶと 1year で当日から再計算する", () => {
    setVaccinesMasterMock([{ id: "7", name: "混合ワクチン", isActive: true, interval: "1年" }]);
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.setVaccineName("7");
    });

    expect(result.current.nextScheduleType).toBe("1year");
    // 2026-08-29 + 1年 = 2027-08-29
    expect(result.current.nextDate).toBe("2027-08-29");
  });

  it("interval 未設定 の種別を選ぶと空欄のまま保つ（捏造しない）", () => {
    setVaccinesMasterMock([{ id: "7", name: "ワクチンエキゾ", isActive: true, interval: "" }]);
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.setVaccineName("7");
    });

    expect(result.current.nextScheduleType).toBe("");
    expect(result.current.nextDate).toBe("");
  });

  it("マスタ取得中・失敗で種別が解決できない場合、既存の type/date を保持する", () => {
    setVaccinesMasterMock([]);
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.setNextScheduleType("1year");
    });
    act(() => {
      result.current.setNextDate("2027-08-29");
    });
    act(() => {
      result.current.setVaccineName("7");
    });

    expect(result.current.nextScheduleType).toBe("1year");
    expect(result.current.nextDate).toBe("2027-08-29");
  });

  it("設定あり種別で計算した後、手入力で nextDate を変えると other になる", () => {
    setVaccinesMasterMock([{ id: "7", name: "混合ワクチン", isActive: true, interval: "1年" }]);
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.setVaccineName("7");
    });
    act(() => {
      result.current.setNextDate("2028-01-10");
    });

    expect(result.current.nextScheduleType).toBe("other");
    expect(result.current.nextDate).toBe("2028-01-10");
  });

  // AC-vaccinations-manual: 明示手入力の次回予定日は種別変更で上書きしない
  it("手入力の nextDate（other＋日付あり）は種別を変更しても保持される", () => {
    setVaccinesMasterMock([
      { id: "7", name: "混合ワクチン", isActive: true, interval: "1年" },
      { id: "8", name: "ワクチンエキゾ", isActive: true, interval: "" },
      { id: "9", name: "狂犬病ワクチン", isActive: true, interval: "1ヶ月" },
    ]);
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.setNextScheduleType("other");
    });
    act(() => {
      result.current.setNextDate("2028-02-01");
    });
    act(() => {
      result.current.setVaccineName("8");
    });
    expect(result.current.nextScheduleType).toBe("other");
    expect(result.current.nextDate).toBe("2028-02-01");

    act(() => {
      result.current.setVaccineName("9");
    });
    expect(result.current.nextScheduleType).toBe("other");
    expect(result.current.nextDate).toBe("2028-02-01");
  });

  // AC-vaccinations-no-setting-default: 取得失敗は「設定なし」と区別して再試行導線を出す
  it("マスタ取得失敗時は isVaccinesMasterError を露出する", () => {
    setVaccinesMasterMock([], true);
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    expect(result.current.isVaccinesMasterError).toBe(true);
    expect(typeof result.current.retryVaccinesMaster).toBe("function");
  });

  it("空欄のまま保存すると next_date=null で送り、保存後も空欄へ戻る", async () => {
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
        next_date: null,
        next_schedule_type: undefined,
      }),
    );

    // 保存後の resetForm も空欄へ戻る（前レコードの値を残さない）
    expect(result.current.isAdding).toBe(false);
    expect(result.current.nextScheduleType).toBe("");
    expect(result.current.nextDate).toBe("");
  });
});

function historyItem(overrides: Partial<VaccinationHistoryItem> = {}): VaccinationHistoryItem {
  return {
    id: 1,
    name: "ワクチンA",
    // 本番形状: date は表示形式（YY/M/D）、dateISO が ISO（VaccinationHistory 参照）
    date: "26/8/29",
    dateISO: "2026-08-29",
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

// EMR-430/EMR-438: 複製は次回予定日だけを写し、間隔は複製元の実施日を基準に再判定する。
// 間隔が解決できる（例: 実施日+1年 = 次回予定日）場合は元の間隔を維持し、新しい実施日
// 入力時に次回予定日を自動再計算する。解決できない（手動）場合は写した日付を残さない。
describe("useMedicalRecordVaccinationForm EMR-430/438 duplicate rejudges interval", () => {
  it("実施日+1年で整合する次回予定日の複製は間隔 1year を維持し、写した日付を表示する", () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.handleDuplicate(
        historyItem({ date: "25/12/24", dateISO: "2025-12-24", nextDate: "2026-12-24" }),
      );
    });

    expect(result.current.isAdding).toBe(true);
    expect(result.current.nextDate).toBe("2026-12-24");
    expect(result.current.nextScheduleType).toBe("1year");
  });

  // 本番の履歴 row は date が表示形式（YY/M/D）で dateISO が ISO。表示形式を渡すと
  // calculateNextDate が解釈できず常に other（＝写した日付が消える）ため、
  // dateISO 経由で解決することを明示的に固定する（Codex P1 回帰）。
  it("dateISO が無く表示形式 date のみの履歴は解決できず other＋日付クリアになる", () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    const item = historyItem({ date: "25/12/24", nextDate: "2026-12-24" });
    delete (item as Partial<VaccinationHistoryItem>).dateISO;

    act(() => {
      result.current.handleDuplicate(item);
    });

    expect(result.current.nextScheduleType).toBe("other");
    expect(result.current.nextDate).toBe("");
  });

  it("間隔が解決できない複製は「以外（手動）」になり、写した日付はクリアする", () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.handleDuplicate(historyItem({ nextDate: "2026-12-24" }));
    });

    expect(result.current.isAdding).toBe(true);
    expect(result.current.nextDate).toBe("");
    expect(result.current.nextScheduleType).toBe("other");
  });

  it("間隔を維持した複製では、新しい実施日入力時に次回予定日が再計算される", () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));

    act(() => {
      result.current.handleDuplicate(
        historyItem({ date: "25/12/24", dateISO: "2025-12-24", nextDate: "2026-12-24" }),
      );
    });
    act(() => {
      result.current.setDate("2026-10-19");
    });

    expect(result.current.nextDate).toBe("2027-10-19");
    expect(result.current.nextScheduleType).toBe("1year");
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
// EMR-430（2026-10-20 本人回答）: 初期 type は空欄。プリセットへ戻した時点で
// 当日から再計算される。
describe("useMedicalRecordVaccinationForm レビュー指摘 — other 切替で次回予定日をクリア", () => {
  it("other に切り替えても空欄を保ち、プリセットに戻すと当日から再計算する", () => {
    const { result } = renderHook(() => useMedicalRecordVaccinationForm("1", "99"));
    expect(result.current.nextScheduleType).toBe("");
    expect(result.current.nextDate).toBe("");

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
