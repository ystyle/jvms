# SDKMAN provider

On Linux and macOS, JVMS uses SDKMAN for Java catalogs, installation, removal,
and switching. Available JDK distributions and versions come from SDKMAN for
the host platform.

## Setup

Requires `bash`, `curl`, `zip`, and `unzip` on `PATH`. Run as your regular user
without `sudo`. JVMS uses `SDKMAN_DIR` if set, otherwise `~/.sdkman`.

Download and extract the matching ZIP from the
[releases page](https://github.com/renniemaharaj/jvms/releases). Archive names
start with `jvms_<version>_` and end with:

| Platform | Archive suffixes |
| --- | --- |
| Linux | `linux_amd64.zip`, `linux_386.zip`, `linux_armv6.zip`, `linux_arm64.zip` |
| macOS | `darwin_amd64.zip` (Intel), `darwin_arm64.zip` (Apple Silicon) |

The ARMv6 binary also supports ARMv7 and ARMv8 in 32 bit mode. Place `jvms` in
a directory on your `PATH`, such as `~/.local/bin`, and make it executable with
`chmod +x ~/.local/bin/jvms`.

With SDKMAN already installed, run `jvms init`. Otherwise follow
[SDKMAN's installation instructions](https://sdkman.io/install/) or enable
JVMS to run its installation script (disabled by default):

```sh
jvms config --allow-auto-install-sdkman=true
jvms init
```

Open a new terminal after SDKMAN installation so its shell setup takes effect.

## Usage

Run `jvms rls` to see SDKMAN Java identifiers and `jvms list` to see installed
versions. Use those identifiers with `jvms install`, `jvms switch`, `jvms use`,
and `jvms remove`, or run `jvms tui` to open the picker. Switching sets SDKMAN's
default Java version.

Installing a JDK through SDKMAN may update both your current Java version and
the default. JVMS shows a warning in the CLI and install confirmations and
records SDKMAN's actual current selection after installation. To hide the warning:

```sh
jvms config --dismiss-sdk-install-warning=true
```
