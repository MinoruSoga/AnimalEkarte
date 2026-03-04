import { axios } from "@/lib/axios";

export const deleteOwner = async (id: string): Promise<void> => {
  await axios.delete(`/v1/owners/${id}`);
};
