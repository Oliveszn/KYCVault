export interface FaceVerification {
  id: string;
  sessionId: string;
  userId: string;

  status: "pending" | "passed" | "failed";

  attemptCount: number;

  matchScore?: number;
  matchThreshold?: number;
  matchPassed?: boolean;

  vendorName?: string;
  vendorRequestId?: string;

  failureReason?: string;
}
