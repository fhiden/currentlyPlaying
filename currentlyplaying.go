package main

import(
	"fmt"
	"net/http"
	"log"
	"io/ioutil"
	"strings"
	"bytes"

	"github.com/fhiden/spotify"
)
const redirectURI = "http://localhost:8080/callback"
var html = ""

var (
	auth  = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState)
	state = "DontHackMePlz"
	ch = make(chan *spotify.Client)
	sclient *spotify.Client
)

func main(){
	b, err := ioutil.ReadFile("currentsong.html") 
	html = string(b)

	http.HandleFunc("/currentSong", requestCurrentSongs)
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/call", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	
	go http.ListenAndServe(":8080", nil)
	
	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify, use following url to login:", url)

	//wait for client auth.
	client := <-ch
	sclient = client

	user, err := client.CurrentUser()
	if err != nil {
		errorMessageLog(err)
		return
	}
	
	fmt.Println("you are logged in as:", user.ID)
	getCurrentlyPlaying(client)
	for{

	}
}
func getCurrentlyPlaying(client *spotify.Client)(*spotify.CurrentlyPlaying){
	//fmt.Println(x)
	song, err := client.PlayerCurrentlyPlayingOpt(nil)
	if err != nil {
		log.Fatal("Error fetching current song.", err, song)	
	}
	 return song
}
func completeAuth(w http.ResponseWriter, r *http.Request) {
	
	tok, err := auth.Token(state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
		return
	}
	if st :=  r.FormValue("state"); st !=state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	
	// use the token to get an authenticated client
	client := auth.NewClient(tok)
	
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Login Completed!")
	
	ch <- &client
}
func errorMessageLog(err error){
	log.Fatal(err)
}
func requestCurrentSongs(w http.ResponseWriter, r *http.Request){
	song := getCurrentlyPlaying(sclient)

	var finalHTML = strings.Replace(html,"[ALBUM_IMAGE]", song.Item.Album.Images[0].URL, -1)
	finalHTML = strings.Replace(finalHTML,"[SONG_NAME]", song.Item.Name, -1)
	var artist bytes.Buffer
	
	fmt.Println(song.Item.Artists[0].Name)
	fmt.Println(len(song.Item.Artists))

	for x:=0;len(song.Item.Artists)>x; x++ {
		if x>0 {
			artist.WriteString(" & ")
		}
		fmt.Println(song.Item.Artists[x].Name)
		artist.WriteString(song.Item.Artists[x].Name);
	}
	fmt.Println( artist.String())
	finalHTML = strings.Replace(finalHTML,"[ARTIST_NAME]", artist.String(), -1)
	fmt.Fprintf(w,finalHTML)
}