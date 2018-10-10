//go:generate goagen bootstrap -d github.com/tinakurian/build-tool-detector/design

package main

import (
	"flag"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	"github.com/tinakurian/build-tool-detector/app"
	controllers "github.com/tinakurian/build-tool-detector/controllers"
	logorus "github.com/tinakurian/build-tool-detector/log"
	"os"
)

var (
	port           = flag.String("PORT", "8080", "port")
	ghClientID     = flag.String("GH_CLIENT_ID", "", "Github Client ID")
	ghClientSecret = flag.String("GH_CLIENT_SECRET", "", "Github Client Secret")
	sentryDSN      = flag.String("SENTRY_DSN", "", "Sentry DSN")
)

func main() {

	flag.Parse()
	if *ghClientID == "" || *ghClientSecret == "" {
		logorus.Logger().
			WithField("ghClientID", ghClientID).
			WithField("ghClientSecret", ghClientSecret).
			Fatalf("Cannot run application without ghClientID and ghClientSecret")
	}

	// Export Sentry DSN for logging
	err := os.Setenv("BUILD_TOOL_DETECTOR_SENTRY_DSN", *sentryDSN)
	if err != nil {
		logorus.Logger().
			WithField("sentryDSN", sentryDSN).
			Fatalf("Failed to set environment variable for sentryDSN: %v", sentryDSN)
	}

	// Create service
	service := goa.New("build-tool-detector")

	// Mount middleware
	service.Use(middleware.RequestID())
	service.Use(middleware.LogRequest(true))
	service.Use(middleware.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	// Mount "build-tool-detector" controller
	c := controllers.NewBuildToolDetectorController(service, *ghClientID, *ghClientSecret)
	app.MountBuildToolDetectorController(service, c)

	cs := controllers.NewSwaggerController(service)
	app.MountSwaggerController(service, cs)

	// Start service
	if err := service.ListenAndServe(":" + *port); err != nil {
		service.LogError("startup", "err", err)
	}
}
