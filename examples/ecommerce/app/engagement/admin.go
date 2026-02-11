package engagement

import (
	"context"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
)

// RegisterAdmin registers engagement models with the admin interface
func RegisterAdmin(ctx context.Context) {
	// RecentlyViewed admin
	admin.Register(&admin.Config[RecentlyViewed]{
		Icon: "Eye",
		ListDisplay: []admin.Field{
			RecentlyViewedFieldsInstance.ID,
			RecentlyViewedFieldsInstance.CustomerID,
			RecentlyViewedFieldsInstance.ProductID,
			RecentlyViewedFieldsInstance.ViewedAt,
			RecentlyViewedFieldsInstance.Source,
			RecentlyViewedFieldsInstance.SessionID,
		},
		ListFilter: []admin.Field{
			RecentlyViewedFieldsInstance.CustomerID,
			RecentlyViewedFieldsInstance.Source,
		},
		SearchFields: []admin.Field{
			RecentlyViewedFieldsInstance.CustomerID,
			RecentlyViewedFieldsInstance.GuestID,
			RecentlyViewedFieldsInstance.SessionID,
		},
		Ordering: []admin.Field{
			RecentlyViewedFieldsInstance.ViewedAt,
		},
		Fieldsets: []admin.Fieldset[RecentlyViewed]{
			{
				Name: "Product Information",
				Fields: []string{"customer_id", "guest_id", "product_id", "variant_id"},
			},
			{
				Name: "Tracking",
				Fields: []string{"viewed_at", "session_id", "user_agent", "ip_address"},
			},
			{
				Name: "Context",
				Fields: []string{"source", "referer_url"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"customer_id", "product_id", "viewed_at"},
		},
		ReadOnlyFields: []admin.Field{
			RecentlyViewedFieldsInstance.ViewedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[RecentlyViewed], user interface{}, obj *RecentlyViewed) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[RecentlyViewed], user interface{}, obj *RecentlyViewed) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[RecentlyViewed], user interface{}, obj *RecentlyViewed) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[RecentlyViewed], user interface{}, obj *RecentlyViewed) bool {
			return true
		},
		Filters: []admin.Filter[RecentlyViewed]{
			{
				Name:  "by_customer",
				Label: "By Customer",
				Type:  admin.FilterTypeObjectID,
				Handler: func(ctx context.Context, qs orm.QuerySet[RecentlyViewed], value interface{}) orm.QuerySet[RecentlyViewed] {
					return qs.Filter("customer_id", value)
				},
			},
			{
				Name:  "by_source",
				Label: "By Source",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "search", Label: "Search"},
					{Value: "browse", Label: "Browse"},
					{Value: "category", Label: "Category"},
					{Value: "product", Label: "Product"},
					{Value: "recommended", Label: "Recommended"},
					{Value: "email", Label: "Email"},
					{Value: "social", Label: "Social"},
					{Value: "direct", Label: "Direct"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[RecentlyViewed], value interface{}) orm.QuerySet[RecentlyViewed] {
					return qs.Filter("source", value)
				},
			},
			{
				Name:  "recent",
				Label: "Last 24 Hours",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[RecentlyViewed], value interface{}) orm.QuerySet[RecentlyViewed] {
					return qs.Filter("viewed_at__gte", value)
				},
			},
		},
		Actions: []admin.Action[RecentlyViewed]{
			{
				Name:         "clear_customer_history",
				Label:        "Clear Customer History",
				Icon:        "Trash2",
				Confirmation: "Are you sure you want to clear the recently viewed history for this customer?",
				Handler: func(ctx context.Context, admin *admin.Admin[RecentlyViewed], ids []interface{}) error {
					for _, id := range ids {
						viewed, err := RecentlyViewedObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						if err := RecentlyViewedObjects.Delete(ctx, viewed); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[RecentlyViewed], user interface{}) bool {
					return true
				},
			},
		},
	})

	// ProductComparison admin
	admin.Register(&admin.Config[ProductComparison]{
		Icon: "GitCompare",
		ListDisplay: []admin.Field{
			ProductComparisonFieldsInstance.ID,
			ProductComparisonFieldsInstance.Name,
			ProductComparisonFieldsInstance.CustomerID,
			ProductComparisonFieldsInstance.ProductIDs,
			ProductComparisonFieldsInstance.IsPublic,
			ProductComparisonFieldsInstance.ViewCount,
			ProductComparisonFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			ProductComparisonFieldsInstance.CustomerID,
			ProductComparisonFieldsInstance.IsPublic,
		},
		SearchFields: []admin.Field{
			ProductComparisonFieldsInstance.Name,
			ProductComparisonFieldsInstance.ShareToken,
		},
		Ordering: []admin.Field{
			ProductComparisonFieldsInstance.CreatedAt,
		},
		Fieldsets: []admin.Fieldset[ProductComparison]{
			{
				Name: "Basic Information",
				Fields: []string{"customer_id", "guest_id", "name"},
			},
			{
				Name: "Products",
				Fields: []string{"product_ids"},
			},
			{
				Name: "Settings",
				Fields: []string{"is_public", "share_token"},
			},
			{
				Name: "Statistics",
				Fields: []string{"view_count"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "is_public", "product_ids"},
		},
		ReadOnlyFields: []admin.Field{
			ProductComparisonFieldsInstance.ViewCount,
			ProductComparisonFieldsInstance.CreatedAt,
			ProductComparisonFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ProductComparison], user interface{}, obj *ProductComparison) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[ProductComparison], user interface{}, obj *ProductComparison) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[ProductComparison], user interface{}, obj *ProductComparison) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[ProductComparison], user interface{}, obj *ProductComparison) bool {
			return true
		},
		Filters: []admin.Filter[ProductComparison]{
			{
				Name:  "by_customer",
				Label: "By Customer",
				Type:  admin.FilterTypeObjectID,
				Handler: func(ctx context.Context, qs orm.QuerySet[ProductComparison], value interface{}) orm.QuerySet[ProductComparison] {
					return qs.Filter("customer_id", value)
				},
			},
			{
				Name:  "public",
				Label: "Public Comparisons",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ProductComparison], value interface{}) orm.QuerySet[ProductComparison] {
					return qs.Filter("is_public", value)
				},
			},
		},
		Actions: []admin.Action[ProductComparison]{
			{
				Name:         "toggle_public",
				Label:        "Toggle Public/Private",
				Icon:        "Globe",
				Confirmation: "Are you sure you want to change the visibility of the selected comparisons?",
				Handler: func(ctx context.Context, admin *admin.Admin[ProductComparison], ids []interface{}) error {
					for _, id := range ids {
						comparison, err := ProductComparisonObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						comparison.IsPublic = !comparison.IsPublic
						if err := ProductComparisonObjects.Update(ctx, comparison); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ProductComparison], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "reset_view_count",
				Label:        "Reset View Count",
				Icon:        "RotateCcw",
				Confirmation: "Are you sure you want to reset the view count for selected comparisons?",
				Handler: func(ctx context.Context, admin *admin.Admin[ProductComparison], ids []interface{}) error {
					for _, id := range ids {
						comparison, err := ProductComparisonObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						comparison.ViewCount = 0
						if err := ProductComparisonObjects.Update(ctx, comparison); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ProductComparison], user interface{}) bool {
					return true
				},
			},
		},
	})

	// Notification admin
	admin.Register(&admin.Config[Notification]{
		Icon: "Bell",
		ListDisplay: []admin.Field{
			NotificationFieldsInstance.ID,
			NotificationFieldsInstance.CustomerID,
			NotificationFieldsInstance.Type,
			NotificationFieldsInstance.Title,
			NotificationFieldsInstance.IsRead,
			NotificationFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			NotificationFieldsInstance.CustomerID,
			NotificationFieldsInstance.Type,
			NotificationFieldsInstance.IsRead,
		},
		SearchFields: []admin.Field{
			NotificationFieldsInstance.Title,
			NotificationFieldsInstance.Message,
		},
		Ordering: []admin.Field{
			NotificationFieldsInstance.CreatedAt,
		},
		Fieldsets: []admin.Fieldset[Notification]{
			{
				Name: "Recipient",
				Fields: []string{"customer_id"},
			},
			{
				Name: "Content",
				Fields: []string{"type", "title", "message"},
			},
			{
				Name: "Action",
				Fields: []string{"action_url", "action_text"},
			},
			{
				Name: "Related Entity",
				Fields: []string{"related_type", "related_id"},
			},
			{
				Name: "Status",
				Fields: []string{"is_read", "read_at"},
			},
			{
				Name: "Channels",
				Fields: []string{"push_enabled", "email_enabled", "sms_enabled"},
			},
			{
				Name: "Scheduling",
				Fields: []string{"scheduled_for", "sent_at"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"type", "title", "is_read"},
		},
		ReadOnlyFields: []admin.Field{
			NotificationFieldsInstance.CreatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Notification], user interface{}, obj *Notification) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[Notification], user interface{}, obj *Notification) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Notification], user interface{}, obj *Notification) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Notification], user interface{}, obj *Notification) bool {
			return true
		},
		Filters: []admin.Filter[Notification]{
			{
				Name:  "by_customer",
				Label: "By Customer",
				Type:  admin.FilterTypeObjectID,
				Handler: func(ctx context.Context, qs orm.QuerySet[Notification], value interface{}) orm.QuerySet[Notification] {
					return qs.Filter("customer_id", value)
				},
			},
			{
				Name:  "by_type",
				Label: "By Type",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "order_confirmation", Label: "Order Confirmation"},
					{Value: "order_shipped", Label: "Order Shipped"},
					{Value: "order_delivered", Label: "Order Delivered"},
					{Value: "order_cancelled", Label: "Order Cancelled"},
					{Value: "payment_received", Label: "Payment Received"},
					{Value: "payment_failed", Label: "Payment Failed"},
					{Value: "review_reminder", Label: "Review Reminder"},
					{Value: "wishlist_price_drop", Label: "Wishlist Price Drop"},
					{Value: "back_in_stock", Label: "Back in Stock"},
					{Value: "abandoned_cart", Label: "Abandoned Cart"},
					{Value: "coupon_available", Label: "Coupon Available"},
					{Value: "account_security", Label: "Account Security"},
					{Value: "system", Label: "System"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[Notification], value interface{}) orm.QuerySet[Notification] {
					return qs.Filter("type", value)
				},
			},
			{
				Name:  "unread",
				Label: "Unread Only",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Notification], value interface{}) orm.QuerySet[Notification] {
					return qs.Filter("is_read", false)
				},
			},
		},
		Actions: []admin.Action[Notification]{
			{
				Name:         "mark_read",
				Label:        "Mark as Read",
				Icon:        "Check",
				Confirmation: "Are you sure you want to mark the selected notifications as read?",
				Handler: func(ctx context.Context, admin *admin.Admin[Notification], ids []interface{}) error {
					for _, id := range ids {
						notification, err := NotificationObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						now := time.Now()
						notification.IsRead = true
						notification.ReadAt = &now
						if err := NotificationObjects.Update(ctx, notification); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[Notification], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "mark_all_read",
				Label:        "Mark All as Read",
				Icon:        "CheckCheck",
				Confirmation: "Are you sure you want to mark ALL notifications as read for a customer?",
				Handler: func(ctx context.Context, admin *admin.Admin[Notification], ids []interface{}) error {
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[Notification], user interface{}) bool {
					return true
				},
			},
		},
	})

	// CustomerActivity admin
	admin.Register(&admin.Config[CustomerActivity]{
		Icon: "Activity",
		ListDisplay: []admin.Field{
			CustomerActivityFieldsInstance.ID,
			CustomerActivityFieldsInstance.CustomerID,
			CustomerActivityFieldsInstance.ActivityType,
			CustomerActivityFieldsInstance.EntityType,
			CustomerActivityFieldsInstance.SessionID,
			CustomerActivityFieldsInstance.Timestamp,
		},
		ListFilter: []admin.Field{
			CustomerActivityFieldsInstance.CustomerID,
			CustomerActivityFieldsInstance.ActivityType,
			CustomerActivityFieldsInstance.EntityType,
		},
		SearchFields: []admin.Field{
			CustomerActivityFieldsInstance.SessionID,
		},
		Ordering: []admin.Field{
			CustomerActivityFieldsInstance.Timestamp,
		},
		Fieldsets: []admin.Fieldset[CustomerActivity]{
			{
				Name: "Activity",
				Fields: []string{"customer_id", "activity_type"},
			},
			{
				Name: "Entity",
				Fields: []string{"entity_type", "entity_id"},
			},
			{
				Name: "Data",
				Fields: []string{"data"},
			},
			{
				Name: "Context",
				Fields: []string{"session_id", "user_agent", "ip_address"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"activity_type", "entity_type"},
		},
		ReadOnlyFields: []admin.Field{
			CustomerActivityFieldsInstance.Timestamp,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[CustomerActivity], user interface{}, obj *CustomerActivity) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[CustomerActivity], user interface{}, obj *CustomerActivity) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[CustomerActivity], user interface{}, obj *CustomerActivity) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[CustomerActivity], user interface{}, obj *CustomerActivity) bool {
			return true
		},
		Filters: []admin.Filter[CustomerActivity]{
			{
				Name:  "by_customer",
				Label: "By Customer",
				Type:  admin.FilterTypeObjectID,
				Handler: func(ctx context.Context, qs orm.QuerySet[CustomerActivity], value interface{}) orm.QuerySet[CustomerActivity] {
					return qs.Filter("customer_id", value)
				},
			},
			{
				Name:  "by_activity_type",
				Label: "By Activity Type",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "page_view", Label: "Page View"},
					{Value: "product_view", Label: "Product View"},
					{Value: "add_to_cart", Label: "Add to Cart"},
					{Value: "remove_from_cart", Label: "Remove from Cart"},
					{Value: "checkout_start", Label: "Checkout Start"},
					{Value: "checkout_complete", Label: "Checkout Complete"},
					{Value: "search", Label: "Search"},
					{Value: "wishlist_add", Label: "Wishlist Add"},
					{Value: "wishlist_remove", Label: "Wishlist Remove"},
					{Value: "review_submit", Label: "Review Submit"},
					{Value: "login", Label: "Login"},
					{Value: "logout", Label: "Logout"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[CustomerActivity], value interface{}) orm.QuerySet[CustomerActivity] {
					return qs.Filter("activity_type", value)
				},
			},
		},
		Actions: []admin.Action[CustomerActivity]{
			{
				Name:         "export_activities",
				Label:        "Export Activities",
				Icon:        "Download",
				Confirmation: "Export customer activities to CSV?",
				Handler: func(ctx context.Context, admin *admin.Admin[CustomerActivity], ids []interface{}) error {
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[CustomerActivity], user interface{}) bool {
					return true
				},
			},
		},
	})

	// AbandonedCartReminder admin
	admin.Register(&admin.Config[AbandonedCartReminder]{
		Icon: "ShoppingCart",
		ListDisplay: []admin.Field{
			AbandonedCartReminderFieldsInstance.ID,
			AbandonedCartReminderFieldsInstance.CartID,
			AbandonedCartReminderFieldsInstance.CustomerID,
			AbandonedCartReminderFieldsInstance.ReminderNumber,
			AbandonedCartReminderFieldsInstance.Status,
			AbandonedCartReminderFieldsInstance.ReminderSentAt,
			AbandonedCartReminderFieldsInstance.RecoveredAt,
		},
		ListFilter: []admin.Field{
			AbandonedCartReminderFieldsInstance.CustomerID,
			AbandonedCartReminderFieldsInstance.Status,
			AbandonedCartReminderFieldsInstance.ReminderNumber,
		},
		SearchFields: []admin.Field{
			AbandonedCartReminderFieldsInstance.GuestEmail,
		},
		Ordering: []admin.Field{
			AbandonedCartReminderFieldsInstance.CreatedAt,
		},
		Fieldsets: []admin.Fieldset[AbandonedCartReminder]{
			{
				Name: "Cart Information",
				Fields: []string{"cart_id", "customer_id", "guest_email"},
			},
			{
				Name: "Reminder Details",
				Fields: []string{"reminder_number", "reminder_sent_at", "reminder_opened_at", "reminder_clicked_at"},
			},
			{
				Name: "Recovery",
				Fields: []string{"recovered_at", "recovered_order_id"},
			},
			{
				Name: "Status",
				Fields: []string{"status"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "reminder_number"},
		},
		ReadOnlyFields: []admin.Field{
			AbandonedCartReminderFieldsInstance.CreatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], user interface{}, obj *AbandonedCartReminder) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], user interface{}, obj *AbandonedCartReminder) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], user interface{}, obj *AbandonedCartReminder) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], user interface{}, obj *AbandonedCartReminder) bool {
			return true
		},
		Filters: []admin.Filter[AbandonedCartReminder]{
			{
				Name:  "by_customer",
				Label: "By Customer",
				Type:  admin.FilterTypeObjectID,
				Handler: func(ctx context.Context, qs orm.QuerySet[AbandonedCartReminder], value interface{}) orm.QuerySet[AbandonedCartReminder] {
					return qs.Filter("customer_id", value)
				},
			},
			{
				Name:  "by_status",
				Label: "By Status",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "pending", Label: "Pending"},
					{Value: "sent", Label: "Sent"},
					{Value: "opened", Label: "Opened"},
					{Value: "clicked", Label: "Clicked"},
					{Value: "recovered", Label: "Recovered"},
					{Value: "cancelled", Label: "Cancelled"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[AbandonedCartReminder], value interface{}) orm.QuerySet[AbandonedCartReminder] {
					return qs.Filter("status", value)
				},
			},
			{
				Name:  "recovered",
				Label: "Recovered Carts",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[AbandonedCartReminder], value interface{}) orm.QuerySet[AbandonedCartReminder] {
					return qs.Filter("status", "recovered")
				},
			},
		},
		Actions: []admin.Action[AbandonedCartReminder]{
			{
				Name:         "send_reminder",
				Label:        "Send Reminder",
				Icon:        "Send",
				Confirmation: "Are you sure you want to send reminders to selected carts?",
				Handler: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], ids []interface{}) error {
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "cancel_reminder",
				Label:        "Cancel Reminder",
				Icon:        "X",
				Confirmation: "Are you sure you want to cancel selected reminders?",
				Handler: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], ids []interface{}) error {
					for _, id := range ids {
						reminder, err := AbandonedCartReminderObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						reminder.Status = "cancelled"
						if err := AbandonedCartReminderObjects.Update(ctx, reminder); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], user interface{}) bool {
					return true
				},
			},
		},
	})

	// UserSegment admin
	admin.Register(&admin.Config[UserSegment]{
		Icon: "Users",
		ListDisplay: []admin.Field{
			UserSegmentFieldsInstance.ID,
			UserSegmentFieldsInstance.Name,
			UserSegmentFieldsInstance.Type,
			UserSegmentFieldsInstance.CustomerCount,
			UserSegmentFieldsInstance.IsActive,
			UserSegmentFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			UserSegmentFieldsInstance.Type,
			UserSegmentFieldsInstance.IsActive,
		},
		SearchFields: []admin.Field{
			UserSegmentFieldsInstance.Name,
			UserSegmentFieldsInstance.Description,
		},
		Ordering: []admin.Field{
			UserSegmentFieldsInstance.Name,
		},
		Fieldsets: []admin.Fieldset[UserSegment]{
			{
				Name: "Basic Information",
				Fields: []string{"name", "description", "type"},
			},
			{
				Name: "Criteria",
				Fields: []string{"criteria"},
			},
			{
				Name: "Rules",
				Fields: []string{"rules"},
			},
			{
				Name: "Membership",
				Fields: []string{"customer_ids", "customer_count"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "type", "is_active", "customer_count"},
		},
		ReadOnlyFields: []admin.Field{
			UserSegmentFieldsInstance.CustomerCount,
			UserSegmentFieldsInstance.CreatedAt,
			UserSegmentFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}, obj *UserSegment) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}, obj *UserSegment) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}, obj *UserSegment) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}, obj *UserSegment) bool {
			return true
		},
		Filters: []admin.Filter[UserSegment]{
			{
				Name:  "by_type",
				Label: "By Type",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "static", Label: "Static"},
					{Value: "dynamic", Label: "Dynamic"},
					{Value: "behavioral", Label: "Behavioral"},
					{Value: "predictive", Label: "Predictive"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[UserSegment], value interface{}) orm.QuerySet[UserSegment] {
					return qs.Filter("type", value)
				},
			},
			{
				Name:  "active",
				Label: "Active Segments",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[UserSegment], value interface{}) orm.QuerySet[UserSegment] {
					return qs.Filter("is_active", value)
				},
			},
		},
		Actions: []admin.Action[UserSegment]{
			{
				Name:         "activate",
				Label:        "Activate Segments",
				Icon:        "Check",
				Confirmation: "Are you sure you want to activate selected segments?",
				Handler: func(ctx context.Context, admin *admin.Admin[UserSegment], ids []interface{}) error {
					for _, id := range ids {
						segment, err := UserSegmentObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						segment.IsActive = true
						if err := UserSegmentObjects.Update(ctx, segment); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "deactivate",
				Label:        "Deactivate Segments",
				Icon:        "X",
				Confirmation: "Are you sure you want to deactivate selected segments?",
				Handler: func(ctx context.Context, admin *admin.Admin[UserSegment], ids []interface{}) error {
					for _, id := range ids {
						segment, err := UserSegmentObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						segment.IsActive = false
						if err := UserSegmentObjects.Update(ctx, segment); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "sync_segment",
				Label:        "Sync Segment Members",
				Icon:        "RefreshCw",
				Confirmation: "Sync customers in selected segments?",
				Handler: func(ctx context.Context, admin *admin.Admin[UserSegment], ids []interface{}) error {
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "evaluate_segment",
				Label:        "Evaluate Segment",
				Icon:        "Activity",
				Confirmation: "Run segment evaluation for selected segment?",
				Handler: func(ctx context.Context, admin *admin.Admin[UserSegment], ids []interface{}) error {
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}) bool {
					return true
				},
			},
		},
	})
}

import (
	"context"

	"github.com/forgego/forge/admin"
	"github.com/forgego/forge/orm"
)

// RegisterAdmin registers engagement models with the admin interface
func RegisterAdmin(ctx context.Context) {
	// RecentlyViewed admin
	admin.Register(&admin.Config[RecentlyViewed]{
		Icon: "Eye",
		ListDisplay: []admin.Field{
			RecentlyViewedFieldsInstance.ID,
			RecentlyViewedFieldsInstance.CustomerID,
			RecentlyViewedFieldsInstance.ProductID,
			RecentlyViewedFieldsInstance.ViewedAt,
			RecentlyViewedFieldsInstance.Source,
			RecentlyViewedFieldsInstance.SessionID,
		},
		ListFilter: []admin.Field{
			RecentlyViewedFieldsInstance.CustomerID,
			RecentlyViewedFieldsInstance.Source,
		},
		SearchFields: []admin.Field{
			RecentlyViewedFieldsInstance.CustomerID,
			RecentlyViewedFieldsInstance.GuestID,
			RecentlyViewedFieldsInstance.SessionID,
		},
		Ordering: []admin.Field{
			RecentlyViewedFieldsInstance.ViewedAt,
		},
		Fieldsets: []admin.Fieldset[RecentlyViewed]{
			{
				Name: "Product Information",
				Fields: []string{"customer_id", "guest_id", "product_id", "variant_id"},
			},
			{
				Name: "Tracking",
				Fields: []string{"viewed_at", "session_id", "user_agent", "ip_address"},
			},
			{
				Name: "Context",
				Fields: []string{"source", "referer_url"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"customer_id", "product_id", "viewed_at"},
		},
		ReadOnlyFields: []admin.Field{
			RecentlyViewedFieldsInstance.ViewedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[RecentlyViewed], user interface{}, obj *RecentlyViewed) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[RecentlyViewed], user interface{}, obj *RecentlyViewed) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[RecentlyViewed], user interface{}, obj *RecentlyViewed) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[RecentlyViewed], user interface{}, obj *RecentlyViewed) bool {
			return true
		},
		Filters: []admin.Filter[RecentlyViewed]{
			{
				Name:  "by_customer",
				Label: "By Customer",
				Type:  admin.FilterTypeObjectID,
				Handler: func(ctx context.Context, qs orm.QuerySet[RecentlyViewed], value interface{}) orm.QuerySet[RecentlyViewed] {
					return qs.Filter("customer_id", value)
				},
			},
			{
				Name:  "by_source",
				Label: "By Source",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "search", Label: "Search"},
					{Value: "browse", Label: "Browse"},
					{Value: "category", Label: "Category"},
					{Value: "product", Label: "Product"},
					{Value: "recommended", Label: "Recommended"},
					{Value: "email", Label: "Email"},
					{Value: "social", Label: "Social"},
					{Value: "direct", Label: "Direct"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[RecentlyViewed], value interface{}) orm.QuerySet[RecentlyViewed] {
					return qs.Filter("source", value)
				},
			},
			{
				Name:  "recent",
				Label: "Last 24 Hours",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[RecentlyViewed], value interface{}) orm.QuerySet[RecentlyViewed] {
					return qs.Filter("viewed_at__gte", value)
				},
			},
		},
		Actions: []admin.Action[RecentlyViewed]{
			{
				Name:         "clear_customer_history",
				Label:        "Clear Customer History",
				Icon:        "Trash2",
				Confirmation: "Are you sure you want to clear the recently viewed history for this customer?",
				Handler: func(ctx context.Context, admin *admin.Admin[RecentlyViewed], ids []interface{}) error {
					for _, id := range ids {
						viewed, err := RecentlyViewedObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						if err := RecentlyViewedObjects.Delete(ctx, viewed); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[RecentlyViewed], user interface{}) bool {
					return true
				},
			},
		},
	})

	// ProductComparison admin
	admin.Register(&admin.Config[ProductComparison]{
		Icon: "GitCompare",
		ListDisplay: []admin.Field{
			ProductComparisonFieldsInstance.ID,
			ProductComparisonFieldsInstance.Name,
			ProductComparisonFieldsInstance.CustomerID,
			ProductComparisonFieldsInstance.ProductIDs,
			ProductComparisonFieldsInstance.IsPublic,
			ProductComparisonFieldsInstance.ViewCount,
			ProductComparisonFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			ProductComparisonFieldsInstance.CustomerID,
			ProductComparisonFieldsInstance.IsPublic,
		},
		SearchFields: []admin.Field{
			ProductComparisonFieldsInstance.Name,
			ProductComparisonFieldsInstance.ShareToken,
		},
		Ordering: []admin.Field{
			ProductComparisonFieldsInstance.CreatedAt,
		},
		Fieldsets: []admin.Fieldset[ProductComparison]{
			{
				Name: "Basic Information",
				Fields: []string{"customer_id", "guest_id", "name"},
			},
			{
				Name: "Products",
				Fields: []string{"product_ids"},
			},
			{
				Name: "Settings",
				Fields: []string{"is_public", "share_token"},
			},
			{
				Name: "Statistics",
				Fields: []string{"view_count"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "is_public", "product_ids"},
		},
		ReadOnlyFields: []admin.Field{
			ProductComparisonFieldsInstance.ViewCount,
			ProductComparisonFieldsInstance.CreatedAt,
			ProductComparisonFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[ProductComparison], user interface{}, obj *ProductComparison) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[ProductComparison], user interface{}, obj *ProductComparison) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[ProductComparison], user interface{}, obj *ProductComparison) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[ProductComparison], user interface{}, obj *ProductComparison) bool {
			return true
		},
		Filters: []admin.Filter[ProductComparison]{
			{
				Name:  "by_customer",
				Label: "By Customer",
				Type:  admin.FilterTypeObjectID,
				Handler: func(ctx context.Context, qs orm.QuerySet[ProductComparison], value interface{}) orm.QuerySet[ProductComparison] {
					return qs.Filter("customer_id", value)
				},
			},
			{
				Name:  "public",
				Label: "Public Comparisons",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[ProductComparison], value interface{}) orm.QuerySet[ProductComparison] {
					return qs.Filter("is_public", value)
				},
			},
		},
		Actions: []admin.Action[ProductComparison]{
			{
				Name:         "toggle_public",
				Label:        "Toggle Public/Private",
				Icon:        "Globe",
				Confirmation: "Are you sure you want to change the visibility of the selected comparisons?",
				Handler: func(ctx context.Context, admin *admin.Admin[ProductComparison], ids []interface{}) error {
					for _, id := range ids {
						comparison, err := ProductComparisonObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						comparison.IsPublic = !comparison.IsPublic
						if err := ProductComparisonObjects.Update(ctx, comparison); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ProductComparison], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "reset_view_count",
				Label:        "Reset View Count",
				Icon:        "RotateCcw",
				Confirmation: "Are you sure you want to reset the view count for selected comparisons?",
				Handler: func(ctx context.Context, admin *admin.Admin[ProductComparison], ids []interface{}) error {
					for _, id := range ids {
						comparison, err := ProductComparisonObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						comparison.ViewCount = 0
						if err := ProductComparisonObjects.Update(ctx, comparison); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[ProductComparison], user interface{}) bool {
					return true
				},
			},
		},
	})

	// Notification admin
	admin.Register(&admin.Config[Notification]{
		Icon: "Bell",
		ListDisplay: []admin.Field{
			NotificationFieldsInstance.ID,
			NotificationFieldsInstance.CustomerID,
			NotificationFieldsInstance.Type,
			NotificationFieldsInstance.Title,
			NotificationFieldsInstance.IsRead,
			NotificationFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			NotificationFieldsInstance.CustomerID,
			NotificationFieldsInstance.Type,
			NotificationFieldsInstance.IsRead,
		},
		SearchFields: []admin.Field{
			NotificationFieldsInstance.Title,
			NotificationFieldsInstance.Message,
		},
		Ordering: []admin.Field{
			NotificationFieldsInstance.CreatedAt,
		},
		Fieldsets: []admin.Fieldset[Notification]{
			{
				Name: "Recipient",
				Fields: []string{"customer_id"},
			},
			{
				Name: "Content",
				Fields: []string{"type", "title", "message"},
			},
			{
				Name: "Action",
				Fields: []string{"action_url", "action_text"},
			},
			{
				Name: "Related Entity",
				Fields: []string{"related_type", "related_id"},
			},
			{
				Name: "Status",
				Fields: []string{"is_read", "read_at"},
			},
			{
				Name: "Channels",
				Fields: []string{"push_enabled", "email_enabled", "sms_enabled"},
			},
			{
				Name: "Scheduling",
				Fields: []string{"scheduled_for", "sent_at"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"type", "title", "is_read"},
		},
		ReadOnlyFields: []admin.Field{
			NotificationFieldsInstance.CreatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[Notification], user interface{}, obj *Notification) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[Notification], user interface{}, obj *Notification) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[Notification], user interface{}, obj *Notification) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[Notification], user interface{}, obj *Notification) bool {
			return true
		},
		Filters: []admin.Filter[Notification]{
			{
				Name:  "by_customer",
				Label: "By Customer",
				Type:  admin.FilterTypeObjectID,
				Handler: func(ctx context.Context, qs orm.QuerySet[Notification], value interface{}) orm.QuerySet[Notification] {
					return qs.Filter("customer_id", value)
				},
			},
			{
				Name:  "by_type",
				Label: "By Type",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "order_confirmation", Label: "Order Confirmation"},
					{Value: "order_shipped", Label: "Order Shipped"},
					{Value: "order_delivered", Label: "Order Delivered"},
					{Value: "order_cancelled", Label: "Order Cancelled"},
					{Value: "payment_received", Label: "Payment Received"},
					{Value: "payment_failed", Label: "Payment Failed"},
					{Value: "review_reminder", Label: "Review Reminder"},
					{Value: "wishlist_price_drop", Label: "Wishlist Price Drop"},
					{Value: "back_in_stock", Label: "Back in Stock"},
					{Value: "abandoned_cart", Label: "Abandoned Cart"},
					{Value: "coupon_available", Label: "Coupon Available"},
					{Value: "account_security", Label: "Account Security"},
					{Value: "system", Label: "System"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[Notification], value interface{}) orm.QuerySet[Notification] {
					return qs.Filter("type", value)
				},
			},
			{
				Name:  "unread",
				Label: "Unread Only",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[Notification], value interface{}) orm.QuerySet[Notification] {
					return qs.Filter("is_read", false)
				},
			},
		},
		Actions: []admin.Action[Notification]{
			{
				Name:         "mark_read",
				Label:        "Mark as Read",
				Icon:        "Check",
				Confirmation: "Are you sure you want to mark the selected notifications as read?",
				Handler: func(ctx context.Context, admin *admin.Admin[Notification], ids []interface{}) error {
					for _, id := range ids {
						notification, err := NotificationObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						now := time.Now()
						notification.IsRead = true
						notification.ReadAt = &now
						if err := NotificationObjects.Update(ctx, notification); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[Notification], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "mark_all_read",
				Label:        "Mark All as Read",
				Icon:        "CheckCheck",
				Confirmation: "Are you sure you want to mark ALL notifications as read for a customer?",
				Handler: func(ctx context.Context, admin *admin.Admin[Notification], ids []interface{}) error {
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[Notification], user interface{}) bool {
					return true
				},
			},
		},
	})

	// CustomerActivity admin
	admin.Register(&admin.Config[CustomerActivity]{
		Icon: "Activity",
		ListDisplay: []admin.Field{
			CustomerActivityFieldsInstance.ID,
			CustomerActivityFieldsInstance.CustomerID,
			CustomerActivityFieldsInstance.ActivityType,
			CustomerActivityFieldsInstance.EntityType,
			CustomerActivityFieldsInstance.SessionID,
			CustomerActivityFieldsInstance.Timestamp,
		},
		ListFilter: []admin.Field{
			CustomerActivityFieldsInstance.CustomerID,
			CustomerActivityFieldsInstance.ActivityType,
			CustomerActivityFieldsInstance.EntityType,
		},
		SearchFields: []admin.Field{
			CustomerActivityFieldsInstance.SessionID,
		},
		Ordering: []admin.Field{
			CustomerActivityFieldsInstance.Timestamp,
		},
		Fieldsets: []admin.Fieldset[CustomerActivity]{
			{
				Name: "Activity",
				Fields: []string{"customer_id", "activity_type"},
			},
			{
				Name: "Entity",
				Fields: []string{"entity_type", "entity_id"},
			},
			{
				Name: "Data",
				Fields: []string{"data"},
			},
			{
				Name: "Context",
				Fields: []string{"session_id", "user_agent", "ip_address"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"activity_type", "entity_type"},
		},
		ReadOnlyFields: []admin.Field{
			CustomerActivityFieldsInstance.Timestamp,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[CustomerActivity], user interface{}, obj *CustomerActivity) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[CustomerActivity], user interface{}, obj *CustomerActivity) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[CustomerActivity], user interface{}, obj *CustomerActivity) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[CustomerActivity], user interface{}, obj *CustomerActivity) bool {
			return true
		},
		Filters: []admin.Filter[CustomerActivity]{
			{
				Name:  "by_customer",
				Label: "By Customer",
				Type:  admin.FilterTypeObjectID,
				Handler: func(ctx context.Context, qs orm.QuerySet[CustomerActivity], value interface{}) orm.QuerySet[CustomerActivity] {
					return qs.Filter("customer_id", value)
				},
			},
			{
				Name:  "by_activity_type",
				Label: "By Activity Type",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "page_view", Label: "Page View"},
					{Value: "product_view", Label: "Product View"},
					{Value: "add_to_cart", Label: "Add to Cart"},
					{Value: "remove_from_cart", Label: "Remove from Cart"},
					{Value: "checkout_start", Label: "Checkout Start"},
					{Value: "checkout_complete", Label: "Checkout Complete"},
					{Value: "search", Label: "Search"},
					{Value: "wishlist_add", Label: "Wishlist Add"},
					{Value: "wishlist_remove", Label: "Wishlist Remove"},
					{Value: "review_submit", Label: "Review Submit"},
					{Value: "login", Label: "Login"},
					{Value: "logout", Label: "Logout"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[CustomerActivity], value interface{}) orm.QuerySet[CustomerActivity] {
					return qs.Filter("activity_type", value)
				},
			},
		},
		Actions: []admin.Action[CustomerActivity]{
			{
				Name:         "export_activities",
				Label:        "Export Activities",
				Icon:        "Download",
				Confirmation: "Export customer activities to CSV?",
				Handler: func(ctx context.Context, admin *admin.Admin[CustomerActivity], ids []interface{}) error {
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[CustomerActivity], user interface{}) bool {
					return true
				},
			},
		},
	})

	// AbandonedCartReminder admin
	admin.Register(&admin.Config[AbandonedCartReminder]{
		Icon: "ShoppingCart",
		ListDisplay: []admin.Field{
			AbandonedCartReminderFieldsInstance.ID,
			AbandonedCartReminderFieldsInstance.CartID,
			AbandonedCartReminderFieldsInstance.CustomerID,
			AbandonedCartReminderFieldsInstance.ReminderNumber,
			AbandonedCartReminderFieldsInstance.Status,
			AbandonedCartReminderFieldsInstance.ReminderSentAt,
			AbandonedCartReminderFieldsInstance.RecoveredAt,
		},
		ListFilter: []admin.Field{
			AbandonedCartReminderFieldsInstance.CustomerID,
			AbandonedCartReminderFieldsInstance.Status,
			AbandonedCartReminderFieldsInstance.ReminderNumber,
		},
		SearchFields: []admin.Field{
			AbandonedCartReminderFieldsInstance.GuestEmail,
		},
		Ordering: []admin.Field{
			AbandonedCartReminderFieldsInstance.CreatedAt,
		},
		Fieldsets: []admin.Fieldset[AbandonedCartReminder]{
			{
				Name: "Cart Information",
				Fields: []string{"cart_id", "customer_id", "guest_email"},
			},
			{
				Name: "Reminder Details",
				Fields: []string{"reminder_number", "reminder_sent_at", "reminder_opened_at", "reminder_clicked_at"},
			},
			{
				Name: "Recovery",
				Fields: []string{"recovered_at", "recovered_order_id"},
			},
			{
				Name: "Status",
				Fields: []string{"status"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"status", "reminder_number"},
		},
		ReadOnlyFields: []admin.Field{
			AbandonedCartReminderFieldsInstance.CreatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], user interface{}, obj *AbandonedCartReminder) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], user interface{}, obj *AbandonedCartReminder) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], user interface{}, obj *AbandonedCartReminder) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], user interface{}, obj *AbandonedCartReminder) bool {
			return true
		},
		Filters: []admin.Filter[AbandonedCartReminder]{
			{
				Name:  "by_customer",
				Label: "By Customer",
				Type:  admin.FilterTypeObjectID,
				Handler: func(ctx context.Context, qs orm.QuerySet[AbandonedCartReminder], value interface{}) orm.QuerySet[AbandonedCartReminder] {
					return qs.Filter("customer_id", value)
				},
			},
			{
				Name:  "by_status",
				Label: "By Status",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "pending", Label: "Pending"},
					{Value: "sent", Label: "Sent"},
					{Value: "opened", Label: "Opened"},
					{Value: "clicked", Label: "Clicked"},
					{Value: "recovered", Label: "Recovered"},
					{Value: "cancelled", Label: "Cancelled"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[AbandonedCartReminder], value interface{}) orm.QuerySet[AbandonedCartReminder] {
					return qs.Filter("status", value)
				},
			},
			{
				Name:  "recovered",
				Label: "Recovered Carts",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[AbandonedCartReminder], value interface{}) orm.QuerySet[AbandonedCartReminder] {
					return qs.Filter("status", "recovered")
				},
			},
		},
		Actions: []admin.Action[AbandonedCartReminder]{
			{
				Name:         "send_reminder",
				Label:        "Send Reminder",
				Icon:        "Send",
				Confirmation: "Are you sure you want to send reminders to selected carts?",
				Handler: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], ids []interface{}) error {
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "cancel_reminder",
				Label:        "Cancel Reminder",
				Icon:        "X",
				Confirmation: "Are you sure you want to cancel selected reminders?",
				Handler: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], ids []interface{}) error {
					for _, id := range ids {
						reminder, err := AbandonedCartReminderObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						reminder.Status = "cancelled"
						if err := AbandonedCartReminderObjects.Update(ctx, reminder); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[AbandonedCartReminder], user interface{}) bool {
					return true
				},
			},
		},
	})

	// UserSegment admin
	admin.Register(&admin.Config[UserSegment]{
		Icon: "Users",
		ListDisplay: []admin.Field{
			UserSegmentFieldsInstance.ID,
			UserSegmentFieldsInstance.Name,
			UserSegmentFieldsInstance.Type,
			UserSegmentFieldsInstance.CustomerCount,
			UserSegmentFieldsInstance.IsActive,
			UserSegmentFieldsInstance.CreatedAt,
		},
		ListFilter: []admin.Field{
			UserSegmentFieldsInstance.Type,
			UserSegmentFieldsInstance.IsActive,
		},
		SearchFields: []admin.Field{
			UserSegmentFieldsInstance.Name,
			UserSegmentFieldsInstance.Description,
		},
		Ordering: []admin.Field{
			UserSegmentFieldsInstance.Name,
		},
		Fieldsets: []admin.Fieldset[UserSegment]{
			{
				Name: "Basic Information",
				Fields: []string{"name", "description", "type"},
			},
			{
				Name: "Criteria",
				Fields: []string{"criteria"},
			},
			{
				Name: "Rules",
				Fields: []string{"rules"},
			},
			{
				Name: "Membership",
				Fields: []string{"customer_ids", "customer_count"},
			},
			{
				Name: "Status",
				Fields: []string{"is_active"},
			},
		},
		HistoryManager: &admin.HistoryManager{
			TrackFields: []string{"name", "type", "is_active", "customer_count"},
		},
		ReadOnlyFields: []admin.Field{
			UserSegmentFieldsInstance.CustomerCount,
			UserSegmentFieldsInstance.CreatedAt,
			UserSegmentFieldsInstance.UpdatedAt,
		},
		HasViewPermission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}, obj *UserSegment) bool {
			return true
		},
		HasAddPermission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}, obj *UserSegment) bool {
			return true
		},
		HasChangePermission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}, obj *UserSegment) bool {
			return true
		},
		HasDeletePermission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}, obj *UserSegment) bool {
			return true
		},
		Filters: []admin.Filter[UserSegment]{
			{
				Name:  "by_type",
				Label: "By Type",
				Type:  admin.FilterTypeChoice,
				Choices: []admin.Choice{
					{Value: "static", Label: "Static"},
					{Value: "dynamic", Label: "Dynamic"},
					{Value: "behavioral", Label: "Behavioral"},
					{Value: "predictive", Label: "Predictive"},
				},
				Handler: func(ctx context.Context, qs orm.QuerySet[UserSegment], value interface{}) orm.QuerySet[UserSegment] {
					return qs.Filter("type", value)
				},
			},
			{
				Name:  "active",
				Label: "Active Segments",
				Type:  admin.FilterTypeBoolean,
				Handler: func(ctx context.Context, qs orm.QuerySet[UserSegment], value interface{}) orm.QuerySet[UserSegment] {
					return qs.Filter("is_active", value)
				},
			},
		},
		Actions: []admin.Action[UserSegment]{
			{
				Name:         "activate",
				Label:        "Activate Segments",
				Icon:        "Check",
				Confirmation: "Are you sure you want to activate selected segments?",
				Handler: func(ctx context.Context, admin *admin.Admin[UserSegment], ids []interface{}) error {
					for _, id := range ids {
						segment, err := UserSegmentObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						segment.IsActive = true
						if err := UserSegmentObjects.Update(ctx, segment); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "deactivate",
				Label:        "Deactivate Segments",
				Icon:        "X",
				Confirmation: "Are you sure you want to deactivate selected segments?",
				Handler: func(ctx context.Context, admin *admin.Admin[UserSegment], ids []interface{}) error {
					for _, id := range ids {
						segment, err := UserSegmentObjects.Get(ctx, id)
						if err != nil {
							return err
						}
						segment.IsActive = false
						if err := UserSegmentObjects.Update(ctx, segment); err != nil {
							return err
						}
					}
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "sync_segment",
				Label:        "Sync Segment Members",
				Icon:        "RefreshCw",
				Confirmation: "Sync customers in selected segments?",
				Handler: func(ctx context.Context, admin *admin.Admin[UserSegment], ids []interface{}) error {
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}) bool {
					return true
				},
			},
			{
				Name:         "evaluate_segment",
				Label:        "Evaluate Segment",
				Icon:        "Activity",
				Confirmation: "Run segment evaluation for selected segment?",
				Handler: func(ctx context.Context, admin *admin.Admin[UserSegment], ids []interface{}) error {
					return nil
				},
				Permission: func(ctx context.Context, admin *admin.Admin[UserSegment], user interface{}) bool {
					return true
				},
			},
		},
	})
}

