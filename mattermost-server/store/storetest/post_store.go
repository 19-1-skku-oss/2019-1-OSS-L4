// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package storetest

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/store"
	"github.com/mattermost/mattermost-server/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostStore(t *testing.T, ss store.Store, s SqlSupplier) {
	t.Run("Save", func(t *testing.T) { testPostStoreSave(t, ss) })
	t.Run("SaveAndUpdateChannelMsgCounts", func(t *testing.T) { testPostStoreSaveChannelMsgCounts(t, ss) })
	t.Run("Get", func(t *testing.T) { testPostStoreGet(t, ss) })
	t.Run("GetSingle", func(t *testing.T) { testPostStoreGetSingle(t, ss) })
	t.Run("GetEtagCache", func(t *testing.T) { testGetEtagCache(t, ss) })
	t.Run("Update", func(t *testing.T) { testPostStoreUpdate(t, ss) })
	t.Run("Delete", func(t *testing.T) { testPostStoreDelete(t, ss) })
	t.Run("Delete1Level", func(t *testing.T) { testPostStoreDelete1Level(t, ss) })
	t.Run("Delete2Level", func(t *testing.T) { testPostStoreDelete2Level(t, ss) })
	t.Run("PermDelete1Level", func(t *testing.T) { testPostStorePermDelete1Level(t, ss) })
	t.Run("PermDelete1Level2", func(t *testing.T) { testPostStorePermDelete1Level2(t, ss) })
	t.Run("GetWithChildren", func(t *testing.T) { testPostStoreGetWithChildren(t, ss) })
	t.Run("GetPostsWithDetails", func(t *testing.T) { testPostStoreGetPostsWithDetails(t, ss) })
	t.Run("GetPostsBeforeAfter", func(t *testing.T) { testPostStoreGetPostsBeforeAfter(t, ss) })
	t.Run("GetPostsSince", func(t *testing.T) { testPostStoreGetPostsSince(t, ss) })
	t.Run("Search", func(t *testing.T) { testPostStoreSearch(t, ss) })
	t.Run("UserCountsWithPostsByDay", func(t *testing.T) { testUserCountsWithPostsByDay(t, ss) })
	t.Run("PostCountsByDay", func(t *testing.T) { testPostCountsByDay(t, ss) })
	t.Run("GetFlaggedPostsForTeam", func(t *testing.T) { testPostStoreGetFlaggedPostsForTeam(t, ss, s) })
	t.Run("GetFlaggedPosts", func(t *testing.T) { testPostStoreGetFlaggedPosts(t, ss) })
	t.Run("GetFlaggedPostsForChannel", func(t *testing.T) { testPostStoreGetFlaggedPostsForChannel(t, ss) })
	t.Run("GetPostsCreatedAt", func(t *testing.T) { testPostStoreGetPostsCreatedAt(t, ss) })
	t.Run("Overwrite", func(t *testing.T) { testPostStoreOverwrite(t, ss) })
	t.Run("GetPostsByIds", func(t *testing.T) { testPostStoreGetPostsByIds(t, ss) })
	t.Run("GetPostsBatchForIndexing", func(t *testing.T) { testPostStoreGetPostsBatchForIndexing(t, ss) })
	t.Run("PermanentDeleteBatch", func(t *testing.T) { testPostStorePermanentDeleteBatch(t, ss) })
	t.Run("GetOldest", func(t *testing.T) { testPostStoreGetOldest(t, ss) })
	t.Run("TestGetMaxPostSize", func(t *testing.T) { testGetMaxPostSize(t, ss) })
	t.Run("GetParentsForExportAfter", func(t *testing.T) { testPostStoreGetParentsForExportAfter(t, ss) })
	t.Run("GetRepliesForExport", func(t *testing.T) { testPostStoreGetRepliesForExport(t, ss) })
	t.Run("GetDirectPostParentsForExportAfter", func(t *testing.T) { testPostStoreGetDirectPostParentsForExportAfter(t, ss, s) })
	t.Run("GetDirectPostParentsForExportAfterDeleted", func(t *testing.T) { testPostStoreGetDirectPostParentsForExportAfterDeleted(t, ss, s) })
	t.Run("GetDirectPostParentsForExportAfterBatched", func(t *testing.T) { testPostStoreGetDirectPostParentsForExportAfterBatched(t, ss, s) })
}

func testPostStoreSave(t *testing.T, ss store.Store) {
	o1 := model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"

	if err := (<-ss.Post().Save(&o1)).Err; err != nil {
		t.Fatal("couldn't save item", err)
	}

	if err := (<-ss.Post().Save(&o1)).Err; err == nil {
		t.Fatal("shouldn't be able to update from save")
	}
}

func testPostStoreSaveChannelMsgCounts(t *testing.T, ss store.Store) {
	c1 := &model.Channel{Name: model.NewId(), DisplayName: "posttestchannel", Type: model.CHANNEL_OPEN}
	res := <-ss.Channel().Save(c1, 1000000)
	require.Nil(t, res.Err)

	o1 := model.Post{}
	o1.ChannelId = c1.Id
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"

	require.Nil(t, (<-ss.Post().Save(&o1)).Err)

	c1, err := ss.Channel().Get(c1.Id, false)
	require.Nil(t, err)
	assert.Equal(t, int64(1), c1.TotalMsgCount, "Message count should update by 1")

	o1.Id = ""
	o1.Type = model.POST_ADD_TO_TEAM
	require.Nil(t, (<-ss.Post().Save(&o1)).Err)

	o1.Id = ""
	o1.Type = model.POST_REMOVE_FROM_TEAM
	require.Nil(t, (<-ss.Post().Save(&o1)).Err)

	c1, err = ss.Channel().Get(c1.Id, false)
	require.Nil(t, err)
	assert.Equal(t, int64(1), c1.TotalMsgCount, "Message count should not update for team add/removed message")

	oldLastPostAt := c1.LastPostAt

	o2 := model.Post{}
	o2.ChannelId = c1.Id
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.CreateAt = int64(7)
	require.Nil(t, (<-ss.Post().Save(&o2)).Err)

	c1, err = ss.Channel().Get(c1.Id, false)
	require.Nil(t, err)
	assert.Equal(t, oldLastPostAt, c1.LastPostAt, "LastPostAt should not update for old message save")
}

func testPostStoreGet(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"

	etag1 := (<-ss.Post().GetEtag(o1.ChannelId, false)).Data.(string)
	if strings.Index(etag1, model.CurrentVersion+".") != 0 {
		t.Fatal("Invalid Etag")
	}

	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	etag2 := (<-ss.Post().GetEtag(o1.ChannelId, false)).Data.(string)
	if strings.Index(etag2, fmt.Sprintf("%v.%v", model.CurrentVersion, o1.UpdateAt)) != 0 {
		t.Fatal("Invalid Etag")
	}

	r1, err := ss.Post().Get(o1.Id)
	if err != nil {
		t.Fatal(err)
	}
	if r1.Posts[o1.Id].CreateAt != o1.CreateAt {
		t.Fatal("invalid returned post")
	}

	if _, err = ss.Post().Get("123"); err == nil {
		t.Fatal("Missing id should have failed")
	}

	if _, err = ss.Post().Get(""); err == nil {
		t.Fatal("should fail for blank post ids")
	}
}

func testPostStoreGetSingle(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"

	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	if r1 := <-ss.Post().GetSingle(o1.Id); r1.Err != nil {
		t.Fatal(r1.Err)
	} else {
		if r1.Data.(*model.Post).CreateAt != o1.CreateAt {
			t.Fatal("invalid returned post")
		}
	}

	if err := (<-ss.Post().GetSingle("123")).Err; err == nil {
		t.Fatal("Missing id should have failed")
	}
}

func testGetEtagCache(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"

	etag1 := (<-ss.Post().GetEtag(o1.ChannelId, true)).Data.(string)
	if strings.Index(etag1, model.CurrentVersion+".") != 0 {
		t.Fatal("Invalid Etag")
	}

	// This one should come from the cache
	etag2 := (<-ss.Post().GetEtag(o1.ChannelId, true)).Data.(string)
	if strings.Index(etag2, model.CurrentVersion+".") != 0 {
		t.Fatal("Invalid Etag")
	}

	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	// We have not invalidated the cache so this should be the same as above
	etag3 := (<-ss.Post().GetEtag(o1.ChannelId, true)).Data.(string)
	if strings.Index(etag3, etag2) != 0 {
		t.Fatal("Invalid Etag")
	}

	ss.Post().InvalidateLastPostTimeCache(o1.ChannelId)

	// Invalidated cache so we should get a good result
	etag4 := (<-ss.Post().GetEtag(o1.ChannelId, true)).Data.(string)
	if strings.Index(etag4, fmt.Sprintf("%v.%v", model.CurrentVersion, o1.UpdateAt)) != 0 {
		t.Fatal("Invalid Etag")
	}
}

