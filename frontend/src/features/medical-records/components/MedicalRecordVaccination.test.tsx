import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";

import { MedicalRecordVaccination } from "./MedicalRecordVaccination";

const { mockCreateVaccination } = vi.hoisted(() => ({
  mockCreateVaccination: vi.fn(),
}));

vi.mock("@/hooks/use-treatment-master", () => ({
  useGetAllVaccinesMaster: () => ({ data: [] }),
}));

vi.mock("@/hooks/use-vaccinations", () => ({
  useCreateVaccination: () => ({ mutateAsync: mockCreateVaccination }),
}));

const { mockHistoryItems } = vi.hoisted(() => ({
  mockHistoryItems: { current: [] as Array<{ id: number; name: string; date: string }> },
}));

vi.mock("../api/get-pet-vaccinations", () => ({
  useGetPetVaccinations: () => ({ data: mockHistoryItems.current, isLoading: false }),
}));

// FE-RC-008: useActionState + <form action> への移行に伴い、"保存" は
// type="submit" の素の button に変更。onSave コールバックは廃止し、実フォームの
// action (親コンポーネント側) を通す。
vi.mock("./VaccinationForm", () => ({
  VaccinationForm: ({
    vaccineName,
    setVaccineName,
    date,
    setDate,
    setSupplemental,
    setNextScheduleType,
    fieldErrors,
  }: {
    vaccineName: string;
    setVaccineName: (value: string) => void;
    date: string;
    setDate: (value: string) => void;
    setSupplemental: (value: string) => void;
    setNextScheduleType: (value: string) => void;
    fieldErrors?: Record<string, string>;
  }) => (
    <div data-testid="vaccination-form">
      <input
        aria-label="ワクチンID"
        value={vaccineName}
        onChange={(event) => setVaccineName(event.target.value)}
      />
      <input
        aria-label="接種日"
        value={date}
        onChange={(event) => setDate(event.target.value)}
      />
      <input aria-label="補助説明" onChange={(event) => setSupplemental(event.target.value)} />
      <input aria-label="次回予定種別" onChange={(event) => setNextScheduleType(event.target.value)} />
      {fieldErrors?.vaccineId ? (
        <p role="alert">{fieldErrors.vaccineId}</p>
      ) : null}
      {fieldErrors?.date ? (
        <p role="alert">{fieldErrors.date}</p>
      ) : null}
      <button type="submit">保存</button>
    </div>
  ),
}));

vi.mock("@/lib/jst-date", () => ({
  todayJSTISO: () => "2026-08-29",
}));

vi.mock("./VaccinationHistory", () => ({
  VaccinationHistory: () => <div data-testid="vaccination-history" />,
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
  mockHistoryItems.current = [];
  vi.mocked(toast.success).mockClear();
});

function openAddForm() {
  fireEvent.click(screen.getByRole("button", { name: "記録を追加" }));
}

async function submitForm() {
  const user = userEvent.setup();
  await user.click(screen.getByRole("button", { name: "保存" }));
}

describe("MedicalRecordVaccination left list (BUG-007)", () => {
  it("接種記録があるときは空状態ではなく一覧を表示する", () => {
    mockHistoryItems.current = [
      { id: 11, name: "混合ワクチン", date: "26/8/1" },
    ];
    render(<MedicalRecordVaccination petId="1" medicalRecordId="99" />);
    expect(screen.getByText("混合ワクチン")).toBeInTheDocument();
    expect(screen.queryByText(/接種記録がありません/)).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "記録を追加" })).toBeInTheDocument();
  });
});

describe("MedicalRecordVaccination responsive layout", () => {
  it("mobileではform/historyを縦積みし、lg以上で5列gridに戻る", () => {
    render(<MedicalRecordVaccination petId="1" medicalRecordId="99" />);

    openAddForm();
    const layout = screen.getByTestId("vaccination-form").parentElement;
    expect(layout).toHaveClass("grid-cols-1", "lg:grid-cols-5");
    expect(layout).not.toHaveClass("grid-cols-12");
    expect(screen.getByTestId("vaccination-history").parentElement).toBe(layout);
  });
});

