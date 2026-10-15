import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { createPetFormData } from "../lib/pet-form-data";
import { PetIdentitySection } from "./PetIdentitySection";

describe("PetIdentitySection required accessibility", () => {
  it("エラー表示前からペット名・動物種・性別を必須と通知し任意欄は必須にしない", () => {
    render(
      <PetIdentitySection
        formData={createPetFormData()}
        setFormData={vi.fn()}
        fieldErrors={{}}
        clearFieldError={vi.fn()}
        animalSpeciesOptions={[{ value: "dog", label: "犬" }]}
        isLoadingSpecies={false}
        speciesPlaceholder="選択してください"
        isEdit={false}
        onAnimalSpeciesChange={vi.fn()}
      />,
    );

    const requiredControls = [
      screen.getByRole("textbox", { name: "ペット名" }),
      screen.getByRole("combobox", { name: "動物種" }),
      screen.getByRole("combobox", { name: "性別" }),
    ];
    for (const control of requiredControls) {
      expect(control).toHaveAttribute("aria-required", "true");
      expect(control).not.toHaveAttribute("required");
      expect(control).not.toHaveAttribute("aria-invalid", "true");
    }
    for (const name of ["ペット名よみ", "生年月日", "名前の由来", "出逢いのストーリー"]) {
      expect(screen.getByRole("textbox", { name })).not.toHaveAttribute("aria-required");
    }
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
