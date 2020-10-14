package fsm

import (
	"context"
	"reflect"
	"testing"
)

// PlantUML for this state machine
/*
@startuml

S1: STATE1
S2: STATE2
S3: STATE3
S4: STATE4
S5: STATE5
S6: STATE6
S7: STATE7
S8: STATE8
S9: STATE9
S10: STATE10

[*] --> S1
S1 --> S2 : v == "a"
S1 --> S1 : v == "b"
S1 --> S3 : v == "c"
S2 --> S3 : v == "d"
S3 --> S4 : v == "e"
S4 --> S5 : v == "f"
S5 --> S6 : v == "g"
S5 --> S5 : v == "h"
S5 --> S4 : v == "i"
S6 --> S1 : v == "j"
S6 --> S7 : v == "k"
S7 --> S3 : v == "l"
S3 --> S8 : v == "m"
S8 --> S2 : v == "n"
S7 --> S9 : v == "o"
S9 --> S10 : v == "p"
S10 --> S10 : v == "q"
S10 --> S7 : v == "r"
S10 --> [*]

@enduml
 */

func TestMachine(t *testing.T) {
	var (
		s1 = NewState("STATE1")
		s2 = NewState("STATE2")
		s3 = NewState("STATE3")
		s4 = NewState("STATE4")
		s5 = NewState("STATE5")
		s6 = NewState("STATE6")
		s7 = NewState("STATE7")
		s8 = NewState("STATE8")
		s9 = NewState("STATE9")
		s10 = NewState("STATE10")
	)

	transitions := []Transition{
		s1.When("v == 'a'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'a', nil
			}
			return false, nil
		}).Then(s2),
		s1.When("v == 'b'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'b', nil
			}
			return false, nil
		}).Then(s1),
		s1.When("v == 'c'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'c', nil
			}
			return false, nil
		}).Then(s3),
		s2.When("v == 'd'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'd', nil
			}
			return false, nil
		}).Then(s3),
		s3.When("v == 'e'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'e', nil
			}
			return false, nil
		}).Then(s4),
		s4.When("v == 'f'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'f', nil
			}
			return false, nil
		}).Then(s5),
		s5.When("v == 'g'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'g', nil
			}
			return false, nil
		}).Then(s6),
		s5.When("v == 'h'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'h', nil
			}
			return false, nil
		}).Then(s5),
		s5.When("v == 'i'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'i', nil
			}
			return false, nil
		}).Then(s4),
		s6.When("v == 'j'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'j', nil
			}
			return false, nil
		}).Then(s1),
		s6.When("v == 'k'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'k', nil
			}
			return false, nil
		}).Then(s7),
		s7.When("v == 'l'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'l', nil
			}
			return false, nil
		}).Then(s3),
		s3.When("v == 'm'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'm', nil
			}
			return false, nil
		}).Then(s8),
		s8.When("v == 'n'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'n', nil
			}
			return false, nil
		}).Then(s2),
		s7.When("v == 'o'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'o', nil
			}
			return false, nil
		}).Then(s9),
		s9.When("v == 'p'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'p', nil
			}
			return false, nil
		}).Then(s10),
		s10.When("v == 'q'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'q', nil
			}
			return false, nil
		}).Then(s10),
		s10.When("v == 'r'", func(ctx context.Context, v interface{}) (bool, error) {
			if b, ok := v.(byte); ok {
				return b == 'r', nil
			}
			return false, nil
		}).Then(s7),
	}

	t.Run("validate", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			m := machine{}
			for _, tr := range transitions {
				m.AddTransition(tr)
			}
			if err := m.SetStart("STATE1"); err != nil {
				t.Fatal(err)
			}
			if err := m.Validate(); err != nil {
				t.Fatal(err)
			}
		})
		t.Run("not valid - duplicate state name", func(t *testing.T) {
			m := machine{}
			for _, tr := range transitions {
				m.AddTransition(tr)
			}
			if err := m.SetStart("STATE1"); err != nil {
				t.Fatal(err)
			}
			s11 := NewState("STATE5")
			m.AddTransition(s11.When("nope", func(_ context.Context, _ interface{}) (bool, error) {
				return false, nil
			}).Then(s10))
			if err := m.Validate(); err == nil {
				t.Fatal("expected invalid machine")
			}
		})
		t.Run("not valid - missing TO state", func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatal("panic expected")
				}
			}()
			m := machine{}
			for _, tr := range transitions {
				m.AddTransition(tr)
			}
			if err := m.SetStart("STATE1"); err != nil {
				t.Fatal(err)
			}
			s11 := NewState("STATE11")
			m.AddTransition(s11.When("nope", func(_ context.Context, _ interface{}) (bool, error) {
				return false, nil
			}))
		})
	})

	t.Run("set start", func(t *testing.T) {
		t.Run("fails - state not found", func(t *testing.T) {
			m := machine{}
			for _, tr := range transitions {
				m.AddTransition(tr)
			}
			if err := m.SetStart("STATE11"); err == nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("set end", func(t *testing.T) {
		t.Run("fails - state not found", func(t *testing.T) {
			m := machine{}
			for _, tr := range transitions {
				m.AddTransition(tr)
			}
			if err := m.SetEndStates("STATE7", "STATE11"); err == nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("is end state", func(t *testing.T) {
		t.Run("fails - state not found", func(t *testing.T) {
			m := machine{}
			for _, tr := range transitions {
				m.AddTransition(tr)
			}
			if err := m.SetEndStates("STATE7", "STATE10"); err != nil {
				t.Fatal(err)
			}

			{
				// move to state 4
				s := "ce"
				ctx := context.Background()

				for i := 0; i < len(s); i++ {
					changed, err := m.Update(ctx, s[i])
					if !changed {
						t.Fatal("change expected")
					}
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
				}
				if m.IsEndState() {
					t.Fatal("unexpected end state")
				}
			}

			{
				// move to state 7
				s := "fgk"
				ctx := context.Background()

				for i := 0; i < len(s); i++ {
					changed, err := m.Update(ctx, s[i])
					if !changed {
						t.Fatal("change expected")
					}
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
				}
				if !m.IsEndState() {
					t.Fatal("expected end state")
				}
			}

			{
				// move to state 10
				s := "op"
				ctx := context.Background()

				for i := 0; i < len(s); i++ {
					changed, err := m.Update(ctx, s[i])
					if !changed {
						t.Fatal("change expected")
					}
					if err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
				}
				if !m.IsEndState() {
					t.Fatal("expected end state")
				}
			}
		})
	})

	t.Run("current", func(t *testing.T) {
		t.Run("default", func(t *testing.T) {
			m := machine{}
			for _, tr := range transitions {
				m.AddTransition(tr)
			}

			s := m.Current()
			if !reflect.DeepEqual(s, s1) {
				t.Fatal("expected default current state")
			}
		})
		t.Run("explicit", func(t *testing.T) {
			m := machine{}
			for _, tr := range transitions {
				m.AddTransition(tr)
			}
			if err := m.SetStart("STATE5"); err != nil {
				t.Fatal(err)
			}

			s := m.Current()
			if !reflect.DeepEqual(s, s5) {
				t.Fatal("expected default current state")
			}
		})
	})

	t.Run("update", func(t *testing.T) {
		m := machine{}
		for _, tr := range transitions {
			m.AddTransition(tr)
		}

		t.Run("adefgkop from 1", func(t *testing.T) {
			if err := m.SetStart("STATE1"); err != nil {
				t.Fatal(err)
			}

			s := "adefgkop"
			ctx := context.Background()

			for i := 0; i < len(s); i++ {
				changed, err := m.Update(ctx, s[i])
				if !changed {
					t.Fatal("change expected")
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			curr := m.Current()
			if !reflect.DeepEqual(curr, s10) {
				t.Fatal("expected s10")
			}
		})

		t.Run("badmndefhifgjbbbcefgkopqqr from 1", func(t *testing.T) {
			if err := m.SetStart("STATE1"); err != nil {
				t.Fatal(err)
			}

			s := "badmndefhifgjbbbcefgkopqqr"
			ctx := context.Background()

			for i := 0; i < len(s); i++ {
				changed, err := m.Update(ctx, s[i])
				if !changed {
					t.Fatal("change expected")
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			curr := m.Current()
			if !reflect.DeepEqual(curr, s7) {
				t.Fatal("expected s7")
			}
		})

		t.Run("roprlefifgjbadefi from 10", func(t *testing.T) {
			if err := m.SetStart("STATE10"); err != nil {
				t.Fatal(err)
			}

			s := "roprlefifgjbadefi"
			ctx := context.Background()

			for i := 0; i < len(s); i++ {
				changed, err := m.Update(ctx, s[i])
				if !changed {
					t.Fatal("change expected")
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			curr := m.Current()
			if !reflect.DeepEqual(curr, s4) {
				t.Fatal("expected s4")
			}
		})

		t.Run("many misses from 1", func(t *testing.T) {
			if err := m.SetStart("STATE1"); err != nil {
				t.Fatal(err)
			}

			s := "zxusqwaffpqndxyvefststhhvvgssk"
			ctx := context.Background()

			var changes, nonChanges int
			for i := 0; i < len(s); i++ {
				changed, err := m.Update(ctx, s[i])
				if changed {
					changes++
				} else {
					nonChanges++
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			curr := m.Current()
			if !reflect.DeepEqual(curr, s7) {
				t.Fatal("expected s7")
			}
			if changes != 8 {
				t.Fatal("expected 8 changes")
			}
			if nonChanges != 22 {
				t.Fatal("expected 22 non-changes")
			}
		})
	})
}
