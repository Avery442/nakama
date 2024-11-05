package server

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/server/evr"
)

type TeamMetadata struct {
	Strength float64 `json:"strength,omitempty"`
}

type slotReservation struct {
	Entrant *EvrMatchPresence
	Expiry  time.Time
}

type MatchLabel struct {
	ID          MatchID          `json:"id"`                    // The Session Id used by EVR (the same as match id)
	Open        bool             `json:"open"`                  // Whether the lobby is open to new players (Matching Only)
	LobbyType   LobbyType        `json:"lobby_type"`            // The type of lobby (Public, Private, Unassigned) (EVR)
	Broadcaster MatchBroadcaster `json:"broadcaster,omitempty"` // The broadcaster's data
	StartTime   time.Time        `json:"start_time,omitempty"`  // The time the match was, or will be started.
	CreatedAt   time.Time        `json:"created_at,omitempty"`  // The time the match was created.
	SpawnedBy   string           `json:"spawned_by,omitempty"`  // The userId of the player that spawned this match.
	GroupID     *uuid.UUID       `json:"group_id,omitempty"`    // The channel id of the broadcaster. (EVR)
	GuildID     string           `json:"guild_id,omitempty"`    // The guild id of the broadcaster. (EVR)
	GuildName   string           `json:"guild_name,omitempty"`  // The guild name of the broadcaster. (EVR)

	Mode             evr.Symbol                `json:"mode,omitempty"`             // The mode of the lobby (Arena, Combat, Social, etc.) (EVR)
	Level            evr.Symbol                `json:"level,omitempty"`            // The level to play on (EVR).
	SessionSettings  *evr.LobbySessionSettings `json:"session_settings,omitempty"` // The session settings for the match (EVR).
	RequiredFeatures []string                  `json:"features,omitempty"`         // The required features for the match. map[feature][hmdtype]isRequired

	MaxSize     int       `json:"limit,omitempty"`        // The total lobby size limit (players + specs)
	Size        int       `json:"size"`                   // The number of players (including spectators) in the match.
	PlayerCount int       `json:"player_count"`           // The number of participants (not including spectators) in the match.
	PlayerLimit int       `json:"player_limit,omitempty"` // The number of players in the match (not including spectators).
	TeamSize    int       `json:"team_size,omitempty"`    // The size of each team in arena/combat (either 4 or 5)
	TeamIndex   TeamIndex `json:"team,omitempty"`         // What team index a player prefers (Used by Matching only)

	TeamMetadata   []TeamMetadata `json:"team_metadata,omitempty"`   // The metadata of the teams in the match.
	Players        []PlayerInfo   `json:"players,omitempty"`         // The displayNames of the players (by team name) in the match.
	TeamAlignments map[string]int `json:"team_alignments,omitempty"` // map[userID]TeamIndex

	GameState *GameState `json:"game_state,omitempty"` // The game state for the match.

	server         runtime.Presence             // The broadcaster's presence
	levelLoaded    bool                         // Whether the server has been sent the start instruction.
	presenceMap    map[string]*EvrMatchPresence // [sessionId]EvrMatchPresence
	reservationMap map[string]*slotReservation  // map[sessionID]slotReservation

	joinTimestamps       map[string]time.Time // The timestamps of when players joined the match. map[sessionId]time.Time
	joinTimeMilliseconds map[string]int64     // The round clock time of when players joined the match. map[sessionId]time.Time
	sessionStartExpiry   int64                // The tick count at which the match will be shut down if it has not started.
	tickRate             int64                // The number of ticks per second.
	emptyTicks           int64                // The number of ticks the match has been empty.
	terminateTick        int64                // The tick count at which the match will be shut down.
}

func (s *MatchLabel) GetPlayerCount() int {
	count := 0
	for _, p := range s.Players {
		if int(p.Team) != evr.TeamSpectator && int(p.Team) != evr.TeamModerator {
			count++
		}
	}
	return count
}

func (s *MatchLabel) GetPlayerByEvrID(evrID evr.EvrId) *PlayerInfo {
	for _, p := range s.Players {
		if p.EvrID == evrID {
			return &p
		}
	}
	return nil
}

func (s *MatchLabel) OpenPlayerSlots() int {
	return s.PlayerLimit - s.GetPlayerCount()
}

func (s *MatchLabel) OpenNonPlayerSlots() int {
	return int(s.MaxSize) - s.PlayerLimit
}

func (s *MatchLabel) OpenSlots() int {
	return int(s.MaxSize) - s.Size
}

func (s *MatchLabel) OpenSlotsByRole(role int) int {
	return s.RoleLimit(role) - s.RoleCount(role)
}

func (s *MatchLabel) String() string {
	return s.GetLabel()
}

func (s *MatchLabel) RoleLimit(role int) int {
	if s.TeamSize == 0 {
		return s.PlayerLimit
	}

	switch role {
	case UnassignedRole:
		return s.TeamSize * 2
	case BlueRole, OrangeRole:
		return s.TeamSize
	case SocialRole:
		return s.PlayerLimit
	case SpectatorRole, ModeratorRole:
		return s.PlayerLimit - s.TeamSize*2
	}

	return s.PlayerLimit
}

func (s *MatchLabel) RoleCount(role int) int {
	count := 0
	for _, p := range s.Players {
		if p.Team == TeamIndex(role) {
			count++
		}
	}
	return count
}

func (s *MatchLabel) Started() bool {
	return !s.StartTime.IsZero() && time.Now().After(s.StartTime)
}

func (s *MatchLabel) GetLabel() string {
	labelJson, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	return string(labelJson)
}

