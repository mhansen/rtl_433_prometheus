package main

import "testing"

func TestChannel(t *testing.T) {
	cases := []struct {
		in   Message
		want string
	}{
		{Message{RawChannel: ""}, ""},
		{Message{RawChannel: "1"}, "1"},
		{Message{RawChannel: "2"}, "2"},
		{Message{RawChannel: 0.0}, "0"},
		{Message{RawChannel: 2.0}, "2"},
	}
	for _, tt := range cases {
		msg := tt.in
		want := tt.want
		got, err := msg.Channel()
		if err != nil {
			t.Errorf("unexpected err=%v", err)
		}
		if got != want {
			t.Errorf("%+v.Channel()=%v, want=%v", msg, got, want)
		}
	}
}
