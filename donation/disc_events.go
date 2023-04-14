package donation

import (
	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/donationBot/discord"
)

func discEvents() {
	discord.MemberJoin.Add(func(s *discordgo.Session, v *discordgo.GuildMemberAdd) bool {
		v.Roles = setDonationRoles(s, c, v.GuildID, v.User.ID, v.Roles)

		return false
	})
}
