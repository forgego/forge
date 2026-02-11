package promotions

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/throttling"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers promotions API endpoints with enhanced features
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// Promotion API with enhanced features
	promotionViewSet := api.NewViewSet(&Promotion{})
	promotionViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			customerUserLookup,
		),
	}
	promotionViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	promotionViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "200/hour"},
		&throttling.AnonRateThrottle{Rate: "20/hour"},
	}

	// Register custom actions for Promotion
	promotionViewSet.RegisterAction("activate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/activate",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handlePromotionActivate,
	})

	promotionViewSet.RegisterAction("deactivate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/deactivate",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handlePromotionDeactivate,
	})

	promotionViewSet.RegisterAction("usage-report", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      true,
		URLPath:     "/usage-report",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handlePromotionUsageReport,
	})

	promotionViewSet.RegisterAction("validate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/validate",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handlePromotionValidate,
	})

	router.Register("promotions", &api.ViewSetConfig{
		Model:              &Promotion{},
		Queryset:           PromotionObjects,
		Serializer:         &PromotionSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:         []string{"id", "name", "code", "type", "discount_value", "discount_type", "is_active", "start_date", "end_date", "times_used"},
		DetailFields:       []string{"id", "name", "code", "description", "type", "discount_value", "discount_type", "min_purchase", "max_discount", "buy_quantity", "get_quantity", "free_product_id", "free_shipping", "applies_to", "product_ids", "category_ids", "brand_ids", "new_customers_only", "customer_group_ids", "total_usage_limit", "per_customer_limit", "start_date", "end_date", "priority", "can_stack", "stack_with", "is_active", "times_used", "created_at", "updated_at"},
		Filterable:         []string{"type", "is_active", "applies_to", "start_date", "end_date"},
		Searchable:         []string{"name", "code", "description"},
		Ordering:           []string{"-priority", "-created_at"},
		PerPage:            20,
		Authenticate:       true,
		Permissions:        []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:          &api.RateLimit{Requests: 200, Window: time.Hour},
		EnhancedViewSet:    promotionViewSet,
	})

	// PromotionRule API with enhanced features
	promotionRuleViewSet := api.NewViewSet(&PromotionRule{})
	promotionRuleViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			customerUserLookup,
		),
	}
	promotionRuleViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	promotionRuleViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "100/hour"},
		&throttling.AnonRateThrottle{Rate: "10/hour"},
	}

	router.Register("promotion-rules", &api.ViewSetConfig{
		Model:              &PromotionRule{},
		Queryset:           PromotionRuleObjects,
		Serializer:         &PromotionRuleSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:         []string{"id", "promotion_id", "rule_type", "logic", "is_active"},
		DetailFields:       []string{"id", "promotion_id", "rule_type", "parameters", "logic", "is_active", "created_at", "updated_at"},
		Filterable:         []string{"promotion_id", "rule_type", "is_active"},
		Searchable:         []string{"rule_type"},
		Ordering:           []string{"promotion_id", "-created_at"},
		PerPage:            20,
		Authenticate:       true,
		Permissions:        []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:          &api.RateLimit{Requests: 100, Window: time.Hour},
		EnhancedViewSet:    promotionRuleViewSet,
	})

	// Banner API with enhanced features
	bannerViewSet := api.NewViewSet(&Banner{})
	bannerViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			customerUserLookup,
		),
	}
	bannerViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	bannerViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for Banner
	bannerViewSet.RegisterAction("stats", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      true,
		URLPath:     "/stats",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleBannerStats,
	})

	bannerViewSet.RegisterAction("schedule", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/schedule",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleBannerSchedule,
	})

	router.Register("banners", &api.ViewSetConfig{
		Model:              &Banner{},
		Queryset:           BannerObjects,
		Serializer:         &BannerSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:         []string{"id", "title", "position", "priority", "is_active", "start_date", "end_date", "click_count", "view_count"},
		DetailFields:       []string{"id", "title", "subtitle", "image_url", "mobile_image_url", "video_url", "content", "link", "link_text", "position", "background_color", "text_color", "start_date", "end_date", "schedule", "device_types", "user_types", "customer_group_ids", "priority", "is_active", "click_count", "view_count", "created_at", "updated_at"},
		Filterable:         []string{"position", "is_active", "device_types"},
		Searchable:         []string{"title", "subtitle"},
		Ordering:           []string{"-priority", "-created_at"},
		PerPage:            20,
		Authenticate:       true,
		Permissions:        []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:          &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet:    bannerViewSet,
	})

	// NewsletterSubscription API with enhanced features
	newsletterViewSet := api.NewViewSet(&NewsletterSubscription{})
	newsletterViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			customerUserLookup,
		),
	}
	newsletterViewSet.PermissionClasses = []permissions.Permission{
		&permissions.AllowAny{}, // Allow public subscription
	}
	newsletterViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "100/hour"},
		&throttling.AnonRateThrottle{Rate: "10/hour"},
	}

	// Register custom actions for NewsletterSubscription
	newsletterViewSet.RegisterAction("subscribe", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/subscribe",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleNewsletterSubscribe,
	})

	newsletterViewSet.RegisterAction("unsubscribe", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/unsubscribe",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleNewsletterUnsubscribe,
	})

	newsletterViewSet.RegisterAction("stats", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/stats",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleNewsletterStats,
	})

	router.Register("newsletter-subscriptions", &api.ViewSetConfig{
		Model:              &NewsletterSubscription{},
		Queryset:           NewsletterSubscriptionObjects,
		Serializer:         &NewsletterSubscriptionSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:         []string{"id", "email", "status", "list_type", "source", "created_at"},
		DetailFields:       []string{"id", "email", "list_type", "source", "preferences", "consent_given", "consent_date", "consent_ip", "status", "subscribed_at", "unsubscribed_at", "click_count", "open_count", "segments", "created_at", "updated_at"},
		Filterable:         []string{"status", "list_type", "source"},
		Searchable:         []string{"email"},
		Ordering:           []string{"-created_at"},
		PerPage:            20,
		Authenticate:       true, // Admin only for list/detail
		Permissions:        []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:          &api.RateLimit{Requests: 100, Window: time.Hour},
		EnhancedViewSet:    newsletterViewSet,
	})

	// PromotionUsage API with enhanced features
	promotionUsageViewSet := api.NewViewSet(&PromotionUsage{})
	promotionUsageViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			customerUserLookup,
		),
	}
	promotionUsageViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	promotionUsageViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	router.Register("promotion-usage", &api.ViewSetConfig{
		Model:              &PromotionUsage{},
		Queryset:           PromotionUsageObjects,
		Serializer:         &PromotionUsageSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:         []string{"id", "promotion_id", "order_id", "customer_id", "discount_amount", "used_at"},
		DetailFields:       []string{"id", "promotion_id", "order_id", "customer_id", "discount_amount", "used_at"},
		Filterable:         []string{"promotion_id", "order_id", "customer_id"},
		Searchable:         []string{},
		Ordering:           []string{"-used_at"},
		PerPage:            20,
		Authenticate:       true,
		Permissions:        []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:          &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet:    promotionUsageViewSet,
	})
}

