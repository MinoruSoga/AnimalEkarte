import type { Dispatch, SetStateAction } from "react";
import { useMemo } from "react";

import { SelectItem } from "@/components/ui/select";

import {
  PetCareSection,
  PetIdentitySection,
  PetPhysicalSection,
  type AnimalSpeciesOption,
  type InsuranceOption,
} from "./PetEditModalFieldSections";
import type { PetFormData } from "../types";

interface PetEditModalFieldsProps {
  formData: PetFormData;
  setFormData: Dispatch<SetStateAction<PetFormData>>;
  fieldErrors: Record<string, string>;
  clearFieldError: (field: string) => void;
  animalSpeciesList: AnimalSpeciesOption[];
  isLoadingSpecies: boolean;
  insuranceList: InsuranceOption[];
  isLoadingInsurances: boolean;
  canEdit: boolean;
  isEdit: boolean;
}

export function PetEditModalFields({
  formData,
  setFormData,
  fieldErrors,
  clearFieldError,
  animalSpeciesList,
  isLoadingSpecies,
  insuranceList,
  isLoadingInsurances,
  canEdit,
  isEdit,
}: PetEditModalFieldsProps) {
  const animalSpeciesSelectItems = useMemo(
    () =>
      animalSpeciesList.map((s) => (
        <SelectItem key={s.id} value={String(s.id)} disabled={s.isInactive}>
          {s.label || s.name}
        </SelectItem>
      )),
    [animalSpeciesList],
  );

  const insuranceSelectItems = useMemo(
    () =>
      insuranceList.map((ins) => (
        <SelectItem key={ins.id} value={String(ins.id)}>
          {ins.name}{ins.coverage_rate != null ? ` (${ins.coverage_rate}%補償)` : ""}
        </SelectItem>
      )),
    [insuranceList],
  );

  const handleAnimalSpeciesChange = (value: string) => {
    const selected = animalSpeciesList.find((s) => String(s.id) === value);
    setFormData((prev) => ({
      ...prev,
      animalSpeciesId: value,
      species: selected?.name ?? prev.species,
      breed: "",
    }));
    clearFieldError("animalSpeciesId");
  };

  const handleInsuranceChange = (value: string) => {
    const actualValue = value === "none" ? "" : value;
    const selected = insuranceList.find((ins) => String(ins.id) === actualValue);
    setFormData((prev) => ({
      ...prev,
      insuranceId: actualValue,
      insuranceName: selected?.name as PetFormData["insuranceName"],
      insuranceDetails:
        actualValue === ""
          ? undefined
          : selected?.coverage_rate != null
            ? `${selected.coverage_rate}%補償`
            : prev.insuranceDetails,
    }));
  };

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <PetIdentitySection
        formData={formData}
        setFormData={setFormData}
        fieldErrors={fieldErrors}
        clearFieldError={clearFieldError}
        animalSpeciesSelectItems={animalSpeciesSelectItems}
        isLoadingSpecies={isLoadingSpecies}
        isEdit={isEdit}
        onAnimalSpeciesChange={handleAnimalSpeciesChange}
      />
      <PetPhysicalSection
        formData={formData}
        setFormData={setFormData}
        fieldErrors={fieldErrors}
        clearFieldError={clearFieldError}
      />
      <PetCareSection
        formData={formData}
        setFormData={setFormData}
        insuranceSelectItems={insuranceSelectItems}
        isLoadingInsurances={isLoadingInsurances}
        canEdit={canEdit}
        onInsuranceChange={handleInsuranceChange}
      />
    </div>
  );
}