func testPostStoreUpdate(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "CCCCCCCCC"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "QQQQQQQQQQ"
	o3 = (<-ss.Post().Save(o3)).Data.(*model.Post)

	r1, err := ss.Post().Get(o1.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro1 := r1.Posts[o1.Id]
	r2, err := ss.Post().Get(o1.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro2 := r2.Posts[o2.Id]
	r3, err := ss.Post().Get(o3.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro3 := r3.Posts[o3.Id]

	if ro1.Message != o1.Message {
		t.Fatal("Failed to save/get")
	}

	o1a := &model.Post{}
	*o1a = *ro1
	o1a.Message = ro1.Message + "BBBBBBBBBB"
	if result := <-ss.Post().Update(o1a, ro1); result.Err != nil {
		t.Fatal(result.Err)
	}

	r1, err = ss.Post().Get(o1.Id)
	if err != nil {
		t.Fatal(err)
	}

	ro1a := r1.Posts[o1.Id]
	if ro1a.Message != o1a.Message {
		t.Fatal("Failed to update/get")
	}

	o2a := &model.Post{}
	*o2a = *ro2
	o2a.Message = ro2.Message + "DDDDDDD"
	if result := <-ss.Post().Update(o2a, ro2); result.Err != nil {
		t.Fatal(result.Err)
	}

	r2, err = ss.Post().Get(o1.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro2a := r2.Posts[o2.Id]

	if ro2a.Message != o2a.Message {
		t.Fatal("Failed to update/get")
	}

	o3a := &model.Post{}
	*o3a = *ro3
	o3a.Message = ro3.Message + "WWWWWWW"
	if result := <-ss.Post().Update(o3a, ro3); result.Err != nil {
		t.Fatal(result.Err)
	}

	r3, err = ss.Post().Get(o3.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro3a := r3.Posts[o3.Id]

	if ro3a.Message != o3a.Message && ro3a.Hashtags != o3a.Hashtags {
		t.Fatal("Failed to update/get")
	}

	o4 := store.Must(ss.Post().Save(&model.Post{
		ChannelId: model.NewId(),
		UserId:    model.NewId(),
		Message:   model.NewId(),
		Filenames: []string{"test"},
	})).(*model.Post)

	r4, err := ss.Post().Get(o4.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro4 := r4.Posts[o4.Id]

	o4a := &model.Post{}
	*o4a = *ro4
	o4a.Filenames = []string{}
	o4a.FileIds = []string{model.NewId()}
	if result := <-ss.Post().Update(o4a, ro4); result.Err != nil {
		t.Fatal(result.Err)
	}

	r4, err = ss.Post().Get(o4.Id)
	if err != nil {
		t.Fatal(err)
	}

	if ro4a := r4.Posts[o4.Id]; len(ro4a.Filenames) != 0 {
		t.Fatal("Failed to clear Filenames")
	} else if len(ro4a.FileIds) != 1 {
		t.Fatal("Failed to set FileIds")
	}
}

func testPostStoreDelete(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	deleteByID := model.NewId()

	etag1 := (<-ss.Post().GetEtag(o1.ChannelId, false)).Data.(string)
	if strings.Index(etag1, model.CurrentVersion+".") != 0 {
		t.Fatal("Invalid Etag")
	}

	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	if r1, err := ss.Post().Get(o1.Id); err != nil {
		t.Fatal(err)
	} else {
		if r1.Posts[o1.Id].CreateAt != o1.CreateAt {
			t.Fatal("invalid returned post")
		}
	}

	if err := ss.Post().Delete(o1.Id, model.GetMillis(), deleteByID); err != nil {
		t.Fatal(err)
	}

	r5 := <-ss.Post().GetPostsCreatedAt(o1.ChannelId, o1.CreateAt)
	post := r5.Data.([]*model.Post)[0]
	actual := post.Props[model.POST_PROPS_DELETE_BY]
	if actual != deleteByID {
		t.Errorf("Expected (*Post).Props[model.POST_PROPS_DELETE_BY] to be %v but got %v.", deleteByID, actual)
	}

	if r3, err := ss.Post().Get(o1.Id); err == nil {
		t.Log(r3)
		t.Fatal("Missing id should have failed")
	}

	etag2 := (<-ss.Post().GetEtag(o1.ChannelId, false)).Data.(string)
	if strings.Index(etag2, model.CurrentVersion+".") != 0 {
		t.Fatal("Invalid Etag")
	}
}

func testPostStoreDelete1Level(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)

	if err := ss.Post().Delete(o1.Id, model.GetMillis(), ""); err != nil {
		t.Fatal(err)
	}

	if _, err := ss.Post().Get(o1.Id); err == nil {
		t.Fatal("Deleted id should have failed")
	}

	if _, err := ss.Post().Get(o2.Id); err == nil {
		t.Fatal("Deleted id should have failed")
	}
}

func testPostStoreDelete2Level(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3.ParentId = o2.Id
	o3.RootId = o1.Id
	o3 = (<-ss.Post().Save(o3)).Data.(*model.Post)

	o4 := &model.Post{}
	o4.ChannelId = model.NewId()
	o4.UserId = model.NewId()
	o4.Message = "zz" + model.NewId() + "b"
	o4 = (<-ss.Post().Save(o4)).Data.(*model.Post)

	if err := ss.Post().Delete(o1.Id, model.GetMillis(), ""); err != nil {
		t.Fatal(err)
	}

	if _, err := ss.Post().Get(o1.Id); err == nil {
		t.Fatal("Deleted id should have failed")
	}

	if _, err := ss.Post().Get(o2.Id); err == nil {
		t.Fatal("Deleted id should have failed")
	}

	if _, err := ss.Post().Get(o3.Id); err == nil {
		t.Fatal("Deleted id should have failed")
	}

	if _, err := ss.Post().Get(o4.Id); err != nil {
		t.Fatal(err)
	}
}

func testPostStorePermDelete1Level(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)

	o3 := &model.Post{}
	o3.ChannelId = model.NewId()
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3 = (<-ss.Post().Save(o3)).Data.(*model.Post)

	if r2 := <-ss.Post().PermanentDeleteByUser(o2.UserId); r2.Err != nil {
		t.Fatal(r2.Err)
	}

	if _, err := ss.Post().Get(o1.Id); err != nil {
		t.Fatal("Deleted id shouldn't have failed")
	}

	if _, err := ss.Post().Get(o2.Id); err == nil {
		t.Fatal("Deleted id should have failed")
	}

	if r2 := <-ss.Post().PermanentDeleteByChannel(o3.ChannelId); r2.Err != nil {
		t.Fatal(r2.Err)
	}

	if _, err := ss.Post().Get(o3.Id); err == nil {
		t.Fatal("Deleted id should have failed")
	}
}

func testPostStorePermDelete1Level2(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)

	o3 := &model.Post{}
	o3.ChannelId = model.NewId()
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3 = (<-ss.Post().Save(o3)).Data.(*model.Post)

	if r2 := <-ss.Post().PermanentDeleteByUser(o1.UserId); r2.Err != nil {
		t.Fatal(r2.Err)
	}

	if _, err := ss.Post().Get(o1.Id); err == nil {
		t.Fatal("Deleted id should have failed")
	}

	if _, err := ss.Post().Get(o2.Id); err == nil {
		t.Fatal("Deleted id should have failed")
	}

	if _, err := ss.Post().Get(o3.Id); err != nil {
		t.Fatal("Deleted id shouldn't have failed")
	}
}

func testPostStoreGetWithChildren(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3.ParentId = o2.Id
	o3.RootId = o1.Id
	o3 = (<-ss.Post().Save(o3)).Data.(*model.Post)

	pl, err := ss.Post().Get(o1.Id)
	if err != nil {
		t.Fatal(err)
	}

	if len(pl.Posts) != 3 {
		t.Fatal("invalid returned post")
	}

	if dErr := ss.Post().Delete(o3.Id, model.GetMillis(), ""); dErr != nil {
		t.Fatal(dErr)
	}

	pl, err = ss.Post().Get(o1.Id)
	if err != nil {
		t.Fatal(err)
	}

	if len(pl.Posts) != 2 {
		t.Fatal("invalid returned post")
	}

	if dErr := ss.Post().Delete(o2.Id, model.GetMillis(), ""); dErr != nil {
		t.Fatal(dErr)
	}

	pl, err = ss.Post().Get(o1.Id)
	if err != nil {
		t.Fatal(err)
	}

	if len(pl.Posts) != 1 {
		t.Fatal("invalid returned post")
	}
}

func testPostStoreGetPostsWithDetails(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	_ = (<-ss.Post().Save(o2)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o2a := &model.Post{}
	o2a.ChannelId = o1.ChannelId
	o2a.UserId = model.NewId()
	o2a.Message = "zz" + model.NewId() + "b"
	o2a.ParentId = o1.Id
	o2a.RootId = o1.Id
	o2a = (<-ss.Post().Save(o2a)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3.ParentId = o1.Id
	o3.RootId = o1.Id
	o3 = (<-ss.Post().Save(o3)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o4 := &model.Post{}
	o4.ChannelId = o1.ChannelId
	o4.UserId = model.NewId()
	o4.Message = "zz" + model.NewId() + "b"
	o4 = (<-ss.Post().Save(o4)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o5 := &model.Post{}
	o5.ChannelId = o1.ChannelId
	o5.UserId = model.NewId()
	o5.Message = "zz" + model.NewId() + "b"
	o5.ParentId = o4.Id
	o5.RootId = o4.Id
	o5 = (<-ss.Post().Save(o5)).Data.(*model.Post)

	r1 := (<-ss.Post().GetPosts(o1.ChannelId, 0, 4, false)).Data.(*model.PostList)

	if r1.Order[0] != o5.Id {
		t.Fatal("invalid order")
	}

	if r1.Order[1] != o4.Id {
		t.Fatal("invalid order")
	}

	if r1.Order[2] != o3.Id {
		t.Fatal("invalid order")
	}

	if r1.Order[3] != o2a.Id {
		t.Fatal("invalid order")
	}

	if len(r1.Posts) != 6 { //the last 4, + o1 (o2a and o3's parent) + o2 (in same thread as o2a and o3)
		t.Fatal("wrong size")
	}

	if r1.Posts[o1.Id].Message != o1.Message {
		t.Fatal("Missing parent")
	}

	r2 := (<-ss.Post().GetPosts(o1.ChannelId, 0, 4, true)).Data.(*model.PostList)

	if r2.Order[0] != o5.Id {
		t.Fatal("invalid order")
	}

	if r2.Order[1] != o4.Id {
		t.Fatal("invalid order")
	}

	if r2.Order[2] != o3.Id {
		t.Fatal("invalid order")
	}

	if r2.Order[3] != o2a.Id {
		t.Fatal("invalid order")
	}

	if len(r2.Posts) != 6 { //the last 4, + o1 (o2a and o3's parent) + o2 (in same thread as o2a and o3)
		t.Fatal("wrong size")
	}

	if r2.Posts[o1.Id].Message != o1.Message {
		t.Fatal("Missing parent")
	}

	// Run once to fill cache
	<-ss.Post().GetPosts(o1.ChannelId, 0, 30, true)

	o6 := &model.Post{}
	o6.ChannelId = o1.ChannelId
	o6.UserId = model.NewId()
	o6.Message = "zz" + model.NewId() + "b"
	_ = (<-ss.Post().Save(o6)).Data.(*model.Post)

	// Should only be 6 since we hit the cache
	r3 := (<-ss.Post().GetPosts(o1.ChannelId, 0, 30, true)).Data.(*model.PostList)
	assert.Equal(t, 6, len(r3.Order))

	ss.Post().InvalidateLastPostTimeCache(o1.ChannelId)

	// Cache was invalidated, we should get all the posts
	r4 := (<-ss.Post().GetPosts(o1.ChannelId, 0, 30, true)).Data.(*model.PostList)
	assert.Equal(t, 7, len(r4.Order))
}

func testPostStoreGetPostsBeforeAfter(t *testing.T, ss store.Store) {
	t.Run("without threads", func(t *testing.T) {
		channelId := model.NewId()
		userId := model.NewId()

		var posts []*model.Post
		for i := 0; i < 10; i++ {
			post := store.Must(ss.Post().Save(&model.Post{
				ChannelId: channelId,
				UserId:    userId,
				Message:   "message",
			})).(*model.Post)

			posts = append(posts, post)

			time.Sleep(time.Millisecond)
		}

		t.Run("should not return anything before the first post", func(t *testing.T) {
			res := <-ss.Post().GetPostsBefore(channelId, posts[0].Id, 10, 0)
			assert.Nil(t, res.Err)

			postList := res.Data.(*model.PostList)
			assert.Equal(t, []string{}, postList.Order)
			assert.Equal(t, map[string]*model.Post{}, postList.Posts)
		})

		t.Run("should return posts before a post", func(t *testing.T) {
			res := <-ss.Post().GetPostsBefore(channelId, posts[5].Id, 10, 0)
			assert.Nil(t, res.Err)

			postList := res.Data.(*model.PostList)
			assert.Equal(t, []string{posts[4].Id, posts[3].Id, posts[2].Id, posts[1].Id, posts[0].Id}, postList.Order)
			assert.Equal(t, map[string]*model.Post{
				posts[0].Id: posts[0],
				posts[1].Id: posts[1],
				posts[2].Id: posts[2],
				posts[3].Id: posts[3],
				posts[4].Id: posts[4],
			}, postList.Posts)
		})

		t.Run("should limit posts before", func(t *testing.T) {
			res := <-ss.Post().GetPostsBefore(channelId, posts[5].Id, 2, 0)
			assert.Nil(t, res.Err)

			postList := res.Data.(*model.PostList)
			assert.Equal(t, []string{posts[4].Id, posts[3].Id}, postList.Order)
			assert.Equal(t, map[string]*model.Post{
				posts[3].Id: posts[3],
				posts[4].Id: posts[4],
			}, postList.Posts)
		})

		t.Run("should not return anything after the last post", func(t *testing.T) {
			res := <-ss.Post().GetPostsAfter(channelId, posts[len(posts)-1].Id, 10, 0)
			assert.Nil(t, res.Err)

			postList := res.Data.(*model.PostList)
			assert.Equal(t, []string{}, postList.Order)
			assert.Equal(t, map[string]*model.Post{}, postList.Posts)
		})

		t.Run("should return posts after a post", func(t *testing.T) {
			res := <-ss.Post().GetPostsAfter(channelId, posts[5].Id, 10, 0)
			assert.Nil(t, res.Err)

			postList := res.Data.(*model.PostList)
			assert.Equal(t, []string{posts[9].Id, posts[8].Id, posts[7].Id, posts[6].Id}, postList.Order)
			assert.Equal(t, map[string]*model.Post{
				posts[6].Id: posts[6],
				posts[7].Id: posts[7],
				posts[8].Id: posts[8],
				posts[9].Id: posts[9],
			}, postList.Posts)
		})

		t.Run("should limit posts after", func(t *testing.T) {
			res := <-ss.Post().GetPostsAfter(channelId, posts[5].Id, 2, 0)
			assert.Nil(t, res.Err)

			postList := res.Data.(*model.PostList)
			assert.Equal(t, []string{posts[7].Id, posts[6].Id}, postList.Order)
			assert.Equal(t, map[string]*model.Post{
				posts[6].Id: posts[6],
				posts[7].Id: posts[7],
			}, postList.Posts)
		})
	})

	t.Run("with threads", func(t *testing.T) {
		channelId := model.NewId()
		userId := model.NewId()

		// This creates a series of posts that looks like:
		// post1
		// post2
		// post3 (in response to post1)
		// post4 (in response to post2)
		// post5
		// post6 (in response to post2)

		post1 := store.Must(ss.Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
		})).(*model.Post)
		time.Sleep(time.Millisecond)

		post2 := store.Must(ss.Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
		})).(*model.Post)
		time.Sleep(time.Millisecond)

		post3 := store.Must(ss.Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			ParentId:  post1.Id,
			RootId:    post1.Id,
			Message:   "message",
		})).(*model.Post)
		time.Sleep(time.Millisecond)

		post4 := store.Must(ss.Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			RootId:    post2.Id,
			ParentId:  post2.Id,
			Message:   "message",
		})).(*model.Post)
		time.Sleep(time.Millisecond)

		post5 := store.Must(ss.Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
		})).(*model.Post)
		time.Sleep(time.Millisecond)

		post6 := store.Must(ss.Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			ParentId:  post2.Id,
			RootId:    post2.Id,
			Message:   "message",
		})).(*model.Post)

		// Adding a post to a thread changes the UpdateAt timestamp of the parent post
		post1.UpdateAt = post3.UpdateAt
		post2.UpdateAt = post6.UpdateAt

		t.Run("should return each post and thread before a post", func(t *testing.T) {
			res := <-ss.Post().GetPostsBefore(channelId, post4.Id, 2, 0)
			assert.Nil(t, res.Err)

			postList := res.Data.(*model.PostList)
			assert.Equal(t, []string{post3.Id, post2.Id}, postList.Order)
			assert.Equal(t, map[string]*model.Post{
				post1.Id: post1,
				post2.Id: post2,
				post3.Id: post3,
				post4.Id: post4,
				post6.Id: post6,
			}, postList.Posts)
		})

		t.Run("should return each post and the root of each thread after a post", func(t *testing.T) {
			res := <-ss.Post().GetPostsAfter(channelId, post4.Id, 2, 0)
			assert.Nil(t, res.Err)

			postList := res.Data.(*model.PostList)
			assert.Equal(t, []string{post6.Id, post5.Id}, postList.Order)
			assert.Equal(t, map[string]*model.Post{
				post2.Id: post2,
				post4.Id: post4,
				post5.Id: post5,
				post6.Id: post6,
			}, postList.Posts)
		})
	})
}

