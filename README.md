# mtorrent
![loc](https://sloc.xyz/github/nektro/mtorrent)
[![discord](https://img.shields.io/discord/551971034593755159.svg?logo=discord)](https://discord.gg/P6Y4zQC)
[![goreportcard](https://goreportcard.com/badge/github.com/nektro/mtorrent)](https://goreportcard.com/report/github.com/nektro/mtorrent)
[![downloads](https://img.shields.io/github/downloads/nektro/mtorrent/total.svg)](https://github.com/nektro/mtorrent/releases)

A totally configurable terminal torrent client.

## Download
```
go get github.com/nektro/mtorrent
```

## Flags
```
Usage of ./mtorrent:
  -c, --concurrency int        Maximum number of torrents to actively download at a time. (default 10)
      --disable-dht            Setting this will disable DHT.
      --do-download            Setting this flag to false will make all torrents idle in client. (default true)
  -d, --done-dir string        Optional directory to move completed torrents to. (default -w)
      --drop-after int         Minutes to drop torrents with no progress after. (disable -1) (default 35)
      --drop-after-force int   Minutes to drop torrents after regardless of progress. (disable -1) (default -1)
  -i, --include-btih-in-dn     If true, folder name will be 'btih dn' instead of 'dn'.
  -m, --magnet stringArray     Magnet Link to download. (Can be passed multiple times.)
      --magnet-file string     Path to a text file with magnet links on each line.
      --mbpp-bar-gradient      Enabling this will make the bar gradient from red/yellow/green.
      --pack-tar               Enabling this will pack torrent folder into a .tar so that it only takes up a single file.
      --peer-id string         Bittorrent peer_id. (default "-lt0D20-")
      --peers-log string       An optional path to log file that will list all peers per torrent.
  -s, --seed-for int           When positive, minutes to seed torrents for. (-1: forever) (0: only leech)
  -t, --torrent stringArray    Path of the torrent file you wish to download. (Can be passed multiple times.)
      --torrent-dir string     Path to directory of torrent files. Will only pick up .torrents.
      --trim-btih int          This will trim the length of the info hash used when --include-btih-in-dn is used. (default 40)
      --user-agent string      HTTP User Agent to use when contacting trackers. (default "rtorrent/0.9.2/0.13.2")
  -w, --working-dir string     Directory to store in-progress torrents. (default "./")
```

## Contact
- hello@nektro.net
- https://twitter.com/nektro

## License
AGPL-3.0
