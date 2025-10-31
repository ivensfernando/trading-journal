package userexchanges

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"vsC1Y2025V01/internal/connectors"
	"vsC1Y2025V01/src/auth"
	"vsC1Y2025V01/src/model"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type mexcConnector interface {
	TestConnection() error
}

var mexcConnectorFactory = func(apiKey, apiSecret string) mexcConnector {
	return connectors.NewMexcConnector(apiKey, apiSecret)
}

func UpsertUserExchangeHandler(logger *logrus.Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok || user == nil {
			logger.Warn("user not found in context while upserting user exchange")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var payload model.UpsertUserExchangePayload
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil {
			logger.WithError(err).Warn("invalid user exchange payload")
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		if payload.ExchangeID == 0 {
			http.Error(w, "exchangeId is required", http.StatusBadRequest)
			return
		}

		payload.APIKey = strings.TrimSpace(payload.APIKey)
		payload.APISecret = strings.TrimSpace(payload.APISecret)
		payload.APIPassphrase = strings.TrimSpace(payload.APIPassphrase)

		exchange, err := getUserExchangeStore().GetExchangeByID(payload.ExchangeID)
		if err != nil {
			if errors.Is(err, ErrExchangeNotFound) {
				http.Error(w, "exchange not found", http.StatusBadRequest)
				return
			}

			logger.WithError(err).Error("failed to load exchange for user exchange upsert")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		userExchange, err := getUserExchangeStore().FindUserExchange(user.ID, payload.ExchangeID)
		if err != nil {
			if errors.Is(err, ErrUserExchangeNotFound) {
				userExchange = &model.UserExchange{
					UserID:     user.ID,
					ExchangeID: payload.ExchangeID,
				}
			} else {
				logger.WithError(err).Error("failed to fetch user exchange")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		if payload.APIKey != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(payload.APIKey), bcrypt.DefaultCost)
			if err != nil {
				logger.WithError(err).Error("failed to hash api key")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			userExchange.APIKeyHash = string(hash)
		}

		if payload.APISecret != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(payload.APISecret), bcrypt.DefaultCost)
			if err != nil {
				logger.WithError(err).Error("failed to hash api secret")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			userExchange.APISecretHash = string(hash)
		}

		if payload.APIPassphrase != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(payload.APIPassphrase), bcrypt.DefaultCost)
			if err != nil {
				logger.WithError(err).Error("failed to hash api passphrase")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			userExchange.APIPassphraseHash = string(hash)
		}

		userExchange.ShowInForms = payload.ShowInForms

		if err := getUserExchangeStore().SaveUserExchange(userExchange); err != nil {
			logger.WithError(err).Error("failed to upsert user exchange")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		userExchange.Exchange = exchange

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(model.NewUserExchangeResponse(userExchange)); err != nil {
			logger.WithError(err).Error("failed to encode user exchange response")
		}
	}
}

func ListFormUserExchangesHandler(logger *logrus.Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok || user == nil {
			logger.Warn("user not found in context while listing form exchanges")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userExchanges, err := getUserExchangeStore().ListFormUserExchanges(user.ID)
		if err != nil {
			logger.WithError(err).Error("failed to list user exchanges for forms")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		responses := make([]model.UserExchangeResponse, 0, len(userExchanges))
		for i := range userExchanges {
			responses = append(responses, model.NewUserExchangeResponse(&userExchanges[i]))
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(responses); err != nil {
			logger.WithError(err).Error("failed to encode user exchange list response")
		}
	}
}

func DeleteUserExchangeHandler(logger *logrus.Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok || user == nil {
			logger.Warn("user not found in context while deleting user exchange")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		exchangeIDStr := chi.URLParam(r, "exchangeID")
		if exchangeIDStr == "" {
			http.Error(w, "exchangeID is required", http.StatusBadRequest)
			return
		}

		exchangeID, err := strconv.ParseUint(exchangeIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid exchangeID", http.StatusBadRequest)
			return
		}

		deleted, err := getUserExchangeStore().DeleteUserExchange(user.ID, uint(exchangeID))
		if err != nil {
			logger.WithError(err).Error("failed to delete user exchange")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !deleted {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func TestMexcConnectionHandler(logger *logrus.Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok || user == nil {
			logger.Warn("user not found in context while testing MEXC connection")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		exchangeIDStr := chi.URLParam(r, "exchangeID")
		if exchangeIDStr == "" {
			http.Error(w, "exchangeID is required", http.StatusBadRequest)
			return
		}

		exchangeID, err := strconv.ParseUint(exchangeIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid exchangeID", http.StatusBadRequest)
			return
		}

		userExchange, err := getUserExchangeStore().FindUserExchange(user.ID, uint(exchangeID))
		if err != nil {
			if errors.Is(err, ErrUserExchangeNotFound) {
				http.Error(w, "user exchange not found", http.StatusNotFound)
				return
			}

			logger.WithError(err).Error("failed to fetch user exchange for MEXC connection test")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		exchange := userExchange.Exchange
		if exchange == nil {
			exchange, err = getUserExchangeStore().GetExchangeByID(uint(exchangeID))
			if err != nil {
				if errors.Is(err, ErrExchangeNotFound) {
					http.Error(w, "exchange not found", http.StatusNotFound)
					return
				}

				logger.WithError(err).Error("failed to load exchange during MEXC connection test")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		if exchange == nil || !strings.EqualFold(exchange.Name, "mexc") {
			http.Error(w, "exchange is not MEXC", http.StatusBadRequest)
			return
		}

		if userExchange.APIKeyHash == "" || userExchange.APISecretHash == "" {
			http.Error(w, "missing API credentials for MEXC", http.StatusBadRequest)
			return
		}

		connector := mexcConnectorFactory(userExchange.APIKeyHash, userExchange.APISecretHash)
		if connector == nil {
			logger.Error("failed to build MEXC connector")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := connector.TestConnection(); err != nil {
			logger.WithError(err).Warn("MEXC connection test failed")
			http.Error(w, "Failed to connect to MEXC", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			logger.WithError(err).Error("failed to encode MEXC test response")
		}
	}
}
