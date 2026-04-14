import axios, {
  AxiosError,
  AxiosInstance,
  InternalAxiosRequestConfig,
} from "axios";
import store from "@/store/store";
import { clearCredentials, setCredentials } from "@/store/auth-slice";

const BASE_URL = import.meta.env.VITE_API_URL ?? "http://localhost:5000/api/v1";

// Tracks a single in-flight refresh so concurrent 401s don't trigger N refreshes.
let refreshPromise: Promise<string> | null = null;

export const apiClient: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  withCredentials: true,
  headers: { "Content-Type": "application/json" },
});

// Request interceptor, this attaches access token
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = store.getState().auth.accessToken;
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error),
);

// Response interceptor this handle 401 with silent refresh
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean;
    };

    const is401 = error.response?.status === 401;
    const isRefreshEndpoint = originalRequest.url?.includes("/auth/refresh");
    const alreadyRetried = originalRequest._retry;

    const isAuthRoute =
      originalRequest.url?.includes("/login") ||
      originalRequest.url?.includes("/refresh");

    if (!is401 || isAuthRoute || alreadyRetried) {
      return Promise.reject(error);
    }

    originalRequest._retry = true;

    try {
      // Deduplicate: if a refresh is already happening, wait for it.
      if (!refreshPromise) {
        refreshPromise = performRefresh();
      }

      const newAccessToken = await refreshPromise;
      refreshPromise = null;

      if (originalRequest.headers) {
        originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
      }

      return apiClient(originalRequest);
    } catch (refreshError) {
      refreshPromise = null;
      // Refresh failed session is dead, logout from store and send to login.
      store.dispatch(clearCredentials());
      // window.location.href = "/login";
      return Promise.reject(refreshError);
    }
  },
);

async function performRefresh(): Promise<string> {
  // POST /auth/refresh browser automatically sends the httpOnly cookie.
  const response = await axios.post<{
    payload: { accessToken: string; expiresIn: number };
  }>(`${BASE_URL}/auth/refresh`, {}, { withCredentials: true });

  const { accessToken, expiresIn } = response.data.payload;

  store.dispatch(setCredentials({ accessToken, expiresIn }));

  return accessToken;
}
