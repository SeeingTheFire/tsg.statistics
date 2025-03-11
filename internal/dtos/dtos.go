package dtos

type ReplayRow struct {
	Rows   []Row  `json:"rows"`
	Total  int    `json:"total"`
	Source string `json:"source"`
	Error  string `json:"error"`
}

type Row struct {
	Name     string   `json:"name"`
	Archive  int      `json:"archive"`
	FileSize int      `json:"fileSize"`
	Array    []string `json:"array"`
}

type Answer struct {
	Json  string `json:"json"`
	Error string `json:"error"`
}

type GameInfo struct {
	Name   string           `json:"name"`
	Date   int              `json:"time"`
	Map    string           `json:"map"`
	GameId string           `json:"gameId"`
	Server string           `json:"server"`
	Squads map[string]Squad `json:"squads"`
}

type Squad struct {
	Tag     string            `json:"tag"`
	Players map[string]Player `json:"players"`
}

type Player struct {
	Name    string     `json:"name"`
	Tag     string     `json:"tag"`
	Kills   []KillInfo `json:"kills"`
	Death   []KillInfo `json:"death"`
	Side    int        `json:"side"`
	Cadet   string     `json:"cadet"`
	SteamId int64      `json:"steamId"`
}

type KillInfo struct {
	Murder        *Player `json:"murder"`
	Killed        *Player `json:"killed"`
	IsTeamKill    bool    `json:"isTeamKill"`
	Weapon        string  `json:"weapon"`
	Time          int     `json:"time"`
	Distance      float32 `json:"distance"`
	IsVehicleKill bool    `json:"isVehicleKill"`
	VehicleName   string  `json:"vehicleName"`
}
type DamageInfo struct {
	Damage            int     `json:"damage"`
	Weapon            string  `json:"weapon"`
	BulletType        string  `json:"bulletType"`
	Time              int     `json:"time"`
	Distance          float32 `json:"distance"`
	DamageToSteamId   int64   `json:"killerSteamId"`
	DamageFromSteamId int64   `json:"killedSteamId"`
}
type Vehicle struct {
	Name  string `json:"name"`
	Class string `json:"class"`
}
