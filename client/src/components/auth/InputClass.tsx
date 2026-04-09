import { cn } from "@/lib/utils";
import { AlertCircle } from "lucide-react";

export const inputClass = (hasError: boolean) =>
  cn(
    "w-full px-4 py-3 rounded-lg text-sm text-gray-900 placeholder-gray-400",
    "bg-white border transition-all duration-150 outline-none",
    "focus:ring-2 focus:ring-offset-0",
    hasError
      ? "border-red-500/50 focus:border-red-500 focus:ring-red-500/20"
      : "border-gray-300 focus:border-blue-500 focus:ring-[#c8f557]/30",
  );

export function Field({
  label,
  error,
  children,
}: {
  label: string;
  error?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-1.5">
      <label className="block text-xs font-medium text-gray-700 tracking-wide uppercase">
        {label}
      </label>
      {children}
      {error && (
        <p className="flex items-center gap-1.5 text-xs text-red-400 animate-in fade-in slide-in-from-top-1 duration-200">
          <AlertCircle size={11} />
          {error}
        </p>
      )}
    </div>
  );
}
