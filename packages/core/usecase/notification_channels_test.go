package usecase

import "testing"

func TestCanonicalChannelType(t *testing.T) {
	if got := CanonicalChannelType("whatsapp"); got != "whatsapp_cloud" {
		t.Fatalf("whatsapp -> %s", got)
	}
	if got := CanonicalChannelType("whatsapp_cloud"); got != "whatsapp_cloud" {
		t.Fatalf("whatsapp_cloud -> %s", got)
	}
	if got := CanonicalChannelType("email"); got != "email_ses" {
		t.Fatalf("email -> %s", got)
	}
	if got := CanonicalChannelType("email_ses"); got != "email_ses" {
		t.Fatalf("email_ses -> %s", got)
	}
	if got := PublicChannelType("whatsapp_cloud"); got != "whatsapp" {
		t.Fatalf("public whatsapp -> %s", got)
	}
	if got := PublicChannelType("email_ses"); got != "email" {
		t.Fatalf("public email -> %s", got)
	}
	names := ChannelLookupNames("whatsapp")
	if len(names) != 2 || names[0] != "whatsapp_cloud" || names[1] != "whatsapp" {
		t.Fatalf("lookup names: %#v", names)
	}
	names = ChannelLookupNames("whatsapp_cloud")
	if len(names) != 2 || names[0] != "whatsapp_cloud" || names[1] != "whatsapp" {
		t.Fatalf("canonical lookup names: %#v", names)
	}
}
