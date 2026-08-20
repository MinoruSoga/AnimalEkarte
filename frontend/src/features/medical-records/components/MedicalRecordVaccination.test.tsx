import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

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

vi.mock("./VaccinationForm", () => ({
  VaccinationForm: ({
    setVaccineName,
    setDate,
    setSupplemental,
    setNextScheduleType,
    fieldErrors,
    onSave,
  }: {
    setVaccineName: (value: string) => void;
    setDate: (value: string) => void;
    setSupplemental: (value: string) => void;
    setNextScheduleType: (value: string) => void;
    fieldErrors?: Record<string, string>;
    onSave?: () => void;
  }) => (
    <div data-testid="vaccination-form">
      <input aria-label="ワクチンID" onChange={(event) => setVaccineName(event.target.value)} />
      <input aria-label="接種日" onChange={(event) => setDate(event.target.value)} />
      <input aria-label="補助説明" onChange={(event) => setSupplemental(event.target.value)} />
      <input aria-label="次回予定種別" onChange={(event) => setNextScheduleType(event.target.value)} />
      {fieldErrors?.vaccineId ? (
        <p role="alert">{fieldErrors.vaccineId}</p>
      ) : null}
      {fieldErrors?.date ? (
        <p role="alert">{fieldErrors.date}</p>
      ) : null}
      <button type="button" onClick={onSave}>
        保存
      </button>
    </div>
  ),
}));

vi.mock("./VaccinationHistory", () => ({
  VaccinationHistory: () => <div data-testid="vaccination-history" />,
}));

beforeEach(() => {
  mockCreateVaccination.mockReset();
  mockCreateVaccination.mockResolvedValue({});
  mockHistoryItems.current = [];
});

function openAddForm() {
  fireEvent.click(screen.getByRole("button", { name: "記録を追加" }));
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
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

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

describe("MedicalRecordVaccination BUG-015 required validation", () => {
  it("ワクチン未選択のまま追加すると明示エラーを出し API を呼ばない", async () => {
    render(<MedicalRecordVaccination petId="1" medicalRecordId="99" />);
    openAddForm();

    fireEvent.change(screen.getByLabelText("接種日"), { target: { value: "2026-07-20" } });
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "ワクチン種別を選択してください",
    );
    expect(mockCreateVaccination).not.toHaveBeenCalled();
  });

  it("接種日未入力のまま追加すると明示エラーを出し API を呼ばない", async () => {
    render(<MedicalRecordVaccination petId="1" medicalRecordId="99" />);
    openAddForm();

    fireEvent.change(screen.getByLabelText("ワクチンID"), { target: { value: "7" } });
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("接種日を入力してください");
    expect(mockCreateVaccination).not.toHaveBeenCalled();
  });

  it("ワクチンと接種日を選択すると create が呼ばれ成功する", async () => {
    render(<MedicalRecordVaccination petId="1" medicalRecordId="99" />);
    openAddForm();

    fireEvent.change(screen.getByLabelText("ワクチンID"), { target: { value: "7" } });
    fireEvent.change(screen.getByLabelText("接種日"), { target: { value: "2026-07-20" } });
    fireEvent.click(screen.getByRole("button", { name: "保存" }));

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
