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

> Make sure the terminal window's width is large enough, otherwise the output will appear all garbled.

## Usage

```shell
roger [kits|instruments] [PackName ...]
```
The first time you run `roger`, it creates a folder on your Desktop named `roger` containing `Kits/`, `Instruments/`, and `Output/` folders, and configuration files. If `Kits/` is empty, `roger` generates example packs so you can see how things work.

With no arguments, all packs in the mode's input directory are processed. Pass one or more pack names to process only those.

## Folder structure

Each top-level folder inside `Kits/` is a pack. You can have as many packs as you like, and they are each converted into a separate MPC Expansion:

```
Kits/
  PackOne/
    ...
  PackTwo/
    ...
```

Within each pack, kits can be organized in two ways:

**Flat pack** — all kits directly under the pack folder:
```
Kits/
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
Kits/
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

`roger` reads filenames to detect drum types (kick, snare, hat, clap, etc.) and assigns them to the appropriate pads automatically.

Each kit should contain up to 16 WAV files. If there are more than 16, the extra files that were not assigned to a pad are ignored.

### Cover image

Place an image file (PNG, JPG, or TIFF) in the top-level pack folder to use it as the expansion cover art:

```
Kits/
  MyPack/
    SomeImage.png   ← used as Expansion.jpg
    Kit 1/
      ...
```

## Configuration

`roger` creates a `config.yaml` in the `roger` folder on your Desktop on first run with sensible defaults. It controls how samples are detected and assigned to pads. You can edit this file to customize the behaviour.

### How sample assignment works

1. `roger` inspects each sample's filename and classifies it as a drum type (e.g. `Kick`, `Snare`) by looking for keywords defined in `drum_types`.
2. It then assigns each sample to one of the 16 pads according to `pad_layout`, which maps each pad to one or more accepted drum types in priority order.
3. Any samples that didn't match their target pad type are placed in the remaining empty pads.

### `drum_types`

Defines the drum types `roger` recognises and the filename keywords used to detect each one:

```yaml
drum_types:
  - name: Kick
    tokens: [kick, kik, bass drum, bd]
  - name: Snare
    tokens: [snare, snr, sd]
  - name: ClosedHiHat
    tokens: [closed hat, closed hi-hat, chh, ch]
  ...
```

If a sample's filename contains any of the tokens for a type (case-insensitive), it is classified as that type. You can add new types, rename existing ones, or extend the token lists to match the naming conventions of your sample packs.

### `pad_layout`

Defines the 16 pads in order, each specifying which drum type(s) it accepts. The first type listed is preferred; the rest are fallbacks used if no sample of the preferred type is available:

```yaml
pad_layout:
  - [Kick]           # pad 1: kick only
  - [Snare, Clap]    # pad 2: snare preferred, clap as fallback
  - [ClosedHiHat]    # pad 3: closed hat only
  ...
```

The type names used here must match names defined in `drum_types`.

### Custom program template

`roger` ships with a default MPC program template. To use your own, export a drum program from the MPC as an `.xpm` file, rename it to `kit.xpm`, and place it in the `roger` folder on your Desktop.

`roger` will use it as the base for all generated programs, preserving your pad colors and other settings.

## Loading into the MPC

Copy the expansion folders from `Output/` into your MPC's `Expansions` folder on its internal storage or an attached drive. The MPC will recognize it as an Expansion, and the programs will appear in its browser.
