package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mdflamingo/fgo-tracker-backend/internal/auth"
	"github.com/mdflamingo/fgo-tracker-backend/internal/logger"
	"github.com/mdflamingo/fgo-tracker-backend/internal/model"
	"github.com/mdflamingo/fgo-tracker-backend/internal/validator"
	"go.uber.org/zap"
)

func parseTaskID(r *http.Request) (uuid.UUID, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return uuid.Nil, errors.New("task ID is required")
	}
	return uuid.Parse(idStr)
}

func parseFilters(r *http.Request) (model.TaskFilter, error) {
	q := r.URL.Query()

	filter := model.TaskFilter{
		Limit:  50,
		Offset: 0,
	}

	filter.Name = q.Get("name")

	if status := q.Get("status"); status != "" {
		filter.Status = model.TaskStatus(status)
	}
	if priority := q.Get("priority"); priority != "" {
		filter.Priority = model.TaskPriority(priority)
	}

	if pid := q.Get("project_id"); pid != "" {
		parsedID, err := uuid.Parse(pid)
		if err != nil {
			return model.TaskFilter{}, fmt.Errorf("invalid project_id format: %w", err)
		}
		filter.ProjectId = parsedID
	}

	if cid := q.Get("creator_id"); cid != "" {
		parsedID, err := uuid.Parse(cid)
		if err != nil {
			return model.TaskFilter{}, fmt.Errorf("invalid creator_id format: %w", err)
		}
		filter.CreatorId = parsedID
	}

	if assignedStr := q.Get("assigned_ids"); assignedStr != "" {
		ids, err := parseUUIDArray(assignedStr)
		if err != nil {
			return model.TaskFilter{}, fmt.Errorf("invalid assigned_ids: %w", err)
		}
		filter.AssignedIds = ids
	}

	if reviewerStr := q.Get("reviewer_ids"); reviewerStr != "" {
		ids, err := parseUUIDArray(reviewerStr)
		if err != nil {
			return model.TaskFilter{}, fmt.Errorf("invalid reviewer_ids: %w", err)
		}
		filter.ReviewerIds = ids
	}

	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			return model.TaskFilter{}, errors.New("invalid limit: must be a positive integer")
		}
		if limit > 100 { // подставлять значения из настроек
			limit = 100
		}
		filter.Limit = limit
	}

	if offsetStr := q.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return model.TaskFilter{}, errors.New("invalid offset: must be a non-negative integer")
		}
		filter.Offset = offset
	}

	if err := validator.GlobalValidator.Struct(filter); err != nil {
		return model.TaskFilter{}, fmt.Errorf("validation failed: %w", err)
	}

	return filter, nil
}

func parseUUIDArray(s string) ([]uuid.UUID, error) {
	parts := strings.Split(s, ",")
	ids := make([]uuid.UUID, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		id, err := uuid.Parse(p)
		if err != nil {
			return nil, fmt.Errorf("invalid uuid '%s'", p)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func parseUUID(w http.ResponseWriter, r *http.Request) (uuid.UUID, error) {
	creatorUUIDStr, err := auth.GetUserIDFromRequest(r)
	if err != nil {
		logger.Log.Error("handler: not parsed UUID from context", zap.Error(err))
		return uuid.Nil, fmt.Errorf("invalid uuid '%s'", err)
	}
	creatorUUID := uuid.MustParse(creatorUUIDStr)

	return creatorUUID, nil
}
