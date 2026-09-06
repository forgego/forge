package users

import (
	"context"

	"github.com/forgego/forge/admin"
)

// RegisterAdmin registers user models with the admin interface
func RegisterAdmin(ctx context.Context) error {
	// User admin
	_, err := admin.Register(&admin.Config[User]{
		ListDisplay: []admin.Field{
			UserFieldsInstance.Username,
			UserFieldsInstance.Email,
			UserFieldsInstance.FirstName,
			UserFieldsInstance.LastName,
			UserFieldsInstance.IsActive,
			UserFieldsInstance.IsStaff,
			UserFieldsInstance.DateJoined,
		},
		ListFilter: []admin.Field{
			UserFieldsInstance.IsActive,
			UserFieldsInstance.IsStaff,
			UserFieldsInstance.IsSuperuser,
		},
		SearchFields: []admin.Field{
			UserFieldsInstance.Username,
			UserFieldsInstance.Email,
			UserFieldsInstance.FirstName,
			UserFieldsInstance.LastName,
		},
		Ordering: []admin.Field{
			UserFieldsInstance.DateJoined,
			UserFieldsInstance.Username,
		},
	})
	if err != nil {
		return err
	}

	// Group admin
	_, err = admin.Register(&admin.Config[Group]{
		ListDisplay: []admin.Field{
			GroupFieldsInstance.Name,
			GroupFieldsInstance.Description,
			GroupFieldsInstance.CreatedAt,
		},
		SearchFields: []admin.Field{
			GroupFieldsInstance.Name,
			GroupFieldsInstance.Description,
		},
		Ordering: []admin.Field{
			GroupFieldsInstance.Name,
		},
	})
	return err
}
