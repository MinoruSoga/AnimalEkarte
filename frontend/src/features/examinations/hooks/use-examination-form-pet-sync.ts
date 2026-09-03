import { useEffect, useRef } from "react";
import { useNavigate } from "react-router";
import { paths } from "@/config/paths";
import type { Pet } from "@/types";

export function useExaminationFormPetSync(input: {
  isEdit: boolean;
  petId: string | null;
  mutationPet: Pet | undefined;
  isPetLoading: boolean;
  setSelectedPets: (pets: Pet[]) => void;
}) {
  const { isEdit, petId, mutationPet, isPetLoading, setSelectedPets } = input;
  const navigate = useNavigate();
  const initializedPetIDRef = useRef<string | null>(null);
  useEffect(() => {
    if (mutationPet && initializedPetIDRef.current !== mutationPet.id) {
      initializedPetIDRef.current = mutationPet.id;
      setSelectedPets([mutationPet]);
      return;
    }
    if (!isEdit && !petId && !isPetLoading) {
      navigate(paths.examinations.selectPet.getHref());
    }
  }, [isEdit, petId, mutationPet, isPetLoading, setSelectedPets, navigate]);
}
