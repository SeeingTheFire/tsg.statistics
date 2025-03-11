package services

import (
	"encoding/json"
	"errors"
	"github.com/SeeingTheFire/tsg.statistics/internal/dtos"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"strconv"
)

const tsgReplays = "https://game.tsgames.ru/ajax.php?a=l&params%5Bf%5D%5B%5D=2&params%5Bf%5D%5B%5D=3&params%5Bf%5D%5B%5D=4&params%5Bf%5D%5B%5D=10&params%5Bf%5D%5B%5D=20%3A1"
const uri = "https://game.tsgames.ru/ajax.php?a=gl&params%5Bf%5D="
const uriEnd = "&params%5Bar%5D=1&params%5Ba%5D=3"

type Parser struct {
	Client *http.Client   `json:"-"`
	Logger *logrus.Logger `json:"-"`
	Ctx    *context.Context
}

func NewParser(logger *logrus.Logger, c *context.Context) *Parser {
	return &Parser{Client: &http.Client{},
		Logger: logger, Ctx: c}
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
	answer := dtos.Answer{}
	err = json.Unmarshal(replay, &answer)
	replayInfo := make([][]interface{}, 100)
	err = json.Unmarshal([]byte(answer.Json), &replayInfo)
	if err != nil {
		return err, nil
	}

	return nil, replayInfo
}

func (p *Parser) ParseReplayInfo(eventsInfo [][]interface{}) (error, *dtos.GameInfo) {
	bots := make(map[int]int, 250)
	vehicles := make(map[int]*dtos.Vehicle, 50)
	players := make(map[int]*dtos.Player, 250)
	if len(eventsInfo) < 2 {
		return errors.New("Not enough events"), &dtos.GameInfo{}
	}

	gameInfo := dtos.GameInfo{}
	botInfoSlice := eventsInfo[1][1].([]interface{})

	for _, botInfoInterface := range botInfoSlice {
		botInfo := botInfoInterface.([]interface{})
		if botInfo != nil {
			eventId := botInfo[0].(float64)
			switch int(eventId) {
			case 1:
				// Добавления коллекции бота
				err := p.parseBot(botInfo, &bots)
				if err != nil {
					continue
				}
			case 2:
				// Добавление техники используемой на миссии
				err := p.parseVehicle(botInfo, &vehicles)
				if err != nil {
					continue
				}
			case 3:
				// Добавление игроков из реплея
				err := p.parsePlayer(botInfo, &bots, &players)
				if err != nil {
					continue
				}
			}
		}
	}

	for i := 2; i < len(eventsInfo); i++ {
		eventCollection := eventsInfo[i][1].([]interface{})
		if len(eventCollection) < 2 {
			i++
			continue
		}

		if eventCollection[i] == nil {
			i++
			continue
		}

		event := eventCollection[i].([]interface{})
		eventId := event[0].(float64)
		switch int(eventId) {
		case 3:
			// Перезаход?

		case 4:
			// Ивент отвечающий за событие убийства
			err := p.GetKillInfo(eventCollection, &players, &vehicles)
			if err != nil {
				p.Logger.Warn(err)
			}
		case 5:
			// Ивент отвечающий за событие нанесения урона без убийства
			err := p.getDamageInfo()
			if err != nil {
				p.Logger.Warn(err)
			}
		case 7:
			// Ивент отвечающий за событие лечения
			err := p.getHealInfo()
			if err != nil {
				p.Logger.Warn(err)
			}
		}

		i++
	}

	return nil, &gameInfo
}

func (p *Parser) getTag(playerName string) (err error, tag string, cadet string, name string) {
	//match, err := regexp.Match(playerName, "ss")
	//if err != nil {
	//	return err, "", "", ""
	//}

	return nil, "", "", ""
}

func (p *Parser) parseBot(botInfo []interface{}, bots *map[int]int) error {
	botId := botInfo[1].(float64)
	side := botInfo[4].(float64)
	(*bots)[int(botId)] = int(side)

	return nil
}
func (p *Parser) parseVehicle(botInfo []interface{}, vehicles *map[int]*dtos.Vehicle) error {
	vehicleId := botInfo[1].(float64)
	name := botInfo[2].(string)
	class := botInfo[3].(string)
	(*vehicles)[int(vehicleId)] = &dtos.Vehicle{Name: name, Class: class}

	return nil
}
func (p *Parser) parsePlayer(botInfo []interface{}, bots *map[int]int, players *map[int]*dtos.Player) error {
	botId := botInfo[1].(float64)
	side := (*bots)[int(botId)]
	playerName := botInfo[3].(string)
	err, tag, cadet, name := p.getTag(playerName)
	if err != nil {
		return err
	}
	steamId, err := strconv.Atoi(botInfo[4].(string))
	if err != nil {
		return err
	}
	if steamId == 0 {
		return err
	}

	(*players)[int(botId)] = &dtos.Player{Name: name, Tag: tag,
		Side: side, Cadet: cadet, SteamId: int64(steamId)}

	return nil
}

func (p *Parser) GetKillInfo(eventInfo []interface{}, players *map[int]*dtos.Player, vehicles *map[int]*dtos.Vehicle) error {
	killedPlayer := (*players)[int(eventInfo[3].(float64))]
	murderPlayer := (*players)[int(eventInfo[2].(float64))]
	var vehicle *dtos.Vehicle
	var isTeamKilled bool

	if murderPlayer == nil {
		return errors.New("murder player not found")
	}

	if killedPlayer == nil {
		vehicle = (*vehicles)[int(eventInfo[3].(float64))]
		if vehicle == nil {
			return errors.New("killed player not found")
		}
	} else {
		isTeamKilled = killedPlayer.Side == murderPlayer.Side
	}

	weapon := eventInfo[4].(string)
	distance := eventInfo[6].(float64)
	killInfo := dtos.KillInfo{Murder: murderPlayer, Killed: killedPlayer,
		IsTeamKill: isTeamKilled, Weapon: weapon, Distance: float32(distance),
	}

	if killedPlayer != nil {
		killedPlayer.Kills = append(killedPlayer.Kills, killInfo)
	}

	murderPlayer.Kills = append(murderPlayer.Kills, killInfo)
	return nil
}

func (p *Parser) getDamageInfo(eventInfo []interface{}, players *map[int]*dtos.Player) error {

	return nil
}

func (p *Parser) getHealInfo(eventInfo []interface{}, players *map[int]*dtos.Player) error {
	return nil
}
