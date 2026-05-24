package privacy

import (
	"strings"
	"testing"
)

func TestRedactSensitive(t *testing.T) {
	input := "OTP is 123456, UPI PIN: 9876, card 4111 1111 1111 1111"
	got := RedactSensitive(input)

	for _, leaked := range []string{"123456", "9876", "4111 1111 1111 1111"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("expected %q to be redacted from %q", leaked, got)
		}
	}
	for _, marker := range []string{"[REDACTED_OTP]", "[REDACTED_PIN]", "[REDACTED_CARD]"} {
		if !strings.Contains(got, marker) {
			t.Fatalf("expected marker %q in %q", marker, got)
		}
	}
}
