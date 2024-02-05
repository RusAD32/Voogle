package main

import (
	"context"
	"github.com/Sogilis/Voogle/src/cmd/encoder/config"
	"github.com/Sogilis/Voogle/src/pkg/clients"
	"github.com/Sogilis/Voogle/src/pkg/ffmpeg"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func downloadFile(ctx context.Context, s3 clients.IS3Client, id, filename string) {
	videoObj, err := s3.GetObject(ctx, id+"/"+filename)
	if err != nil {
		log.Fatal("Fail get video ", err)
	}
	vidFile, err := os.Create(filename)
	_, err = io.Copy(vidFile, videoObj)
	if err != nil {
		log.Fatal("Fail write video ", err)
	}
	err = vidFile.Close()
	if err != nil {
		log.Fatal("Fail close video ", err)
	}
}

func main() {
	log.Info("Starting Voogle encoder")

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("Failed to parse Env var ", err)
	}
	if cfg.DevMode {
		log.SetLevel(log.DebugLevel)
	}
	os.Mkdir("tmp", 0755)
	vidIds := []string{"b6b863b1-14a1-4685-8097-5ac834b742f8"}

	// S3 client to access the videos
	s3Client, err := clients.NewS3Client(cfg.S3Host, cfg.S3Region, cfg.S3Bucket, cfg.S3AuthKey, cfg.S3AuthPwd)
	if err != nil {
		log.Fatal("Fail to create S3Client ", err)
	}
	ctx := context.Background()
	for _, vid := range vidIds {
		log.Info("Downloading ", vid, "...")
		downloadFile(ctx, s3Client, vid, "source.mp4")
		source := "source.mp4"
		res, err := ffmpeg.ExtractResolution(source)
		if err != nil {
			log.Fatal("Fail to get resolution ", err)
		}
		targetRes := ffmpeg.Resolution{854, 480, 800000}
		log.Info("Converting ", vid, "...")
		err = ffmpeg.ConvertToHLSWithDownsample(source, res, targetRes)
		if err != nil {
			panic(err)
		}
		playlistObj, err := s3Client.GetObject(ctx, vid+"/master.m3u8")
		if err != nil {
			log.Fatal("Fail get playlist ", err)
		}
		f, err := os.Open("tmp/master.m3u8")
		if err != nil {
			log.Fatal("Fail read new playlist ", err)
		}
		newPlaylist, err := io.ReadAll(f)
		master, err := io.ReadAll(playlistObj)
		masterLines := strings.Split(string(master), "\n")[:5]
		nps := strings.Split(string(newPlaylist), "\n")[2:]
		masterLines = append(masterLines, nps...)
		f.Close()

		err = os.WriteFile("tmp/master.m3u8", []byte(strings.Join(masterLines, "\n")), 0644)
		if err != nil {
			log.Fatal("Fail writing new playlist ", err)
		}
		err = filepath.WalkDir("tmp",
			func(path string, info os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if path == "." || (!strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".m3u8") && !strings.HasSuffix(path, ".jpeg")) {
					log.Debug("Skipping ", path)
					return nil
				}
				f, err := os.Open(path)
				if err != nil {
					return err
				}
				defer func() { _ = f.Close() }()
				return s3Client.PutObjectInput(context.Background(), f, strings.Replace(filepath.Join(vid, path[4:]), "\\", "/", -1))
			})
		if err != nil {
			panic(err)
		}
		os.RemoveAll("tmp")
		os.Remove("source.mp4")
	}
}
