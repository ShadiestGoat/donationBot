package config

import (
	"strconv"
	"strings"

	"github.com/shadiestgoat/log"
)

// Roles. min:max:role_id|min2:max2:role_id_2
// For donations. The min is non inclusive, so if min == 0, then non donors are not accepted into this category
// If max == -1, then there is no upper limit to it.
func parseDonationRole(inp string) any {
	roles := []*DonationRole{}
	rolesOverall := strings.Split(inp, "|")

	for _, r := range rolesOverall {
		raw := strings.Split(r, ":")
		if len(raw) != 3 {
			log.Warn("Couldn't parse '%v': format should be 'min:max:role_id'", raw)
			continue
		}
		min, err := strconv.ParseFloat(raw[0], 64)
		if err != nil {
			log.Warn("Couldn't parse '%v': invalid float", err)
			continue
		}
		max, err := strconv.ParseFloat(raw[1], 64)
		if err != nil {
			log.Warn("Couldn't parse '%v': invalid float", err)
			continue
		}

		roleID := raw[2]

		if (min < 0 || max < 0) && max != -1 {
			log.Warn("Role '%s' has min or max < 0, which is illegal!", roleID)
			continue
		}

		if max != -1 && min < max {
			log.Warn("Role '%s' min < max, which is very illegal (you're going to jail btw)!", roleID)
			continue
		}

		roles = append(roles, &DonationRole{
			Min:          min,
			Max:          max,
			RoleID:       roleID,
		})
	}

	return roles
}

type DonationRole struct {
	Min          float64
	Max          float64
	RoleID       string
}
