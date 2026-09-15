package core

import (
	"context"
	"reflect"
)

func isSuperuser(user interface{}) bool {
	if user == nil {
		return false
	}
	switch u := user.(type) {
	case map[string]interface{}:
		if role, ok := u["role"].(string); ok && (role == "superuser" || role == "admin") {
			return true
		}
		if isSuper, ok := u["is_superuser"].(bool); ok && isSuper {
			return true
		}
	case map[string]string:
		if u["role"] == "superuser" || u["role"] == "admin" {
			return true
		}
	}

	val := reflect.ValueOf(user)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return false
		}
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		for _, fieldName := range []string{"IsSuperuser", "IsAdmin", "Role"} {
			f := val.FieldByName(fieldName)
			if f.IsValid() {
				if f.Kind() == reflect.Bool && f.Bool() {
					return true
				}
				if f.Kind() == reflect.String && (f.String() == "superuser" || f.String() == "admin") {
					return true
				}
			}
		}
	}
	return false
}

// Permission methods

func (a *Admin[T]) HasAddPermission(ctx context.Context, user interface{}) bool {
	if a.config.HasAddPermission != nil {
		return a.config.HasAddPermission(ctx, a, user)
	}
	if a.config.PermissionChecker != nil {
		return a.config.PermissionChecker.HasPermission(ctx, user, GetPermissionName(a.name, PermAdd))
	}
	if isSuperuser(user) {
		return true
	}
	return false // Default deny: configure Has*Permission or PermissionChecker to grant access
}

func (a *Admin[T]) checkObjectPermission(
	ctx context.Context,
	user interface{},
	obj interface{},
	hook func(context.Context, *Admin[T], interface{}, *T) bool,
	permType PermissionType,
) bool {
	var typedObj *T
	if obj != nil {
		var ok bool
		typedObj, ok = obj.(*T)
		if !ok {
			return false
		}
	}

	if hook != nil {
		return hook(ctx, a, user, typedObj)
	}
	if a.config.PermissionChecker != nil {
		return a.config.PermissionChecker.HasPermission(ctx, user, GetPermissionName(a.name, permType))
	}
	if isSuperuser(user) {
		return true
	}
	return false // Default deny: configure Has*Permission or PermissionChecker to grant access
}

func (a *Admin[T]) HasChangePermission(ctx context.Context, user interface{}, obj interface{}) bool {
	return a.checkObjectPermission(ctx, user, obj, a.config.HasChangePermission, PermChange)
}

func (a *Admin[T]) HasDeletePermission(ctx context.Context, user interface{}, obj interface{}) bool {
	return a.checkObjectPermission(ctx, user, obj, a.config.HasDeletePermission, PermDelete)
}

func (a *Admin[T]) HasViewPermission(ctx context.Context, user interface{}, obj interface{}) bool {
	return a.checkObjectPermission(ctx, user, obj, a.config.HasViewPermission, PermView)
}

func (a *Admin[T]) HasModulePermission(ctx context.Context, user interface{}) bool {
	if a.config.HasModulePermission != nil {
		return a.config.HasModulePermission(ctx, a, user)
	}
	if a.config.PermissionChecker != nil {
		return a.config.PermissionChecker.HasPermission(ctx, user, GetPermissionName(a.name, PermView))
	}
	if isSuperuser(user) {
		return true
	}
	return a.HasViewPermission(ctx, user, nil) ||
		a.HasAddPermission(ctx, user) ||
		a.HasChangePermission(ctx, user, nil) ||
		a.HasDeletePermission(ctx, user, nil)
}

func (a *Admin[T]) getPermissionsMetadata(ctx context.Context, user interface{}) PermissionMetadata {
	return PermissionMetadata{
		Add:    a.HasAddPermission(ctx, user),
		Change: a.HasChangePermission(ctx, user, nil),
		Delete: a.HasDeletePermission(ctx, user, nil),
		View:   a.HasViewPermission(ctx, user, nil),
	}
}
