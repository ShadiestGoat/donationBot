package donation

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	donations "github.com/shadiestgoat/donation-api-wrapper"
	"github.com/shadiestgoat/donationBot/discutils"
	"github.com/shadiestgoat/donationBot/utils"
	"github.com/shadiestgoat/log"
)

func embedFund(f *donations.Fund) *discordgo.MessageEmbed {
	emb := discutils.BaseEmbed
	emb.Title = "The '" + f.ShortTitle + "' Fund"
	emb.Description = "[" + f.Title + "](" + c.FundURL(f) + ")"

	if f.Goal != 0 {
		if f.Amount == nil {
			tmpF, err := c.FundByID(f.ID)
			if log.ErrorIfErr(err, "fetching fund for embed <3") {
				return nil
			}

			*f = *tmpF
		}

		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Goal",
			Value:  utils.TextProgressBar(f.Goal, *f.Amount, "0", fmt.Sprint(f.Goal), 40),
			Inline: false,
		})
	}

	return &emb
}
