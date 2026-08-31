package domain

import "errors"

var (
	// ErrProfileNotFound is returned when a learner profile row does not exist.
	ErrProfileNotFound = errors.New("learner profile not found")
)

const (
	AvailabilityLessThan5 = "lt_5"
	Availability5To10     = "5_10"
	Availability10To20    = "10_20"
	AvailabilityOver20    = "gt_20"
)

const (
	ExperienceBeginner     = "beginner"
	ExperienceIntermediate = "intermediate"
	ExperienceAdvanced     = "advanced"
)

const (
	FormatVideo       = "video"
	FormatArticle     = "article"
	FormatInteractive = "interactive"
)

var (
	validAvailability = map[string]bool{
		AvailabilityLessThan5: true, Availability5To10: true,
		Availability10To20: true, AvailabilityOver20: true,
	}
	validExperience = map[string]bool{
		ExperienceBeginner: true, ExperienceIntermediate: true, ExperienceAdvanced: true,
	}
	validFormats = map[string]bool{
		FormatVideo: true, FormatArticle: true, FormatInteractive: true,
	}
)

func ValidAvailability(v string) bool { return validAvailability[v] }

func ValidExperience(v string) bool { return validExperience[v] }

func ValidFormat(v string) bool { return validFormats[v] }

// DailyMinutes maps the UI weekly-hours buckets to a daily learning budget used
// by the daily-task scheduler.
func DailyMinutes(availability string) int {
	switch availability {
	case AvailabilityLessThan5:
		return 30
	case Availability5To10:
		return 60
	case Availability10To20:
		return 120
	case AvailabilityOver20:
		return 180
	default:
		return 60
	}
}

type LearnerProfile struct {
	UserID           string
	TimeAvailability string
	FormatPreference string
	PriorExperience  string
	Gender           string
	AvatarURL        string
	Role             string
}
