type status = "pending" | "passed" | "failed";
export interface FaceVerification {
  id: string;
  sessionId: string;
  userId: string;

  status: status;

  attemptCount: number;

  matchScore?: number;
  matchThreshold?: number;
  matchPassed?: boolean;

  vendorName?: string;
  vendorRequestId?: string;

  failureReason?: string;
}
