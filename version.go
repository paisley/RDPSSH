package main

// Variables to be set via ldflags or defaults
var (
	AppName     = "RDPSSH"
	AppVersion  = "v1.0"
	BuildCommit = "unknown"
	BuildDate   = "unknown"
	DocURL      = "https://github.com/paisley/rdpssh"
	IssueURL    = "https://github.com/paisley/rdpssh/issues"

	// Icon Paths (Relative to binary)
	IconPathApp          = "icons/app.png"
	IconPathIdle         = "icons/idle.png"
	IconPathConnected    = "icons/connected.png"
	IconPathDisconnected = "icons/disconnected.png"

	// Status Text
	StatusTextConnected    = "Status: Connected to %s"
	StatusTextDisconnected = "Status: Disconnected"

	// About Text
	AboutText = "A simple RDP-over-SSH tunnel manager with certificate authentication.\n\nBuild: " + BuildCommit + " (" + BuildDate + ")\n\n" + DocURL
)
