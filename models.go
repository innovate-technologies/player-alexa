package main

type AudioStartResponse struct {
	Version  string `json:"version"`
	Response struct {
		ShouldEndSession bool             `json:"shouldEndSession"`
		Directives       []AudioDirective `json:"directives"`
	} `json:"response"`
}

type AudioDirective struct {
	Type         string   `json:"type"`
	PlayBehavior string   `json:"playBehavior"`
	AudioItem    AudoItem `json:"audioItem"`
}

type AudoItem struct {
	Stream Stream `json:"stream"`
}

type Stream struct {
	URL                   string      `json:"url"`
	Token                 string      `json:"token"`
	ExpectedPreviousToken interface{} `json:"expectedPreviousToken"`
	OffsetInMilliseconds  int         `json:"offsetInMilliseconds"`
}

func NewAudioStartResponse() AudioStartResponse {
	out := AudioStartResponse{}

	out.Version = "1.0"

	return out
}
