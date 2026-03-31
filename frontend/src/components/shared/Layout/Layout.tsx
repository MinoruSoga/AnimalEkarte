import { Navigate, Outlet, useLocation, useNavigation } from "react-router";
import { Sidebar } from "./Sidebar";
import { useAuth } from "@/features/auth/hooks/use-auth";
import { C } from "@/lib/design-tokens";

export function Layout() {
  const { isAuthenticated, isLoading } = useAuth();
  const location = useLocation();
  const navigation = useNavigation();

  if (isLoading) return null;

  if (!isAuthenticated) {
    // BUG-047: Redirect back to original destination after login
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  const isNavigating = navigation.state === "loading";

  return (
    <div className={`flex h-full ${C.bgSubtle} overflow-hidden relative`}>
      {/* Global Navigation Progress Bar */}
      {isNavigating ? (
        <div className="fixed top-0 left-0 right-0 z-[9999] h-[2px] bg-transparent overflow-hidden">
          <div className={`h-full ${C.bgBrand} animate-progress-indeterminate origin-left shadow-[0_0_8px_rgba(3,139,148,0.5)]`} />
        </div>
      ) : null}

      <Sidebar />
      <main className="flex-1 flex flex-col overflow-y-auto">
        <Outlet />
      </main>
    </div>
  );
}
