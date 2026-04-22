import { useSessionQueue, useStatusCounts } from "@/hooks/useKyc";
import { useNavigate } from "react-router-dom";
import { Eye } from "lucide-react";
import { KYCSessionResponse } from "@/types/kyc";
import { tiles } from "@/config/adminTiles";

export default function AdminQueuePage() {
  const { data: queue, isLoading } = useSessionQueue();
  const { data: counts } = useStatusCounts();
  const navigate = useNavigate();

  const totalCount = counts
    ? Object.values(counts).reduce((a, b) => a + b, 0)
    : 0;

  return (
    <div className="min-h-screen bg-background text-foreground">
      <nav className="border-b border-border px-6 py-4">
        <div className="max-w-6xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-7 h-7 bg-blue-500 rounded-sm flex items-center justify-center">
              <span className="text-white font-black text-xs">K</span>
            </div>
            <span className="font-semibold tracking-wide text-sm">
              KYCVault Admin
            </span>
          </div>
          <span className="text-xs text-muted-foreground font-mono px-2 py-1 bg-muted rounded">
            ADMIN
          </span>
        </div>
      </nav>

      <main className="max-w-6xl mx-auto px-6 py-10 space-y-8">
        <div>
          <p className="text-xs font-mono tracking-widest uppercase text-muted-foreground mb-1">
            Admin
          </p>
          <h1 className="text-3xl font-bold tracking-tight">Review Queue</h1>
        </div>

        {/* Status tiles */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {tiles.map(({ label, key, icon: Icon, color }) => (
            <div
              key={label}
              className="border border-border rounded-xl p-4 space-y-2"
            >
              <div
                className={`w-8 h-8 rounded-lg flex items-center justify-center ${color}`}
              >
                <Icon size={16} />
              </div>
              <p className="text-2xl font-bold">
                {key ? (counts?.[key] ?? 0) : totalCount}
              </p>
              <p className="text-xs text-muted-foreground">{label}</p>
            </div>
          ))}
        </div>

        {/* Queue table */}
        <div className="border border-border rounded-xl overflow-hidden">
          <div className="px-5 py-4 border-b border-border">
            <h2 className="text-sm font-semibold">Pending Review</h2>
          </div>

          {isLoading && (
            <div className="px-5 py-10 text-sm text-muted-foreground text-center">
              Loading queue...
            </div>
          )}

          {!isLoading && queue?.sessions?.length === 0 && (
            <div className="px-5 py-10 text-sm text-muted-foreground text-center">
              No sessions pending review.
            </div>
          )}

          {!isLoading && queue?.sessions && queue.sessions.length > 0 && (
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border bg-muted/50">
                  <th className="text-left px-5 py-3 text-xs font-medium text-muted-foreground">
                    Session ID
                  </th>
                  <th className="text-left px-5 py-3 text-xs font-medium text-muted-foreground">
                    Country
                  </th>
                  <th className="text-left px-5 py-3 text-xs font-medium text-muted-foreground">
                    ID Type
                  </th>
                  <th className="text-left px-5 py-3 text-xs font-medium text-muted-foreground">
                    Attempt
                  </th>
                  <th className="text-left px-5 py-3 text-xs font-medium text-muted-foreground">
                    Submitted
                  </th>
                  <th className="px-5 py-3" />
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {queue.sessions.map((session: KYCSessionResponse) => (
                  <tr
                    key={session.id}
                    className="hover:bg-muted/30 transition-colors"
                  >
                    <td className="px-5 py-3.5 font-mono text-xs text-muted-foreground">
                      {session.id.slice(0, 8)}...
                    </td>
                    <td className="px-5 py-3.5">{session.country}</td>
                    <td className="px-5 py-3.5 capitalize">
                      {session.id_type?.replace("_", " ")}
                    </td>
                    <td className="px-5 py-3.5">#{session.attempt_number}</td>
                    <td className="px-5 py-3.5 text-muted-foreground">
                      {new Date(session.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-5 py-3.5 text-right">
                      <button
                        onClick={() =>
                          navigate(`/admin/sessions/${session.id}`)
                        }
                        className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-border text-xs font-medium hover:bg-muted transition-colors ml-auto"
                      >
                        <Eye size={12} />
                        Review
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </main>
    </div>
  );
}
