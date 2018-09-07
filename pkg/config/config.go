package config

import "github.com/gin-gonic/gin"

const (
	// AppName is the properly formatted app name.
	AppName = "Blaster"
	// CookiePrefix is the app-specific prefix for cookie names.
	CookiePrefix = "blaster_"
)

// IsDebugging determines whether app is running in debug mode.
var IsDebugging = gin.IsDebugging()
