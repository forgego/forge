package engagement

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/throttling"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers engagement API endpoints with enhanced features
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// RecentlyViewed API with enhanced features
	recentlyViewedViewSet := api.NewViewSet(&RecentlyViewed{})
	recentlyViewedViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			engagementUserLookup,
		),
	}
	recentlyViewedViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	recentlyViewedViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "1000/hour"},
		&throttling.AnonRateThrottle{Rate: "100/hour"},
	}

	// Register custom actions for RecentlyViewed
	recentlyViewedViewSet.RegisterAction("add", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/add",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleRecentlyViewedAdd,
	})

	recentlyViewedViewSet.RegisterAction("clear", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/clear",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleRecentlyViewedClear,
	})

	recentlyViewedViewSet.RegisterAction("recent", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/recent",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleRecentlyViewedRecent,
	})

	router.Register("recently-viewed", &api.ViewSetConfig{
		Model:           &RecentlyViewed{},
		Queryset:        RecentlyViewedObjects,
		Serializer:      &RecentlyViewedSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "customer_id", "product_id", "variant_id", "viewed_at", "source"},
		DetailFields:    []string{"id", "customer_id", "guest_id", "product_id", "variant_id", "viewed_at", "session_id", "user_agent", "ip_address", "source", "referer_url"},
		Filterable:      []string{"customer_id", "product_id", "source", "viewed_at"},
		Searchable:      []string{"guest_id", "session_id"},
		Ordering:        []string{"-viewed_at"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 1000, Window: time.Hour},
		EnhancedViewSet: recentlyViewedViewSet,
	})

	// ProductComparison API with enhanced features
	productComparisonViewSet := api.NewViewSet(&ProductComparison{})
	productComparisonViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			engagementUserLookup,
		),
	}
	productComparisonViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	productComparisonViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for ProductComparison
	productComparisonViewSet.RegisterAction("share", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/share",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleProductComparisonShare,
	})

	productComparisonViewSet.RegisterAction("duplicate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/duplicate",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleProductComparisonDuplicate,
	})

	productComparisonViewSet.RegisterAction("add-product", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/add-product",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleProductComparisonAddProduct,
	})

	productComparisonViewSet.RegisterAction("remove-product", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/remove-product",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleProductComparisonRemoveProduct,
	})

	router.Register("product-comparisons", &api.ViewSetConfig{
		Model:           &ProductComparison{},
		Queryset:        ProductComparisonObjects,
		Serializer:      &ProductComparisonSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "name", "customer_id", "product_ids", "is_public", "view_count", "created_at"},
		DetailFields:    []string{"id", "name", "customer_id", "guest_id", "product_ids", "is_public", "share_token", "view_count", "created_at", "updated_at"},
		Filterable:      []string{"customer_id", "is_public", "type"},
		Searchable:      []string{"name", "share_token"},
		Ordering:        []string{"-created_at"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: productComparisonViewSet,
	})

	// Notification API with enhanced features
	notificationViewSet := api.NewViewSet(&Notification{})
	notificationViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			engagementUserLookup,
		),
	}
	notificationViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
	}
	notificationViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "200/hour"},
	}

	// Register custom actions for Notification
	notificationViewSet.RegisterAction("mark-read", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/mark-read",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleNotificationMarkRead,
	})

	notificationViewSet.RegisterAction("mark-all-read", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/mark-all-read",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleNotificationMarkAllRead,
	})

	notificationViewSet.RegisterAction("send", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/send",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleNotificationSend,
	})

	router.Register("notifications", &api.ViewSetConfig{
		Model:           &Notification{},
		Queryset:        NotificationObjects,
		Serializer:      &NotificationSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "customer_id", "type", "title", "is_read", "created_at"},
		DetailFields:    []string{"id", "customer_id", "type", "title", "message", "action_url", "action_text", "related_type", "related_id", "is_read", "read_at", "push_enabled", "email_enabled", "sms_enabled", "scheduled_for", "sent_at", "created_at"},
		Filterable:      []string{"customer_id", "type", "is_read", "scheduled_for"},
		Searchable:      []string{"title", "message"},
		Ordering:        []string{"-created_at"},
		PerPage:         30,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}},
		RateLimit:       &api.RateLimit{Requests: 200, Window: time.Hour},
		EnhancedViewSet: notificationViewSet,
	})

	// CustomerActivity API with enhanced features
	customerActivityViewSet := api.NewViewSet(&CustomerActivity{})
	customerActivityViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			engagementUserLookup,
		),
	}
	customerActivityViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	customerActivityViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "1000/hour"},
	}

	// Register custom actions for CustomerActivity
	customerActivityViewSet.RegisterAction("log", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/log",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleCustomerActivityLog,
	})

	customerActivityViewSet.RegisterAction("export", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/export",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleCustomerActivityExport,
	})

	customerActivityViewSet.RegisterAction("by-customer", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/by-customer",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleCustomerActivityByCustomer,
	})

	router.Register("customer-activities", &api.ViewSetConfig{
		Model:           &CustomerActivity{},
		Queryset:        CustomerActivityObjects,
		Serializer:      &CustomerActivitySerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "customer_id", "activity_type", "entity_type", "timestamp"},
		DetailFields:    []string{"id", "customer_id", "activity_type", "entity_type", "entity_id", "data", "session_id", "user_agent", "ip_address", "timestamp"},
		Filterable:      []string{"customer_id", "activity_type", "entity_type", "timestamp"},
		Searchable:      []string{"session_id"},
		Ordering:        []string{"-timestamp"},
		PerPage:         50,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 1000, Window: time.Hour},
		EnhancedViewSet: customerActivityViewSet,
	})

	// AbandonedCartReminder API with enhanced features
	abandonedCartReminderViewSet := api.NewViewSet(&AbandonedCartReminder{})
	abandonedCartReminderViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			engagementUserLookup,
		),
	}
	abandonedCartReminderViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	abandonedCartReminderViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "100/hour"},
	}

	// Register custom actions for AbandonedCartReminder
	abandonedCartReminderViewSet.RegisterAction("send", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/send",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleAbandonedCartReminderSend,
	})

	abandonedCartReminderViewSet.RegisterAction("schedule", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/schedule",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleAbandonedCartReminderSchedule,
	})

	abandonedCartReminderViewSet.RegisterAction("recover", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/recover",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleAbandonedCartReminderRecover,
	})

	router.Register("abandoned-cart-reminders", &api.ViewSetConfig{
		Model:           &AbandonedCartReminder{},
		Queryset:        AbandonedCartReminderObjects,
		Serializer:      &AbandonedCartReminderSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "cart_id", "customer_id", "reminder_number", "status", "reminder_sent_at", "recovered_at"},
		DetailFields:    []string{"id", "cart_id", "customer_id", "guest_email", "reminder_number", "reminder_sent_at", "reminder_opened_at", "reminder_clicked_at", "recovered_at", "recovered_order_id", "status", "created_at"},
		Filterable:      []string{"customer_id", "status", "reminder_number"},
		Searchable:      []string{"guest_email"},
		Ordering:        []string{"-created_at"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 100, Window: time.Hour},
		EnhancedViewSet: abandonedCartReminderViewSet,
	})

	// UserSegment API with enhanced features
	userSegmentViewSet := api.NewViewSet(&UserSegment{})
	userSegmentViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			engagementUserLookup,
		),
	}
	userSegmentViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	userSegmentViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "100/hour"},
	}

	// Register custom actions for UserSegment
	userSegmentViewSet.RegisterAction("evaluate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/evaluate",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleUserSegmentEvaluate,
	})

	userSegmentViewSet.RegisterAction("sync", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/sync",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleUserSegmentSync,
	})

	userSegmentViewSet.RegisterAction("export", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      true,
		URLPath:     "/export",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleUserSegmentExport,
	})

	router.Register("user-segments", &api.ViewSetConfig{
		Model:           &UserSegment{},
		Queryset:        UserSegmentObjects,
		Serializer:      &UserSegmentSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "name", "type", "customer_count", "is_active", "created_at"},
		DetailFields:    []string{"id", "name", "description", "type", "criteria", "rules", "customer_ids", "customer_count", "is_active", "created_at", "updated_at"},
		Filterable:      []string{"type", "is_active"},
		Searchable:      []string{"name", "description"},
		Ordering:        []string{"name"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 100, Window: time.Hour},
		EnhancedViewSet: userSegmentViewSet,
	})
}

