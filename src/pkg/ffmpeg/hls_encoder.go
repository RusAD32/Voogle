package ffmpeg

import (
	"fmt"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func ConvertToHLS(source string, res resolution) error {
	cmd, args, err := generateCommand(source, res)
	if err != nil {
		return err
	}

	log.Debug("FFMPEG command: ", cmd, strings.Join(args, " "))
	rawOutput, err := exec.Command(cmd, args...).CombinedOutput()
	log.Debug("FFMPEG output: ", string(rawOutput[:]))
	return err
}

func generateCommand(filepath string, res resolution) (string, []string, error) {
	// working under assumption that uploaded video is already of correct format
	// do only minimal processing for the sake of speed
	if res.x < 640 && res.y < 480 {
		return "", nil, fmt.Errorf("resolution (%d,%d) is below minimal resolution (640x480)", res.x, res.y)
	}

	command := "ffmpeg"
	args := []string{"-y", "-i", filepath,  "-vcodec", "copy", "-preset", "fast"}
	sound := []string{"-map", "0:0", "-map", "0:1"}
	streamMap := "v:0,a:0"

	args = append(args, sound...)
	args = append(args, resolutionTarget...)
	args = append(args, "-c:a", "copy")
	args = append(args, "-var_stream_map", streamMap)
	args = append(args, "-master_pl_name", "master.m3u8", "-f", "hls", "-hls_time", "6", "-hls_list_size", "0", "-hls_segment_filename", "v%v/segment%d.ts", "v%v/segment_index.m3u8")

	return command, args, nil
}
