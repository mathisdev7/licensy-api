package utils

func FormatIconURL(guildID string, iconHash *string) *string {
	if iconHash == nil {
		return nil
	}
	url := "https://cdn.discordapp.com/icons/" + guildID + "/" + *iconHash + ".png"
	return &url
}
