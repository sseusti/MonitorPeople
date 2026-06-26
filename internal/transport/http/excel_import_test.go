package httpapi

import "testing"

func TestStudentFIOFromRowPrefersCyrillicName(t *testing.T) {
	fio := studentFIOFromRow([]string{
		"0",
		"ABASOVA LEISAN TEMIRLANOVNA",
		"Абасова Лейсан Темирлановна",
	}, 1, 2)

	if fio != "Абасова Лейсан Темирлановна" {
		t.Fatalf("expected cyrillic FIO, got %q", fio)
	}
}
