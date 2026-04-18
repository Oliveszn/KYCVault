type KYCStatus =
  | "initiated"
  | "doc_upload"
  | "face_verify"
  | "in_review"
  | "approved"
  | "rejected";
type IDType =
  | "national_id"
  | "drivers_license"
  | "passport"
  | "residence_permit";
export interface InitiateSessionPayload {
  country: string;
  id_type: IDType;
}

export interface ApiResponse<T> {
  message: string;
  payload: T;
}

export type ApiError = {
  message: string;
  code?: string;
};

export interface KYCSessionResponse {
  id: string;
  userId: string;
  status: KYCStatus;
  country: string;
  id_type: IDType;
  attempt_number: number;
  expires_at: string;
  created_at: string;
}

export interface SessionHistoryResponse {
  sessions: KYCSessionResponse[];
  count: number;
}

export interface SessionQueueResponse {
  sessions: KYCSessionResponse[];
  total: number;
  limit: number;
  offset: number;
}

export type StatusCounts = Record<string, number>;

export interface ApproveSessionPayload {
  note: string;
}

export interface RejectSessionPayload {
  note: string;
  reason: string;
}
