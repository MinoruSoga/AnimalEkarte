import { axios } from "@/lib/axios";
import type { AuthUser } from "../types";
import { type BackendMeResponse, mapMeToAuthUser } from "./transforms";

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user: AuthUser;
}

interface BackendLoginResponse {
  token: string;
  expires_at: number;
  user_type: string;
}

export async function login(
  email: string,
  password: string,
): Promise<LoginResponse> {
  const { data: loginData } = await axios.post<BackendLoginResponse>(
    "/v1/auth/login",
    { email, password },
  );

  // JWT をストレージに保存（axios インターセプターが自動で Authorization ヘッダに注入）
  localStorage.setItem("token", loginData.token);

  const { data: meData } = await axios.get<BackendMeResponse>("/v1/me");
  const user = mapMeToAuthUser(meData);

  return {
    accessToken: loginData.token,
    refreshToken: "",
    user,
  };
}
