package domain

import (
	"time"

)

// DeliveryStatus represents the status of a webhook delivery
type DeliveryStatus string

const (
	DeliveryStatusPending  DeliveryStatus = "pending"
	DeliveryStatusSuccess  DeliveryStatus = "success"
	DeliveryStatusFailed   DeliveryStatus = "failed"
	DeliveryStatusRetrying DeliveryStatus = "retrying"
)

// Delivery represents a webhook delivery (one event sent to one webhook)
type Delivery struct {
	ID             string         
	WebhookID      string         
	EventID        string         
	EventType      string         
	EventPayload   string          // JSON string
	Status         DeliveryStatus 
	AttemptCount   int            
	NextRetryAt    *time.Time     `json:"next_retry_at,omitempty"`
	LastAttemptAt  *time.Time     `json:"last_attempt_at,omitempty"`
	CompletedAt    *time.Time     `json:"completed_at,omitempty"`
	LastError      string         
	CreatedAt      time.Time      
	UpdatedAt      time.Time      
	DeletedAt      *time.Time 

	// Relations
	Webhook  *Webhook           
	Attempts []DeliveryAttempt  
}


// DeliveryAttempt represents a single delivery attempt
type DeliveryAttempt struct {
	ID               string         
	DeliveryID       string         
	AttemptNumber    int            
	RequestURL       string         
	RequestHeaders   string          // JSON string
	RequestBody      string         
	ResponseStatus   int            `json:"response_status"`
	ResponseHeaders  string          // JSON string
	ResponseBody     string         
	Success          bool           `json:"success"`
	Error            string         
	DurationMs       int            `json:"duration_ms"`
	CreatedAt        time.Time      
	DeletedAt        *time.Time 

	// Relations
	Delivery *Delivery 
}

