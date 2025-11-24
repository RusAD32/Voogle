package ffmpeg

import (
	"fmt"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func ConvertToHLS(source string, res Resolution) error {
	cmd, args, err := generateCommand(source, res)
	if err != nil {
		return err
	}

	log.Debug("FFMPEG command: ", cmd, strings.Join(args, " "))
	rawOutput, err := exec.Command(cmd, args...).CombinedOutput()
	log.Debug("FFMPEG output: ", strings.Replace(string(rawOutput[:]), "\\\\", "\\", -1))
	return err
}

func generateCommand(filepath string, res Resolution) (string, []string, error) {
	// working under assumption that uploaded video is already of correct format
	// do only minimal processing for the sake of speed
	if res.X < 640 && res.Y < 480 {
		return "", nil, fmt.Errorf("Resolution (%d,%d) is below minimal Resolution (640x480)", res.X, res.Y)
	}

	command := "ffmpeg"
	args := []string{"-y", "-i", filepath, "-vcodec", "copy", "-preset", "fast"}
	sound := []string{"-map", "0:0", "-map", "0:1"}
	resolutionTarget := []string{"-c:v:0", "copy"}
	streamMap := "v:0,a:0"
	if (Resolution{X: 640, Y: 480}).GreaterResolution(res) {
		sound = append(sound, "-map", "0:0", "-map", "0:1")

		resolutionTarget = append(resolutionTarget, "-s:v:1", "640x480", "-c:v:1", "libx264", "-crf", "23")
		streamMap += " v:1,a:1"
	}

	if res.GreaterResolution(Resolution{X: 1920, Y: 1080}) {
		sound = append(sound, "-map", "0:0", "-map", "0:1")
		resolutionTarget = append(resolutionTarget, "-s:v:1", "1920x1080", "-c:v:1", "libx264", "-crf", "23")
		streamMap = streamMap + " v:2,a:2"
	}

	args = append(args, sound...)
	args = append(args, resolutionTarget...)
	args = append(args, "-c:a", "copy")
	args = append(args, "-var_stream_map", streamMap)
	args = append(args, "-master_pl_name", "master.m3u8", "-f", "hls", "-hls_time", "6", "-hls_playlist_type", "vod", "-hls_segment_type", "fmp4", "-hls_list_size", "0", "-hls_segment_filename", "v%v/segment%d.m4s", "v%v/segment_index.m3u8")

	return command, args, nil
}

func ConvertToHLSWithDownsample(source string, res Resolution, resTargets ...Resolution) error {
	cmd, args, err := generateCommandWithDownsampleNvidia(source, res, resTargets...)
	if err != nil {
		return err
	}

	log.Debug("FFMPEG command: ", cmd, strings.Join(args, " "))
	rawOutput, err := exec.Command(cmd, args...).CombinedOutput()
	log.Debug("FFMPEG output: ", string(rawOutput[:]))
	return err
}

func generateCommandWithDownsampleNvidia(filepath string, res Resolution, resTargets ...Resolution) (string, []string, error) {
	// Example of the biggest command that can be generated
	// ffmpeg -y -i <filepath> \
	//              -pix_fmt yuv420p \
	//              -vcodec libx264 \
	//              -preset fast \
	//              -g 48 -sc_threshold 0 \
	//              -map 0:0 -map 0:1 -map 0:0 -map 0:1 -map 0:0 -map 0:1 -map 0:0 -map 0:1 \
	//              -s:v:0 640x480 -c:v:0 libx264 -b:v:0 1000k \
	//              -s:v:1 1280x720 -c:v:1 libx264 -b:v:1 2000k  \
	//              -s:v:2 1920x1080 -c:v:2 libx264 -b:v:2 4000k  \
	//              -s:v:3 3840x2160 -c:v:3 libx264 -b:v:3 8000k  \
	//              -c:a aac -b:a 128k -ac 2 \
	//              -var_stream_map "v:0,a:0 v:1,a:1 v:2,a:2 v:3,a:3" \
	//              -master_pl_name master.m3u8 \
	//              -f hls -hls_time 6 -hls_list_size 0 \
	//              -hls_segment_filename "v%v/segment%d.ts" \
	//              v%v/segment_index.m3u8

	if res.X < 640 && res.Y < 480 {
		return "", nil, fmt.Errorf("Resolution (%d,%d) is below minimal Resolution (640x480)", res.X, res.Y)
	}

	command := "ffmpeg"
	args := []string{"-y", "-i", filepath, "-vcodec", "h264_nvenc", "-preset", "fast", "-g", "48", "-sc_threshold", "0"}
	sound := []string{}
	resolutionTarget := []string{}
	streamMap := ""
	for i, target := range resTargets {
		sound = append(sound, "-map", "0:0", "-map", "0:1")
		resolutionTarget = append(resolutionTarget,
			fmt.Sprintf("-s:v:%d", i),
			fmt.Sprintf("%dx%d", target.X, target.Y),
			fmt.Sprintf("-c:v:%d", i),
			"h264_nvenc",
			fmt.Sprintf("-b:v:%d", i),
			fmt.Sprintf("%d", target.Bitrate),
		)
		streamMap = streamMap + fmt.Sprintf(" v:%d,a:%d", i, i)
	}

	args = append(args, sound...)
	args = append(args, resolutionTarget...)
	args = append(args, "-c:a", "copy")
	args = append(args, "-var_stream_map", streamMap)
	args = append(args, "-master_pl_name", "master.m3u8", "-f", "hls", "-hls_time", "6", "-hls_list_size", "0", "-hls_playlist_type", "vod", "-hls_segment_type", "fmp4", "-hls_segment_filename", "tmp/v0%v/segment%d.m4s", "tmp/v0%v/segment_index.m3u8")

	return command, args, nil
}
