import { Provider } from "react-redux";
import store from "./store/store";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { AuthProvider } from "./AuthProvider";
import {
  ProtectedRoute,
  PublicOnlyRoute,
} from "./components/common/ProtectedRoutes";
import LoginPage from "./pages/auth/LoginPage";
import RegisterPage from "./pages/auth/RegisterPage";
import DashboardPage from "./pages/dashboard/DashboardPage";
import KycForm from "./pages/KYCWizard/KycForm";
import InitiateSession from "./components/KycForm/Step1";
import UploadDocument from "./components/KycForm/Step2";
import FaceVerification from "./components/KycForm/Step3";
import { Toaster } from "./components/ui/sonner";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error: unknown) => {
        // Never retry 401s the interceptor handles refresh.
        // Never retry 403s user genuinely lacks permission.
        const status = (error as { response?: { status: number } })?.response
          ?.status;
        if (status === 401 || status === 403) return false;
        return failureCount < 2;
      },
      staleTime: 30_000,
    },
  },
});
export default function App() {
  return (
    <Provider store={store}>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <Toaster position="top-right" richColors />
          <AuthProvider>
            <Routes>
              {/* Public routes redirects to /dashboard if already auth */}
              <Route element={<PublicOnlyRoute />}>
                <Route path="/login" element={<LoginPage />} />
                <Route path="/register" element={<RegisterPage />} />
              </Route>

              {/* Protected routes redirect to /login if not authed */}
              <Route element={<ProtectedRoute />}>
                <Route path="/dashboard" element={<DashboardPage />} />
                <Route path="/kyc" element={<KycForm />}>
                  <Route
                    index
                    element={<Navigate to="/kyc/sessions/new" replace />}
                  />

                  <Route path="sessions">
                    <Route path="new" element={<InitiateSession />} />
                    <Route path=":id/documents" element={<UploadDocument />} />
                    <Route path=":id/face" element={<FaceVerification />} />
                  </Route>
                </Route>
              </Route>

              {/* <Route element={<ProtectedRoute allowedRoles={["admin"]} />}>
                <Route path="/admin" element={<AdminPage />} />
              </Route> */}

              <Route path="/" element={<Navigate to="/dashboard" replace />} />
              <Route path="*" element={<Navigate to="/dashboard" replace />} />
            </Routes>
          </AuthProvider>
        </BrowserRouter>
      </QueryClientProvider>
    </Provider>
  );
}
