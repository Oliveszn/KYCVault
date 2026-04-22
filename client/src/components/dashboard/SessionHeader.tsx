import { AdminSessionResponse } from "@/types/kyc";

type Props = {
  session: AdminSessionResponse;
};
export default function SessionHeader({ session }: Props) {
  return (
    <div className="flex items-start justify-between">
      <div>
        <p className="text-xs font-mono text-muted-foreground mb-1">
          Session {session.id.slice(0, 8)}...
        </p>
        <h1 className="text-2xl font-bold">Identity Review</h1>
      </div>
      <span
        className={`px-3 py-1 rounded-full text-xs font-medium capitalize ${
          session.status === "approved"
            ? "bg-green-500/10 text-green-500"
            : session.status === "rejected"
              ? "bg-red-500/10 text-red-500"
              : "bg-yellow-500/10 text-yellow-500"
        }`}
      >
        {session.status.replace("_", " ")}
      </span>
    </div>
  );
}
