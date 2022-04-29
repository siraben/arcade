package arcade

import (
	"sort"
	"sync"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

type GamesListView struct {
	View

	mu sync.RWMutex

	lobbies     map[string]*Lobby
	selectedRow int
}

var header = []string{
	"▄▀█ █▀ █▀▀ █ █   ▄▀█ █▀█ █▀▀ ▄▀█ █▀▄ █▀▀",
	"█▀█ ▄█ █▄▄ █ █   █▀█ █▀▄ █▄▄ █▀█ █▄▀ ██▄",
}

var footer = []string{
	"[C]reate new lobby      [J]oin selected lobby by IP address",
}

var glv_join_box = ""
var selectedLobbyKey = ""

const (
	nameColX    = 4
	gameColX    = 30
	playersColX = 40
	pingColX    = 70

	tableX1 = 3
	tableY1 = 7
	tableX2 = 76
	tableY2 = 19

	joinbox_X1 = 20
	joinbox_Y1 = 9
	joinbox_X2 = 59
	joinbox_Y2 = 15
)

var glv_code_input_string = ""
var glv_code = ""
var glv_code_editing = false

func NewGamesListView() *GamesListView {
	return &GamesListView{
		lobbies: make(map[string]*Lobby),
	}
}

func (v *GamesListView) Init() {
	server.RLock()
	clients := server.clients
	server.RUnlock()

	for _, client := range clients {
		if client.Distributor {
			continue
		}

		client.Send(NewHelloMessage())
	}
}

func (v *GamesListView) ProcessEvent(evt tcell.Event) {
	switch evt := evt.(type) {
	case *tcell.EventKey:
		switch evt.Key() {
		case tcell.KeyDown:
			v.selectedRow++

			v.mu.RLock()
			if v.selectedRow > len(v.lobbies)-1 {
				v.selectedRow = len(v.lobbies) - 1
			}
			v.mu.RUnlock()
		case tcell.KeyUp:
			v.selectedRow--

			if v.selectedRow < 0 {
				v.selectedRow = 0
			}
		case tcell.KeyBackspace:
			if glv_join_box != "" {
				glv_code_input_string = glv_code_input_string[:len(glv_code_input_string)-1]
			}
		case tcell.KeyEnter:
			if glv_join_box == "join_code" {
				if len(glv_code_input_string) == 4 {
					glv_code = glv_code_input_string
					selectedLobby := v.lobbies[selectedLobbyKey]
					self, _ := server.GetClient(selectedLobby.HostID)

					joinPlayer := Player{
						ClientID: server.ID,
						Username: "joiningjoanna",
						Host:     false,
					}
					go self.Send(NewJoinMessage(glv_code, joinPlayer))
				} else {
					glv_join_box = "join_code_short"
				}

			}
		case tcell.KeyRune:
			if glv_join_box == "" {
				switch evt.Rune() {
				case 'c':
					glv_join_box = ""
					mgr.SetView(NewLobbyCreateView())
				case 't':
					// tg := CreateGame("bruh", false, "Tron", 8, "1", "1")
					// tg.AddPlayer(&Player{Client: *NewClient("addr1"), Username: "bob", Status:  "chillin",Host: true})
					// mgr.SetView(tg)
				case 'j':
					if len(v.lobbies) != 0 {
						v.mu.RLock()

						keys := make([]string, 0, len(v.lobbies))

						for k := range v.lobbies {
							keys = append(keys, k)
						}
						sort.Strings(keys)

						selectedLobbyKey = keys[v.selectedRow]
						selectedLobby := v.lobbies[keys[v.selectedRow]]
						if selectedLobby.Private {
							glv_join_box = "join_code"
						} else {
							self, _ := server.GetClient(selectedLobby.HostID)

							joinPlayer := Player{
								ClientID: server.ID,
								Username: "joiningjoanna",
								Host:     false,
							}
							go self.Send(NewJoinMessage("", joinPlayer))
						}
						v.mu.RUnlock()

					}

				}
			} else {
				if len(glv_code_input_string) < 4 {
					glv_code_input_string += string(evt.Rune())
				}
			}
		}
	}
}

func (v *GamesListView) ProcessMessage(from *Client, p interface{}) interface{} {
	switch p := p.(type) {
	case LobbyInfoMessage:
		v.mu.Lock()
		v.lobbies[p.Lobby.ID] = p.Lobby
		v.mu.Unlock()
	}

	return nil
}

func (v *GamesListView) Render(s *Screen) {
	width, height := s.Size()

	// Green text on default background
	sty := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorGreen)

	// Draw ASCII ARCADE header
	headerX := (width - utf8.RuneCountInString(header[0])) / 2
	s.DrawText(headerX, 1, sty, header[0])
	s.DrawText(headerX, 2, sty, header[1])

	// Draw box surrounding games list
	s.DrawBox(tableX1-1, 4, tableX2+1, tableY2+1, sty, true)

	// Draw footer with navigation keystrokes
	s.DrawText((width-len(footer[0]))/2, height-2, sty, footer[0])

	// Draw column headers
	s.DrawText(nameColX, 5, sty, "NAME")
	s.DrawText(gameColX, 5, sty, "GAME")
	s.DrawText(playersColX, 5, sty, "PLAYERS")
	s.DrawText(pingColX, 5, sty, "PING")

	// Draw border below column headers
	s.DrawLine(3, 6, tableX2, 6, sty, true)
	s.DrawText(2, 6, sty, "╠")
	s.DrawText(width-3, 6, sty, "╣")

	// Draw selected row
	selectedSty := tcell.StyleDefault.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorWhite)
	sty_bold := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorDarkGreen)

	i := 0
	v.mu.RLock()

	keys := make([]string, 0, len(v.lobbies))

	for k := range v.lobbies {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, lobbyID := range keys {
		lobby := v.lobbies[lobbyID]
		y := tableY1 + i
		rowSty := sty

		if i == v.selectedRow {
			rowSty = selectedSty
		}

		name := lobby.Name
		game := "Pong"
		players := "1/2"
		ping := "25ms"

		s.DrawEmpty(tableX1, y, nameColX-1, y, rowSty)
		s.DrawText(nameColX, y, rowSty, name)
		s.DrawEmpty(nameColX+len(name), y, gameColX-1, y, rowSty)
		s.DrawText(gameColX, y, rowSty, game)
		s.DrawEmpty(gameColX+len(game), y, playersColX-1, y, rowSty)
		s.DrawText(playersColX, y, rowSty, players)
		s.DrawEmpty(playersColX+len(players), y, pingColX-1, y, rowSty)
		s.DrawText(pingColX, y, rowSty, ping)
		s.DrawEmpty(pingColX+len(ping), y, tableX2, y, rowSty)

		i++
	}

	if glv_join_box != "" {
		selectedLobby := v.lobbies[selectedLobbyKey]
		// Draw box surrounding games list
		s.DrawBox(joinbox_X1, joinbox_Y1, joinbox_X2, joinbox_Y2, sty, true)

		joinheader := "Joining private game " + selectedLobby.Name
		s.DrawText((width-len(joinheader))/2, joinbox_Y1+1, sty, "Joining private game ")
		s.DrawText((width-len(joinheader))/2+len(selectedLobby.Name), joinbox_Y1+1, sty_bold, selectedLobby.Name)
		codeHeader := "Enter code: "
		s.DrawText((width-len(codeHeader)-5)/2, joinbox_Y1+2, sty, codeHeader)
		s.DrawText((width-len(codeHeader)-5)/2+len(codeHeader), joinbox_Y1+2, sty_bold, glv_code_input_string)
		if glv_join_box == "join_code_short" {
			shortString := "Code must be four characters long."
			s.DrawText((width-len(shortString))/2, joinbox_Y1+3, sty_bold, shortString)
		}
	}
	v.mu.RUnlock()
}
