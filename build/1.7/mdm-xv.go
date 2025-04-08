// mdm-xv.go
// CLI tool for token generation and device lookup (Jamf-agnostic)

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

type TokenResponse struct {
	Token   string `json:"token"`
	Expires string `json:"expires"`
}

const (
	serviceName   = "mdm_xv_token_tool"
	tokenFileName = "mdm_token.json"
	usernameKey   = "username"
	passwordKey   = "password"
	urlKey        = "url"
)

func getTokenFilePath() string {
	dir, _ := os.UserHomeDir()
	return filepath.Join(dir, "Library", "Application Support", "mdm-xv", tokenFileName)
}

func readCachedToken() (*TokenResponse, error) {
	filePath := getTokenFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var token TokenResponse
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	expTime, err := time.Parse(time.RFC3339, token.Expires)
	if err != nil || time.Now().UTC().After(expTime) {
		return nil, fmt.Errorf("token expired")
	}
	return &token, nil
}

func writeCachedToken(token TokenResponse) error {
	filePath := getTokenFilePath()
	os.MkdirAll(filepath.Dir(filePath), 0755)
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

func getCredentialsFromKeyring() (string, string, string, error) {
	u, err := keyring.Get(serviceName, usernameKey)
	if err != nil {
		return "", "", "", err
	}
	p, err := keyring.Get(serviceName, passwordKey)
	if err != nil {
		return "", "", "", err
	}
	url, err := keyring.Get(serviceName, urlKey)
	if err != nil {
		return "", "", "", err
	}
	return u, p, url, nil
}

func saveCredentialsToKeyring(username, password, url string) error {
	if err := keyring.Set(serviceName, usernameKey, username); err != nil {
		return err
	}
	if err := keyring.Set(serviceName, passwordKey, password); err != nil {
		return err
	}
	return keyring.Set(serviceName, urlKey, url)
}

func promptCredentials() (string, string, string) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter username: ")
	u, _ := reader.ReadString('\n')
	u = strings.TrimSpace(u)

	fmt.Print("Enter password: ")
	bp, _ := term.ReadPassword(int(os.Stdin.Fd()))
	p := strings.TrimSpace(string(bp))
	fmt.Println()

	fmt.Print("Enter server URL: ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)

	return u, p, url
}

func getBearerToken(username, password, url string) (*TokenResponse, error) {
	if url == "" {
		return nil, errors.New("URL is empty")
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", url+"/api/v1/auth/token", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("auth failed: %s", resp.Status)
	}
	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

func getNested(data map[string]interface{}, path string) string {
	keys := strings.Split(path, ".")
	current := data
	for _, key := range keys {
		if val, ok := current[key]; ok {
			if m, ok := val.(map[string]interface{}); ok {
				current = m
			} else {
				return fmt.Sprintf("%v", val)
			}
		} else {
			return ""
		}
	}
	return ""
}

func queryDevicesByFilter(url, token, filter string) ([]byte, error) {
	cmd := exec.Command("curl", "-G",
		"-H", "Authorization: Bearer "+token,
		"--data-urlencode", "filter="+filter,
		"--data-urlencode", "section=GENERAL",
		"--data-urlencode", "section=HARDWARE",
		"--data-urlencode", "section=USER_AND_LOCATION",
		"--data-urlencode", "section=OPERATING_SYSTEM",
		"--data-urlencode", "section=DISK_ENCRYPTION",
		url+"/api/v1/computers-inventory",
	)
	return cmd.Output()
}

func resetCredentials() error {
	_ = keyring.Delete(serviceName, usernameKey)
	_ = keyring.Delete(serviceName, passwordKey)
	_ = keyring.Delete(serviceName, urlKey)
	return os.Remove(getTokenFilePath())
}

func printDeviceTable(devices []map[string]interface{}) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Name", "Serial", "Managed", "Model", "OS", "Last Check-In",
		"Inventory", "Enrollment", "RAM (MB)", "MAC", "FileVault", "Email",
	})
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
	)

	for _, d := range devices {
		fv := getNested(d, "diskEncryption.fileVault2Enabled")
		fvColor := tablewriter.Colors{}
		if fv == "true" {
			fvColor = tablewriter.Colors{tablewriter.FgHiGreenColor}
		} else {
			fvColor = tablewriter.Colors{tablewriter.FgHiRedColor}
		}
		row := []string{
			getNested(d, "general.name"),
			getNested(d, "hardware.serialNumber"),
			getNested(d, "general.remoteManagement.managed"),
			getNested(d, "hardware.model"),
			getNested(d, "operatingSystem.version"),
			getNested(d, "general.lastContactTime"),
			getNested(d, "general.reportDate"),
			getNested(d, "general.lastEnrolledDate"),
			getNested(d, "hardware.totalRamMegabytes"),
			getNested(d, "hardware.macAddress"),
			fv,
			getNested(d, "userAndLocation.email"),
		}
		table.Rich(row, []tablewriter.Colors{
			{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, fvColor, {},
		})
	}
	table.Render()
}

