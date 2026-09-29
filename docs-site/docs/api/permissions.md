---
sidebar_position: 5
description: Permission classes for API access control.
image: /social-card.png
---

# Permissions

Permissions control access to viewsets and objects.

## Built-in permissions

- AllowAny
- IsAuthenticated
- IsAdminUser
- IsOwnerOrReadOnly

## Denied requests

A request that fails a permission is answered like Django REST framework:

- **Unauthenticated, challenge available: 401.** When no authentication class
  identified the caller and the first configured class can issue a
  `WWW-Authenticate` challenge, the response is `401 Not Authenticated` with
  that header: `Token` for `TokenAuthentication`, `Bearer` for
  `JWTAuthentication`, `Basic realm="api"` for `BasicAuthentication`.
- **Unauthenticated, no challenge: 403.** Session and API key authentication
  issue no challenge, so with one of them first (or with no authentication
  classes) the response is `403 Permission Denied`.
- **Authenticated: 403.** A caller who authenticated but lacks the permission
  always gets `403 Permission Denied`.

Invalid credentials are rejected with 401 before permissions run, with the
same challenge header when available.

## Next steps

- [Throttling](/docs/api/throttling/)
- [Errors](/docs/api/errors/)
