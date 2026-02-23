package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/icholy/digest"
)

// Camera represents a HikVision IP camera accessible over HTTP.
type Camera struct {
	Host     string
	Username string
	Password string
	client   *http.Client
}

// NewCamera creates a Camera with an HTTP client configured for digest auth.
func NewCamera(host, username, password string) *Camera {
	return &Camera{
		Host:     host,
		Username: username,
		Password: password,
		client: &http.Client{
			Transport: &digest.Transport{
				Username: username,
				Password: password,
			},
		},
	}
}

// hardwareService is the root XML envelope returned by GET /ISAPI/System/Hardware.
type hardwareService struct {
	XMLName      xml.Name      `xml:"HardwareService"`
	IrLightSwitch irLightSwitch `xml:"IrLightSwitch"`
}

// irLightSwitch is the IR LED control element nested inside HardwareService.
type irLightSwitch struct {
	Mode string `xml:"mode"`
}

// SetIRLight turns the IR illuminator on (true) or off (false).
// Calls PUT /ISAPI/System/Hardware with an IrLightSwitch XML body.
func (c *Camera) SetIRLight(on bool) error {
	mode := "close"
	if on {
		mode = "open"
	}

	payload, err := xml.Marshal(hardwareService{IrLightSwitch: irLightSwitch{Mode: mode}})
	if err != nil {
		return fmt.Errorf("marshal xml: %w", err)
	}

	url := fmt.Sprintf("http://%s/ISAPI/System/Hardware", c.Host)
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(xml.Header+string(payload)))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/xml")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("PUT %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("camera returned %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// GetIRLight returns true if the IR illuminator is currently enabled.
// Calls GET /ISAPI/System/Hardware and parses the IrLightSwitch mode.
func (c *Camera) GetIRLight() (bool, error) {
	url := fmt.Sprintf("http://%s/ISAPI/System/Hardware", c.Host)
	resp, err := c.client.Get(url)
	if err != nil {
		return false, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("camera returned %d: %s", resp.StatusCode, string(body))
	}

	var result hardwareService
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("decode response: %w", err)
	}

	return result.IrLightSwitch.Mode == "open", nil
}

func main() {
	host := flag.String("host", "", "Camera IP address (required)")
	user := flag.String("user", "admin", "Camera username")
	pass := flag.String("pass", "", "Camera password (required)")
	action := flag.String("action", "", "Action: on | off | status (required)")
	flag.Parse()

	if *host == "" || *pass == "" || *action == "" {
		fmt.Fprintf(os.Stderr, "Usage: hikvision-ir --host <IP> --user <user> --pass <pass> --action on|off|status\n")
		os.Exit(1)
	}

	cam := NewCamera(*host, *user, *pass)

	switch *action {
	case "on":
		if err := cam.SetIRLight(true); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("IR light: on")

	case "off":
		if err := cam.SetIRLight(false); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("IR light: off")

	case "status":
		on, err := cam.GetIRLight()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if on {
			fmt.Println("IR light: on")
		} else {
			fmt.Println("IR light: off")
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown action %q — must be on, off, or status\n", *action)
		os.Exit(1)
	}
}
