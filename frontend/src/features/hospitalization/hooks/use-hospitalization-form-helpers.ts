import { useEffect, useRef } from "react";
import { useNavigate } from "react-router";
import { toast } from "sonner";
import { paths } from "@/config/paths";
import type { EntityReadResult } from "@/lib/entity-read-result";
import type { Pet, HospitalizationTreatmentPlan } from "@/types";
import type { TreatmentPlanResponse } from "@/types/generated/hospitalization-responses";
import type { BackendHospitalization } from "../api/types";
import type { HospitalizationFormData } from "../types";
import {
  buildHospitalizationFormDataFromRecord,
  buildSelectedPetFromHospitalization,
  buildTreatmentPlansFromRecord,
} from "./use-hospitalization-form-model";

export function useHospitalizationFormHydration(input: {
  isEdit: boolean;
  id: string | undefined;
  petId: string | null;
  isPetLoading: boolean;
  petFromQuery: Pet | undefined;
  entityRead: EntityReadResult<BackendHospitalization>;
  treatmentPlanWire: readonly TreatmentPlanResponse[] | undefined;
  isTreatmentPlansSuccess: boolean;
  selectedPetRef: { current: Pet | undefined };
  setFormData: (updater: (prev: HospitalizationFormData) => HospitalizationFormData) => void;
  setSelectedPets: (pets: Pet[]) => void;
  setTreatmentPlans: (plans: HospitalizationTreatmentPlan[]) => void;
}) {
  const {
    isEdit,
    id,
    petId,
    isPetLoading,
    petFromQuery,
    entityRead,
    treatmentPlanWire,
    isTreatmentPlansSuccess,
    selectedPetRef,
    setFormData,
    setSelectedPets,
    setTreatmentPlans,
  } = input;
  const navigate = useNavigate();
  const hydratedHospitalizationId = useRef<number | undefined>(undefined);

  useEffect(() => {
    if (entityRead.status !== "found") return;
    const record = entityRead.data;
    if (hydratedHospitalizationId.current === record.id) return;
    hydratedHospitalizationId.current = record.id;
    setFormData((prev) => buildHospitalizationFormDataFromRecord(prev, record));
    const selectedPet = buildSelectedPetFromHospitalization(record);
    if (selectedPet) {
      selectedPetRef.current = selectedPet;
      setSelectedPets([selectedPet]);
    }
  }, [entityRead, setSelectedPets, selectedPetRef, setFormData]);

  const hydratedTreatmentPlansForId = useRef<string | undefined>(undefined);
  useEffect(() => {
    if (!isEdit || !id || !isTreatmentPlansSuccess || treatmentPlanWire === undefined) return;
    if (entityRead.status !== "found") return;
    if (hydratedTreatmentPlansForId.current === id) return;
    hydratedTreatmentPlansForId.current = id;
    setTreatmentPlans(buildTreatmentPlansFromRecord(treatmentPlanWire));
  }, [isEdit, id, isTreatmentPlansSuccess, treatmentPlanWire, entityRead.status, setTreatmentPlans]);

  useEffect(() => {
    if (!petId || id) return;
    if (isPetLoading) return;
    if (petFromQuery) {
      selectedPetRef.current = petFromQuery;
      setSelectedPets([petFromQuery]);
    } else {
      toast.error("ペット情報の取得に失敗しました");
      navigate(paths.hospitalization.selectPet.getHref());
    }
  }, [petId, id, petFromQuery, isPetLoading, setSelectedPets, navigate, selectedPetRef]);
}
