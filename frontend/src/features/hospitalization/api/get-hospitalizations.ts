import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Hospitalization } from "@/types";
import { transformHospitalization } from "./transforms";
import type { BackendHospitalization } from "./types";

export const getHospitalizations = async (): Promise<Hospitalization[]> => {
  const { data } = await axios.get<BackendHospitalization[]>(
    "/v1/hospitalizations"
  );
  return data.map(transformHospitalization);
};

export const useGetHospitalizations = () => {
  return useQuery({
    queryKey: ["hospitalizations"],
    queryFn: getHospitalizations,
  });
};
