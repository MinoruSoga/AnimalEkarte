import { Navigate } from "react-router";
import { useAuth } from "../hooks/useAuth";
import { LoginForm } from "../components/LoginForm";

/* rendering-hoist-jsx: 静的JSXをコンポーネント外に hoist して再生成を防ぐ */
const LOADING_VIEW = (
  <div className="min-h-screen flex items-center justify-center bg-[#F7F6F3]">
    <div className="size-8 border-2 border-[#038B94] border-t-transparent rounded-full animate-spin" />
  </div>
);

const LOGIN_PAGE = (
  <div className="min-h-screen flex items-center justify-center bg-[#F7F6F3] p-4">
    <LoginForm />
  </div>
);

export function Login() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) return LOADING_VIEW;
  if (isAuthenticated) return <Navigate to="/" replace />;

  return LOGIN_PAGE;
}
