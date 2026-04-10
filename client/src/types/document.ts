export interface KYCDocument {
  id: string;
  sessionId: string;
  userId: string;
  side: "front" | "back";
  status: "pending" | "accepted" | "rejected";
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
