// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package app

import (
	"bytes"
	"image"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dyatlov/go-opengraph/opengraph"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/utils"
	"github.com/mattermost/mattermost-server/utils/imgutils"
	"github.com/mattermost/mattermost-server/utils/markdown"
)

const LINK_CACHE_SIZE = 10000
const LINK_CACHE_DURATION = 3600
const MaxMetadataImageSize = MaxOpenGraphResponseSize

var linkCache = utils.NewLru(LINK_CACHE_SIZE)

func (a *App) InitPostMetadata() {
	// Dump any cached links if the proxy settings have changed so image URLs can be updated
	a.AddConfigListener(func(before, after *model.Config) {
		if (before.ImageProxySettings.Enable != after.ImageProxySettings.Enable) ||
			(before.ImageProxySettings.ImageProxyType != after.ImageProxySettings.ImageProxyType) ||
			(before.ImageProxySettings.RemoteImageProxyURL != after.ImageProxySettings.RemoteImageProxyURL) ||
			(before.ImageProxySettings.RemoteImageProxyOptions != after.ImageProxySettings.RemoteImageProxyOptions) {
			linkCache.Purge()
		}
	})
}

func (a *App) PreparePostListForClient(originalList *model.PostList) *model.PostList {
	list := &model.PostList{
		Posts: make(map[string]*model.Post, len(originalList.Posts)),
		Order: originalList.Order, // Note that this uses the original Order array, so it isn't a deep copy
	}

	for id, originalPost := range originalList.Posts {
		post := a.PreparePostForClient(originalPost, false, false)

		list.Posts[id] = post
	}

	return list
}

func (a *App) PreparePostForClient(originalPost *model.Post, isNewPost bool, isEditPost bool) *model.Post {
	post := originalPost.Clone()

	// Proxy image links before constructing metadata so that requests go through the proxy
	post = a.PostWithProxyAddedToImageURLs(post)

	if *a.Config().ExperimentalSettings.DisablePostMetadata {
		return post
	}

	post.Metadata = &model.PostMetadata{}

	// Emojis and reaction counts
	if emojis, reactions, err := a.getEmojisAndReactionsForPost(post); err != nil {
		mlog.Warn("Failed to get emojis and reactions for a post", mlog.String("post_id", post.Id), mlog.Err(err))
	} else {
		post.Metadata.Emojis = emojis
		post.Metadata.Reactions = reactions
	}

	// Files
	if fileInfos, err := a.getFileMetadataForPost(post, isNewPost || isEditPost); err != nil {
		mlog.Warn("Failed to get files for a post", mlog.String("post_id", post.Id), mlog.Err(err))
	} else {
		post.Metadata.Files = fileInfos
	}

	// Embeds and image dimensions
	firstLink, images := getFirstLinkAndImages(post.Message)

	if embed, err := a.getEmbedForPost(post, firstLink, isNewPost); err != nil {
		mlog.Debug("Failed to get embedded content for a post", mlog.String("post_id", post.Id), mlog.Err(err))
	} else if embed == nil {
		post.Metadata.Embeds = []*model.PostEmbed{}
	} else {
		post.Metadata.Embeds = []*model.PostEmbed{embed}
	}

	post.Metadata.Images = a.getImagesForPost(post, images, isNewPost)

	return post
}

func (a *App) getFileMetadataForPost(post *model.Post, fromMaster bool) ([]*model.FileInfo, *model.AppError) {
	if len(post.FileIds) == 0 {
		return nil, nil
	}

	return a.GetFileInfosForPost(post.Id, fromMaster)
}

func (a *App) getEmojisAndReactionsForPost(post *model.Post) ([]*model.Emoji, []*model.Reaction, *model.AppError) {
	var reactions []*model.Reaction
	if post.HasReactions {
		var err *model.AppError
		reactions, err = a.GetReactionsForPost(post.Id)
		if err != nil {
			return nil, nil, err
		}
	}

	emojis, err := a.getCustomEmojisForPost(post, reactions)
	if err != nil {
		return nil, nil, err
	}

	return emojis, reactions, nil
}

