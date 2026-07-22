import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { MedicalRecordVaccination } from "./MedicalRecordVaccination";

vi.mock("@/hooks/use-treatment-master", () => ({
  useGetAllVaccinesMaster: () => ({ data: [] }),
}));

vi.mock("@/hooks/use-vaccinations", () => ({
  useCreateVaccination: () => ({ mutateAsync: vi.fn() }),
}));

vi.mock("../api/get-pet-vaccinations", () => ({
  useGetPetVaccinations: () => ({ data: [], isLoading: false }),
}));

vi.mock("./VaccinationForm", () => ({
  VaccinationForm: () => <div data-testid="vaccination-form" />,
}));

vi.mock("./VaccinationHistory", () => ({
  VaccinationHistory: () => <div data-testid="vaccination-history" />,
}));

describe("MedicalRecordVaccination responsive layout", () => {
  it("mobileではform/historyを縦積みし、lg以上で既存の12列gridに戻る", () => {
    render(<MedicalRecordVaccination petId="1" />);

    const layout = screen.getByTestId("vaccination-form").parentElement;
    expect(layout).toHaveClass("grid-cols-1", "lg:grid-cols-12");
    expect(layout).not.toHaveClass("grid-cols-12");
    expect(screen.getByTestId("vaccination-history").parentElement).toBe(layout);
  });
});
