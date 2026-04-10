type side = "front" | "back";
type status = "pending" | "accepted" | "rejected";
export interface KYCDocument {
  id: string;
  sessionId: string;
  userId: string;
  side: side;
  status: status;
  storageKey: string;
  storageBucket: string;
  originalFilename: string;
  mimeType: string;
  fileSizeBytes: number;
  checksum: string;
}

export interface DocumentListResponse {
  documents: KYCDocument[];
  count: number;
}

export interface PresignedURLResponse {
  url: string;
  expires_in: string;
}
