# Backend handoff — admin metadata & list serialization

**Audience:** the agent that owns the Go side of this repo.
**Author:** the frontend/design agent. Nothing here has been implemented on the Go
side by this session; the frontend has shipped only what it can do without backend
changes, and each item below says exactly what the frontend already does today.

Two items. They are independent and can ship in either order.

---

## BH-1 — `read_only` is not set on auto-managed timestamp fields

### Symptom

On a **create** form, auto-managed columns (`created_at`, `updated_at`, and any
field the ORM populates itself) render as editable inputs. A user can type into
them. Whatever they type is discarded by the backend, so the form silently lies.

### What this is NOT

The other review of this UI attributed the bug to
`src/pages/ModelUpsertPage.tsx:481`. That is wrong — that line already guards
correctly:

```tsx
// ModelUpsertPage.tsx — already correct, do not change
if (field.read_only && !isCreate) return null;
```

The frontend honours `read_only` faithfully. The defect is that the metadata
emitter never sets the flag for these fields, so the frontend has nothing to
honour.

### Root cause

The admin metadata emitter builds `FieldMetadata` from the model's struct tags. It
sets `read_only` only when the field carries an explicit `admin:"readonly"` tag. It
does not consult whether the ORM manages the column's value (auto-now / auto-now-add
semantics, DB-side `DEFAULT`, or a generated/identity column).

### Required change

In the metadata emitter, set `read_only: true` for any field where **any** of the
following holds:

1. the field is populated by the ORM on insert or update (auto-now-add / auto-now);
2. the column is DB-generated (identity, serial, or `GENERATED ... AS`);
3. the field is the primary key **and** the PK is not user-assigned.

Keep the existing explicit `admin:"readonly"` tag working — the new rule is an
addition, not a replacement.

### Contract the frontend depends on

`FieldMetadata.read_only` is already in the wire type
(`forge/admin/ui/web/src/api/types.ts:9`). No TypeScript change is needed. Setting
the flag correctly is sufficient; the UI will immediately stop rendering those
inputs on create and will render them as static values on edit.

### How to verify

Fetch `GET /admin/api/meta/<model>` for a model with an auto timestamp and assert
`read_only == true` on that field. Then load the admin create form for the same
model and confirm the field is no longer an input.

---

## BH-2 — relation display values in list serialization

### Symptom

Every foreign-key column in the admin list view shows a bare integer. A user
scanning an orders table sees `#4182` where they need a customer name.

### What the frontend does today (phase 1, already shipped)

`src/pages/ModelListPage.tsx` renders an FK cell as a **linked id**: a monospace
`#<id>` button that navigates to the related record's detail page. That is a real
improvement over dead text and it is the most that can be done client-side. It is
deliberately built so that phase 2 is a drop-in: the cell renderer will prefer a
display value when one is present and fall back to the linked id when it is not.

### Why the frontend cannot solve this alone

The only relation endpoint is `GET /{model}/autocomplete`, and it is a **text
search** gated on `query.length > 0`. It cannot batch-resolve a set of ids. Three
client-side options were considered and rejected:

| Option | Why rejected |
|---|---|
| Fetch the whole related table and join in the browser | Unbounded. A 200k-row lookup table breaks the page. |
| Add a `/labels?ids=` endpoint and call it per FK column | Still N extra round trips per page render, and a new endpoint to maintain. |
| One request per visible FK cell | Catastrophic; 25 rows × 3 FK columns = 75 requests. |

The correct place to solve it is the list serializer, which already has the row in
hand and can join once. This is what Django admin does.

### Required change

When serializing a list row, for each relation field also emit a sibling key
`<field>__display` carrying the related object's human label:

```json
{
  "id": 4182,
  "customer_id": 91,
  "customer_id__display": "Acme Industrial GmbH",
  "total": "1249.00"
}
```

Rules:

- The label should be the related model's string representation — the same value
  `/{model}/autocomplete` returns in `AutocompleteItem.label`, so the two surfaces
  agree.
- Emit the key **only** when the relation resolves. A null FK emits neither the
  display key nor an empty string; the frontend already renders an em-dash for a
  missing value and must not be handed `""`.
- Resolve with a single joined query or one batched lookup per relation per page —
  never one query per row. This is the whole point of the change.
- Apply it to `foreign_key` and `one_to_one`. `many_to_many` is out of scope here.

### Suggested phasing

`core.Metadata` has six callers plus tests, so the change was split so each half
ships independently and the two are compatible in either deployment order:

- **BH-2a — serializer only.** Emit `<field>__display` from the list serializer.
  Additive, so it breaks no existing consumer. The frontend can start reading it
  as soon as it appears.
- **BH-2b — metadata flag.** Add a boolean to the relation metadata announcing that
  the backend emits display values, so the frontend can distinguish "this backend
  is old" from "this particular FK is null". Optional; the frontend's fallback is
  correct without it.

### Contract the frontend will consume

The frontend change is one branch in the ModelListPage cell renderer:

```
if (obj[`${fieldName}__display`]) -> render the label, linked
else if (val != null)             -> render #<id>, linked      (today's behaviour)
else                              -> render <EmptyValue/>
```

No new TypeScript type is required — list rows are already `Record<string, any>`.

### How to verify

`GET /admin/api/<model>?page=1` on a model with a populated FK returns the
`__display` sibling key, and the same request on a model with a null FK omits it.
Confirm the query count does not scale with page size.
