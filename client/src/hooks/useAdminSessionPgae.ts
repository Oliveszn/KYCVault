import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  useAdminSession,
  useApproveSession,
  useRejectSession,
} from "@/hooks/useKyc";
import { toast } from "sonner";
import { useFaceVerificationAdmin } from "./useFaceVerify";

export const useAdminSessionPage = (id: string) => {
  const navigate = useNavigate();

  const { data: session, isLoading } = useAdminSession(id);
  const { data: faceVerification } = useFaceVerificationAdmin(id);
  const approveSession = useApproveSession();
  const rejectSession = useRejectSession();

  const [showRejectForm, setShowRejectForm] = useState(false);
  const [note, setNote] = useState("");
  const [reason, setReason] = useState("");

  const handleApprove = () => {
    approveSession.mutate(
      { sessionId: id, payload: { note } },
      {
        onSuccess: () => {
          toast.success("Session approved");
          navigate("/admin");
        },
      },
    );
  };

  const handleReject = () => {
    if (!reason.trim()) {
      toast.error("Rejection reason is required");
      return;
    }

    rejectSession.mutate(
      { sessionId: id, payload: { note, reason } },
      {
        onSuccess: () => {
          toast.success("Session rejected");
          navigate("/admin");
        },
      },
    );
  };

  const frontDoc = session?.documents?.find((d) => d.side === "front");
  const backDoc = session?.documents?.find((d) => d.side === "back");

  const isTerminal = ["approved", "rejected"].includes(session?.status ?? "");
  const isPending = approveSession.isPending || rejectSession.isPending;

  return {
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
  };
};
