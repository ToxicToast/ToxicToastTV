package mapper

import (
	"toxictoast/services/webhook-service/internal/domain"
	"toxictoast/services/webhook-service/internal/repository/entity"
)

func DeliveryAttemptToEntity(attempt *domain.DeliveryAttempt) *entity.DeliveryAttemptEntity {
	if attempt == nil {
		return nil
	}

	return &entity.DeliveryAttemptEntity{
		ID:              attempt.ID,
		DeliveryID:      attempt.DeliveryID,
		AttemptNumber:   attempt.AttemptNumber,
		RequestURL:      attempt.RequestURL,
		RequestHeaders:  attempt.RequestHeaders,
		RequestBody:     attempt.RequestBody,
		ResponseStatus:  attempt.ResponseStatus,
		ResponseHeaders: attempt.ResponseHeaders,
		ResponseBody:    attempt.ResponseBody,
		Success:         attempt.Success,
		Error:           attempt.Error,
		DurationMs:      attempt.DurationMs,
		CreatedAt:       attempt.CreatedAt,
	}
}

func DeliveryAttemptToDomain(e *entity.DeliveryAttemptEntity) *domain.DeliveryAttempt {
	if e == nil {
		return nil
	}

	attempt := &domain.DeliveryAttempt{
		ID:              e.ID,
		DeliveryID:      e.DeliveryID,
		AttemptNumber:   e.AttemptNumber,
		RequestURL:      e.RequestURL,
		RequestHeaders:  e.RequestHeaders,
		RequestBody:     e.RequestBody,
		ResponseStatus:  e.ResponseStatus,
		ResponseHeaders: e.ResponseHeaders,
		ResponseBody:    e.ResponseBody,
		Success:         e.Success,
		Error:           e.Error,
		DurationMs:      e.DurationMs,
		CreatedAt:       e.CreatedAt,
	}

	if e.DeletedAt.Valid {
		deletedAt := e.DeletedAt.Time
		attempt.DeletedAt = &deletedAt
	}

	return attempt
}

func DeliveryAttemptsToDomain(entities []entity.DeliveryAttemptEntity) []domain.DeliveryAttempt {
	attempts := make([]domain.DeliveryAttempt, len(entities))
	for i, e := range entities {
		attempt := DeliveryAttemptToDomain(&e)
		if attempt != nil {
			attempts[i] = *attempt
		}
	}
	return attempts
}
