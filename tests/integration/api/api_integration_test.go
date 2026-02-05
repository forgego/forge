package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"testing"

	forgeapi "github.com/forgego/forge/api"
	apitesting "github.com/forgego/forge/api/testing"
	forgehttp "github.com/forgego/forge/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sampleItem struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
	IsActive bool    `json:"is_active"`
}

type sampleStore struct {
	items  []sampleItem
	nextID int
}

func newSampleStore(seed []sampleItem) *sampleStore {
	items := make([]sampleItem, len(seed))
	copy(items, seed)
	maxID := 0
	for _, item := range items {
		if item.ID > maxID {
			maxID = item.ID
		}
	}
	return &sampleStore{
		items:  items,
		nextID: maxID + 1,
	}
}

func (s *sampleStore) list() []sampleItem {
	items := make([]sampleItem, len(s.items))
	copy(items, s.items)
	return items
}

func (s *sampleStore) find(id int) (sampleItem, bool) {
	for _, item := range s.items {
		if item.ID == id {
			return item, true
		}
	}
	return sampleItem{}, false
}

func (s *sampleStore) create(item sampleItem) sampleItem {
	item.ID = s.nextID
	s.nextID++
	s.items = append(s.items, item)
	return item
}

func (s *sampleStore) update(id int, updated sampleItem) (sampleItem, bool) {
	for i, item := range s.items {
		if item.ID == id {
			updated.ID = id
			s.items[i] = updated
			return updated, true
		}
	}
	return sampleItem{}, false
}

func (s *sampleStore) patch(id int, payload itemPayload) (sampleItem, bool) {
	for i, item := range s.items {
		if item.ID == id {
			if payload.Name != nil {
				item.Name = strings.TrimSpace(*payload.Name)
			}
			if payload.Category != nil {
				item.Category = strings.TrimSpace(*payload.Category)
			}
			if payload.Price != nil {
				item.Price = *payload.Price
			}
			if payload.IsActive != nil {
				item.IsActive = *payload.IsActive
			}
			s.items[i] = item
			return item, true
		}
	}
	return sampleItem{}, false
}

func (s *sampleStore) delete(id int) bool {
	for i, item := range s.items {
		if item.ID == id {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return true
		}
	}
	return false
}

type itemPayload struct {
	Name     *string  `json:"name"`
	Category *string  `json:"category"`
	Price    *float64 `json:"price"`
	IsActive *bool    `json:"is_active"`
}

type sampleViewSet struct {
	store *sampleStore
}

