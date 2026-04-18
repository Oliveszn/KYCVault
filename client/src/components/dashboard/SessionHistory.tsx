import { KYCSessionResponse } from "@/types/kyc";
import { format } from "date-fns";

type Props = {
  sessions: KYCSessionResponse[];
};

export default function SessionHistory({ sessions }: Props) {
  return (
    <div className="space-y-4">
      <h2 className="text-lg font-semibold tracking-tight">Session History</h2>

      <div className="space-y-3">
        {sessions.map((session) => (
          <div
            key={session.id}
            className="p-4 rounded-xl border border-[#1a1a1a] bg-[#0f0f0f] hover:border-[#2a2a2a] transition"
          >
            <div className="flex items-center justify-between">
              <p className="text-sm font-medium text-white">
                {session.id_type.replace("_", " ")}
              </p>

              <span
                className={`text-xs px-2 py-1 rounded-full ${
                  session.status === "approved"
                    ? "bg-green-500/10 text-green-400"
                    : "bg-yellow-500/10 text-yellow-400"
                }`}
              >
                {session.status}
              </span>
            </div>

            <div className="mt-2 text-xs text-[#888] space-y-1">
              <p>Country: {session.country}</p>
              <p>Attempt: #{session.attempt_number}</p>
              <p>Created: {format(new Date(session.created_at), "PPP p")}</p>
              <p>Expires: {format(new Date(session.expires_at), "PPP p")}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
