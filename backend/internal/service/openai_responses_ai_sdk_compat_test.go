package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestEnsureOpenAIResponsesAISDKCompatibilityAddsRequiredMessageFields(t *testing.T) {
	body := []byte(`{"id":"resp_abc","object":"response","model":"gpt-5.4-mini","status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"OK"}]}],"usage":{"input_tokens":1,"output_tokens":1}}`)

	got := ensureOpenAIResponsesAISDKCompatibility(body)

	require.Equal(t, "msg_resp_abc_0", gjson.GetBytes(got, "output.0.id").String())
	require.True(t, gjson.GetBytes(got, "output.0.content.0.annotations").IsArray())
	require.JSONEq(t, `[]`, gjson.GetBytes(got, "output.0.content.0.annotations").Raw)
	require.Equal(t, "OK", gjson.GetBytes(got, "output.0.content.0.text").String())
}

func TestEnsureOpenAIResponsesAISDKCompatibilityPreservesExistingFields(t *testing.T) {
	body := []byte(`{"id":"resp_abc","output":[{"type":"message","id":"msg_existing","role":"assistant","content":[{"type":"output_text","text":"OK","annotations":[{"type":"url_citation","start_index":0,"end_index":2,"url":"https://example.com","title":"Example"}]}]}]}`)

	got := ensureOpenAIResponsesAISDKCompatibility(body)

	require.Equal(t, "msg_existing", gjson.GetBytes(got, "output.0.id").String())
	require.Len(t, gjson.GetBytes(got, "output.0.content.0.annotations").Array(), 1)
}
