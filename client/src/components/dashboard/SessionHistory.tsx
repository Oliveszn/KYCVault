import { statusConfig } from "@/config/statusConfig";
import { cn } from "@/lib/utils";
import { KYCSessionResponse } from "@/types/kyc";
import { format } from "date-fns";
import { Globe, Hash, Calendar, Clock, ChevronRight } from "lucide-react";
import { useNavigate } from "react-router-dom";

type Props = {
  sessions: KYCSessionResponse[];
};

function StatusBadge({ status }: { status: string }) {
  const cfg = statusConfig[status] ?? {
    label: status,
    dot: "bg-gray-400",
    text: "text-gray-500",
    bg: "bg-gray-100",
  };
  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[11px] font-medium ${cfg.text} ${cfg.bg}`}
    >
      <span className={`w-1.5 h-1.5 rounded-full ${cfg.dot}`} />
      {cfg.label}
    </span>
  );
}

const getResumeRoute = (session: KYCSessionResponse) => {
  switch (session.status) {
    case "initiated":
      return `/kyc/sessions/${session.id}/documents`;
    case "doc_upload":
      return `/kyc/sessions/${session.id}/face`;
    case "face_verify":
      return `/kyc/sessions/${session.id}/face`;
    default:
      return null; // terminal states — no resume
  }
};

export default function SessionHistory({ sessions }: Props) {
  const navigate = useNavigate();

  const handleClick = (session: KYCSessionResponse) => {
    const route = getResumeRoute(session);
    if (route) navigate(route);
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-5">
        <h2
          className="text-xs tracking-[0.2em] uppercase text-gray-400"
          style={{ fontFamily: "'DM Mono', monospace" }}
        >
          Session History
        </h2>
        <span
          className="text-xs text-gray-400"
          style={{ fontFamily: "'DM Mono', monospace" }}
        >
          {sessions.length} {sessions.length === 1 ? "record" : "records"}
        </span>
      </div>

      <div className="space-y-2">
        {sessions.map((session, i) => (
          <div
            key={session.id}
            onClick={() => handleClick(session)}
            className={cn(
              "group relative p-4 rounded-xl border border-gray-100 bg-gray-50/60 transition-all duration-200",
              getResumeRoute(session)
                ? "hover:bg-white hover:border-gray-200 hover:shadow-sm cursor-pointer"
                : "cursor-default opacity-80",
            )}
            style={{ animationDelay: `${i * 60}ms` }}
          >
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-3 mb-3">
                  <p className="text-sm font-medium text-gray-800 capitalize">
                    {session.id_type?.replace(/_/g, " ")}
                  </p>
                  <StatusBadge status={session.status} />
                </div>

                <div className="grid grid-cols-2 gap-x-6 gap-y-1.5">
                  <div className="flex items-center gap-2 text-xs text-gray-400">
                    <Globe size={11} className="shrink-0" />
                    <span>{session.country}</span>
                  </div>
                  <div className="flex items-center gap-2 text-xs text-gray-400">
                    <Hash size={11} className="shrink-0" />
                    <span>Attempt {session.attempt_number}</span>
                  </div>
                  <div className="flex items-center gap-2 text-xs text-gray-400">
                    <Calendar size={11} className="shrink-0" />
                    <span>
                      {format(new Date(session.created_at), "dd MMM yyyy")}
                    </span>
                  </div>
                  {session.expires_at && (
                    <div className="flex items-center gap-2 text-xs text-gray-400">
                      <Clock size={11} className="shrink-0" />
                      <span>
                        Exp.{" "}
                        {format(new Date(session.expires_at), "dd MMM yyyy")}
                      </span>
                    </div>
                  )}
                </div>
              </div>

              <div className="flex flex-col items-end gap-2 shrink-0">
                {getResumeRoute(session) ? (
                  <span className="text-[10px] text-blue-500 font-medium group-hover:underline">
                    Continue →
                  </span>
                ) : (
                  <ChevronRight
                    size={14}
                    className="text-gray-200 group-hover:text-gray-400 transition-colors"
                  />
                )}
                <span className="text-[10px] text-gray-300 font-mono">
                  {session.id.slice(0, 8)}
                </span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
