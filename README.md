# roger

`roger` batch-converts drum sample packs into MPC Expansions. It's intended for use with sample packs that already provide samples split into _kits_, where one folder contains all samples for a single kit. Examples of such packs are the Samples From Mars packs.

- Creates 16-pad drum programs with samples automatically assigned to pads by type (kick, snare, hihat, etc.)
- Creates _Multikit_ programs with one kit per bank, for browsing multiple kits without switching programs
- Generates audio previews for each program, so you can audition kits in the MPC browser
- Use your own drum program as a template to keep pad colors and other settings consistent across all programs
- Configurable pad layout and drum type detection

<video src="https://github.com/user-attachments/assets/294e9a2f-f1e7-45c1-bd61-4143ef09cd09" controls></video>

## Installation

[Download the latest release](https://github.com/psrpinto/roger/releases/latest) for your platform and extract the zip. `roger` is a terminal-based application, so you launch it from the terminal.

If you're not used to working with terminals, you can also just double-click the executable file and it will launch in a terminal window.

> On MacOS you will likely need to go into Settings -> Privacy & Security, scroll down to Security and click "Open Anyway".

## Usage

```shell
roger [PackName ...]
```
The first time you run `roger`, it creates a workspace on your Desktop at `~/Desktop/roger/` with `Input/` and `Output/` folders. If `Input/` is empty, `roger` generates example packs so you can see how things work.

With no arguments, all packs in `Input/` are processed. Pass one or more pack names to process only those.

## Folder structure

Organize your samples inside `Input/` like this:

**Flat pack** — one group of kits:
```
Input/
  MyPack/
    Kit 1/
      Kick.wav
      Snare.wav
      ...
    Kit 2/
      Kick.wav
      Snare.wav
      ...
```

**Grouped pack** — kits organized into named groups:
```
Input/
  MyPack/
    Group A/
      Kit 1/
        Kick.wav
        ...
    Group B/
      Kit 1/
        Kick.wav
        ...
```

Each kit should contain up to 16 WAV files. `roger` reads filenames to detect drum types (kick, snare, hat, clap, etc.) and assigns them to the appropriate pads automatically.

### Cover image

Place an image file (PNG, JPG, or TIFF) in the top-level pack folder to use it as the expansion cover art:

```
Input/
  MyPack/
    cover.png   ← used as Expansion.jpg
    Kit 1/
      ...
```

## Configuration

`roger` creates a `config.yaml` in `~/Desktop/roger/` on first run. You can edit it to customize:

- **`drum_types`** — the list of drum types and the filename keywords used to detect each one
- **`pad_layout`** — which drum type(s) each of the 16 pads accepts, in priority order

Example pad layout entry:
```yaml
pad_layout:
  - [Kick]           # pad 1: kick only
  - [Snare, Clap]    # pad 2: snare preferred, clap as fallback
  - [ClosedHiHat]    # pad 3: closed hat only
  ...
```

### Custom program template

`roger` ships with a default MPC program template. To use your own, export a drum program from MPC as an `.xpm` file and place it at `~/Desktop/roger/template.xpm`. `roger` will use it as the base for all generated programs, preserving your pad colors and other settings.

## Loading into the MPC

Copy the pack folder from `Output/` into your MPC's `Expansions` directory on its internal storage or an attached drive. The MPC will recognize it as an Expansion and the programs will appear in its browser.
