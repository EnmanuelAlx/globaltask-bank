package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/globaltask/bank/internal/application/service"
	"github.com/globaltask/bank/internal/domain/entity"
	"github.com/globaltask/bank/internal/infrastructure/middleware"
	"github.com/globaltask/bank/internal/infrastructure/repository"
)

type LoanApplicationHandler struct {
	loanService *service.LoanApplicationService
}

func NewLoanApplicationHandler(svc *service.LoanApplicationService) *LoanApplicationHandler {
	return &LoanApplicationHandler{loanService: svc}
}

type CreateApplicationRequest struct {
	CountryID        int     `json:"country_id" binding:"required"`
	BorrowerName     string  `json:"borrower_name"`
	IdentityDocument string  `json:"identity_document" binding:"required"`
	RequestedAmount  float64 `json:"requested_amount" binding:"required,gt=0"`
	MonthlyIncome    float64 `json:"monthly_income" binding:"required,gt=0"`
}

// CreateApplication handles POST /applications
func (h *LoanApplicationHandler) CreateApplication(c *gin.Context) {
	var req CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	app, err := h.loanService.CreateApplication(c.Request.Context(), &service.CreateApplicationInput{
		UserID:           userID,
		CountryID:        req.CountryID,
		BorrowerName:     req.BorrowerName,
		IdentityDocument: req.IdentityDocument,
		RequestedAmount:  req.RequestedAmount,
		MonthlyIncome:    req.MonthlyIncome,
		UserRole:         role,
	})
	if err != nil {
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, app)
}

// ListApplications handles GET /applications
func (h *LoanApplicationHandler) ListApplications(c *gin.Context) {
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)

	filter := repository.ListFilter{
		Limit: 20,
	}

	// Parse query params
	if countryID := c.Query("country_id"); countryID != "" {
		if id, err := strconv.Atoi(countryID); err == nil {
			filter.CountryID = &id
		}
	}
	if status := c.Query("status"); status != "" {
		s := entity.LoanApplicationStatus(status)
		filter.Status = &s
	}
	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}
	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o > 0 {
			filter.Offset = o
		}
	}

	if borrowerName := c.Query("borrower_name"); borrowerName != "" {
		filter.BorrowerName = &borrowerName
	}
	if idDoc := c.Query("identity_document"); idDoc != "" {
		filter.IdentityDocument = &idDoc
	}
	if minAmount := c.Query("min_amount"); minAmount != "" {
		if val, err := strconv.ParseFloat(minAmount, 64); err == nil {
			filter.MinAmount = &val
		}
	}
	if maxAmount := c.Query("max_amount"); maxAmount != "" {
		if val, err := strconv.ParseFloat(maxAmount, 64); err == nil {
			filter.MaxAmount = &val
		}
	}
	if minIncome := c.Query("min_income"); minIncome != "" {
		if val, err := strconv.ParseFloat(minIncome, 64); err == nil {
			filter.MinIncome = &val
		}
	}
	if maxIncome := c.Query("max_income"); maxIncome != "" {
		if val, err := strconv.ParseFloat(maxIncome, 64); err == nil {
			filter.MaxIncome = &val
		}
	}

	// Regular users can only see their own applications
	if role != "ADMIN" {
		filter.UserID = &userID
	}

	apps, err := h.loanService.ListApplications(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   apps,
		"total":  len(apps),
		"filter": filter,
	})
}

// GetApplication handles GET /applications/:id
func (h *LoanApplicationHandler) GetApplication(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	app, err := h.loanService.GetApplicationByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if app == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
		return
	}

	// Check ownership for non-admin users
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	if role != "ADMIN" && app.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, app)
}

type UpdateApplicationRequest struct {
	Status          *entity.LoanApplicationStatus `json:"status"`
	BankInformation *map[string]interface{}       `json:"bank_information"`
}

// UpdateApplication handles PATCH /applications/:id
func (h *LoanApplicationHandler) UpdateApplication(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var req UpdateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.loanService.UpdateApplication(c.Request.Context(), id, &service.UpdateApplicationInput{
		Status:          req.Status,
		BankInformation: req.BankInformation,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated successfully"})
}

// HandleBankWebhook handles POST /webhook/bank-update
func (h *LoanApplicationHandler) HandleBankWebhook(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	applicationID, ok := payload["application_id"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing application_id"})
		return
	}

	err := h.loanService.HandleBankWebhook(c.Request.Context(), applicationID, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook received"})
}
