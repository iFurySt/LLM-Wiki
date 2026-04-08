package ingest

import "testing"

func TestDetectURLSource(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{url: "https://x.com/openai/status/1", want: SourceX.ID},
		{url: "https://twitter.com/openai/status/1", want: SourceX.ID},
		{url: "https://mp.weixin.qq.com/s/demo", want: SourceWechat.ID},
		{url: "https://www.youtube.com/watch?v=demo", want: SourceYouTube.ID},
		{url: "https://www.zhihu.com/question/1", want: SourceZhihu.ID},
		{url: "https://www.xiaohongshu.com/explore/demo", want: SourceXHS.ID},
		{url: "https://example.com/post", want: SourceWeb.ID},
	}

	for _, tt := range tests {
		got := DetectURLSource(tt.url)
		if got.ID != tt.want {
			t.Fatalf("DetectURLSource(%q) = %q, want %q", tt.url, got.ID, tt.want)
		}
	}
}

func TestSourceFromInlineText(t *testing.T) {
	source := SourceFromInlineText()
	if source == nil || source.ID != SourcePlainText.ID {
		t.Fatalf("expected inline source metadata")
	}
	if source.InputMode != "text" {
		t.Fatalf("expected text input mode, got %q", source.InputMode)
	}
}

func TestSlugify(t *testing.T) {
	if got := Slugify("Launch Plan v2"); got != "launch-plan-v2" {
		t.Fatalf("Slugify returned %q", got)
	}
	if got := Slugify("  "); got != "" {
		t.Fatalf("expected empty slug, got %q", got)
	}
}
