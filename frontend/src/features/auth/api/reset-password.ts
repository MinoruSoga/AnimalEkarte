import { axios } from "@/lib/axios";

export interface ResetPasswordInput {
  token: string;
  password: string;
}

export async function resetPassword(input: ResetPasswordInput): Promise<void> {
  await axios.post("/v1/auth/reset-password", input);
}
