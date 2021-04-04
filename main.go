package main

import (
	"archive/tar"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	logg "github.com/anacrolix/log"
	"github.com/anacrolix/torrent"
	"github.com/nektro/go-util/mbpp"
	"github.com/nektro/go-util/util"
	"github.com/spf13/pflag"
	"golang.org/x/sync/semaphore"
)

var (
	wg    = new(sync.WaitGroup)
	dones = map[string]bool{}
	mtx   = new(sync.Mutex)
	wwg   = new(sync.WaitGroup)

	workingDir  string
	doneDir     string
	dropAfter   int
	dropAfterF  int
	doDownload  bool
	peerLog     *os.File
	seedFor     int
	includeHash bool
	hashTrimLen int
	packTar     bool

	ctx   = context.TODO()
	guard *semaphore.Weighted
)

func main() {

	flagWD := pflag.StringP("working-dir", "w", "./", "Directory to store in-progress torrents.")
	flagDD := pflag.StringP("done-dir", "d", "", "Optional directory to move completed torrents to. (default -w)")
	flagTR := pflag.StringArrayP("torrent", "t", []string{}, "Path of the torrent file you wish to download. (Can be passed multiple times.)")
	flagTD := pflag.String("torrent-dir", "", "Path to directory of torrent files. Will only pick up .torrents.")
	flagMG := pflag.StringArrayP("magnet", "m", []string{}, "Magnet Link to download. (Can be passed multiple times.)")
	flagCR := pflag.IntP("concurrency", "c", 10, "Maximum number of torrents to actively download at a time.")
	flagDT := pflag.Int("drop-after", 35, "Minutes to drop torrents with no progress after. (disable -1)")
	flagUA := pflag.String("user-agent", "rtorrent/0.9.2/0.13.2", "HTTP User Agent to use when contacting trackers.")
	flagPI := pflag.String("peer-id", "-lt0D20-", "Bittorrent peer_id.")
	flagDF := pflag.Int("drop-after-force", -1, "Minutes to drop torrents after regardless of progress. (disable -1)")
	flagDL := pflag.Bool("do-download", true, "Setting this flag to false will make all torrents idle in client.")
	flagPL := pflag.String("peers-log", "", "An optional path to log file that will list all peers per torrent.")
	flagSF := pflag.IntP("seed-for", "s", 0, "When positive, minutes to seed torrents for. (-1: forever) (0: only leech)")
	flagIH := pflag.BoolP("include-btih-in-dn", "i", false, "If true, folder name will be 'btih dn' instead of 'dn'.")
	flagTB := pflag.Int("trim-btih", 40, "This will trim the length of the info hash used when --include-btih-in-dn is used.")
	flagDH := pflag.Bool("disable-dht", false, "Setting this will disable DHT.")
	flagMF := pflag.String("magnet-file", "", "Path to a text file with magnet links on each line.")
	flagPT := pflag.Bool("pack-tar", false, "Enabling this will pack torrent folder into a .tar so that it only takes up a single file.")
	pflag.Parse()

	//
	//

	workingDir = *flagWD
	workingDir, _ = filepath.Abs(workingDir)
	util.DieOnError(util.Assert(util.DoesDirectoryExist(workingDir), "Path of --working-dir is '"+workingDir+"/' and must exist!"))

	doneDir = *flagDD
	if doneDir == "" {
		doneDir = workingDir
	}
	doneDir, _ = filepath.Abs(doneDir)
	util.DieOnError(util.Assert(util.DoesDirectoryExist(doneDir), "Path of --done-dir is '"+doneDir+"/' and must exist!"))

	if len(*flagPL) > 0 {
		plp := *flagPL
		plp, _ = filepath.Abs(plp)
		peerLog, _ = os.OpenFile(plp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	}

	dropAfter = *flagDT
	guard = semaphore.NewWeighted(int64(*flagCR))
	dropAfterF = *flagDF
	doDownload = *flagDL
	seedFor = *flagSF
	includeHash = *flagIH
	hashTrimLen = *flagTB
	packTar = *flagPT

	//
	//

	util.Log("Working directory:\t", workingDir)
	util.Log("Finish directory:\t", doneDir)
	util.Log("--do-download:", doDownload)

	//
	//

	mbpp.Init(*flagCR)

	util.Log("Starting up client...")
	cf := torrent.NewDefaultClientConfig()
	cf.NoDHT = *flagDH
	cf.DisablePEX = true
	cf.Debug = false
	cf.Bep20 = *flagPI + randomHex(20-len(*flagPI))
	cf.ExtendedHandshakeClientVersion = "3.00+"
	cf.DataDir = workingDir
	cf.DisableIPv6 = true
	cf.ListenPort = 0
	cf.Logger = logg.Discard
	cf.HTTPUserAgent = *flagUA
	cf.Seed = *flagSF != 0
	util.Log("peer_id:", cf.Bep20)

	c, err := torrent.NewClient(cf)
	util.DieOnError(err)
	defer c.Close()

	//

	magFilePth := *flagMF
	if len(magFilePth) > 0 {
		magFilePth, _ := filepath.Abs(magFilePth)
		util.DieOnError(util.Assert(util.DoesFileExist(magFilePth), "Path of --magnet-file is '"+magFilePth+"' and must exist!"))
		util.ReadFileLines(magFilePth, func(s string) {
			addT(c.AddMagnet, s)
		})
	}

	torrentPths := *flagTR
	if len(torrentPths) > 0 {
		for _, item := range torrentPths {
			torrentP, _ := filepath.Abs(item)
			util.DieOnError(util.Assert(util.DoesFileExist(torrentP), "Path of --torrent is '"+torrentP+"' and must exist!"))
			addT(c.AddTorrentFromFile, torrentP)
		}
	}

	torrentDir := *flagTD
	if len(torrentDir) > 0 {
		if util.DoesDirectoryExist(torrentDir) {
			filepath.Walk(torrentDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info == nil {
					return nil
				}
				if strings.HasSuffix(info.Name(), ".torrent") {
					addT(c.AddTorrentFromFile, path)
				}
				if strings.HasSuffix(info.Name(), ".magnet.txt") {
					addT(c.AddMagnet, string(util.ReadFile(path)))
				}
				return nil
			})
		}
	}

	magnetLnks := *flagMG
	if len(magnetLnks) > 0 {
		for _, item := range magnetLnks {
			addT(c.AddMagnet, item)
		}
	}

	//

	time.Sleep(time.Second)
	c.WaitAll()
	wg.Wait()
	wwg.Wait()
	util.Log("Completed after", len(dones), "torrents downloaded.")
}