func testPostStoreGetPostsSince(t *testing.T, ss store.Store) {
	o0 := &model.Post{}
	o0.ChannelId = model.NewId()
	o0.UserId = model.NewId()
	o0.Message = "zz" + model.NewId() + "b"
	_ = (<-ss.Post().Save(o0)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	_ = (<-ss.Post().Save(o2)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o2a := &model.Post{}
	o2a.ChannelId = o1.ChannelId
	o2a.UserId = model.NewId()
	o2a.Message = "zz" + model.NewId() + "b"
	o2a.ParentId = o1.Id
	o2a.RootId = o1.Id
	o2a = (<-ss.Post().Save(o2a)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3.ParentId = o1.Id
	o3.RootId = o1.Id
	o3 = (<-ss.Post().Save(o3)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o4 := &model.Post{}
	o4.ChannelId = o1.ChannelId
	o4.UserId = model.NewId()
	o4.Message = "zz" + model.NewId() + "b"
	o4 = (<-ss.Post().Save(o4)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o5 := &model.Post{}
	o5.ChannelId = o1.ChannelId
	o5.UserId = model.NewId()
	o5.Message = "zz" + model.NewId() + "b"
	o5.ParentId = o4.Id
	o5.RootId = o4.Id
	o5 = (<-ss.Post().Save(o5)).Data.(*model.Post)

	r1 := (<-ss.Post().GetPostsSince(o1.ChannelId, o1.CreateAt, false)).Data.(*model.PostList)

	if r1.Order[0] != o5.Id {
		t.Fatal("invalid order")
	}

	if r1.Order[1] != o4.Id {
		t.Fatal("invalid order")
	}

	if r1.Order[2] != o3.Id {
		t.Fatal("invalid order")
	}

	if r1.Order[3] != o2a.Id {
		t.Fatal("invalid order")
	}

	if len(r1.Posts) != 6 {
		t.Fatal("wrong size")
	}

	if r1.Posts[o1.Id].Message != o1.Message {
		t.Fatal("Missing parent")
	}

	r2 := (<-ss.Post().GetPostsSince(o1.ChannelId, o5.UpdateAt, true)).Data.(*model.PostList)

	if len(r2.Order) != 0 {
		t.Fatal("wrong size ", len(r2.Posts))
	}
}

func testPostStoreSearch(t *testing.T, ss store.Store) {
	teamId := model.NewId()
	userId := model.NewId()

	c1 := &model.Channel{}
	c1.TeamId = teamId
	c1.DisplayName = "Channel1"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	c1 = (<-ss.Channel().Save(c1, -1)).Data.(*model.Channel)

	m1 := model.ChannelMember{}
	m1.ChannelId = c1.Id
	m1.UserId = userId
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	store.Must(ss.Channel().SaveMember(&m1))

	c2 := &model.Channel{}
	c2.TeamId = teamId
	c2.DisplayName = "Channel1"
	c2.Name = "zz" + model.NewId() + "b"
	c2.Type = model.CHANNEL_OPEN
	c2 = (<-ss.Channel().Save(c2, -1)).Data.(*model.Channel)

	c3 := &model.Channel{}
	c3.TeamId = teamId
	c3.DisplayName = "Channel1"
	c3.Name = "zz" + model.NewId() + "b"
	c3.Type = model.CHANNEL_OPEN
	c3 = (<-ss.Channel().Save(c3, -1)).Data.(*model.Channel)
	ss.Channel().Delete(c3.Id, model.GetMillis())

	m3 := model.ChannelMember{}
	m3.ChannelId = c3.Id
	m3.UserId = userId
	m3.NotifyProps = model.GetDefaultChannelNotifyProps()
	store.Must(ss.Channel().SaveMember(&m3))

	o1 := &model.Post{}
	o1.ChannelId = c1.Id
	o1.UserId = model.NewId()
	o1.Message = "corey mattermost new york"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	o1a := &model.Post{}
	o1a.ChannelId = c1.Id
	o1a.UserId = model.NewId()
	o1a.Message = "corey mattermost new york"
	o1a.Type = model.POST_JOIN_CHANNEL
	_ = (<-ss.Post().Save(o1a)).Data.(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = c1.Id
	o2.UserId = model.NewId()
	o2.Message = "New Jersey is where John is from"
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)

	o3 := &model.Post{}
	o3.ChannelId = c2.Id
	o3.UserId = model.NewId()
	o3.Message = "New Jersey is where John is from corey new york"
	_ = (<-ss.Post().Save(o3)).Data.(*model.Post)

	o4 := &model.Post{}
	o4.ChannelId = c1.Id
	o4.UserId = model.NewId()
	o4.Hashtags = "#hashtag"
	o4.Message = "(message)blargh"
	o4 = (<-ss.Post().Save(o4)).Data.(*model.Post)

	o5 := &model.Post{}
	o5.ChannelId = c1.Id
	o5.UserId = model.NewId()
	o5.Hashtags = "#secret #howdy"
	o5 = (<-ss.Post().Save(o5)).Data.(*model.Post)

	o6 := &model.Post{}
	o6.ChannelId = c3.Id
	o6.UserId = model.NewId()
	o6.Hashtags = "#hashtag"
	o6 = (<-ss.Post().Save(o6)).Data.(*model.Post)

	o7 := &model.Post{}
	o7.ChannelId = c3.Id
	o7.UserId = model.NewId()
	o7.Message = "New Jersey is where John is from corey new york"
	o7 = (<-ss.Post().Save(o7)).Data.(*model.Post)

	o8 := &model.Post{}
	o8.ChannelId = c3.Id
	o8.UserId = model.NewId()
	o8.Message = "Deleted"
	o8 = (<-ss.Post().Save(o8)).Data.(*model.Post)

	tt := []struct {
		name                     string
		searchParams             *model.SearchParams
		extectedResultsCount     int
		expectedMessageResultIds []string
	}{
		{
			"normal-search-1",
			&model.SearchParams{Terms: "corey"},
			1,
			[]string{o1.Id},
		},
		{
			"normal-search-2",
			&model.SearchParams{Terms: "new"},
			2,
			[]string{o1.Id, o2.Id},
		},
		{
			"normal-search-3",
			&model.SearchParams{Terms: "john"},
			1,
			[]string{o2.Id},
		},
		{
			"wildcard-search",
			&model.SearchParams{Terms: "matter*"},
			1,
			[]string{o1.Id},
		},
		{
			"hashtag-search",
			&model.SearchParams{Terms: "#hashtag", IsHashtag: true},
			1,
			[]string{o4.Id},
		},
		{
			"hashtag-search-2",
			&model.SearchParams{Terms: "#secret", IsHashtag: true},
			1,
			[]string{o5.Id},
		},
		{
			"no-match-mention",
			&model.SearchParams{Terms: "@thisshouldmatchnothing", IsHashtag: true},
			0,
			[]string{},
		},
		{
			"no-results-search",
			&model.SearchParams{Terms: "mattermost jersey"},
			0,
			[]string{},
		},
		{
			"multiple-words-search",
			&model.SearchParams{Terms: "corey new york"},
			1,
			[]string{o1.Id},
		},
		{
			"multiple-wildcard-search",
			&model.SearchParams{Terms: "matter* jer*"},
			0,
			[]string{},
		},
		{
			"search-with-work-next-to-a-symbol",
			&model.SearchParams{Terms: "message blargh"},
			1,
			[]string{o4.Id},
		},
		{
			"search-with-or",
			&model.SearchParams{Terms: "Jersey corey", OrTerms: true},
			2,
			[]string{o1.Id, o2.Id},
		},
		{
			"search-with-or-and-deleted",
			&model.SearchParams{Terms: "Jersey corey", OrTerms: true, IncludeDeletedChannels: true},
			3,
			[]string{o1.Id, o2.Id, o7.Id},
		},
		{
			"search-hashtag-deleted",
			&model.SearchParams{Terms: "#hashtag", IsHashtag: true, IncludeDeletedChannels: true},
			2,
			[]string{o4.Id, o6.Id},
		},
		{
			"search-deleted-only",
			&model.SearchParams{Terms: "Deleted", IncludeDeletedChannels: true},
			1,
			[]string{o8.Id},
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := (<-ss.Post().Search(teamId, userId, tc.searchParams)).Data.(*model.PostList)
			require.Len(t, result.Order, tc.extectedResultsCount)
			for _, expectedMessageResultId := range tc.expectedMessageResultIds {
				assert.Contains(t, result.Order, expectedMessageResultId)
			}
		})
	}
}

func testUserCountsWithPostsByDay(t *testing.T, ss store.Store) {
	t1 := &model.Team{}
	t1.DisplayName = "DisplayName"
	t1.Name = "zz" + model.NewId() + "b"
	t1.Email = MakeEmail()
	t1.Type = model.TEAM_OPEN
	t1, err := ss.Team().Save(t1)
	require.Nil(t, err)

	c1 := &model.Channel{}
	c1.TeamId = t1.Id
	c1.DisplayName = "Channel2"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	c1 = store.Must(ss.Channel().Save(c1, -1)).(*model.Channel)

	o1 := &model.Post{}
	o1.ChannelId = c1.Id
	o1.UserId = model.NewId()
	o1.CreateAt = utils.MillisFromTime(utils.Yesterday())
	o1.Message = "zz" + model.NewId() + "b"
	o1 = store.Must(ss.Post().Save(o1)).(*model.Post)

	o1a := &model.Post{}
	o1a.ChannelId = c1.Id
	o1a.UserId = model.NewId()
	o1a.CreateAt = o1.CreateAt
	o1a.Message = "zz" + model.NewId() + "b"
	_ = store.Must(ss.Post().Save(o1a)).(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = c1.Id
	o2.UserId = model.NewId()
	o2.CreateAt = o1.CreateAt - (1000 * 60 * 60 * 24)
	o2.Message = "zz" + model.NewId() + "b"
	o2 = store.Must(ss.Post().Save(o2)).(*model.Post)

	o2a := &model.Post{}
	o2a.ChannelId = c1.Id
	o2a.UserId = o2.UserId
	o2a.CreateAt = o1.CreateAt - (1000 * 60 * 60 * 24)
	o2a.Message = "zz" + model.NewId() + "b"
	_ = store.Must(ss.Post().Save(o2a)).(*model.Post)

	if r1 := <-ss.Post().AnalyticsUserCountsWithPostsByDay(t1.Id); r1.Err != nil {
		t.Fatal(r1.Err)
	} else {
		row1 := r1.Data.(model.AnalyticsRows)[0]
		if row1.Value != 2 {
			t.Fatal("wrong value")
		}

		row2 := r1.Data.(model.AnalyticsRows)[1]
		if row2.Value != 1 {
			t.Fatal("wrong value")
		}
	}
}

func testPostCountsByDay(t *testing.T, ss store.Store) {
	t1 := &model.Team{}
	t1.DisplayName = "DisplayName"
	t1.Name = "zz" + model.NewId() + "b"
	t1.Email = MakeEmail()
	t1.Type = model.TEAM_OPEN
	t1, err := ss.Team().Save(t1)
	require.Nil(t, err)

	c1 := &model.Channel{}
	c1.TeamId = t1.Id
	c1.DisplayName = "Channel2"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	c1 = store.Must(ss.Channel().Save(c1, -1)).(*model.Channel)

	o1 := &model.Post{}
	o1.ChannelId = c1.Id
	o1.UserId = model.NewId()
	o1.CreateAt = utils.MillisFromTime(utils.Yesterday())
	o1.Message = "zz" + model.NewId() + "b"
	o1 = store.Must(ss.Post().Save(o1)).(*model.Post)

	o1a := &model.Post{}
	o1a.ChannelId = c1.Id
	o1a.UserId = model.NewId()
	o1a.CreateAt = o1.CreateAt
	o1a.Message = "zz" + model.NewId() + "b"
	_ = store.Must(ss.Post().Save(o1a)).(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = c1.Id
	o2.UserId = model.NewId()
	o2.CreateAt = o1.CreateAt - (1000 * 60 * 60 * 24 * 2)
	o2.Message = "zz" + model.NewId() + "b"
	o2 = store.Must(ss.Post().Save(o2)).(*model.Post)

	o2a := &model.Post{}
	o2a.ChannelId = c1.Id
	o2a.UserId = o2.UserId
	o2a.CreateAt = o1.CreateAt - (1000 * 60 * 60 * 24 * 2)
	o2a.Message = "zz" + model.NewId() + "b"
	_ = store.Must(ss.Post().Save(o2a)).(*model.Post)

	time.Sleep(1 * time.Second)

	if r1 := <-ss.Post().AnalyticsPostCountsByDay(t1.Id); r1.Err != nil {
		t.Fatal(r1.Err)
	} else {
		row1 := r1.Data.(model.AnalyticsRows)[0]
		if row1.Value != 2 {
			t.Fatal(row1)
		}

		row2 := r1.Data.(model.AnalyticsRows)[1]
		if row2.Value != 2 {
			t.Fatal("wrong value")
		}
	}

	if r1 := <-ss.Post().AnalyticsPostCount(t1.Id, false, false); r1.Err != nil {
		t.Fatal(r1.Err)
	} else {
		if r1.Data.(int64) != 4 {
			t.Fatal("wrong value")
		}
	}
}

func testPostStoreGetFlaggedPostsForTeam(t *testing.T, ss store.Store, s SqlSupplier) {
	c1 := &model.Channel{}
	c1.TeamId = model.NewId()
	c1.DisplayName = "Channel1"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	c1 = store.Must(ss.Channel().Save(c1, -1)).(*model.Channel)

	o1 := &model.Post{}
	o1.ChannelId = c1.Id
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3.DeleteAt = 1
	o3 = (<-ss.Post().Save(o3)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o4 := &model.Post{}
	o4.ChannelId = model.NewId()
	o4.UserId = model.NewId()
	o4.Message = "zz" + model.NewId() + "b"
	o4 = (<-ss.Post().Save(o4)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	c2 := &model.Channel{}
	c2.DisplayName = "DMChannel1"
	c2.Name = "zz" + model.NewId() + "b"
	c2.Type = model.CHANNEL_DIRECT

	m1 := &model.ChannelMember{}
	m1.ChannelId = c2.Id
	m1.UserId = o1.UserId
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()

	m2 := &model.ChannelMember{}
	m2.ChannelId = c2.Id
	m2.UserId = model.NewId()
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()

	c2, err := ss.Channel().SaveDirectChannel(c2, m1, m2)
	require.Nil(t, err)

	o5 := &model.Post{}
	o5.ChannelId = c2.Id
	o5.UserId = m2.UserId
	o5.Message = "zz" + model.NewId() + "b"
	o5 = (<-ss.Post().Save(o5)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	r1 := (<-ss.Post().GetFlaggedPosts(o1.ChannelId, 0, 2)).Data.(*model.PostList)

	if len(r1.Order) != 0 {
		t.Fatal("should be empty")
	}

	preferences := model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o1.Id,
			Value:    "true",
		},
	}

	store.Must(ss.Preference().Save(&preferences))

	r2 := (<-ss.Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 0, 2)).Data.(*model.PostList)

	if len(r2.Order) != 1 {
		t.Fatal("should have 1 post")
	}

	preferences = model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o2.Id,
			Value:    "true",
		},
	}

	store.Must(ss.Preference().Save(&preferences))

	r3 := (<-ss.Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 0, 1)).Data.(*model.PostList)

	if len(r3.Order) != 1 {
		t.Fatal("should have 1 post")
	}

	r3 = (<-ss.Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 1, 1)).Data.(*model.PostList)

	if len(r3.Order) != 1 {
		t.Fatal("should have 1 post")
	}

	r3 = (<-ss.Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 1000, 10)).Data.(*model.PostList)

	if len(r3.Order) != 0 {
		t.Fatal("should be empty")
	}

	r4 := (<-ss.Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 0, 2)).Data.(*model.PostList)

	if len(r4.Order) != 2 {
		t.Fatal("should have 2 posts")
	}

	preferences = model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o3.Id,
			Value:    "true",
		},
	}

	store.Must(ss.Preference().Save(&preferences))

	r4 = (<-ss.Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 0, 2)).Data.(*model.PostList)

	if len(r4.Order) != 2 {
		t.Fatal("should have 2 posts")
	}

	preferences = model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o4.Id,
			Value:    "true",
		},
	}
	store.Must(ss.Preference().Save(&preferences))

	r4 = (<-ss.Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 0, 2)).Data.(*model.PostList)

	if len(r4.Order) != 2 {
		t.Fatal("should have 2 posts")
	}

	r4 = (<-ss.Post().GetFlaggedPostsForTeam(o1.UserId, model.NewId(), 0, 2)).Data.(*model.PostList)

	if len(r4.Order) != 0 {
		t.Fatal("should have 0 posts")
	}

	preferences = model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o5.Id,
			Value:    "true",
		},
	}
	store.Must(ss.Preference().Save(&preferences))

	r4 = (<-ss.Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 0, 10)).Data.(*model.PostList)

	if len(r4.Order) != 3 {
		t.Fatal("should have 3 posts")
	}

	// Manually truncate Channels table until testlib can handle cleanups
	s.GetMaster().Exec("TRUNCATE Channels")
}

