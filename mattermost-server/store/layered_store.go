// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package store

import (
	"context"

	"github.com/mattermost/mattermost-server/einterfaces"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
)

const (
	ENABLE_EXPERIMENTAL_REDIS = false
)

type LayeredStoreDatabaseLayer interface {
	LayeredStoreSupplier
	Store
}

type LayeredStore struct {
	TmpContext      context.Context
	ReactionStore   ReactionStore
	RoleStore       RoleStore
	SchemeStore     SchemeStore
	DatabaseLayer   LayeredStoreDatabaseLayer
	LocalCacheLayer *LocalCacheSupplier
	RedisLayer      *RedisSupplier
	LayerChainHead  LayeredStoreSupplier
	GroupStore      GroupStore
}

func NewLayeredStore(db LayeredStoreDatabaseLayer, metrics einterfaces.MetricsInterface, cluster einterfaces.ClusterInterface) Store {
	store := &LayeredStore{
		TmpContext:      context.TODO(),
		DatabaseLayer:   db,
		LocalCacheLayer: NewLocalCacheSupplier(metrics, cluster),
	}

	store.ReactionStore = &LayeredReactionStore{store}
	store.RoleStore = &LayeredRoleStore{store}
	store.SchemeStore = &LayeredSchemeStore{store}
	store.GroupStore = &LayeredGroupStore{store}

	// Setup the chain
	if ENABLE_EXPERIMENTAL_REDIS {
		mlog.Debug("Experimental redis enabled.")
		store.RedisLayer = NewRedisSupplier()
		store.RedisLayer.SetChainNext(store.DatabaseLayer)
		store.LayerChainHead = store.RedisLayer
	} else {
		store.LocalCacheLayer.SetChainNext(store.DatabaseLayer)
		store.LayerChainHead = store.LocalCacheLayer
	}

	return store
}

type QueryFunction func(LayeredStoreSupplier) *LayeredStoreSupplierResult

func (s *LayeredStore) RunQuery(queryFunction QueryFunction) StoreChannel {
	storeChannel := make(StoreChannel)

	go func() {
		result := queryFunction(s.LayerChainHead)
		storeChannel <- result.StoreResult
	}()

	return storeChannel
}

func (s *LayeredStore) Team() TeamStore {
	return s.DatabaseLayer.Team()
}

func (s *LayeredStore) Channel() ChannelStore {
	return s.DatabaseLayer.Channel()
}

func (s *LayeredStore) Post() PostStore {
	return s.DatabaseLayer.Post()
}

func (s *LayeredStore) User() UserStore {
	return s.DatabaseLayer.User()
}

func (s *LayeredStore) Bot() BotStore {
	return s.DatabaseLayer.Bot()
}

func (s *LayeredStore) Audit() AuditStore {
	return s.DatabaseLayer.Audit()
}

func (s *LayeredStore) ClusterDiscovery() ClusterDiscoveryStore {
	return s.DatabaseLayer.ClusterDiscovery()
}

func (s *LayeredStore) Compliance() ComplianceStore {
	return s.DatabaseLayer.Compliance()
}

func (s *LayeredStore) Session() SessionStore {
	return s.DatabaseLayer.Session()
}

func (s *LayeredStore) OAuth() OAuthStore {
	return s.DatabaseLayer.OAuth()
}

func (s *LayeredStore) System() SystemStore {
	return s.DatabaseLayer.System()
}

func (s *LayeredStore) Webhook() WebhookStore {
	return s.DatabaseLayer.Webhook()
}

func (s *LayeredStore) Command() CommandStore {
	return s.DatabaseLayer.Command()
}

func (s *LayeredStore) CommandWebhook() CommandWebhookStore {
	return s.DatabaseLayer.CommandWebhook()
}

func (s *LayeredStore) Preference() PreferenceStore {
	return s.DatabaseLayer.Preference()
}

func (s *LayeredStore) License() LicenseStore {
	return s.DatabaseLayer.License()
}

func (s *LayeredStore) Token() TokenStore {
	return s.DatabaseLayer.Token()
}

func (s *LayeredStore) Emoji() EmojiStore {
	return s.DatabaseLayer.Emoji()
}

func (s *LayeredStore) Status() StatusStore {
	return s.DatabaseLayer.Status()
}

func (s *LayeredStore) FileInfo() FileInfoStore {
	return s.DatabaseLayer.FileInfo()
}

func (s *LayeredStore) Reaction() ReactionStore {
	return s.ReactionStore
}

func (s *LayeredStore) Job() JobStore {
	return s.DatabaseLayer.Job()
}

func (s *LayeredStore) UserAccessToken() UserAccessTokenStore {
	return s.DatabaseLayer.UserAccessToken()
}

func (s *LayeredStore) ChannelMemberHistory() ChannelMemberHistoryStore {
	return s.DatabaseLayer.ChannelMemberHistory()
}

