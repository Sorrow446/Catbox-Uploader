# Catbox-Uploader
CLI catbox.moe uploader written in Go.
![](https://i.imgur.com/zCnTc7c.png)
[Windows, Linux, macOS and Android binaries](https://github.com/Sorrow446/Catbox-Uploader/releases)

# Usage
Upload a file and print URL to console:   
`catbox_ul_x64.exe G:\1.bin`

Upload two files, print URLs to console and append to out.txt:   
`catbox_ul_x64.exe G:\1.bin G:\2.bin -o out.txt`

You can use the -w flag to wipe the output text file before writing to it instead of appending.
```
Usage: catbox_ul_x64.exe [--outpath OUTPATH] [--wipe] PATHS [PATHS ...]

Positional arguments:
  PATHS

Options:
  --outpath OUTPATH, -o OUTPATH
  --wipe, -w
  --help, -h             display this help and exit
  ```
