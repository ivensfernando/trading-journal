package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"vsC1Y2025V01/src/payloads"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"vsC1Y2025V01/src/auth"
	"vsC1Y2025V01/src/db"
	"vsC1Y2025V01/src/model"
)

func CreateWebhookHandler(logger *logrus.Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok || user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var payload payloads.CreateWebhookPayload
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil {
			logger.WithError(err).Warn("invalid webhook payload")
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		payload.Name = strings.TrimSpace(payload.Name)
		payload.Type = strings.TrimSpace(payload.Type)
		if payload.Name == "" || payload.Type == "" {
			http.Error(w, "Name and type are required", http.StatusBadRequest)
			return
		}

		token, err := generateToken()
		if err != nil {
			logger.WithError(err).Error("failed to generate webhook token")
			http.Error(w, "Unable to create webhook", http.StatusInternalServerError)
			return
		}

		webhook := model.Webhook{
			UserID:      user.ID,
			Name:        payload.Name,
			Description: payload.Description,
			Style:       "",
			Type:        payload.Type,
			Tickers:     payload.Tickers,
			Token:       token,
			CreatedAt:   time.Time{},
			UpdatedAt:   time.Time{},
		}

		if err := db.DB.Create(&webhook).Error; err != nil {
			logger.WithError(err).Error("failed to persist webhook")
			http.Error(w, "Unable to create webhook", http.StatusInternalServerError)
			return
		}

		scheme := "https"
		if r.TLS == nil {
			scheme = "http"
		}

		response := map[string]interface{}{
			"webhook": webhook,
			"url":     fmt.Sprintf("%s://%s/trading/webhook/%s", scheme, r.Host),
			"token":   webhook.Token,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.WithError(err).Error("failed to encode webhook response")
		}
	}
}

func AlertHandler(logger *logrus.Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		if strings.TrimSpace(token) == "" {
			http.Error(w, "Webhook token is required", http.StatusBadRequest)
			return
		}

		var webhook model.Webhook
		if err := db.DB.Where("token = ?", token).First(&webhook).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "Webhook not found", http.StatusNotFound)
				return
			}
			logger.WithError(err).Error("failed to load webhook")
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		var payload payloads.AlertPayload
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil {
			logger.WithError(err).Warn("invalid alert payload")
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		var alertTime *time.Time
		if payload.Time != "" {
			parsedTime, err := time.Parse(time.RFC3339, payload.Time)
			if err != nil {
				logger.WithError(err).Warn("invalid alert Time")
				http.Error(w, "time must be RFC3339 formatted", http.StatusBadRequest)
				return
			}
			alertTime = &parsedTime
		}

		quantity, err := stringToFloat64(payload.Quantity)
		if err != nil {
			logger.WithError(err).Warn("invalid alert quantity")
			http.Error(w, "quantity must be Float", http.StatusBadRequest)
			return
		}

		price, err := stringToFloat64(payload.Price)
		if err != nil {
			logger.WithError(err).Warn("invalid alert price")
			http.Error(w, "price must be Float", http.StatusBadRequest)
			return
		}

		alert := model.WebhookAlert{
			WebhookID:  webhook.ID,
			UserID:     webhook.UserID,
			Ticker:     payload.Ticker,
			Action:     payload.Action,
			Sentiment:  payload.Sentiment,
			Quantity:   quantity,
			Price:      price,
			Interval:   payload.Interval,
			AlertTime:  alertTime,
			ReceivedAt: time.Now(),
		}

		if err := db.DB.Create(&alert).Error; err != nil {
			logger.WithError(err).Error("failed to store webhook alert")
			http.Error(w, "Unable to store alert", http.StatusInternalServerError)
			return
		}

		//w.Header().Set("Content-Type", "application/json")
		//if err := json.NewEncoder(w).Encode(alert); err != nil {
		//	logger.WithError(err).Error("failed to encode alert response")
		//}
		w.WriteHeader(http.StatusOK)
		return
	}
}