func (s *LayeredStore) Plugin() PluginStore {
	return s.DatabaseLayer.Plugin()
}

func (s *LayeredStore) Role() RoleStore {
	return s.RoleStore
}

func (s *LayeredStore) TermsOfService() TermsOfServiceStore {
	return s.DatabaseLayer.TermsOfService()
}

func (s *LayeredStore) UserTermsOfService() UserTermsOfServiceStore {
	return s.DatabaseLayer.UserTermsOfService()
}

func (s *LayeredStore) Scheme() SchemeStore {
	return s.SchemeStore
}

func (s *LayeredStore) Group() GroupStore {
	return s.GroupStore
}

func (s *LayeredStore) LinkMetadata() LinkMetadataStore {
	return s.DatabaseLayer.LinkMetadata()
}

func (s *LayeredStore) MarkSystemRanUnitTests() {
	s.DatabaseLayer.MarkSystemRanUnitTests()
}

func (s *LayeredStore) Close() {
	s.DatabaseLayer.Close()
}

func (s *LayeredStore) LockToMaster() {
	s.DatabaseLayer.LockToMaster()
}

func (s *LayeredStore) UnlockFromMaster() {
	s.DatabaseLayer.UnlockFromMaster()
}

func (s *LayeredStore) DropAllTables() {
	defer s.LocalCacheLayer.Invalidate()
	s.DatabaseLayer.DropAllTables()
}

func (s *LayeredStore) TotalMasterDbConnections() int {
	return s.DatabaseLayer.TotalMasterDbConnections()
}

func (s *LayeredStore) TotalReadDbConnections() int {
	return s.DatabaseLayer.TotalReadDbConnections()
}

func (s *LayeredStore) TotalSearchDbConnections() int {
	return s.DatabaseLayer.TotalSearchDbConnections()
}

type LayeredReactionStore struct {
	*LayeredStore
}

func (s *LayeredReactionStore) Save(reaction *model.Reaction) (*model.Reaction, *model.AppError) {
	return s.LayerChainHead.ReactionSave(s.TmpContext, reaction)
}

func (s *LayeredReactionStore) Delete(reaction *model.Reaction) (*model.Reaction, *model.AppError) {
	return s.LayerChainHead.ReactionDelete(s.TmpContext, reaction)
}

func (s *LayeredReactionStore) GetForPost(postId string, allowFromCache bool) ([]*model.Reaction, *model.AppError) {
	return s.LayerChainHead.ReactionGetForPost(s.TmpContext, postId)
}

func (s *LayeredReactionStore) BulkGetForPosts(postIds []string) ([]*model.Reaction, *model.AppError) {
	return s.LayerChainHead.ReactionsBulkGetForPosts(s.TmpContext, postIds)
}

func (s *LayeredReactionStore) DeleteAllWithEmojiName(emojiName string) *model.AppError {
	return s.LayerChainHead.ReactionDeleteAllWithEmojiName(s.TmpContext, emojiName)
}

func (s *LayeredReactionStore) PermanentDeleteBatch(endTime int64, limit int64) (int64, *model.AppError) {
	return s.LayerChainHead.ReactionPermanentDeleteBatch(s.TmpContext, endTime, limit)
}

type LayeredRoleStore struct {
	*LayeredStore
}

func (s *LayeredRoleStore) Save(role *model.Role) (*model.Role, *model.AppError) {
	return s.LayerChainHead.RoleSave(s.TmpContext, role)
}

func (s *LayeredRoleStore) Get(roleId string) (*model.Role, *model.AppError) {
	return s.LayerChainHead.RoleGet(s.TmpContext, roleId)
}

func (s *LayeredRoleStore) GetAll() ([]*model.Role, *model.AppError) {
	return s.LayerChainHead.RoleGetAll(s.TmpContext)
}

func (s *LayeredRoleStore) GetByName(name string) (*model.Role, *model.AppError) {
	return s.LayerChainHead.RoleGetByName(s.TmpContext, name)
}

func (s *LayeredRoleStore) GetByNames(names []string) ([]*model.Role, *model.AppError) {
	return s.LayerChainHead.RoleGetByNames(s.TmpContext, names)
}

func (s *LayeredRoleStore) Delete(roldId string) (*model.Role, *model.AppError) {
	return s.LayerChainHead.RoleDelete(s.TmpContext, roldId)
}

func (s *LayeredRoleStore) PermanentDeleteAll() *model.AppError {
	return s.LayerChainHead.RolePermanentDeleteAll(s.TmpContext)
}

type LayeredSchemeStore struct {
	*LayeredStore
}

func (s *LayeredSchemeStore) Save(scheme *model.Scheme) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.SchemeSave(s.TmpContext, scheme)
	})
}

func (s *LayeredSchemeStore) Get(schemeId string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.SchemeGet(s.TmpContext, schemeId)
	})
}

func (s *LayeredSchemeStore) GetByName(schemeName string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.SchemeGetByName(s.TmpContext, schemeName)
	})
}

