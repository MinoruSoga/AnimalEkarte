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

vi.mock("../api/get-pet-vaccinations", () => ({
  useGetPetVaccinations: () => ({ data: [], isLoading: false }),
}));

vi.mock("./VaccinationForm", () => ({
  VaccinationForm: ({
    setVaccineName,
    setDate,
    setSupplemental,
    setNextScheduleType,
    onSave,
  }: {
    setVaccineName: (value: string) => void;
    setDate: (value: string) => void;
    setSupplemental: (value: string) => void;
    setNextScheduleType: (value: string) => void;
    onSave?: () => void;
  }) => (
    <div data-testid="vaccination-form">
      <input aria-label="ワクチンID" onChange={(event) => setVaccineName(event.target.value)} />
      <input aria-label="接種日" onChange={(event) => setDate(event.target.value)} />
      <input aria-label="補助説明" onChange={(event) => setSupplemental(event.target.value)} />
      <input aria-label="次回予定種別" onChange={(event) => setNextScheduleType(event.target.value)} />
      <button type="button" onClick={onSave}>保存</button>
    </div>
  ),
}));

vi.mock("./VaccinationHistory", () => ({
  VaccinationHistory: () => <div data-testid="vaccination-history" />,
}));

beforeEach(() => {
  mockCreateVaccination.mockReset();
  mockCreateVaccination.mockResolvedValue({});
});

describe("MedicalRecordVaccination responsive layout", () => {
  it("mobileではform/historyを縦積みし、lg以上で既存の12列gridに戻る", () => {
    render(<MedicalRecordVaccination petId="1" />);

    const layout = screen.getByTestId("vaccination-form").parentElement;
    expect(layout).toHaveClass("grid-cols-1", "lg:grid-cols-12");
    expect(layout).not.toHaveClass("grid-cols-12");
    expect(screen.getByTestId("vaccination-history").parentElement).toBe(layout);
  });
});

describe("MedicalRecordVaccination vaccination payload", () => {
  it("embedded 保存 payload に supplemental と next_schedule_type を含める（サイレント消失防止）", async () => {
    render(<MedicalRecordVaccination petId="1" />);

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
