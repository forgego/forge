---
sidebar_position: 32
description: Security policy and reporting.
---

# Security

## API response visibility

`api.BaseViewSet` omits schema fields marked `WriteOnly()` or `Serialize(false)` from list, detail, create, update, and partial update responses, including nested model relations. Serializer exclusions and `ExcludeResponseFields` can remove more fields. Custom API handlers should serialize models with `api.SerializeModel` or `api.SerializeMany` to apply schema visibility.

List responses check object permissions for each matching row. Denied rows are removed before pagination; `count` and page links describe only rows the requester can see. Detail requests for a denied row continue to return a permission error.

For reporting channels and disclosure guidance, see the full security policy:

- https://github.com/forgego/forge/blob/master/SECURITY.md

CSV exports prefix text cells beginning with `=`, `+`, `-`, `@`, a tab, or a carriage return with a single quote so spreadsheet applications treat them as text. Go numeric values, including negative numbers, remain numeric. Admin autocomplete and global search only return objects allowed by `HasViewPermission` for the current user.

Identity user endpoints enforce actor and target authorization. Ordinary users
may read and edit only their own profile; staff may list and read other users,
and superusers may manage other accounts and status flags. A forbidden user
update returns HTTP 403 before any changes are saved. Logout checks session
ownership before revoking a session.