func (s *LayeredSchemeStore) Delete(schemeId string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.SchemeDelete(s.TmpContext, schemeId)
	})
}

func (s *LayeredSchemeStore) GetAllPage(scope string, offset int, limit int) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.SchemeGetAllPage(s.TmpContext, scope, offset, limit)
	})
}

func (s *LayeredSchemeStore) PermanentDeleteAll() StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.SchemePermanentDeleteAll(s.TmpContext)
	})
}

type LayeredGroupStore struct {
	*LayeredStore
}

func (s *LayeredGroupStore) Create(group *model.Group) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupCreate(s.TmpContext, group)
	})
}

func (s *LayeredGroupStore) Get(groupID string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupGet(s.TmpContext, groupID)
	})
}

func (s *LayeredGroupStore) GetByRemoteID(remoteID string, groupSource model.GroupSource) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupGetByRemoteID(s.TmpContext, remoteID, groupSource)
	})
}

func (s *LayeredGroupStore) GetAllBySource(groupSource model.GroupSource) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupGetAllBySource(s.TmpContext, groupSource)
	})
}

func (s *LayeredGroupStore) Update(group *model.Group) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupUpdate(s.TmpContext, group)
	})
}

func (s *LayeredGroupStore) Delete(groupID string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupDelete(s.TmpContext, groupID)
	})
}

func (s *LayeredGroupStore) GetMemberUsers(groupID string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupGetMemberUsers(s.TmpContext, groupID)
	})
}

func (s *LayeredGroupStore) GetMemberUsersPage(groupID string, offset int, limit int) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupGetMemberUsersPage(s.TmpContext, groupID, offset, limit)
	})
}

func (s *LayeredGroupStore) GetMemberCount(groupID string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupGetMemberCount(s.TmpContext, groupID)
	})
}

func (s *LayeredGroupStore) CreateOrRestoreMember(groupID string, userID string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupCreateOrRestoreMember(s.TmpContext, groupID, userID)
	})
}

func (s *LayeredGroupStore) DeleteMember(groupID string, userID string) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupDeleteMember(s.TmpContext, groupID, userID)
	})
}

func (s *LayeredGroupStore) CreateGroupSyncable(groupSyncable *model.GroupSyncable) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupCreateGroupSyncable(s.TmpContext, groupSyncable)
	})
}

func (s *LayeredGroupStore) GetGroupSyncable(groupID string, syncableID string, syncableType model.GroupSyncableType) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupGetGroupSyncable(s.TmpContext, groupID, syncableID, syncableType)
	})
}

func (s *LayeredGroupStore) GetAllGroupSyncablesByGroupId(groupID string, syncableType model.GroupSyncableType) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupGetAllGroupSyncablesByGroup(s.TmpContext, groupID, syncableType)
	})
}

func (s *LayeredGroupStore) UpdateGroupSyncable(groupSyncable *model.GroupSyncable) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupUpdateGroupSyncable(s.TmpContext, groupSyncable)
	})
}

func (s *LayeredGroupStore) DeleteGroupSyncable(groupID string, syncableID string, syncableType model.GroupSyncableType) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GroupDeleteGroupSyncable(s.TmpContext, groupID, syncableID, syncableType)
	})
}

func (s *LayeredGroupStore) TeamMembersToAdd(since int64) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.TeamMembersToAdd(s.TmpContext, since)
	})
}

func (s *LayeredGroupStore) ChannelMembersToAdd(since int64) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.ChannelMembersToAdd(s.TmpContext, since)
	})
}

func (s *LayeredGroupStore) TeamMembersToRemove() StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.TeamMembersToRemove(s.TmpContext)
	})
}

func (s *LayeredGroupStore) ChannelMembersToRemove() StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.ChannelMembersToRemove(s.TmpContext)
	})
}

func (s *LayeredGroupStore) GetGroupsByChannel(channelId string, opts model.GroupSearchOpts) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GetGroupsByChannel(s.TmpContext, channelId, opts)
	})
}

func (s *LayeredGroupStore) CountGroupsByChannel(channelId string, opts model.GroupSearchOpts) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.CountGroupsByChannel(s.TmpContext, channelId, opts)
	})
}

func (s *LayeredGroupStore) GetGroupsByTeam(teamId string, opts model.GroupSearchOpts) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GetGroupsByTeam(s.TmpContext, teamId, opts)
	})
}

func (s *LayeredGroupStore) CountGroupsByTeam(teamId string, opts model.GroupSearchOpts) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.CountGroupsByTeam(s.TmpContext, teamId, opts)
	})
}

func (s *LayeredGroupStore) GetGroups(page, perPage int, opts model.GroupSearchOpts) StoreChannel {
	return s.RunQuery(func(supplier LayeredStoreSupplier) *LayeredStoreSupplierResult {
		return supplier.GetGroups(s.TmpContext, page, perPage, opts)
	})
}
