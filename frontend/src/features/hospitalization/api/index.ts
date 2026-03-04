export {
  getHospitalizations,
  useGetHospitalizations,
} from "./get-hospitalizations";
export {
  getHospitalization,
  useGetHospitalization,
  getHospitalizationsByPetId,
  useGetHospitalizationsByPetId,
  getHospitalizationsByOwnerId,
  useGetHospitalizationsByOwnerId,
  getHospitalizationsByStatus,
  useGetHospitalizationsByStatus,
} from "./get-hospitalization";
export {
  createHospitalization,
  useCreateHospitalization,
} from "./create-hospitalization";
export {
  updateHospitalization,
  useUpdateHospitalization,
} from "./update-hospitalization";
export {
  deleteHospitalization,
  useDeleteHospitalization,
} from "./delete-hospitalization";
export type {
  BackendHospitalization,
  CreateHospitalizationRequest,
  UpdateHospitalizationRequest,
} from "./types";
