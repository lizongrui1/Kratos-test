package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"student/internal/data/model"

	"student/internal/data"

	"gorm.io/gorm"
)

type Message struct {
	ID     int32  `json:"id"`
	Name   string `json:"name"`
	Info   string `json:"info"`
	Status string `json:"status"`
}

func getData(db *data.Data) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		var records []Message
		if err := db.DB().Model(&model.Student{}).Where("id = ?", id).First(&records).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(w, "Record not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(records)
	}
}

func serviceB(db *data.Data) {
	http.HandleFunc("/serviceB", getData(db))
	http.ListenAndServe(":8081", nil)
}
