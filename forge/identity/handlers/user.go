package handlers

import (
	"net/http"
	"strconv"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/identity/serializers"
	"github.com/forgego/forge/identity/service"
	forgehttp "github.com/forgego/forge/server"
)

// UserHandler handles user management endpoints
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// List handles GET /users/
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get pagination parameters
	page, pageSize, _ := api.ParsePaginationParams(r, 20)

	// Build filters
	filters := &service.UserFilters{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	// Get query parameters
	if email := r.URL.Query().Get("email"); email != "" {
		filters.Email = email
	}
	if username := r.URL.Query().Get("username"); username != "" {
		filters.Username = username
	}
	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = search
	}
	if isActive := r.URL.Query().Get("is_active"); isActive != "" {
		active := isActive == "true"
		filters.IsActive = &active
	}

	// Get users
	users, count, err := h.userService.ListUsers(ctx, filters)
	if err != nil {
		forgehttp.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Serialize users
	serialized := make([]map[string]interface{}, len(users))
	for i, user := range users {
		serialized[i] = serializers.FromUser(user).ToJSONMap()
	}

	// Send paginated response
	api.SendPaginatedResponse(w, r, serialized, int(count), page, pageSize)
}

// Create handles POST /users/
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var data map[string]interface{}
	if err := forgehttp.GetJSON(r, &data); err != nil {
		forgehttp.SendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Deserialize and validate
	serializer := serializers.NewUserSerializer()
	serializer.BaseSerializer = api.NewBaseSerializer(data)
	if err := serializer.Validate(); err != nil {
		forgehttp.SendJSON(w, http.StatusBadRequest, map[string]interface{}{
			"errors": serializer.Errors(),
		})
		return
	}

	// Create user request
	req := &service.CreateUserRequest{
		Username:  serializer.GetUsername(),
		Email:     serializer.GetEmail(),
		Password:  serializer.GetPassword(),
		FirstName: serializer.GetFirstName(),
		LastName:  serializer.GetLastName(),
		IsActive:  serializer.GetBool("is_active"),
		IsStaff:   serializer.GetBool("is_staff"),
	}

	// Create user
	user, err := h.userService.CreateUser(ctx, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrEmailExists || err == service.ErrUsernameExists {
			status = http.StatusConflict
		}
		forgehttp.SendError(w, status, err.Error())
		return
	}

	// Serialize response
	response := serializers.FromUser(user).ToJSONMap()
	forgehttp.SendJSON(w, http.StatusCreated, response)
}

// Retrieve handles GET /users/{id}/
func (h *UserHandler) Retrieve(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get ID from URL
	idStr := forgehttp.GetParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		forgehttp.SendError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Get user
	user, err := h.userService.GetUser(ctx, id)
	if err != nil {
		if err == service.ErrUserNotFound {
			forgehttp.SendError(w, http.StatusNotFound, "User not found")
		} else {
			forgehttp.SendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Serialize response
	response := serializers.FromUser(user).ToJSONMap()
	forgehttp.SendJSON(w, http.StatusOK, response)
}

// Update handles PUT /users/{id}/
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get ID from URL
	idStr := forgehttp.GetParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		forgehttp.SendError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Parse request
	var data map[string]interface{}
	if err := forgehttp.GetJSON(r, &data); err != nil {
		forgehttp.SendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Build update request
	req := &service.UpdateUserRequest{}
	if email, ok := data["email"].(string); ok {
		req.Email = &email
	}
	if firstName, ok := data["first_name"].(string); ok {
		req.FirstName = &firstName
	}
	if lastName, ok := data["last_name"].(string); ok {
		req.LastName = &lastName
	}
	if isActive, ok := data["is_active"].(bool); ok {
		req.IsActive = &isActive
	}
	if isStaff, ok := data["is_staff"].(bool); ok {
		req.IsStaff = &isStaff
	}
	if isLocked, ok := data["is_locked"].(bool); ok {
		req.IsLocked = &isLocked
	}

	// Update user
	user, err := h.userService.UpdateUser(ctx, id, req)
	if err != nil {
		if err == service.ErrUserNotFound {
			forgehttp.SendError(w, http.StatusNotFound, "User not found")
		} else if err == service.ErrEmailExists {
			forgehttp.SendError(w, http.StatusConflict, err.Error())
		} else {
			forgehttp.SendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Serialize response
	response := serializers.FromUser(user).ToJSONMap()
	forgehttp.SendJSON(w, http.StatusOK, response)
}

// Delete handles DELETE /users/{id}/
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get ID from URL
	idStr := forgehttp.GetParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		forgehttp.SendError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	// Delete user
	if err := h.userService.DeleteUser(ctx, id); err != nil {
		if err == service.ErrUserNotFound {
			forgehttp.SendError(w, http.StatusNotFound, "User not found")
		} else {
			forgehttp.SendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
