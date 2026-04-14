package handlers

import (
	"callsign/middleware"
	"callsign/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// =====================
// Messaging (SMS/MMS)
// =====================

func (h *Handler) ListConversations(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var conversations []models.Conversation
	h.DB.Where("tenant_id = ?", tenantID).
		Order("last_message_at DESC").
		Limit(50).
		Find(&conversations)

	return c.JSON(conversations)
}

func (h *Handler) GetConversation(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var conversation models.Conversation
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Messages").
		First(&conversation).Error; err != nil {
		h.logWarn("MESSAGING", "GetConversation: Conversation not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Conversation not found"})
	}

	return c.JSON(conversation)
}

func (h *Handler) SendMessage(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var msg models.Message
	if err := c.BodyParser(&msg); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	msg.TenantID = tenantID
	msg.Direction = "outbound"
	msg.Status = "pending"

	if err := h.DB.Create(&msg).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Queue for delivery via messaging provider
	if h.MsgManager != nil {
		go h.MsgManager.SendMessage(tenantID, msg.From, msg.To, msg.Body, nil, msg.ProviderID)
	}

	return c.Status(http.StatusCreated).JSON(msg)
}

// =====================
// Contacts
// =====================

func (h *Handler) ListContacts(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	search := c.Query("search")

	query := h.DB.Where("tenant_id = ?", tenantID)
	if search != "" {
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR phone ILIKE ? OR email ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var contacts []models.Contact
	query.Order("last_name, first_name").Limit(100).Find(&contacts)
	return c.JSON(contacts)
}

func (h *Handler) CreateContact(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var contact models.Contact
	if err := c.BodyParser(&contact); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	contact.TenantID = tenantID

	if err := h.DB.Create(&contact).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(contact)
}

func (h *Handler) GetContact(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		h.logWarn("MESSAGING", "GetContact: Contact not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Contact not found"})
	}

	return c.JSON(contact)
}

func (h *Handler) UpdateContact(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		h.logWarn("MESSAGING", "UpdateContact: Contact not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Contact not found"})
	}

	if err := c.BodyParser(&contact); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	contact.TenantID = tenantID
	h.DB.Save(&contact)
	return c.JSON(contact)
}

func (h *Handler) DeleteContact(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		h.logWarn("MESSAGING", "DeleteContact: Contact not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Contact not found"})
	}

	h.DB.Delete(&contact)
	c.Status(http.StatusNoContent)
	return nil
}

func (h *Handler) SyncContact(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var contact models.Contact
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).First(&contact).Error; err != nil {
		h.logWarn("MESSAGING", "SyncContact: Contact not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Contact not found"})
	}

	// Guard: Contact must have an external source configured
	if contact.ExternalSource == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Contact has no external source configured"})
	}

	// Guard: Contact must have an external ID
	if contact.ExternalID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Contact has no external ID"})
	}

	// Find the webhook configuration for this source
	var webhook models.ContactWebhook
	if err := h.DB.Where("tenant_id = ? AND source = ? AND enabled = true", tenantID, contact.ExternalSource).
		First(&webhook).Error; err != nil {
		h.logWarn("MESSAGING", "SyncContact: Webhook not found", h.reqFields(c, map[string]interface{}{
			"contact_id": id,
			"source":     contact.ExternalSource,
		}))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Webhook source not found"})
	}

	// Guard: Webhook must have a fetch URL
	if webhook.FetchURL == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Webhook has no fetch URL configured"})
	}

	// Build the fetch URL by substituting the external ID
	fetchURL := replaceURLTemplate(webhook.FetchURL, contact.ExternalID)

	// Prepare the HTTP request
	req, err := http.NewRequest(http.MethodGet, fetchURL, nil)
	if err != nil {
		h.logError("MESSAGING", "SyncContact: Failed to create request", h.reqFields(c, map[string]interface{}{
			"contact_id": id,
			"error":      err.Error(),
		}))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create webhook request"})
	}

	// Set default headers
	req.Header.Set("Accept", "application/json")

	// Apply custom headers from webhook configuration
	if webhook.FetchHeaders != "" {
		applyWebhookHeaders(req, webhook.FetchHeaders)
	}

	// Execute the webhook request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.logError("MESSAGING", "SyncContact: Webhook request failed", h.reqFields(c, map[string]interface{}{
			"contact_id": id,
			"url":        fetchURL,
			"error":      err.Error(),
		}))
		return c.Status(http.StatusBadGateway).JSON(fiber.Map{"error": "Webhook request failed"})
	}
	defer resp.Body.Close()

	// Guard: Validate response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		h.logWarn("MESSAGING", "SyncContact: Webhook returned error", h.reqFields(c, map[string]interface{}{
			"contact_id": id,
			"status":     resp.StatusCode,
		}))
		return c.Status(http.StatusBadGateway).JSON(fiber.Map{"error": fmt.Sprintf("Webhook returned status %d", resp.StatusCode)})
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logError("MESSAGING", "SyncContact: Failed to read response", h.reqFields(c, map[string]interface{}{
			"contact_id": id,
			"error":      err.Error(),
		}))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read webhook response"})
	}

	// Parse the response data
	var webhookData map[string]interface{}
	if err := json.Unmarshal(body, &webhookData); err != nil {
		h.logWarn("MESSAGING", "SyncContact: Failed to parse webhook response", h.reqFields(c, map[string]interface{}{
			"contact_id": id,
			"error":      err.Error(),
		}))
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid webhook response format"})
	}

	// Apply field mapping if configured
	updatedContact := applyFieldMapping(contact, webhookData, webhook.FieldMapping)

	// Update sync timestamp
	now := time.Now()
	updatedContact.LastSyncAt = &now

	// Save the updated contact
	if err := h.DB.Save(&updatedContact).Error; err != nil {
		h.logError("MESSAGING", "SyncContact: Failed to save contact", h.reqFields(c, map[string]interface{}{
			"contact_id": id,
			"error":      err.Error(),
		}))
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update contact"})
	}

	// Update webhook's last sync timestamp
	h.DB.Model(&webhook).Updates(map[string]interface{}{
		"last_sync_at":     now,
		"last_sync_status": "success",
	})

	h.logInfo("MESSAGING", "SyncContact: Contact synced successfully", h.reqFields(c, map[string]interface{}{
		"contact_id": id,
		"source":     contact.ExternalSource,
	}))

	return c.JSON(updatedContact)
}

