import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { authApi } from "@/lib/api/auth";
import { useAppDispatch } from "@/store/hooks";
import type { LoginPayload, RegisterPayload } from "@/types/auth";
import { clearCredentials, setCredentials, setUser } from "@/store/auth-slice";
import { toast } from "sonner";
import { ApiError } from "@/types/kyc";
import { AxiosError } from "axios";

export const authKeys = {
  me: ["auth", "me"] as const,
};

export const useRegister = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: RegisterPayload) => authApi.register(payload),
    onSuccess: async (data) => {
      const { payload, message } = data;

      toast.success(message);

      dispatch(
        setCredentials({
          accessToken: payload.accessToken,
          expiresIn: payload.expiresIn,
        }),
      );

      const user = await authApi.me();
      dispatch(setUser(user));
      queryClient.setQueryData(authKeys.me, user);

      navigate("/dashboard");
    },
    onError: (err: AxiosError<ApiError>) => {
      const message = err.response?.data?.message;

      toast.error(message);
    },
  });
};

export const useLogin = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (payload: LoginPayload) => authApi.login(payload),
    onSuccess: async (data) => {
      const { payload, message } = data;

      toast.success(message);

      dispatch(
        setCredentials({
          accessToken: payload.accessToken,
          expiresIn: payload.expiresIn,
        }),
      );

      const user = await authApi.me();
      dispatch(setUser(user));
      queryClient.setQueryData(authKeys.me, user);

      navigate("/dashboard");
    },
    onError: (err: AxiosError<ApiError>) => {
      const message = err.response?.data?.message;

      toast.error(message);
    },
  });
};

export const useLogout = () => {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => authApi.logout(),
    onSettled: () => {
      // Always clear local state, even if the server call fails.
      dispatch(clearCredentials());
      queryClient.clear();
      navigate("/login");
    },
  });
};

export const useMe = (enabled = true) =>
  useQuery({
    queryKey: authKeys.me,
    queryFn: authApi.me,
    enabled,
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
