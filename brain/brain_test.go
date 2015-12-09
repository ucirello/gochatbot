package brain

import "testing"

func TestBrain(t *testing.T) {
	brain := Brain()
	brain.Save("fake", "key", []byte("1"))
	if len(brain.items) != 1 {
		t.Error("error storing values in Brain. Expected 1 item. Got:", len(brain.items))
	}

	realKey := fullKeyName("fake", "key")
	if string(brain.items[realKey]) != "1" {
		t.Error("corruption of information stored in Brain. Expected 1 item. Got:", brain.items[realKey])
	}

	impossibleValue := brain.Read("invalid", "key")
	if string(impossibleValue) != "" {
		t.Error("corruption of information stored in Brain. Expected empty slice. Got:", impossibleValue)
	}
}
