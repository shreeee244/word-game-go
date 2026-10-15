package game

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// TestScore covers the scoring rules, with special attention to duplicate
// letters — the case the brief calls out as "the hard part".
//
// The rule: each letter in the ANSWER is a physical tile that can colour at
// most ONE letter of the guess. Greens claim first; yellows only if a copy of
// that letter is still unclaimed.
func TestScore(t *testing.T) {
	tests := []struct {
		name   string
		guess  string
		answer string
		want   []Mark
	}{
		{
			name:   "all correct",
			guess:  "CRANE",
			answer: "CRANE",
			want:   []Mark{Correct, Correct, Correct, Correct, Correct},
		},
		{
			name:   "all absent",
			guess:  "MOIST",
			answer: "CRANE",
			want:   []Mark{Absent, Absent, Absent, Absent, Absent},
		},
		{
			name:   "mixed present and absent",
			guess:  "SLATE",
			answer: "CRANE",
						// pos2 A==A → Correct
			// pos4 E==E → Correct
			want:   []Mark{Absent, Absent, Correct, Absent, Correct},
		},
		{
			name:   "greens consume the answer's copy before yellows",
			guess:  "ABBEY",
			answer: "BABES",
			// A absent (A not in BABES? yes it is: B A B E S — A at pos1, not pos0 → Present)
			// Recompute carefully: guess A B B E Y vs answer B A B E S
			//   pos0: A vs B → not correct
			//   pos1: B vs A → not correct
			//   pos2: B vs B → Correct
			//   pos3: E vs E → Correct
			//   pos4: Y vs S → not correct
			// remaining after greens (pass1 consumes a[2]=B, a[3]=E; a[0]=B, a[1]=A, a[4]=S remain):
			//   remaining = {B:1, A:1, S:1}  (B at pos0 of answer)
			// pass2:
			//   pos0: A → remaining[A]=1 → Present, remaining[A]=0
			//   pos1: B → remaining[B]=1 → Present, remaining[B]=0
			//   pos4: Y → remaining[Y]=0 → Absent
			want: []Mark{Present, Present, Correct, Correct, Absent},
		},
		 		{
			// THE HARD ONE from the brief.
			// Answer CRANE has exactly one E.  Guess EERIE has three E's.
			// Only one of the guess's three E's is coloured (the green at
			// pos 4).  The other two E's are Absent, NOT Present.
			//
			// The R at pos 2 is Present because the answer's R (pos 1) was
			// never consumed as a green — the guess has R at pos 2, so it
			// matches the answer's unclaimed R.
			name:   "duplicate letters: only one E may be coloured",
			guess:  "EERIE",
			answer: "CRANE",
			want:   []Mark{Absent, Absent, Present, Absent, Correct},
		},
		{
			name:   "duplicate yellows: answer has two L, guess has three",
			guess:  "LLLXX",
			answer: "HELLO",
			// pos2: L vs L → Correct (answer pos2 = L)
			// remaining after pass1: a[0]=H, a[1]=E, a[3]=L, a[4]=O → {H:1,E:1,L:1,O:1}
			// pass2:
			//   pos0: L → remaining[L]=1 → Present, remaining[L]=0
			//   pos1: L → remaining[L]=0 → Absent
			//   pos3: X → Absent
			//   pos4: X → Absent
			want: []Mark{Present, Absent, Correct, Absent, Absent},
		},
		{
			name:   "lowercase inputs are uppercased",
			guess:  "crane",
			answer: "CRANE",
			want:   []Mark{Correct, Correct, Correct, Correct, Correct},
		},
		{
			name:   "guess shorter than answer",
			guess:  "HI",
			answer: "CRANE",
			// Score builds marks of len(guess); pass1 only marks Correct where
			// i < len(answer) AND equal. pass2 fills the rest from remaining.
			// remaining after pass1: a[0..4] = C,R,A,N,E all added (none matched
			// since guess[0]=H != C, guess[1]=I != R) → {C:1,R:1,A:1,N:1,E:1}
			// pass2:
			//   pos0: H → remaining[H]=0 → Absent
			//   pos1: I → remaining[I]=0 → Absent
			want: []Mark{Absent, Absent},
		},
		{
			name:   "empty guess",
			guess:  "",
			answer: "CRANE",
			want:   []Mark{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Score(tc.guess, tc.answer)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Score(%q, %q) = %v, want %v", tc.guess, tc.answer, got, tc.want)
			}
		})
	}
}

// TestIsWin is trivial but exercises both branches and the empty case.
func TestIsWin(t *testing.T) {
	cases := []struct {
		name  string
		marks []Mark
		want  bool
	}{
		{"all correct", []Mark{Correct, Correct, Correct, Correct, Correct}, true},
		{"one present", []Mark{Correct, Present, Correct, Correct, Correct}, false},
		{"one absent", []Mark{Correct, Correct, Absent, Correct, Correct}, false},
		{"empty", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsWin(tc.marks); got != tc.want {
				t.Fatalf("IsWin(%v) = %v, want %v", tc.marks, got, tc.want)
			}
		})
	}
}

// TestValidateGuess exercises each of the four error paths and the happy path.
func TestValidateGuess(t *testing.T) {
	allowed := map[string]bool{"CRANE": true, "SLATE": true}

	cases := []struct {
		name    string
		guess   string
		allowed map[string]bool
		wantErr error
	}{
		{"valid uppercase", "CRANE", allowed, nil},
		{"valid lowercase", "crane", allowed, nil},
		{"valid with whitespace", "  slate  ", allowed, nil},
		{"too short", "HI", allowed, ErrWrongLength},
		{"too long", "CRANES", allowed, ErrWrongLength},
		{"empty", "", allowed, ErrWrongLength},
		{"digits", "CR4NE", allowed, ErrNotLetters},
		{"punctuation", "CR-NE", allowed, ErrNotLetters},
		{"not in word list", "ZZZZZ", allowed, ErrNotAWord},
		{"nil allowlist skips membership check", "ZZZZZ", nil, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateGuess(tc.guess, tc.allowed)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("ValidateGuess(%q, ...) = %v, want %v", tc.guess, err, tc.wantErr)
			}
		})
	}
}

// TestConstants pins the public constants in case they drift.
func TestConstants(t *testing.T) {
	if WordLength != 5 {
		t.Fatalf("WordLength = %d, want 5", WordLength)
	}
	if MaxGuesses != 6 {
		t.Fatalf("MaxGuesses = %d, want 6", MaxGuesses)
	}
}

// TestMarkValues pins the string values of the Mark type; they appear in JSON
// responses so changing them is a breaking change.
func TestMarkValues(t *testing.T) {
	if string(Correct) != "correct" {
		t.Fatalf("Correct = %q, want %q", Correct, "correct")
	}
	if string(Present) != "present" {
		t.Fatalf("Present = %q, want %q", Present, "present")
	}
	if string(Absent) != "absent" {
		t.Fatalf("Absent = %q, want %q", Absent, "absent")
	}
}

// Compile-time guard: Score's signature must stay (string, string) []Mark.
var _ func(string, string) []Mark = Score

// Keep imports honest even if the test list shrinks later.
var _ = strings.ToUpper