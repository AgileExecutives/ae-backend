package main

// This file provides Swagger documentation stubs for booking module endpoints.
// The actual implementations live in the booking module (github.com/unburdy/booking-module).
// These stubs exist only so that `swag init` run in this repository can include the paths.

import (
	baseAPI "github.com/ae-base-server/api"
	_ "github.com/ae-base-server/pkg/bootstrap" // dummy underscore import to satisfy build (optional)
	bookingEntities "github.com/unburdy/booking-module/entities"
)

// CreateBookingTemplate swagger docs stub
// @Summary Create a new booking configuration
// @Description Create a new booking configuration/template for a user's calendar
// @Tags booking-templates
// @Accept json
// @Produce json
// @Param configuration body bookingEntities.CreateBookingTemplateRequest true "Booking configuration data (includes allowed_start_minutes)"
// @Success 201 {object} baseAPI.APIResponse{data=bookingEntities.BookingTemplateResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates [post]
func CreateBookingTemplateSwaggerStub() {
	// reference types so imports are used
	_ = bookingEntities.CreateBookingTemplateRequest{}
	_ = baseAPI.APIResponse{}
}

// GetBookingTemplate swagger docs stub
// @Summary Get a booking configuration by ID
// @Description Retrieve a specific booking configuration by ID
// @Tags booking-templates
// @Produce json
// @Param id path int true "Configuration ID"
// @Success 200 {object} baseAPI.APIResponse{data=bookingEntities.BookingTemplateResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates/{id} [get]
func GetBookingTemplateSwaggerStub() {
	_ = bookingEntities.BookingTemplateResponse{}
}

// ListBookingTemplates swagger docs stub
// @Summary Get all booking configurations
// @Description Retrieve all booking configurations for the tenant
// @Tags booking-templates
// @Produce json
// @Success 200 {object} baseAPI.APIResponse{data=[]bookingEntities.BookingTemplateResponse}
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates [get]
func ListBookingTemplatesSwaggerStub() {
	_ = []bookingEntities.BookingTemplateResponse{}
}

// UpdateBookingTemplate swagger docs stub
// @Summary Update a booking configuration
// @Description Update an existing booking configuration
// @Tags booking-templates
// @Accept json
// @Produce json
// @Param id path int true "Configuration ID"
// @Param configuration body bookingEntities.UpdateBookingTemplateRequest true "Updated configuration data (allowed_start_minutes optional)"
// @Success 200 {object} baseAPI.APIResponse{data=bookingEntities.BookingTemplateResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates/{id} [put]
func UpdateBookingTemplateSwaggerStub() {
	_ = bookingEntities.UpdateBookingTemplateRequest{}
}

// DeleteBookingTemplate swagger docs stub
// @Summary Delete a booking configuration
// @Description Soft delete a booking configuration by ID
// @Tags booking-templates
// @Produce json
// @Param id path int true "Configuration ID"
// @Success 200 {object} baseAPI.APIResponse
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/templates/{id} [delete]
func DeleteBookingTemplateSwaggerStub() {}

// CreateBookingLink swagger docs stub
// @Summary Create a booking link
// @Description Generate a booking link token for a client to book appointments
// @Tags booking-templates
// @Accept json
// @Produce json
// @Param link body bookingEntities.CreateBookingLinkRequest true "Booking link data"
// @Success 201 {object} baseAPI.APIResponse{data=bookingEntities.BookingLinkResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Security BearerAuth
// @Router /booking/link [post]
func CreateBookingLinkSwaggerStub() {
	_ = bookingEntities.CreateBookingLinkRequest{}
	_ = bookingEntities.BookingLinkResponse{}
}

// GetFreeSlots swagger docs stub (public via booking token)
// @Summary Get available time slots for booking
// @Description Retrieve available time slots based on a booking link token. Token is validated and must not be blacklisted.
// @Tags booking-slots
// @Produce json
// @Param token path string true "Booking link token"
// @Param start query string false "Start date for slot search (YYYY-MM-DD)" example("2025-11-01")
// @Param end query string false "End date for slot search (YYYY-MM-DD)" example("2025-11-30")
// @Success 200 {object} baseAPI.APIResponse{data=bookingEntities.FreeSlotsResponse}
// @Failure 400 {object} baseAPI.APIResponse
// @Failure 401 {object} baseAPI.APIResponse
// @Failure 404 {object} baseAPI.APIResponse
// @Failure 500 {object} baseAPI.APIResponse
// @Router /booking/freeslots/{token} [get]
func GetFreeSlotsSwaggerStub() {
	_ = bookingEntities.FreeSlotsResponse{}
}