func main() {
	app := &cli.App{
		Name:  "mdm-xv",
		Usage: "Extended view of Apple device data via Jamf API",
		Commands: []*cli.Command{
			{
				Name:  "token",
				Usage: "Generate or reuse bearer token",
				Action: func(c *cli.Context) error {
					if cached, err := readCachedToken(); err == nil {
						expTime, _ := time.Parse(time.RFC3339, cached.Expires)
						fmt.Println("✅ Valid token found:", cached.Token)
						fmt.Println("Expires:", expTime.Format("2006-01-02 15:04:05 MST"))
						return nil
					}
					u, p, url, err := getCredentialsFromKeyring()
					if err != nil || url == "" {
						u, p, url = promptCredentials()
						saveCredentialsToKeyring(u, p, url)
					}
					token, err := getBearerToken(u, p, url)
					if err != nil {
						return err
					}
					writeCachedToken(*token)
					expTime, _ := time.Parse(time.RFC3339, token.Expires)
					fmt.Println("✅ New token:", token.Token)
					fmt.Println("Expires:", expTime.Format("2006-01-02 15:04:05 MST"))
					return nil
				},
			},
			{
				Name:  "lookup",
				Usage: "Look up a device by serial number",
				Action: func(c *cli.Context) error {
					tokenData, err := readCachedToken()
					if err != nil {
						return err
					}
					url, err := keyring.Get(serviceName, urlKey)
					if err != nil {
						return err
					}
					reader := bufio.NewReader(os.Stdin)
					fmt.Print("🔍 Enter Serial Number: ")
					serial, _ := reader.ReadString('\n')
					serial = strings.TrimSpace(serial)
					out, err := queryDevicesByFilter(url, tokenData.Token, fmt.Sprintf(`hardware.serialNumber=="%s"`, serial))
					if err != nil {
						return err
					}
					var response struct {
						Results []map[string]interface{} `json:"results"`
					}
					json.Unmarshal(out, &response)
					if len(response.Results) == 0 {
						fmt.Println("❌ No device found.")
						return nil
					}
					printDeviceTable(response.Results)
					return nil
				},
			},
			{
				Name:  "email",
				Usage: "Look up devices by user email",
				Action: func(c *cli.Context) error {
					tokenData, err := readCachedToken()
					if err != nil {
						return err
					}
					url, err := keyring.Get(serviceName, urlKey)
					if err != nil {
						return err
					}
					reader := bufio.NewReader(os.Stdin)
					fmt.Print("📧 Enter Email Address: ")
					email, _ := reader.ReadString('\n')
					email = strings.TrimSpace(email)
					out, err := queryDevicesByFilter(url, tokenData.Token, fmt.Sprintf(`userAndLocation.email=="%s"`, email))
					if err != nil {
						return err
					}
					var response struct {
						Results []map[string]interface{} `json:"results"`
					}
					json.Unmarshal(out, &response)
					if len(response.Results) == 0 {
						fmt.Println("❌ No devices found for this email.")
						return nil
					}
					printDeviceTable(response.Results)
					return nil
				},
			},
			{
				Name:  "reset",
				Usage: "Clear all stored credentials and token",
				Action: func(c *cli.Context) error {
					fmt.Println("⚠️  Resetting credentials and token...")
					return resetCredentials()
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println("❌ Error:", err)
		os.Exit(1)
	}
}
