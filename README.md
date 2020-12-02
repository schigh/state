# state

`state` implements lightweight and concurrent state machines for various uses.

## TL;DR examples

The following example is the classic use case for a finite state machine. 
The machine evaluates if a series of characters match the pattern `a+bc`.

A start state is created, as well as states for all acceptable transitions. 
The `c` state is set as one of the end states (there can be more than one).

For each string, we reset the state machine and feed it into the start state, 
popping the first character off for each evaluation, until we either reach an 
invalid state, or there are no more characters to evaluate.
```go
package main
import (
    "context"
    "fmt"
    
    "github.com/schigh/state/fsm"
)

func main() {
    s1 := fsm.NewState("start")
    s2 := fsm.NewState("a")
    s3 := fsm.NewState("b")
    s4 := fsm.NewState("c")

    // the start state.  since this will be the first one added,
    // it will be set as the start state implicitly
    // this state will transition to the next state if the first
    // character in our word is 'a'
    case1 := s1.When("starts with a", func(ctx context.Context, v interface{}) ( bool,  error) {
        s, _ := v.(string)
        return len(s) > 0 && s[0] == 'a', nil
    }).Then(s2)

    // this state will transition to the next state if the
    // first character in our word is 'b'
    case2 := s2.When("head is b", func(ctx context.Context, v interface{}) ( bool,  error) {
        s, _ := v.(string)
        return len(s) > 0 && s[0] == 'b', nil
    }).Then(s3)

    // this condition represents a loop in the 'a' state, where any string that starts
    // with 1...N instances of 'a' is valid
    case3 := s2.When("head is a", func(ctx context.Context, v interface{}) ( bool,  error) {
        s, _ := v.(string)
        return len(s) > 0 && s[0] == 'a', nil
    }).Then(s2)

    // this will transition to the final state 'c'
    case4 := s3.When("head is c", func(ctx context.Context, v interface{}) ( bool,  error) {
        s, _ := v.(string)
        return len(s) > 0 && s[0] == 'c', nil
    }).Then(s4)

    machine := fsm.NewMachine(fsm.WithTransitions(case1, case2, case3, case4))
    if err := machine.SetEndStates("c"); err != nil {
        panic(err)
    }
    if err := machine.Validate(); err != nil {
        panic(err)
    }

    words := []string{
        "abc",
        "cba",
        "aaaaaabc",
        "aaababc",
        "ab",
    }
    ctx := context.Background()

    for _, word := range words {
        if resetErr := machine.Reset(); resetErr != nil {
            panic(resetErr)
        }

        valid := true
        var err error

        test := word
        for valid {
            valid, err = machine.Update(ctx, test)
            if err != nil {
                panic(err)
            }
            if len(test) <= 1 {
                break
            }
            test = test[1:]
        }

        if valid && machine.IsEndState(){
            fmt.Printf("'%s' matches the pattern\n", word)
            continue
        }
        fmt.Printf("'%s' DOES NOT MATCH the pattern\n", word)
    }
}
```
This is the output from the above example:
```text
'abc' matches the pattern
'cba' DOES NOT MATCH the pattern
'aaaaaabc' matches the pattern
'aaababc' DOES NOT MATCH the pattern
'ab' DOES NOT MATCH the pattern
```

## Packages

### fsm

The `fsm` package implements a more traditional finite state machine, where the 
transition from one state to another is triggered by a targeted evaluation.  
See the [README](./fsm/README.md) for detailed docs.

### flipflop

The `flipflop` package implements a set of toggle switches, which can trigger 
downstream events when a toggle occurs.  See the [README](./flipflop/README.md) 
for more details.
