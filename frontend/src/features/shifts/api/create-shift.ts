import { axios } from "@/lib/axios";
import type { CreateShiftInput } from "../types";
import { transformShift } from "./transforms";
import type { Shift } from "./transforms";
import type { BackendShift } from "./types";

export async function createShift(input: CreateShiftInput): Promise<Shift> {
  const { data } = await axios.post<BackendShift>("/v1/shifts", {
    ...input,
    staff_id: Number(input.staff_id),
  });
  return transformShift(data);
}
