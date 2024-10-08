package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"student/internal/data"
	"student/internal/data/model"

	"gorm.io/gorm"
)

type Score struct {
	Score int `json:"score"`
}

func getScore(db *data.Data) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var score int32
		if err := db.DB().Model(&model.Student{}).Select("score").Where("id = ?", id).First(&score).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "Record not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
			}
			return
		}
		response := Score{
			Score: int(score),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func ServiceC(db *data.Data) {
	http.HandleFunc("/serviceC", getScore(db))
	http.ListenAndServe(":8082", nil)
}
