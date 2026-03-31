import { axios } from "@/lib/axios";

export interface ForgotPasswordInput {
  email: string;
}

export async function forgotPassword(input: ForgotPasswordInput): Promise<void> {
  await axios.post("/v1/auth/forgot-password", input);
}