describe("MedicalRecordVaccination vaccination payload", () => {
  it("embedded 保存 payload に supplemental と next_schedule_type を含める（サイレント消失防止）", async () => {
    render(<MedicalRecordVaccination petId="1" medicalRecordId="99" />);
    openAddForm();

    fireEvent.change(screen.getByLabelText("ワクチンID"), { target: { value: "7" } });
    fireEvent.change(screen.getByLabelText("接種日"), { target: { value: "2026-07-20" } });
    fireEvent.change(screen.getByLabelText("補助説明"), { target: { value: "補助説明テキスト" } });
    fireEvent.change(screen.getByLabelText("次回予定種別"), { target: { value: "3weeks" } });
    await submitForm();

    await waitFor(() => {
      expect(mockCreateVaccination).toHaveBeenCalledWith(
        expect.objectContaining({
          supplemental: "補助説明テキスト",
          next_schedule_type: "3weeks",
        }),
      );
    });
  });
});

describe("MedicalRecordVaccination BUG-501 実施日 default", () => {
  it("記録を追加で開いたフォームの接種日は JST 当日で初期表示される", () => {
    render(<MedicalRecordVaccination petId="1" medicalRecordId="99" />);
    openAddForm();

    expect(screen.getByLabelText("接種日")).toHaveValue("2026-08-29");
  });
});

describe("MedicalRecordVaccination BUG-015 required validation", () => {
  it("ワクチン未選択のまま追加すると明示エラーを出し API を呼ばない", async () => {
    render(<MedicalRecordVaccination petId="1" medicalRecordId="99" />);
    openAddForm();

    fireEvent.change(screen.getByLabelText("接種日"), { target: { value: "2026-07-20" } });
    await submitForm();

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "ワクチン種別を選択してください",
    );
    expect(mockCreateVaccination).not.toHaveBeenCalled();
  });

  it("接種日を明示クリアしたまま追加すると明示エラーを出し API を呼ばない", async () => {
    render(<MedicalRecordVaccination petId="1" medicalRecordId="99" />);
    openAddForm();

    fireEvent.change(screen.getByLabelText("ワクチンID"), { target: { value: "7" } });
    // BUG-501: 実施日は当日デフォルトのため、未入力検証は明示クリアが必要
    fireEvent.change(screen.getByLabelText("接種日"), { target: { value: "" } });
    await submitForm();

    expect(await screen.findByRole("alert")).toHaveTextContent("接種日を入力してください");
    expect(mockCreateVaccination).not.toHaveBeenCalled();
  });

  it("ワクチンと接種日を選択すると create が呼ばれ成功する", async () => {
    render(<MedicalRecordVaccination petId="1" medicalRecordId="99" />);
    openAddForm();

    fireEvent.change(screen.getByLabelText("ワクチンID"), { target: { value: "7" } });
    fireEvent.change(screen.getByLabelText("接種日"), { target: { value: "2026-07-20" } });
    await submitForm();

    await waitFor(() => {
      expect(mockCreateVaccination).toHaveBeenCalledWith(
        expect.objectContaining({
          pet_id: 1,
          medical_record_id: 99,
          vaccine_id: 7,
          date: "2026-07-20",
        }),
      );
    });
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});

describe("MedicalRecordVaccination BUG-001 inner save toast", () => {
  it("接種記録の追加に成功したら成功トーストを出す", async () => {
    render(<MedicalRecordVaccination petId="1" medicalRecordId="99" />);
    openAddForm();

    fireEvent.change(screen.getByLabelText("ワクチンID"), { target: { value: "7" } });
    fireEvent.change(screen.getByLabelText("接種日"), { target: { value: "2026-07-20" } });
    await submitForm();

    await waitFor(() => {
      expect(mockCreateVaccination).toHaveBeenCalled();
    });
    expect(toast.success).toHaveBeenCalledWith("接種記録を追加しました");
  });

  it("接種記録の追加が失敗したら成功トーストを出さずフォームを残す", async () => {
    mockCreateVaccination.mockRejectedValue(new Error("create failed"));
    render(<MedicalRecordVaccination petId="1" medicalRecordId="99" />);
    openAddForm();

    fireEvent.change(screen.getByLabelText("ワクチンID"), { target: { value: "7" } });
    fireEvent.change(screen.getByLabelText("接種日"), { target: { value: "2026-07-20" } });
    await submitForm();

    await waitFor(() => {
      expect(mockCreateVaccination).toHaveBeenCalled();
    });
    expect(toast.success).not.toHaveBeenCalled();
    expect(screen.getByTestId("vaccination-form")).toBeInTheDocument();
    expect(screen.getByLabelText("ワクチンID")).toBeInTheDocument();
    expect(screen.getByLabelText("接種日")).toBeInTheDocument();
  });
});