func (s *MatchLabel) GetLabelIndented() string {
	labelJson, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return ""
	}
	return string(labelJson)
}

func (s *MatchLabel) GetGroupID() uuid.UUID {
	if s.GroupID == nil {
		return uuid.Nil
	}
	return *s.GroupID
}

func (s *MatchLabel) GetEndpoint() evr.Endpoint {
	return s.Broadcaster.Endpoint
}

func (s *MatchLabel) GetEntrantConnectMessage(role int, isPCVR bool, disableEncryption bool, disableMAC bool) *evr.LobbySessionSuccessv5 {
	return evr.NewLobbySessionSuccess(s.Mode, s.ID.UUID, s.GetGroupID(), s.Broadcaster.Endpoint, int16(role), isPCVR, disableEncryption, disableMAC).Version5()
}

func (s *MatchLabel) MetricsTags() map[string]string {
	return map[string]string{
		"mode":     s.Mode.String(),
		"level":    s.Level.String(),
		"type":     s.LobbyType.String(),
		"group_id": s.GetGroupID().String(),
	}
}

// rebuildCache is called after the presences map is updated.
func (s *MatchLabel) rebuildCache() {
	presences := make([]*EvrMatchPresence, 0, len(s.presenceMap))
	for _, p := range s.presenceMap {
		presences = append(presences, p)
	}

	// Include the reservations in the cache.
	for id, r := range s.reservationMap {
		if r.Expiry.Before(time.Now()) {
			delete(s.reservationMap, id)
			continue
		}
		// Include the reservation in the cache.
		presences = append(presences, r.Entrant)
	}

	// Rebuild the lookup tables.
	s.Size = len(presences)
	s.Players = make([]PlayerInfo, 0, s.Size)
	s.PlayerCount = 0
	// Construct Player list
	team1 := make(RatedTeam, 0, s.TeamSize)
	team2 := make(RatedTeam, 0, s.TeamSize)

	for _, p := range presences {
		// Do not include spectators or moderators in player count
		if p.RoleAlignment != evr.TeamSpectator && p.RoleAlignment != evr.TeamModerator {
			s.PlayerCount++
		}

		playerinfo := PlayerInfo{
			UserID:        p.UserID.String(),
			Username:      p.Username,
			DisplayName:   p.DisplayName,
			EvrID:         p.EvrID,
			Team:          TeamIndex(p.RoleAlignment),
			ClientIP:      p.ClientIP,
			DiscordID:     p.DiscordID,
			PartyID:       p.PartyID.String(),
			JoinTime:      s.joinTimeMilliseconds[p.SessionID.String()],
			Rating:        p.Rating,
			IsReservation: s.reservationMap[p.SessionID.String()] != nil,
		}

		s.Players = append(s.Players, playerinfo)

		switch s.Mode {
		case evr.ModeArenaPublic:
			switch p.RoleAlignment {
			case BlueRole:
				team1 = append(team1, p.Rating)
			case OrangeRole:
				team2 = append(team2, p.Rating)
			}
		case evr.ModeArenaPrivate, evr.ModeCombatPrivate:
			playerinfo.Team = TeamIndex(UnassignedRole)
		}
	}
	s.TeamMetadata = make([]TeamMetadata, 0, 2)

	if s.Mode == evr.ModeArenaPublic {
		s.TeamMetadata = []TeamMetadata{
			{
				Strength: team1.Strength(),
			},
			{
				Strength: team2.Strength(),
			},
		}
	}

	sort.SliceStable(s.Players, func(i, j int) bool {
		return s.Players[i].Team < s.Players[j].Team
	})
}

func (l *MatchLabel) PublicView() *MatchLabel {
	// Remove private data
	v := &MatchLabel{
		LobbyType:        l.LobbyType,
		ID:               l.ID,
		Open:             l.Open,
		GameState:        l.GameState,
		StartTime:        l.StartTime,
		CreatedAt:        l.CreatedAt,
		GroupID:          l.GroupID,
		GuildID:          l.GuildID,
		SpawnedBy:        l.SpawnedBy,
		Mode:             l.Mode,
		Level:            l.Level,
		RequiredFeatures: l.RequiredFeatures,
		MaxSize:          l.MaxSize,
		Size:             l.Size,
		PlayerCount:      l.PlayerCount,
		PlayerLimit:      l.PlayerLimit,
		TeamSize:         l.TeamSize,
		Broadcaster: MatchBroadcaster{
			OperatorID:  l.Broadcaster.OperatorID,
			GroupIDs:    l.Broadcaster.GroupIDs,
			VersionLock: l.Broadcaster.VersionLock,
			Regions:     l.Broadcaster.Regions,
			Tags:        l.Broadcaster.Tags,
			Features:    l.Broadcaster.Features,
		},
		Players:      make([]PlayerInfo, 0),
		TeamMetadata: l.TeamMetadata,
	}
	if l.LobbyType == PrivateLobby || l.LobbyType == UnassignedLobby {
		// Set the last bytes to FF to hide the ID
		for i := 12; i < 16; i++ {
			v.ID.UUID[i] = 0xFF
		}
	} else {
		for i := range l.Players {
			v.Players = append(v.Players, PlayerInfo{
				UserID:      l.Players[i].UserID,
				Username:    l.Players[i].Username,
				DisplayName: l.Players[i].DisplayName,
				EvrID:       l.Players[i].EvrID,
				Team:        l.Players[i].Team,
				DiscordID:   l.Players[i].DiscordID,
				PartyID:     l.Players[i].PartyID,
				JoinTime:    l.Players[i].JoinTime,
				Rating:      l.Players[i].Rating,
			})
		}

	}
	return v
}
