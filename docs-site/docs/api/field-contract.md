---
sidebar_position: 3
description: Which fields a BaseViewSet reads from requests and writes to responses, unknown request keys, null and omitted values, and PUT versus PATCH.
image: /social-card.png
---

# Request & Response Field Contract

This page describes what `api.BaseViewSet` does with each field of a model: whether it appears in responses, whether clients can write it, and what happens to request keys that match no field. The generated `<Model>ViewSet` uses the strictest settings, which are shown in [Generated viewsets](#generated-viewsets).

---

## Schema flags

A model's `Fields()` schema sets two flags on each field. Together they give the field one of four roles:

| `Editable` | `Serialize` | Role | Responses | Requests |
| :--- | :--- | :--- | :--- | :--- |
| `true` | `true` | read-write (the default) | included | written |
| `false` | `true` | read-only | included | ignored |
| `true` | `false` | write-only, such as `schema.WriteOnly()` for a password | omitted | written |
| `false` | `false` | hidden | omitted | ignored; rejected as unknown in strict mode |

Some fields are read-only even with `Editable: true`, because the database owns their values: an auto-increment primary key, `AutoNow`, `AutoNowAdd` and `Generated` columns.

The viewset enforces these roles only through two fields that you set from the schema. The generated viewset sets both:

```go
vs.ExcludeResponseFields = api.NonSerializableFields(&Product{}) // drop Serialize(false) fields from responses
vs.ReadOnlyRequestFields = api.NonEditableFields(&Product{})     // ignore read-only and database-owned fields in requests
```

---

## Responses

A response contains every exported struct field that has a `json` tag other than `-`, under its JSON name. The viewset then removes the following keys, in order:

1. `ExcludeResponseFields`.
2. Keys not listed in the serializer's `Fields()`, when that list is non-empty. A nil or empty `Fields()` keeps every field.
3. The serializer's `Exclude()` and `WriteOnlyFields()`.

`BaseViewSet` has no computed response fields. Every key comes from a struct field. A value derived from other columns belongs in a `Generated` schema field, which the database computes and the viewset treats as read-only. Values computed in Go belong in a custom handler or a [router action](/docs/api/viewsets/).

## Requests

`POST`, `PUT` and `PATCH` bodies must be JSON objects. For each key, the viewset does one of the following:

- **Writes it** when it names a field by its JSON name or, for a field in the schema, by its schema name, column name, `db` tag or Go name. The value is converted to the field's Go type, and a value that cannot be converted fails with `400`.
- **Ignores it** when it names a read-only field: a key in `ReadOnlyRequestFields`, one of the serializer's `ReadOnlyFields()`, or the primary key. These are matched without regard to case. The primary key in a `PUT` or `PATCH` body never changes which row is written. That row comes from the URL.
- **Ignores or rejects it** when it names nothing. See [Unknown keys](#unknown-keys).

The serializer's `Fields()` only trims responses. It does not limit which fields a request can write. Use `ReadOnlyRequestFields` or `ReadOnlyFields()` for that.

### Unknown keys

A key is known when it names a field that the viewset writes or ignores, as listed above, or when the serializer declares it in `Fields()`, `ReadOnlyFields()` or `WriteOnlyFields()`. A hidden field (`Editable: false`, `Serialize: false`) is not known, under any spelling.

By default, unknown keys are ignored, as in Django REST Framework. Setting `RejectUnknownRequestFields` rejects them instead:

```go
vs.RejectUnknownRequestFields = true
```

A request with unknown keys then fails with `400` before anything else runs, and the error names each key (response abbreviated):

```json
{
  "status": 400,
  "code": "VALIDATION_ERROR",
  "detail": "Validation failed",
  "errors": {
    "titel": [{ "message": "Unknown field.", "code": "INVALID_FIELD" }]
  }
}
```

Read-only keys that clients echo back from an earlier response, such as `id` and `created_at`, are known, so they are still ignored silently. A hidden field is reported exactly like a key that does not exist, so the response reveals nothing about it. If your serializer reads an input that has no model field, such as `password_confirm`, declare it in `WriteOnlyFields()` so that strict mode accepts it.

### Null, omitted and zero

| Request value | Effect |
| :--- | :--- |
| key omitted | The field keeps its stored value. On create, it gets its schema default, or the Go zero value if there is none. |
| explicit zero (`0`, `""`, `false`) | The zero value is written. |
| `null` on a pointer, slice, map or interface field | The field is cleared to nil. |
| `null` on any other field, such as `string`, `int64` or `bool` | The value is ignored, so the field keeps its current value. It is not set to zero. |
| `null` on a `TypeJSON` field | JSON `null` is written: the bytes `null` for a `[]byte` field, otherwise the zero value. |

### PUT and PATCH

`PUT` is a full update. Its body must name every required, writable schema field that has no default, and a missing one fails with `400` naming the field. Optional fields left out of a `PUT` keep their stored values; `PUT` does not reset them. `PATCH` is a partial update and writes only the keys present. After the keys are applied, the whole object is validated, so a `PATCH` that leaves a required field empty also fails with `400`.

---

## Validation order

`create`, `update` and `partial_update` run their steps in the order below. A failure at any step returns `400` and stops the request. Business hooks (`BeforeCreate`, `BeforeSave`, `BeforeUpdate`) and persistence run only in the last step, inside the manager, so a failed request leaves stored data unchanged and runs no hooks.

1. Parse the JSON body.
2. Reject unknown keys, when `RejectUnknownRequestFields` is set.
3. Run the serializer's `Validate()`.
4. Load the stored object (update only) and apply the body, converting each value to its field's type.
5. Validate the model: `Clean()`, schema `Clean` hooks, `Validate()`, and schema constraints such as `Required` and `MaxLength`.
6. Call the manager's `Create` or `Update`, which runs hooks and writes.

---

## Generated viewsets

`forge generate --api` emits a viewset for each model with this contract:

```go
vs.ExcludeResponseFields = api.NonSerializableFields(&Product{})
vs.ReadOnlyRequestFields = api.NonEditableFields(&Product{})
vs.RejectUnknownRequestFields = true
```

A generated API rejects a misspelled key with `400` instead of silently dropping it. Hand-written `BaseViewSet` and `ViewSetConfig` values keep the default, which ignores unknown keys. Both have a `RejectUnknownRequestFields` field for opting in.
