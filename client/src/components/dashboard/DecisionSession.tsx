import { CheckCircle, Loader2, XCircle } from "lucide-react";

type DecisionSectionProps = {
  isTerminal: boolean;
  showRejectForm: boolean;
  setShowRejectForm: (value: boolean) => void;

  note: string;
  setNote: (value: string) => void;

  reason: string;
  setReason: (value: string) => void;

  handleApprove: () => void;
  handleReject: () => void;

  isPending: boolean;

  session: {
    status: string;
    rejection_reason?: string;
  };
};

export default function DecisionSection({
  isTerminal,
  showRejectForm,
  setShowRejectForm,
  note,
  setNote,
  reason,
  setReason,
  handleApprove,
  handleReject,
  isPending,
  session,
}: DecisionSectionProps) {
  if (isTerminal) {
    return (
      <div
        className={`rounded-xl p-4 text-sm font-medium ${
          session.status === "approved"
            ? "bg-green-500/10 text-green-600"
            : "bg-red-500/10 text-red-600"
        }`}
      >
        {session.status === "approved"
          ? "This session has been approved."
          : `This session was rejected: ${session.rejection_reason}`}
      </div>
    );
  }

  return (
    <div className="border border-border rounded-xl p-6 space-y-4">
      <h2 className="text-sm font-semibold">Decision</h2>

      <div>
        <label className="text-xs text-muted-foreground mb-1.5 block">
          Internal note (optional)
        </label>
        <textarea
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="Add an internal note..."
          rows={2}
          className="w-full text-sm px-3 py-2 rounded-lg border border-border bg-background resize-none focus:outline-none focus:ring-1 focus:ring-ring"
        />
      </div>

      {showRejectForm && (
        <div>
          <label className="text-xs text-muted-foreground mb-1.5 block">
            Rejection reason <span className="text-destructive">*</span>
          </label>
          <textarea
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="This will be shown to the user..."
            rows={2}
            className="w-full text-sm px-3 py-2 rounded-lg border border-destructive/50 bg-background resize-none focus:outline-none focus:ring-1 focus:ring-destructive"
          />
        </div>
      )}

      <div className="flex gap-3">
        {!showRejectForm ? (
          <>
            <button
              onClick={handleApprove}
              disabled={isPending}
              className="flex items-center gap-2 px-5 py-2.5 rounded-lg bg-green-600 hover:bg-green-700 text-white text-sm font-medium transition-colors disabled:opacity-50"
            >
              {isPending ? (
                <Loader2 size={14} className="animate-spin" />
              ) : (
                <CheckCircle size={14} />
              )}
              Approve
            </button>
            <button
              onClick={() => setShowRejectForm(true)}
              disabled={isPending}
              className="flex items-center gap-2 px-5 py-2.5 rounded-lg border border-destructive/50 text-destructive hover:bg-destructive/5 text-sm font-medium transition-colors disabled:opacity-50"
            >
              <XCircle size={14} />
              Reject
            </button>
          </>
        ) : (
          <>
            <button
              onClick={handleReject}
              disabled={isPending}
              className="flex items-center gap-2 px-5 py-2.5 rounded-lg bg-destructive hover:bg-destructive/90 text-white text-sm font-medium transition-colors disabled:opacity-50"
            >
              {isPending ? (
                <Loader2 size={14} className="animate-spin" />
              ) : (
                <XCircle size={14} />
              )}
              Confirm Rejection
            </button>
            <button
              onClick={() => {
                setShowRejectForm(false);
                setReason("");
              }}
              disabled={isPending}
              className="px-5 py-2.5 rounded-lg border border-border text-sm font-medium hover:bg-muted transition-colors"
            >
              Cancel
            </button>
          </>
        )}
      </div>
    </div>
  );
}
