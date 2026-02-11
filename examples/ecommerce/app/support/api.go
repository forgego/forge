package support

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

// RegisterAPI registers support API endpoints with enhanced features
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// SupportTicket API with enhanced features
	supportTicketViewSet := api.NewViewSet(&SupportTicket{})
	supportTicketViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			supportUserLookup,
		),
	}
	supportTicketViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	supportTicketViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for SupportTicket
	supportTicketViewSet.RegisterAction("assign", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/assign",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleTicketAssign,
	})

	supportTicketViewSet.RegisterAction("escalate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/escalate",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleTicketEscalate,
	})

	supportTicketViewSet.RegisterAction("resolve", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/resolve",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleTicketResolve,
	})

	supportTicketViewSet.RegisterAction("reopen", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/reopen",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleTicketReopen,
	})

	supportTicketViewSet.RegisterAction("add-message", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/add-message",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleTicketAddMessage,
	})

	router.Register("support-tickets", &api.ViewSetConfig{
		Model:           &SupportTicket{},
		Queryset:        SupportTicketObjects,
		Serializer:      &SupportTicketSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "ticket_number", "type", "priority", "status", "assigned_group", "customer_id", "created_at"},
		DetailFields:    []string{"id", "ticket_number", "customer_id", "type", "subject", "description", "priority", "status", "assigned_to", "assigned_group", "related_type", "related_id", "tags", "resolution", "resolved_at", "customer_satisfaction", "customer_feedback", "is_escalated", "escalated_at", "escalation_reason", "first_response_at", "created_at", "updated_at"},
		Filterable:      []string{"status", "priority", "type", "assigned_group", "is_escalated", "customer_id"},
		Searchable:      []string{"ticket_number", "subject", "description"},
		Ordering:        []string{"-created_at", "priority"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: supportTicketViewSet,
	})

	// ReturnRequest API with enhanced features
	returnRequestViewSet := api.NewViewSet(&ReturnRequest{})
	returnRequestViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			supportUserLookup,
		),
	}
	returnRequestViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	returnRequestViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for ReturnRequest
	returnRequestViewSet.RegisterAction("approve", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/approve",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleReturnApprove,
	})

	returnRequestViewSet.RegisterAction("reject", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/reject",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleReturnReject,
	})

	returnRequestViewSet.RegisterAction("process-refund", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/process-refund",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleReturnProcessRefund,
	})

	returnRequestViewSet.RegisterAction("ship-return", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/ship-return",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleReturnShip,
	})

	returnRequestViewSet.RegisterAction("receive-inspect", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/receive-inspect",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleReturnInspect,
	})

	router.Register("return-requests", &api.ViewSetConfig{
		Model:           &ReturnRequest{},
		Queryset:        ReturnRequestObjects,
		Serializer:      &ReturnRequestSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "return_number", "order_id", "reason", "resolution_type", "status", "refund_amount", "created_at"},
		DetailFields:    []string{"id", "return_number", "order_id", "customer_id", "reason", "reason_detail", "items", "resolution_type", "resolution_detail", "refund_amount", "refund_method", "refund_status", "refund_processed_at", "return_shipping_method", "return_shipping_label_url", "return_tracking_number", "return_carrier", "status", "status_history", "inspected_at", "inspection_notes", "inspection_result", "approved_at", "returned_at", "completed_at", "created_at", "updated_at"},
		Filterable:      []string{"status", "reason", "resolution_type", "refund_status", "customer_id", "order_id"},
		Searchable:      []string{"return_number"},
		Ordering:        []string{"-created_at"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: returnRequestViewSet,
	})

	// LiveChatSession API with enhanced features
	liveChatViewSet := api.NewViewSet(&LiveChatSession{})
	liveChatViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			supportUserLookup,
		),
	}
	liveChatViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	liveChatViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "1000/hour"},
		&throttling.AnonRateThrottle{Rate: "100/hour"},
	}

	// Register custom actions for LiveChatSession
	liveChatViewSet.RegisterAction("start", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/start",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleChatStart,
	})

	liveChatViewSet.RegisterAction("end", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/end",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleChatEnd,
	})

	liveChatViewSet.RegisterAction("transfer", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/transfer",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleChatTransfer,
	})

	liveChatViewSet.RegisterAction("send-message", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/send-message",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleChatSendMessage,
	})

	router.Register("live-chat-sessions", &api.ViewSetConfig{
		Model:           &LiveChatSession{},
		Queryset:        LiveChatSessionObjects,
		Serializer:      &LiveChatSessionSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "session_id", "status", "type", "customer_id", "agent_id", "wait_time_seconds", "chat_duration_seconds", "started_at"},
		DetailFields:    []string{"id", "session_id", "customer_id", "guest_name", "guest_email", "agent_id", "agent_name", "status", "type", "subject", "messages", "end_reason", "end_notes", "customer_satisfaction", "wait_time_seconds", "chat_duration_seconds", "related_ticket_id", "related_order_id", "started_at", "ended_at"},
		Filterable:      []string{"status", "type", "agent_id", "customer_id"},
		Searchable:      []string{"session_id", "guest_name", "guest_email"},
		Ordering:        []string{"-started_at"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 1000, Window: time.Hour},
		EnhancedViewSet: liveChatViewSet,
	})

	// FAQ API with enhanced features
	faqViewSet := api.NewViewSet(&FAQ{})
	faqViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			supportUserLookup,
		),
	}
	faqViewSet.PermissionClasses = []permissions.Permission{
		&permissions.AllowAny{},
	}
	faqViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "1000/hour"},
		&throttling.AnonRateThrottle{Rate: "500/hour"},
	}

	// Register custom actions for FAQ
	faqViewSet.RegisterAction("search", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/search",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleFAQSearch,
	})

	faqViewSet.RegisterAction("mark-helpful", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/mark-helpful",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleFAQMarkHelpful,
	})

	router.Register("faqs", &api.ViewSetConfig{
		Model:           &FAQ{},
		Queryset:        FAQObjects,
		Serializer:      &FAQSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "question", "category", "is_featured", "is_visible", "display_order", "view_count", "helpful_count"},
		DetailFields:    []string{"id", "question", "answer", "category", "keywords", "display_order", "is_visible", "is_featured", "view_count", "helpful_count", "not_helpful_count", "created_at", "updated_at"},
		Filterable:      []string{"category", "is_visible", "is_featured"},
		Searchable:      []string{"question", "answer", "keywords"},
		Ordering:        []string{"display_order", "category"},
		PerPage:         20,
		Authenticate:     false,
		Permissions:     []api.Permission{&permissions.AllowAny{}},
		RateLimit:       &api.RateLimit{Requests: 1000, Window: time.Hour},
		EnhancedViewSet: faqViewSet,
	})

	// SupportMessage API
	supportMessageViewSet := api.NewViewSet(&SupportMessage{})
	supportMessageViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			supportUserLookup,
		),
	}
	supportMessageViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
	}

	router.Register("support-messages", &api.ViewSetConfig{
		Model:           &SupportMessage{},
		Queryset:        SupportMessageObjects,
		Serializer:      &SupportMessageSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "ticket_id", "sender_type", "is_internal_note", "created_at"},
		DetailFields:    []string{"id", "ticket_id", "sender_type", "customer_id", "agent_id", "content", "content_type", "attachments", "is_internal_note", "is_automated", "automation_rule_id", "created_at"},
		Filterable:      []string{"ticket_id", "sender_type", "is_internal_note"},
		Searchable:      []string{"content"},
		Ordering:        []string{"created_at"},
		PerPage:         50,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: supportMessageViewSet,
	})
}

