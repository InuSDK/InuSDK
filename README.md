# InuSDK

InuSDK is a SDK Manager oriented for multiple platforms, not limiting itself to Linux.

## Installation

### Shell/bash

To install the SDK and test it run the `install.sh` script. After installation, run `inusdk`

Right now, most of the commands are under development.

### Powershell/Windows terminal

to install in Windows, run the powershell script `install.ps1`, as mentioned in [Shell/Bash](https://github.com/InuzDev/InuSDK#shellbash), run `inusdk`

## Developer/Quick test

If you desire to give a quick test of the program, you can run in your local machine `install-dev.ps1`

> Under development; `install-dev.sh` for Linux/Darwin users.

## Commands

- `use [version]`: This command specify the SDK and the version you want to use in the moment.
- `install [sdkname] [version] <flag>`: This will install an SDK from any version available, if the user doesn't specify; it will install the latest release of the SDK. You can use `--pr` to get a pre-release version.
- `list [sdkname]`: This will list every single SDK installed in your machine. You can specify an sdk name to check every version of that specific SDK.
- `reset`: This will reset to default settings the SDK manager.
- `uninstall [sdkname] [version] <flag>`: This command is used to uninstall an SDK, using the flag `--all` it will delete every single SDK and with `--force` it will delete even the active SDK.
- `remove`: This command uninstall completely the SDK manager, be careful with this one. It will always prompt for confirmation, it doesn't have flags to avoid accidents.
