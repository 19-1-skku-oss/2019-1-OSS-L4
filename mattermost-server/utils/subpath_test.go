package utils_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/utils"
)

func TestUpdateAssetsSubpath(t *testing.T) {
	t.Run("no client dir", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "test_update_assets_subpath")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)
		os.Chdir(tempDir)

		err = utils.UpdateAssetsSubpath("/")
		require.Error(t, err)
	})

	t.Run("valid", func(t *testing.T) {
		tempDir, err := ioutil.TempDir("", "test_update_assets_subpath")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)
		os.Chdir(tempDir)

		err = os.Mkdir(model.CLIENT_DIR, 0700)
		require.NoError(t, err)

		testCases := []struct {
			Description          string
			RootHTML             string
			MainCSS              string
			ManifestJSON         string
			Subpath              string
			ExpectedError        error
			ExpectedRootHTML     string
			ExpectedMainCSS      string
			ExpectedManifestJSON string
		}{
			{
				"no changes required, empty subpath provided",
				baseRootHtml,
				baseCss,
				baseManifestJson,
				"",
				nil,
				baseRootHtml,
				baseCss,
				baseManifestJson,
			},
			{
				"no changes required",
				baseRootHtml,
				baseCss,
				baseManifestJson,
				"/",
				nil,
				baseRootHtml,
				baseCss,
				baseManifestJson,
			},
			{
				"content security policy not found (missing quotes)",
				contentSecurityPolicyNotFoundHtml,
				baseCss,
				baseManifestJson,
				"/subpath",
				fmt.Errorf("failed to find 'Content-Security-Policy' meta tag to rewrite"),
				contentSecurityPolicyNotFoundHtml,
				baseCss,
				baseManifestJson,
			},
			{
				"content security policy not found (missing unsafe-eval)",
				contentSecurityPolicyNotFound2Html,
				baseCss,
				baseManifestJson,
				"/subpath",
				fmt.Errorf("failed to find 'Content-Security-Policy' meta tag to rewrite"),
				contentSecurityPolicyNotFound2Html,
				baseCss,
				baseManifestJson,
			},
			{
				"subpath",
				baseRootHtml,
				baseCss,
				baseManifestJson,
				"/subpath",
				nil,
				subpathRootHtml,
				subpathCss,
				subpathManifestJson,
			},
			{
				"new subpath from old",
				subpathRootHtml,
				subpathCss,
				subpathManifestJson,
				"/nested/subpath",
				nil,
				newSubpathRootHtml,
				newSubpathCss,
				newSubpathManifestJson,
			},
			{
				"resetting to /",
				subpathRootHtml,
				subpathCss,
				baseManifestJson,
				"/",
				nil,
				baseRootHtml,
				baseCss,
				baseManifestJson,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Description, func(t *testing.T) {
				ioutil.WriteFile(filepath.Join(tempDir, model.CLIENT_DIR, "root.html"), []byte(testCase.RootHTML), 0700)
				ioutil.WriteFile(filepath.Join(tempDir, model.CLIENT_DIR, "main.css"), []byte(testCase.MainCSS), 0700)
				ioutil.WriteFile(filepath.Join(tempDir, model.CLIENT_DIR, "manifest.json"), []byte(testCase.ManifestJSON), 0700)
				err := utils.UpdateAssetsSubpath(testCase.Subpath)
				if testCase.ExpectedError != nil {
					require.Equal(t, testCase.ExpectedError, err)
				} else {
					require.NoError(t, err)
				}

				contents, err := ioutil.ReadFile(filepath.Join(tempDir, model.CLIENT_DIR, "root.html"))
				require.NoError(t, err)

				// Rewrite the expected and contents for simpler diffs when failed.
				expectedRootHTML := strings.Replace(testCase.ExpectedRootHTML, ">", ">\n", -1)
				contentsStr := strings.Replace(string(contents), ">", ">\n", -1)
				require.Equal(t, expectedRootHTML, contentsStr)

				contents, err = ioutil.ReadFile(filepath.Join(tempDir, model.CLIENT_DIR, "main.css"))
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedMainCSS, string(contents))

				contents, err = ioutil.ReadFile(filepath.Join(tempDir, model.CLIENT_DIR, "manifest.json"))
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedManifestJSON, string(contents))
			})
		}
	})
}

