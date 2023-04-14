package discutils

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/log"
)

func GetMember(s *discordgo.Session, guildID, userID string) *discordgo.Member {
	m, err := s.State.Member(guildID, userID)
	if err == nil && m != nil {
		return m
	}
	m, err = s.GuildMember(guildID, userID)
	if err != nil {
		log.ErrorIfErr(err, "fetching member '%s' in guild '%s'", userID, guildID)
		m = nil
	}

	return m
}

