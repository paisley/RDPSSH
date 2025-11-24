# RDPSSH

> **R**emote **D**esktop **P**rotocol over **SSH** - A secure tunnel manager with certificate authentication

RDPSSH is a Windows desktop application that simplifies connecting to remote Linux hosts via RDP through an SSH tunnel, using PKCS#12 certificate authentication.

## Features

- 🔒 **Certificate-Based SSH Authentication** - Uses P12/PFX certificates instead of passwords
- 🚇 **Automatic SSH Tunneling** - Creates local port forwarding to remote RDP (port 3389)
- 🖥️ **Integrated RDP Launch** - Automatically launches `mstsc.exe` with tunnel configuration
- 💾 **Configuration Persistence** - Saves connection settings for quick reconnection
- 📊 **Activity Logging** - Comprehensive logging with save/export functionality
- 🔑 **Key Export** - Export private/public keys from certificates in OpenSSH format
- 🎨 **System Tray Integration** - Minimize to tray, connect/disconnect from tray menu
- ✅ **Connection Testing** - Test SSH connectivity before establishing full tunnel

## Use Cases

- **Secure Remote Administration** - Access RDP through SSH tunnels for enhanced security
- **Certificate-Based Workflows** - Leverage existing PKI infrastructure for SSH authentication

## Prerequisites

- **Windows** (tested on Windows 10/11)
- **Go 1.21+** (for building from source)
- Valid PKCS#12 (.p12 or .pfx) certificate with private key
- SSH access to remote host with certificate authentication configured
- RDP enabled on remote host

## Installation

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/paisley/rdpssh/releases).

### Build from Source

```powershell
# Clone the repository
git clone https://github.com/paisley/rdpssh.git
cd rdpssh

# Development build
make dev

# Production build
make prod

# Or use go directly
go build -ldflags "-X main.AppVersion=v1.0.0" -o rdpssh.exe
```

## Usage

### First Time Setup

1. **Launch RDPSSH**

2. **Configure Connection**
   - **Remote Host**: IP address or hostname of SSH server (e.g., `192.168.1.100`)
   - **SSH Username**: Your SSH username
   - **Local Port**: Local port for tunnel (default: `33890`, range: 33890-65000)
   - **Certificate File**: Browse and select your `.p12` or `.pfx` certificate
   - **Certificate Password**: Enter the password for your certificate

3. **Test Connection**
   - Click "Test Connection" to verify SSH connectivity
   - Check the activity log for detailed connection information

4. **Connect & Launch**
   - Click "Connect & Launch" to establish tunnel and open RDP client
   - The application will remain in the system tray while connected
   - Close the RDP window to disconnect

### System Tray

- **Show App** - Restore main window
- **Connect** - Quick connect with saved settings
- **Disconnect** - Close active tunnel
- **Quit** - Exit application (warns if tunnel active)

### File Menu

- **File Menu**
  - View Activity Log
  - Export Private Key (OpenSSH format)
  - Export Public Key (OpenSSH authorized_keys format)
  - Quit

## Configuration

Settings are automatically saved to:
```
%APPDATA%\rdpssh\config.json
```
## Building

### Development Build
```powershell
make dev
```

### Production Build
```powershell
make prod VERSION=v1.2.3
```
- Optimized binary
- Stripped debug info (`-trimpath`)
- Custom version number

### CI/CD Build
```powershell
# Set version from environment
$env:VERSION="v1.0.0"
$env:COMMIT="abc1234"
make ci
```

### Custom Build Flags
```powershell
go build -ldflags "-X 'main.AppName=RDPSSH' -X 'main.AppVersion=v1.0.0'" -o rdpssh.exe
```

## Project Structure

```
rdpssh/
├── main.go           # Application entry point and UI
├── config.go         # Configuration management
├── p12.go            # PKCS#12 certificate parsing
├── ssh_client.go     # SSH tunnel and RDP launch logic
├── theme.go          # Custom Fyne theme (the default green was horrible)
├── version.go        # Version and constants
├── icons/            # Application icons
│   ├── app.png
│   ├── idle.png      # systray
│   ├── connected.png      |   
│   └── disconnected.png   v
├── Makefile          # Build automation
├── go.mod            # Go module definition
└── README.md         # This file
```

## Troubleshooting

### Certificate Issues

**Problem**: "failed to decode p12" error

**Solutions**:
- Verify certificate password is correct
- Ensure certificate contains both private key and public certificate
- Try re-exporting certificate from your certificate store with "Export Private Key" enabled

### SSH Connection Failures

**Problem**: "connection failed" or authentication errors

**Solutions**:
- Verify remote host allows SSH key authentication
- Check if your certificate's public key is in `~/.ssh/authorized_keys` on remote host
- Export public key via File → Export Public Key and add to remote host
- Verify SSH is running on remote host (default port 22)
- Check firewall rules allow SSH connections

### RDP Launch Issues

**Problem**: RDP client doesn't open or can't connect

**Solutions**:
- Verify local port isn't already in use
- Check RDP is enabled on remote Windows host
- Try different local port (33890-65000)

### Log Files

Opening File > Activity Log will show you a running log of the current session.

Application logs are saved to:
```
%APPDATA%\rdpssh\app.log
```

## Security Considerations

⚠️ **Important Security Notes**:

- RDPSSH currently uses `ssh.InsecureIgnoreHostKey()` which **does not verify SSH host keys**
- This makes the connection vulnerable to man-in-the-middle attacks
- For production use, implement proper SSH host key verification
- 
## License

This project is open source. See LICENSE file for details.

## Attribution
- Red X Circle Icon: [Delete icons created by Pixel perfect - Flaticon](https://www.flaticon.com/free-icons/delete)
- Black X Circle Icon: [Delete icons created by Pixel perfect - Flaticon](https://www.flaticon.com/free-icons/delete)
- Green Check Circle Icon: [Success icons created by hqrloveq - Flaticon](https://www.flaticon.com/free-icons/success)
- App Icon: [No connection icons created by andinur - Flaticon](https://www.flaticon.com/free-icons/no-connection)

---