func TestGetSubpathFromConfig(t *testing.T) {
	sToP := func(s string) *string {
		return &s
	}

	testCases := []struct {
		Description     string
		SiteURL         *string
		ExpectedError   bool
		ExpectedSubpath string
	}{
		{
			"empty SiteURL",
			sToP(""),
			false,
			"/",
		},
		{
			"invalid SiteURL",
			sToP("cache_object:foo/bar"),
			true,
			"",
		},
		{
			"nil SiteURL",
			nil,
			false,
			"/",
		},
		{
			"no trailing slash",
			sToP("http://localhost:8065"),
			false,
			"/",
		},
		{
			"trailing slash",
			sToP("http://localhost:8065/"),
			false,
			"/",
		},
		{
			"subpath, no trailing slash",
			sToP("http://localhost:8065/subpath"),
			false,
			"/subpath",
		},
		{
			"trailing slash",
			sToP("http://localhost:8065/subpath/"),
			false,
			"/subpath",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			config := &model.Config{
				ServiceSettings: model.ServiceSettings{
					SiteURL: testCase.SiteURL,
				},
			}

			subpath, err := utils.GetSubpathFromConfig(config)
			if testCase.ExpectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, testCase.ExpectedSubpath, subpath)
		})
	}
}

const contentSecurityPolicyNotFoundHtml = `<!DOCTYPE html> <html lang=en> <head> <meta charset=utf-8> <meta http-equiv=Content-Security-Policy content="script-src 'self' cdn.segment.com/analytics.js/"> <meta http-equiv=X-UA-Compatible content="IE=edge"> <meta name=viewport content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=0"> <meta name=robots content="noindex, nofollow"> <meta name=referrer content=no-referrer> <title>Mattermost</title> <meta name=apple-mobile-web-app-capable content=yes> <meta name=apple-mobile-web-app-status-bar-style content=default> <meta name=mobile-web-app-capable content=yes> <meta name=apple-mobile-web-app-title content=Mattermost> <meta name=application-name content=Mattermost> <meta name=format-detection content="telephone=no"> <link rel=apple-touch-icon sizes=57x57 href=/static/files/78b7e73b41b8731ce2c41c870ecc8886.png> <link rel=apple-touch-icon sizes=60x60 href=/static/files/51d00ffd13afb6d74fd8f6dfdeef768a.png> <link rel=apple-touch-icon sizes=72x72 href=/static/files/23645596f8f78f017bd4d457abb855c4.png> <link rel=apple-touch-icon sizes=76x76 href=/static/files/26e9d72f472663a00b4b206149459fab.png> <link rel=apple-touch-icon sizes=144x144 href=/static/files/7bd91659bf3fc8c68fcd45fc1db9c630.png> <link rel=apple-touch-icon sizes=120x120 href=/static/files/fa69ffe11eb334aaef5aece8d848ca62.png> <link rel=apple-touch-icon sizes=152x152 href=/static/files/f046777feb6ab12fc43b8f9908b1db35.png> <link rel=icon type=image/png sizes=16x16 href=/static/files/02b96247d275680adaaabf01c71c571d.png> <link rel=icon type=image/png sizes=32x32 href=/static/files/1d9020f201a6762421cab8d30624fdd8.png> <link rel=icon type=image/png sizes=96x96 href=/static/files/fe23af39ae98d77dc26ae8586565970f.png> <link rel=icon type=image/png sizes=192x192 href=/static/files/d7ff68a7675f84337cc154c3d4abe713.png> <link rel=manifest href=/static/files/a985ad72552ad069537d6eea81e719c7.json> <link rel=stylesheet class=code_theme> <style>.error-screen{font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;padding-top:50px;max-width:750px;font-size:14px;color:#333;margin:auto;display:none;line-height:1.5}.error-screen h2{font-size:30px;font-weight:400;line-height:1.2}.error-screen ul{padding-left:15px;line-height:1.7;margin-top:0;margin-bottom:10px}.error-screen hr{color:#ddd;margin-top:20px;margin-bottom:20px;border:0;border-top:1px solid #eee}.error-screen-visible{display:block}</style> <link href="/static/main.364fd054d7a6d741efc6.css" rel="stylesheet"><script type="text/javascript" src="/static/main.e49599ac425584ffead5.js"></script></head> <body class=font--open_sans> <div id=root> <div class=error-screen> <h2>Cannot connect to Mattermost</h2> <hr/> <p>We're having trouble connecting to Mattermost. If refreshing this page (Ctrl+R or Command+R) does not work, please verify that your computer is connected to the internet.</p> <br/> </div> <div class=loading-screen style=position:relative> <div class=loading__content> <div class="round round-1"></div> <div class="round round-2"></div> <div class="round round-3"></div> </div> </div> </div> <noscript> To use Mattermost, please enable JavaScript. </noscript> </body> </html>`

