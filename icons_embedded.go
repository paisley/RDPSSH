package main

import (
    _ "embed"

    "fyne.io/fyne/v2"
)

//go:embed icons/app.png
var appIconBytes []byte

//go:embed icons/idle.png
var iconIdleBytes []byte

//go:embed icons/connected.png
var iconConnectedBytes []byte

//go:embed icons/disconnected.png
var iconDisconnectedBytes []byte

var (
    embeddedAppIcon          fyne.Resource = fyne.NewStaticResource("app.png", appIconBytes)
    embeddedIconIdle         fyne.Resource = fyne.NewStaticResource("idle.png", iconIdleBytes)
    embeddedIconConnected    fyne.Resource = fyne.NewStaticResource("connected.png", iconConnectedBytes)
    embeddedIconDisconnected fyne.Resource = fyne.NewStaticResource("disconnected.png", iconDisconnectedBytes)
)