func testPostStoreGetFlaggedPosts(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3.DeleteAt = 1
	o3 = (<-ss.Post().Save(o3)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	r1 := (<-ss.Post().GetFlaggedPosts(o1.UserId, 0, 2)).Data.(*model.PostList)

	if len(r1.Order) != 0 {
		t.Fatal("should be empty")
	}

	preferences := model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o1.Id,
			Value:    "true",
		},
	}

	store.Must(ss.Preference().Save(&preferences))

	r2 := (<-ss.Post().GetFlaggedPosts(o1.UserId, 0, 2)).Data.(*model.PostList)

	if len(r2.Order) != 1 {
		t.Fatal("should have 1 post")
	}

	preferences = model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o2.Id,
			Value:    "true",
		},
	}

	store.Must(ss.Preference().Save(&preferences))

	r3 := (<-ss.Post().GetFlaggedPosts(o1.UserId, 0, 1)).Data.(*model.PostList)

	if len(r3.Order) != 1 {
		t.Fatal("should have 1 post")
	}

	r3 = (<-ss.Post().GetFlaggedPosts(o1.UserId, 1, 1)).Data.(*model.PostList)

	if len(r3.Order) != 1 {
		t.Fatal("should have 1 post")
	}

	r3 = (<-ss.Post().GetFlaggedPosts(o1.UserId, 1000, 10)).Data.(*model.PostList)

	if len(r3.Order) != 0 {
		t.Fatal("should be empty")
	}

	r4 := (<-ss.Post().GetFlaggedPosts(o1.UserId, 0, 2)).Data.(*model.PostList)

	if len(r4.Order) != 2 {
		t.Fatal("should have 2 posts")
	}

	preferences = model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o3.Id,
			Value:    "true",
		},
	}

	store.Must(ss.Preference().Save(&preferences))

	r4 = (<-ss.Post().GetFlaggedPosts(o1.UserId, 0, 2)).Data.(*model.PostList)

	if len(r4.Order) != 2 {
		t.Fatal("should have 2 posts")
	}
}