// supportUserLookup looks up a user from JWT claims
func supportUserLookup(claims authentication.JWTClaims) (interface{}, error) {
	userID, ok := claims["subject"].(string)
	if !ok {
		if subject, ok := claims["subject"].(float64); ok {
			userID = strconv.FormatFloat(subject, 'f', -1, 64)
		} else {
			return nil, nil
		}
	}
	// In a real implementation, you would query the database
	return nil, nil
}

// Custom action handlers for SupportTicket

func handleTicketAssign(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	ticketID := vars["id"]

	var req struct {
		AgentID      string `json:"agent_id"`
		AssignedGroup string `json:"assigned_group"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"ticket_id":       ticketID,
		"agent_id":       req.AgentID,
		"assigned_group": req.AssignedGroup,
		"status":         "assigned",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTicketEscalate(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	ticketID := vars["id"]

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"ticket_id":          ticketID,
		"is_escalated":      true,
		"escalation_reason": req.Reason,
		"escalated_at":      time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTicketResolve(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	ticketID := vars["id"]

	var req struct {
		Resolution string `json:"resolution"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"ticket_id":   ticketID,
		"status":      "resolved",
		"resolution":  req.Resolution,
		"resolved_at": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTicketReopen(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	ticketID := vars["id"]

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"ticket_id": ticketID,
		"status":    "open",
		"reason":    req.Reason,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTicketAddMessage(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	ticketID := vars["id"]

	var req struct {
		Content        string `json:"content"`
		ContentType    string `json:"content_type"`
		IsInternalNote bool   `json:"is_internal_note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"message_id":      "msg_" + ticketID,
		"ticket_id":       ticketID,
		"content":         req.Content,
		"content_type":    req.ContentType,
		"is_internal_note": req.IsInternalNote,
		"created_at":      time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for ReturnRequest

func handleReturnApprove(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	returnID := vars["id"]

	response := map[string]interface{}{
		"return_id":  returnID,
		"status":     "approved",
		"approved_at": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleReturnReject(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	returnID := vars["id"]

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"return_id": returnID,
		"status":    "rejected",
		"reason":    req.Reason,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleReturnProcessRefund(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	returnID := vars["id"]

	var req struct {
		Method string `json:"method"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"return_id":         returnID,
		"refund_status":     "processing",
		"refund_method":     req.Method,
		"refund_processed_at": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleReturnShip(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	returnID := vars["id"]

	var req struct {
		TrackingNumber string `json:"tracking_number"`
		Carrier        string `json:"carrier"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"return_id":           returnID,
		"status":             "in_transit",
		"return_tracking_number": req.TrackingNumber,
		"return_carrier":     req.Carrier,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleReturnInspect(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	returnID := vars["id"]

	var req struct {
		Notes   string `json:"notes"`
		Result  string `json:"result"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"return_id":         returnID,
		"status":            "received_inspected",
		"inspection_notes":  req.Notes,
		"inspection_result": req.Result,
		"inspected_at":      time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for LiveChatSession

func handleChatStart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID string `json:"customer_id"`
		GuestName  string `json:"guest_name"`
		GuestEmail string `json:"guest_email"`
		Type       string `json:"type"`
		Subject    string `json:"subject"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"session_id":        "chat_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		"status":            "waiting",
		"customer_id":       req.CustomerID,
		"guest_name":       req.GuestName,
		"guest_email":      req.GuestEmail,
		"type":             req.Type,
		"subject":          req.Subject,
		"wait_time_seconds": 0,
		"started_at":       time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleChatEnd(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	sessionID := vars["id"]

	var req struct {
		Reason  string `json:"reason"`
		Notes   string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"session_id":  sessionID,
		"status":      "ended",
		"end_reason":  req.Reason,
		"end_notes":   req.Notes,
		"ended_at":    time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleChatTransfer(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	sessionID := vars["id"]

	var req struct {
		NewAgentID string `json:"new_agent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"session_id":   sessionID,
		"status":       "transferred",
		"new_agent_id": req.NewAgentID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleChatSendMessage(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	sessionID := vars["id"]

	var req struct {
		SenderType  string `json:"sender_type"`
		SenderID    string `json:"sender_id"`
		SenderName  string `json:"sender_name"`
		Content     string `json:"content"`
		ContentType string `json:"content_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"message_id":   "msg_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		"session_id":   sessionID,
		"sender_type":  req.SenderType,
		"sender_id":    req.SenderID,
		"sender_name":  req.SenderName,
		"content":      req.Content,
		"content_type": req.ContentType,
		"timestamp":    time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for FAQ

func handleFAQSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")

	// Return FAQs matching the query (placeholder)
	response := map[string]interface{}{
		"query":    query,
		"category": category,
		"results": []interface{}{
			map[string]interface{}{
				"id":       "1",
				"question": "How do I track my order?",
				"answer":   "You can track your order by logging into your account...",
				"category": "orders",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleFAQMarkHelpful(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	faqID := vars["id"]

	var req struct {
		Helpful bool `json:"helpful"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Helpful {
		response := map[string]interface{}{
			"faq_id":         faqID,
			"helpful_count":  1,
			"not_helpful_count": 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		response := map[string]interface{}{
			"faq_id":            faqID,
			"helpful_count":     0,
			"not_helpful_count": 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
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

// RegisterAPI registers support API endpoints with enhanced features
func RegisterAPI(ctx context.Context, router *api.Router, database *db.DB) {
	// SupportTicket API with enhanced features
	supportTicketViewSet := api.NewViewSet(&SupportTicket{})
	supportTicketViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			supportUserLookup,
		),
	}
	supportTicketViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	supportTicketViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for SupportTicket
	supportTicketViewSet.RegisterAction("assign", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/assign",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleTicketAssign,
	})

	supportTicketViewSet.RegisterAction("escalate", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/escalate",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleTicketEscalate,
	})

	supportTicketViewSet.RegisterAction("resolve", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/resolve",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleTicketResolve,
	})

	supportTicketViewSet.RegisterAction("reopen", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/reopen",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleTicketReopen,
	})

	supportTicketViewSet.RegisterAction("add-message", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/add-message",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleTicketAddMessage,
	})

	router.Register("support-tickets", &api.ViewSetConfig{
		Model:           &SupportTicket{},
		Queryset:        SupportTicketObjects,
		Serializer:      &SupportTicketSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "ticket_number", "type", "priority", "status", "assigned_group", "customer_id", "created_at"},
		DetailFields:    []string{"id", "ticket_number", "customer_id", "type", "subject", "description", "priority", "status", "assigned_to", "assigned_group", "related_type", "related_id", "tags", "resolution", "resolved_at", "customer_satisfaction", "customer_feedback", "is_escalated", "escalated_at", "escalation_reason", "first_response_at", "created_at", "updated_at"},
		Filterable:      []string{"status", "priority", "type", "assigned_group", "is_escalated", "customer_id"},
		Searchable:      []string{"ticket_number", "subject", "description"},
		Ordering:        []string{"-created_at", "priority"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: supportTicketViewSet,
	})

	// ReturnRequest API with enhanced features
	returnRequestViewSet := api.NewViewSet(&ReturnRequest{})
	returnRequestViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			supportUserLookup,
		),
	}
	returnRequestViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	returnRequestViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "500/hour"},
		&throttling.AnonRateThrottle{Rate: "50/hour"},
	}

	// Register custom actions for ReturnRequest
	returnRequestViewSet.RegisterAction("approve", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/approve",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleReturnApprove,
	})

	returnRequestViewSet.RegisterAction("reject", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/reject",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleReturnReject,
	})

	returnRequestViewSet.RegisterAction("process-refund", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/process-refund",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleReturnProcessRefund,
	})

	returnRequestViewSet.RegisterAction("ship-return", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/ship-return",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleReturnShip,
	})

	returnRequestViewSet.RegisterAction("receive-inspect", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/receive-inspect",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleReturnInspect,
	})

	router.Register("return-requests", &api.ViewSetConfig{
		Model:           &ReturnRequest{},
		Queryset:        ReturnRequestObjects,
		Serializer:      &ReturnRequestSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "return_number", "order_id", "reason", "resolution_type", "status", "refund_amount", "created_at"},
		DetailFields:    []string{"id", "return_number", "order_id", "customer_id", "reason", "reason_detail", "items", "resolution_type", "resolution_detail", "refund_amount", "refund_method", "refund_status", "refund_processed_at", "return_shipping_method", "return_shipping_label_url", "return_tracking_number", "return_carrier", "status", "status_history", "inspected_at", "inspection_notes", "inspection_result", "approved_at", "returned_at", "completed_at", "created_at", "updated_at"},
		Filterable:      []string{"status", "reason", "resolution_type", "refund_status", "customer_id", "order_id"},
		Searchable:      []string{"return_number"},
		Ordering:        []string{"-created_at"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: returnRequestViewSet,
	})

	// LiveChatSession API with enhanced features
	liveChatViewSet := api.NewViewSet(&LiveChatSession{})
	liveChatViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			supportUserLookup,
		),
	}
	liveChatViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
		&permissions.IsAdmin{},
	}
	liveChatViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "1000/hour"},
		&throttling.AnonRateThrottle{Rate: "100/hour"},
	}

	// Register custom actions for LiveChatSession
	liveChatViewSet.RegisterAction("start", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      false,
		URLPath:     "/start",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleChatStart,
	})

	liveChatViewSet.RegisterAction("end", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/end",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleChatEnd,
	})

	liveChatViewSet.RegisterAction("transfer", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/transfer",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		Handler:     handleChatTransfer,
	})

	liveChatViewSet.RegisterAction("send-message", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/send-message",
		Permissions: []api.Permission{&permissions.IsAuthenticated{}},
		Handler:     handleChatSendMessage,
	})

	router.Register("live-chat-sessions", &api.ViewSetConfig{
		Model:           &LiveChatSession{},
		Queryset:        LiveChatSessionObjects,
		Serializer:      &LiveChatSessionSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "session_id", "status", "type", "customer_id", "agent_id", "wait_time_seconds", "chat_duration_seconds", "started_at"},
		DetailFields:    []string{"id", "session_id", "customer_id", "guest_name", "guest_email", "agent_id", "agent_name", "status", "type", "subject", "messages", "end_reason", "end_notes", "customer_satisfaction", "wait_time_seconds", "chat_duration_seconds", "related_ticket_id", "related_order_id", "started_at", "ended_at"},
		Filterable:      []string{"status", "type", "agent_id", "customer_id"},
		Searchable:      []string{"session_id", "guest_name", "guest_email"},
		Ordering:        []string{"-started_at"},
		PerPage:         20,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 1000, Window: time.Hour},
		EnhancedViewSet: liveChatViewSet,
	})

	// FAQ API with enhanced features
	faqViewSet := api.NewViewSet(&FAQ{})
	faqViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			supportUserLookup,
		),
	}
	faqViewSet.PermissionClasses = []permissions.Permission{
		&permissions.AllowAny{},
	}
	faqViewSet.ThrottleClasses = []throttling.Throttle{
		&throttling.UserRateThrottle{Rate: "1000/hour"},
		&throttling.AnonRateThrottle{Rate: "500/hour"},
	}

	// Register custom actions for FAQ
	faqViewSet.RegisterAction("search", &api.ActionConfig{
		Methods:     []string{"GET"},
		Detail:      false,
		URLPath:     "/search",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleFAQSearch,
	})

	faqViewSet.RegisterAction("mark-helpful", &api.ActionConfig{
		Methods:     []string{"POST"},
		Detail:      true,
		URLPath:     "/mark-helpful",
		Permissions: []api.Permission{&permissions.AllowAny{}},
		Handler:     handleFAQMarkHelpful,
	})

	router.Register("faqs", &api.ViewSetConfig{
		Model:           &FAQ{},
		Queryset:        FAQObjects,
		Serializer:      &FAQSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "question", "category", "is_featured", "is_visible", "display_order", "view_count", "helpful_count"},
		DetailFields:    []string{"id", "question", "answer", "category", "keywords", "display_order", "is_visible", "is_featured", "view_count", "helpful_count", "not_helpful_count", "created_at", "updated_at"},
		Filterable:      []string{"category", "is_visible", "is_featured"},
		Searchable:      []string{"question", "answer", "keywords"},
		Ordering:        []string{"display_order", "category"},
		PerPage:         20,
		Authenticate:     false,
		Permissions:     []api.Permission{&permissions.AllowAny{}},
		RateLimit:       &api.RateLimit{Requests: 1000, Window: time.Hour},
		EnhancedViewSet: faqViewSet,
	})

	// SupportMessage API
	supportMessageViewSet := api.NewViewSet(&SupportMessage{})
	supportMessageViewSet.AuthenticationClasses = []authentication.Authentication{
		authentication.NewJWTAuthentication(
			[]byte("your-jwt-secret-key"),
			supportUserLookup,
		),
	}
	supportMessageViewSet.PermissionClasses = []permissions.Permission{
		&permissions.IsAuthenticated{},
	}

	router.Register("support-messages", &api.ViewSetConfig{
		Model:           &SupportMessage{},
		Queryset:        SupportMessageObjects,
		Serializer:      &SupportMessageSerializer{BaseSerializer: api.NewBaseSerializer(nil)},
		ListFields:      []string{"id", "ticket_id", "sender_type", "is_internal_note", "created_at"},
		DetailFields:    []string{"id", "ticket_id", "sender_type", "customer_id", "agent_id", "content", "content_type", "attachments", "is_internal_note", "is_automated", "automation_rule_id", "created_at"},
		Filterable:      []string{"ticket_id", "sender_type", "is_internal_note"},
		Searchable:      []string{"content"},
		Ordering:        []string{"created_at"},
		PerPage:         50,
		Authenticate:     true,
		Permissions:     []api.Permission{&permissions.IsAuthenticated{}, &permissions.IsAdmin{}},
		RateLimit:       &api.RateLimit{Requests: 500, Window: time.Hour},
		EnhancedViewSet: supportMessageViewSet,
	})
}

