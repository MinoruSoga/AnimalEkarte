import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Hospitalization } from "@/types";
import { transformHospitalization } from "./transforms";
import type { BackendHospitalization } from "./types";

interface HospitalizationPaginatedResponse {
  data: BackendHospitalization[];
  total: number;
  page: number;
  limit: number;
}

export const getHospitalizations = async (): Promise<Hospitalization[]> => {
  const { data } = await axios.get<HospitalizationPaginatedResponse>(
    "/v1/hospitalizations"
  );
  return data.data.map(transformHospitalization);
};

export const useGetHospitalizations = () => {
  return useQuery({
    queryKey: ["hospitalizations"],
    queryFn: getHospitalizations,
  });
};
