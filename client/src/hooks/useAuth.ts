import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { authApi } from "@/lib/api/auth";
import { useAppDispatch } from "@/store/hooks";
import type { LoginPayload, RegisterPayload } from "@/types/auth";
import { clearCredentials, setCredentials, setUser } from "@/store/auth-slice";

export const authKeys = {
  me: ["auth", "me"] as const,
};

export const useRegister = () => {
  const navigate = useNavigate();

  return useMutation({
    mutationFn: (payload: RegisterPayload) => authApi.register(payload),
    onSuccess: () => {
      navigate("/login?registered=true");
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
      dispatch(
        setCredentials({
          accessToken: data.accessToken,
          expiresIn: data.expiresIn,
        }),
      );

      ///fetch user profile and cache
      const user = await authApi.me();
      dispatch(setUser(user));
      queryClient.setQueryData(authKeys.me, user);

      navigate("/dashboard");
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
