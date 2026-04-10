Auth (jwt) Mental model
Short expiry on access tokens (15-60 mins)
Refresh tokens stored securely (httpOnly cookie)
Refresh tokens rotated on every use
Token blacklist in place for logout
Sensitive data NOT stored in payload
Algorithm explicitly set (avoid "none")
Tokens validated on every request