const contentSecurityPolicyNotFound2Html = `<!DOCTYPE html> <html lang=en> <head> <meta charset=utf-8> <meta http-equiv=Content-Security-Policy content="script-src 'self' cdn.segment.com/analytics.js/ 'unsafe-eval'"> <meta http-equiv=X-UA-Compatible content="IE=edge"> <meta name=viewport content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=0"> <meta name=robots content="noindex, nofollow"> <meta name=referrer content=no-referrer> <title>Mattermost</title> <meta name=apple-mobile-web-app-capable content=yes> <meta name=apple-mobile-web-app-status-bar-style content=default> <meta name=mobile-web-app-capable content=yes> <meta name=apple-mobile-web-app-title content=Mattermost> <meta name=application-name content=Mattermost> <meta name=format-detection content="telephone=no"> <link rel=apple-touch-icon sizes=57x57 href=/static/files/78b7e73b41b8731ce2c41c870ecc8886.png> <link rel=apple-touch-icon sizes=60x60 href=/static/files/51d00ffd13afb6d74fd8f6dfdeef768a.png> <link rel=apple-touch-icon sizes=72x72 href=/static/files/23645596f8f78f017bd4d457abb855c4.png> <link rel=apple-touch-icon sizes=76x76 href=/static/files/26e9d72f472663a00b4b206149459fab.png> <link rel=apple-touch-icon sizes=144x144 href=/static/files/7bd91659bf3fc8c68fcd45fc1db9c630.png> <link rel=apple-touch-icon sizes=120x120 href=/static/files/fa69ffe11eb334aaef5aece8d848ca62.png> <link rel=apple-touch-icon sizes=152x152 href=/static/files/f046777feb6ab12fc43b8f9908b1db35.png> <link rel=icon type=image/png sizes=16x16 href=/static/files/02b96247d275680adaaabf01c71c571d.png> <link rel=icon type=image/png sizes=32x32 href=/static/files/1d9020f201a6762421cab8d30624fdd8.png> <link rel=icon type=image/png sizes=96x96 href=/static/files/fe23af39ae98d77dc26ae8586565970f.png> <link rel=icon type=image/png sizes=192x192 href=/static/files/d7ff68a7675f84337cc154c3d4abe713.png> <link rel=manifest href=/static/files/a985ad72552ad069537d6eea81e719c7.json> <link rel=stylesheet class=code_theme> <style>.error-screen{font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;padding-top:50px;max-width:750px;font-size:14px;color:#333;margin:auto;display:none;line-height:1.5}.error-screen h2{font-size:30px;font-weight:400;line-height:1.2}.error-screen ul{padding-left:15px;line-height:1.7;margin-top:0;margin-bottom:10px}.error-screen hr{color:#ddd;margin-top:20px;margin-bottom:20px;border:0;border-top:1px solid #eee}.error-screen-visible{display:block}</style> <link href="/static/main.364fd054d7a6d741efc6.css" rel="stylesheet"><script type="text/javascript" src="/static/main.e49599ac425584ffead5.js"></script></head> <body class=font--open_sans> <div id=root> <div class=error-screen> <h2>Cannot connect to Mattermost</h2> <hr/> <p>We're having trouble connecting to Mattermost. If refreshing this page (Ctrl+R or Command+R) does not work, please verify that your computer is connected to the internet.</p> <br/> </div> <div class=loading-screen style=position:relative> <div class=loading__content> <div class="round round-1"></div> <div class="round round-2"></div> <div class="round round-3"></div> </div> </div> </div> <noscript> To use Mattermost, please enable JavaScript. </noscript> </body> </html>`