// replaceURLTemplate substitutes {{external_id}} placeholder in the URL
func replaceURLTemplate(templateURL, externalID string) string {
	templateURL = strings.ReplaceAll(templateURL, "{{external_id}}", url.PathEscape(externalID))
	return templateURL
}

// applyWebhookHeaders parses and applies custom headers to the request
func applyWebhookHeaders(req *http.Request, headersJSON string) {
	var headers map[string]string
	if err := json.Unmarshal([]byte(headersJSON), &headers); err != nil {
		return
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

// applyFieldMapping maps webhook data to contact fields based on field mapping config
func applyFieldMapping(contact models.Contact, data map[string]interface{}, fieldMappingJSON string) models.Contact {
	// If no field mapping, try auto-mapping based on common field names
	if fieldMappingJSON == "" {
		contact = autoMapFields(contact, data)
		return contact
	}

	// Parse field mapping configuration
	var fieldMapping map[string]string
	if err := json.Unmarshal([]byte(fieldMappingJSON), &fieldMapping); err != nil {
		// Fall back to auto-mapping on parse error
		return autoMapFields(contact, data)
	}

	// Apply configured mappings
	for contactField, dataKey := range fieldMapping {
		if value, exists := data[dataKey]; exists {
			contact = setContactField(contact, contactField, value)
		}
	}

	return contact
}

// autoMapFields attempts to map common field names automatically
func autoMapFields(contact models.Contact, data map[string]interface{}) models.Contact {
	// Map common field names from various CRM/webhook systems
	fieldMappings := map[string][]string{
		"FirstName":   {"first_name", "firstName", "given_name", "givenName", "name"},
		"LastName":    {"last_name", "lastName", "family_name", "familyName", "surname"},
		"DisplayName": {"display_name", "displayName", "full_name", "fullName", "name"},
		"Company":     {"company", "company_name", "companyName", "organization", "org"},
		"Title":       {"title", "job_title", "jobTitle", "position"},
		"Email":       {"email", "email_address", "emailAddress"},
		"Phone":       {"phone", "phone_number", "phoneNumber", "telephone"},
		"MobilePhone": {"mobile_phone", "mobilePhone", "cell_phone", "cellPhone", "mobile"},
		"Address1":    {"address1", "address_1", "street_address", "streetAddress", "address"},
		"City":        {"city", "town"},
		"State":       {"state", "province", "region"},
		"PostalCode":  {"postal_code", "postalCode", "zip", "zip_code", "zipCode"},
		"Country":     {"country"},
	}

	for contactField, dataKeys := range fieldMappings {
		for _, dataKey := range dataKeys {
			if value, exists := data[dataKey]; exists {
				contact = setContactField(contact, contactField, value)
				break
			}
		}
	}

	return contact
}

// setContactField sets a field on the contact by name
func setContactField(contact models.Contact, fieldName string, value interface{}) models.Contact {
	strValue := fmt.Sprintf("%v", value)

	switch fieldName {
	case "FirstName":
		contact.FirstName = strValue
	case "LastName":
		contact.LastName = strValue
	case "DisplayName":
		contact.DisplayName = strValue
	case "Company":
		contact.Company = strValue
	case "Title":
		contact.Title = strValue
	case "Email":
		contact.Email = strValue
	case "Phone":
		contact.Phone = strValue
	case "PhoneAlt":
		contact.PhoneAlt = strValue
	case "MobilePhone":
		contact.MobilePhone = strValue
	case "Address1":
		contact.Address1 = strValue
	case "Address2":
		contact.Address2 = strValue
	case "City":
		contact.City = strValue
	case "State":
		contact.State = strValue
	case "PostalCode":
		contact.PostalCode = strValue
	case "Country":
		contact.Country = strValue
	case "Notes":
		contact.Notes = strValue
	case "PreferredChannel":
		contact.PreferredChannel = strValue
	case "PreferredLanguage":
		contact.PreferredLanguage = strValue
	case "Timezone":
		contact.Timezone = strValue
	case "Status":
		contact.Status = strValue
	}

	return contact
}

func (h *Handler) GetContactByPhone(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	phone := c.Query("phone")

	var contact models.Contact
	if err := h.DB.Where("tenant_id = ? AND (phone = ? OR mobile_phone = ? OR phone_alt = ?)",
		tenantID, phone, phone, phone).First(&contact).Error; err != nil {
		h.logWarn("MESSAGING", "GetContactByPhone: Contact not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Contact not found"})
	}

	return c.JSON(contact)
}

// =====================
// Chat System
// =====================

func (h *Handler) ListChatThreads(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	channel := c.Query("channel", "")
	status := c.Query("status", "open")

	query := h.DB.Where("tenant_id = ?", tenantID)
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var threads []models.ChatThread
	query.Order("last_message_at DESC").Limit(50).Find(&threads)
	return c.JSON(threads)
}

func (h *Handler) CreateChatThread(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var thread models.ChatThread
	if err := c.BodyParser(&thread); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	thread.TenantID = tenantID

	if err := h.DB.Create(&thread).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(thread)
}

func (h *Handler) GetChatThread(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	id, _ := strconv.ParseUint(c.Params("id"), 10, 64)

	var thread models.ChatThread
	if err := h.DB.Where("id = ? AND tenant_id = ?", id, tenantID).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC").Limit(100)
		}).
		Preload("Messages.Attachments").
		First(&thread).Error; err != nil {
		h.logWarn("MESSAGING", "GetChatThread: Thread not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Thread not found"})
	}

	return c.JSON(thread)
}

