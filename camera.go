package videoinsight

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// CameraTime is a time.Time wrapper for unmarshalling the camera API's non-standard time format
type CameraTime time.Time

// UnmarshalJSON implements json.Unmarshaler
func (t *CameraTime) UnmarshalJSON(b []byte) error {
	if string(b) == `"0001-01-01T00:00:00"` {
		*t = CameraTime(time.Time{})
		return nil
	}

	tm, err := time.Parse(`"2006-01-02T15:04:05"`, string(b))
	if err != nil {
		return err
	}

	*t = CameraTime(tm)
	return nil
}

// Stream represents a Camera's video stream
type Stream struct {
	ProfileNumber int
	Width         int
	Height        int
	FrameRate     float64
	Bandwidth     float64
	Codec         string `json:"CodecFourCC"`
	LastRefreshed CameraTime
}

// Camera represents a camera
type Camera struct {
	ID                   int
	Name                 string
	IPAddress            net.IP
	Model                string
	ServerID             int
	Type                 string `json:"CameraType"`
	Is360                bool
	IsPTZ                bool `json:"IsPtz"`
	DisplayOrder         int
	Streams              map[string]*Stream `json:"StreamsInfo,omitempty"`
	IsNvr                bool
	IsLiveAudioEnabled   bool
	IsRecordAudioEnabled bool
}

// Cameras returns a list of Cameras
func (c *Client) Cameras() ([]*Camera, error) {
	cameraURL := &url.URL{
		Scheme: c.proto,
		Host:   fmt.Sprintf("%s:%d", c.host, c.port),
		Path:   "/api/v1/cameras",
	}
	if c.token != "" {
		cameraURL.RawQuery = url.Values{"token": []string{c.token}}.Encode()
	}

	resp, err := http.Get(cameraURL.String())
	if err != nil {
		return nil, fmt.Errorf("could not complete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, AuthenticationError(resp.Status)
		}
		return nil, UnknownError(resp.Status)
	}

	var cameras []*Camera
	d := json.NewDecoder(resp.Body)
	if err = d.Decode(&cameras); err != nil {
		return nil, fmt.Errorf("could not parse response: %w", err)
	}

	return cameras, nil
}

// Snapshot returns a live snapshot (JPEG bytes) of the camera with the given id
func (c *Client) Snapshot(id int) ([]byte, error) {
	snapshotURL := &url.URL{
		Scheme: c.proto,
		Host:   fmt.Sprintf("%s:%d", c.host, c.port),
		Path:   fmt.Sprintf("/api/v1/video/%d/keyframe", id),
	}
	if c.token != "" {
		snapshotURL.RawQuery = url.Values{"token": []string{c.token}}.Encode()
	}

	resp, err := http.Get(snapshotURL.String())
	if err != nil {
		return nil, fmt.Errorf("could not complete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, AuthenticationError(resp.Status)
		}
		return nil, UnknownError(resp.Status)
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read body: %w", err)
	}

	return buf, nil
}
