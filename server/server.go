package server

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"github.com/google/go-github/github"
	"io/ioutil"
	"net/http"
	"strconv"
)

type ReleaseEvent struct {
	Action  string                   `json:"action,omitempty"`
	Release github.RepositoryRelease `json:"release,omitempty"`
	Repo    github.Repository        `json:"repository,omitempty"`
	Sender  github.User              `json:"sender,omitempty"`
}

func (e *ReleaseEvent) String() (output string) {
	output += "action:   " + e.Action + "\n"
	output += "sender:  " + e.Sender.String() + "\n"
	output += "repo:   " + e.Repo.String() + "\n"
	output += "tag:   " + *e.Release.TagName + "\n"
	output += "body:   " + *e.Release.Body + "\n"

	return
}

type Server struct {
	Port   int               // Port to listen on. Defaults to 80
	Path   string            // Path to receive on. Defaults to "/event"
	Secret string            // Option secret key for authenticating via HMAC
	Events chan ReleaseEvent // Channel of events. Read from this channel to get push events as they happen.
}

// Create a new server with sensible defaults.
// By default the Port is set to 80 and the Path is set to `/event`
func NewServer() *Server {
	return &Server{
		Port:   80,
		Path:   "/event",
		Events: make(chan ReleaseEvent, 10), // buffered to 10 items
	}
}

// Spin up the server and listen for github webhook push events. The events
// will be passed to Server.Events channel.
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(":"+strconv.Itoa(s.Port), s)
}

// Inside a go-routine, spin up the server and listen for github webhook release
// events. The events will be passed to Server.Events channel.
func (s *Server) GoListenAndServe() {
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()
}

// Satisfies the http.Handler interface.
// Instead of calling Server.ListenAndServe you can integrate server.Server
// inside your own http server. If you are using server.Server in his way
// Server.Path should be set to match your mux pattern and Server.Port will be
// ignored.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != "POST" {
		http.Error(w, "405 Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if req.URL.Path != s.Path {
		http.Error(w, "404 Not found", http.StatusNotFound)
		return
	}

	eventType := req.Header.Get("X-GitHub-Event")
	if eventType == "" {
		http.Error(w, "400 Bad Request - Missing X-GitHub-Event Header", http.StatusBadRequest)
		return
	}
	if eventType != "release" {
		http.Error(w, "400 Bad Request - Unknown Event Type "+eventType, http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If we have a Secret set, we should check the MAC
	if s.Secret != "" {
		sig := req.Header.Get("X-Hub-Signature")

		if sig == "" {
			http.Error(w, "403 Forbidden - Missing X-Hub-Signature required for HMAC verification", http.StatusForbidden)
			return
		}

		mac := hmac.New(sha1.New, []byte(s.Secret))
		mac.Write(body)
		expectedMAC := mac.Sum(nil)
		expectedSig := "sha1=" + hex.EncodeToString(expectedMAC)
		if !hmac.Equal([]byte(expectedSig), []byte(sig)) {
			http.Error(w, "403 Forbidden - HMAC verification failed", http.StatusForbidden)
			return
		}
	}

	var releaseEvent ReleaseEvent

	if eventType == "release" {
		json.Unmarshal(body, &releaseEvent)
	} else {
		http.Error(w, "Unknown Event Type "+eventType, http.StatusInternalServerError)
		return
	}

	// We've built our releaseEvent - put it into the channel and we're done
	go func() {
		s.Events <- releaseEvent
	}()

	w.Write([]byte(releaseEvent.String()))
}
