package donation

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	donations "github.com/shadiestgoat/donation-api-wrapper"
	"github.com/shadiestgoat/donationBot/config"
	"github.com/shadiestgoat/donationBot/discord"
	"github.com/shadiestgoat/donationBot/discutils"
	"github.com/shadiestgoat/log"
)

var c *donations.Client

var fundNotFound = ""

var minAliasL = 3

func Register() *donations.Client {
	cmdEditFund()
	cmdAddFund()
	cmdFund()
	cmdDonor()
	discEvents()

	discord.RegisterAutocomplete("fund", autocompleteFunds)
	discord.RegisterAutocomplete("editfund", autocompleteFunds)
	discord.RegisterAutocomplete("donate", autocompleteFunds)

	fundNotFound = "Couldn't find the fund!\nYou [can view all the funds here](https://" + config.LOCATION + "/funds)"

	c = donations.NewClient(config.TOKEN, donations.WithCustomLocation(config.LOCATION))

	return c
}

type guildInfo struct {
	AnnouncementChannel []string
	FundChannels        []string

	RolesPersistent []*config.DonationRole
	RolesMonthly    []*config.DonationRole
}

var guildMap = map[string]*guildInfo{}

func PrepareConfig(s *discordgo.Session) {
	persistent := map[string]*config.DonationRole{}
	monthly := map[string]*config.DonationRole{}

	for _, r := range config.ROLES_PERSISTENT {
		persistent[r.RoleID] = r
	}

	for _, r := range config.ROLES_MONTHLY {
		monthly[r.RoleID] = r
	}

	fundChanMap := map[string]bool{}
	annChanMap := map[string]bool{}

	for _, c := range config.CHANNELS_FUND {
		fundChanMap[c] = true
	}

	for _, c := range config.CHANNELS_ANNOUNCEMENTS {
		annChanMap[c] = true
	}

	for _, guildID := range config.GUILD_IDS {
		info := &guildInfo{
			AnnouncementChannel: []string{},
			FundChannels:        []string{},
			RolesPersistent:     []*config.DonationRole{},
			RolesMonthly:        []*config.DonationRole{},
		}

		guild, err := s.Guild(guildID)
		if log.ErrorIfErr(err, "fetching guild '%s'", guildID) {
			log.Warn("Couldn't fetch the guild '%s', are you sure I am in it?", guildID)
			continue
		}

		for _, role := range guild.Roles {
			if persistent[role.ID] != nil {
				info.RolesPersistent = append(info.RolesPersistent, persistent[role.ID])
			}
			if monthly[role.ID] != nil {
				info.RolesMonthly = append(info.RolesMonthly, monthly[role.ID])
			}
		}

		channels, err := s.GuildChannels(guildID)
		log.FatalIfErr(err, "fetch guild channels for guild '%s'", guildID)

		for _, c := range channels {
			if fundChanMap[c.ID] {
				info.FundChannels = append(info.FundChannels, c.ID)
			}
			if annChanMap[c.ID] {
				info.AnnouncementChannel = append(info.AnnouncementChannel, c.ID)
			}
		}

		guildMap[guildID] = info
	}
}

func Init(Discord *discordgo.Session) {
	c.AddHandler(func(c *donations.Client, v *donations.EventOpen) {
		log.Debug("Opening the donation WS conn")

		after := ""
		errors := 0

		for _, guildID := range config.GUILD_IDS {
			for {
				if errors > 5 {
					log.Fatal("Got too many errors when initiating the donations")
				}

				members, err := Discord.GuildMembers(guildID, after, 100)

				if log.ErrorIfErr(err, "fetching guild members in donation setup") {
					errors++
					time.Sleep(5 * time.Second)
				} else {

					for i, mem := range members {
						setDonationRoles(Discord, c, guildID, mem.User.ID, mem.Roles)
						if i == 99 {
							after = mem.User.ID
						}
					}

					if len(members) != 100 {
						break
					}

					errors = 0
				}
			}
		}

		log.Debug("Finished the guild member donation setup")
	})

	c.AddHandler(func(c *donations.Client, v *donations.EventClose) {
		log.Warn("The donation API had to be shut down due to '%v', restarting in 30s...", v.Err)
		time.Sleep(30 * time.Second)
		c.OpenWS()
	})

	c.AddHandler(func(c *donations.Client, v *donations.EventNewDonation) {
		donor, err := c.DonorByID(v.Donor, false)
		if log.ErrorIfErr(err, "fetching donor '%s'", v.Donor) {
			return
		}

		fund, err := c.FundByID(v.FundID)
		if log.ErrorIfErr(err, "fetching fund '%s'", v.FundID) {
			return
		}

		donorDiscord := ""

		for _, d := range donor.Donors {
			if d.DiscordID != "" {
				donorDiscord = d.DiscordID
				break
			}
		}

		emb := discutils.BaseEmbed
		emb.Title = "New Donation!"

		discordID := ""

		if donorDiscord != "" {
			discordID = donorDiscord
			donorDiscord = "<@" + donorDiscord + ">"
		} else {
			donorDiscord = "Someone"
		}

		emb.Description = fmt.Sprintf("%s donated %.2f Euro for the [%s](%s) fund!", donorDiscord, v.Amount, fund.ShortTitle, c.FundURL(fund))
		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "ID",
			Value:  "`" + v.ID + "`",
			Inline: false,
		})

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Message",
			Value:  v.Message,
			Inline: false,
		})

		for guildID, info := range guildMap {
			if discordID != "" {
				mem := discutils.GetMember(Discord, guildID, discordID)

				if mem == nil {
					emb.Author = nil
				} else {
					setDonationRoles(Discord, c, guildID, discordID, mem.Roles)

					emb.Author = &discordgo.MessageEmbedAuthor{
						Name:    discutils.MemberName(mem),
						IconURL: mem.AvatarURL("256"),
					}
				}
			}

			for _, chanID := range info.AnnouncementChannel {
				discutils.SendMessage(Discord, chanID, &discordgo.MessageSend{
					Embeds: []*discordgo.MessageEmbed{
						&emb,
					},
				})
			}

		}
	})

	c.AddHandler(func(c *donations.Client, v *donations.EventNewFund) {
		goalStr := ""

		if v.Goal != 0 {
			goalStr = " with a goal of " + fmt.Sprint(v.Goal) + " Euros"
		}

		emb := discutils.BaseEmbed
		emb.Title = "Fund '" + v.ShortTitle + "' has been created" + goalStr + "!"
		emb.Description = v.Title + "\n[You can look at it here](" + c.FundURL(v.Fund) + ")"

		for _, info := range guildMap {
			for _, chanID := range info.FundChannels {
				discutils.SendMessage(Discord, chanID, &discordgo.MessageSend{
					Embeds: []*discordgo.MessageEmbed{
						&emb,
					},
				})
			}
		}
	})

	c.OpenWS()

	go donationQueue.Loop(Discord, c)
}
