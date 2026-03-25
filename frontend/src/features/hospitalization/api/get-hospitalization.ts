import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { Hospitalization } from "@/types";
import { transformHospitalization } from "./transforms";
import type { BackendHospitalization } from "./types";

export const getHospitalization = async (
  id: string
): Promise<Hospitalization> => {
  const { data } = await axios.get<BackendHospitalization>(
    `/v1/hospitalizations/${id}`
  );
  return transformHospitalization(data);
};

export const useGetHospitalization = (id: string) => {
  return useQuery({
    queryKey: ["hospitalization", id],
    queryFn: () => getHospitalization(id),
    enabled: !!id,
  });
};

export const getHospitalizationRaw = async (
  id: string
): Promise<BackendHospitalization> => {
  const { data } = await axios.get<BackendHospitalization>(
    `/v1/hospitalizations/${id}`
  );
  return data;
};

export const useGetHospitalizationRaw = (id: string | undefined) => {
  return useQuery({
    queryKey: ["hospitalization", "raw", id],
    queryFn: () => getHospitalizationRaw(id!),
    enabled: !!id,
  });
};

