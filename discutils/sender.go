package discutils

import "github.com/bwmarrin/discordgo"

func SendMessage(s *discordgo.Session, channelID string, msg *discordgo.MessageSend) (resp *discordgo.Message, err error) {
	if s == nil {
		return nil, nil
	}
	resp, err = s.ChannelMessageSendComplex(channelID, msg)
	if err != nil {
		return
	}
	channel, chanErr := s.State.Channel(channelID)
	if chanErr != nil {
		channel, chanErr = s.Channel(channelID)
	}
	if chanErr != nil || channel == nil || (channel.Type != discordgo.ChannelTypeGuildNews && channel.Type != discordgo.ChannelTypeGuildNewsThread) {
		return
	}

	_, err = s.ChannelMessageCrosspost(channelID, resp.ID)
	return
}
