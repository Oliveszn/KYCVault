export interface Notification {
  ID: string;
  UserID: string;
  SessionId: string;
  Type: "kyc_approved" | "kyc_rejected";
  Message: string;
  Read: boolean;
  created_at: string;
}

export interface WSMessage {
  type: "notification";
  data: Notification;
}
