import { axios } from "@/lib/axios";
import type { Owner } from "@/types/owner";
import { transformOwner } from "./transforms";
import type { BackendOwner } from "./types";

interface OwnersResponse {
  data: BackendOwner[];
  total: number;
  page: number;
  limit: number;
}

export const getOwners = async (): Promise<Owner[]> => {
  const { data } = await axios.get<OwnersResponse>("/v1/owners");
  return data.data.map(transformOwner);
};
