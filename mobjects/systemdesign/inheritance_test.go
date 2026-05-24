package systemdesign

import (
	"testing"

	"github.com/ankitsinghchadda/goanim/core/style"
)

// TestPacketLabelInheritsFontFamily — regression test: Packet.label
// must NOT force FontHandDrawn, which would defeat the scene's
// FontSans default in the Crisp preset.
//
// The general rule this enforces: a composite mobject's child must
// resolve against the scene default for any attribute the child doesn't
// explicitly need to override.
func TestPacketLabelInheritsFontFamily(t *testing.T) {
	pkt := NewPacket(1, "GET /user")

	// Inject a scene default that uses FontSans, as PresetCrisp does.
	ctx := style.NewContext()
	ctx.SceneDefault = style.Style{FontFamily: style.FontSans}

	got := ctx.Resolve(*pkt.label.Style())
	if got.FontFamily != style.FontSans {
		t.Fatalf("packet label FontFamily not inheriting from scene default: got %v, want FontSans", got.FontFamily)
	}
}

// TestClientLabelInheritsStyleFromScene — the same general property
// for the box-based mobjects (Client, Server).
func TestClientLabelInheritsStyleFromScene(t *testing.T) {
	c := NewClient(2, "Client")

	ctx := style.NewContext()
	ctx.SceneDefault = style.Style{
		FontFamily: style.FontSans,
		FontSize:   style.FontLarge,
	}

	got := ctx.Resolve(*c.box.label.Style())
	if got.FontFamily != style.FontSans {
		t.Fatalf("client label FontFamily inheritance broken: got %v, want FontSans", got.FontFamily)
	}
	if got.FontSize != style.FontLarge {
		t.Fatalf("client label FontSize inheritance broken: got %v, want FontLarge", got.FontSize)
	}
}

// TestDatabaseLabelInheritsStyleFromScene — same property for Database.
func TestDatabaseLabelInheritsStyleFromScene(t *testing.T) {
	d := NewDatabase(3, "Postgres")

	ctx := style.NewContext()
	ctx.SceneDefault = style.Style{FontFamily: style.FontSans}

	got := ctx.Resolve(*d.label.Style())
	if got.FontFamily != style.FontSans {
		t.Fatalf("database label FontFamily inheritance broken: got %v, want FontSans", got.FontFamily)
	}
}
