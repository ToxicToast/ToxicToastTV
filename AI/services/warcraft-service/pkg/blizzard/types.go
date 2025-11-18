package blizzard

// CharacterProfile contains the full character data from Blizzard API
type CharacterProfile struct {
	// Character API fields
	Name   string
	Realm  string
	Region string

	// Character Details from API
	DisplayName       string
	DisplayRealm      string
	Level             int
	ItemLevel         int
	ClassName         string
	ClassID           int
	RaceName          string
	RaceID            int
	FactionType       string
	GuildName         *string
	GuildRealm        *string
	ThumbnailURL      *string
	AchievementPoints int
}

// GuildProfile contains the full guild data from Blizzard API
type GuildProfile struct {
	// Guild API fields
	Name   string
	Realm  string
	Region string

	// Guild Details from API
	FactionType       string
	MemberCount       int
	AchievementPoints int
	Crest             string
}
