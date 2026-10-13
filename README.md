# Elarsi | User Guide

**Elarsi 1.2 • Windows 10/11, 64-bit • Portable • Offline**

Elarsi is an LRC lyric synchronization tool. Load a song and lyrics, listen, and press a key to record timestamps. It supports line-by-line and word-by-word synchronization, importing existing LRC files for correction, and a read-only playback preview.

## 1. Install and launch

1. Download the Elarsi ZIP and choose **Extract All** in Windows Explorer.
2. Open the extracted `Elarsi` folder.
3. Double-click `Elarsi.exe`. The interface opens in your default browser.
4. Keep the EXE running while using the app. Use **Quit Elarsi** when finished.

There is no installer, account, or internet connection required. The EXE starts a local-only server at `127.0.0.1` to display its interface in your browser. Audio and lyrics remain on your computer. Exports are downloaded through your browser, normally into Downloads unless your browser asks where to save.

**Security note:** Windows or your browser may warn about newly built or unsigned EXEs, even inside a ZIP. ZIP distribution does not guarantee that SmartScreen warnings disappear. Only run software you trust; do not disable security protections to use Elarsi. This build is currently unsigned.

## 2. First synchronization: line-by-line

1. Use **AUDIO > Browse** to open your song.
2. Use **LYRICS > Browse** to open a `.txt` file containing one lyric per line, or a `.lrc` file.
3. Choose **Line Edit** in Synchronization.
4. The **CURRENT** area starts at `START`; **NEXT** contains the first lyric to stamp.
5. Start the song using **Play**, then press **Space** as the first word of **NEXT** begins.
6. Elarsi writes the audio playhead time to **NEXT**, moves it into **CURRENT**, and queues the following line.
7. Continue until **NEXT** displays `END`.
8. Switch to **Playback Preview** to check timing, then choose **Export LRC**.

Example input TXT:

```text
Three hundred hours
Not a single regret
Until I saw something
```

Example output LRC (times are examples):

```lrc
[ti:Example Song]
[ar:Example Artist]
[offset:0]

[00:04.26]Three hundred hours
[00:07.48]Not a single regret
[00:11.26]Until I saw something
```

Blank input lines are separators, not lyrics requiring stamps.

## 3. Word Edit (enhanced LRC)

Switch to **Word Edit** to queue and timestamp individual words. **NEXT** is the upcoming word; each press records a `<mm:ss.xx>` word timestamp and advances one word. Stamping the first word can initialize that line's `[mm:ss.xx]` timestamp. The line and word timestamps are stored independently thereafter, which makes separate mass corrections possible.

Example enhanced LRC:

```lrc
[00:04.26]<00:04.26>Three <00:04.60>hundred <00:05.01>hours
```

Word segmentation uses the words from your lyric text; punctuation stays with its word. You can return to Line Edit without discarding stored word timing.

## 4. Playback Preview

**Playback Preview** is read-only. Play, pause, or seek audio to review saved line timing and word highlighting. In this mode, displayed lyrics follow the audio, unlike Edit Mode. Editing controls are disabled and visually muted.

## 5. The queue and playhead are independent

This is a core rule in Elarsi:

- Moving the **audio playhead**, playing, pausing, changing speed, or seeking does **not** change which lyric or word is queued.
- Selecting, scrolling, or navigating to another **lyric/word** does **not** move the audio playhead.
- **Stamping** reads the current audio time, writes it to **NEXT**, and advances unless Single Correction is enabled.

To fix a passage in an existing LRC: import it, select the desired lyric in the list (or navigate the queue), position audio manually near that passage, and stamp again. Untouched timestamps remain in place. When **Single Correction** is enabled, stamping keeps the same lyric/word queued so you can refine it repeatedly. This is a toggle in Settings; hover for an explanation.

## 6. Controls and shortcuts

| Control | Action |
|---|---|
| Space (default) | Stamp **NEXT** lyric or word in Edit Mode |
| Enter | Play / pause |
| Left / Right arrow | Seek backward / forward 2 seconds |
| Up / Down arrow | Queue previous / next lyric or word in Edit Mode |
| Ctrl+Z | Undo last document change |
| Ctrl+Y | Redo |
| Ctrl+S | Save `.elarsi` project |
| `↶ 5s` / `5s ↷` | Rewind / advance audio 5 seconds |
| Mouse wheel over the queued lyric | Choose previous / next queued lyric or word |
| Click the waveform or drag the seek slider | Move the playhead without changing the queue |
| Click a lyric in the Lyrics list | Queue that lyric without seeking |

Change the stamp key with **Settings > Stamp key**. Click it, press the new key, or press Escape to cancel. Avoid choosing a key you normally use for playback controls.

## 7. Song Information and Settings (left column)

