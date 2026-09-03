import { axios } from "@/lib/axios";
import type { AuthUser } from "../types";
import { type BackendMeResponse, mapMeToAuthUser } from "./transforms";

export interface LoginResponse {
  user: AuthUser;
}

/** バックエンドのログインレスポンス。JWT は httpOnly Cookie で設定されるためボディに含まれない */
interface BackendLoginResponse {
  is_system_admin: boolean;
  user: BackendMeResponse;
}

export async function login(email: string, password: string): Promise<LoginResponse> {
  // async-parallel 適用済み: バックエンドがログインレスポンスにユーザー情報を含むため
  // /me への2番目のリクエストは不要（ウォーターフォール解消）
  const { data } = await axios.post<BackendLoginResponse>("/v1/login", { email, password });

  // JWT トークンはバックエンドが httpOnly Cookie で設定するため
  // フロントエンド側での sessionStorage 保存は不要

  const user = mapMeToAuthUser(data.user);

  return { user };
}
