import { LoginForm } from "../components/LoginForm";

/* rendering-hoist-jsx: 静的JSXをコンポーネント外に hoist して再生成を防ぐ */
// AuthProvider はログイン後の保護ルート側にのみ配置するため、このページでは useAuth() を使わない。
// ログイン済みユーザーが /login にアクセスした場合は LoginForm 内で / へリダイレクトする。
const LOGIN_PAGE = (
  <div className="min-h-screen flex items-center justify-center bg-[#F1F0EE] p-4">
    <LoginForm />
  </div>
);

export function Login() {
  return LOGIN_PAGE;
}
