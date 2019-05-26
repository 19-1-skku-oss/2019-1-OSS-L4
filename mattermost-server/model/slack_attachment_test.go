package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSlackAttachment(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		post := &Post{}
		attachments := []*SlackAttachment{}

		ParseSlackAttachment(post, attachments)

		expectedPost := &Post{
			Type: POST_SLACK_ATTACHMENT,
			Props: map[string]interface{}{
				"attachments": []*SlackAttachment{},
			},
		}
		assert.Equal(t, expectedPost, post)
	})

	t.Run("list with nil", func(t *testing.T) {
		post := &Post{}
		attachments := []*SlackAttachment{
			nil,
		}

		ParseSlackAttachment(post, attachments)

		expectedPost := &Post{
			Type: POST_SLACK_ATTACHMENT,
			Props: map[string]interface{}{
				"attachments": []*SlackAttachment{},
			},
		}
		assert.Equal(t, expectedPost, post)
	})
}
