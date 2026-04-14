import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Eye, EyeOff, UserPlus, AlertCircle, Check } from "lucide-react";
import { cn } from "@/lib/utils";
import { useRegister } from "@/hooks/useAuth";
import {
  RegisterFormValues,
  registerSchema,
} from "@/utils/validation/authSchema";
import { Field, inputClass } from "@/components/auth/InputClass";
import PasswordStrength from "@/components/auth/PasswordStrength";

export default function RegisterPage() {
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const register_ = useRegister();
  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      firstName: "",
      lastName: "",
      email: "",
      password: "",
      confirmPassword: "",
    },
    mode: "onTouched",
  });

  const watchedPassword = watch("password", "");

  const onSubmit = (values: RegisterFormValues) => {
    register_.mutate({
      email: values.email,
      password: values.password,
      confirmPassword: values.confirmPassword,
      firstName: values.firstName,
      lastName: values.lastName,
    });
  };

  return (
    <div className="min-h-screen flex">
      <div className="hidden lg:flex lg:w-1/2 relative flex-col justify-between p-14 overflow-hidden">
        <div
          className="absolute inset-0 pointer-events-none"
          style={{
            backgroundImage:
              "linear-gradient(rgba(200,245,87,0.1) 1px, transparent 1px), linear-gradient(90deg, rgba(200,245,87,0.1) 1px, transparent 1px)",
            backgroundSize: "40px 40px",
          }}
        />
        <div className="absolute bottom-0 left-0 w-[600px] h-[600px] bg-blue-500 opacity-20 rounded-full blur-[140px] -translate-x-1/3 translate-y-1/3 pointer-events-none" />

        <div className="relative z-10 flex items-center gap-3">
          <div className="w-8 h-8 bg-blue-500 rounded-sm flex items-center justify-center">
            K
          </div>
          <span className="text-black font-semibold tracking-wide text-sm">
            KYCVault
          </span>
        </div>

        <div className="relative z-10 space-y-6">
          <p className="text-black text-sm font-mono tracking-[0.2em] uppercase">
            Get started
          </p>
          <h1 className="text-black font-bold text-5xl leading-[1.1] tracking-tight">
            Verify your identity
            <br />
            in minutes.
          </h1>
          <p className="text-black text-base leading-relaxed max-w-sm">
            Securely complete your KYC process with real-time identity checks,
            document verification, and facial recognition.
          </p>
        </div>

        <div className="relative z-10 text-[#333] text-xs">
          © {new Date().getFullYear()} KYCVault. All rights reserved.
        </div>
      </div>

      <div className="w-full lg:w-1/2 flex items-center justify-center p-6 lg:p-14 overflow-y-auto">
        <div className="w-full max-w-md space-y-7 py-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
          <div className="flex items-center gap-3 lg:hidden">
            <div className="w-7 h-7 bg-blue-500 rounded-sm flex items-center justify-center">
              <span className="text-black font-black text-xs">KV</span>
            </div>
            <span className="text-black font-semibold tracking-wide text-sm">
              KYCVault
            </span>
          </div>

          <div className="space-y-1.5">
            <h2 className="text-black text-3xl font-bold tracking-tight">
              Create account
            </h2>
            <p className="text-[#555] text-sm">
              Already have one?{" "}
              <Link
                to="/login"
                className="text-blue-500 hover:text-white transition-colors font-medium"
              >
                Sign in
              </Link>
            </p>
          </div>

          {/* {register_.isError && (
            <div className="flex items-center gap-3 p-3.5 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm animate-in fade-in duration-300">
              <AlertCircle className="w-4 h-4 shrink-0" />
              {(register_.error as Error)?.message ??
                "Something went wrong. Please try again."}
            </div>
          )} */}

          <form
            onSubmit={handleSubmit(onSubmit)}
            className="space-y-5"
            noValidate
          >
            {/* Name row */}
            <div className="grid grid-cols-2 gap-3">
              <Field label="First name" error={errors.firstName?.message}>
                <input
                  {...register("firstName")}
                  type="text"
                  autoComplete="given-name"
                  placeholder="Jane"
                  className={inputClass(!!errors.firstName)}
                />
              </Field>
              <Field label="Last name" error={errors.lastName?.message}>
                <input
                  {...register("lastName")}
                  type="text"
                  autoComplete="family-name"
                  placeholder="Smith"
                  className={inputClass(!!errors.lastName)}
                />
              </Field>
            </div>

            {/* Email */}
            <Field label="Email address" error={errors.email?.message}>
              <input
                {...register("email")}
                type="email"
                autoComplete="email"
                placeholder="you@example.com"
                className={inputClass(!!errors.email)}
                autoFocus
              />
            </Field>

            {/* Password */}
            <Field label="Password" error={errors.password?.message}>
              <div className="space-y-3">
                <div className="relative">
                  <input
                    {...register("password")}
                    type={showPassword ? "text" : "password"}
                    autoComplete="new-password"
                    placeholder="••••••••"
                    className={cn(inputClass(!!errors.password), "pr-11")}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword((v) => !v)}
                    className="absolute right-3.5 top-1/2 -translate-y-1/2 text-[#444] hover:text-blue-500 transition-colors"
                    tabIndex={-1}
                  >
                    {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                  </button>
                </div>
                <PasswordStrength password={watchedPassword} />
              </div>
            </Field>

            {/* Confirm password */}
            <Field
              label="Confirm password"
              error={errors.confirmPassword?.message}
            >
              <div className="relative">
                <input
                  {...register("confirmPassword")}
                  type={showConfirm ? "text" : "password"}
                  autoComplete="new-password"
                  placeholder="••••••••"
                  className={cn(inputClass(!!errors.confirmPassword), "pr-11")}
                />
                <button
                  type="button"
                  onClick={() => setShowConfirm((v) => !v)}
                  className="absolute right-3.5 top-1/2 -translate-y-1/2 text-[#444] hover:text-blue-500 transition-colors"
                  tabIndex={-1}
                >
                  {showConfirm ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
            </Field>

            <p className="text-[#333] text-xs leading-relaxed">
              By creating an account you agree to our{" "}
              <span className="text-[#555] hover:text-blue-500 cursor-pointer transition-colors">
                Terms of Service
              </span>{" "}
              and{" "}
              <span className="text-[#555] hover:text-blue-500 cursor-pointer transition-colors">
                Privacy Policy
              </span>
              .
            </p>

            <button
              type="submit"
              disabled={register_.isPending}
              className={cn(
                "w-full flex items-center justify-center gap-2.5 px-5 py-3.5 rounded-lg",
                "bg-blue-500 text-white text-sm font-semibold tracking-wide cursor-pointer",
                "hover:bg-blue-400 active:scale-[0.98] transition-all duration-150",
                "disabled:opacity-50 disabled:cursor-not-allowed disabled:active:scale-100",
                "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2 focus-visible:ring-offset-[#080808]",
              )}
            >
              {register_.isPending ? (
                <span className="w-4 h-4 border-2 border-black/30 border-t-black rounded-full animate-spin" />
              ) : (
                <>
                  <UserPlus size={15} />
                  Create account
                </>
              )}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
