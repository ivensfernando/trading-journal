package lookup

import (
	"encoding/json"
	"fmt"
	"net/http"
	"vsC1Y2025V01/src/db"
	"vsC1Y2025V01/src/model"

	"github.com/sirupsen/logrus"
)

// GET /exchanges
func ListExchanges(logger *logrus.Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var exchanges []model.Exchange
		var total int64

		// Count total first
		if err := db.DB.Model(&model.Exchange{}).Count(&total).Error; err != nil {
			logger.WithError(err).Error("Failed to count exchanges")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Now fetch data (optionally add pagination later)
		if err := db.DB.Find(&exchanges).Error; err != nil {
			logger.WithError(err).Error("Failed to fetch exchanges")
			http.Error(w, "Error fetching exchanges", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Expose-Headers", "X-Total-Count")
		w.Header().Set("X-Total-Count", fmt.Sprintf("%d", total))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(exchanges)
	}
}

// GET /pairs
func ListPairs(logger *logrus.Entry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pairs []model.PairsCoins
		var total int64

		// Count total first
		if err := db.DB.Model(&model.PairsCoins{}).Count(&total).Error; err != nil {
			logger.WithError(err).Error("Failed to count exchanges")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Now fetch data (optionally add pagination later)
		if err := db.DB.Find(&pairs).Error; err != nil {
			logger.WithError(err).Error("Failed to fetch exchanges")
			http.Error(w, "Error fetching exchanges", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Expose-Headers", "X-Total-Count")
		w.Header().Set("X-Total-Count", fmt.Sprintf("%d", total))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(pairs)
	}
}
