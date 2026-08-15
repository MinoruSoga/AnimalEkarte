import type { Pet } from "@/types";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ExaminationPatientChangeDialog } from "./ExaminationPatientChangeDialog";

const mocks = vi.hoisted(() => ({ selectionTable: vi.fn() }));

vi.mock(
  "@/components/shared/ReservationFormModal/PatientSelectionTable",
  () => ({
    PatientSelectionTable: (props: {
      includeDeceased?: boolean;
      onSelect: (pet: Pet) => void;
    }) => {
      mocks.selectionTable(props);
      const pet = {
        id: "84",
        clinicId: undefined,
        ownerId: "owner-test",
        ownerNumber: undefined,
        ownerName: "テスト飼主",
        ownerNameKana: undefined,
        address: undefined,
        phone: "",
        petNumber: undefined,
        name: "テスト患者",
        petNameKana: undefined,
        species: "犬",
        animalSpeciesId: undefined,
        status: "生存",
        breed: "",
        color: "",
        bloodType: undefined,
        microchipNumber: undefined,
        gender: "不明",
        birthDate: undefined,
        neuteredDate: undefined,
        weight: undefined,
        food: "",
        environment: "",
        acquisitionType: undefined,
        dangerLevel: undefined,
        dangerReason: "",
        lastVisit: undefined,
        insuranceId: undefined,
        insuranceName: undefined,
        insuranceDetails: undefined,
        remarks: "",
        deceasedAt: undefined,
      } satisfies Pet;
      return (
        <>
          <button type="button" onClick={() => props.onSelect(pet)}>
            テスト患者を選択
          </button>
          <button
            type="button"
            onClick={() => props.onSelect({ ...pet, status: "不明" })}
          >
            状態不明患者を選択
          </button>
        </>
      );
    },
  }),
);

describe("ExaminationPatientChangeDialog", () => {
  it("死亡sentinelを含む検索を要求し、生存患者の選択後に閉じる", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(
      <ExaminationPatientChangeDialog
        selectedPet={undefined}
        onSelect={onSelect}
      />,
    );

    await user.click(screen.getByRole("button", { name: "患者を変更" }));
    expect(mocks.selectionTable).toHaveBeenLastCalledWith(
      expect.objectContaining({ includeDeceased: true }),
    );
    await user.click(screen.getByRole("button", { name: "テスト患者を選択" }));

    expect(onSelect).toHaveBeenCalledWith(
      expect.objectContaining({ id: "84", status: "生存" }),
    );
    expect(
      screen.queryByRole("button", { name: "テスト患者を選択" }),
    ).not.toBeInTheDocument();
  });

  it("状態不明の患者を選択せずdialogを開いたままにする", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(
      <ExaminationPatientChangeDialog
        selectedPet={undefined}
        onSelect={onSelect}
      />,
    );

    await user.click(screen.getByRole("button", { name: "患者を変更" }));
    await user.click(
      screen.getByRole("button", { name: "状態不明患者を選択" }),
    );

    expect(onSelect).not.toHaveBeenCalled();
    expect(
      screen.getByRole("button", { name: "状態不明患者を選択" }),
    ).toBeInTheDocument();
  });
});
