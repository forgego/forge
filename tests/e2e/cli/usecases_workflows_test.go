package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func commerceJourney(t *testing.T, app *journeyApp, client *journeyClient, product int64) {
	customer := createConsumer(t, client, "customers", map[string]any{"name": "Buyer", "email": "buyer@example.com"})
	order := createConsumer(t, client, "orders", map[string]any{"customer_id": customer, "owner_id": 1, "total_cents": 5997})
	line := createConsumer(t, client, "order-lines", map[string]any{"order_id": order, "product_id": product, "quantity": 3, "unit_price_cents": 1999})
	assert.Equal(t, 1, app.count(`SELECT count(*) FROM orders o JOIN order_lines l ON l.order_id=o.id WHERE o.id=$1 AND o.total_cents=l.quantity*l.unit_price_cents`, order))
	status, body := client.do(http.MethodPost, "/order-lines/", "token-alice", map[string]any{"order_id": order, "product_id": product, "quantity": -1, "unit_price_cents": 1999})
	assert.Equal(t, http.StatusBadRequest, status, "%v", body)
	status, body = client.do(http.MethodPost, "/order-lines/", "token-alice", map[string]any{"order_id": order, "product_id": 999999, "quantity": 1, "unit_price_cents": 1999})
	assert.Equal(t, http.StatusBadRequest, status, "%v", body)
	assert.Zero(t, app.count(`SELECT count(*) FROM order_lines WHERE product_id=999999`))
	status, body = client.do(http.MethodDelete, fmt.Sprintf("/products/%d", product), "token-alice", nil)
	assert.Equal(t, http.StatusBadRequest, status, "referenced product is protected: %v", body)
	checkOwnedRecord(t, app, client, "orders", "orders", order)
	status, _ = client.do(http.MethodDelete, fmt.Sprintf("/orders/%d", order), "token-alice", nil)
	require.Equal(t, http.StatusNoContent, status)
	assert.Zero(t, app.count(`SELECT count(*) FROM order_lines WHERE id=$1`, line), "dependent line cascades")
}

func crmJourney(t *testing.T, app *journeyApp, client *journeyClient, organization int64) {
	contact := createConsumer(t, client, "contacts", map[string]any{"organization_id": organization, "name": "Customer contact", "email": "contact@example.com"})
	ticket := createConsumer(t, client, "tickets", map[string]any{"contact_id": contact, "owner_id": 1, "title": "Cannot sign in", "priority": 0})
	comment := createConsumer(t, client, "comments", map[string]any{"ticket_id": ticket, "body": "Investigating the sign-in problem"})
	status, body := client.do(http.MethodGet, fmt.Sprintf("/tickets/?contact_id=%d&priority=0", contact), "token-alice", nil)
	require.Equal(t, http.StatusOK, status)
	assert.EqualValues(t, 1, body["count"])
	checkOwnedRecord(t, app, client, "tickets", "tickets", ticket)
	status, body = client.do(http.MethodPatch, fmt.Sprintf("/tickets/%d", ticket), "token-alice", map[string]any{"status": "closed"})
	require.Equal(t, http.StatusOK, status, "%v", body)
	assert.Equal(t, "closed", body["status"])
	assert.Equal(t, 1, app.count(`SELECT count(*) FROM comments c JOIN tickets t ON t.id=c.ticket_id WHERE c.id=$1 AND t.status='closed'`, comment))
}

func bookingJourney(t *testing.T, app *journeyApp, client *journeyClient, resource int64) {
	start, end := "2030-01-02T10:00:00Z", "2030-01-02T11:00:00Z"
	slot := createConsumer(t, client, "slots", map[string]any{"resource_id": resource, "starts_at": start, "ends_at": end})
	status, body := client.do(http.MethodPost, "/slots/", "token-alice", map[string]any{"resource_id": resource, "starts_at": end, "ends_at": start})
	assert.Equal(t, http.StatusBadRequest, status, "CHECK must reject reversed interval: %v", body)
	status, body = client.do(http.MethodPost, "/slots/", "token-alice", map[string]any{"resource_id": resource, "starts_at": start, "ends_at": end})
	assert.Equal(t, http.StatusConflict, status, "composite unique slot: %v", body)
	// Separate clients avoid shared response-header state between goroutines.
	statuses := make(chan int, 2)
	var wg sync.WaitGroup
	for i, who := range []string{"alice", "bob"} {
		wg.Add(1)
		go func(i int, who string) {
			defer wg.Done()
			payload, err := json.Marshal(map[string]any{"slot_id": slot, "owner_id": i + 1, "reference": "RES-" + who})
			if err != nil {
				statuses <- 0
				return
			}
			request, err := http.NewRequestWithContext(app.ctx, http.MethodPost, client.base+"/bookings/", bytes.NewReader(payload))
			if err != nil {
				statuses <- 0
				return
			}
			request.Header.Set("Authorization", "Token token-"+who)
			request.Header.Set("Content-Type", "application/json")
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				statuses <- 0
				return
			}
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			statuses <- response.StatusCode
		}(i, who)
	}
	wg.Wait()
	close(statuses)
	var got []int
	for status := range statuses {
		got = append(got, status)
	}
	assert.ElementsMatch(t, []int{http.StatusCreated, http.StatusConflict}, got)
	assert.Equal(t, 1, app.count(`SELECT count(*) FROM bookings WHERE slot_id=$1`, slot))
}

