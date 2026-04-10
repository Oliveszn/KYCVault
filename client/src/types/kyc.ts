export interface InitiateSessionPayload {
  country: string;
  IDType: string;
}

export interface ApiResponse<T> {
  message: string;
  payload: T;
}

export interface KYCSessionResponse {
  id: string;
  userId: string;
  status: string;
  country: string;
  idType: string;
  attemptNumber: number;
  expiresAt: string;
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
