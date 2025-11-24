package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	nativeDialog "github.com/sqweek/dialog"
)

var (
	logWindow fyne.Window

	statusBinding binding.String
	logBinding    binding.String
	logMutex      sync.Mutex

	connectFunc    func()
	disconnectFunc func()
	isTunnelActive bool
)

// uiWriter redirects log output to the UI log window
type uiWriter struct {
	logFunc func(string)
}

func (w *uiWriter) Write(p []byte) (n int, err error) {
	w.logFunc(strings.TrimRight(string(p), "\n"))
	return len(p), nil
}

func main() {
	// Enforce single instance via TCP listener
	l, err := net.Listen("tcp", "127.0.0.1:44444")
	if err != nil {
		nativeDialog.Message("Application is already running. Only one instance can run at a time. Check the system tray if the app is not visible.").Title("Error").Error()
		os.Exit(1)
	}
	defer l.Close()

	cfg, _ := LoadConfig()
	cfgPath, _ := GetConfigPath()

	a := app.New()
	a.Settings().SetTheme(&DarkGreenTheme{})

	var appIcon fyne.Resource
	if icon, err := fyne.LoadResourceFromPath(IconPathApp); err == nil {
		appIcon = icon
		a.SetIcon(icon)
	}

	logFilePath := filepath.Join(filepath.Dir(cfgPath), "app.log")
	logFile, err := os.Create(logFilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log file: %v\n", err)
	}

	statusBinding = binding.NewString()
	statusBinding.Set("Status: Ready")

	logBinding = binding.NewString()
	logBinding.Set(fmt.Sprintf("[%s] App started...\n", time.Now().Format("2006-01-02 15:04:05")))

	// Aggregate log messages from multiple sources (UI, file, stderr) with timestamps
	uiLogMessage := func(msg string) {
		ts := time.Now().Format("2006-01-02 15:04:05")
		line := fmt.Sprintf("[%s] %s\n", ts, msg)

		logMutex.Lock()
		defer logMutex.Unlock()

		current, _ := logBinding.Get()
		logBinding.Set(current + line)
	}

	var writers []io.Writer
	writers = append(writers, &uiWriter{logFunc: uiLogMessage})
	if logFile != nil {
		writers = append(writers, logFile)
	}
	writers = append(writers, os.Stderr)

	multiWriter := io.MultiWriter(writers...)
	log.SetOutput(multiWriter)
	log.SetFlags(0)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC RECOVERED: %v", r)
			if logFile != nil {
				logFile.Sync()
			}
			nativeDialog.Message("Application crashed. Check log file.").Title("Critical Error").Error()
			os.Exit(1)
		}
	}()

	w := a.NewWindow(fmt.Sprintf("%s %s", AppName, AppVersion))
	w.SetFixedSize(true)

	updateStatus := func(msg string) {
		statusBinding.Set(strings.ReplaceAll(msg, "\n", " "))
	}

	var desk desktop.App
	var updateTray func(string)

	if d, ok := a.(desktop.App); ok {
		desk = d
		iconIdle, _ := fyne.LoadResourceFromPath(IconPathIdle)
		iconConnected, _ := fyne.LoadResourceFromPath(IconPathConnected)
		iconDisconnected, _ := fyne.LoadResourceFromPath(IconPathDisconnected)

		if iconIdle == nil {
			iconIdle = a.Icon()
		}
		if iconConnected == nil {
			iconConnected = a.Icon()
		}
		if iconDisconnected == nil {
			iconDisconnected = a.Icon()
		}

		itemShow := fyne.NewMenuItem("Show App", func() {
			w.Show()
			w.RequestFocus()
		})
		itemConnect := fyne.NewMenuItem("Connect", func() {
			if connectFunc != nil {
				connectFunc()
			}
		})
		itemDisconnect := fyne.NewMenuItem("Disconnect", func() {
			if disconnectFunc != nil {
				disconnectFunc()
			}
		})
		itemQuit := fyne.NewMenuItem("Quit", func() {
			if isTunnelActive {
				dialog.ShowConfirm("Active Connection", "A tunnel is currently active. Quitting will disconnect it. Continue?", func(ok bool) {
					if ok {
						a.Quit()
					}
				}, w)
				w.Show()
			} else {
				a.Quit()
			}
		})

		trayMenu := fyne.NewMenu("Tray",
			itemShow,
			fyne.NewMenuItemSeparator(),
			itemConnect,
			itemDisconnect,
			fyne.NewMenuItemSeparator(),
			itemQuit,
		)

		updateTray = func(state string) {
			switch state {
			case "connected":
				desk.SetSystemTrayIcon(iconConnected)
				itemConnect.Disabled = true
				itemDisconnect.Disabled = false
			case "disconnected":
				desk.SetSystemTrayIcon(iconDisconnected)
				itemConnect.Disabled = false
				itemDisconnect.Disabled = true
			default:
				desk.SetSystemTrayIcon(iconIdle)
				itemConnect.Disabled = false
				itemDisconnect.Disabled = true
			}
			trayMenu.Refresh()
		}

		a.Lifecycle().SetOnStarted(func() {
			desk.SetSystemTrayMenu(trayMenu)
			updateTray("idle")
		})

		w.SetCloseIntercept(func() {
			w.Hide()
		})
	} else {
		updateTray = func(s string) {}
	}

	hostEntry := widget.NewEntry()
	hostEntry.SetPlaceHolder("e.g. 192.168.1.100")
	hostEntry.SetText(cfg.RemoteHost)

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("e.g. jdoe")
	userEntry.SetText(cfg.RemoteUser)

	localPortEntry := widget.NewEntry()
	localPortEntry.SetPlaceHolder("e.g. 33890")
	localPortEntry.SetText(cfg.LocalPort)

	p12Label := widget.NewLabel(filepath.Base(cfg.P12Path))
	if cfg.P12Path == "" {
		p12Label.SetText("Select certificate file...")
	}
	p12Label.Truncation = fyne.TextTruncateEllipsis

	p12PassEntry := widget.NewPasswordEntry()
	p12PassEntry.SetPlaceHolder("Certificate Password")

	status := widget.NewLabelWithData(statusBinding)
	status.Truncation = fyne.TextTruncateEllipsis
	statusBar := container.NewBorder(nil, nil, nil, nil, status)

	validateInputs := func() error {
		if hostEntry.Text == "" {
			return fmt.Errorf("remote host is required")
		}
		if userEntry.Text == "" {
			return fmt.Errorf("ssh username is required")
		}
		if localPortEntry.Text == "" {
			return fmt.Errorf("local port is required")
		}

		port, err := strconv.Atoi(localPortEntry.Text)
		if err != nil {
			return fmt.Errorf("local port must be a number")
		}
		// Port range: ephemeral ports typically above 32768, capped for safety
		if port < 33890 || port > 65000 {
			return fmt.Errorf("local port must be between 33890 and 65000")
		}

		return nil
	}

	validateP12 := func() (*P12Info, error) {
		updateStatus("Status: Validating certificate...")

		if err := validateInputs(); err != nil {
			updateStatus("Status: Error - " + err.Error())
			return nil, err
		}

		if cfg.P12Path == "" {
			err := fmt.Errorf("no P12 file selected")
			updateStatus("Status: Error - " + err.Error())
			return nil, err
		}
		if p12PassEntry.Text == "" {
			err := fmt.Errorf("certificate password required")
			updateStatus("Status: Error - " + err.Error())
			return nil, err
		}

		info, err := ParseP12(cfg.P12Path, p12PassEntry.Text)
		if err != nil {
			updateStatus("Status: Invalid Cert - " + err.Error())
			return nil, err
		}

		log.Print("Certificate Loaded:")
		log.Printf("  Subject: %s", info.Certificate.Subject)
		log.Printf("  Issuer: %s", info.Certificate.Issuer)
		log.Printf("  Serial: %s", info.Certificate.SerialNumber)
		log.Printf("  Valid: %s to %s", info.Certificate.NotBefore, info.Certificate.NotAfter)
		if info.UPN != "" {
			log.Printf("  UPN: %s", info.UPN)
		}

		msg := fmt.Sprintf("Status: Valid Cert (CN: %s", info.CommonName)
		if info.UPN != "" {
			msg += fmt.Sprintf(" | UPN: %s", info.UPN)
		}
		msg += ")"
		updateStatus(msg)

		return info, nil
	}

	showLog := func() {
		if logWindow != nil {
			logWindow.RequestFocus()
			return
		}
		logWindow = a.NewWindow("Activity Log")

		logLabel := widget.NewLabelWithData(logBinding)
		logLabel.Wrapping = fyne.TextWrapWord

		scroll := container.NewVScroll(logLabel)

		saveBtn := widget.NewButtonWithIcon("Save Log", theme.DocumentSaveIcon(), func() {
			val, _ := logBinding.Get()

			filename, err := nativeDialog.File().Save()
			if err != nil {
				if err != nativeDialog.Cancelled {
					dialog.ShowError(err, logWindow)
				}
				return
			}

			if filepath.Ext(filename) == "" {
				filename += ".txt"
			}

			if err := os.WriteFile(filename, []byte(val), 0644); err != nil {
				dialog.ShowError(err, logWindow)
				return
			}
			log.Printf("Log saved to %s", filename)
		})

		clearBtn := widget.NewButtonWithIcon("Clear Log", theme.DeleteIcon(), func() {
			dialog.ShowConfirm("Clear Log", "Are you sure you want to clear the activity log?", func(ok bool) {
				if ok {
					logBinding.Set("")
					log.Print("Log cleared.")
				}
			}, logWindow)
		})

		copyBtn := widget.NewButtonWithIcon("Copy Logs", theme.ContentCopyIcon(), func() {
			val, _ := logBinding.Get()
			a.Clipboard().SetContent(val)
		})

		btnBar := container.NewHBox(saveBtn, clearBtn, layout.NewSpacer(), copyBtn)
		content := container.NewBorder(nil, btnBar, nil, nil, scroll)

		logWindow.SetContent(content)
		logWindow.Resize(fyne.NewSize(800, 600))
		logWindow.CenterOnScreen()
		logWindow.SetOnClosed(func() {
			logWindow = nil
		})
		logWindow.Show()
	}

	exportKey := func(isPrivate bool) {
		info, err := validateP12()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		var data []byte
		var defaultName string

		if isPrivate {
			data, err = ExportPrivateKey(info.PrivateKey)
			defaultName = "id_rsa"
		} else {
			data, err = ExportPublicKey(info.PrivateKey)
			defaultName = "id_rsa.pub"
		}

		if err != nil {
			dialog.ShowError(fmt.Errorf("export failed: %v", err), w)
			return
		}

		filename, err := nativeDialog.File().SetStartFile(defaultName).Save()
		if err != nil {
			if err != nativeDialog.Cancelled {
				dialog.ShowError(err, w)
			}
			return
		}

		if err := os.WriteFile(filename, data, 0600); err != nil {
			dialog.ShowError(err, w)
			return
		}
		log.Printf("Key exported to %s", filename)
		dialog.ShowInformation("Export Success", "Key exported successfully.", w)
	}

	openUrl := func(raw string) {
		if u, err := url.Parse(raw); err == nil {
			_ = fyne.CurrentApp().OpenURL(u)
		}
	}

	quitApp := func() {
		if isTunnelActive {
			dialog.ShowConfirm("Active Connection", "A tunnel is currently active. Quitting will disconnect it. Continue?", func(ok bool) {
				if ok {
					a.Quit()
				}
			}, w)
		} else {
			a.Quit()
		}
	}

	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("View Activity Log", showLog),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Export Private Key", func() { exportKey(true) }),
		fyne.NewMenuItem("Export Public Key", func() { exportKey(false) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", quitApp),
	)
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Documentation", func() { openUrl(DocURL) }),
		fyne.NewMenuItem("Report Issue", func() { openUrl(IssueURL) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("About", func() {
			dialog.ShowInformation(AppName+" "+AppVersion, AboutText, w)
		}),
	)
	w.SetMainMenu(fyne.NewMainMenu(fileMenu, helpMenu))

	var headerIcon *canvas.Image
	if appIcon != nil {
		headerIcon = canvas.NewImageFromResource(appIcon)
		headerIcon.FillMode = canvas.ImageFillContain
		headerIcon.SetMinSize(fyne.NewSize(49, 48))
	}

	title := canvas.NewText(AppName, theme.Color(theme.ColorNameForeground))
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}

	var headerTop *fyne.Container
	if headerIcon != nil {
		headerTop = container.NewHBox(headerIcon, title)
	} else {
		headerTop = container.NewHBox(title)
	}

	desc := widget.NewLabel("Connect to a remote host via RDP tunneled through SSH.")
	desc.Wrapping = fyne.TextWrapWord
	header := container.NewVBox(headerTop, widget.NewSeparator(), desc)

	browse := widget.NewButton("...", func() {
		filename, err := nativeDialog.File().Filter("PKCS#12 Certificate", "p12", "pfx").Load()
		if err != nil {
			if err != nativeDialog.Cancelled {
				dialog.ShowError(err, w)
			}
			return
		}
		cfg.P12Path = filename
		p12Label.SetText(filepath.Base(filename))
		updateStatus("Status: Certificate selected. Enter password and click Test.")
		log.Printf("Selected certificate file: %s", filename)
	})

	p12Row := container.NewBorder(nil, nil, nil, browse, p12Label)

	grid := container.NewGridWithColumns(2)

	add := func(label string, input fyne.CanvasObject) {
		lbl := widget.NewLabel(label)
		lbl.Alignment = fyne.TextAlignLeading
		grid.Add(lbl)
		grid.Add(input)
	}

	add("Remote Host", hostEntry)
	add("SSH Username", userEntry)
	add("Local Port", localPortEntry)
	add("Certificate File", p12Row)
	add("Certificate Password", p12PassEntry)

	setInputsEnabled := func(enabled bool) {
		if enabled {
			hostEntry.Enable()
			userEntry.Enable()
			localPortEntry.Enable()
			p12PassEntry.Enable()
			browse.Enable()
		} else {
			hostEntry.Disable()
			userEntry.Disable()
			localPortEntry.Disable()
			p12PassEntry.Disable()
			browse.Disable()
		}
	}

	var testBtn *widget.Button
	var connectBtn *widget.Button

	testBtn = widget.NewButton("Test Connection", func() {
		testBtn.Disable()
		log.Print("--- Starting Connection Test ---")

		info, err := validateP12()
		if err != nil {
			log.Printf("Validation failed: %v", err)
			testBtn.Enable()
			return
		}

		updateStatus("Status: Testing SSH connection...")
		go func() {
			res, err := TestConnection(hostEntry.Text, userEntry.Text, info)

			fyne.Do(func() {
				if err != nil {
					log.Printf("Test failed: %v", err)
					updateStatus("Status: Test Failed - " + err.Error())
				} else {
					log.Printf("Test success: %s", res)
					updateStatus("Status: " + res)
				}
				testBtn.Enable()
			})
		}()
	})
	testBtn.Importance = widget.HighImportance

	var cancelTunnel context.CancelFunc

	disconnectFunc = func() {
		if cancelTunnel != nil {
			log.Print("Disconnect requested by user.")
			cancelTunnel()
		}
	}

	connectFunc = func() {
		if isTunnelActive {
			dialog.ShowConfirm("Disconnect", "Are you sure you want to disconnect?", func(ok bool) {
				if ok {
					disconnectFunc()
				}
			}, w)
			return
		}

		connectBtn.Disable()
		testBtn.Disable()
		setInputsEnabled(false)
		log.Print("--- Initiating Connection Sequence ---")

		cfg.RemoteHost = hostEntry.Text
		cfg.RemoteUser = userEntry.Text
		cfg.LocalPort = localPortEntry.Text
		_ = SaveConfig(cfg)

		info, err := validateP12()
		if err != nil {
			log.Printf("Validation failed: %v", err)
			w.Show()
			w.RequestFocus()
			dialog.ShowError(fmt.Errorf("authentication failed: %v", err), w)
			connectBtn.Enable()
			testBtn.Enable()
			setInputsEnabled(true)
			return
		}

		updateStatus("Status: Connecting...")

		var ctx context.Context
		ctx, cancelTunnel = context.WithCancel(context.Background())

		onReady := func() {
			fyne.Do(func() {
				isTunnelActive = true
				updateStatus(fmt.Sprintf(StatusTextConnected, hostEntry.Text))
				log.Print("Tunnel Ready. RDP Client Launched.")

				updateTray("connected")

				connectBtn.SetText("Disconnect")
				connectBtn.Importance = widget.DangerImportance
				connectBtn.Refresh()
				connectBtn.Enable()
			})
		}

		go func() {
			err := StartTunnel(ctx, hostEntry.Text, userEntry.Text, localPortEntry.Text, info, func(s string) { log.Print(s) }, onReady)

			fyne.Do(func() {
				isTunnelActive = false
				cancelTunnel = nil

				// Ignore expected exit codes from user-cancelled RDP sessions
				isCancel := err == context.Canceled || (err != nil && strings.Contains(err.Error(), "exit status 1") && ctx.Err() == context.Canceled)

				if err != nil && !isCancel {
					log.Printf("Tunnel error: %v", err)
					updateStatus("Status: Connection Error")
					dialog.ShowError(err, w)
				} else {
					log.Print("Session ended normally.")
					updateStatus(StatusTextDisconnected)
				}

				updateTray("disconnected")

				connectBtn.SetText("Connect & Launch")
				connectBtn.Importance = widget.SuccessImportance
				connectBtn.Refresh()
				connectBtn.Enable()
				testBtn.Enable()
				setInputsEnabled(true)
			})
		}()
	}

	connectBtn = widget.NewButton("Connect & Launch", connectFunc)
	connectBtn.Importance = widget.SuccessImportance

	btnRow := container.NewHBox(
		layout.NewSpacer(),
		testBtn,
		widget.NewLabel(""),
		widget.NewLabel(""),
		connectBtn,
		layout.NewSpacer(),
	)

	paddedBtnRow := container.NewVBox(
		widget.NewLabel(""),
		btnRow,
		widget.NewLabel(""),
	)

	centerContent := container.NewVBox(
		grid,
		paddedBtnRow,
	)

	bottomContent := container.NewVBox(
		widget.NewSeparator(),
		statusBar,
	)

	content := container.NewBorder(
		header,
		bottomContent,
		nil,
		nil,
		centerContent,
	)

	w.SetContent(container.NewPadded(content))

	w.SetCloseIntercept(func() {
		if cfg.MinimizeToTrayWarning {
			check := widget.NewCheck("Don't show this again", nil)
			content := container.NewVBox(
				widget.NewLabel("The application will continue running in the system tray."),
				check,
			)

			d := dialog.NewCustomConfirm("Minimize to Tray", "OK", "Cancel", content, func(ok bool) {
				if ok {
					if check.Checked {
						cfg.MinimizeToTrayWarning = false
						_ = SaveConfig(cfg)
					}
					w.Hide()
				}
			}, w)
			d.Show()
		} else {
			w.Hide()
		}
	})

	w.Resize(fyne.NewSize(480, 420))
	w.ShowAndRun()
}
