package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	ListenAddr    = flag.String("addr", ":80", "The ADDRESS:PORT to listen on")
	DocumentRoot  = flag.String("dir", "", "The directory to serve up")
	Logging       = flag.Bool("log", false, "Enable/disable logging")
	LogPrefix     = flag.String("log-prefix", "uhttpd", "Set the logging prefix")
	LogPath       = flag.String("log-path", "", "Log to file (leave blank for STDOUT)")
	EnableTLS     = flag.Bool("tls", false, "Enable/disable TLS server")
	TLSListenAddr = flag.String("tls-addr", ":443", "The ADDRESS:PORT for TLS to listen on")
	SvrCertFile   = flag.String("cert", "", "Server certificate file")
	SvrKeyFile    = flag.String("key", "", "Key file for server certificate)")
)

func main() {
	flag.Parse()
	log.SetPrefix(*LogPrefix + " ")
	if *DocumentRoot == "" {
		log.Fatalln("You must specify a directory to serve, with '-dir=\"...\"'")
	}

	if *EnableTLS && (*SvrCertFile == "" || *SvrKeyFile == "") {
		log.Fatalln("You must specify both certificate and key flie for TLS enabled server.")
	}

	handler := http.FileServer(http.Dir(*DocumentRoot))
	if *Logging {
		// Set the logger output.
		var output io.Writer
		if *LogPath == "" {
			output = os.Stdout
		} else {
			var err error
			flags := os.O_CREATE | os.O_APPEND | os.O_WRONLY
			output, err = os.OpenFile(*LogPath, flags, 0644)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
		handler = LoggingHandler(output, http.FileServer(http.Dir(*DocumentRoot)))
		log.Printf("Serving %q", *DocumentRoot)
		log.Printf("Listening on %q for HTTP requests", *ListenAddr)
		if *EnableTLS {
			log.Printf("Listening on %q for HTTPS requests", *TLSListenAddr)
		}
	}

	if *EnableTLS {
		go func() {
			if err := http.ListenAndServe(*ListenAddr, handler); err != nil {
				log.Fatalln(err)
			}
		}()
		if err := http.ListenAndServeTLS(*TLSListenAddr, *SvrCertFile, *SvrKeyFile, handler); err != nil {
			log.Fatalln(err)
		}
	} else {
		if err := http.ListenAndServe(*ListenAddr, handler); err != nil {
			log.Fatalln(err)
		}
	}

	return
}

func LoggingHandler(w io.Writer, h http.Handler) http.Handler {
	logger := log.New(w, *LogPrefix+" ", log.LstdFlags)
	return loggingHandler{Logger: logger, Handler: h}
}

type loggingHandler struct {
	Logger  *log.Logger
	Handler http.Handler
}

func (h loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	h.Handler.ServeHTTP(w, r)
	d := time.Since(t)
	h.Logger.Printf("%s %s %s %s %s (%dms.)",
		r.Proto, r.Method, r.URL.RequestURI(), r.RemoteAddr, r.UserAgent(), d.Nanoseconds()/1e3)
}
