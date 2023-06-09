package discutils

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/shadiestgoat/donationBot/utils"
	"github.com/shadiestgoat/log"
)

func ParseCommand(d discordgo.ApplicationCommandInteractionData) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	return parseCmd(d.Options, 1)
}

func parseCmd(d []*discordgo.ApplicationCommandInteractionDataOption, layer int) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	r := map[string]*discordgo.ApplicationCommandInteractionDataOption{}

	for _, opt := range d {
		switch opt.Type {
		case discordgo.ApplicationCommandOptionSubCommandGroup, discordgo.ApplicationCommandOptionSubCommand:
			tmp := parseCmd(opt.Options, layer+1)
			for key, val := range tmp {
				if r[key] != nil {
					panic(key + " is not unique!")
				}
				r[key] = val
			}
			r[fmt.Sprintf("cmd-%d", layer)] = opt
		default:
			r[opt.Name] = opt
		}
	}

	return r
}

func ParseModal(d discordgo.ModalSubmitInteractionData) map[string]string {
	m := map[string]string{}

	for _, row := range d.Components {
		if row.Type() != discordgo.ActionsRowComponent {
			log.Error("Non-ActionsRowComponent ???")
			continue
		}
		r := row.(*discordgo.ActionsRow)
		for _, inp := range r.Components {
			if inp.Type() != discordgo.TextInputComponent {
				continue
			}
			c := inp.(*discordgo.TextInput)
			m[c.CustomID] = c.Value
		}
	}

	return m
}

func errEmbed(err string) *discordgo.MessageEmbed {
	emb := BaseEmbed
	emb.Color = COLOR_DANGER
	emb.Title = "Error!"
	emb.Description = err

	return &emb
}

// WARNING: IS_EPHEMERAL is reversed in here! That is, by default (no opts) it has IS_EPHEMERAL, but, if IS_EPHEMERAL is present, it will make it _not_ ephemeral!
func IError(s *discordgo.Session, i *discordgo.Interaction, err string, opts ...InteractionOpt) {
	parsedOpts := I_NONE
	for _, o := range opts {
		parsedOpts |= o
	}

	IErrorComponents(s, i, err, []discordgo.MessageComponent{}, parsedOpts)
}

// Generate an error response but keep content and ephemeral status.
// Should only be used on buttons!
func IErrorBtn(s *discordgo.Session, i *discordgo.Interaction, err string, keepComps bool) {
	comps := []discordgo.MessageComponent{}
	if keepComps {
		comps = i.Message.Components
	}

	opts := I_UPDATE

	if utils.BitMask(i.Message.Flags, discordgo.MessageFlagsEphemeral) {
		opts |= I_EPHEMERAL
	}

	IResp(s, i, &IRespOpts{
		Embed:   errEmbed(err),
		Comps:   comps,
		Content: &i.Message.Content,
	}, opts)
}

// Same as IError() but allows for components
//
// WARNING: IS_EPHEMERAL is reversed in here! That is, by default (no opts) it has IS_EPHEMERAL, but, if IS_EPHEMERAL is present, it will make it _not_ ephemeral!
// Can auto resolved comps into Action Row (so you can directly pass buttons)
func IErrorComponents(s *discordgo.Session, i *discordgo.Interaction, err string, comps []discordgo.MessageComponent, opts InteractionOpt) {

	opts ^= I_EPHEMERAL

	IResp(s, i, &IRespOpts{
		Embed: errEmbed(err),
		Comps: comps,
	}, opts)

	msg, errGiven := s.InteractionResponse(i)
	if log.ErrorIfErr(errGiven, "fetching interaction") {
		return
	}

	if !utils.BitMask(opts, I_EPHEMERAL) {
		errGiven = s.MessageReactionAdd(msg.ChannelID, msg.ID, EMOJI_CROSS)
		log.ErrorIfErr(errGiven, "adding reaction for '%s'", msg.ID)
	}
}

type InteractionOpt int

const (
	I_NONE      InteractionOpt = 0
	I_EPHEMERAL InteractionOpt = 1 << iota
	I_UPDATE
	I_DEFERRED
)

func IDefer(s *discordgo.Session, i *discordgo.Interaction, opts InteractionOpt) {
	t := discordgo.InteractionResponseDeferredChannelMessageWithSource
	var flags discordgo.MessageFlags

	if utils.BitMask(opts, I_UPDATE) {
		t = discordgo.InteractionResponseDeferredMessageUpdate
	}
	if utils.BitMask(opts, I_EPHEMERAL) {
		flags = discordgo.MessageFlagsEphemeral
	}

	err := s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: t,
		Data: &discordgo.InteractionResponseData{
			Flags: flags,
		},
	})

	log.ErrorIfErr(err, "deferring reaction")
}

// Note: This removes components
func IEmbed(s *discordgo.Session, i *discordgo.Interaction, emb *discordgo.MessageEmbed, opts InteractionOpt) {
	var flags discordgo.MessageFlags
	t := discordgo.InteractionResponseChannelMessageWithSource

	if utils.BitMask(opts, I_DEFERRED) {
		_, err := s.InteractionResponseEdit(i, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{
				emb,
			},
			Components: &[]discordgo.MessageComponent{},
		})
		log.ErrorIfErr(err, "sending deferred resp")
		return
	}

	if utils.BitMask(opts, I_UPDATE) {
		t = discordgo.InteractionResponseUpdateMessage
	}
	if utils.BitMask(opts, I_EPHEMERAL) {
		flags = discordgo.MessageFlagsEphemeral
	}

	err := s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: t,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				emb,
			},
			Components: []discordgo.MessageComponent{},
			Flags:      flags,
		},
	})

	log.ErrorIfErr(err, "creating interaction")
}

type IRespOpts struct {
	Embed  *discordgo.MessageEmbed
	Embeds []*discordgo.MessageEmbed

	Comps   []discordgo.MessageComponent
	Content *string
}

// Can auto resolved comps into Action Row (so you can directly pass buttons)
func IResp(s *discordgo.Session, i *discordgo.Interaction, conf *IRespOpts, opts InteractionOpt) {
	var flags discordgo.MessageFlags
	t := discordgo.InteractionResponseChannelMessageWithSource

	embeds := []*discordgo.MessageEmbed{}

	if len(conf.Embeds) != 0 {
		embeds = conf.Embeds
	}

	if conf.Embed != nil {
		embeds = append(embeds, conf.Embed)
	}

	components := conf.Comps

	if len(components) > 0 {
		_, ok1 := components[0].(discordgo.ActionsRow)
		_, ok2 := components[0].(*discordgo.ActionsRow)

		if !ok1 || !ok2 {
			components = []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: components,
				},
			}
		}
	}

	if utils.BitMask(opts, I_DEFERRED) {
		_, err := s.InteractionResponseEdit(i, &discordgo.WebhookEdit{
			Embeds:     &embeds,
			Components: &components,
			Content:    conf.Content,
		})

		log.ErrorIfErr(err, "sending deferred resp")
		return
	}

	if utils.BitMask(opts, I_UPDATE) {
		t = discordgo.InteractionResponseUpdateMessage
	}
	if utils.BitMask(opts, I_EPHEMERAL) {
		flags = discordgo.MessageFlagsEphemeral
	}

	content := ""
	if conf.Content != nil {
		content = *conf.Content
	}

	err := s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: t,
		Data: &discordgo.InteractionResponseData{
			Embeds:     embeds,
			Content:    content,
			Components: components,
			Flags:      flags,
		},
	})

	log.ErrorIfErr(err, "creating interaction w/ components")
}