- **Title:** exported to the `[ti:]` LRC header field.
- **Author / Artist:** exported to `[ar:]`.
- **LRC metadata offset (ms):** exported as `[offset:]`; does **not** rewrite stored timestamps.
- **Stamp key:** remappable shortcut, Space by default.
- **Export timestamp precision:** hundredths (`mm:ss.xx`) or milliseconds (`mm:ss.xxx`). Internally, timestamps are stored in integer milliseconds.
- **Single Correction:** when enabled, stamping does not advance to the next queued item.
- **Output Filenames:** choose **Audio filename** or **Lyrics / LRC filename** to set the basename used for both LRC exports and `.elarsi` project saves. The on-screen preview shows the result. If the preferred source is missing, Elarsi falls back to the available imported filename and then the song title.

The chosen output filename is independent of the song's **Title** metadata; renaming one does not have to rename the other.

## 8. Adjust Timestamps (right column)

These are **permanent, document-wide edits**, distinct from `[offset:]` metadata.

- **Line offset (ms):** shifts every existing `[ ]` line timestamp, leaving all `< >` word timestamps unchanged.
- **Word offset (ms):** shifts every existing `< >` word timestamp, leaving all `[ ]` line timestamps unchanged.

For example, `+150` line / `-75` word moves each line time forward 150 ms and each word time backward 75 ms. The panel shows the number of affected timestamps and a **BEFORE / AFTER** preview. Check it before choosing **Apply Adjustments**. The operation is one Undo step, even when many timestamps change. Adjustments that would make a timestamp negative are blocked.

## 9. Document Status and warnings (right column)

Document Status reports incomplete timing, timestamp-order warnings, and other potential issues. Expand an issue group to see affected lines and press **Go to** to select the lyric and scroll to it in the editor. **Go to** never moves audio. Warning icons appear beside affected lyrics.

A timestamp-order warning means a line is earlier than the preceding line **in document order**. It does not mean repeated words or a repeated chorus are invalid. Overlapping vocals may make these warnings intentional. Elarsi warns before exporting files with validation issues, but does not automatically change your lyrics or times.

## 10. Files, projects, and source locations

**Import:** TXT, standard LRC, and enhanced LRC. Audio playback depends on the formats your installed browser supports; common formats include MP3 and WAV, while other formats may vary.

**Export LRC:** produces `.lrc` text with `[ti:]`, `[ar:]`, `[offset:]` headers and available line/word timestamps. Untimed lines are omitted on LRC export. Review export warnings to avoid accidentally excluding unfinished lines.

**Save Project:** produces an `.elarsi` JSON project containing lyric text, timestamps, metadata, and editing state. **Audio is not embedded**; reattach the original audio when reopening the project. Older `.lyricsync` project files can also be opened.

**Source folder buttons:** the folder icons beside Audio and Lyrics and the **Project Source** control reveal the source file in Windows Explorer, when the source path is known. Files chosen through Elarsi's native Windows Browse/Open Project dialogs can be located this way. A file dragged into the browser does not usually expose its original disk path because of browser security restrictions. The source-folder buttons point to **imported originals**, not to exported files.

**Autosave recovery:** Elarsi can offer to recover lyrics from an earlier editing session. Explicitly save projects for reliable backups and portability.

## 11. Troubleshooting

**EXE icon looks wrong:** Windows Explorer may be showing a cached icon. Refresh Explorer or move the app into its permanent folder.

**Browser says it cannot connect:** start `Elarsi.exe` again; the UI depends on its local server running.

**No audio plays:** re-open the audio through Browse, and try an MP3 or WAV if the format isn't supported by your browser.

**Stamp key does nothing:** ensure you are in Line Edit or Word Edit, audio is loaded, and focus is not inside a text-entry field. Check your mapped key in Settings.

**Folder button cannot locate a source:** reimport via Browse rather than browser drag-and-drop.

**Files are saved to an unexpected place:** Elarsi's export actions use the browser's download destination/settings. The imported source folder is not automatically the export folder.

**A timestamp-order warning appears:** expand it in Document Status to compare the affected line with the preceding line's timestamp; corrections do not require retiming the entire song.

## 12. What's inside this package

```text
Elarsi/
  Elarsi.exe     Windows executable with embedded app icon
  README.md      Setup instructions and user documentation (this file)
```

This is the **user distribution**, not the source package. Source code can be distributed separately.

## Building the Windows executable with its icon

The GitHub source includes `build.bat`, `embed_icon.py`, and `elarsi.ico` so the Windows Explorer icon is embedded automatically whenever you rebuild the EXE. This is separate from the in-app logo and browser favicon in `web/elarsi-icon.png`.

1. Install [Go](https://go.dev/dl/) and Python 3 on Windows. Ensure `go` and the Python `py` launcher are on your PATH.
2. Open the project root and double-click `build.bat`, or run it in Command Prompt so you can see any error output.
3. Find `dist/Elarsi.exe` and the same `README.md` in `dist/`. The EXE contains the multi-size Windows icon resource.
4. For distribution, ZIP the contents of `dist/`. Do not distribute `embed_icon.py`, `elarsi.ico`, or development files to end users.

The Python icon embedding script uses **only the Python standard library**, with no pip dependencies. Do not rerun it on an already patched EXE; rerun `build.bat` to build afresh. If you later digitally sign the executable, sign **after** the icon is embedded.

Both the source archive and the user release ZIP contain this same README/manual.
