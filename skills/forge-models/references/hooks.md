# Model Hooks and Validation

Lifecycle hooks run inside `orm.Manager` write paths (`Create`, `Update`, `Save`, `Delete`).

## Defining `Hooks()`

```go
package users

import (
	"context"
	"strings"

	"github.com/forgego/forge/identity"
	"github.com/forgego/forge/schema"
)

type User struct {
	UserGenerated
}

func (User) Hooks() *schema.ModelHooks {
	return &schema.ModelHooks{
		BeforeCreate: func(ctx context.Context, instance interface{}) error {
			u := instance.(*User)
			if u.Password != "" && !strings.HasPrefix(u.Password, "$2a$") {
				hashed, err := identity.HashPassword(u.Password)
				if err != nil {
					return err
				}
				u.Password = hashed
			}
			return nil
		},
		BeforeSave: func(ctx context.Context, instance interface{}) error {
			u := instance.(*User)
			u.Email = strings.TrimSpace(strings.ToLower(u.Email))
			return nil
		},
		Clean: func(instance interface{}) error {
			// Note: Clean takes (instance interface{}) error — no context.Context.
			return nil
		},
	}
}
```

## Exact Hook Execution Order (`orm/hook_order_contract_test.go`)

- **Create**: `BeforeCreate` -> `BeforeSave` -> `Clean` -> `INSERT` -> `AfterCreate` -> `AfterSave`
- **Update**: `BeforeUpdate` -> `BeforeSave` -> `Clean` -> `UPDATE` -> `AfterUpdate` -> `AfterSave`
- **Delete**: `BeforeDelete` -> `DELETE` -> `AfterDelete`

Note: `schema.ModelHooks` has `BeforeCreate`, `AfterCreate`, `BeforeUpdate`, `AfterUpdate`, `BeforeSave`, `AfterSave`, `BeforeDelete`, `AfterDelete`, and `Clean`. There is no `Validate` field on `ModelHooks`; validation runs via `Clean`, generated `Validate()`, and `validate/`.
