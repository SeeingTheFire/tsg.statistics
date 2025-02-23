package services

import (
	"encoding/json"
	"fmt"
	"github.com/SeeingTheFire/tsg.statistics/internal/dtos"
	"io"
	"net/http"
)

const tsgReplays = "https://game.tsgames.ru/ajax.php?a=l&params%5Bf%5D%5B%5D=2&params%5Bf%5D%5B%5D=3&params%5Bf%5D%5B%5D=4&params%5Bf%5D%5B%5D=10&params%5Bf%5D%5B%5D=20%3A1"
const uri = "https://game.tsgames.ru/ajax.php?a=gl&params%5Bf%5D="
const uriEnd = "&params%5Bar%5D=1&params%5Ba%5D=3"

type Parser struct {
	Client *http.Client `json:"http_._client"`
}

func NewParser() *Parser {
	return &Parser{Client: &http.Client{}}
}

func (p *Parser) ParseRows() (error, dtos.ReplayRow) {
	res, err := p.Client.Get(tsgReplays)
	if err != nil {
		return err, dtos.ReplayRow{}
	}
	defer res.Body.Close()
	message, _ := io.ReadAll(res.Body)
	repl := dtos.ReplayRow{}
	err = json.Unmarshal(message, &repl)
	if err != nil {
		return err, dtos.ReplayRow{}
	}

	return nil, repl
}

func (p *Parser) ParseReplay(replayName string) (error, [][]interface{}) {

	ss := uri + replayName + uriEnd
	res, err := p.Client.Get(ss)

	if err != nil {
		return err, nil
	}

	replay, _ := io.ReadAll(res.Body)

	str := string(replay)

	fmt.Println(str)
	answer := dtos.Answer{}
	err = json.Unmarshal(replay, &answer)

	replst := make([][]interface{}, 100)
	err = json.Unmarshal([]byte(answer.Json), &replst)
	if err != nil {
		return err, nil
	}

	return nil, replst
}
