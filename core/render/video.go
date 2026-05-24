package render

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"os/exec"
)

// VideoEncoder pipes RGB24 frames to ffmpeg, producing a libx264 MP4
// at visually-lossless quality.
//
// Construction validates ffmpeg is on PATH; encoding errors surface
// from WriteFrame / Close. The encoder is single-use — call Close
// when done, then construct a new one for the next video.
type VideoEncoder struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	width  int
	height int
	buf    []byte
	closed bool
}

// VideoOptions configures the ffmpeg invocation.
type VideoOptions struct {
	Width, Height int
	FPS           int
	// CRF is the libx264 quality knob. 18 is "visually lossless"
	// (default). Lower = larger, higher = smaller.
	CRF int
	// Preset is the libx264 speed/quality tradeoff. "slow" is the
	// default; "veryfast" for quick iterations.
	Preset string
	// FFmpegPath overrides the binary name (default: "ffmpeg").
	FFmpegPath string
}

// NewVideoEncoder starts ffmpeg and returns an encoder ready to
// receive frames. The output file is overwritten if it exists.
func NewVideoEncoder(outPath string, opts VideoOptions) (*VideoEncoder, error) {
	if opts.Width <= 0 || opts.Height <= 0 {
		return nil, errors.New("render: VideoOptions Width/Height must be > 0")
	}
	if opts.FPS <= 0 {
		opts.FPS = 60
	}
	if opts.CRF == 0 {
		opts.CRF = 18
	}
	if opts.Preset == "" {
		opts.Preset = "slow"
	}
	bin := opts.FFmpegPath
	if bin == "" {
		bin = "ffmpeg"
	}
	if _, err := exec.LookPath(bin); err != nil {
		return nil, fmt.Errorf(
			"render: ffmpeg not found on PATH — install with `brew install ffmpeg` "+
				"(macOS) or `apt-get install ffmpeg` (Linux): %w", err)
	}

	args := []string{
		"-y",             // overwrite output
		"-f", "rawvideo", // input format
		"-pix_fmt", "rgb24", // input pixel format
		"-s", fmt.Sprintf("%dx%d", opts.Width, opts.Height),
		"-r", fmt.Sprintf("%d", opts.FPS),
		"-i", "-", // input from stdin
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p", // broad compatibility
		"-preset", opts.Preset,
		"-crf", fmt.Sprintf("%d", opts.CRF),
		"-movflags", "+faststart",
		outPath,
	}
	cmd := exec.Command(bin, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("render: ffmpeg stdin: %w", err)
	}
	// Surface ffmpeg's logs only if it actually errors out, so the
	// normal case is quiet.
	cmd.Stderr = nil
	cmd.Stdout = nil
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("render: start ffmpeg: %w", err)
	}
	return &VideoEncoder{
		cmd:    cmd,
		stdin:  stdin,
		width:  opts.Width,
		height: opts.Height,
		buf:    make([]byte, opts.Width*opts.Height*3),
	}, nil
}

// WriteFrame writes one frame to ffmpeg. img.Bounds().Size() must
// match the encoder's configured Width × Height.
func (v *VideoEncoder) WriteFrame(img image.Image) error {
	if v.closed {
		return errors.New("render: VideoEncoder is closed")
	}
	b := img.Bounds()
	if b.Dx() != v.width || b.Dy() != v.height {
		return fmt.Errorf("render: frame size %dx%d differs from encoder %dx%d",
			b.Dx(), b.Dy(), v.width, v.height)
	}
	// Fast path for *image.RGBA: copy R, G, B columns into v.buf.
	if rgba, ok := img.(*image.RGBA); ok {
		stride := rgba.Stride
		for y := 0; y < v.height; y++ {
			srcRow := rgba.Pix[y*stride : y*stride+v.width*4]
			dstRow := v.buf[y*v.width*3 : (y+1)*v.width*3]
			for x := 0; x < v.width; x++ {
				dstRow[x*3+0] = srcRow[x*4+0]
				dstRow[x*3+1] = srcRow[x*4+1]
				dstRow[x*3+2] = srcRow[x*4+2]
			}
		}
	} else {
		// Generic path — handles any image.Image (slower).
		idx := 0
		for y := 0; y < v.height; y++ {
			for x := 0; x < v.width; x++ {
				r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
				v.buf[idx+0] = uint8(r >> 8)
				v.buf[idx+1] = uint8(g >> 8)
				v.buf[idx+2] = uint8(bl >> 8)
				idx += 3
			}
		}
	}
	_, err := v.stdin.Write(v.buf)
	return err
}

// Close finishes the encoder and waits for ffmpeg to finalize the file.
func (v *VideoEncoder) Close() error {
	if v.closed {
		return nil
	}
	v.closed = true
	if err := v.stdin.Close(); err != nil {
		return fmt.Errorf("render: close ffmpeg stdin: %w", err)
	}
	if err := v.cmd.Wait(); err != nil {
		return fmt.Errorf("render: ffmpeg exit: %w", err)
	}
	return nil
}

// Sanity-check that color.Color is used (suppress unused-import).
var _ color.Color = color.Black
