import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAppSelector } from "@/store/hooks";
import { useAuthContext } from "@/AuthProvider";
import { selectIsAuthenticated } from "@/store/auth-slice";

interface ProtectedRouteProps {
  //if allowed roles is omitted any auth user passes
  allowedRoles?: string[];
  redirectTo?: string;
}

//wraps private routes
//while silent refresh is firing, render nothing
//if theres a role mismatch send to correct page
//if authenticated and roke is ok render child throught outlets
export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({
  allowedRoles,
  redirectTo = "/login",
}) => {
  const { isBootstrapping } = useAuthContext();
  const isAuthenticated = useAppSelector(selectIsAuthenticated);
  const user = useAppSelector((s) => s.auth.user);
  const location = useLocation();

  // Don't render anything while we're resolving the session prevents
  // the flash of the login page on hard refresh for authenticated users.
  if (isBootstrapping) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[#0a0a0a]">
        <div className="w-6 h-6 border-2 border-[#c8f557] border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to={redirectTo} state={{ from: location }} replace />;
  }

  if (allowedRoles && user && !allowedRoles.includes(user.role)) {
    if (user.role === "admin") {
      return <Navigate to="/admin" replace />;
    }
    return <Navigate to="/dashboard" replace />;
  }

  return <Outlet />;
};

//redirects already auth users away from login and register page
export const PublicOnlyRoute: React.FC = () => {
  const { isBootstrapping } = useAuthContext();
  const isAuthenticated = useAppSelector(selectIsAuthenticated);
  const user = useAppSelector((s) => s.auth.user);

  if (isBootstrapping) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[#0a0a0a]">
        <div className="w-6 h-6 border-2 border-[#c8f557] border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (isAuthenticated && user) {
    return (
      <Navigate to={user.role === "admin" ? "/admin" : "/dashboard"} replace />
    );
  }

  return <Outlet />;
};
