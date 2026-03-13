import { Navigate, Outlet } from "react-router";
import { Sidebar } from "./Sidebar";
import { useAuth } from "@/features/auth/hooks/useAuth";

export function Layout() {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) return null;

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return (
    <div className="flex h-full bg-[#fafafa] overflow-hidden">
      <Sidebar />
      <main className="flex-1 flex flex-col overflow-y-auto">
        <Outlet />
      </main>
    </div>
  );
}
