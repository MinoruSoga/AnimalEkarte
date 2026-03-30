import { Navigate, Outlet, useLocation } from "react-router";
import { Sidebar } from "./Sidebar";
import { useAuth } from "@/features/auth/hooks/use-auth";

export function Layout() {
  const { isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) return null;

  if (!isAuthenticated) {
    // BUG-047: Redirect back to original destination after login
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
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
