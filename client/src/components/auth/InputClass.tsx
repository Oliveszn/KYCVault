import { cn } from "@/lib/utils";
import { AlertCircle } from "lucide-react";

export const inputClass = (hasError: boolean) =>
  cn(
    "w-full px-4 py-3 rounded-lg text-sm text-white placeholder-[#383838]",
    "bg-[#111] border transition-all duration-150 outline-none",
    "focus:ring-2 focus:ring-offset-0",
    hasError
      ? "border-red-500/50 focus:border-red-500 focus:ring-red-500/20"
      : "border-[#222] focus:border-[#c8f557] focus:ring-[#c8f557]/15",
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
      <label className="block text-xs font-medium text-[#888] tracking-wide uppercase">
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
