import { QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "@/components/ui/sonner";
import { queryClient } from "@/lib/react-query";

interface AppProviderProps {
  children: React.ReactNode;
}

// FE-RC-053: AuthProvider はこの AppProvider ではなく router.tsx の root route element に
// アプリ全体（/login 含む）へ配置されている。BUG-031 以降、/login での GET /v1/me 401 は
// AuthProvider 内部のセッション復元処理が password-recovery 経路（isPasswordRecoveryPublicPath）
// のみ復元をスキップすることで回避しており、配置レベルでの除外ではない。
export function AppProvider({ children }: AppProviderProps) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <Toaster />
    </QueryClientProvider>
  );
}
