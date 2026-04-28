import { LogOut, Plus, ShieldCheck } from "lucide-react";
import { cn } from "@/lib/utils";
import { useAppSelector } from "@/store/hooks";
import { useLogout } from "@/hooks/useAuth";
import { selectUser } from "@/store/auth-slice";
import { useSessionHistory } from "@/hooks/useKyc";
import SessionHistory from "@/components/dashboard/SessionHistory";
import { useNavigate } from "react-router-dom";
import NotificationBell from "@/components/dashboard/NotificationBell";

export default function DashboardPage() {
  const user = useAppSelector(selectUser);
  const logout = useLogout();
  const { data, isLoading, isError } = useSessionHistory();
  const navigate = useNavigate();
  return (
    <div
      className="min-h-screen bg-gray-50 text-black"
      style={{ fontFamily: "'DM Sans', sans-serif" }}
    >
      <style>{`
        @import url('https://fonts.googleapis.com/css2?family=DM+Sans:wght@300;400;500;600&family=DM+Mono:wght@400;500&display=swap');
        .glow-line { background: linear-gradient(90deg, transparent, rgba(96,165,250,0.35), transparent); }
        @keyframes fadeUp { from { opacity: 0; transform: translateY(16px); } to { opacity: 1; transform: translateY(0); } }
        .fade-up-1 { animation: fadeUp 0.5s 0.1s ease both; }
        .fade-up-2 { animation: fadeUp 0.5s 0.2s ease both; }
        .fade-up-3 { animation: fadeUp 0.5s 0.3s ease both; }
      `}</style>

      {/* Nav */}
      <nav className="border-b border-gray-100 px-6 py-4 bg-white/80 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-4xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-2.5">
            <div className="w-7 h-7 bg-blue-500 rounded-md flex items-center justify-center shadow-md shadow-blue-500/20">
              <ShieldCheck size={14} className="text-white" />
            </div>
            <span
              className="font-semibold text-sm tracking-wide text-gray-800"
              style={{ fontFamily: "'DM Mono', monospace" }}
            >
              KYCVault
            </span>
          </div>

          <div className="flex items-center gap-2">
            <NotificationBell />
            <button
              onClick={() => logout.mutate()}
              disabled={logout.isPending}
              className={cn(
                "flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium cursor-pointer",
                "text-gray-400 hover:text-gray-600 border border-gray-200 hover:border-gray-300 bg-white",
                "transition-all duration-200 disabled:opacity-30",
              )}
            >
              {logout.isPending ? (
                <span className="w-3 h-3 border border-gray-300 border-t-gray-600 rounded-full animate-spin" />
              ) : (
                <LogOut size={12} />
              )}
              Sign out
            </button>
          </div>
        </div>
      </nav>

      <main className="max-w-4xl mx-auto px-6 py-14">
        {/* Header */}
        <div className="mb-12 fade-up-1">
          <p
            className="text-[10px] tracking-[0.25em] uppercase text-gray-400 mb-3"
            style={{ fontFamily: "'DM Mono', monospace" }}
          >
            Identity Vault
          </p>
          <h1 className="text-[2.5rem] font-semibold leading-[1.15] tracking-tight text-gray-900">
            {user?.firstName ? (
              <>
                Hey, <span className="text-blue-500">{user.firstName}</span>.
              </>
            ) : (
              "Your dashboard."
            )}
          </h1>
          <p className="text-gray-400 text-sm mt-2 font-light">
            Manage your identity verification sessions.
          </p>
        </div>

        {/* CTA */}
        <div className="fade-up-2 mb-10">
          <button
            onClick={() => navigate("/kyc/sessions/new")}
            className="group flex items-center gap-2.5 px-5 py-2.5 rounded-xl bg-blue-500 hover:bg-blue-600 text-white text-sm font-medium transition-all duration-200 shadow-md shadow-blue-500/20 hover:shadow-blue-600/25"
          >
            <Plus
              size={15}
              className="group-hover:rotate-90 transition-transform duration-200"
            />
            New Verification
          </button>
        </div>

        {/* Divider */}
        <div className="glow-line h-px mb-10 fade-up-2" />

        {/* Session history */}
        <div className="fade-up-3">
          {isLoading && (
            <div className="flex items-center gap-3 text-sm text-gray-400 py-8">
              <span className="w-4 h-4 border border-gray-200 border-t-gray-500 rounded-full animate-spin" />
              Loading sessions...
            </div>
          )}
          {isError && (
            <div className="text-sm text-red-400 py-8">
              Failed to load sessions.
            </div>
          )}
          {!isLoading && !isError && data?.count === 0 && (
            <div className="py-16 text-center border border-gray-100 rounded-2xl bg-gray-50/50">
              <div className="w-10 h-10 rounded-xl bg-gray-100 border border-gray-200 flex items-center justify-center mx-auto mb-4">
                <ShieldCheck size={18} className="text-gray-300" />
              </div>
              <p className="text-sm text-gray-400 font-light">
                No verification sessions yet.
              </p>
              <p className="text-xs text-gray-300 mt-1">
                Start one to verify your identity.
              </p>
            </div>
          )}
          {!isLoading && !isError && data && data.sessions?.length > 0 && (
            <SessionHistory sessions={data.sessions} />
          )}
        </div>
      </main>
    </div>
  );
}
