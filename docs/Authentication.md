# Authentication
Minitwit uses JWT to authenticate users without relying on storing state on the server. This saves us a round trip to the database. 

## Middleware
The system uses two different authentication middlewares depending on if authentication is required to access the endpoint.  
**RequiredAuth**: Ensures that user is authenticated by checking for a valid JWT and redirecting to the login page if no valid JWT is found.  
**OptionalAuth**: Authenticates user if a valid JWT exists, but allows the request to go through to the handler no matter what. 

## Specifications about the JWT
**Claims in JWT**:
- UserId
- Username
- Email
- Issue time
- Expirery time (1 day after issue time)

**Hashing algorithm**: HS256 (HMAC with SHA-256)  
We use symmetric signing system since only the minitwit servers need to validate the JWT, and symmetric cryptography is faster.