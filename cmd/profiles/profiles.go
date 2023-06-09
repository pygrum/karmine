package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/pygrum/karmine/config"
	"github.com/pygrum/karmine/datastore"
	log "github.com/sirupsen/logrus"
)

var (
	app    = kingpin.New("profiles", "view malware profiles")
	view   = app.Command("view", "view all profiles")
	remove = app.Command("remove", "remove a profile")
	name   = remove.Arg("name", "name(s, comma separated) of profile(s) to remove, or 'all' to remove all").Required().String()
)

func main() {
	db, err := datastore.New()
	if err != nil {
		log.Fatal(err)
	}
	conf, err := config.GetFullConfig()
	if err != nil {
		log.Fatal(err)
	}
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case view.FullCommand():
		fmt.Printf("\n")
		str := fmt.Sprintf("%-36s | %-12s | %s", "uuid", "name", "strain")
		fmt.Println(str)
		fmt.Println(strings.Repeat("=", len(str)))
		for _, p := range conf.Hosts {
			entry := fmt.Sprintf("%-36s | %-12s | %s", p.UUID, p.Name, p.Strain)
			fmt.Println(entry)
			fmt.Println(strings.Repeat("-", len(str)))
		}
	case remove.FullCommand():
		if *name == "all" {
			conf, err := config.GetFullConfig()
			if err != nil {
				log.Fatal(err)
			}
			for _, h := range conf.Hosts {
				if err = db.RemoveProfile(h.Name); err != nil {
					log.Error(err)
				}
			}
			return
		}
		for _, t := range strings.Split(*name, ",") {
			if err = db.RemoveProfile(t); err != nil {
				log.Error(err)
			}
		}
	}
}
