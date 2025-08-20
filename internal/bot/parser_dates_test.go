package bot

import "testing"

func TestMessageParser_RussianDates(t *testing.T) {
    p := NewMessageParser()
    r1, _ := p.ParseMessage("сегодня 100 кофе")
    if !r1.IsValid || r1.OccurredAt == nil { t.Fatalf("сегодня should set date") }
    r2, _ := p.ParseMessage("вчера 100 кофе")
    if !r2.IsValid || r2.OccurredAt == nil { t.Fatalf("вчера should set date") }
    r3, _ := p.ParseMessage("позавчера 100 кофе")
    if !r3.IsValid || r3.OccurredAt == nil { t.Fatalf("позавчера should set date") }
}

func TestMessageParser_DDMM(t *testing.T) {
    p := NewMessageParser()
    r, _ := p.ParseMessage("12.08 100 такси")
    if !r.IsValid || r.OccurredAt == nil { t.Fatalf("DD.MM should parse") }
}