const baseRootHtml = `<!DOCTYPE html> <html lang=en> <head> <meta charset=utf-8> <meta http-equiv="Content-Security-Policy" content="script-src 'self' cdn.segment.com/analytics.js/"> <meta http-equiv=X-UA-Compatible content="IE=edge"> <meta name=viewport content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=0"> <meta name=robots content="noindex, nofollow"> <meta name=referrer content=no-referrer> <title>Mattermost</title> <meta name=apple-mobile-web-app-capable content=yes> <meta name=apple-mobile-web-app-status-bar-style content=default> <meta name=mobile-web-app-capable content=yes> <meta name=apple-mobile-web-app-title content=Mattermost> <meta name=application-name content=Mattermost> <meta name=format-detection content="telephone=no"> <link rel=apple-touch-icon sizes=57x57 href=/static/files/78b7e73b41b8731ce2c41c870ecc8886.png> <link rel=apple-touch-icon sizes=60x60 href=/static/files/51d00ffd13afb6d74fd8f6dfdeef768a.png> <link rel=apple-touch-icon sizes=72x72 href=/static/files/23645596f8f78f017bd4d457abb855c4.png> <link rel=apple-touch-icon sizes=76x76 href=/static/files/26e9d72f472663a00b4b206149459fab.png> <link rel=apple-touch-icon sizes=144x144 href=/static/files/7bd91659bf3fc8c68fcd45fc1db9c630.png> <link rel=apple-touch-icon sizes=120x120 href=/static/files/fa69ffe11eb334aaef5aece8d848ca62.png> <link rel=apple-touch-icon sizes=152x152 href=/static/files/f046777feb6ab12fc43b8f9908b1db35.png> <link rel=icon type=image/png sizes=16x16 href=/static/files/02b96247d275680adaaabf01c71c571d.png> <link rel=icon type=image/png sizes=32x32 href=/static/files/1d9020f201a6762421cab8d30624fdd8.png> <link rel=icon type=image/png sizes=96x96 href=/static/files/fe23af39ae98d77dc26ae8586565970f.png> <link rel=icon type=image/png sizes=192x192 href=/static/files/d7ff68a7675f84337cc154c3d4abe713.png> <link rel=manifest href=/static/files/a985ad72552ad069537d6eea81e719c7.json> <link rel=stylesheet class=code_theme> <style>.error-screen{font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;padding-top:50px;max-width:750px;font-size:14px;color:#333;margin:auto;display:none;line-height:1.5}.error-screen h2{font-size:30px;font-weight:400;line-height:1.2}.error-screen ul{padding-left:15px;line-height:1.7;margin-top:0;margin-bottom:10px}.error-screen hr{color:#ddd;margin-top:20px;margin-bottom:20px;border:0;border-top:1px solid #eee}.error-screen-visible{display:block}</style> <link href="/static/main.364fd054d7a6d741efc6.css" rel="stylesheet"><script type="text/javascript" src="/static/main.e49599ac425584ffead5.js"></script></head> <body class=font--open_sans> <div id=root> <div class=error-screen> <h2>Cannot connect to Mattermost</h2> <hr/> <p>We're having trouble connecting to Mattermost. If refreshing this page (Ctrl+R or Command+R) does not work, please verify that your computer is connected to the internet.</p> <br/> </div> <div class=loading-screen style=position:relative> <div class=loading__content> <div class="round round-1"></div> <div class="round round-2"></div> <div class="round round-3"></div> </div> </div> </div> <noscript> To use Mattermost, please enable JavaScript. </noscript> </body> </html>`

const baseCss = `@font-face{font-family:FontAwesome;src:url(/static/files/674f50d287a8c48dc19ba404d20fe713.eot);src:url(/static/files/674f50d287a8c48dc19ba404d20fe713.eot?#iefix&v=4.7.0) format("embedded-opentype"),url(/static/files/af7ae505a9eed503f8b8e6982036873e.woff2) format("woff2"),url(/static/files/fee66e712a8a08eef5805a46892932ad.woff) format("woff"),url(/static/files/b06871f281fee6b241d60582ae9369b9.ttf) format("truetype"),url(/static/files/677433a0892aaed7b7d2628c313c9775.svg#fontawesomeregular) format("svg");font-weight:400;font-style:normal}`