func randomHex(n int) string {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)[:n]
}

func addT(f func(string) (*torrent.Torrent, error), s string) {
	wg.Add(1)
	guard.Acquire(ctx, 1)
	go func() {
		t, err := f(s)
		if err != nil {
			wg.Done()
			guard.Release(1)
			return
		}

		if dropAfter > 0 {
			go func() {
				time.Sleep(time.Minute * time.Duration(dropAfter))
				if t.BytesCompleted() == 0 {
					closeT(t, true)
				}
			}()
		}
		if dropAfterF > 0 {
			go func() {
				time.Sleep(time.Minute * time.Duration(dropAfterF))
				closeT(t, true)
			}()
		}

		<-t.GotInfo()
		if includeHash {
			t.Info().Name = (t.InfoHash().HexString()[:hashTrimLen] + " " + t.Info().Name)
		}
		name := t.Info().Name

		if packTar {
			if util.DoesFileExist(doneDir + "/" + name + ".tar") {
				closeT(t, true)
				return
			}
		} else {
			if util.DoesFileExist(doneDir + "/" + name) {
				go func() {
					if seedFor > 0 {
						os.Rename(doneDir+"/"+name, workingDir+"/"+name)
						t.VerifyData()
						time.Sleep(time.Minute * time.Duration(seedFor))
						os.Rename(workingDir+"/"+name, doneDir+"/"+name)
					}
					closeT(t, true)
				}()
				return
			}
		}

		if doDownload {
			go func() {
				mbpp.CreateJob(t.Name(), func(bar *mbpp.BarProxy) {
					bar.AddToTotal(int64(t.NumPieces()))
					prev_completed := 0
					for {
						this_completed := 0
						for _, r := range t.PieceStateRuns() {
							if r.Complete {
								this_completed += r.Length
							}
						}
						bar.Increment(this_completed - prev_completed)
						if this_completed == t.NumPieces() {
							break
						}
						prev_completed = this_completed
					}
				})
				if seedFor >= 0 {
					wwg.Add(1)
					go func() {
						time.Sleep(time.Minute * time.Duration(seedFor))
						closeT(t, false)

						workName := workingDir + "/" + name
						doneName := doneDir + "/" + name

						if packTar {
							stat, _ := os.Stat(workName)
							out1, _ := os.Create(workName + ".tar")
							out2 := tar.NewWriter(out1)

							if stat.IsDir() {
								filepath.Walk(workName, func(pathS string, info os.FileInfo, err error) error {
									if util.DoesDirectoryExist(pathS) {
										return nil
									}
									name := strings.TrimPrefix(pathS, workName+"/")
									writeTarFile(out2, workName, name)
									return nil
								})
							} else {
								writeTarFile(out2, workingDir, name)
							}
							out2.Close()

							os.RemoveAll(workName)
							workName += ".tar"
							doneName += ".tar"
						}
						if doneDir != workingDir {
							os.Rename(workName, doneName)
						}
						wwg.Done()
					}()
				}
			}()
			t.DownloadAll()
		}
	}()
}

func closeT(t *torrent.Torrent, early bool) {
	mtx.Lock()
	defer mtx.Unlock()
	h := t.InfoHash().HexString()
	b, ok := dones[h]
	if ok && b {
		return
	}
	if peerLog != nil {
		peers := []string{}
		for _, p := range t.PeerConns() {
			peers = append(peers, p.RemoteAddr.String())
		}
		fmt.Fprintln(peerLog, h, peers)
	}
	dones[h] = true
	t.Drop()
	wg.Done()
	guard.Release(1)
}

func writeTarFile(w *tar.Writer, dir, name string) {
	content := util.ReadFile(dir + "/" + name)
	w.WriteHeader(&tar.Header{
		Name: name,
		Mode: 0600,
		Size: int64(len(content)),
	})
	w.Write(content)
}
