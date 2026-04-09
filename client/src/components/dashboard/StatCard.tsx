import { cn } from "@/lib/utils";

export default function StatCard({
  icon,
  label,
  value,
  highlight,
  truncate,
}: {
  icon: React.ReactNode;
  label: string;
  value: string;
  highlight?: boolean;
  truncate?: boolean;
}) {
  return (
    <div className="rounded-xl border border-gray-200 bg-white p-5 space-y-3 shadow-sm">
      <div className="flex items-center gap-2 text-gray-400 text-xs uppercase tracking-widest">
        {icon}
        {label}
      </div>
      <p
        className={cn(
          "text-lg font-semibold tracking-tight capitalize",
          truncate && "truncate",
          highlight ? "text-[#c8f557]" : "text-gray-900",
        )}
      >
        {value}
      </p>
    </div>
  );
}
