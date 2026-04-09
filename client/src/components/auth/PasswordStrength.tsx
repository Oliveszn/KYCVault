import { cn } from "@/lib/utils";
import { Check } from "lucide-react";

const strengthChecks = [
  { label: "8+ characters", test: (p: string) => p.length >= 8 },
  { label: "Uppercase letter", test: (p: string) => /[A-Z]/.test(p) },
  { label: "Number", test: (p: string) => /[0-9]/.test(p) },
  { label: "Special character", test: (p: string) => /[^A-Za-z0-9]/.test(p) },
];

export default function PasswordStrength({ password }: { password: string }) {
  const passed = strengthChecks.filter((c) => c.test(password)).length;
  const colors = ["#333", "#ef4444", "#f97316", "#eab308", "#c8f557"];

  if (!password) return null;

  return (
    <div className="space-y-2.5 animate-in fade-in duration-200">
      <div className="flex gap-1">
        {Array.from({ length: 4 }).map((_, i) => (
          <div
            key={i}
            className="h-0.5 flex-1 rounded-full transition-all duration-300"
            style={{ backgroundColor: i < passed ? colors[passed] : "#222" }}
          />
        ))}
      </div>
      <div className="grid grid-cols-2 gap-x-4 gap-y-1">
        {strengthChecks.map((c) => {
          const ok = c.test(password);
          return (
            <div key={c.label} className="flex items-center gap-1.5 text-xs">
              <Check
                size={10}
                className={cn(
                  "transition-colors duration-200",
                  ok ? "text-[#c8f557]" : "text-[#333]",
                )}
              />
              <span
                className={cn(
                  "transition-colors duration-200",
                  ok ? "text-[#888]" : "text-[#444]",
                )}
              >
                {c.label}
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}
