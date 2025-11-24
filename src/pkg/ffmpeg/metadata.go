package ffmpeg

import (
	"os/exec"
	"strconv"
	"strings"
)

type Resolution struct {
	X       uint64
	Y       uint64
	Bitrate uint64
}

func (r Resolution) GreaterOrEqualResolution(input Resolution) bool {
	return r.X >= input.X && r.Y >= input.Y
}

func (r Resolution) GreaterResolution(input Resolution) bool {
	return r.X > input.X && r.Y > input.Y
}

func CheckContainsSound(filepath string) (bool, error) {
	// sh -c "ffmpeg -i <filepath> 2>&1 | grep Audio | awk '{print $0}' | tr -d ,"
	rawOutput, err := exec.Command("sh", "-c", "ffmpeg -i "+filepath+" 2>&1 | grep Audio | awk '{print $0}' | tr -d ,").CombinedOutput()
	if err != nil {
		return false, err
	}
	haveSound := len(rawOutput) != 0
	return haveSound, err
}

// Extract Resolution of the video
func ExtractResolution(filepath string) (Resolution, error) {
	// ffprobe -v error -select_streams v:0 -show_entries stream=width,height -of csv=s=x:p=0 <filepath>
	rawOutput, err := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height,bit_rate", "-of", "csv=s=x:p=0", filepath).Output()
	if err != nil {
		return Resolution{}, err
	}
	output := string(rawOutput[:])

	//Sometimes, ffprobe return several Resolution despite the video only have one video track
	firstLine := strings.Trim(strings.Split(output, "\n")[0], "\r") // We get: XRESxYRES

	splitResolution := strings.Split(firstLine, "x")
	var x, y, br uint64
	if x, err = strconv.ParseUint(splitResolution[0], 10, 32); err != nil {
		return Resolution{}, err
	}
	if y, err = strconv.ParseUint(splitResolution[1], 10, 32); err != nil {
		return Resolution{}, err
	}
	if br, err = strconv.ParseUint(splitResolution[2], 10, 32); err != nil {
		return Resolution{}, err
	}

	return Resolution{x, y, br}, nil
}
