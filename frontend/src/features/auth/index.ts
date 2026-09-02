// AuthProvider は @/features/auth/provider からのみ export する（router.tsx が唯一の配置点）。
export { useAuth } from "@/hooks/use-auth";
export { usePermission } from "@/hooks/use-permission";
export { ME_QUERY_KEY } from "@/lib/query-keys";
export { Login } from "./routes/Login";
// 実装は共有レイヤ（@/components/shared/ChangePasswordDialog/）。
// auth ドメインの公開 API として feature index から再エクスポートする。
export { ChangePasswordDialog } from "@/components/shared/ChangePasswordDialog/ChangePasswordDialog";
export { ForgotPasswordPage } from "./routes/ForgotPasswordPage";
export { ResetPasswordPage } from "./routes/ResetPasswordPage";
export type { ResourceAction } from "./types";