func (vs *sampleViewSet) List(w http.ResponseWriter, r *http.Request) {
	items := vs.store.list()

	category := strings.TrimSpace(r.URL.Query().Get("category"))
	if category != "" {
		filtered := make([]sampleItem, 0, len(items))
		for _, item := range items {
			if strings.EqualFold(item.Category, category) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	if isActiveRaw := r.URL.Query().Get("is_active"); isActiveRaw != "" {
		isActive, err := strconv.ParseBool(isActiveRaw)
		if err == nil {
			filtered := make([]sampleItem, 0, len(items))
			for _, item := range items {
				if item.IsActive == isActive {
					filtered = append(filtered, item)
				}
			}
			items = filtered
		}
	}

	if priceMinRaw := r.URL.Query().Get("price_gte"); priceMinRaw != "" {
		if priceMin, err := strconv.ParseFloat(priceMinRaw, 64); err == nil {
			filtered := make([]sampleItem, 0, len(items))
			for _, item := range items {
				if item.Price >= priceMin {
					filtered = append(filtered, item)
				}
			}
			items = filtered
		}
	}

	if search := strings.TrimSpace(r.URL.Query().Get("search")); search != "" {
		filtered := make([]sampleItem, 0, len(items))
		search = strings.ToLower(search)
		for _, item := range items {
			if strings.Contains(strings.ToLower(item.Name), search) {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}

	ordering := strings.TrimSpace(r.URL.Query().Get("ordering"))
	if ordering != "" {
		desc := strings.HasPrefix(ordering, "-")
		field := strings.TrimPrefix(ordering, "-")
		sort.Slice(items, func(i, j int) bool {
			if field == "price" {
				if desc {
					return items[i].Price > items[j].Price
				}
				return items[i].Price < items[j].Price
			}
			if desc {
				return items[i].Name > items[j].Name
			}
			return items[i].Name < items[j].Name
		})
	}

	page, pageSize, offset := forgeapi.ParsePaginationParams(r, 2)
	totalCount := len(items)
	if offset > totalCount {
		offset = totalCount
	}
	end := offset + pageSize
	if end > totalCount {
		end = totalCount
	}
	items = items[offset:end]

	_ = forgeapi.SendPaginatedResponse(w, r, items, totalCount, page, pageSize)
}

func (vs *sampleViewSet) Create(w http.ResponseWriter, r *http.Request) {
	payload, ok := decodePayload(w, r)
	if !ok {
		return
	}

	item, ok := buildItemFromPayload(payload, true)
	if !ok {
		_ = forgehttp.SendError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	created := vs.store.create(item)
	_ = forgehttp.SendJSON(w, http.StatusCreated, created)
}

func (vs *sampleViewSet) Retrieve(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	item, found := vs.store.find(id)
	if !found {
		_ = forgehttp.SendError(w, http.StatusNotFound, "Item not found")
		return
	}

	_ = forgehttp.SendJSON(w, http.StatusOK, item)
}

func (vs *sampleViewSet) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	payload, ok := decodePayload(w, r)
	if !ok {
		return
	}

	item, ok := buildItemFromPayload(payload, true)
	if !ok {
		_ = forgehttp.SendError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	updated, found := vs.store.update(id, item)
	if !found {
		_ = forgehttp.SendError(w, http.StatusNotFound, "Item not found")
		return
	}

	_ = forgehttp.SendJSON(w, http.StatusOK, updated)
}

func (vs *sampleViewSet) PartialUpdate(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	payload, ok := decodePayload(w, r)
	if !ok {
		return
	}

	if !hasAnyField(payload) {
		_ = forgehttp.SendError(w, http.StatusBadRequest, "No fields to update")
		return
	}

	patched, found := vs.store.patch(id, payload)
	if !found {
		_ = forgehttp.SendError(w, http.StatusNotFound, "Item not found")
		return
	}

	_ = forgehttp.SendJSON(w, http.StatusOK, patched)
}

func (vs *sampleViewSet) Destroy(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if !vs.store.delete(id) {
		_ = forgehttp.SendError(w, http.StatusNotFound, "Item not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	idRaw := forgehttp.GetParam(r, "id")
	id, err := strconv.Atoi(idRaw)
	if err != nil {
		_ = forgehttp.SendError(w, http.StatusBadRequest, "Invalid id")
		return 0, false
	}
	return id, true
}

func decodePayload(w http.ResponseWriter, r *http.Request) (itemPayload, bool) {
	var payload itemPayload
	if err := forgehttp.GetJSON(r, &payload); err != nil {
		_ = forgehttp.SendError(w, http.StatusBadRequest, "Invalid JSON")
		return itemPayload{}, false
	}
	return payload, true
}

func buildItemFromPayload(payload itemPayload, requireAll bool) (sampleItem, bool) {
	if requireAll && (payload.Name == nil || payload.Category == nil || payload.Price == nil) {
		return sampleItem{}, false
	}
	item := sampleItem{}
	if payload.Name != nil {
		item.Name = strings.TrimSpace(*payload.Name)
	}
	if payload.Category != nil {
		item.Category = strings.TrimSpace(*payload.Category)
	}
	if payload.Price != nil {
		item.Price = *payload.Price
	}
	if payload.IsActive != nil {
		item.IsActive = *payload.IsActive
	}
	if requireAll && (item.Name == "" || item.Category == "") {
		return sampleItem{}, false
	}
	return item, true
}

func hasAnyField(payload itemPayload) bool {
	return payload.Name != nil || payload.Category != nil || payload.Price != nil || payload.IsActive != nil
}

type sampleApp struct {
	client  *apitesting.APIClient
	handler http.Handler
}

func newSampleApp() *sampleApp {
	seed := []sampleItem{
		{ID: 1, Name: "Laptop Pro", Category: "Electronics", Price: 1999.99, IsActive: true},
		{ID: 2, Name: "Laptop Air", Category: "Electronics", Price: 1299.99, IsActive: true},
		{ID: 3, Name: "Coffee Grinder", Category: "Kitchen", Price: 89.99, IsActive: true},
		{ID: 4, Name: "Blender Plus", Category: "Kitchen", Price: 149.99, IsActive: false},
		{ID: 5, Name: "Noise Cancelling Headphones", Category: "Electronics", Price: 299.99, IsActive: true},
	}

	store := newSampleStore(seed)
	viewset := &sampleViewSet{store: store}

	apiRouter := forgeapi.NewRouter("/api/v1")
	apiRouter.Register("items", viewset)

	httpRouter := forgehttp.NewRouter()
	apiRouter.RegisterRoutes(httpRouter)

	return &sampleApp{
		client:  apitesting.NewAPIClient(httpRouter),
		handler: httpRouter,
	}
}

type paginatedItemsResponse struct {
	Count   int          `json:"count"`
	Results []sampleItem `json:"results"`
}

func decodePaginatedItems(t *testing.T, response *apitesting.Response) paginatedItemsResponse {
	t.Helper()
	var payload paginatedItemsResponse
	err := json.Unmarshal(response.Body, &payload)
	require.NoError(t, err)
	return payload
}

func decodeItem(t *testing.T, response *apitesting.Response) sampleItem {
	t.Helper()
	var payload sampleItem
	err := json.Unmarshal(response.Body, &payload)
	require.NoError(t, err)
	return payload
}

func decodeProblem(t *testing.T, response *apitesting.Response) map[string]interface{} {
	t.Helper()
	var payload map[string]interface{}
	err := json.Unmarshal(response.Body, &payload)
	require.NoError(t, err)
	return payload
}

func rawRequest(t *testing.T, handler http.Handler, method, path, body string) *apitesting.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return &apitesting.Response{
		StatusCode: recorder.Code,
		Body:       recorder.Body.Bytes(),
		Headers:    recorder.Header(),
	}
}

func TestAPI_ListFiltersPagination(t *testing.T) {
	app := newSampleApp()

	response := app.client.Get("/api/v1/items/?category=Electronics&is_active=true&price_gte=200&ordering=-price&page=1&page_size=2")
	require.Equal(t, http.StatusOK, response.Status())

	payload := decodePaginatedItems(t, response)
	assert.Equal(t, 3, payload.Count)
	require.Len(t, payload.Results, 2)

	for _, item := range payload.Results {
		assert.Equal(t, "Electronics", item.Category)
		assert.True(t, item.IsActive)
		assert.GreaterOrEqual(t, item.Price, 200.0)
	}
	assert.GreaterOrEqual(t, payload.Results[0].Price, payload.Results[1].Price)

	searchResponse := app.client.Get("/api/v1/items/?search=grinder&page=1&page_size=5")
	require.Equal(t, http.StatusOK, searchResponse.Status())
	searchPayload := decodePaginatedItems(t, searchResponse)
	require.Len(t, searchPayload.Results, 1)
	assert.Equal(t, "Coffee Grinder", searchPayload.Results[0].Name)
}

func TestAPI_CRUDFlow(t *testing.T) {
	app := newSampleApp()

	createPayload := map[string]interface{}{
		"name":      "Standing Desk",
		"category":  "Office",
		"price":     399.50,
		"is_active": true,
	}
	createResponse := app.client.Post("/api/v1/items/", createPayload)
	require.Equal(t, http.StatusCreated, createResponse.Status())
	created := decodeItem(t, createResponse)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "Standing Desk", created.Name)

	detailResponse := app.client.Get("/api/v1/items/" + strconv.Itoa(created.ID))
	require.Equal(t, http.StatusOK, detailResponse.Status())
	detail := decodeItem(t, detailResponse)
	assert.Equal(t, created.ID, detail.ID)

	updatePayload := map[string]interface{}{
		"name":      "Standing Desk XL",
		"category":  "Office",
		"price":     499.00,
		"is_active": false,
	}
	updateResponse := app.client.Put("/api/v1/items/"+strconv.Itoa(created.ID), updatePayload)
	require.Equal(t, http.StatusOK, updateResponse.Status())
	updated := decodeItem(t, updateResponse)
	assert.Equal(t, "Standing Desk XL", updated.Name)
	assert.False(t, updated.IsActive)

	patchPayload := map[string]interface{}{
		"price": 459.00,
	}
	patchResponse := app.client.Patch("/api/v1/items/"+strconv.Itoa(created.ID), patchPayload)
	require.Equal(t, http.StatusOK, patchResponse.Status())
	patched := decodeItem(t, patchResponse)
	assert.Equal(t, 459.00, patched.Price)

	deleteResponse := app.client.Delete("/api/v1/items/" + strconv.Itoa(created.ID))
	require.Equal(t, http.StatusNoContent, deleteResponse.Status())

	missingResponse := app.client.Get("/api/v1/items/" + strconv.Itoa(created.ID))
	require.Equal(t, http.StatusNotFound, missingResponse.Status())
}

func TestAPI_ErrorResponses(t *testing.T) {
	app := newSampleApp()

	notFoundResponse := app.client.Get("/api/v1/items/9999")
	require.Equal(t, http.StatusNotFound, notFoundResponse.Status())
	assert.Equal(t, "application/problem+json", notFoundResponse.Headers.Get("Content-Type"))
	problem := decodeProblem(t, notFoundResponse)
	assert.Equal(t, float64(http.StatusNotFound), problem["status"])
	assert.Equal(t, "Item not found", problem["detail"])
	assert.NotEmpty(t, problem["type"])
	assert.NotEmpty(t, problem["title"])
	assert.NotEmpty(t, problem["code"])

	invalidResponse := rawRequest(t, app.handler, http.MethodPost, "/api/v1/items/", "{")
	require.Equal(t, http.StatusBadRequest, invalidResponse.Status())
	invalidProblem := decodeProblem(t, invalidResponse)
	assert.Equal(t, float64(http.StatusBadRequest), invalidProblem["status"])
	assert.Equal(t, "Invalid JSON", invalidProblem["detail"])
}