// engagementUserLookup looks up a user from JWT claims
func engagementUserLookup(claims authentication.JWTClaims) (interface{}, error) {
	userID, ok := claims["subject"].(string)
	if !ok {
		if subject, ok := claims["subject"].(float64); ok {
			userID = strconv.FormatFloat(subject, 'f', -1, 64)
		} else {
			return nil, nil
		}
	}
	return nil, nil
}

// Custom action handlers for RecentlyViewed

func handleRecentlyViewedAdd(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID string `json:"customer_id"`
		GuestID    string `json:"guest_id"`
		ProductID  string `json:"product_id"`
		VariantID  string `json:"variant_id"`
		Source     string `json:"source"`
		SessionID  string `json:"session_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Product added to recently viewed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleRecentlyViewedClear(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "success",
		"message": "Recently viewed history cleared",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleRecentlyViewedRecent(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer_id")
	guestID := r.URL.Query().Get("guest_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	response := map[string]interface{}{
		"customer_id": customerID,
		"guest_id":    guestID,
		"limit":       limit,
		"items":       []interface{}{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for ProductComparison

func handleProductComparisonShare(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	comparisonID := vars["id"]

	response := map[string]interface{}{
		"comparison_id": comparisonID,
		"share_url":     "/compare/" + comparisonID + "?token=shared",
		"share_token":   "shared_" + comparisonID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleProductComparisonDuplicate(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	comparisonID := vars["id"]

	response := map[string]interface{}{
		"original_id":  comparisonID,
		"duplicate_id": "new_" + comparisonID,
		"status":      "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleProductComparisonAddProduct(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	comparisonID := vars["id"]

	response := map[string]interface{}{
		"comparison_id": comparisonID,
		"status":       "success",
		"message":      "Product added to comparison",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleProductComparisonRemoveProduct(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	comparisonID := vars["id"]

	response := map[string]interface{}{
		"comparison_id": comparisonID,
		"status":        "success",
		"message":       "Product removed from comparison",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for Notification

func handleNotificationMarkRead(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	notificationID := vars["id"]

	response := map[string]interface{}{
		"notification_id": notificationID,
		"is_read":         true,
		"read_at":         time.Now(),
		"status":          "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleNotificationMarkAllRead(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":         "success",
		"mark_all_read":  true,
		"updated_count":  0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleNotificationSend(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "success",
		"message": "Notification sent",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for CustomerActivity

func handleCustomerActivityLog(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID   string                 `json:"customer_id"`
		ActivityType string                 `json:"activity_type"`
		EntityType   string                 `json:"entity_type"`
		EntityID     string                 `json:"entity_id"`
		Data         map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"status":        "success",
		"activity_id":   "new_activity_id",
		"activity_type": req.ActivityType,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCustomerActivityExport(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer_id")
	format := r.URL.Query().Get("format")

	response := map[string]interface{}{
		"customer_id": customerID,
		"format":      format,
		"url":         "/exports/activities_" + customerID + "." + format,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCustomerActivityByCustomer(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	response := map[string]interface{}{
		"customer_id": customerID,
		"limit":        limit,
		"offset":       offset,
		"activities":  []interface{}{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for AbandonedCartReminder

func handleAbandonedCartReminderSend(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	reminderID := vars["id"]

	response := map[string]interface{}{
		"reminder_id": reminderID,
		"status":      "sent",
		"sent_at":     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleAbandonedCartReminderSchedule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CartID    string `json:"cart_id"`
		SendAt    string `json:"send_at"`
		Template  string `json:"template"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"cart_id":      req.CartID,
		"scheduled_at": req.SendAt,
		"template":     req.Template,
		"reminder_id":  "new_reminder_id",
		"status":       "scheduled",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleAbandonedCartReminderRecover(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	reminderID := vars["id"]

	response := map[string]interface{}{
		"reminder_id": reminderID,
		"status":      "recovered",
		"recovered":  true,
		"recovered_at": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for UserSegment

func handleUserSegmentEvaluate(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	segmentID := vars["id"]

	response := map[string]interface{}{
		"segment_id":      segmentID,
		"status":          "evaluated",
		"customer_count":  0,
		"evaluated_at":    time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleUserSegmentSync(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	segmentID := vars["id"]

	response := map[string]interface{}{
		"segment_id":      segmentID,
		"status":          "synced",
		"synced_count":    0,
		"synced_at":       time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleUserSegmentExport(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	segmentID := vars["id"]
	format := r.URL.Query().Get("format")

	response := map[string]interface{}{
		"segment_id": segmentID,
		"format":     format,
		"url":        "/exports/segment_" + segmentID + "." + format,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/throttling"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers engagement API endpoints with enhanced features
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// RecentlyViewed API with enhanced features
	recentlyViewedViewSet := api.NewViewSet(&RecentlyViewed{})
	recentlyViewedViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			engagementUserLookup,
		),
	}
	recentlyViewedViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	recentlyViewedViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "1000/hour"},
		&throttling.AnonRateThrottle{Rate: "100/hour"},
	}

	// Register custom actions for RecentlyViewed
	recentlyViewedViewSet.RegisterAction("add", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/add",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleRecentlyViewedAdd,
	})

	recentlyViewedViewSet.RegisterAction("clear", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/clear",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleRecentlyViewedClear,
	})

	recentlyViewedViewSet.RegisterAction("recent", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/recent",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleRecentlyViewedRecent,
	})

	router.Register("recently-viewed", &api.ViewSetConfig{
		Model:           &RecentlyViewed{},
		Queryset:        RecentlyViewedObjects,
		Serializer:      &RecentlyViewedSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "customer_id", "product_id", "variant_id", "viewed_at", "source"},
		DetailFields:    []string{"id", "customer_id", "guest_id", "product_id", "variant_id", "viewed_at", "session_id", "user_agent", "ip_address", "source", "referer_url"},
		Filterable:      []string{"customer_id", "product_id", "source", "viewed_at"},
		Searchable:      []string{"guest_id", "session_id"},
		Ordering:        []string{"-viewed_at"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 1000, Window: time.Hour},
		EnhancedViewSet: recentlyViewedViewSet,
	})

	// ProductComparison API with enhanced features
	productComparisonViewSet := api.NewViewSet(&ProductComparison{})
	productComparisonViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			engagementUserLookup,
		),
	}
	productComparisonViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	productComparisonViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for ProductComparison
	productComparisonViewSet.RegisterAction("share", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/share",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleProductComparisonShare,
	})

	productComparisonViewSet.RegisterAction("duplicate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/duplicate",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleProductComparisonDuplicate,
	})

	productComparisonViewSet.RegisterAction("add-product", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/add-product",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleProductComparisonAddProduct,
	})

	productComparisonViewSet.RegisterAction("remove-product", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/remove-product",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleProductComparisonRemoveProduct,
	})

	router.Register("product-comparisons", &api.ViewSetConfig{
		Model:           &ProductComparison{},
		Queryset:        ProductComparisonObjects,
		Serializer:      &ProductComparisonSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "name", "customer_id", "product_ids", "is_public", "view_count", "created_at"},
		DetailFields:    []string{"id", "name", "customer_id", "guest_id", "product_ids", "is_public", "share_token", "view_count", "created_at", "updated_at"},
		Filterable:      []string{"customer_id", "is_public", "type"},
		Searchable:      []string{"name", "share_token"},
		Ordering:        []string{"-created_at"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: productComparisonViewSet,
	})

	// Notification API with enhanced features
	notificationViewSet := api.NewViewSet(&Notification{})
	notificationViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			engagementUserLookup,
		),
	}
	notificationViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
	}
	notificationViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "200/hour"},
	}

	// Register custom actions for Notification
	notificationViewSet.RegisterAction("mark-read", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/mark-read",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleNotificationMarkRead,
	})

	notificationViewSet.RegisterAction("mark-all-read", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/mark-all-read",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleNotificationMarkAllRead,
	})

	notificationViewSet.RegisterAction("send", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/send",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleNotificationSend,
	})

	router.Register("notifications", &api.ViewSetConfig{
		Model:           &Notification{},
		Queryset:        NotificationObjects,
		Serializer:      &NotificationSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "customer_id", "type", "title", "is_read", "created_at"},
		DetailFields:    []string{"id", "customer_id", "type", "title", "message", "action_url", "action_text", "related_type", "related_id", "is_read", "read_at", "push_enabled", "email_enabled", "sms_enabled", "scheduled_for", "sent_at", "created_at"},
		Filterable:      []string{"customer_id", "type", "is_read", "scheduled_for"},
		Searchable:      []string{"title", "message"},
		Ordering:        []string{"-created_at"},
		PerPage:         30,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}},
		RateLimit:       &api.RateLimit{Requests: 200, Window: time.Hour},
		EnhancedViewSet: notificationViewSet,
	})

	// CustomerActivity API with enhanced features
	customerActivityViewSet := api.NewViewSet(&CustomerActivity{})
	customerActivityViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			engagementUserLookup,
		),
	}
	customerActivityViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	customerActivityViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "1000/hour"},
	}

	// Register custom actions for CustomerActivity
	customerActivityViewSet.RegisterAction("log", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/log",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleCustomerActivityLog,
	})

	customerActivityViewSet.RegisterAction("export", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/export",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleCustomerActivityExport,
	})

	customerActivityViewSet.RegisterAction("by-customer", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/by-customer",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleCustomerActivityByCustomer,
	})

	router.Register("customer-activities", &api.ViewSetConfig{
		Model:           &CustomerActivity{},
		Queryset:        CustomerActivityObjects,
		Serializer:      &CustomerActivitySerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "customer_id", "activity_type", "entity_type", "timestamp"},
		DetailFields:    []string{"id", "customer_id", "activity_type", "entity_type", "entity_id", "data", "session_id", "user_agent", "ip_address", "timestamp"},
		Filterable:      []string{"customer_id", "activity_type", "entity_type", "timestamp"},
		Searchable:      []string{"session_id"},
		Ordering:        []string{"-timestamp"},
		PerPage:         50,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 1000, Window: time.Hour},
		EnhancedViewSet: customerActivityViewSet,
	})

	// AbandonedCartReminder API with enhanced features
	abandonedCartReminderViewSet := api.NewViewSet(&AbandonedCartReminder{})
	abandonedCartReminderViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			engagementUserLookup,
		),
	}
	abandonedCartReminderViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	abandonedCartReminderViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "100/hour"},
	}

	// Register custom actions for AbandonedCartReminder
	abandonedCartReminderViewSet.RegisterAction("send", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/send",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleAbandonedCartReminderSend,
	})

	abandonedCartReminderViewSet.RegisterAction("schedule", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/schedule",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleAbandonedCartReminderSchedule,
	})

	abandonedCartReminderViewSet.RegisterAction("recover", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/recover",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleAbandonedCartReminderRecover,
	})

	router.Register("abandoned-cart-reminders", &api.ViewSetConfig{
		Model:           &AbandonedCartReminder{},
		Queryset:        AbandonedCartReminderObjects,
		Serializer:      &AbandonedCartReminderSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "cart_id", "customer_id", "reminder_number", "status", "reminder_sent_at", "recovered_at"},
		DetailFields:    []string{"id", "cart_id", "customer_id", "guest_email", "reminder_number", "reminder_sent_at", "reminder_opened_at", "reminder_clicked_at", "recovered_at", "recovered_order_id", "status", "created_at"},
		Filterable:      []string{"customer_id", "status", "reminder_number"},
		Searchable:      []string{"guest_email"},
		Ordering:        []string{"-created_at"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 100, Window: time.Hour},
		EnhancedViewSet: abandonedCartReminderViewSet,
	})

	// UserSegment API with enhanced features
	userSegmentViewSet := api.NewViewSet(&UserSegment{})
	userSegmentViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			engagementUserLookup,
		),
	}
	userSegmentViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	userSegmentViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "100/hour"},
	}

	// Register custom actions for UserSegment
	userSegmentViewSet.RegisterAction("evaluate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/evaluate",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleUserSegmentEvaluate,
	})

	userSegmentViewSet.RegisterAction("sync", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/sync",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleUserSegmentSync,
	})

	userSegmentViewSet.RegisterAction("export", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      true,
		URLPath:     "/export",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleUserSegmentExport,
	})

	router.Register("user-segments", &api.ViewSetConfig{
		Model:           &UserSegment{},
		Queryset:        UserSegmentObjects,
		Serializer:      &UserSegmentSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "name", "type", "customer_count", "is_active", "created_at"},
		DetailFields:    []string{"id", "name", "description", "type", "criteria", "rules", "customer_ids", "customer_count", "is_active", "created_at", "updated_at"},
		Filterable:      []string{"type", "is_active"},
		Searchable:      []string{"name", "description"},
		Ordering:        []string{"name"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 100, Window: time.Hour},
		EnhancedViewSet: userSegmentViewSet,
	})
}