func (a *App) getEmbedForPost(post *model.Post, firstLink string, isNewPost bool) (*model.PostEmbed, error) {
	if _, ok := post.Props["attachments"]; ok {
		return &model.PostEmbed{
			Type: model.POST_EMBED_MESSAGE_ATTACHMENT,
		}, nil
	}

	if firstLink == "" || !*a.Config().ServiceSettings.EnableLinkPreviews {
		return nil, nil
	}

	og, image, err := a.getLinkMetadata(firstLink, post.CreateAt, isNewPost)
	if err != nil {
		return nil, err
	}

	if og != nil {
		return &model.PostEmbed{
			Type: model.POST_EMBED_OPENGRAPH,
			URL:  firstLink,
			Data: og,
		}, nil
	}

	if image != nil {
		// Note that we're not passing the image info here since it'll be part of the PostMetadata.Images field
		return &model.PostEmbed{
			Type: model.POST_EMBED_IMAGE,
			URL:  firstLink,
		}, nil
	}

	return nil, nil
}

func (a *App) getImagesForPost(post *model.Post, imageURLs []string, isNewPost bool) map[string]*model.PostImage {
	images := map[string]*model.PostImage{}

	for _, embed := range post.Metadata.Embeds {
		switch embed.Type {
		case model.POST_EMBED_IMAGE:
			// These dimensions will generally be cached by a previous call to getEmbedForPost
			imageURLs = append(imageURLs, embed.URL)

		case model.POST_EMBED_MESSAGE_ATTACHMENT:
			imageURLs = append(imageURLs, getImagesInMessageAttachments(post)...)

		case model.POST_EMBED_OPENGRAPH:
			for _, image := range embed.Data.(*opengraph.OpenGraph).Images {
				var imageURL string
				if image.SecureURL != "" {
					imageURL = image.SecureURL
				} else if image.URL != "" {
					imageURL = image.URL
				}

				if imageURL == "" {
					continue
				}

				imageURLs = append(imageURLs, imageURL)
			}
		}
	}

	// Removing duplicates isn't strictly since images is a map, but it feels safer to do it beforehand
	if len(imageURLs) > 1 {
		imageURLs = model.RemoveDuplicateStrings(imageURLs)
	}

	for _, imageURL := range imageURLs {
		if _, image, err := a.getLinkMetadata(imageURL, post.CreateAt, isNewPost); err != nil {
			mlog.Debug("Failed to get dimensions of an image in a post",
				mlog.String("post_id", post.Id), mlog.String("image_url", imageURL), mlog.Err(err))
		} else if image != nil {
			images[imageURL] = image
		}
	}

	return images
}

func getEmojiNamesForString(s string) []string {
	names := model.EMOJI_PATTERN.FindAllString(s, -1)

	for i, name := range names {
		names[i] = strings.Trim(name, ":")
	}

	return names
}

func getEmojiNamesForPost(post *model.Post, reactions []*model.Reaction) []string {
	// Post message
	names := getEmojiNamesForString(post.Message)

	// Reactions
	for _, reaction := range reactions {
		names = append(names, reaction.EmojiName)
	}

	// Post attachments
	for _, attachment := range post.Attachments() {
		if attachment.Text != "" {
			names = append(names, getEmojiNamesForString(attachment.Text)...)
		}

		if attachment.Pretext != "" {
			names = append(names, getEmojiNamesForString(attachment.Pretext)...)
		}

		for _, field := range attachment.Fields {
			if value, ok := field.Value.(string); ok {
				names = append(names, getEmojiNamesForString(value)...)
			}
		}
	}

	// Remove duplicates
	names = model.RemoveDuplicateStrings(names)

	return names
}

func (a *App) getCustomEmojisForPost(post *model.Post, reactions []*model.Reaction) ([]*model.Emoji, *model.AppError) {
	if !*a.Config().ServiceSettings.EnableCustomEmoji {
		// Only custom emoji are returned
		return []*model.Emoji{}, nil
	}

	names := getEmojiNamesForPost(post, reactions)

	if len(names) == 0 {
		return []*model.Emoji{}, nil
	}

	return a.GetMultipleEmojiByName(names)
}