func cmsJourney(t *testing.T, app *journeyApp, client *journeyClient, author int64) {
	post := createConsumer(t, client, "posts", map[string]any{"author_id": author, "owner_id": 1, "title": "A Persian and Unicode story", "slug": "unicode-story", "body": "سلام دنیا — café", "published": false})
	tag := createConsumer(t, client, "tags", map[string]any{"name": "Engineering"})
	createConsumer(t, client, "post-tags", map[string]any{"post_id": post, "tag_id": tag})
	status, body := client.do(http.MethodPost, "/post-tags/", "token-alice", map[string]any{"post_id": post, "tag_id": tag})
	assert.Equal(t, http.StatusConflict, status, "%v", body)
	createConsumer(t, client, "revisions", map[string]any{"post_id": post, "number": 1, "body": "سلام دنیا — café"})
	checkOwnedRecord(t, app, client, "posts", "posts", post)
	status, body = client.do(http.MethodPatch, fmt.Sprintf("/posts/%d", post), "token-alice", map[string]any{"published": true, "published_at": "2030-01-02T10:00:00Z"})
	require.Equal(t, http.StatusOK, status, "%v", body)
	assert.Equal(t, true, body["published"])
	assert.Equal(t, "سلام دنیا — café", body["body"])
	assert.Equal(t, 1, app.count(`SELECT count(*) FROM posts p JOIN post_tags pt ON p.id=pt.post_id JOIN tags t ON t.id=pt.tag_id WHERE p.id=$1 AND t.name='Engineering' AND p.published`, post))
}

func expensesJourney(t *testing.T, app *journeyApp, client *journeyClient, employee int64) {
	report := createConsumer(t, client, "expense-reports", map[string]any{"employee_id": employee, "owner_id": 1, "title": "Conference expenses", "total_cents": 10001})
	for _, amount := range []int{3333, 6668} {
		createConsumer(t, client, "expenses", map[string]any{"report_id": report, "amount_cents": amount, "currency": "USD", "description": "Travel and supplies"})
	}
	assert.Equal(t, 1, app.count(`SELECT count(*) FROM expense_reports r WHERE r.id=$1 AND r.total_cents=(SELECT sum(amount_cents) FROM expenses e WHERE e.report_id=r.id)`, report))
	checkOwnedRecord(t, app, client, "expense-reports", "expense_reports", report)
	status, body := client.do(http.MethodPost, "/approvals/", "token-alice", map[string]any{"report_id": report, "approver_id": 1, "decision": "approved"})
	assert.Equal(t, http.StatusForbidden, status, "employee cannot approve own report: %v", body)
	assert.Zero(t, app.count(`SELECT count(*) FROM approvals`))
	status, body = client.do(http.MethodPost, "/approvals/", "token-staff", map[string]any{"report_id": report, "approver_id": 3, "decision": "approved", "note": "Receipts checked"})
	require.Equal(t, http.StatusCreated, status, "%v", body)
	assert.Equal(t, 1, app.count(`SELECT count(*) FROM approvals WHERE report_id=$1 AND approver_id=3 AND decision='approved'`, report))
}

func checkOwnedRecord(t *testing.T, app *journeyApp, client *journeyClient, route, table string, id int64) {
	t.Helper()
	for _, method := range []string{http.MethodPatch, http.MethodDelete} {
		var payload map[string]any
		if method == http.MethodPatch {
			payload = map[string]any{"owner_id": 2}
		}
		status, body := client.do(method, fmt.Sprintf("/%s/%d", route, id), "token-bob", payload)
		assert.Equal(t, http.StatusForbidden, status, "%v", body)
		assert.Equal(t, 1, app.count("SELECT count(*) FROM "+table+" WHERE id=$1 AND owner_id=1", id))
	}
	for _, key := range []string{"owner_id", "OwnerId"} {
		status, body := client.do(http.MethodPatch, fmt.Sprintf("/%s/%d", route, id), "token-alice", map[string]any{key: 2})
		assert.Equal(t, http.StatusForbidden, status, "owner cannot transfer ownership through %s: %v", key, body)
		assert.Equal(t, 1, app.count("SELECT count(*) FROM "+table+" WHERE id=$1 AND owner_id=1", id))
	}
}
