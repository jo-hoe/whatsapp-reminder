package whatsapp

import "testing"

func TestLinkCreator_CreateWhatsappLink(t *testing.T) {
	type args struct {
		phoneNumber string
		messageText string
	}
	tests := []struct {
		name     string
		args     args
		wantLink string
	}{
		{
			name:     "Default",
			wantLink: "https://wa.me/15551234567?text=I%27m%20interested%20in%20your%20car%20for%20sale",
			args: args{
				phoneNumber: "15551234567",
				messageText: "I'm interested in your car for sale",
			},
		}, {
			name:     "Empty Phone Number",
			wantLink: "https://wa.me/?text=I%27m%20interested%20in%20your%20car%20for%20sale",
			args: args{
				phoneNumber: "",
				messageText: "I'm interested in your car for sale",
			},
		}, {
			name:     "Empty Phone Text",
			wantLink: "https://wa.me/15551234567",
			args: args{
				phoneNumber: "15551234567",
				messageText: "",
			},
		}, {
			name:     "Nil Case",
			wantLink: "https://wa.me/",
			args: args{
				phoneNumber: "",
				messageText: "",
			},
		}, {
			name:     "Emoji Case",
			wantLink: "https://wa.me/?text=Hallo%20%F0%9F%98%89",
			args: args{
				phoneNumber: "",
				messageText: "Hallo 😉",
			},
		}, {
			name:     "Whitespace in phone number",
			wantLink: "https://wa.me/+49123456789?text=test",
			args: args{
				phoneNumber: " +49\t123\r456\n78 9 ",
				messageText: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLink := CreateWhatsappLink(tt.args.phoneNumber, tt.args.messageText)
			if gotLink != tt.wantLink {
				t.Errorf("CreateWhatsappLink() = %v, want %v", gotLink, tt.wantLink)
			}
		})
	}
}