// Given a string, returns the first autolinked URL in the string as well as an array of all Markdown
// images of the form ![alt text](image url). Note that this does not return Markdown links of the
// form [text](url).
func getFirstLinkAndImages(str string) (string, []string) {
	firstLink := ""
	images := []string{}

	markdown.Inspect(str, func(blockOrInline interface{}) bool {
		switch v := blockOrInline.(type) {
		case *markdown.Autolink:
			if firstLink == "" {
				firstLink = v.Destination()
			}
		case *markdown.InlineImage:
			images = append(images, v.Destination())
		case *markdown.ReferenceImage:
			images = append(images, v.ReferenceDefinition.Destination())
		}

		return true
	})

	return firstLink, images
}

func getImagesInMessageAttachments(post *model.Post) []string {
	var images []string

	for _, attachment := range post.Attachments() {
		_, imagesInText := getFirstLinkAndImages(attachment.Text)
		images = append(images, imagesInText...)

		_, imagesInPretext := getFirstLinkAndImages(attachment.Pretext)
		images = append(images, imagesInPretext...)

		for _, field := range attachment.Fields {
			if value, ok := field.Value.(string); ok {
				_, imagesInFieldValue := getFirstLinkAndImages(value)
				images = append(images, imagesInFieldValue...)
			}
		}

		if attachment.AuthorIcon != "" {
			images = append(images, attachment.AuthorIcon)
		}

		if attachment.ImageURL != "" {
			images = append(images, attachment.ImageURL)
		}

		if attachment.ThumbURL != "" {
			images = append(images, attachment.ThumbURL)
		}

		if attachment.FooterIcon != "" {
			images = append(images, attachment.FooterIcon)
		}
	}

	return images
}

func (a *App) getLinkMetadata(requestURL string, timestamp int64, isNewPost bool) (*opengraph.OpenGraph, *model.PostImage, error) {
	requestURL = resolveMetadataURL(requestURL, a.GetSiteURL())

	timestamp = model.FloorToNearestHour(timestamp)

	// Check cache
	og, image, ok := getLinkMetadataFromCache(requestURL, timestamp)
	if ok {
		return og, image, nil
	}

	// Check the database if this isn't a new post. If it is a new post and the data is cached, it should be in memory.
	if !isNewPost {
		og, image, ok = a.getLinkMetadataFromDatabase(requestURL, timestamp)
		if ok {
			cacheLinkMetadata(requestURL, timestamp, og, image)

			return og, image, nil
		}
	}

	// Make request for a web page or an image
	request, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, nil, err
	}

	var body io.ReadCloser
	var contentType string

	if (request.URL.Scheme+"://"+request.URL.Host) == a.GetSiteURL() && request.URL.Path == "/api/v4/image" {
		// /api/v4/image requires authentication, so bypass the API by hitting the proxy directly
		body, contentType, err = a.ImageProxy.GetImageDirect(a.ImageProxy.GetUnproxiedImageURL(request.URL.String()))
	} else {
		request.Header.Add("Accept", "image/*, text/html")

		client := a.HTTPService.MakeClient(false)
		client.Timeout = time.Duration(*a.Config().ExperimentalSettings.LinkMetadataTimeoutMilliseconds) * time.Millisecond

		var res *http.Response
		res, err = client.Do(request)

		if res != nil {
			body = res.Body
			contentType = res.Header.Get("Content-Type")
		}
	}

	if body != nil {
		defer body.Close()
	}

	if err == nil {
		// Parse the data
		og, image, err = a.parseLinkMetadata(requestURL, body, contentType)
	}

	// Write back to cache and database, even if there was an error and the results are nil
	cacheLinkMetadata(requestURL, timestamp, og, image)

	a.saveLinkMetadataToDatabase(requestURL, timestamp, og, image)

	return og, image, err
}

