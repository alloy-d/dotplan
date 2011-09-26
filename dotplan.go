package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/inotify"
	"path"

	"github.com/alloy-d/go140"
	"github.com/alloy-d/goauth"
)

var home string
var authFile string
func init() {
	home = os.Getenv("HOME")
	authFile = path.Join(home, ".dotplan.goauth")
}

func processError(err os.Error) {
	if err != nil {
		log.Fatal(err)
	}
}

func authorize(api *go140.API) (err os.Error) {
	api.ConsumerKey = "no6XLKksvxtRHtS3aorzg"
	api.ConsumerSecret = "EJr4Nr9qcT7zjpuLTnmEnMPgNd0I0QEiDh6ksrjLWI"
	api.SignatureMethod = oauth.HMAC_SHA1

	api.RequestTokenURL = "https://api.twitter.com/oauth/request_token"
	api.OwnerAuthURL = "https://api.twitter.com/oauth/authorize"
	api.AccessTokenURL = "https://api.twitter.com/oauth/access_token"
	api.Callback = "oob"

	api.Root = "https://api.twitter.com"

	err = api.Load(authFile)
	if err != nil {
		err = api.GetRequestToken()
		processError(err)
		url, err := api.AuthorizationURL()
		processError(err)
		fmt.Printf("Please visit the following URL for authorization:\n%s\n", url)

		var verifier string
		fmt.Printf("PIN: ")
		fmt.Scanf("%s", &verifier)
		err = api.GetAccessToken(verifier)
		processError(err)

		err = api.Save(authFile)
		if err != nil {
			log.Println("Couldn't save authorization information", err)
		}
	}

	return nil
}

func update(filename string, api *go140.API) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	_, err = api.Update(string(contents))
	if err != nil {
		log.Println(err)
	}
}

func main() {
	api := new(go140.API)
	authorize(api)

	file := path.Join(os.Getenv("HOME"), ".plan")
	in_args := inotify.IN_CLOSE_WRITE

	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	err = watcher.AddWatch(os.Getenv("HOME"), in_args)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case ev := <-watcher.Event:
			if ev.Mask == inotify.IN_CLOSE_WRITE ||
					ev.Mask == inotify.IN_CREATE {
				if ev.Name == file {
					log.Println("Updating.")
					update(file, api)
				}
			}
		case err := <-watcher.Error:
			log.Println("error:", err)
		}
	}
}
