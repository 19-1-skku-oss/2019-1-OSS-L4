// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package storetest

import (
	"net/http"
	"testing"
	"time"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/store"
	"github.com/stretchr/testify/require"
)

func TestWebhookStore(t *testing.T, ss store.Store) {
	t.Run("SaveIncoming", func(t *testing.T) { testWebhookStoreSaveIncoming(t, ss) })
	t.Run("UpdateIncoming", func(t *testing.T) { testWebhookStoreUpdateIncoming(t, ss) })
	t.Run("GetIncoming", func(t *testing.T) { testWebhookStoreGetIncoming(t, ss) })
	t.Run("GetIncomingList", func(t *testing.T) { testWebhookStoreGetIncomingList(t, ss) })
	t.Run("GetIncomingByTeam", func(t *testing.T) { testWebhookStoreGetIncomingByTeam(t, ss) })
	t.Run("DeleteIncoming", func(t *testing.T) { testWebhookStoreDeleteIncoming(t, ss) })
	t.Run("DeleteIncomingByChannel", func(t *testing.T) { testWebhookStoreDeleteIncomingByChannel(t, ss) })
	t.Run("DeleteIncomingByUser", func(t *testing.T) { testWebhookStoreDeleteIncomingByUser(t, ss) })
	t.Run("SaveOutgoing", func(t *testing.T) { testWebhookStoreSaveOutgoing(t, ss) })
	t.Run("GetOutgoing", func(t *testing.T) { testWebhookStoreGetOutgoing(t, ss) })
	t.Run("GetOutgoingList", func(t *testing.T) { testWebhookStoreGetOutgoingList(t, ss) })
	t.Run("GetOutgoingByChannel", func(t *testing.T) { testWebhookStoreGetOutgoingByChannel(t, ss) })
	t.Run("GetOutgoingByTeam", func(t *testing.T) { testWebhookStoreGetOutgoingByTeam(t, ss) })
	t.Run("DeleteOutgoing", func(t *testing.T) { testWebhookStoreDeleteOutgoing(t, ss) })
	t.Run("DeleteOutgoingByChannel", func(t *testing.T) { testWebhookStoreDeleteOutgoingByChannel(t, ss) })
	t.Run("DeleteOutgoingByUser", func(t *testing.T) { testWebhookStoreDeleteOutgoingByUser(t, ss) })
	t.Run("UpdateOutgoing", func(t *testing.T) { testWebhookStoreUpdateOutgoing(t, ss) })
	t.Run("CountIncoming", func(t *testing.T) { testWebhookStoreCountIncoming(t, ss) })
	t.Run("CountOutgoing", func(t *testing.T) { testWebhookStoreCountOutgoing(t, ss) })
}

func testWebhookStoreSaveIncoming(t *testing.T, ss store.Store) {
	o1 := buildIncomingWebhook()

	if _, err := ss.Webhook().SaveIncoming(o1); err != nil {
		t.Fatal("couldn't save item", err)
	}

	if _, err := ss.Webhook().SaveIncoming(o1); err == nil {
		t.Fatal("shouldn't be able to update from save")
	}
}

func testWebhookStoreUpdateIncoming(t *testing.T, ss store.Store) {

	var err *model.AppError

	o1 := buildIncomingWebhook()
	o1, err = ss.Webhook().SaveIncoming(o1)
	if err != nil {
		t.Fatal("unable to save webhook", err)
	}

	previousUpdatedAt := o1.UpdateAt

	o1.DisplayName = "TestHook"
	time.Sleep(10 * time.Millisecond)

	webhook, err := ss.Webhook().UpdateIncoming(o1)
	require.Nil(t, err)

	if webhook.UpdateAt == previousUpdatedAt {
		t.Fatal("should have updated the UpdatedAt of the hook")
	}

	if webhook.DisplayName != "TestHook" {
		t.Fatal("display name is not updated")
	}

}