const subpathRootHtml = `<!DOCTYPE html> <html lang=en> <head> <meta charset=utf-8> <meta http-equiv="Content-Security-Policy" content="script-src 'self' cdn.segment.com/analytics.js/ 'sha256-tPOjw+tkVs9axL78ZwGtYl975dtyPHB6LYKAO2R3gR4='"> <meta http-equiv=X-UA-Compatible content="IE=edge"> <meta name=viewport content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=0"> <meta name=robots content="noindex, nofollow"> <meta name=referrer content=no-referrer> <title>Mattermost</title> <meta name=apple-mobile-web-app-capable content=yes> <meta name=apple-mobile-web-app-status-bar-style content=default> <meta name=mobile-web-app-capable content=yes> <meta name=apple-mobile-web-app-title content=Mattermost> <meta name=application-name content=Mattermost> <meta name=format-detection content="telephone=no"> <link rel=apple-touch-icon sizes=57x57 href=/subpath/static/files/78b7e73b41b8731ce2c41c870ecc8886.png> <link rel=apple-touch-icon sizes=60x60 href=/subpath/static/files/51d00ffd13afb6d74fd8f6dfdeef768a.png> <link rel=apple-touch-icon sizes=72x72 href=/subpath/static/files/23645596f8f78f017bd4d457abb855c4.png> <link rel=apple-touch-icon sizes=76x76 href=/subpath/static/files/26e9d72f472663a00b4b206149459fab.png> <link rel=apple-touch-icon sizes=144x144 href=/subpath/static/files/7bd91659bf3fc8c68fcd45fc1db9c630.png> <link rel=apple-touch-icon sizes=120x120 href=/subpath/static/files/fa69ffe11eb334aaef5aece8d848ca62.png> <link rel=apple-touch-icon sizes=152x152 href=/subpath/static/files/f046777feb6ab12fc43b8f9908b1db35.png> <link rel=icon type=image/png sizes=16x16 href=/subpath/static/files/02b96247d275680adaaabf01c71c571d.png> <link rel=icon type=image/png sizes=32x32 href=/subpath/static/files/1d9020f201a6762421cab8d30624fdd8.png> <link rel=icon type=image/png sizes=96x96 href=/subpath/static/files/fe23af39ae98d77dc26ae8586565970f.png> <link rel=icon type=image/png sizes=192x192 href=/subpath/static/files/d7ff68a7675f84337cc154c3d4abe713.png> <link rel=manifest href=/subpath/static/files/a985ad72552ad069537d6eea81e719c7.json> <link rel=stylesheet class=code_theme> <style>.error-screen{font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;padding-top:50px;max-width:750px;font-size:14px;color:#333;margin:auto;display:none;line-height:1.5}.error-screen h2{font-size:30px;font-weight:400;line-height:1.2}.error-screen ul{padding-left:15px;line-height:1.7;margin-top:0;margin-bottom:10px}.error-screen hr{color:#ddd;margin-top:20px;margin-bottom:20px;border:0;border-top:1px solid #eee}.error-screen-visible{display:block}</style><script>window.publicPath='/subpath/static/'</script> <link href="/subpath/static/main.364fd054d7a6d741efc6.css" rel="stylesheet"><script type="text/javascript" src="/subpath/static/main.e49599ac425584ffead5.js"></script></head> <body class=font--open_sans> <div id=root> <div class=error-screen> <h2>Cannot connect to Mattermost</h2> <hr/> <p>We're having trouble connecting to Mattermost. If refreshing this page (Ctrl+R or Command+R) does not work, please verify that your computer is connected to the internet.</p> <br/> </div> <div class=loading-screen style=position:relative> <div class=loading__content> <div class="round round-1"></div> <div class="round round-2"></div> <div class="round round-3"></div> </div> </div> </div> <noscript> To use Mattermost, please enable JavaScript. </noscript> </body> </html>`

const subpathCss = `@font-face{font-family:FontAwesome;src:url(/subpath/static/files/674f50d287a8c48dc19ba404d20fe713.eot);src:url(/subpath/static/files/674f50d287a8c48dc19ba404d20fe713.eot?#iefix&v=4.7.0) format("embedded-opentype"),url(/subpath/static/files/af7ae505a9eed503f8b8e6982036873e.woff2) format("woff2"),url(/subpath/static/files/fee66e712a8a08eef5805a46892932ad.woff) format("woff"),url(/subpath/static/files/b06871f281fee6b241d60582ae9369b9.ttf) format("truetype"),url(/subpath/static/files/677433a0892aaed7b7d2628c313c9775.svg#fontawesomeregular) format("svg");font-weight:400;font-style:normal}`

