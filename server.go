package videoinsight

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// ServerTime is a time.Time wrapper for unmarshalling the server API's non-standard time format
type ServerTime time.Time

// UnmarshalJSON implements json.Unmarshaler
func (t *ServerTime) UnmarshalJSON(b []byte) error {
	if string(b) == `"0001-01-01T00:00:00"` {
		*t = ServerTime(time.Time{})
		return nil
	}
	tm, err := time.Parse(`"2006-01-02T15:04:05-07:00"`, string(b))
	if err != nil {
		return err
	}

	*t = ServerTime(tm)
	return nil
}

// NetworkProfile represents network information for the server
type NetworkProfile struct {
	CommandPort            int
	DataPort               int
	MaximumFramesPerSecond float64
	ID                     int
	IPAddress              net.IP
	Name                   string
}

// CameraStatus represents status information about a camera
type CameraStatus struct {
	ID               int
	VideoFormat      string
	IsDisabled       bool
	Name             string `json:"CameraName"`
	FrameRate        float64
	Bandwidth        float64 `json:"NetworkBandwidth"`
	Height           int     `json:"FrameHeight"`
	Width            int     `json:"FrameWidth"`
	Codec            int     `json:"FourCC"`
	FrameRateTwo     float64 `json:"SecondStreamFrameRate"`
	BandwidthTwo     float64 `json:"SecondStreamBandwidth"`
	HeightTwo        int     `json:"SecondStreamFrameHeight"`
	WidthTwo         int     `json:"SecondStreamFrameWidth"`
	CodecTwo         int     `json:"SecondStreamFourCC"`
	IsReceivingVideo bool
	IsWritingVideo   bool
	LastReceived     ServerTime `json:"LastReceivedTime"`
	LastWrite        ServerTime
}

// ServerStatus represents status information about a server
type ServerStatus struct {
	AvailableCameras int
	Cameras          []*CameraStatus `json:"CameraStatus"`
	CPUUsage         float64
	DiskSpace        string
	TotalMemory      string
	MaxCameras       int
	MemoryUsage      float64
	SerialNumber     string
	Version          string
	NetworkStatus    int
}

// Server represents a VideoInsight server
type Server struct {
	ID                    int
	IPAddress             net.IP
	CommandPort           int
	DataPort              int
	Name                  string
	NetworkProfiles       []*NetworkProfile
	SecurityEnabled       bool
	RecordingPath         string
	ServerStatus          *ServerStatus
	LicenseType           string
	TimezoneOffsetSeconds int
}

// Servers returns a list of servers
func (c *Client) Servers() ([]*Server, error) {
	serverURL := &url.URL{
		Scheme: c.proto,
		Host:   fmt.Sprintf("%s:%d", c.host, c.port),
		Path:   "/api/v1/server",
	}
	resp, err := http.Get(serverURL.String())
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

	var servers []*Server
	d := json.NewDecoder(resp.Body)
	if err = d.Decode(&servers); err != nil {
		return nil, fmt.Errorf("could not parse response: %w", err)
	}

	return servers, nil
}
