package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/b4fun/prom-snapshot/pkg/adapterprom"
)

var (
	flagAddr = kingpin.Flag("api.addr", "prometheus v1 api addr").
			Default(":8080").
			String()
	flagSnapshots = kingpin.Flag("snapshot", "snapshot to load ({name}={path})").
			Short('s').
			Strings()
)

func main() {
	kingpin.Parse()

	snapshotOpts := map[string]*adapterprom.Opts{}
	for _, snapshot := range *flagSnapshots {
		name, opts, err := parseSnapshotSettings(snapshot)
		if err != nil {
			kingpin.Fatalf("invalid snapshot settings: %s", err)
			return
		}
		snapshotOpts[name] = opts
	}
	if len(snapshotOpts) == 0 {
		kingpin.Fatalf("empty snapshot settings")
		return
	}

	mux := http.NewServeMux()
	for snapshot, opts := range snapshotOpts {
		prefix := fmt.Sprintf("/%s", strings.TrimPrefix(snapshot, "/"))
		if err := opts.RegisterV1API(prefix, mux); err != nil {
			kingpin.Fatalf("failed to register snapshot %s: %s", snapshot, err)
			return
		}
	}

	if err := http.ListenAndServe(*flagAddr, mux); err != nil {
		kingpin.Fatalf("server http: %s", err)
		return
	}
}

func parseSnapshotSettings(s string) (string, *adapterprom.Opts, error) {
	ss := strings.Split(s, "=")
	if len(ss) != 2 {
		return "", nil, fmt.Errorf("snapshot settings should be {name}={path} format, got: %s", s)
	}

	name, path := ss[0], ss[1]

	stat, err := os.Stat(path)
	if err != nil {
		return "", nil, fmt.Errorf("invalid snapshot path %s: %w", path, err)
	}
	if !stat.IsDir() {
		return "", nil, fmt.Errorf("snapshot path %s should be dir", path)
	}

	return name, adapterprom.DefaultOpts(path), nil
}