// engagementUserLookup looks up a user from JWT claims
func engagementUserLookup(claims authentication.JWTClaims) (interface{}, error) {
	userID, ok := claims["subject"].(string)
	if !ok {
		if subject, ok := claims["subject"].(float64); ok {
			userID = strconv.FormatFloat(subject, 'f', -1, 64)
		} else {
			return nil, nil
		}
	}
	return nil, nil
}

// Custom action handlers for RecentlyViewed

func handleRecentlyViewedAdd(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID string `json:"customer_id"`
		GuestID    string `json:"guest_id"`
		ProductID  string `json:"product_id"`
		VariantID  string `json:"variant_id"`
		Source     string `json:"source"`
		SessionID  string `json:"session_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"status":  "success",
		"message": "Product added to recently viewed",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleRecentlyViewedClear(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "success",
		"message": "Recently viewed history cleared",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleRecentlyViewedRecent(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer_id")
	guestID := r.URL.Query().Get("guest_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	response := map[string]interface{}{
		"customer_id": customerID,
		"guest_id":    guestID,
		"limit":       limit,
		"items":       []interface{}{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for ProductComparison

func handleProductComparisonShare(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	comparisonID := vars["id"]

	response := map[string]interface{}{
		"comparison_id": comparisonID,
		"share_url":     "/compare/" + comparisonID + "?token=shared",
		"share_token":   "shared_" + comparisonID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleProductComparisonDuplicate(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	comparisonID := vars["id"]

	response := map[string]interface{}{
		"original_id":  comparisonID,
		"duplicate_id": "new_" + comparisonID,
		"status":      "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleProductComparisonAddProduct(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	comparisonID := vars["id"]

	response := map[string]interface{}{
		"comparison_id": comparisonID,
		"status":       "success",
		"message":      "Product added to comparison",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleProductComparisonRemoveProduct(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	comparisonID := vars["id"]

	response := map[string]interface{}{
		"comparison_id": comparisonID,
		"status":        "success",
		"message":       "Product removed from comparison",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for Notification

func handleNotificationMarkRead(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	notificationID := vars["id"]

	response := map[string]interface{}{
		"notification_id": notificationID,
		"is_read":         true,
		"read_at":         time.Now(),
		"status":          "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleNotificationMarkAllRead(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":         "success",
		"mark_all_read":  true,
		"updated_count":  0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleNotificationSend(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "success",
		"message": "Notification sent",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for CustomerActivity

func handleCustomerActivityLog(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID   string                 `json:"customer_id"`
		ActivityType string                 `json:"activity_type"`
		EntityType   string                 `json:"entity_type"`
		EntityID     string                 `json:"entity_id"`
		Data         map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"status":        "success",
		"activity_id":   "new_activity_id",
		"activity_type": req.ActivityType,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCustomerActivityExport(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer_id")
	format := r.URL.Query().Get("format")

	response := map[string]interface{}{
		"customer_id": customerID,
		"format":      format,
		"url":         "/exports/activities_" + customerID + "." + format,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleCustomerActivityByCustomer(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("customer_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	response := map[string]interface{}{
		"customer_id": customerID,
		"limit":        limit,
		"offset":       offset,
		"activities":  []interface{}{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for AbandonedCartReminder

func handleAbandonedCartReminderSend(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	reminderID := vars["id"]

	response := map[string]interface{}{
		"reminder_id": reminderID,
		"status":      "sent",
		"sent_at":     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleAbandonedCartReminderSchedule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CartID    string `json:"cart_id"`
		SendAt    string `json:"send_at"`
		Template  string `json:"template"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"cart_id":      req.CartID,
		"scheduled_at": req.SendAt,
		"template":     req.Template,
		"reminder_id":  "new_reminder_id",
		"status":       "scheduled",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleAbandonedCartReminderRecover(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	reminderID := vars["id"]

	response := map[string]interface{}{
		"reminder_id": reminderID,
		"status":      "recovered",
		"recovered":  true,
		"recovered_at": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for UserSegment

func handleUserSegmentEvaluate(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	segmentID := vars["id"]

	response := map[string]interface{}{
		"segment_id":      segmentID,
		"status":          "evaluated",
		"customer_count":  0,
		"evaluated_at":    time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleUserSegmentSync(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	segmentID := vars["id"]

	response := map[string]interface{}{
		"segment_id":      segmentID,
		"status":          "synced",
		"synced_count":    0,
		"synced_at":       time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleUserSegmentExport(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	segmentID := vars["id"]
	format := r.URL.Query().Get("format")

	response := map[string]interface{}{
		"segment_id": segmentID,
		"format":     format,
		"url":        "/exports/segment_" + segmentID + "." + format,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

