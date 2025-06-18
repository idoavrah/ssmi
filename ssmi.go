package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/google/uuid"
	"github.com/idoavrah/ssmi/internal"
	"github.com/posthog/posthog-go"
)

var version = "X.Y.Z"

func main() {

	versionFlag := flag.Bool("version", false, "Print the version")
	helpFlag := flag.Bool("help", false, "Show help message")
	offlineFlag := flag.Bool("offline", false, "Run in offline mode (no telemetry)")

	flag.Parse()

	if *helpFlag {
		println("SSMI v" + version + "\n")
		flag.Usage()
		os.Exit(0)
	}

	if *versionFlag {
		println("SSMI v" + version + "\n")
		os.Exit(0)
	}

	if !*offlineFlag {
		POSTHOG_API_KEY := "phc_tjGzx7V6Y85JdNfOFWxQLXo5wtUs6MeVLvoVfybqz09"
		disableGeoIP := false

		client, _ := posthog.NewWithConfig(POSTHOG_API_KEY, posthog.Config{Endpoint: "https://app.posthog.com", DisableGeoIP: &disableGeoIP})
		defer client.Close()

		client.Enqueue(posthog.Capture{
			DistinctId: uuid.New().String(),
			Event:      "ssmi started",
			Properties: posthog.NewProperties().
				Set("ssmi_version", version).
				Set("platform", runtime.GOOS+"-"+runtime.GOARCH)})
	}

	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		fmt.Println("AWS_PROFILE is not set")
		os.Exit(1)
	}

	internal.StartApplication(profile)
}
