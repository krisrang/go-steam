package steam

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

var (
	user  string
	limit int

	apiRoot = "http://steamcommunity.com/id/"
)

type User struct {
	SteamID64      string `xml:"steamID64"`
	SteamID        string `xml:"steamID"`
	OnlineState    string `xml:"onlineState"`
	StateMessage   string `xml:"stateMessage"`
	AvatarIcon     string `xml:"avatarIcon"`
	AvatarMedium   string `xml:"avatarMedium"`
	AvatarFull     string `xml:"avatarFull"`
	CustomURL      string `xml:"customURL"`
	MemberSince    string `xml:"memberSince"`
	SteamRating    string `xml:"steamRating"`
	HoursPlayed2Wk string `xml:"hoursPlayed2Wk"`
	Location       string `xml:"location"`
	Realname       string `xml:"realname"`
	Summary        string `xml:"summary"`
	GameCount      int
	// MostPlayed     []MostPlayedGame `xml:"mostPlayedGames>mostPlayedGame"`
}

func (u User) FullURL() string {
	return apiRoot + u.CustomURL
}

func (u User) RatingDescription() string {
	title := "Playing on PS3"

	switch u.SteamRating {
	case "10":
		title = "EAGLES SCREAM"
	case "9":
		title = "Still not 10"
	case "8":
		title = "COBRA KAI!"
	case "7":
		title = "Wax on, Wax off"
	case "6":
		title = "Oooh! Shiny!"
	case "5":
		title = "Halfway Cool"
	case "4":
		title = "Master of Nothing"
	case "3":
		title = "Shooting Blanks"
	case "2":
		title = "Nearly Lifeless"
	case "1":
		title = "El Terrible!"
	}

	return title
}

// type MostPlayedGame struct {
//  Name          string `xml:"gameName"`
//  Link          string `xml:"gameLink"`
//  Icon          string `xml:"gameIcon"`
//  Logo          string `xml:"gameLogo"`
//  LogoSmall     string `xml:"gameLogoSmall"`
//  HoursPlayed   string `xml:"hoursPlayed"`
//  HoursOnRecord string `xml:"hoursOnRecord"`
// }

type GamesList struct {
	Games []Game `xml:"games>game"`
}

func (g GamesList) RecentlyPlayed() *GamesList {
	playedLast2Weeks := []Game{}
	notPlayedLastWeeks := []Game{}
	games := g.Games

	for i := range games {
		game := games[i]

		if game.HoursLast2Weeks != "" {
			playedLast2Weeks = append(playedLast2Weeks, game)
		} else {
			notPlayedLastWeeks = append(notPlayedLastWeeks, game)
		}
	}

	sort.Sort(GamesByLast2Weeks{Games: playedLast2Weeks})
	sort.Sort(GamesByHours{Games: notPlayedLastWeeks})

	sortedGames := append(playedLast2Weeks, notPlayedLastWeeks...)

	if len(g.Games) > limit && limit > 0 {
		g.Games = sortedGames[:limit]
	} else {
		g.Games = sortedGames
	}

	return &g
}

func (g GamesList) HoursPlayed2Wk() float64 {
	hours := 0.0
	for i := range g.Games {
		game := g.Games[i]
		hrs, _ := strconv.ParseFloat(game.HoursLast2Weeks, 64)
		hours += hrs
	}

	return hours
}

type Game struct {
	AppID           string `xml:"appID"`
	Name            string `xml:"name"`
	Logo            string `xml:"logo"`
	StoreLink       string `xml:"storeLink"`
	HoursLast2Weeks string `xml:"hoursLast2Weeks"`
	HoursOnRecord   string `xml:"hoursOnRecord"`
}

type GamesByLast2Weeks GamesList
type GamesByHours GamesList

// yes, this is bad
func (s GamesByLast2Weeks) Less(i, j int) bool {
	a, _ := strconv.ParseFloat(s.Games[i].HoursLast2Weeks, 64)
	b, _ := strconv.ParseFloat(s.Games[j].HoursLast2Weeks, 64)

	return a > b
}

func (s GamesByHours) Less(i, j int) bool {
	a, _ := strconv.ParseFloat(s.Games[i].HoursOnRecord, 64)
	b, _ := strconv.ParseFloat(s.Games[j].HoursOnRecord, 64)

	return a > b
}

func (s GamesByLast2Weeks) Len() int      { return len(s.Games) }
func (s GamesByLast2Weeks) Swap(i, j int) { s.Games[i], s.Games[j] = s.Games[j], s.Games[i] }

func (s GamesByHours) Len() int      { return len(s.Games) }
func (s GamesByHours) Swap(i, j int) { s.Games[i], s.Games[j] = s.Games[j], s.Games[i] }

func SetConfig(u string, l int) {
	user = u
	limit = l
}

func GetUser() *User {
	uri := apiRoot + user + "?xml=1"
	userdata := &User{}
	getData(uri, userdata)

	userdata.Summary = strings.Replace(userdata.Summary, "<br>", "", -1)
	userdata.GameCount = len(GetGames().Games)

	return userdata
}

func GetGames() *GamesList {
	uri := apiRoot + user + "/games/?xml=1"
	gamedata := &GamesList{}
	getData(uri, gamedata)
	return gamedata
}

// PRIVATE

func getData(uri string, i interface{}) {
	data := getRequest(uri)
	xmlUnmarshal(data, i)
}

func getRequest(uri string) []byte {
	res, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	return body
}

func xmlUnmarshal(b []byte, i interface{}) {
	err := xml.Unmarshal(b, i)
	if err != nil {
		log.Fatal(err)
	}
}
