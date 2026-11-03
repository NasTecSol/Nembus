package usecase

import "strings"

// CanonicalChannelType maps UI and dispatcher names onto the stored channel.
// Settings save "whatsapp" and "email"; the sender looks up "whatsapp_cloud" and "email_ses".
func CanonicalChannelType(channelType string) string {
	switch strings.ToLower(strings.TrimSpace(channelType)) {
	case "whatsapp", "whatsapp_cloud":
		return "whatsapp_cloud"
	case "email", "email_ses":
		return "email_ses"
	default:
		return strings.TrimSpace(channelType)
	}
}

// PublicChannelType is the name the settings UI compares against.
func PublicChannelType(channelType string) string {
	switch CanonicalChannelType(channelType) {
	case "whatsapp_cloud":
		return "whatsapp"
	case "email_ses":
		return "email"
	default:
		return strings.TrimSpace(channelType)
	}
}

// ChannelLookupNames returns the stored name first, then the settings-screen alias.
func ChannelLookupNames(channelType string) []string {
	switch CanonicalChannelType(channelType) {
	case "whatsapp_cloud":
		return []string{"whatsapp_cloud", "whatsapp"}
	case "email_ses":
		return []string{"email_ses", "email"}
	default:
		name := strings.TrimSpace(channelType)
		if name == "" {
			return nil
		}
		return []string{name}
	}
}
