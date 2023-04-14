package discord

import (
	"encoding/json"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/donationBot/discutils"
	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/donation-api-wrapper"
)

var s *discordgo.Session

func Init(donClient *donations.Client) *discordgo.Session {
	RegisterComponent("cancel", func(s *discordgo.Session, i *discordgo.InteractionCreate, d *discordgo.MessageComponentInteractionData) {
		emb := discutils.BaseEmbed

		emb.Title = "Cancelled"
		emb.Description = "This operation was cancelled!"
		emb.Color = discutils.COLOR_DANGER

		discutils.IResp(s, i.Interaction, &discutils.IRespOpts{
			Embed:   &discordgo.MessageEmbed{},
			Comps:   []discordgo.MessageComponent{},
			Content: new(string),
		}, discutils.I_UPDATE)
	})

	discToken, err := donClient.DiscordToken()
	log.FatalIfErr(err, "Fetching discord token")

	s, err = discordgo.New("Bot " + discToken)
	log.FatalIfErr(err, "Creating a discord session")

	s.Identify.Intents = discordgo.IntentGuildPresences |
		discordgo.IntentGuilds |
		discordgo.IntentGuildVoiceStates |
		discordgo.IntentGuildMessages |
		discordgo.IntentGuildPresences |
		discordgo.IntentMessageContent |
		discordgo.IntentGuildMembers

	s.AddHandler(handleInteraction)
	s.AddHandler(MessageReactionAdd.Handle)
	s.AddHandler(MemberJoin.Handle)
	s.AddHandler(MessageReactionRemove.Handle)
	s.AddHandler(MessageCreate.Handle)
	s.AddHandler(MessageRemove.Handle)

	Ready.Add(func(s *discordgo.Session, e *discordgo.Ready) bool {
		log.Success("Discord connected as '%s'", e.User.Username)
		return false
	})

	Ready.Add(func(s *discordgo.Session, v *discordgo.Ready) bool {
		f, err := os.ReadFile("status.json")

		if err != nil {
			log.Warn("Couldn't read the status.json file! As such, we are not setting a status!")
			return false
		}

		usp := discordgo.UpdateStatusData{}

		err = json.Unmarshal(f, &usp)

		if err != nil {
			log.Warn("Couldn't parse the status.json file! As such, we are not setting a status!")
			return false
		}

		err = s.UpdateStatusComplex(usp)

		log.ErrorIfErr(err, "setting a status for the bot")

		return false
	})

	s.AddHandlerOnce(Ready.Handle)

	s.StateEnabled = true
	s.State.MaxMessageCount = 150

	err = s.Open()
	log.FatalIfErr(err, "Opening the Discord connection")

	appID := s.State.User.ID

	curCommands, err := s.ApplicationCommands(appID, "")
	log.FatalIfErr(err, "fetching application commands")

	for _, cmd := range curCommands {
		if _, ok := commands[cmd.Name]; !ok {
			err := s.ApplicationCommandDelete(appID, "", cmd.ID)
			if log.ErrorIfErr(err, "deleting command '%s' (id: '%s')", cmd.Name, cmd.ID) {
				log.Warn("This means that the command '%s' will still be there, but it has no handler!", cmd.Name)
			} else {
				log.Debug("Removed command '%s' as non existent", cmd.Name)
			}
		}
	}

	log.Debug("Deprecated commands removed")

	for _, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		log.FatalIfErr(err, "creating/updating command '%s' (id: '%s')", v.Name, v.ID)
		log.Debug("Uploaded command '%s'", cmd.Name)

		// Update with an ID, etc
		commands[cmd.Name] = cmd
	}

	return s
}

func Close() {
	s.Close()
}
