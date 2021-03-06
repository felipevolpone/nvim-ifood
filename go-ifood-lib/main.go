package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"strings"
	"time"

	"github.com/neovim/go-client/nvim"
	"github.com/neovim/go-client/nvim/plugin"
	"github.com/tidwall/gjson"
)

var buf nvim.Buffer
var win nvim.Window
var accessToken string
var refreshToken string
var selectedAddress Address
var addresses []Address
var home gjson.Result
var merchants gjson.Result
var selectedCardHomeID string

func main() {
	plugin.Main(func(p *plugin.Plugin) error {

		p.HandleCommand(&plugin.CommandOptions{Name: "IfoodLogin"},
			func() (string, error) {
				login(p)
				return "", nil
			})

		p.HandleCommand(&plugin.CommandOptions{Name: "IfoodAddress"},
			func() (string, error) {
				listAddress(p)
				mappings := map[string]string{
					"<cr>": ":IfoodPickAddress<cr>",
				}
				setMappings(p, mappings)
				return "", nil
			})
		p.HandleCommand(&plugin.CommandOptions{Name: "IfoodPickAddress"},
			func() (string, error) {
				pickAddress(p)
				return "", nil
			})

		p.HandleCommand(&plugin.CommandOptions{Name: "IfoodHome"},
			func() (string, error) {
				showHome(p)
				mappings := map[string]string{
					"<cr>": ":IfoodPickHomeCard<cr>",
				}
				setMappings(p, mappings)
				return "", nil
			})
		p.HandleCommand(&plugin.CommandOptions{Name: "IfoodPickHomeCard"},
			func() (string, error) {
				pickHomeCard(p)
				return "", nil
			})

		p.HandleCommand(&plugin.CommandOptions{Name: "IfoodMerchants"},
			func() (string, error) {
				showMerchants(p)
				return "", nil
			})
		return nil
	})

}

func listAddress(p *plugin.Plugin) {
	addresses = ListAddress()
	repl := [][]byte{
		[]byte("Choose your address:"),
		[]byte("-------------------------------"),
		[]byte(""),
	}

	for _, address := range addresses {
		a := fmt.Sprintf("%s %s %s %s", address.StreetName, address.StreetNumber, address.Complement, address.Neighborhood)
		repl = append(repl, []byte(a))
	}

	createWindow(p, "pick an address", repl)
}

func createWindow(p *plugin.Plugin, title string, content [][]byte) {
	buf, err := p.Nvim.CreateBuffer(false, true)
	if err != nil {
		fmt.Println("err", err)
	}

	p.Nvim.SetBufferOption(buf, "bufhidden", "wipe")
	p.Nvim.SetBufferOption(buf, "filetype", "whid")

	var width int
	p.Nvim.Option("columns", &width)

	var height int
	p.Nvim.Option("columns", &height)

	winHeight := int(math.Ceil(float64(height) * 0.3))
	winWidth := int(math.Ceil(float64(width) * float64(0.5)))

	row := math.Ceil((float64(height)-float64(winHeight))/2 - 1)
	col := math.Ceil((float64(width)-float64(winWidth))/2 - 1)

	p.Nvim.SetBufferLines(buf, 0, -1, false, content)
	p.Nvim.AddBufferHighlight(buf, -1, "WhidHeader", 0, 0, -1)

	opts := nvim.WindowConfig{
		Relative: "editor",
		Width:    winWidth,
		Height:   winHeight,
		Row:      row,
		Col:      col,
		Style:    "minimal",
		Anchor:   "NW",
	}

	win, err := p.Nvim.OpenWindow(buf, true, &opts)
	if err != nil {
		fmt.Println("err", err)
	}

	p.Nvim.SetWindowOption(win, "cursorline", true)
	p.Nvim.SetWindowCursor(win, [2]int{4, 0})
	p.Nvim.SetBufferOption(buf, "modifiable", false)
}

func setMappings(p *plugin.Plugin, mappings map[string]string) {
	opts := map[string]bool{"nowait": true, "noremap": true, "silent": true}
	for k, v := range mappings {
		err := p.Nvim.SetBufferKeyMap(buf, "n", k, v, opts)
		if err != nil {
			fmt.Println("err", err)
		}
	}
}

