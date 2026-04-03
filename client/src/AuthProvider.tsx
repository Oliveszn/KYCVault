import React, {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import axios from "axios";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import {
  clearCredentials,
  selectExpiresAt,
  setCredentials,
  setUser,
} from "./store/auth-slice";
import { authApi } from "./lib/api/auth";

const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080/api/v1";

//the threshold for us to refresh before access expires
const REFRESH_THRESHOLD_MS = 60_000; // 1 min

interface AuthContextValue {
  /** True while the initial silent refresh attempt is in flight. */
  isBootstrapping: boolean;
}

const AuthContext = createContext<AuthContextValue>({ isBootstrapping: true });

export const useAuthContext = () => useContext(AuthContext);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const dispatch = useAppDispatch();
  const expiresAt = useAppSelector(selectExpiresAt);
  const [isBootstrapping, setIsBootstrapping] = useState(true);
  const refreshTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  ////slient refresh helper
  const silentRefresh = async (): Promise<boolean> => {
    try {
      const response = await axios.post<{
        payload: { accessToken: string; expiresIn: number };
      }>(`${BASE_URL}/auth/refresh`, {}, { withCredentials: true });

      const { accessToken, expiresIn } = response.data.payload;
      dispatch(setCredentials({ accessToken, expiresIn }));

      // Also hydrate user profile.
      const user = await authApi.me();
      dispatch(setUser(user));

      return true;
    } catch {
      dispatch(clearCredentials());
      return false;
    }
  };

  // Schedule proactive refresh
  const scheduleRefresh = (expiry: number) => {
    if (refreshTimerRef.current) {
      clearTimeout(refreshTimerRef.current);
    }

    const msUntilRefresh = expiry - Date.now() - REFRESH_THRESHOLD_MS;

    if (msUntilRefresh <= 0) {
      //we call the refresh when token is close to expiry
      silentRefresh().then((ok) => {
        if (ok && expiresAt) scheduleRefresh(expiresAt);
      });
      return;
    }

    refreshTimerRef.current = setTimeout(async () => {
      const ok = await silentRefresh();
      // After a successful refresh, expiresAt in Redux will update and
      // the useEffect below will schedule the next cycle automatically.
      if (!ok) {
        dispatch(clearCredentials());
      }
    }, msUntilRefresh);
  };

  // try to restore session on mount
  useEffect(() => {
    silentRefresh().finally(() => setIsBootstrapping(false));

    return () => {
      if (refreshTimerRef.current) clearTimeout(refreshTimerRef.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Re-schedule whenever expiry changes (login / refresh cycle)
  useEffect(() => {
    if (expiresAt) {
      scheduleRefresh(expiresAt);
    }
    return () => {
      if (refreshTimerRef.current) clearTimeout(refreshTimerRef.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [expiresAt]);

  return (
    <AuthContext.Provider value={{ isBootstrapping }}>
      {children}
    </AuthContext.Provider>
  );
};
