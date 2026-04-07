Auth (jwt) Mental model
Short expiry on access tokens (15-60 mins)
Refresh tokens stored securely (httpOnly cookie)
Refresh tokens rotated on every use
Token blacklist in place for logout
Sensitive data NOT stored in payload
Algorithm explicitly set (avoid "none")
Tokens validated on every request

Perfect! Since your code involves document uploads with KYC sessions, S3 storage, and session status updates, we can write a set of Postman test cases to cover the main flows. I’ll give you practical endpoints and checks. You can adapt the URLs and payloads based on your actual API.

1️⃣ Create a KYC session

Purpose: Ensure session creation works.

Method: POST
URL: http://localhost:8000/api/kyc/session
Body (JSON):
{
"user_id": "c56a4180-65aa-42ec-a945-5fd21dec0538",
"type": "individual"
}
Expected Response: 201 Created
{
"id": "generated-session-uuid",
"status": "initiated"
}
Test Assertions:
Status code = 201
Body contains status = "initiated"
Body contains id field
2️⃣ Upload a document (front side)

Purpose: Test document upload and S3 integration.

Method: POST
URL: http://localhost:8000/api/kyc/document
Headers:
Authorization: Bearer <access_token>
Body: form-data
session_id: UUID from step 1
side: front
file: <select file>
Expected Response: 201 Created
{
"id": "document-uuid",
"status": "pending",
"side": "front"
}
Test Assertions:
Status code = 201
Body contains side = "front"
Body contains status = "pending"
3️⃣ Upload a document (back side)
Same as step 2, but side = "back"
Expected:
Two documents exist for the session (front + back)
BothSidesAccepted should eventually return true if statuses are updated manually to accepted.
4️⃣ Advance session to doc_upload

Purpose: Test atomic update with AdvanceStatusIfCurrent.

Method: POST
URL: http://localhost:8000/api/kyc/session/advance
Body (JSON):
{
"session_id": "UUID-from-step-1",
"from_status": "initiated",
"to_status": "doc_upload"
}
Expected Response:
{
"updated": true,
"status": "doc_upload"
}
Test Assertions:
updated = true (if no one else advanced)
If you call again → updated = false
5️⃣ Check BothSidesAccepted

Purpose: Ensure your race-safe count query works.

Method: GET
URL: http://localhost:8000/api/kyc/session/<session_id>/both_sides_accepted
Expected Response:
{
"both_sides_accepted": false
}
After setting both front/back status = accepted in DB:
{
"both_sides_accepted": true
}
6️⃣ Advance to face_verify
Method: POST
URL: http://localhost:8000/api/kyc/session/advance
Body (JSON):
{
"session_id": "UUID-from-step-1",
"from_status": "doc_upload",
"to_status": "face_verify"
}
Expected:
Returns updated = true only if BothSidesAccepted = true
Otherwise updated = false
Optional: Edge Cases
Upload same side again → should return 409 Conflict (ErrDocumentSideExists)
Advance session from wrong current status → should return updated = false
Upload document without auth → 401 Unauthorized

If you want, I can draft a ready-to-import Postman collection with all these test cases prefilled so you can just hit “Send” and validate everything end-to-end.

Do you want me to do that?
