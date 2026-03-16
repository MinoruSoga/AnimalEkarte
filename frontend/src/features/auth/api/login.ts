import { axios } from "@/lib/axios";
import type { AuthUser } from "../types";
import { type BackendMeResponse, mapMeToAuthUser } from "./transforms";

export interface LoginResponse {
  user: AuthUser;
}

/** バックエンドのログインレスポンス。token はレスポンスボディに含まれる */
interface BackendLoginResponse {
  token: string;
  expires_at: number;
  user_type: string;
  user: BackendMeResponse;
}

export async function login(
  email: string,
  password: string,
): Promise<LoginResponse> {
  // async-parallel 適用済み: バックエンドがログインレスポンスにユーザー情報を含むため
  // /me への2番目のリクエストは不要（ウォーターフォール解消）
  const { data } = await axios.post<BackendLoginResponse>(
    "/v1/login",
    { email, password },
  );

  // JWT トークンを sessionStorage に保存（Authorization Bearer ヘッダで送信用）
  sessionStorage.setItem("auth_token", data.token);

  const user = mapMeToAuthUser(data.user);

  return { user };
}
