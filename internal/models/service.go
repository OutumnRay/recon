package models

import "time"

// ServiceStatus представляет статус сервиса
type ServiceStatus string

const (
	ServiceStatusRunning ServiceStatus = "running" // Работает
	ServiceStatusStopped ServiceStatus = "stopped" // Остановлен
	ServiceStatusError   ServiceStatus = "error"   // Ошибка
	ServiceStatusUnknown ServiceStatus = "unknown" // Неизвестен
)

// ServiceType представляет тип сервиса
type ServiceType string

const (
	ServiceTypeManagingPortal      ServiceType = "managing-portal"      // Портал управления
	ServiceTypeUserPortal          ServiceType = "user-portal"          // Пользовательский портал
	ServiceTypeTranscriptionWorker ServiceType = "transcription-worker" // Воркер транскрипции
	ServiceTypeSummarizationWorker ServiceType = "summarization-worker" // Воркер суммаризации
)

// ServiceInfo представляет информацию о сервисе
type ServiceInfo struct {
	// ID - уникальный идентификатор сервиса
	ID          string        `json:"id"`
	// Name - название сервиса
	Name        string        `json:"name"`
	// Type - тип сервиса
	Type        ServiceType   `json:"type"`
	// Status - текущий статус сервиса
	Status      ServiceStatus `json:"status"`
	// Version - версия сервиса
	Version     string        `json:"version"`
	// StartedAt - время запуска сервиса
	StartedAt   time.Time     `json:"started_at"`
	// LastHeartbeat - время последнего сигнала жизни
	LastHeartbeat time.Time   `json:"last_heartbeat"`
	// Metadata - дополнительные метаданные сервиса
	Metadata    map[string]interface{} `json:"metadata"`
}

// SystemStatus представляет общий статус системы
type SystemStatus struct {
	// Status - общий статус системы
	Status       string                  `json:"status"`
	// Services - информация о сервисах
	Services     map[string]ServiceInfo  `json:"services"`
	// Infrastructure - информация об инфраструктуре
	Infrastructure map[string]interface{} `json:"infrastructure"`
	// Timestamp - время получения статуса
	Timestamp    time.Time               `json:"timestamp"`
}

// HealthResponse представляет ответ на проверку здоровья
type HealthResponse struct {
	// Status - статус здоровья
	Status    string    `json:"status"`
	// Timestamp - время проверки
	Timestamp time.Time `json:"timestamp"`
	// Version - версия сервиса
	Version   string    `json:"version"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	// Error - код ошибки
	Error   string    `json:"error"`
	// Message - сообщение об ошибке
	Message string    `json:"message"`
	// Code - HTTP код ошибки
	Code    int       `json:"code"`
	// Timestamp - время возникновения ошибки
	Timestamp time.Time `json:"timestamp"`
}
