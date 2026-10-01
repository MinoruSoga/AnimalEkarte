import { LoginForm } from "../components/LoginForm";
import { C } from "@/lib/design-tokens";

/* rendering-hoist-jsx: 静的JSXをコンポーネント外に hoist して再生成を防ぐ */
// LoginForm が useAuth() を使用し、ログイン済みなら <Navigate to="/" /> を返す。
const LOGIN_PAGE = (
  <div
    className={`h-dvh overflow-y-auto flex items-center-safe justify-center-safe ${C.bgPage} p-4`}
  >
    <LoginForm />
  </div>
);

export function Login() {
  return LOGIN_PAGE;
}
