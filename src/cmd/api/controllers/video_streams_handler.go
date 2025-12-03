package controllers

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Sogilis/Voogle/src/cmd/api/db/dao"
	"github.com/Sogilis/Voogle/src/cmd/api/metrics"
	"github.com/Sogilis/Voogle/src/pkg/clients"
	"github.com/Sogilis/Voogle/src/pkg/transformer/v1"
)

type VideoGetMasterHandler struct {
	S3Client clients.IS3Client
	UUIDGen  clients.IUUIDGenerator
}

// VideoGetMasterHandler godoc
// @Summary Get video master
// @Description Get video master
// @Tags video
// @Produce plain
// @Param id path string true "Video ID"
// @Success 200 {string} string "HLS video master"
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /api/v1/videos/{id}/streams/master.m3u8 [get]
func (v VideoGetMasterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Debug("GET VideoGetMasterHandler - parameters ", vars)

	id := vars["id"]
	if !v.UUIDGen.IsValidUUID(id) {
		log.Error("Invalid id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	object, err := v.S3Client.GetObject(r.Context(), id+"/master.m3u8")
	if err != nil {
		log.Error("Failed to open video "+id+"/master.m3u8 ", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if _, err = io.Copy(w, object); err != nil {
		log.Error("Unable to stream video master", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type VideoGetSourceHandler struct {
	S3Client  clients.IS3Client
	VideosDAO *dao.VideosDAO
	UUIDGen   clients.IUUIDGenerator
}

// VideoGetSourceHandler godoc
// @Summary Get video master
// @Description Get video master
// @Tags video
// @Produce plain
// @Param id path string true "Video ID"
// @Success 200 {string} string "MP4 video source"
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /api/v1/videos/{id}/streams/source.mp4 [get]
func (v VideoGetSourceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Debug("GET VideoGetMasterHandler - parameters ", vars)

	id := vars["id"]
	if !v.UUIDGen.IsValidUUID(id) {
		log.Error("Invalid id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	video, err := v.VideosDAO.GetVideo(r.Context(), id)
	if err != nil {
		log.Error("Failed to read video details "+id+"/source.mp4 ", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	videoTitle := video.Title
	if !strings.HasSuffix(videoTitle, ".mp4") {
		videoTitle += ".mp4"
	}
	_range := r.Header.Get("Range")
	var object *s3.GetObjectOutput
	if _range == "" {
		object, err = v.S3Client.GetObjectFull(r.Context(), id+"/source.mp4")
	} else {

		object, err = v.S3Client.GetObjectRange(r.Context(), id+"/source.mp4", _range)

	}

	if err != nil {
		log.Error("Failed to open video "+id+"/source.mp4 ", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", object.ContentLength))
	w.Header().Set("Content-Type", "video/mp4")
	if _range != "" {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", object.ContentLength))
		if object.ContentRange != nil {
			w.Header().Set("Content-Range", *object.ContentRange)
		}
		if object.AcceptRanges != nil {
			w.Header().Set("Accept-Ranges", *object.AcceptRanges)
		}
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.Header().Set("Content-Disposition", "attachment; filename="+videoTitle)
	}
	var n int64
	if n, err = io.Copy(w, object.Body); err != nil {
		fmt.Println(n)
		log.Error("Unable to stream video master", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else {
		fmt.Println(n)
	}
}

type VideoGetSubPartHandler struct {
	S3Client         clients.IS3Client
	UUIDGen          clients.IUUIDGenerator
	ServiceDiscovery clients.ServiceDiscovery
}

// VideoGetSubPartHandler godoc
// @Summary Get sub part stream video
// @Description Get sub part stream video
// @Tags video
// @Produce plain
// @Param id path string true "Video ID"
// @Param quality path string true "Video quality"
// @Param filename path string true "Video sub part name"
// @Param filter query []string false "List of required filters"
// @Success 200 {string} string "Video sub part (.ts)"
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /api/v1/videos/{id}/streams/{quality}/{filename} [get]
func (v VideoGetSubPartHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := r.URL.Query()
	log.Debug("GET VideoGetSubPartHandler - Parameters: ", vars)

	id := vars["id"]
	if !v.UUIDGen.IsValidUUID(id) {
		log.Error("Invalid id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	quality := vars["quality"]
	filename := vars["filename"]
	transformers := query["filter"]
	s3VideoPath := id + "/" + quality + "/" + filename

	if strings.Contains(filename, "segment_index") || transformers == nil {
		object, err := v.S3Client.GetObject(r.Context(), id+"/"+quality+"/"+filename)
		if err != nil {
			log.Error("Failed to open video videoPath", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if _, err := io.Copy(w, object); err != nil {
			log.Error("Unable to stream subpart", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		// Add metrics (should be move into transformations service implem)
		for _, service := range transformers {
			if service == "gray" {
				metrics.CounterVideoTransformGray.Inc()
			} else if service == "flip" {
				metrics.CounterVideoTransformFlip.Inc()
			}
		}
		_range := r.Header.Get("Range")
		videoPart, err := v.getVideoPart(r.Context(), s3VideoPath, _range, transformers, w)
		if err != nil {
			log.Error("Cannot get video part : ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(w, videoPart); err != nil {
			log.Error("Unable to stream subpart", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (v VideoGetSubPartHandler) getVideoPart(ctx context.Context, s3VideoPath, rangeBytes string, transformers []string, w http.ResponseWriter) (io.Reader, error) {
	if len(transformers) == 0 {
		// Retrieve the video part from aws S3
		var err error
		var videoPart io.Reader
		if rangeBytes != "" {
			var video *s3.GetObjectOutput
			video, err = v.S3Client.GetObjectRange(ctx, s3VideoPath, rangeBytes)
			if err != nil {
				log.Error("Failed to get video from S3 : ", err)
				return nil, err
			}
			videoPart = video.Body
			w.WriteHeader(http.StatusPartialContent)
			w.Header().Set("Content-Length", fmt.Sprintf("%d", video.ContentLength))
			if video.ContentRange != nil {
				w.Header().Set("Content-Range", *video.ContentRange)
			}
			if video.AcceptRanges != nil {
				w.Header().Set("Accept-Ranges", *video.AcceptRanges)
			}
		} else {
			videoPart, err = v.S3Client.GetObject(ctx, s3VideoPath)
		}
		if err != nil {
			log.Error("Failed to get video from S3 : ", err)
			return nil, err
		}
		return videoPart, nil

	} else {
		// Ask for video part transformation
		start := time.Now()

		// Connect to RPC Client
		clientRPC, err := v.connectClientRPC(transformers[len(transformers)-1])
		if err != nil {
			log.Error("Cannot connect to RPC client : ", err)
			return nil, err
		}

		// Ask RPC Client for video transformation
		request := transformer.TransformVideoRequest{
			Videopath:       s3VideoPath,
			TransformerList: transformers,
		}
		streamResponse, err := clientRPC.TransformVideo(ctx, &request)
		if err != nil {
			log.Error("Failed to transform video : ", err)
			return nil, err
		}

		var videoPart bytes.Buffer
		for {
			res, err := streamResponse.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Error("Failed to receive stream : ", err)
				return nil, err
			}

			if res != nil {
				_, err := videoPart.Write(res.Chunk)
				if err != nil {
					log.Error("Failed to write : ", err)
					return nil, err
				}
			}
		}

		log.Debug("transformation execution time : ", time.Since(start).Seconds())
		metrics.StoreTranformationTime(start, transformers)
		return &videoPart, nil
	}
}

func (v VideoGetSubPartHandler) connectClientRPC(clientName string) (transformer.TransformerServiceClient, error) {
	// Retrieve service address and port
	tfServices, err := v.ServiceDiscovery.GetTransformationService(clientName)
	if err != nil {
		log.Errorf("Cannot get address for service name %v : %v", clientName, err)
		return nil, err
	}

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(tfServices, opts)
	if err != nil {
		log.Errorf("Cannot open TCP connection with grpc %v transformer server : %v", clientName, err)
		return nil, err
	}

	return transformer.NewTransformerServiceClient(conn), nil
}

type VideoGetSubtitlesHandler struct {
	S3Client         clients.IS3Client
	UUIDGen          clients.IUUIDGenerator
	ServiceDiscovery clients.ServiceDiscovery
}

// VideoGetSubtitlesHandler godoc
// @Summary Get subtitles for video
// @Description Get subtitles for video
// @Tags video, subtitles
// @Produce plain
// @Param id path string true "Video ID"
// @Param filename path string true "Subtitles file nams"
// @Success 200 {string} string "Video subtitles"
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /api/v1/videos/{id}/subtitles/{filename} [get]
func (v VideoGetSubtitlesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Debug("GET VideoGetSubtitlesHandler - Parameters: ", vars)

	id := vars["id"]
	if !v.UUIDGen.IsValidUUID(id) {
		log.Error("Invalid id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filename := vars["filename"]
	s3VideoPath := id + "/" + filename

	var err error
	videoPart, err := v.S3Client.GetObject(r.Context(), s3VideoPath)
	if err != nil {
		log.Error("Failed to get video from S3 : ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(w, videoPart); err != nil {
		log.Error("Unable to stream subpart", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

type VideoEditDataHandler struct {
	S3Client         clients.IS3Client
	UUIDGen          clients.IUUIDGenerator
	ServiceDiscovery clients.ServiceDiscovery
	VideosDAO        *dao.VideosDAO
}

// VideoEditDataHandler godoc
// @Summary Upload subtitles for video
// @Description Upload subtitles for video
// @Tags video, subtitles
// @Accept multipart/form-data
// @Produce plain
// @Param id path string true "Video ID"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 404 {string} string
// @Failure 500 {string} string
// @Router /api/v1/videos/{id}/subtitles [get]
func (v VideoEditDataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Debug("GET VideoEditDataHandler - Parameters: ", vars)

	// Fetch title
	title := r.FormValue("title")
	if title == "" {
		log.Error("Missing file title ")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Infof("Receive video upload request with title : '%v'", title)

	id := vars["id"]
	if !v.UUIDGen.IsValidUUID(id) {
		log.Error("Invalid id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Fetch cover image. Not mandatory
	fileCover, fileHandlerCover, err := r.FormFile("cover")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		log.Error("File cover error ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if fileCover != nil {
		defer fileCover.Close()

		// Check if the received file cover is a supported image type
		if !isSupportedCoverType(fileCover) {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		// Upload cover image (if exists) on S3, update database
		coverPath, err := v.uploadCover(r.Context(), fileCover, id, fileHandlerCover)
		if err != nil {
			log.Error("Cannot upload cover image : ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = v.VideosDAO.UpdateVideoCover(r.Context(), id, coverPath)
		if err != nil {
			log.Error("Cannot update video with cover image : ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Fetch subtitles. Not mandatory
	subtitles, subtitileHandler, err := r.FormFile("subs")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		log.Error("Subtitle file error ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if subtitles != nil {
		defer subtitles.Close()
		//TODO check subtitles are valid
	}

	// Upload subtitles (if file exists) on S3
	// TODO update database?
	_, err = v.uploadSubtitles(r.Context(), subtitles, id, subtitileHandler)
	if err != nil {
		log.Error("Cannot upload subtitles : ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Check if a video with this id exists
	video, err := v.VideosDAO.GetVideo(r.Context(), id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if video.Title != title {
		// Check if a video with this title already exists
		videoConflict, err := v.VideosDAO.GetVideoFromTitle(r.Context(), title)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if videoConflict != nil && videoConflict.ID != id {
			w.WriteHeader(http.StatusConflict)
			return
		}
		err = v.VideosDAO.UpdateVideoTitle(r.Context(), id, title)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (v VideoEditDataHandler) uploadSubtitles(ctx context.Context, cover multipart.File, videoID string, fileHandler *multipart.FileHeader) (string, error) {
	subtitlesPath := ""
	if cover != nil {
		subtitlesPath = videoID + "/" + "subs" + filepath.Ext(fileHandler.Filename)
		if err := v.S3Client.PutObjectInput(ctx, cover, subtitlesPath); err != nil {
			log.Error("Cannot upload subtitles : ", err)
			return "", err
		}
	}
	return subtitlesPath, nil
}

func (v VideoEditDataHandler) uploadCover(ctx context.Context, cover multipart.File, videoID string, fileHandler *multipart.FileHeader) (string, error) {
	coverPath := ""
	if cover != nil {
		coverPath = videoID + "/" + "cover" + filepath.Ext(fileHandler.Filename)
		if err := v.S3Client.PutObjectInput(ctx, cover, coverPath); err != nil {
			log.Error("Cannot upload cover : ", err)
			return "", err
		}
	}
	return coverPath, nil
}
