// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package store

import "github.com/mattermost/mattermost-server/model"
import "context"

type LayeredStoreSupplierResult struct {
	StoreResult
}

func NewSupplierResult() *LayeredStoreSupplierResult {
	return &LayeredStoreSupplierResult{}
}

type LayeredStoreSupplier interface {
	//
	// Control
	//
	SetChainNext(LayeredStoreSupplier)
	Next() LayeredStoreSupplier

	//
	// Reactions
	//), hints ...LayeredStoreHint)
	ReactionSave(ctx context.Context, reaction *model.Reaction, hints ...LayeredStoreHint) (*model.Reaction, *model.AppError)
	ReactionDelete(ctx context.Context, reaction *model.Reaction, hints ...LayeredStoreHint) (*model.Reaction, *model.AppError)
	ReactionGetForPost(ctx context.Context, postId string, hints ...LayeredStoreHint) ([]*model.Reaction, *model.AppError)
	ReactionDeleteAllWithEmojiName(ctx context.Context, emojiName string, hints ...LayeredStoreHint) *model.AppError
	ReactionPermanentDeleteBatch(ctx context.Context, endTime int64, limit int64, hints ...LayeredStoreHint) (int64, *model.AppError)
	ReactionsBulkGetForPosts(ctx context.Context, postIds []string, hints ...LayeredStoreHint) ([]*model.Reaction, *model.AppError)

	// Roles
	RoleSave(ctx context.Context, role *model.Role, hints ...LayeredStoreHint) (*model.Role, *model.AppError)
	RoleGet(ctx context.Context, roleId string, hints ...LayeredStoreHint) (*model.Role, *model.AppError)
	RoleGetAll(ctx context.Context, hints ...LayeredStoreHint) ([]*model.Role, *model.AppError)
	RoleGetByName(ctx context.Context, name string, hints ...LayeredStoreHint) (*model.Role, *model.AppError)
	RoleGetByNames(ctx context.Context, names []string, hints ...LayeredStoreHint) ([]*model.Role, *model.AppError)
	RoleDelete(ctx context.Context, roldId string, hints ...LayeredStoreHint) (*model.Role, *model.AppError)
	RolePermanentDeleteAll(ctx context.Context, hints ...LayeredStoreHint) *model.AppError

	// Schemes
	SchemeSave(ctx context.Context, scheme *model.Scheme, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	SchemeGet(ctx context.Context, schemeId string, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	SchemeGetByName(ctx context.Context, schemeName string, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	SchemeDelete(ctx context.Context, schemeId string, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	SchemeGetAllPage(ctx context.Context, scope string, offset int, limit int, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	SchemePermanentDeleteAll(ctx context.Context, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
}
