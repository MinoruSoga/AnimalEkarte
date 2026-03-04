import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import type { MasterItem } from "@/types";
import { transformMasterItem } from "./transforms";
import type { BackendMasterItem } from "./types";

export const getMasterItems = async (): Promise<MasterItem[]> => {
  const { data } = await axios.get<BackendMasterItem[]>("/v1/master-items");
  return data.map(transformMasterItem);
};

export const useGetMasterItems = () => {
  return useQuery({
    queryKey: ["masterItems"],
    queryFn: getMasterItems,
  });
};