// resolveMetadataURL resolves a given URL relative to the server's site URL.
func resolveMetadataURL(requestURL string, siteURL string) string {
	base, err := url.Parse(siteURL)
	if err != nil {
		return ""
	}

	resolved, err := base.Parse(requestURL)
	if err != nil {
		return ""
	}

	return resolved.String()
}

func getLinkMetadataFromCache(requestURL string, timestamp int64) (*opengraph.OpenGraph, *model.PostImage, bool) {
	cached, ok := linkCache.Get(model.GenerateLinkMetadataHash(requestURL, timestamp))
	if !ok {
		return nil, nil, false
	}

	switch v := cached.(type) {
	case *opengraph.OpenGraph:
		return v, nil, true
	case *model.PostImage:
		return nil, v, true
	default:
		return nil, nil, true
	}
}

func (a *App) getLinkMetadataFromDatabase(requestURL string, timestamp int64) (*opengraph.OpenGraph, *model.PostImage, bool) {
	result := <-a.Srv.Store.LinkMetadata().Get(requestURL, timestamp)
	if result.Err != nil {
		return nil, nil, false
	}

	data := result.Data.(*model.LinkMetadata).Data

	switch v := data.(type) {
	case *opengraph.OpenGraph:
		return v, nil, true
	case *model.PostImage:
		return nil, v, true
	default:
		return nil, nil, true
	}
}

func (a *App) saveLinkMetadataToDatabase(requestURL string, timestamp int64, og *opengraph.OpenGraph, image *model.PostImage) {
	metadata := &model.LinkMetadata{
		URL:       requestURL,
		Timestamp: timestamp,
	}

	if og != nil {
		metadata.Type = model.LINK_METADATA_TYPE_OPENGRAPH
		metadata.Data = og
	} else if image != nil {
		metadata.Type = model.LINK_METADATA_TYPE_IMAGE
		metadata.Data = image
	} else {
		metadata.Type = model.LINK_METADATA_TYPE_NONE
	}

	result := <-a.Srv.Store.LinkMetadata().Save(metadata)
	if result.Err != nil {
		mlog.Warn("Failed to write link metadata", mlog.String("request_url", requestURL), mlog.Err(result.Err))
	}
}

func cacheLinkMetadata(requestURL string, timestamp int64, og *opengraph.OpenGraph, image *model.PostImage) {
	var val interface{}
	if og != nil {
		val = og
	} else if image != nil {
		val = image
	}

	linkCache.AddWithExpiresInSecs(model.GenerateLinkMetadataHash(requestURL, timestamp), val, LINK_CACHE_DURATION)
}

func (a *App) parseLinkMetadata(requestURL string, body io.Reader, contentType string) (*opengraph.OpenGraph, *model.PostImage, error) {
	if strings.HasPrefix(contentType, "image") {
		image, err := parseImages(io.LimitReader(body, MaxMetadataImageSize))
		return nil, image, err
	} else if strings.HasPrefix(contentType, "text/html") {
		og := a.ParseOpenGraphMetadata(requestURL, body, contentType)

		// The OpenGraph library and Go HTML library don't error for malformed input, so check that at least
		// one of these required fields exists before returning the OpenGraph data
		if og.Title != "" || og.Type != "" || og.URL != "" {
			return og, nil, nil
		} else {
			return nil, nil, nil
		}
	} else {
		// Not an image or web page with OpenGraph information
		return nil, nil, nil
	}
}

func parseImages(body io.Reader) (*model.PostImage, error) {
	// Store any data that is read for the config for any further processing
	buf := &bytes.Buffer{}
	t := io.TeeReader(body, buf)

	// Read the image config to get the format and dimensions
	config, format, err := image.DecodeConfig(t)
	if err != nil {
		return nil, err
	}

	image := &model.PostImage{
		Width:  config.Width,
		Height: config.Height,
		Format: format,
	}

	if format == "gif" {
		// Decoding the config may have read some of the image data, so re-read the data that has already been read first
		frameCount, err := imgutils.CountFrames(io.MultiReader(buf, body))
		if err != nil {
			return nil, err
		}

		image.FrameCount = frameCount
	}

	return image, nil
}
