package controllers

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/Sogilis/Voogle/src/cmd/api/db/dao"
	"github.com/Sogilis/Voogle/src/pkg/clients"
)

type VideoCoverHandler struct {
	S3Client  clients.IS3Client
	VideosDAO *dao.VideosDAO
	UUIDGen   clients.IUUIDGenerator
}

// VideoCoverHandler godoc
// @Summary Get video cover image in base64
// @Description Get video cover image in base64
// @Tags video
// @Accept plain
// @Produce plain
// @Param id path string true "Video ID"
// @Success 200 {string} string "video cover image in base64"
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /api/v1/videos/{id}/cover [get]
func (v VideoCoverHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Debug("GET VideoCoverHandler - parameters ", vars)

	id := vars["id"]
	if !v.UUIDGen.IsValidUUID(id) {
		log.Error("Invalid id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Fetch video cover path from DB
	video, err := v.VideosDAO.GetVideo(r.Context(), id)
	if err != nil {
		log.Error("Failed to get video "+id+" info from DB: ", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// If there is no cover file, return nothing, but no error
	if video.CoverPath != "" {
		if r.Method == "head" {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Fetch cover image from S3
		object, err := v.S3Client.GetObject(r.Context(), video.CoverPath)
		if err != nil {
			log.Error("Failed to open video cover "+video.CoverPath+": ", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/octet-stream")

		if _, err := io.Copy(w, object); err != nil {
			log.Error("Unable to stream cover", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}