// supportUserLookup looks up a user from JWT claims
func supportUserLookup(claims authentication.JWTClaims) (interface{}, error) {
	userID, ok := claims["subject"].(string)
	if !ok {
		if subject, ok := claims["subject"].(float64); ok {
			userID = strconv.FormatFloat(subject, 'f', -1, 64)
		} else {
			return nil, nil
		}
	}
	// In a real implementation, you would query the database
	return nil, nil
}

// Custom action handlers for SupportTicket

func handleTicketAssign(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	ticketID := vars["id"]

	var req struct {
		AgentID      string `json:"agent_id"`
		AssignedGroup string `json:"assigned_group"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"ticket_id":       ticketID,
		"agent_id":       req.AgentID,
		"assigned_group": req.AssignedGroup,
		"status":         "assigned",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTicketEscalate(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	ticketID := vars["id"]

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"ticket_id":          ticketID,
		"is_escalated":      true,
		"escalation_reason": req.Reason,
		"escalated_at":      time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTicketResolve(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	ticketID := vars["id"]

	var req struct {
		Resolution string `json:"resolution"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"ticket_id":   ticketID,
		"status":      "resolved",
		"resolution":  req.Resolution,
		"resolved_at": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTicketReopen(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	ticketID := vars["id"]

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"ticket_id": ticketID,
		"status":    "open",
		"reason":    req.Reason,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTicketAddMessage(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	ticketID := vars["id"]

	var req struct {
		Content        string `json:"content"`
		ContentType    string `json:"content_type"`
		IsInternalNote bool   `json:"is_internal_note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"message_id":      "msg_" + ticketID,
		"ticket_id":       ticketID,
		"content":         req.Content,
		"content_type":    req.ContentType,
		"is_internal_note": req.IsInternalNote,
		"created_at":      time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for ReturnRequest

func handleReturnApprove(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	returnID := vars["id"]

	response := map[string]interface{}{
		"return_id":  returnID,
		"status":     "approved",
		"approved_at": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleReturnReject(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	returnID := vars["id"]

	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"return_id": returnID,
		"status":    "rejected",
		"reason":    req.Reason,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleReturnProcessRefund(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	returnID := vars["id"]

	var req struct {
		Method string `json:"method"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"return_id":         returnID,
		"refund_status":     "processing",
		"refund_method":     req.Method,
		"refund_processed_at": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleReturnShip(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	returnID := vars["id"]

	var req struct {
		TrackingNumber string `json:"tracking_number"`
		Carrier        string `json:"carrier"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"return_id":           returnID,
		"status":             "in_transit",
		"return_tracking_number": req.TrackingNumber,
		"return_carrier":     req.Carrier,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleReturnInspect(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	returnID := vars["id"]

	var req struct {
		Notes   string `json:"notes"`
		Result  string `json:"result"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"return_id":         returnID,
		"status":            "received_inspected",
		"inspection_notes":  req.Notes,
		"inspection_result": req.Result,
		"inspected_at":      time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for LiveChatSession

func handleChatStart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CustomerID string `json:"customer_id"`
		GuestName  string `json:"guest_name"`
		GuestEmail string `json:"guest_email"`
		Type       string `json:"type"`
		Subject    string `json:"subject"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"session_id":        "chat_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		"status":            "waiting",
		"customer_id":       req.CustomerID,
		"guest_name":       req.GuestName,
		"guest_email":      req.GuestEmail,
		"type":             req.Type,
		"subject":          req.Subject,
		"wait_time_seconds": 0,
		"started_at":       time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleChatEnd(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	sessionID := vars["id"]

	var req struct {
		Reason  string `json:"reason"`
		Notes   string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"session_id":  sessionID,
		"status":      "ended",
		"end_reason":  req.Reason,
		"end_notes":   req.Notes,
		"ended_at":    time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleChatTransfer(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	sessionID := vars["id"]

	var req struct {
		NewAgentID string `json:"new_agent_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"session_id":   sessionID,
		"status":       "transferred",
		"new_agent_id": req.NewAgentID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleChatSendMessage(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	sessionID := vars["id"]

	var req struct {
		SenderType  string `json:"sender_type"`
		SenderID    string `json:"sender_id"`
		SenderName  string `json:"sender_name"`
		Content     string `json:"content"`
		ContentType string `json:"content_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"message_id":   "msg_" + strconv.FormatInt(time.Now().UnixNano(), 10),
		"session_id":   sessionID,
		"sender_type":  req.SenderType,
		"sender_id":    req.SenderID,
		"sender_name":  req.SenderName,
		"content":      req.Content,
		"content_type": req.ContentType,
		"timestamp":    time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom action handlers for FAQ

func handleFAQSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")

	// Return FAQs matching the query (placeholder)
	response := map[string]interface{}{
		"query":    query,
		"category": category,
		"results": []interface{}{
			map[string]interface{}{
				"id":       "1",
				"question": "How do I track my order?",
				"answer":   "You can track your order by logging into your account...",
				"category": "orders",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleFAQMarkHelpful(w http.ResponseWriter, r *http.Request) {
	vars := api.GetURLVars(r)
	faqID := vars["id"]

	var req struct {
		Helpful bool `json:"helpful"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Helpful {
		response := map[string]interface{}{
			"faq_id":         faqID,
			"helpful_count":  1,
			"not_helpful_count": 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		response := map[string]interface{}{
			"faq_id":            faqID,
			"helpful_count":     0,
			"not_helpful_count": 1,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

