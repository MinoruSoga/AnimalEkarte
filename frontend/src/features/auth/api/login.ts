import { axios } from "@/lib/axios";
import type { AuthUser } from "../types";
import { type BackendMeResponse, mapMeToAuthUser } from "./transforms";

export interface LoginResponse {
  user: AuthUser;
}

/** バックエンドのログインレスポンス。token は httpOnly Cookie で返るためボディには含まれない */
interface BackendLoginResponse {
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
  // token は httpOnly Cookie でセットされるため localStorage への保存も不要（XSS耐性）
  const { data } = await axios.post<BackendLoginResponse>(
    "/v1/login",
    { email, password },
  );

  const user = mapMeToAuthUser(data.user);

  return { user };
}
