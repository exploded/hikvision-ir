# hikvision-ir

Controls the IR LED illuminators on HikVision IP cameras via the ISAPI HTTP API.

Tested on DS-2CD2343G0-I and DS-2CD2335-I. Should work on any HikVision camera with the Hardware IR light switch in the web UI at Configuration → System → Maintenance → System Service → Hardware.

## Usage

```
hikvision-ir --host <IP> --user <user> --pass <pass> --action on|off|status
```

```sh
# Check current state
hikvision-ir --host 192.168.1.4 --pass yourpassword --action status

# Turn IR off
hikvision-ir --host 192.168.1.4 --pass yourpassword --action off

# Turn IR on
hikvision-ir --host 192.168.1.4 --pass yourpassword --action on
```

`--user` defaults to `admin`.

## Build

```sh
go build -o hikvision-ir .
```

## How it works

Uses `PUT /ISAPI/System/Hardware` with an `IrLightSwitch` XML payload and HTTP Digest authentication. This is the same endpoint the camera web UI uses for the Hardware IR light switch toggle.
