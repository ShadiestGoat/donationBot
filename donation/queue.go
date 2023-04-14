package donation

import (
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	donations "github.com/shadiestgoat/donation-api-wrapper"
	"github.com/shadiestgoat/donationBot/discutils"
	"github.com/shadiestgoat/donationBot/utils"
	"github.com/shadiestgoat/log"
)

type DonationQueueStore struct {
	*sync.Mutex
	Queue []string
}

var donationQueue = &DonationQueueStore{
	Mutex: &sync.Mutex{},
	// Format: {guildID}-{userID}
	Queue: []string{},
}

func (s *DonationQueueStore) Add(guildID, userID string) {
	s.Lock()
	defer s.Unlock()

	if utils.BinarySearch(s.Queue, guildID+"-"+userID) != -1 {
		return
	}

	s.Queue = append(s.Queue, userID)
}

func (store *DonationQueueStore) Loop(s *discordgo.Session, c *donations.Client) {
	for {
		log.Debug("Looking through donation loop")
		store.Lock()

		for _, rawInfo := range store.Queue {
			spl := strings.Split(rawInfo, "-")

			guildID := spl[0]
			userID := spl[1]

			m := discutils.GetMember(s, guildID, userID)
			if m == nil {
				continue
			}

			go setDonationRoles(s, c, guildID, m.User.ID, m.Roles)
		}

		store.Unlock()
		time.Sleep(12 * time.Hour)
	}
}
