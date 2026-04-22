import { useParams, useNavigate } from "react-router-dom";
import { ChevronLeft, Loader2 } from "lucide-react";
import SelfiePreview from "@/components/dashboard/AdminSelfiePreview";
import SessionHeader from "@/components/dashboard/SessionHeader";
import SessionInfo from "@/components/dashboard/SessionInfo";
import { useAdminSessionPage } from "@/hooks/useAdminSessionPgae";
import DecisionSection from "@/components/dashboard/DecisionSession";
import DocumentsSection from "@/components/dashboard/DocumentSection";

export default function AdminSessionPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const {
    session,
    isLoading,
    frontDoc,
    backDoc,
    isTerminal,
    isPending,
    showRejectForm,
    setShowRejectForm,
    note,
    setNote,
    reason,
    setReason,
    handleApprove,
    handleReject,
    faceVerification,
  } = useAdminSessionPage(id!);

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Loader2 className="animate-spin" />
      </div>
    );
  }

  if (!session) return null;
  return (
    <div className="min-h-screen bg-background text-foreground">
      <nav className="border-b border-border px-6 py-4">
        <div className="max-w-5xl mx-auto flex items-center gap-4">
          <button
            onClick={() => navigate("/admin")}
            className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
          >
            <ChevronLeft size={16} />
            Back to queue
          </button>
        </div>
      </nav>

      {session && (
        <main className="max-w-5xl mx-auto px-6 py-10 space-y-8">
          <SessionHeader session={session} />

          <SessionInfo session={session} />

          <DocumentsSection frontDoc={frontDoc} backDoc={backDoc} />

          {faceVerification?.id ? (
            <SelfiePreview verificationId={faceVerification.id} />
          ) : (
            <p className="text-sm text-muted-foreground">No selfie available</p>
          )}

          <DecisionSection
            session={session}
            isTerminal={isTerminal}
            showRejectForm={showRejectForm}
            setShowRejectForm={setShowRejectForm}
            note={note}
            setNote={setNote}
            reason={reason}
            setReason={setReason}
            handleApprove={handleApprove}
            handleReject={handleReject}
            isPending={isPending}
          />
        </main>
      )}
    </div>
  );
}
