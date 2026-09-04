import { axios } from "@/lib/axios";
import type { UpdateShiftInput } from "../types";
import { transformShift } from "./transforms";
import type { Shift } from "./transforms";
import type { BackendShift } from "./types";

export async function updateShift(id: string, input: UpdateShiftInput): Promise<Shift> {
  const { data } = await axios.patch<BackendShift>(`/v1/shifts/${id}`, input);
  return transformShift(data);
}
