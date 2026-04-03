import { useState } from "react";
import { formatDistanceToNow, format } from "date-fns";
import {
  LogOut,
  LogOutIcon,
  User,
  Shield,
  Clock,
  Activity,
  ChevronRight,
  Zap,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAppSelector } from "@/store/hooks";
import { useLogout } from "@/hooks/useAuth";
import { selectExpiresAt, selectUser } from "@/store/auth-slice";

export default function DashboardPage() {
  const user = useAppSelector(selectUser);
  const expiresAt = useAppSelector(selectExpiresAt);
  const logout = useLogout();
  const [loggingOutAll, setLoggingOutAll] = useState(false);

  const expiresIn = expiresAt
    ? formatDistanceToNow(new Date(expiresAt), { addSuffix: true })
    : "Unknown";

  const expiresPercent = expiresAt
    ? Math.max(
        0,
        Math.min(100, ((expiresAt - Date.now()) / (15 * 60 * 1000)) * 100),
      )
    : 0;

  return (
    <div className="min-h-screen bg-[#080808] text-white">
      <nav className="border-b border-[#161616] px-6 py-4">
        <div className="max-w-5xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-7 h-7 bg-[#c8f557] rounded-sm flex items-center justify-center">
              <span className="text-black font-black text-xs">A</span>
            </div>
            <span className="text-white font-semibold tracking-wide text-sm">
              ACME
            </span>
          </div>

          <div className="flex items-center gap-3">
            <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full bg-[#c8f557]/10 border border-[#c8f557]/20">
              <div className="w-1.5 h-1.5 rounded-full bg-[#c8f557] animate-pulse" />
              <span className="text-[#c8f557] text-xs font-medium">
                Session active
              </span>
            </div>

            <button
              onClick={() => logout.mutate()}
              disabled={logout.isPending}
              className={cn(
                "flex items-center gap-2 px-3.5 py-2 rounded-lg text-sm",
                "text-[#555] hover:text-white border border-[#1a1a1a] hover:border-[#2a2a2a]",
                "transition-all duration-150 disabled:opacity-50",
              )}
            >
              {logout.isPending ? (
                <span className="w-3.5 h-3.5 border border-[#555] border-t-white rounded-full animate-spin" />
              ) : (
                <LogOut size={14} />
              )}
              Sign out
            </button>
          </div>
        </div>
      </nav>

      <main className="max-w-5xl mx-auto px-6 py-12 space-y-10">
        <div className="space-y-1 animate-in fade-in slide-in-from-bottom-3 duration-500">
          <p className="text-[#c8f557] text-sm font-mono tracking-[0.2em] uppercase">
            Dashboard
          </p>
          <h1 className="text-4xl font-bold tracking-tight">
            Welcome back
            {user?.firstName ? (
              <span className="text-[#c8f557]">, {user.firstName}</span>
            ) : null}
            .
          </h1>
          <p className="text-[#555] text-sm pt-1">
            You're authenticated. Your session renews automatically.
          </p>
        </div>

        <div
          className="grid grid-cols-1 sm:grid-cols-3 gap-4 animate-in fade-in slide-in-from-bottom-3 duration-500"
          style={{ animationDelay: "100ms" }}
        >
          <StatCard
            icon={<User size={16} />}
            label="Role"
            value={user?.role ?? "—"}
            highlight={user?.role === "admin"}
          />
          <StatCard
            icon={<Shield size={16} />}
            label="User ID"
            value={user?.id ? `#${user.id}` : "—"}
          />
          <StatCard
            icon={<Activity size={16} />}
            label="Account"
            value={user?.email ?? "—"}
            truncate
          />
        </div>

        <div
          className="rounded-xl border border-[#161616] bg-[#0d0d0d] p-6 space-y-5 animate-in fade-in slide-in-from-bottom-3 duration-500"
          style={{ animationDelay: "200ms" }}
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <Zap size={15} className="text-[#c8f557]" />
              <span className="text-sm font-medium text-white">
                Access Token
              </span>
            </div>
            <span className="text-xs text-[#444] font-mono">
              HS256 · Bearer
            </span>
          </div>

          <div className="space-y-2">
            <div className="flex items-center justify-between text-xs">
              <div className="flex items-center gap-1.5 text-[#555]">
                <Clock size={11} />
                Expires {expiresIn}
              </div>
              <span className="text-[#333]">
                {expiresAt ? format(new Date(expiresAt), "HH:mm:ss") : ""}
              </span>
            </div>
            <div className="h-1 bg-[#161616] rounded-full overflow-hidden">
              <div
                className="h-full rounded-full transition-all duration-1000"
                style={{
                  width: `${expiresPercent}%`,
                  backgroundColor:
                    expiresPercent > 50
                      ? "#c8f557"
                      : expiresPercent > 20
                        ? "#eab308"
                        : "#ef4444",
                }}
              />
            </div>
            <p className="text-[#333] text-xs">
              Will silently refresh 1 minute before expiry — no action needed.
            </p>
          </div>
        </div>

        <div
          className="rounded-xl border border-[#161616] bg-[#0d0d0d] divide-y divide-[#161616] animate-in fade-in slide-in-from-bottom-3 duration-500"
          style={{ animationDelay: "300ms" }}
        >
          <div className="px-6 py-4">
            <h3 className="text-sm font-medium text-white">
              Session Management
            </h3>
            <p className="text-xs text-[#444] mt-0.5">
              Control your active sessions across all devices
            </p>
          </div>

          <ActionRow
            icon={<LogOut size={14} />}
            title="Sign out this device"
            description="Revokes the current session's refresh token"
            onClick={() => logout.mutate()}
            loading={logout.isPending}
          />

          <ActionRow
            icon={<LogOutIcon size={14} />}
            title="Sign out all devices"
            description="Revokes every active refresh token for your account"
            onClick={() => {
              setLoggingOutAll(true);
              logout.mutate();
            }}
            loading={loggingOutAll && logout.isPending}
            danger
          />
        </div>

        <div
          className="rounded-xl border border-[#c8f557]/10 bg-[#c8f557]/[0.03] p-5 animate-in fade-in slide-in-from-bottom-3 duration-500"
          style={{ animationDelay: "400ms" }}
        >
          <p className="text-xs text-[#555] leading-relaxed">
            <span className="text-[#c8f557] font-medium">
              How your session works:{" "}
            </span>
            Your access token lives in Redux memory (not localStorage) and
            expires in 15 minutes. A refresh token in an{" "}
            <span className="text-[#888]">httpOnly cookie</span> is used to
            silently issue a new one — you'll never be interrupted. On logout,
            the token is revoked server-side.
          </p>
        </div>
      </main>
    </div>
  );
}

function StatCard({
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
    <div className="rounded-xl border border-[#161616] bg-[#0d0d0d] p-5 space-y-3">
      <div className="flex items-center gap-2 text-[#444] text-xs uppercase tracking-widest">
        {icon}
        {label}
      </div>
      <p
        className={cn(
          "text-lg font-semibold tracking-tight",
          truncate && "truncate",
          highlight ? "text-[#c8f557]" : "text-white",
        )}
      >
        {value}
      </p>
    </div>
  );
}

function ActionRow({
  icon,
  title,
  description,
  onClick,
  loading,
  danger,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
  onClick: () => void;
  loading?: boolean;
  danger?: boolean;
}) {
  return (
    <button
      onClick={onClick}
      disabled={loading}
      className={cn(
        "w-full px-6 py-4 flex items-center justify-between gap-4 text-left",
        "hover:bg-white/[0.02] transition-colors duration-150 disabled:opacity-50 disabled:cursor-not-allowed",
        "group",
      )}
    >
      <div className="flex items-center gap-3">
        <div
          className={cn(
            "w-8 h-8 rounded-lg flex items-center justify-center",
            danger ? "bg-red-500/10 text-red-400" : "bg-[#161616] text-[#555]",
            "group-hover:scale-105 transition-transform duration-150",
          )}
        >
          {loading ? (
            <span className="w-3.5 h-3.5 border border-current border-t-transparent rounded-full animate-spin" />
          ) : (
            icon
          )}
        </div>
        <div>
          <p
            className={cn(
              "text-sm font-medium",
              danger ? "text-red-400" : "text-white",
            )}
          >
            {title}
          </p>
          <p className="text-xs text-[#444] mt-0.5">{description}</p>
        </div>
      </div>
      <ChevronRight
        size={14}
        className="text-[#333] group-hover:text-[#555] transition-colors shrink-0"
      />
    </button>
  );
}
