import { axios } from "@/lib/axios";
import type { Owner } from "@/types/owner";
import { transformOwner } from "./transforms";
import type { Owner as BackendOwner } from "@/types/generated/models";

export const getOwner = async (id: string): Promise<Owner> => {
  const { data } = await axios.get<BackendOwner>(`/v1/owners/${id}`);
  return transformOwner(data);
};