func pickHomeCard(p *plugin.Plugin) {
	lineBytes, err := p.Nvim.CurrentLine()
	if err != nil {
		fmt.Println("err", err)
	}

	line := string(lineBytes)
	err = p.Nvim.CloseWindow(win, true)
	if err != nil {
		fmt.Println("err", err)
	}
	selectedCardHomeID = homeCardFromString(line)

	p.Nvim.Command("IfoodMerchants")
}

func pickAddress(p *plugin.Plugin) {
	lineBytes, err := p.Nvim.CurrentLine()
	if err != nil {
		fmt.Println("err", err)
	}

	line := string(lineBytes)
	err = p.Nvim.CloseWindow(win, true)
	if err != nil {
		fmt.Println("err", err)
	}
	selectedAddress = adressFromString(line)

	p.Nvim.Command("IfoodHome")
}

func listMerchants(p *plugin.Plugin) {
	content := [][]byte{
		[]byte("pick a merchant"),
		[]byte(""),
		[]byte("hamburgueria xpto"),
		[]byte("pizarria foo bar"),
	}

	mappings := map[string]string{
		"<cr>": ":call IfoodShowMerchantHome()<cr>",
	}
	createWindow(p, "any", content)
	setMappings(p, mappings)
}

func showHome(p *plugin.Plugin) {
	home = GetHome()

	repl := [][]byte{
		[]byte("Feeling hungry? What do you want to eat?"),
		[]byte("----------------------------------"),
	}

	for _, c := range home.Get("sections.0.cards.0.data.contents").Array() {
		repl = append(repl, []byte(c.Get("title").String()))
	}

	createWindow(p, "pick a list", repl)
}

func login(p *plugin.Plugin) {
	/*
		// for debug purpose
		openCreds()
		p.Nvim.Command("IfoodAddress")
		return
	*/

	var email string
	err := p.Nvim.Call("input", &email, "Lets login. Write your email: ")
	if err != nil {
		fmt.Println("err", err)
	}

	token := AskOtpCode(email)

	err = p.Nvim.Echo([]nvim.TextChunk{{Text: "get the code code in your email"}}, false, map[string]interface{}{})
	if err != nil {
		fmt.Println("err", err)
	}

	time.Sleep(time.Millisecond * 700)

	var otpCode string
	err = p.Nvim.Call("input", &otpCode, "Now, insert the code: ")
	if err != nil {
		fmt.Println("err", err)
	}

	accessToken = ClaimOtpCode(otpCode, token)
	err = p.Nvim.Echo([]nvim.TextChunk{{Text: "You're logged in"}}, false, map[string]interface{}{})
	if err != nil {
		fmt.Println("err", err)
	}

	accessToken, refreshToken = Auth(email, accessToken)
	saveLocal()

	p.Nvim.Command("IfoodAddress")
}

func saveLocal() {
	type creds struct {
		AccessToken  string
		RefreshToken string
	}

	data := creds{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	file, _ := json.MarshalIndent(data, "", " ")
	ioutil.WriteFile("cred.json", file, 0644)
}

func openCreds() {
	file, _ := ioutil.ReadFile("cred.json")
	type creds struct {
		AccessToken  string
		RefreshToken string
	}
	data := creds{}
	_ = json.Unmarshal([]byte(file), &data)
	accessToken = data.AccessToken
	refreshToken = data.RefreshToken

	accessToken, refreshToken = RefreshToken()
}

func adressFromString(address string) Address {
	for _, a := range addresses {
		line := fmt.Sprintf("%s %s %s %s", a.StreetName, a.StreetNumber, a.Complement, a.Neighborhood)
		if line == address {
			return a
		}
	}
	return Address{}
}

func homeCardFromString(selectedCard string) string {
	for _, c := range home.Get("sections.0.cards.0.data.contents").Array() {
		title := c.Get("title").String()
		if title == selectedCard {
			action := c.Get("action").String()
			identifier := strings.Split(action, "?")[1]
			x := strings.Split(identifier, "=")[1]
			return strings.Split(x, "&")[0]
		}
	}
	return ""
}

func showMerchants(p *plugin.Plugin) {
	merchants := ShowMerchants(selectedCardHomeID, selectedAddress.Coordinates.Latitude, selectedAddress.Coordinates.Longitude)

	repl := [][]byte{
		[]byte("Now it's time to pick a merchant! Go on!"),
		[]byte("----------------------------------"),
	}

	for _, c := range merchants.Get("sections.0.cards.0.data.contents").Array() {
		title := c.Get("name").String()
		repl = append(repl, []byte(title))
	}

	createWindow(p, "now you can choose the merchant", repl)
}
