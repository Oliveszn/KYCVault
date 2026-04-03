import { useEffect } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Eye, EyeOff, LogIn, AlertCircle, CheckCircle2 } from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { useLogin } from "@/hooks/useAuth";
import { LoginFormValues, loginSchema } from "@/utils/validation/authSchema";
import { Field, inputClass } from "@/components/auth/InputClass";

export default function LoginPage() {
  const [showPassword, setShowPassword] = useState(false);
  const [searchParams] = useSearchParams();
  const justRegistered = searchParams.get("registered") === "true";
  const login = useLogin();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    setFocus,
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  });

  useEffect(() => {
    setFocus("email");
  }, [setFocus]);

  const onSubmit = (values: LoginFormValues) => {
    login.mutate(values);
  };

  return (
    <div className="min-h-screen bg-[#080808] flex">
      <div className="hidden lg:flex lg:w-1/2 relative flex-col justify-between p-14 overflow-hidden">
        <div
          className="absolute inset-0"
          style={{
            backgroundImage:
              "linear-gradient(rgba(200,245,87,0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(200,245,87,0.04) 1px, transparent 1px)",
            backgroundSize: "40px 40px",
          }}
        />
        <div className="absolute bottom-0 left-0 w-[600px] h-[600px] bg-[#c8f557] opacity-[0.06] rounded-full blur-[120px] -translate-x-1/3 translate-y-1/3" />

        <div className="relative z-10">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-[#c8f557] rounded-sm flex items-center justify-center">
              <span className="text-black font-black text-sm tracking-tighter">
                A
              </span>
            </div>
            <span className="text-white font-semibold tracking-wide text-sm">
              ACME
            </span>
          </div>
        </div>

        <div className="relative z-10 space-y-6">
          <div className="space-y-2">
            <p className="text-[#c8f557] text-sm font-mono tracking-[0.2em] uppercase">
              Welcome back
            </p>
            <h1 className="text-white font-bold text-5xl leading-[1.1] tracking-tight">
              Pick up right
              <br />
              where you left.
            </h1>
          </div>
          <p className="text-[#555] text-base leading-relaxed max-w-sm">
            Your session is protected with rotating refresh tokens and
            short-lived access tokens.
          </p>

          <div className="flex flex-col gap-2 pt-2">
            {[
              "httpOnly cookie session",
              "Silent token refresh",
              "Auto logout on expiry",
            ].map((f) => (
              <div
                key={f}
                className="flex items-center gap-2.5 text-sm text-[#666]"
              >
                <div className="w-1.5 h-1.5 rounded-full bg-[#c8f557]" />
                {f}
              </div>
            ))}
          </div>
        </div>

        <div className="relative z-10 text-[#333] text-xs">
          © {new Date().getFullYear()} ACME Corp. All rights reserved.
        </div>
      </div>

      <div className="w-full lg:w-1/2 flex items-center justify-center p-6 lg:p-14">
        <div className="w-full max-w-md space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
          <div className="flex items-center gap-3 lg:hidden">
            <div className="w-7 h-7 bg-[#c8f557] rounded-sm flex items-center justify-center">
              <span className="text-black font-black text-xs">A</span>
            </div>
            <span className="text-white font-semibold tracking-wide text-sm">
              ACME
            </span>
          </div>

          <div className="space-y-1.5">
            <h2 className="text-white text-3xl font-bold tracking-tight">
              Sign in
            </h2>
            <p className="text-[#555] text-sm">
              Don't have an account?{" "}
              <Link
                to="/register"
                className="text-[#c8f557] hover:text-white transition-colors font-medium"
              >
                Create one
              </Link>
            </p>
          </div>

          {/* Registered success banner */}
          {justRegistered && (
            <div className="flex items-center gap-3 p-3.5 rounded-lg bg-[#c8f557]/10 border border-[#c8f557]/20 text-[#c8f557] text-sm animate-in fade-in duration-300">
              <CheckCircle2 className="w-4 h-4 shrink-0" />
              Account created! Sign in to continue.
            </div>
          )}

          {/* Server error */}
          {login.isError && (
            <div className="flex items-center gap-3 p-3.5 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm animate-in fade-in duration-300">
              <AlertCircle className="w-4 h-4 shrink-0" />
              {(login.error as Error)?.message ?? "Invalid email or password"}
            </div>
          )}

          <form
            onSubmit={handleSubmit(onSubmit)}
            className="space-y-5"
            noValidate
          >
            {/* Email */}
            <Field label="Email address" error={errors.email?.message}>
              <input
                {...register("email")}
                type="email"
                autoComplete="email"
                placeholder="you@example.com"
                className={inputClass(!!errors.email)}
              />
            </Field>

            {/* Password */}
            <Field label="Password" error={errors.password?.message}>
              <div className="relative">
                <input
                  {...register("password")}
                  type={showPassword ? "text" : "password"}
                  autoComplete="current-password"
                  placeholder="••••••••"
                  className={cn(inputClass(!!errors.password), "pr-11")}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((v) => !v)}
                  className="absolute right-3.5 top-1/2 -translate-y-1/2 text-[#444] hover:text-[#c8f557] transition-colors"
                  tabIndex={-1}
                >
                  {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
            </Field>

            <div className="flex justify-end">
              <Link
                to="/forgot-password"
                className="text-xs text-[#444] hover:text-[#c8f557] transition-colors"
              >
                Forgot password?
              </Link>
            </div>

            <button
              type="submit"
              disabled={isSubmitting || login.isPending}
              className={cn(
                "w-full flex items-center justify-center gap-2.5 px-5 py-3.5 rounded-lg",
                "bg-[#c8f557] text-black text-sm font-semibold tracking-wide",
                "hover:bg-[#d9f87a] active:scale-[0.98] transition-all duration-150",
                "disabled:opacity-50 disabled:cursor-not-allowed disabled:active:scale-100",
                "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#c8f557] focus-visible:ring-offset-2 focus-visible:ring-offset-[#080808]",
              )}
            >
              {login.isPending ? (
                <span className="w-4 h-4 border-2 border-black/30 border-t-black rounded-full animate-spin" />
              ) : (
                <>
                  <LogIn size={15} />
                  Sign in
                </>
              )}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
