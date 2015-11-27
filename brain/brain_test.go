package brain

import "testing"

func TestBrain(t *testing.T) {
	brain := New()
	brain.Save("fake", "key", 1)
	if len(brain.items) != 1 {
		t.Error("error storing values in Brain. Expected 1 item. Got:", len(brain.items))
	}

	realKey := fullKeyName("fake", "key")
	if brain.items[realKey].(int) != 1 {
		t.Error("corruption of information stored in Brain. Expected 1 item. Got:", brain.items[realKey])
	}

	impossibleValue := brain.Read("invalid", "key")
	if impossibleValue != nil {
		t.Error("corruption of information stored in Brain. Expected nil. Got:", impossibleValue)
	}
}
