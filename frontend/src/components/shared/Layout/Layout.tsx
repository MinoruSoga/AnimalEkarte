import { Navigate, Outlet, useLocation, useNavigation } from "react-router";
import { Sidebar } from "./Sidebar";
import { useAuth } from "@/hooks/use-auth";
import { C, STYLE, Z_CLASS } from "@/lib/design-tokens";
import { paths } from "@/config/paths";

export function Layout() {
  const { isAuthenticated, isLoading } = useAuth();
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
        <div className={`fixed top-0 left-0 right-0 ${Z_CLASS.overlay} h-[2px] bg-transparent overflow-hidden`}>
          <div className={`h-full ${C.bgActionPrimary} animate-progress-indeterminate origin-left ${STYLE.primaryGlow}`} />
        </div>
      ) : null}

      <Sidebar />
      <main className="flex-1 flex flex-col overflow-y-auto relative">
        <Outlet />
      </main>
    </div>
  );
}