const newSubpathRootHtml = `<!DOCTYPE html> <html lang=en> <head> <meta charset=utf-8> <meta http-equiv="Content-Security-Policy" content="script-src 'self' cdn.segment.com/analytics.js/ 'sha256-mbRaPRRpWz6MNkX9SyXWMJ8XnWV4w/DoqK2M0ryUAvc='"> <meta http-equiv=X-UA-Compatible content="IE=edge"> <meta name=viewport content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=0"> <meta name=robots content="noindex, nofollow"> <meta name=referrer content=no-referrer> <title>Mattermost</title> <meta name=apple-mobile-web-app-capable content=yes> <meta name=apple-mobile-web-app-status-bar-style content=default> <meta name=mobile-web-app-capable content=yes> <meta name=apple-mobile-web-app-title content=Mattermost> <meta name=application-name content=Mattermost> <meta name=format-detection content="telephone=no"> <link rel=apple-touch-icon sizes=57x57 href=/nested/subpath/static/files/78b7e73b41b8731ce2c41c870ecc8886.png> <link rel=apple-touch-icon sizes=60x60 href=/nested/subpath/static/files/51d00ffd13afb6d74fd8f6dfdeef768a.png> <link rel=apple-touch-icon sizes=72x72 href=/nested/subpath/static/files/23645596f8f78f017bd4d457abb855c4.png> <link rel=apple-touch-icon sizes=76x76 href=/nested/subpath/static/files/26e9d72f472663a00b4b206149459fab.png> <link rel=apple-touch-icon sizes=144x144 href=/nested/subpath/static/files/7bd91659bf3fc8c68fcd45fc1db9c630.png> <link rel=apple-touch-icon sizes=120x120 href=/nested/subpath/static/files/fa69ffe11eb334aaef5aece8d848ca62.png> <link rel=apple-touch-icon sizes=152x152 href=/nested/subpath/static/files/f046777feb6ab12fc43b8f9908b1db35.png> <link rel=icon type=image/png sizes=16x16 href=/nested/subpath/static/files/02b96247d275680adaaabf01c71c571d.png> <link rel=icon type=image/png sizes=32x32 href=/nested/subpath/static/files/1d9020f201a6762421cab8d30624fdd8.png> <link rel=icon type=image/png sizes=96x96 href=/nested/subpath/static/files/fe23af39ae98d77dc26ae8586565970f.png> <link rel=icon type=image/png sizes=192x192 href=/nested/subpath/static/files/d7ff68a7675f84337cc154c3d4abe713.png> <link rel=manifest href=/nested/subpath/static/files/a985ad72552ad069537d6eea81e719c7.json> <link rel=stylesheet class=code_theme> <style>.error-screen{font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;padding-top:50px;max-width:750px;font-size:14px;color:#333;margin:auto;display:none;line-height:1.5}.error-screen h2{font-size:30px;font-weight:400;line-height:1.2}.error-screen ul{padding-left:15px;line-height:1.7;margin-top:0;margin-bottom:10px}.error-screen hr{color:#ddd;margin-top:20px;margin-bottom:20px;border:0;border-top:1px solid #eee}.error-screen-visible{display:block}</style><script>window.publicPath='/nested/subpath/static/'</script> <link href="/nested/subpath/static/main.364fd054d7a6d741efc6.css" rel="stylesheet"><script type="text/javascript" src="/nested/subpath/static/main.e49599ac425584ffead5.js"></script></head> <body class=font--open_sans> <div id=root> <div class=error-screen> <h2>Cannot connect to Mattermost</h2> <hr/> <p>We're having trouble connecting to Mattermost. If refreshing this page (Ctrl+R or Command+R) does not work, please verify that your computer is connected to the internet.</p> <br/> </div> <div class=loading-screen style=position:relative> <div class=loading__content> <div class="round round-1"></div> <div class="round round-2"></div> <div class="round round-3"></div> </div> </div> </div> <noscript> To use Mattermost, please enable JavaScript. </noscript> </body> </html>`

const newSubpathCss = `@font-face{font-family:FontAwesome;src:url(/nested/subpath/static/files/674f50d287a8c48dc19ba404d20fe713.eot);src:url(/nested/subpath/static/files/674f50d287a8c48dc19ba404d20fe713.eot?#iefix&v=4.7.0) format("embedded-opentype"),url(/nested/subpath/static/files/af7ae505a9eed503f8b8e6982036873e.woff2) format("woff2"),url(/nested/subpath/static/files/fee66e712a8a08eef5805a46892932ad.woff) format("woff"),url(/nested/subpath/static/files/b06871f281fee6b241d60582ae9369b9.ttf) format("truetype"),url(/nested/subpath/static/files/677433a0892aaed7b7d2628c313c9775.svg#fontawesomeregular) format("svg");font-weight:400;font-style:normal}`

