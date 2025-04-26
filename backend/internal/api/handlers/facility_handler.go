package handlers

import (
	"net/http"
	"strconv"

	"github.com/fasim/backend/internal/models"
	"github.com/fasim/backend/internal/repositories"
	"github.com/labstack/echo/v4"
)

type FacilityHandler struct {
	facilityRepo repositories.FacilityRepository
	itemRepo     repositories.ItemRepository
}

func NewFacilityHandler(facilityRepo repositories.FacilityRepository, itemRepo repositories.ItemRepository) *FacilityHandler {
	return &FacilityHandler{
		facilityRepo: facilityRepo,
		itemRepo:     itemRepo,
	}
}

type inputRequirementRequest struct {
	ItemID   int `json:"itemId"`
	Quantity int `json:"quantity"`
}

type outputDefinitionRequest struct {
	ItemID   int `json:"itemId"`
	Quantity int `json:"quantity"`
}

type createFacilityRequest struct {
	Name           string                   `json:"name"`
	Description    string                   `json:"description"`
	ProcessingTime int64                    `json:"processingTime"`
	Inputs         []inputRequirementRequest  `json:"inputs"`
	Outputs        []outputDefinitionRequest  `json:"outputs"`
}

type updateFacilityRequest struct {
	Name           string                   `json:"name"`
	Description    string                   `json:"description"`
	ProcessingTime int64                    `json:"processingTime"`
	Inputs         []inputRequirementRequest  `json:"inputs"`
	Outputs        []outputDefinitionRequest  `json:"outputs"`
}

type inputRequirementResponse struct {
	ItemID   int `json:"itemId"`
	Quantity int `json:"quantity"`
}

type outputDefinitionResponse struct {
	ItemID   int `json:"itemId"`
	Quantity int `json:"quantity"`
}

type facilityResponse struct {
	ID             int                       `json:"id"`
	Name           string                    `json:"name"`
	Description    string                    `json:"description"`
	ProcessingTime int64                     `json:"processingTime"`
	Inputs         []inputRequirementResponse  `json:"inputs"`
	Outputs        []outputDefinitionResponse  `json:"outputs"`
}

func toInputRequirementResponse(req *models.InputRequirement) inputRequirementResponse {
	return inputRequirementResponse{
		ItemID:   req.Item().ID(),
		Quantity: req.Quantity(),
	}
}

func toOutputDefinitionResponse(def *models.OutputDefinition) outputDefinitionResponse {
	return outputDefinitionResponse{
		ItemID:   def.Item().ID(),
		Quantity: def.Quantity(),
	}
}

func toFacilityResponse(facility *models.Facility) facilityResponse {
	inputs := make([]inputRequirementResponse, len(facility.InputRequirements()))
	for i, req := range facility.InputRequirements() {
		inputs[i] = toInputRequirementResponse(req)
	}

	outputs := make([]outputDefinitionResponse, len(facility.OutputDefinitions()))
	for i, def := range facility.OutputDefinitions() {
		outputs[i] = toOutputDefinitionResponse(def)
	}

	return facilityResponse{
		ID:             facility.ID(),
		Name:           facility.Name(),
		Description:    facility.Description(),
		ProcessingTime: facility.ProcessingTime(),
		Inputs:         inputs,
		Outputs:        outputs,
	}
}

// List handles GET /api/facilities
func (h *FacilityHandler) List(c echo.Context) error {
	facilities, err := h.facilityRepo.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	responses := make([]facilityResponse, len(facilities))
	for i, facility := range facilities {
		responses[i] = toFacilityResponse(facility)
	}

	return c.JSON(http.StatusOK, responses)
}

// Get handles GET /api/facilities/:id
func (h *FacilityHandler) Get(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid facility ID")
	}

	facility, err := h.facilityRepo.Get(c.Request().Context(), id)
	if err != nil || facility == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Facility not found")
	}

	return c.JSON(http.StatusOK, toFacilityResponse(facility))
}

// Create handles POST /api/facilities
func (h *FacilityHandler) Create(c echo.Context) error {
	var req createFacilityRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	facility := models.NewFacility(req.Name, req.Description, req.ProcessingTime)

	// Add input requirements
	for _, input := range req.Inputs {
		item, err := h.itemRepo.Get(c.Request().Context(), input.ItemID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid input item ID")
		}
		facility.AddInputRequirement(models.NewInputRequirement(item, input.Quantity))
	}

	// Add output definitions
	for _, output := range req.Outputs {
		item, err := h.itemRepo.Get(c.Request().Context(), output.ItemID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid output item ID")
		}
		facility.AddOutputDefinition(models.NewOutputDefinition(item, output.Quantity))
	}

	if err := h.facilityRepo.Create(c.Request().Context(), facility); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Get the created facility to ensure we have the correct ID and relationships
	createdFacility, err := h.facilityRepo.Get(c.Request().Context(), facility.ID())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, toFacilityResponse(createdFacility))
}

// Update handles PUT /api/facilities/:id
func (h *FacilityHandler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid facility ID")
	}

	var req updateFacilityRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	existingFacility, err := h.facilityRepo.Get(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if existingFacility == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Facility not found")
	}

	// Create input requirements
	inputReqs := make([]*models.InputRequirement, len(req.Inputs))
	for i, input := range req.Inputs {
		item, err := h.itemRepo.Get(c.Request().Context(), input.ItemID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid input item ID")
		}
		inputReqs[i] = models.NewInputRequirement(item, input.Quantity)
	}

	// Create output definitions
	outputDefs := make([]*models.OutputDefinition, len(req.Outputs))
	for i, output := range req.Outputs {
		item, err := h.itemRepo.Get(c.Request().Context(), output.ItemID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid output item ID")
		}
		outputDefs[i] = models.NewOutputDefinition(item, output.Quantity)
	}

	updatedFacility := models.NewFacilityFromParams(
		id,
		req.Name,
		req.Description,
		inputReqs,
		outputDefs,
		req.ProcessingTime,
	)

	if err := h.facilityRepo.Update(c.Request().Context(), updatedFacility); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, toFacilityResponse(updatedFacility))
}

// Delete handles DELETE /api/facilities/:id
func (h *FacilityHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid facility ID")
	}

	if err := h.facilityRepo.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