func testWebhookStoreGetIncoming(t *testing.T, ss store.Store) {
	var err *model.AppError

	o1 := buildIncomingWebhook()
	o1, err = ss.Webhook().SaveIncoming(o1)
	if err != nil {
		t.Fatal("unable to save webhook", err)
	}

	webhook, err := ss.Webhook().GetIncoming(o1.Id, false)
	require.Nil(t, err)
	if webhook.CreateAt != o1.CreateAt {
		t.Fatal("invalid returned webhook")
	}

	webhook, err = ss.Webhook().GetIncoming(o1.Id, true)
	require.Nil(t, err)
	if webhook.CreateAt != o1.CreateAt {
		t.Fatal("invalid returned webhook")
	}

	if _, err = ss.Webhook().GetIncoming("123", false); err == nil {
		t.Fatal("Missing id should have failed")
	}

	if _, err = ss.Webhook().GetIncoming("123", true); err == nil {
		t.Fatal("Missing id should have failed")
	}

	if _, err = ss.Webhook().GetIncoming("123", true); err.StatusCode != http.StatusNotFound {
		t.Fatal("Should have set the status as not found for missing id")
	}
}

func testWebhookStoreGetIncomingList(t *testing.T, ss store.Store) {
	o1 := &model.IncomingWebhook{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.TeamId = model.NewId()

	var err *model.AppError
	o1, err = ss.Webhook().SaveIncoming(o1)
	if err != nil {
		t.Fatal("unable to save webhook", err)
	}

	if hooks, err := ss.Webhook().GetIncomingList(0, 1000); err != nil {
		t.Fatal(err)
	} else {
		found := false
		for _, hook := range hooks {
			if hook.Id == o1.Id {
				found = true
			}
		}
		if !found {
			t.Fatal("missing webhook")
		}
	}

	if hooks, err := ss.Webhook().GetIncomingList(0, 1); err != nil {
		t.Fatal(err)
	} else {
		if len(hooks) != 1 {
			t.Fatal("only 1 should be returned")
		}
	}
}

func testWebhookStoreGetIncomingByTeam(t *testing.T, ss store.Store) {
	var err *model.AppError

	o1 := buildIncomingWebhook()
	o1, err = ss.Webhook().SaveIncoming(o1)
	require.Nil(t, err)

	if hooks, err := ss.Webhook().GetIncomingByTeam(o1.TeamId, 0, 100); err != nil {
		t.Fatal(err)
	} else {
		if hooks[0].CreateAt != o1.CreateAt {
			t.Fatal("invalid returned webhook")
		}
	}

	if hooks, err := ss.Webhook().GetIncomingByTeam("123", 0, 100); err != nil {
		t.Fatal(err)
	} else {
		if len(hooks) != 0 {
			t.Fatal("no webhooks should have returned")
		}
	}
}

func testWebhookStoreGetIncomingByChannel(t *testing.T, ss store.Store) {
	o1 := buildIncomingWebhook()

	o1, err := ss.Webhook().SaveIncoming(o1)
	if err != nil {
		t.Fatal("unable to save webhook")
	}

	webhooks, err := ss.Webhook().GetIncomingByChannel(o1.ChannelId)
	require.Nil(t, err)
	if webhooks[0].CreateAt != o1.CreateAt {
		t.Fatal("invalid returned webhook")
	}

	if webhooks, err = ss.Webhook().GetIncomingByChannel("123"); err != nil {
		t.Fatal(err)
	} else {
		if len(webhooks) != 0 {
			t.Fatal("no webhooks should have returned")
		}
	}
}

func testWebhookStoreDeleteIncoming(t *testing.T, ss store.Store) {
	var err *model.AppError

	o1 := buildIncomingWebhook()
	o1, err = ss.Webhook().SaveIncoming(o1)
	if err != nil {
		t.Fatal("unable to save webhook", err)
	}

	webhook, err := ss.Webhook().GetIncoming(o1.Id, true)
	require.Nil(t, err)
	if webhook.CreateAt != o1.CreateAt {
		t.Fatal("invalid returned webhook")
	}

	if err = ss.Webhook().DeleteIncoming(o1.Id, model.GetMillis()); err != nil {
		t.Fatal(err)
	}

	webhook, err = ss.Webhook().GetIncoming(o1.Id, true)
	require.NotNil(t, err)
}

func testWebhookStoreDeleteIncomingByChannel(t *testing.T, ss store.Store) {
	var err *model.AppError

	o1 := buildIncomingWebhook()
	o1, err = ss.Webhook().SaveIncoming(o1)
	if err != nil {
		t.Fatal("unable to save webhook", err)
	}

	webhook, err := ss.Webhook().GetIncoming(o1.Id, true)
	require.Nil(t, err)
	if webhook.CreateAt != o1.CreateAt {
		t.Fatal("invalid returned webhook")
	}

	if err = ss.Webhook().PermanentDeleteIncomingByChannel(o1.ChannelId); err != nil {
		t.Fatal(err)
	}

	if _, err = ss.Webhook().GetIncoming(o1.Id, true); err == nil {
		t.Fatal("Missing id should have failed")
	}
}

func testWebhookStoreDeleteIncomingByUser(t *testing.T, ss store.Store) {
	var err *model.AppError

	o1 := buildIncomingWebhook()
	o1, err = ss.Webhook().SaveIncoming(o1)
	if err != nil {
		t.Fatal("unable to save webhook", err)
	}

	webhook, err := ss.Webhook().GetIncoming(o1.Id, true)
	require.Nil(t, err)
	if webhook.CreateAt != o1.CreateAt {
		t.Fatal("invalid returned webhook")
	}

	if err = ss.Webhook().PermanentDeleteIncomingByUser(o1.UserId); err != nil {
		t.Fatal(err)
	}

	if _, err = ss.Webhook().GetIncoming(o1.Id, true); err == nil {
		t.Fatal("Missing id should have failed")
	}
}

func buildIncomingWebhook() *model.IncomingWebhook {
	o1 := &model.IncomingWebhook{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.TeamId = model.NewId()

	return o1
}

func testWebhookStoreSaveOutgoing(t *testing.T, ss store.Store) {
	o1 := model.OutgoingWebhook{}
	o1.ChannelId = model.NewId()
	o1.CreatorId = model.NewId()
	o1.TeamId = model.NewId()
	o1.CallbackURLs = []string{"http://nowhere.com/"}
	o1.Username = "test-user-name"
	o1.IconURL = "http://nowhere.com/icon"

	if _, err := ss.Webhook().SaveOutgoing(&o1); err != nil {
		t.Fatal("couldn't save item", err)
	}

	if _, err := ss.Webhook().SaveOutgoing(&o1); err == nil {
		t.Fatal("shouldn't be able to update from save")
	}
}

func testWebhookStoreGetOutgoing(t *testing.T, ss store.Store) {
	o1 := &model.OutgoingWebhook{}
	o1.ChannelId = model.NewId()
	o1.CreatorId = model.NewId()
	o1.TeamId = model.NewId()
	o1.CallbackURLs = []string{"http://nowhere.com/"}
	o1.Username = "test-user-name"
	o1.IconURL = "http://nowhere.com/icon"

	o1, _ = ss.Webhook().SaveOutgoing(o1)

	webhook, err := ss.Webhook().GetOutgoing(o1.Id)
	require.Nil(t, err)
	if webhook.CreateAt != o1.CreateAt {
		t.Fatal("invalid returned webhook")
	}

	if _, err := ss.Webhook().GetOutgoing("123"); err == nil {
		t.Fatal("Missing id should have failed")
	}
}

func testWebhookStoreGetOutgoingList(t *testing.T, ss store.Store) {
	o1 := &model.OutgoingWebhook{}
	o1.ChannelId = model.NewId()
	o1.CreatorId = model.NewId()
	o1.TeamId = model.NewId()
	o1.CallbackURLs = []string{"http://nowhere.com/"}

	o1, _ = ss.Webhook().SaveOutgoing(o1)

	o2 := &model.OutgoingWebhook{}
	o2.ChannelId = model.NewId()
	o2.CreatorId = model.NewId()
	o2.TeamId = model.NewId()
	o2.CallbackURLs = []string{"http://nowhere.com/"}

	o2, _ = ss.Webhook().SaveOutgoing(o2)

	if r1, err := ss.Webhook().GetOutgoingList(0, 1000); err != nil {
		t.Fatal(err)
	} else {
		hooks := r1
		found1 := false
		found2 := false

		for _, hook := range hooks {
			if hook.CreateAt != o1.CreateAt {
				found1 = true
			}

			if hook.CreateAt != o2.CreateAt {
				found2 = true
			}
		}

		if !found1 {
			t.Fatal("missing hook1")
		}
		if !found2 {
			t.Fatal("missing hook2")
		}
	}

	if result, err := ss.Webhook().GetOutgoingList(0, 2); err != nil {
		t.Fatal(err)
	} else {
		if len(result) != 2 {
			t.Fatal("wrong number of hooks returned")
		}
	}
}

func testWebhookStoreGetOutgoingByChannel(t *testing.T, ss store.Store) {
	o1 := &model.OutgoingWebhook{}
	o1.ChannelId = model.NewId()
	o1.CreatorId = model.NewId()
	o1.TeamId = model.NewId()
	o1.CallbackURLs = []string{"http://nowhere.com/"}

	o1, _ = ss.Webhook().SaveOutgoing(o1)

	if r1, err := ss.Webhook().GetOutgoingByChannel(o1.ChannelId, 0, 100); err != nil {
		t.Fatal(err)
	} else {
		if r1[0].CreateAt != o1.CreateAt {
			t.Fatal("invalid returned webhook")
		}
	}

	if result, err := ss.Webhook().GetOutgoingByChannel("123", -1, -1); err != nil {
		t.Fatal(err)
	} else {
		if len(result) != 0 {
			t.Fatal("no webhooks should have returned")
		}
	}
}

func testWebhookStoreGetOutgoingByTeam(t *testing.T, ss store.Store) {
	o1 := &model.OutgoingWebhook{}
	o1.ChannelId = model.NewId()
	o1.CreatorId = model.NewId()
	o1.TeamId = model.NewId()
	o1.CallbackURLs = []string{"http://nowhere.com/"}

	o1, _ = ss.Webhook().SaveOutgoing(o1)

	if r1, err := ss.Webhook().GetOutgoingByTeam(o1.TeamId, 0, 100); err != nil {
		t.Fatal(err)
	} else {
		if r1[0].CreateAt != o1.CreateAt {
			t.Fatal("invalid returned webhook")
		}
	}

	if result, err := ss.Webhook().GetOutgoingByTeam("123", -1, -1); err != nil {
		t.Fatal(err)
	} else {
		if len(result) != 0 {
			t.Fatal("no webhooks should have returned")
		}
	}
}

func testWebhookStoreDeleteOutgoing(t *testing.T, ss store.Store) {
	o1 := &model.OutgoingWebhook{}
	o1.ChannelId = model.NewId()
	o1.CreatorId = model.NewId()
	o1.TeamId = model.NewId()
	o1.CallbackURLs = []string{"http://nowhere.com/"}

	o1, _ = ss.Webhook().SaveOutgoing(o1)

	webhook, err := ss.Webhook().GetOutgoing(o1.Id)
	require.Nil(t, err)
	if webhook.CreateAt != o1.CreateAt {
		t.Fatal("invalid returned webhook")
	}

	if err := ss.Webhook().DeleteOutgoing(o1.Id, model.GetMillis()); err != nil {
		t.Fatal(err)
	}

	if _, err := ss.Webhook().GetOutgoing(o1.Id); err == nil {
		t.Fatal("Missing id should have failed")
	}
}

func testWebhookStoreDeleteOutgoingByChannel(t *testing.T, ss store.Store) {
	o1 := &model.OutgoingWebhook{}
	o1.ChannelId = model.NewId()
	o1.CreatorId = model.NewId()
	o1.TeamId = model.NewId()
	o1.CallbackURLs = []string{"http://nowhere.com/"}

	o1, _ = ss.Webhook().SaveOutgoing(o1)

	webhook, err := ss.Webhook().GetOutgoing(o1.Id)
	require.Nil(t, err)
	if webhook.CreateAt != o1.CreateAt {
		t.Fatal("invalid returned webhook")
	}

	if err := ss.Webhook().PermanentDeleteOutgoingByChannel(o1.ChannelId); err != nil {
		t.Fatal(err)
	}

	if _, err := ss.Webhook().GetOutgoing(o1.Id); err == nil {
		t.Fatal("Missing id should have failed")
	}
}

func testWebhookStoreDeleteOutgoingByUser(t *testing.T, ss store.Store) {
	o1 := &model.OutgoingWebhook{}
	o1.ChannelId = model.NewId()
	o1.CreatorId = model.NewId()
	o1.TeamId = model.NewId()
	o1.CallbackURLs = []string{"http://nowhere.com/"}

	o1, _ = ss.Webhook().SaveOutgoing(o1)

	webhook, err := ss.Webhook().GetOutgoing(o1.Id)
	require.Nil(t, err)
	if webhook.CreateAt != o1.CreateAt {
		t.Fatal("invalid returned webhook")
	}

	if err := ss.Webhook().PermanentDeleteOutgoingByUser(o1.CreatorId); err != nil {
		t.Fatal(err)
	}

	if _, err := ss.Webhook().GetOutgoing(o1.Id); err == nil {
		t.Fatal("Missing id should have failed")
	}
}

func testWebhookStoreUpdateOutgoing(t *testing.T, ss store.Store) {
	o1 := &model.OutgoingWebhook{}
	o1.ChannelId = model.NewId()
	o1.CreatorId = model.NewId()
	o1.TeamId = model.NewId()
	o1.CallbackURLs = []string{"http://nowhere.com/"}
	o1.Username = "test-user-name"
	o1.IconURL = "http://nowhere.com/icon"

	o1, _ = ss.Webhook().SaveOutgoing(o1)

	o1.Token = model.NewId()
	o1.Username = "another-test-user-name"

	if _, err := ss.Webhook().UpdateOutgoing(o1); err != nil {
		t.Fatal(err)
	}
}

func testWebhookStoreCountIncoming(t *testing.T, ss store.Store) {
	o1 := &model.IncomingWebhook{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.TeamId = model.NewId()

	_, _ = ss.Webhook().SaveIncoming(o1)

	c, err := ss.Webhook().AnalyticsIncomingCount("")
	if err != nil {
		t.Fatal(err)
	}

	if c == 0 {
		t.Fatal("should have at least 1 incoming hook")
	}
}

func testWebhookStoreCountOutgoing(t *testing.T, ss store.Store) {
	o1 := &model.OutgoingWebhook{}
	o1.ChannelId = model.NewId()
	o1.CreatorId = model.NewId()
	o1.TeamId = model.NewId()
	o1.CallbackURLs = []string{"http://nowhere.com/"}

	ss.Webhook().SaveOutgoing(o1)

	if r, err := ss.Webhook().AnalyticsOutgoingCount(""); err != nil {
		t.Fatal(err)
	} else {
		if r == 0 {
			t.Fatal("should have at least 1 outgoing hook")
		}
	}
}
