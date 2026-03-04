import { axios } from "@/lib/axios";
import type { Owner } from "@/types/owner";
import { transformOwner } from "./transforms";
import type { BackendOwner } from "./types";

export const getOwners = async (): Promise<Owner[]> => {
  const { data } = await axios.get<BackendOwner[]>("/v1/owners");
  return data.map(transformOwner);
};