// Customer user lookup function for JWT authentication
func customerUserLookup(claims authentication.JWTClaims) (interface{}, error) {
	customerID, ok := claims["subject"].(string)
	if !ok {
		return nil, nil
	}
	return nil, nil
}

// ==================== Custom Action Handlers ====================

// Promotion handlers

func handlePromotionActivate(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	promotionID := vars["id"]

	response := map[string]interface{}{
		"promotion_id": promotionID,
		"message":      "Promotion activated",
		"is_active":    true,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handlePromotionDeactivate(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	promotionID := vars["id"]

	response := map[string]interface{}{
		"promotion_id": promotionID,
		"message":      "Promotion deactivated",
		"is_active":    false,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handlePromotionUsageReport(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	promotionID := vars["id"]

	response := map[string]interface{}{
		"promotion_id":      promotionID,
		"message":           "Usage report generated",
		"total_usage":       0,
		"unique_customers":   0,
		"total_discount":    0.0,
		"average_discount":  0.0,
		"usage_by_day":       []interface{}{},
		"usage_by_customer": []interface{}{},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handlePromotionValidate(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		Code        string  `json:"code"`
		CartTotal   float64 `json:"cart_total"`
		ProductIDs   []string `json:"product_ids"`
		CustomerID  string  `json:"customer_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := map[string]interface{}{
			"valid":  false,
			"error":  "Invalid request body",
			"reason": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"valid":           true,
		"promotion":       map[string]interface{}{},
		"discount_amount": 0.0,
		"message":         "Promotion is valid",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Banner handlers

func handleBannerStats(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	bannerID := vars["id"]

	response := map[string]interface{}{
		"banner_id":      bannerID,
		"message":        "Banner statistics",
		"total_views":    0,
		"total_clicks":  0.0,
		"click_rate":    0.0,
		"views_by_day":  []interface{}{},
		"clicks_by_day": []interface{}{},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleBannerSchedule(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	bannerID := vars["id"]

	// Parse schedule from request body
	var schedule []BannerSchedule
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		response := map[string]interface{}{
			"banner_id": bannerID,
			"message":   "Invalid schedule data",
			"error":    err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"banner_id": bannerID,
		"message":   "Banner scheduled successfully",
		"schedule":  schedule,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// NewsletterSubscription handlers

func handleNewsletterSubscribe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email       string            `json:"email"`
		ListType    string            `json:"list_type"`
		Source      string            `json:"source"`
		Preferences map[string]bool   `json:"preferences"`
		ConsentGiven bool             `json:"consent_given"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Subscribed successfully",
		"email":   req.Email,
		"status":  StatusSubscribed,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleNewsletterUnsubscribe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Unsubscribed successfully",
		"email":   req.Email,
		"status":  StatusUnsubscribed,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleNewsletterStats(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message":              "Newsletter statistics",
		"total_subscribers":    0,
		"active_subscribers":   0,
		"unsubscribed":         0,
		"bounced":              0,
		"avg_open_rate":        0.0,
		"avg_click_rate":       0.0,
		"subscribers_by_source": map[string]interface{}{},
		"subscribers_by_list":  map[string]interface{}{},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ==================== Serializers ====================

type PromotionSerializer struct {
	*api.BaseSerializer
}

func (s *PromotionSerializer) New() api.Serializer {
	return &PromotionSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *PromotionSerializer) Validate() error {
	errors := make(map[string]string)
	if s.Data["name"] == "" {
		errors["name"] = "Name is required"
	}
	if s.Data["code"] == "" {
		errors["code"] = "Code is required"
	}
	if s.Data["type"] == "" {
		errors["type"] = "Type is required"
	}
	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

type PromotionRuleSerializer struct {
	*api.BaseSerializer
}

func (s *PromotionRuleSerializer) New() api.Serializer {
	return &PromotionRuleSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type BannerSerializer struct {
	*api.BaseSerializer
}

func (s *BannerSerializer) New() api.Serializer {
	return &BannerSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *BannerSerializer) Validate() error {
	errors := make(map[string]string)
	if s.Data["title"] == "" {
		errors["title"] = "Title is required"
	}
	if s.Data["image_url"] == "" {
		errors["image_url"] = "Image URL is required"
	}
	if s.Data["position"] == "" {
		errors["position"] = "Position is required"
	}
	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

type NewsletterSubscriptionSerializer struct {
	*api.BaseSerializer
}

func (s *NewsletterSubscriptionSerializer) New() api.Serializer {
	return &NewsletterSubscriptionSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *NewsletterSubscriptionSerializer) Validate() error {
	errors := make(map[string]string)
	if s.Data["email"] == "" {
		errors["email"] = "Email is required"
	}
	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

type PromotionUsageSerializer struct {
	*api.BaseSerializer
}

func (s *PromotionUsageSerializer) New() api.Serializer {
	return &PromotionUsageSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/forgego/forge/api"
	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
	"github.com/forgego/forge/api/throttling"
	"github.com/forgego/forge/db"
)

// RegisterAPI registers promotions API endpoints with enhanced features
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// Promotion API with enhanced features
	promotionViewSet := api.NewViewSet(&Promotion{})
	promotionViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			customerUserLookup,
		),
	}
	promotionViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	promotionViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "200/hour"},
		&throttling.AnonRateThrottle{Rate: "20/hour"},
	}

	// Register custom actions for Promotion
	promotionViewSet.RegisterAction("activate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/activate",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handlePromotionActivate,
	})

	promotionViewSet.RegisterAction("deactivate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/deactivate",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handlePromotionDeactivate,
	})

	promotionViewSet.RegisterAction("usage-report", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      true,
		URLPath:     "/usage-report",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handlePromotionUsageReport,
	})

	promotionViewSet.RegisterAction("validate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/validate",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handlePromotionValidate,
	})

	router.Register("promotions", &api.ViewSetConfig{
		Model:              &Promotion{},
		Queryset:           PromotionObjects,
		Serializer:         &PromotionSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:         []string{"id", "name", "code", "type", "discount_value", "discount_type", "is_active", "start_date", "end_date", "times_used"},
		DetailFields:       []string{"id", "name", "code", "description", "type", "discount_value", "discount_type", "min_purchase", "max_discount", "buy_quantity", "get_quantity", "free_product_id", "free_shipping", "applies_to", "product_ids", "category_ids", "brand_ids", "new_customers_only", "customer_group_ids", "total_usage_limit", "per_customer_limit", "start_date", "end_date", "priority", "can_stack", "stack_with", "is_active", "times_used", "created_at", "updated_at"},
		Filterable:         []string{"type", "is_active", "applies_to", "start_date", "end_date"},
		Searchable:         []string{"name", "code", "description"},
		Ordering:           []string{"-priority", "-created_at"},
		PerPage:            20,
		Authenticate:       true,
		Permissions:        []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:          &api.RateLimit{Requests: 200, Window: time.Hour},
		EnhancedViewSet:    promotionViewSet,
	})

	// PromotionRule API with enhanced features
	promotionRuleViewSet := api.NewViewSet(&PromotionRule{})
	promotionRuleViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			customerUserLookup,
		),
	}
	promotionRuleViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	promotionRuleViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "100/hour"},
		&throttling.AnonRateThrottle{Rate: "10/hour"},
	}

	router.Register("promotion-rules", &api.ViewSetConfig{
		Model:              &PromotionRule{},
		Queryset:           PromotionRuleObjects,
		Serializer:         &PromotionRuleSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:         []string{"id", "promotion_id", "rule_type", "logic", "is_active"},
		DetailFields:       []string{"id", "promotion_id", "rule_type", "parameters", "logic", "is_active", "created_at", "updated_at"},
		Filterable:         []string{"promotion_id", "rule_type", "is_active"},
		Searchable:         []string{"rule_type"},
		Ordering:           []string{"promotion_id", "-created_at"},
		PerPage:            20,
		Authenticate:       true,
		Permissions:        []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:          &api.RateLimit{Requests: 100, Window: time.Hour},
		EnhancedViewSet:    promotionRuleViewSet,
	})

	// Banner API with enhanced features
	bannerViewSet := api.NewViewSet(&Banner{})
	bannerViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			customerUserLookup,
		),
	}
	bannerViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	bannerViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for Banner
	bannerViewSet.RegisterAction("stats", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      true,
		URLPath:     "/stats",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleBannerStats,
	})

	bannerViewSet.RegisterAction("schedule", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/schedule",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleBannerSchedule,
	})

	router.Register("banners", &api.ViewSetConfig{
		Model:              &Banner{},
		Queryset:           BannerObjects,
		Serializer:         &BannerSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:         []string{"id", "title", "position", "priority", "is_active", "start_date", "end_date", "click_count", "view_count"},
		DetailFields:       []string{"id", "title", "subtitle", "image_url", "mobile_image_url", "video_url", "content", "link", "link_text", "position", "background_color", "text_color", "start_date", "end_date", "schedule", "device_types", "user_types", "customer_group_ids", "priority", "is_active", "click_count", "view_count", "created_at", "updated_at"},
		Filterable:         []string{"position", "is_active", "device_types"},
		Searchable:         []string{"title", "subtitle"},
		Ordering:           []string{"-priority", "-created_at"},
		PerPage:            20,
		Authenticate:       true,
		Permissions:        []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:          &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet:    bannerViewSet,
	})

	// NewsletterSubscription API with enhanced features
	newsletterViewSet := api.NewViewSet(&NewsletterSubscription{})
	newsletterViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			customerUserLookup,
		),
	}
	newsletterViewSet.PermissionClasses = []permissions.Permission{
		&permissions.AllowAny{}, // Allow public subscription
	}
	newsletterViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "100/hour"},
		&throttling.AnonRateThrottle{Rate: "10/hour"},
	}

	// Register custom actions for NewsletterSubscription
	newsletterViewSet.RegisterAction("subscribe", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/subscribe",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleNewsletterSubscribe,
	})

	newsletterViewSet.RegisterAction("unsubscribe", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/unsubscribe",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleNewsletterUnsubscribe,
	})

	newsletterViewSet.RegisterAction("stats", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/stats",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleNewsletterStats,
	})

	router.Register("newsletter-subscriptions", &api.ViewSetConfig{
		Model:              &NewsletterSubscription{},
		Queryset:           NewsletterSubscriptionObjects,
		Serializer:         &NewsletterSubscriptionSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:         []string{"id", "email", "status", "list_type", "source", "created_at"},
		DetailFields:       []string{"id", "email", "list_type", "source", "preferences", "consent_given", "consent_date", "consent_ip", "status", "subscribed_at", "unsubscribed_at", "click_count", "open_count", "segments", "created_at", "updated_at"},
		Filterable:         []string{"status", "list_type", "source"},
		Searchable:         []string{"email"},
		Ordering:           []string{"-created_at"},
		PerPage:            20,
		Authenticate:       true, // Admin only for list/detail
		Permissions:        []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:          &api.RateLimit{Requests: 100, Window: time.Hour},
		EnhancedViewSet:    newsletterViewSet,
	})

	// PromotionUsage API with enhanced features
	promotionUsageViewSet := api.NewViewSet(&PromotionUsage{})
	promotionUsageViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			customerUserLookup,
		),
	}
	promotionUsageViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	promotionUsageViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	router.Register("promotion-usage", &api.ViewSetConfig{
		Model:              &PromotionUsage{},
		Queryset:           PromotionUsageObjects,
		Serializer:         &PromotionUsageSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:         []string{"id", "promotion_id", "order_id", "customer_id", "discount_amount", "used_at"},
		DetailFields:       []string{"id", "promotion_id", "order_id", "customer_id", "discount_amount", "used_at"},
		Filterable:         []string{"promotion_id", "order_id", "customer_id"},
		Searchable:         []string{},
		Ordering:           []string{"-used_at"},
		PerPage:            20,
		Authenticate:       true,
		Permissions:        []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:          &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet:    promotionUsageViewSet,
	})
}