func testPostStoreGetFlaggedPostsForChannel(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	// deleted post
	o3 := &model.Post{}
	o3.ChannelId = model.NewId()
	o3.UserId = o1.ChannelId
	o3.Message = "zz" + model.NewId() + "b"
	o3.DeleteAt = 1
	o3 = (<-ss.Post().Save(o3)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	o4 := &model.Post{}
	o4.ChannelId = model.NewId()
	o4.UserId = model.NewId()
	o4.Message = "zz" + model.NewId() + "b"
	o4 = (<-ss.Post().Save(o4)).Data.(*model.Post)
	time.Sleep(2 * time.Millisecond)

	r := (<-ss.Post().GetFlaggedPostsForChannel(o1.UserId, o1.ChannelId, 0, 10)).Data.(*model.PostList)

	if len(r.Order) != 0 {
		t.Fatal("should be empty")
	}

	preference := model.Preference{
		UserId:   o1.UserId,
		Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
		Name:     o1.Id,
		Value:    "true",
	}

	store.Must(ss.Preference().Save(&model.Preferences{preference}))

	r = (<-ss.Post().GetFlaggedPostsForChannel(o1.UserId, o1.ChannelId, 0, 10)).Data.(*model.PostList)

	if len(r.Order) != 1 {
		t.Fatal("should have 1 post")
	}

	preference.Name = o2.Id
	store.Must(ss.Preference().Save(&model.Preferences{preference}))

	preference.Name = o3.Id
	store.Must(ss.Preference().Save(&model.Preferences{preference}))

	r = (<-ss.Post().GetFlaggedPostsForChannel(o1.UserId, o1.ChannelId, 0, 1)).Data.(*model.PostList)

	if len(r.Order) != 1 {
		t.Fatal("should have 1 post")
	}

	r = (<-ss.Post().GetFlaggedPostsForChannel(o1.UserId, o1.ChannelId, 1, 1)).Data.(*model.PostList)

	if len(r.Order) != 1 {
		t.Fatal("should have 1 post")
	}

	r = (<-ss.Post().GetFlaggedPostsForChannel(o1.UserId, o1.ChannelId, 1000, 10)).Data.(*model.PostList)

	if len(r.Order) != 0 {
		t.Fatal("should be empty")
	}

	r = (<-ss.Post().GetFlaggedPostsForChannel(o1.UserId, o1.ChannelId, 0, 10)).Data.(*model.PostList)

	if len(r.Order) != 2 {
		t.Fatal("should have 2 posts")
	}

	preference.Name = o4.Id
	store.Must(ss.Preference().Save(&model.Preferences{preference}))

	r = (<-ss.Post().GetFlaggedPostsForChannel(o1.UserId, o4.ChannelId, 0, 10)).Data.(*model.PostList)

	if len(r.Order) != 1 {
		t.Fatal("should have 1 post")
	}
}

func testPostStoreGetPostsCreatedAt(t *testing.T, ss store.Store) {
	createTime := model.GetMillis() + 1

	o0 := &model.Post{}
	o0.ChannelId = model.NewId()
	o0.UserId = model.NewId()
	o0.Message = "zz" + model.NewId() + "b"
	o0.CreateAt = createTime
	o0 = (<-ss.Post().Save(o0)).Data.(*model.Post)

	o1 := &model.Post{}
	o1.ChannelId = o0.ChannelId
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1.CreateAt = createTime
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2.CreateAt = createTime + 1
	_ = (<-ss.Post().Save(o2)).Data.(*model.Post)

	o3 := &model.Post{}
	o3.ChannelId = model.NewId()
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3.CreateAt = createTime
	_ = (<-ss.Post().Save(o3)).Data.(*model.Post)

	r1 := (<-ss.Post().GetPostsCreatedAt(o1.ChannelId, createTime)).Data.([]*model.Post)
	assert.Equal(t, 2, len(r1))
}

func testPostStoreOverwrite(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "CCCCCCCCC"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "QQQQQQQQQQ"
	o3 = (<-ss.Post().Save(o3)).Data.(*model.Post)

	r1, err := ss.Post().Get(o1.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro1 := r1.Posts[o1.Id]
	r2, err := ss.Post().Get(o1.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro2 := r2.Posts[o2.Id]
	r3, err := ss.Post().Get(o3.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro3 := r3.Posts[o3.Id]

	if ro1.Message != o1.Message {
		t.Fatal("Failed to save/get")
	}

	o1a := &model.Post{}
	*o1a = *ro1
	o1a.Message = ro1.Message + "BBBBBBBBBB"
	_, err = ss.Post().Overwrite(o1a)
	if err != nil {
		t.Fatal(err)
	}

	r1, err = ss.Post().Get(o1.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro1a := r1.Posts[o1.Id]

	if ro1a.Message != o1a.Message {
		t.Fatal("Failed to overwrite/get")
	}

	o2a := &model.Post{}
	*o2a = *ro2
	o2a.Message = ro2.Message + "DDDDDDD"
	_, err = ss.Post().Overwrite(o2a)
	if err != nil {
		t.Fatal(err)
	}

	r2, err = ss.Post().Get(o1.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro2a := r2.Posts[o2.Id]

	if ro2a.Message != o2a.Message {
		t.Fatal("Failed to overwrite/get")
	}

	o3a := &model.Post{}
	*o3a = *ro3
	o3a.Message = ro3.Message + "WWWWWWW"
	_, err = ss.Post().Overwrite(o3a)
	if err != nil {
		t.Fatal(err)
	}

	r3, err = ss.Post().Get(o3.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro3a := r3.Posts[o3.Id]

	if ro3a.Message != o3a.Message && ro3a.Hashtags != o3a.Hashtags {
		t.Fatal("Failed to overwrite/get")
	}

	o4 := store.Must(ss.Post().Save(&model.Post{
		ChannelId: model.NewId(),
		UserId:    model.NewId(),
		Message:   model.NewId(),
		Filenames: []string{"test"},
	})).(*model.Post)

	r4, err := ss.Post().Get(o4.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro4 := r4.Posts[o4.Id]

	o4a := &model.Post{}
	*o4a = *ro4
	o4a.Filenames = []string{}
	o4a.FileIds = []string{model.NewId()}
	_, err = ss.Post().Overwrite(o4a)
	if err != nil {
		t.Fatal(err)
	}

	r4, err = ss.Post().Get(o4.Id)
	if err != nil {
		t.Fatal(err)
	}

	ro4a := r4.Posts[o4.Id]

	if len(ro4a.Filenames) != 0 {
		t.Fatal("Failed to clear Filenames")
	} else if len(ro4a.FileIds) != 1 {
		t.Fatal("Failed to set FileIds")
	}
}

func testPostStoreGetPostsByIds(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "CCCCCCCCC"
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "QQQQQQQQQQ"
	o3 = (<-ss.Post().Save(o3)).Data.(*model.Post)

	r1, err := ss.Post().Get(o1.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro1 := r1.Posts[o1.Id]
	r2, err := ss.Post().Get(o2.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro2 := r2.Posts[o2.Id]
	r3, err := ss.Post().Get(o3.Id)
	if err != nil {
		t.Fatal(err)
	}
	ro3 := r3.Posts[o3.Id]

	postIds := []string{
		ro1.Id,
		ro2.Id,
		ro3.Id,
	}

	if ro4 := store.Must(ss.Post().GetPostsByIds(postIds)).([]*model.Post); len(ro4) != 3 {
		t.Fatalf("Expected 3 posts in results. Got %v", len(ro4))
	}

	if err := ss.Post().Delete(ro1.Id, model.GetMillis(), ""); err != nil {
		t.Fatal(err)
	}

	if ro5 := store.Must(ss.Post().GetPostsByIds(postIds)).([]*model.Post); len(ro5) != 3 {
		t.Fatalf("Expected 3 posts in results. Got %v", len(ro5))
	}
}

func testPostStoreGetPostsBatchForIndexing(t *testing.T, ss store.Store) {
	c1 := &model.Channel{}
	c1.TeamId = model.NewId()
	c1.DisplayName = "Channel1"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	c1 = (<-ss.Channel().Save(c1, -1)).Data.(*model.Channel)

	c2 := &model.Channel{}
	c2.TeamId = model.NewId()
	c2.DisplayName = "Channel2"
	c2.Name = "zz" + model.NewId() + "b"
	c2.Type = model.CHANNEL_OPEN
	c2 = (<-ss.Channel().Save(c2, -1)).Data.(*model.Channel)

	o1 := &model.Post{}
	o1.ChannelId = c1.Id
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = c2.Id
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "CCCCCCCCC"
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)

	o3 := &model.Post{}
	o3.ChannelId = c1.Id
	o3.UserId = model.NewId()
	o3.ParentId = o1.Id
	o3.RootId = o1.Id
	o3.Message = "zz" + model.NewId() + "QQQQQQQQQQ"
	o3 = (<-ss.Post().Save(o3)).Data.(*model.Post)

	if r := store.Must(ss.Post().GetPostsBatchForIndexing(o1.CreateAt, model.GetMillis()+100000, 100)).([]*model.PostForIndexing); len(r) != 3 {
		t.Fatalf("Expected 3 posts in results. Got %v", len(r))
	} else {
		for _, p := range r {
			if p.Id == o1.Id {
				if p.TeamId != c1.TeamId {
					t.Fatalf("Unexpected team ID")
				}
				if p.ParentCreateAt != nil {
					t.Fatalf("Unexpected parent create at")
				}
			} else if p.Id == o2.Id {
				if p.TeamId != c2.TeamId {
					t.Fatalf("Unexpected team ID")
				}
				if p.ParentCreateAt != nil {
					t.Fatalf("Unexpected parent create at")
				}
			} else if p.Id == o3.Id {
				if p.TeamId != c1.TeamId {
					t.Fatalf("Unexpected team ID")
				}
				if *p.ParentCreateAt != o1.CreateAt {
					t.Fatalf("Unexpected parent create at")
				}
			} else {
				t.Fatalf("unexpected post returned")
			}
		}
	}
}

func testPostStorePermanentDeleteBatch(t *testing.T, ss store.Store) {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o1.CreateAt = 1000
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = model.NewId()
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o2.CreateAt = 1000
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)

	o3 := &model.Post{}
	o3.ChannelId = model.NewId()
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o3.CreateAt = 100000
	o3 = (<-ss.Post().Save(o3)).Data.(*model.Post)

	store.Must(ss.Post().PermanentDeleteBatch(2000, 1000))

	if _, err := ss.Post().Get(o1.Id); err == nil {
		t.Fatalf("Should have not found post 1 after purge")
	}

	if _, err := ss.Post().Get(o2.Id); err == nil {
		t.Fatalf("Should have not found post 2 after purge")
	}

	if _, err := ss.Post().Get(o3.Id); err != nil {
		t.Fatalf("Should have not found post 3 after purge")
	}
}

func testPostStoreGetOldest(t *testing.T, ss store.Store) {
	o0 := &model.Post{}
	o0.ChannelId = model.NewId()
	o0.UserId = model.NewId()
	o0.Message = "zz" + model.NewId() + "b"
	o0.CreateAt = 3
	o0 = (<-ss.Post().Save(o0)).Data.(*model.Post)

	o1 := &model.Post{}
	o1.ChannelId = o0.Id
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1.CreateAt = 2
	o1 = (<-ss.Post().Save(o1)).Data.(*model.Post)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.CreateAt = 1
	o2 = (<-ss.Post().Save(o2)).Data.(*model.Post)

	r1 := (<-ss.Post().GetOldest()).Data.(*model.Post)

	assert.EqualValues(t, o2.Id, r1.Id)
}

func testGetMaxPostSize(t *testing.T, ss store.Store) {
	assert.Equal(t, model.POST_MESSAGE_MAX_RUNES_V2, ss.Post().GetMaxPostSize())
	assert.Equal(t, model.POST_MESSAGE_MAX_RUNES_V2, ss.Post().GetMaxPostSize())
}

func testPostStoreGetParentsForExportAfter(t *testing.T, ss store.Store) {
	t1 := model.Team{}
	t1.DisplayName = "Name"
	t1.Name = model.NewId()
	t1.Email = MakeEmail()
	t1.Type = model.TEAM_OPEN
	_, err := ss.Team().Save(&t1)
	require.Nil(t, err)

	c1 := model.Channel{}
	c1.TeamId = t1.Id
	c1.DisplayName = "Channel1"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	store.Must(ss.Channel().Save(&c1, -1))

	u1 := model.User{}
	u1.Username = model.NewId()
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	store.Must(ss.User().Save(&u1))

	p1 := &model.Post{}
	p1.ChannelId = c1.Id
	p1.UserId = u1.Id
	p1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	p1.CreateAt = 1000
	p1 = (<-ss.Post().Save(p1)).Data.(*model.Post)

	r1 := <-ss.Post().GetParentsForExportAfter(10000, strings.Repeat("0", 26))
	assert.Nil(t, r1.Err)
	d1 := r1.Data.([]*model.PostForExport)

	found := false
	for _, p := range d1 {
		if p.Id == p1.Id {
			found = true
			assert.Equal(t, p.Id, p1.Id)
			assert.Equal(t, p.Message, p1.Message)
			assert.Equal(t, p.Username, u1.Username)
			assert.Equal(t, p.TeamName, t1.Name)
			assert.Equal(t, p.ChannelName, c1.Name)
		}
	}
	assert.True(t, found)
}

func testPostStoreGetRepliesForExport(t *testing.T, ss store.Store) {
	t1 := model.Team{}
	t1.DisplayName = "Name"
	t1.Name = model.NewId()
	t1.Email = MakeEmail()
	t1.Type = model.TEAM_OPEN
	_, err := ss.Team().Save(&t1)
	require.Nil(t, err)

	c1 := model.Channel{}
	c1.TeamId = t1.Id
	c1.DisplayName = "Channel1"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	store.Must(ss.Channel().Save(&c1, -1))

	u1 := model.User{}
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	store.Must(ss.User().Save(&u1))

	p1 := &model.Post{}
	p1.ChannelId = c1.Id
	p1.UserId = u1.Id
	p1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	p1.CreateAt = 1000
	p1 = (<-ss.Post().Save(p1)).Data.(*model.Post)

	p2 := &model.Post{}
	p2.ChannelId = c1.Id
	p2.UserId = u1.Id
	p2.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	p2.CreateAt = 1001
	p2.ParentId = p1.Id
	p2.RootId = p1.Id
	p2 = (<-ss.Post().Save(p2)).Data.(*model.Post)

	r1 := <-ss.Post().GetRepliesForExport(p1.Id)
	assert.Nil(t, r1.Err)

	d1 := r1.Data.([]*model.ReplyForExport)
	assert.Len(t, d1, 1)

	reply1 := d1[0]
	assert.Equal(t, reply1.Id, p2.Id)
	assert.Equal(t, reply1.Message, p2.Message)
	assert.Equal(t, reply1.Username, u1.Username)

	// Checking whether replies by deleted user are exported
	u1.DeleteAt = 1002
	store.Must(ss.User().Update(&u1, false))

	r1 = <-ss.Post().GetRepliesForExport(p1.Id)
	assert.Nil(t, r1.Err)

	d1 = r1.Data.([]*model.ReplyForExport)
	assert.Len(t, d1, 1)

	reply1 = d1[0]
	assert.Equal(t, reply1.Id, p2.Id)
	assert.Equal(t, reply1.Message, p2.Message)
	assert.Equal(t, reply1.Username, u1.Username)

}

func testPostStoreGetDirectPostParentsForExportAfter(t *testing.T, ss store.Store, s SqlSupplier) {
	teamId := model.NewId()

	o1 := model.Channel{}
	o1.TeamId = teamId
	o1.DisplayName = "Name"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_DIRECT

	u1 := &model.User{}
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	store.Must(ss.User().Save(u1))
	store.Must(ss.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1))

	u2 := &model.User{}
	u2.Email = MakeEmail()
	u2.Nickname = model.NewId()
	store.Must(ss.User().Save(u2))
	store.Must(ss.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}, -1))

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = u1.Id
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()

	m2 := model.ChannelMember{}
	m2.ChannelId = o1.Id
	m2.UserId = u2.Id
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()

	ss.Channel().SaveDirectChannel(&o1, &m1, &m2)

	p1 := &model.Post{}
	p1.ChannelId = o1.Id
	p1.UserId = u1.Id
	p1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	p1.CreateAt = 1000
	p1 = (<-ss.Post().Save(p1)).Data.(*model.Post)

	r1 := <-ss.Post().GetDirectPostParentsForExportAfter(10000, strings.Repeat("0", 26))
	assert.Nil(t, r1.Err)
	d1 := r1.Data.([]*model.DirectPostForExport)

	assert.Equal(t, p1.Message, d1[0].Message)

	// Manually truncate Channels table until testlib can handle cleanups
	s.GetMaster().Exec("TRUNCATE Channels")
}

func testPostStoreGetDirectPostParentsForExportAfterDeleted(t *testing.T, ss store.Store, s SqlSupplier) {
	teamId := model.NewId()

	o1 := model.Channel{}
	o1.TeamId = teamId
	o1.DisplayName = "Name"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_DIRECT

	u1 := &model.User{}
	u1.DeleteAt = 1
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	store.Must(ss.User().Save(u1))
	store.Must(ss.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1))

	u2 := &model.User{}
	u2.DeleteAt = 1
	u2.Email = MakeEmail()
	u2.Nickname = model.NewId()
	store.Must(ss.User().Save(u2))
	store.Must(ss.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}, -1))

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = u1.Id
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()

	m2 := model.ChannelMember{}
	m2.ChannelId = o1.Id
	m2.UserId = u2.Id
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()

	ss.Channel().SaveDirectChannel(&o1, &m1, &m2)

	o1.DeleteAt = 1
	err := ss.Channel().SetDeleteAt(o1.Id, 1, 1)
	assert.Nil(t, err)

	p1 := &model.Post{}
	p1.ChannelId = o1.Id
	p1.UserId = u1.Id
	p1.Message = "zz" + model.NewId() + "BBBBBBBBBBBB"
	p1.CreateAt = 1000
	p1 = (<-ss.Post().Save(p1)).Data.(*model.Post)

	o1a := &model.Post{}
	*o1a = *p1
	o1a.DeleteAt = 1
	o1a.Message = p1.Message + "BBBBBBBBBB"
	if result := <-ss.Post().Update(o1a, p1); result.Err != nil {
		t.Fatal(result.Err)
	}

	r1 := <-ss.Post().GetDirectPostParentsForExportAfter(10000, strings.Repeat("0", 26))
	assert.Nil(t, r1.Err)
	d1 := r1.Data.([]*model.DirectPostForExport)

	assert.Equal(t, 0, len(d1))

	// Manually truncate Channels table until testlib can handle cleanups
	s.GetMaster().Exec("TRUNCATE Channels")
}

func testPostStoreGetDirectPostParentsForExportAfterBatched(t *testing.T, ss store.Store, s SqlSupplier) {
	teamId := model.NewId()

	o1 := model.Channel{}
	o1.TeamId = teamId
	o1.DisplayName = "Name"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_DIRECT

	var postIds []string
	for i := 0; i < 150; i++ {
		u1 := &model.User{}
		u1.Email = MakeEmail()
		u1.Nickname = model.NewId()
		store.Must(ss.User().Save(u1))
		store.Must(ss.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1))

		u2 := &model.User{}
		u2.Email = MakeEmail()
		u2.Nickname = model.NewId()
		store.Must(ss.User().Save(u2))
		store.Must(ss.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}, -1))

		m1 := model.ChannelMember{}
		m1.ChannelId = o1.Id
		m1.UserId = u1.Id
		m1.NotifyProps = model.GetDefaultChannelNotifyProps()

		m2 := model.ChannelMember{}
		m2.ChannelId = o1.Id
		m2.UserId = u2.Id
		m2.NotifyProps = model.GetDefaultChannelNotifyProps()

		ss.Channel().SaveDirectChannel(&o1, &m1, &m2)

		p1 := &model.Post{}
		p1.ChannelId = o1.Id
		p1.UserId = u1.Id
		p1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
		p1.CreateAt = 1000
		p1 = (<-ss.Post().Save(p1)).Data.(*model.Post)
		postIds = append(postIds, p1.Id)
	}
	sort.Slice(postIds, func(i, j int) bool { return postIds[i] < postIds[j] })

	// Get all posts
	r1 := <-ss.Post().GetDirectPostParentsForExportAfter(10000, strings.Repeat("0", 26))
	assert.Nil(t, r1.Err)
	d1 := r1.Data.([]*model.DirectPostForExport)
	assert.Equal(t, len(postIds), len(d1))
	var exportedPostIds []string
	for i := range d1 {
		exportedPostIds = append(exportedPostIds, d1[i].Id)
	}
	sort.Slice(exportedPostIds, func(i, j int) bool { return exportedPostIds[i] < exportedPostIds[j] })
	assert.ElementsMatch(t, postIds, exportedPostIds)

	// Get 100
	r1 = <-ss.Post().GetDirectPostParentsForExportAfter(100, strings.Repeat("0", 26))
	assert.Nil(t, r1.Err)
	d1 = r1.Data.([]*model.DirectPostForExport)
	assert.Equal(t, 100, len(d1))
	exportedPostIds = []string{}
	for i := range d1 {
		exportedPostIds = append(exportedPostIds, d1[i].Id)
	}
	sort.Slice(exportedPostIds, func(i, j int) bool { return exportedPostIds[i] < exportedPostIds[j] })
	assert.ElementsMatch(t, postIds[:100], exportedPostIds)

	// Manually truncate Channels table until testlib can handle cleanups
	s.GetMaster().Exec("TRUNCATE Channels")
}
