package main

import (
	"testing"
	"time"
)

var testsPrettyDuration = []struct {
	in  time.Duration
	out string
}{
	{time.Second * 0, "0s ago"},
	{time.Second * 1, "1s ago"},
	{time.Second * 2, "2s ago"},
	{time.Minute*1 + time.Second*0, "1m ago"},
	{time.Minute*1 + time.Second*1, "1m ago"},
	{time.Minute*1 + time.Second*2, "1m ago"},
	{time.Minute*2 + time.Second*0, "2m ago"},
	{time.Minute*2 + time.Second*1, "2m ago"},
	{time.Minute*2 + time.Second*2, "2m ago"},
	{time.Hour*1 + time.Minute*0, "1h ago"},
	{time.Hour*1 + time.Minute*1, "1h 1m ago"},
	{time.Hour*1 + time.Minute*2, "1h 2m ago"},
	{time.Hour*2 + time.Minute*0, "2h ago"},
	{time.Hour*2 + time.Minute*1, "2h 1m ago"},
	{time.Hour*2 + time.Minute*2, "2h 2m ago"},
	{time.Hour*24 + time.Minute*0, "1d ago"},
	{time.Hour*24 + time.Minute*1, "1d ago"},
	{time.Hour*24 + time.Minute*2, "1d ago"},
	{time.Hour*24 + time.Minute*60, "1d 1h ago"},
	{time.Hour*24 + time.Minute*120, "1d 2h ago"},
	{time.Hour*24*2 + time.Minute*0, "2d ago"},
	{time.Hour*24*2 + time.Minute*1, "2d ago"},
	{time.Hour*24*2 + time.Minute*2, "2d ago"},
	{time.Hour*24*2 + time.Minute*60, "2d 1h ago"},
	{time.Hour*24*2 + time.Minute*120, "2d 2h ago"},
	{time.Hour*24*6 + time.Minute*0, "6d ago"},
	{time.Hour*24*6 + time.Minute*1, "6d ago"},
	{time.Hour*24*6 + time.Minute*2, "6d ago"},
	{time.Hour*24*6 + time.Minute*60, "6d 1h ago"},
	{time.Hour*24*6 + time.Minute*120, "6d 2h ago"},
	{time.Hour*24*7 + time.Minute*0, "7d ago"},
	{time.Hour*24*7 + time.Minute*1, "7d ago"},
	{time.Hour*24*7 + time.Minute*2, "7d ago"},
	{time.Hour*24*7 + time.Minute*60, "7d ago"},
	{time.Hour*24*7 + time.Minute*120, "7d ago"},
	{time.Hour*24*14 + time.Minute*0, "2w ago"},
	{time.Hour*24*14 + time.Minute*1, "2w ago"},
	{time.Hour*24*365 + time.Minute*0, "1y 0w ago"},
	{time.Hour*24*(365+7*1) + time.Minute*0, "1y 1w ago"},
	{time.Hour*24*(365+7*2) + time.Minute*0, "1y 2w ago"},
}

func TestPrettyDuration(t *testing.T) {
	for _, tt := range testsPrettyDuration {
		out := PrettyDuration(tt.in)
		if out != tt.out {
			t.Errorf("pretty_duration(%q) => %q, want %q", tt.in, out, tt.out)
		}
	}
}

var testsNormalizeURL = []struct {
	in  string
	out string
}{
	{"https://example.org", "http://example.org"},
	{"http://example.org:80", "http://example.org"},
	{"https://example.org:443", "http://example.org"},
	{"http://example.org/", "http://example.org"},
	{"http://example.org/bar/", "http://example.org/bar"},
	{"http://example.org/bar", "http://example.org/bar"},
	{"http://example.org/bar/../quux", "http://example.org/quux"},
	{"http://example.org/b%61r", "http://example.org/bar"},
	{"http://example.org/b%6F%6f", "http://example.org/boo"},
	{"http://bob:s3cr3t@example.org/", "http://example.org"},
}

func TestNormalizeURL(t *testing.T) {
	for _, tt := range testsNormalizeURL {
		out := NormalizeURL(tt.in)
		if out != tt.out {
			t.Errorf("normalizeURL(%q) => %q, want %q", tt.in, out, tt.out)
		}
	}
}