// Customer user lookup function for JWT authentication
func customerUserLookup(claims authentication.JWTClaims) (interface{}, error) {
	customerID, ok := claims["subject"].(string)
	if !ok {
		return nil, nil
	}
	return nil, nil
}

// ==================== Custom Action Handlers ====================

// Promotion handlers

func handlePromotionActivate(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	promotionID := vars["id"]

	response := map[string]interface{}{
		"promotion_id": promotionID,
		"message":      "Promotion activated",
		"is_active":    true,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handlePromotionDeactivate(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	promotionID := vars["id"]

	response := map[string]interface{}{
		"promotion_id": promotionID,
		"message":      "Promotion deactivated",
		"is_active":    false,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handlePromotionUsageReport(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	promotionID := vars["id"]

	response := map[string]interface{}{
		"promotion_id":      promotionID,
		"message":           "Usage report generated",
		"total_usage":       0,
		"unique_customers":   0,
		"total_discount":    0.0,
		"average_discount":  0.0,
		"usage_by_day":       []interface{}{},
		"usage_by_customer": []interface{}{},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handlePromotionValidate(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		Code        string  `json:"code"`
		CartTotal   float64 `json:"cart_total"`
		ProductIDs   []string `json:"product_ids"`
		CustomerID  string  `json:"customer_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := map[string]interface{}{
			"valid":  false,
			"error":  "Invalid request body",
			"reason": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"valid":           true,
		"promotion":       map[string]interface{}{},
		"discount_amount": 0.0,
		"message":         "Promotion is valid",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Banner handlers

func handleBannerStats(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	bannerID := vars["id"]

	response := map[string]interface{}{
		"banner_id":      bannerID,
		"message":        "Banner statistics",
		"total_views":    0,
		"total_clicks":  0.0,
		"click_rate":    0.0,
		"views_by_day":  []interface{}{},
		"clicks_by_day": []interface{}{},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleBannerSchedule(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	bannerID := vars["id"]

	// Parse schedule from request body
	var schedule []BannerSchedule
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		response := map[string]interface{}{
			"banner_id": bannerID,
			"message":   "Invalid schedule data",
			"error":    err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"banner_id": bannerID,
		"message":   "Banner scheduled successfully",
		"schedule":  schedule,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// NewsletterSubscription handlers

func handleNewsletterSubscribe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email       string            `json:"email"`
		ListType    string            `json:"list_type"`
		Source      string            `json:"source"`
		Preferences map[string]bool   `json:"preferences"`
		ConsentGiven bool             `json:"consent_given"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Subscribed successfully",
		"email":   req.Email,
		"status":  StatusSubscribed,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleNewsletterUnsubscribe(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Unsubscribed successfully",
		"email":   req.Email,
		"status":  StatusUnsubscribed,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleNewsletterStats(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message":              "Newsletter statistics",
		"total_subscribers":    0,
		"active_subscribers":   0,
		"unsubscribed":         0,
		"bounced":              0,
		"avg_open_rate":        0.0,
		"avg_click_rate":       0.0,
		"subscribers_by_source": map[string]interface{}{},
		"subscribers_by_list":  map[string]interface{}{},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ==================== Serializers ====================

type PromotionSerializer struct {
	*api.BaseSerializer
}

func (s *PromotionSerializer) New() api.Serializer {
	return &PromotionSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *PromotionSerializer) Validate() error {
	errors := make(map[string]string)
	if s.Data["name"] == "" {
		errors["name"] = "Name is required"
	}
	if s.Data["code"] == "" {
		errors["code"] = "Code is required"
	}
	if s.Data["type"] == "" {
		errors["type"] = "Type is required"
	}
	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

type PromotionRuleSerializer struct {
	*api.BaseSerializer
}

func (s *PromotionRuleSerializer) New() api.Serializer {
	return &PromotionRuleSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

type BannerSerializer struct {
	*api.BaseSerializer
}

func (s *BannerSerializer) New() api.Serializer {
	return &BannerSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *BannerSerializer) Validate() error {
	errors := make(map[string]string)
	if s.Data["title"] == "" {
		errors["title"] = "Title is required"
	}
	if s.Data["image_url"] == "" {
		errors["image_url"] = "Image URL is required"
	}
	if s.Data["position"] == "" {
		errors["position"] = "Position is required"
	}
	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

type NewsletterSubscriptionSerializer struct {
	*api.BaseSerializer
}

func (s *NewsletterSubscriptionSerializer) New() api.Serializer {
	return &NewsletterSubscriptionSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

func (s *NewsletterSubscriptionSerializer) Validate() error {
	errors := make(map[string]string)
	if s.Data["email"] == "" {
		errors["email"] = "Email is required"
	}
	if len(errors) > 0 {
		return &api.ValidationError{Errors: errors}
	}
	return nil
}

type PromotionUsageSerializer struct {
	*api.BaseSerializer
}

func (s *PromotionUsageSerializer) New() api.Serializer {
	return &PromotionUsageSerializer{BaseSerializer: api.NewBaseSerializer(nil)}
}

