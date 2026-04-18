import { Navigate, Outlet, useLocation, useNavigation } from "react-router";
import { Sidebar } from "./Sidebar";
import { useAuth } from "@/hooks/use-auth";
import { C } from "@/lib/design-tokens";
import { paths } from "@/config/paths";

export function Layout() {
  const { isAuthenticated, isLoading, isSwitchingClinic } = useAuth();
  const location = useLocation();
  const navigation = useNavigation();

  if (isLoading) return null;

  if (!isAuthenticated) {
    // BUG-047: Redirect back to original destination after login
    return <Navigate to={paths.auth.login.getHref()} replace state={{ from: location.pathname }} />;
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
      <main className="flex-1 flex flex-col overflow-y-auto relative">
        <Outlet />
        {isSwitchingClinic ? (
          <div className="absolute inset-0 z-50 flex items-center justify-center bg-white/70 backdrop-blur-[1px]">
            <div className="flex flex-col items-center gap-3">
              <div className={`size-8 border-2 ${C.borderBrand} border-t-transparent rounded-full animate-spin`} />
              <span className={`text-sm ${C.text50}`}>医院を切り替えています...</span>
            </div>
          </div>
        ) : null}
      </main>
    </div>
  );
}