const resetRootHtml = `<!DOCTYPE html> <html lang=en> <head> <meta charset=utf-8> <meta http-equiv="Content-Security-Policy" content="script-src 'self' cdn.segment.com/analytics.js/ 'sha256-VFw7U/t/OI+I9YMja3c2GDwEQbnlOq/L5+GealgesK8='"> <meta http-equiv=X-UA-Compatible content="IE=edge"> <meta name=viewport content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=0"> <meta name=robots content="noindex, nofollow"> <meta name=referrer content=no-referrer> <title>Mattermost</title> <meta name=apple-mobile-web-app-capable content=yes> <meta name=apple-mobile-web-app-status-bar-style content=default> <meta name=mobile-web-app-capable content=yes> <meta name=apple-mobile-web-app-title content=Mattermost> <meta name=application-name content=Mattermost> <meta name=format-detection content="telephone=no"> <link rel=apple-touch-icon sizes=57x57 href=/static/files/78b7e73b41b8731ce2c41c870ecc8886.png> <link rel=apple-touch-icon sizes=60x60 href=/static/files/51d00ffd13afb6d74fd8f6dfdeef768a.png> <link rel=apple-touch-icon sizes=72x72 href=/static/files/23645596f8f78f017bd4d457abb855c4.png> <link rel=apple-touch-icon sizes=76x76 href=/static/files/26e9d72f472663a00b4b206149459fab.png> <link rel=apple-touch-icon sizes=144x144 href=/static/files/7bd91659bf3fc8c68fcd45fc1db9c630.png> <link rel=apple-touch-icon sizes=120x120 href=/static/files/fa69ffe11eb334aaef5aece8d848ca62.png> <link rel=apple-touch-icon sizes=152x152 href=/static/files/f046777feb6ab12fc43b8f9908b1db35.png> <link rel=icon type=image/png sizes=16x16 href=/static/files/02b96247d275680adaaabf01c71c571d.png> <link rel=icon type=image/png sizes=32x32 href=/static/files/1d9020f201a6762421cab8d30624fdd8.png> <link rel=icon type=image/png sizes=96x96 href=/static/files/fe23af39ae98d77dc26ae8586565970f.png> <link rel=icon type=image/png sizes=192x192 href=/static/files/d7ff68a7675f84337cc154c3d4abe713.png> <link rel=manifest href=/static/files/a985ad72552ad069537d6eea81e719c7.json> <link rel=stylesheet class=code_theme> <style>.error-screen{font-family:'Helvetica Neue',Helvetica,Arial,sans-serif;padding-top:50px;max-width:750px;font-size:14px;color:#333;margin:auto;display:none;line-height:1.5}.error-screen h2{font-size:30px;font-weight:400;line-height:1.2}.error-screen ul{padding-left:15px;line-height:1.7;margin-top:0;margin-bottom:10px}.error-screen hr{color:#ddd;margin-top:20px;margin-bottom:20px;border:0;border-top:1px solid #eee}.error-screen-visible{display:block}</style><script>window.publicPath='/static/'</script> <link href="/static/main.364fd054d7a6d741efc6.css" rel="stylesheet"><script type="text/javascript" src="/static/main.e49599ac425584ffead5.js"></script></head> <body class=font--open_sans> <div id=root> <div class=error-screen> <h2>Cannot connect to Mattermost</h2> <hr/> <p>We're having trouble connecting to Mattermost. If refreshing this page (Ctrl+R or Command+R) does not work, please verify that your computer is connected to the internet.</p> <br/> </div> <div class=loading-screen style=position:relative> <div class=loading__content> <div class="round round-1"></div> <div class="round round-2"></div> <div class="round round-3"></div> </div> </div> </div> <noscript> To use Mattermost, please enable JavaScript. </noscript> </body> </html>`

