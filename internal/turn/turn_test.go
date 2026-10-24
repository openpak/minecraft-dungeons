package turn

import (
	"testing"
	"time"
)

func TestMintValid(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	c := Mint("s3cret", "dungeons-1", time.Hour, []string{"turn:x:3478"}, now)
	if c.Username != "1800003600:dungeons-1" {
		t.Fatalf("username %q", c.Username)
	}
	if !Valid("s3cret", c.Username, c.Password, now) {
		t.Fatal("own credential rejected")
	}
	if Valid("other", c.Username, c.Password, now) || Valid("s3cret", c.Username, c.Password, now.Add(2*time.Hour)) {
		t.Fatal("wrong secret or expired credential accepted")
	}
}
