import { axios } from "@/lib/axios";
import { transformHospitalization } from "./transforms";
import type { Hospitalization } from "./transforms";
import type { BackendHospitalization, CreateHospitalizationRequest } from "./types";

function toOptionalUint(value: string | undefined): number | undefined {
  if (!value) return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : undefined;
}

export const createHospitalization = async (
  req: CreateHospitalizationRequest,
): Promise<Hospitalization> => {
  const { data } = await axios.post<BackendHospitalization>("/v1/hospitalizations", {
    ...req,
    pet_id: Number(req.pet_id),
    owner_id: Number(req.owner_id),
    cage_id: req.cage_id ? Number(req.cage_id) : undefined,
    doctor_id: toOptionalUint(req.doctor_id),
  });
  return transformHospitalization(data);
};
