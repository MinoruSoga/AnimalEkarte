import { QueryClientProvider } from "@tanstack/react-query";
import { Toaster } from "@/components/ui/sonner";
import { queryClient } from "@/lib/react-query";

interface AppProviderProps {
  children: React.ReactNode;
}

// AuthProvider はログインページを含まない保護ルート側にのみ配置する（router.tsx 参照）。
// ここに置くとアプリ起動時に /login でも GET /v1/me が実行され 401 が発生するため除外。
export function AppProvider({ children }: AppProviderProps) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <Toaster />
    </QueryClientProvider>
  );
}
