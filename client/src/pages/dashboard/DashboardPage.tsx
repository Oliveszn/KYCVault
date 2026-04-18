import { LogOut } from "lucide-react";
import { cn } from "@/lib/utils";
import { useAppSelector } from "@/store/hooks";
import { useLogout } from "@/hooks/useAuth";
import { selectUser } from "@/store/auth-slice";
import { useSessionHistory } from "@/hooks/useKyc";
import SessionHistory from "@/components/dashboard/SessionHistory";
import { Button } from "@/components/ui/button";
import { useNavigate } from "react-router-dom";

export default function DashboardPage() {
  const user = useAppSelector(selectUser);
  const logout = useLogout();
  const { data, isLoading, isError, error } = useSessionHistory();
  const navigate = useNavigate();
  return (
    <div className="min-h-screen text-black">
      <nav className="border-b border-[#161616] px-6 py-4">
        <div className="max-w-5xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-7 h-7 bg-blue-500 rounded-sm flex items-center justify-center">
              <span className="text-black font-black text-xs">K</span>
            </div>
            <span className="text-black font-semibold tracking-wide text-sm">
              KYCVault
            </span>
          </div>

          <div className="flex items-center gap-3">
            <div className="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full bg-blue-500/10 border border-blue-400/20">
              <div className="w-1.5 h-1.5 rounded-full bg-blue-500 animate-pulse" />
              <span className="text-blue-500 text-xs font-medium">
                Session active
              </span>
            </div>

            <button
              onClick={() => logout.mutate()}
              disabled={logout.isPending}
              className={cn(
                "flex items-center gap-2 px-3.5 py-2 rounded-lg text-sm",
                "text-[#555] hover:text-red-500 border border-[#1a1a1a] hover:border-[#2a2a2a] cursor-pointer",
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

      <main className="max-w-7xl mx-auto px-6 py-12 space-y-10">
        <div className="space-y-1 animate-in fade-in slide-in-from-bottom-3 duration-500">
          <p className="text-black text-sm font-mono tracking-[0.2em] uppercase">
            Dashboard
          </p>
          <h1 className="text-4xl font-bold tracking-tight">
            Welcome back
            {user?.firstName ? (
              <span className="text-[#c8f557]">, {user.firstName}</span>
            ) : null}
            .
          </h1>
          <Button
            onClick={() => navigate("/kyc/sessions/new")}
            className="cursor-pointer"
          >
            Initiate Session
          </Button>
        </div>

        <div className="animate-in fade-in slide-in-from-bottom-3 duration-500">
          {/* Loading */}
          {isLoading && (
            <div className="text-sm text-[#666]">Loading sessions...</div>
          )}

          {/* Error */}
          {isError && (
            <div className="text-sm text-red-400">Failed to load sessions</div>
          )}

          {/* Empty state */}
          {!isLoading && !isError && data?.count === 0 && (
            <div className="p-6 border border-[#1a1a1a] rounded-xl text-center text-sm text-[#777]">
              No sessions yet.
            </div>
          )}

          {/* Success */}
          {!isLoading && !isError && data && data?.sessions?.length > 0 && (
            <SessionHistory sessions={data.sessions} />
          )}
        </div>
      </main>
    </div>
  );
}
