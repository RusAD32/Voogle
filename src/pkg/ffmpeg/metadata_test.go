package ffmpeg

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ExtractResolution(t *testing.T) {
	//  LFS Github quota is not really fair, is uses bandwidth even when we use it within the Github Actions CI
	//  So until we've found an alternative, we won't test the video processing part
	t.SkipNow()
	cases := []struct {
		GivenPath        string
		GivenFilename    string
		ExpectResolution Resolution
		ExpectError      bool
	}{
		{
			GivenPath:        "../../../../samples/", // FIXME(JPR): Root of the project from the test file (We need may need a better way to address these)
			GivenFilename:    "320x240_testvideo.mp4",
			ExpectResolution: Resolution{320, 240, 0},
			ExpectError:      false,
		},
		{
			GivenPath:        "../../../../samples/", // FIXME(JPR): Root of the project from the test file (We need may need a better way to address these)
			GivenFilename:    "960x400_ocean_with_audio.avi",
			ExpectResolution: Resolution{960, 400, 0},
			ExpectError:      false,
		},
		{
			GivenPath:        "../../../../samples/", // FIXME(JPR): Root of the project from the test file (We need may need a better way to address these)
			GivenFilename:    "4K-10bit.mkv",
			ExpectResolution: Resolution{3840, 2160, 0},
			ExpectError:      false,
		},
		{
			GivenPath:        "../../../../samples/", // FIXME(JPR): Root of the project from the test file (We need may need a better way to address these)
			GivenFilename:    "960x400_ocean_with_audio.mkv",
			ExpectResolution: Resolution{960, 400, 0},
			ExpectError:      false,
		},
		{
			GivenPath:        "../../../../samples/", // FIXME(JPR): Root of the project from the test file (We need may need a better way to address these)
			GivenFilename:    "1280x720_2mb.mp4",
			ExpectResolution: Resolution{1280, 720, 0},
			ExpectError:      false,
		},
	}

	for _, tt := range cases {
		t.Run("Extract Resolution from video "+tt.GivenFilename, func(t *testing.T) {
			res, err := ExtractResolution(tt.GivenPath + tt.GivenFilename)
			if tt.ExpectError {
				require.NotNil(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, res.X, tt.ExpectResolution.X)
			require.Equal(t, res.Y, tt.ExpectResolution.Y)
		})
	}
}

func Test_videoHaveSound(t *testing.T) {
	t.SkipNow()
	cases := []struct {
		Name          string
		GivenFilepath string
		ExpectSound   bool
		ExpectError   bool
	}{
		{Name: "With Sound", GivenFilepath: "../../../samples/1280x720_2mb.mp4", ExpectSound: true, ExpectError: false},
		{Name: "Without Sound", GivenFilepath: "../../../samples/video_without_sound.mp4", ExpectSound: false, ExpectError: false},
	}

	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			sound, err := CheckContainsSound(tt.GivenFilepath)
			if tt.ExpectError {
				require.NotNil(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.ExpectSound, sound)
		})
	}
}