func (h *Handler) SendChatMessage(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	threadIDu64, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	threadID := uint(threadIDu64)

	// Verify thread exists
	var thread models.ChatThread
	if err := h.DB.Where("id = ? AND tenant_id = ?", threadID, tenantID).First(&thread).Error; err != nil {
		h.logWarn("MESSAGING", "SendChatMessage: Thread not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Thread not found"})
	}

	var msg models.ChatMessage
	if err := c.BodyParser(&msg); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	msg.TenantID = tenantID
	msg.ThreadID = threadID

	if err := h.DB.Create(&msg).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Update thread's last message time
	h.DB.Model(&thread).Update("last_message_at", msg.CreatedAt)

	return c.Status(http.StatusCreated).JSON(msg)
}

func (h *Handler) ListChatRooms(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var rooms []models.ChatRoom
	h.DB.Where("tenant_id = ? AND archived = false", tenantID).Find(&rooms)

	return c.JSON(rooms)
}

func (h *Handler) CreateChatRoom(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	claims := middleware.GetClaims(c)

	var room models.ChatRoom
	if err := c.BodyParser(&room); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	room.TenantID = tenantID
	if claims != nil {
		room.CreatedByID = claims.UserID
	}

	if err := h.DB.Create(&room).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(room)
}

func (h *Handler) JoinChatRoom(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)
	roomIDu64, _ := strconv.ParseUint(c.Params("id"), 10, 64)
	roomID := uint(roomIDu64)
	claims := middleware.GetClaims(c)

	var room models.ChatRoom
	if err := h.DB.Where("id = ? AND tenant_id = ?", roomID, tenantID).First(&room).Error; err != nil {
		h.logWarn("MESSAGING", "JoinChatRoom: Room not found", h.reqFields(c, nil))
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Room not found"})
	}

	// Check if already member
	var existing models.ChatRoomMember
	if err := h.DB.Where("room_id = ? AND extension_id = ?", roomID, claims.UserID).First(&existing).Error; err == nil {
		return c.JSON(fiber.Map{"message": "Already a member"})
	}

	member := models.ChatRoomMember{
		RoomID:      roomID,
		ExtensionID: claims.UserID,
		Role:        "member",
	}
	h.DB.Create(&member)

	return c.Status(http.StatusCreated).JSON(member)
}

func (h *Handler) ListChatQueues(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var queues []models.ChatQueue
	h.DB.Where("tenant_id = ?", tenantID).Preload("Agents").Find(&queues)

	return c.JSON(queues)
}

func (h *Handler) CreateChatQueue(c *fiber.Ctx) error {
	tenantID := middleware.GetTenantID(c)

	var queue models.ChatQueue
	if err := c.BodyParser(&queue); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	queue.TenantID = tenantID

	if err := h.DB.Create(&queue).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(queue)
}