const baseManifestJson = `{
  "icons": [
    {
      "src": "/static/icon_96x96.png",
      "sizes": "96x96",
      "type": "image/png"
    },
    {
      "src": "/static/icon_32x32.png",
      "sizes": "32x32",
      "type": "image/png"
    },
    {
      "src": "/static/icon_16x16.png",
      "sizes": "16x16",
      "type": "image/png"
    },
    {
      "src": "/static/icon_76x76.png",
      "sizes": "76x76",
      "type": "image/png"
    },
    {
      "src": "/static/icon_72x72.png",
      "sizes": "72x72",
      "type": "image/png"
    },
    {
      "src": "/static/icon_60x60.png",
      "sizes": "60x60",
      "type": "image/png"
    },
    {
      "src": "/static/icon_57x57.png",
      "sizes": "57x57",
      "type": "image/png"
    },
    {
      "src": "/static/icon_152x152.png",
      "sizes": "152x152",
      "type": "image/png"
    },
    {
      "src": "/static/icon_144x144.png",
      "sizes": "144x144",
      "type": "image/png"
    },
    {
      "src": "/static/icon_120x120.png",
      "sizes": "120x120",
      "type": "image/png"
    },
    {
      "src": "/static/icon_192x192.png",
      "sizes": "192x192",
      "type": "image/png"
    }
  ],
  "name": "Mattermost",
  "short_name": "Mattermost",
  "orientation": "any",
  "display": "standalone",
  "start_url": ".",
  "description": "Mattermost is an open source, self-hosted Slack-alternative",
  "background_color": "#ffffff"
}
`

const subpathManifestJson = `{
  "icons": [
    {
      "src": "/subpath/static/icon_96x96.png",
      "sizes": "96x96",
      "type": "image/png"
    },
    {
      "src": "/subpath/static/icon_32x32.png",
      "sizes": "32x32",
      "type": "image/png"
    },
    {
      "src": "/subpath/static/icon_16x16.png",
      "sizes": "16x16",
      "type": "image/png"
    },
    {
      "src": "/subpath/static/icon_76x76.png",
      "sizes": "76x76",
      "type": "image/png"
    },
    {
      "src": "/subpath/static/icon_72x72.png",
      "sizes": "72x72",
      "type": "image/png"
    },
    {
      "src": "/subpath/static/icon_60x60.png",
      "sizes": "60x60",
      "type": "image/png"
    },
    {
      "src": "/subpath/static/icon_57x57.png",
      "sizes": "57x57",
      "type": "image/png"
    },
    {
      "src": "/subpath/static/icon_152x152.png",
      "sizes": "152x152",
      "type": "image/png"
    },
    {
      "src": "/subpath/static/icon_144x144.png",
      "sizes": "144x144",
      "type": "image/png"
    },
    {
      "src": "/subpath/static/icon_120x120.png",
      "sizes": "120x120",
      "type": "image/png"
    },
    {
      "src": "/subpath/static/icon_192x192.png",
      "sizes": "192x192",
      "type": "image/png"
    }
  ],
  "name": "Mattermost",
  "short_name": "Mattermost",
  "orientation": "any",
  "display": "standalone",
  "start_url": ".",
  "description": "Mattermost is an open source, self-hosted Slack-alternative",
  "background_color": "#ffffff"
}
`

const newSubpathManifestJson = `{
  "icons": [
    {
      "src": "/nested/subpath/static/icon_96x96.png",
      "sizes": "96x96",
      "type": "image/png"
    },
    {
      "src": "/nested/subpath/static/icon_32x32.png",
      "sizes": "32x32",
      "type": "image/png"
    },
    {
      "src": "/nested/subpath/static/icon_16x16.png",
      "sizes": "16x16",
      "type": "image/png"
    },
    {
      "src": "/nested/subpath/static/icon_76x76.png",
      "sizes": "76x76",
      "type": "image/png"
    },
    {
      "src": "/nested/subpath/static/icon_72x72.png",
      "sizes": "72x72",
      "type": "image/png"
    },
    {
      "src": "/nested/subpath/static/icon_60x60.png",
      "sizes": "60x60",
      "type": "image/png"
    },
    {
      "src": "/nested/subpath/static/icon_57x57.png",
      "sizes": "57x57",
      "type": "image/png"
    },
    {
      "src": "/nested/subpath/static/icon_152x152.png",
      "sizes": "152x152",
      "type": "image/png"
    },
    {
      "src": "/nested/subpath/static/icon_144x144.png",
      "sizes": "144x144",
      "type": "image/png"
    },
    {
      "src": "/nested/subpath/static/icon_120x120.png",
      "sizes": "120x120",
      "type": "image/png"
    },
    {
      "src": "/nested/subpath/static/icon_192x192.png",
      "sizes": "192x192",
      "type": "image/png"
    }
  ],
  "name": "Mattermost",
  "short_name": "Mattermost",
  "orientation": "any",
  "display": "standalone",
  "start_url": ".",
  "description": "Mattermost is an open source, self-hosted Slack-alternative",
  "background_color": "#ffffff"
}
`
