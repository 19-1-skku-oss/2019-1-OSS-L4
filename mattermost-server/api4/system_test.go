package api4

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
	"github.com/stretchr/testify/assert"
)

func TestGetPing(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	goRoutineHealthThreshold := *th.App.Config().ServiceSettings.GoroutineHealthThreshold
	defer func() {
		th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.GoroutineHealthThreshold = goRoutineHealthThreshold })
	}()

	status, resp := Client.GetPing()
	CheckNoError(t, resp)
	if status != "OK" {
		t.Fatal("should return OK")
	}

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.GoroutineHealthThreshold = 10 })
	status, resp = th.SystemAdminClient.GetPing()
	CheckInternalErrorStatus(t, resp)
	if status != "unhealthy" {
		t.Fatal("should return unhealthy")
	}
}

func TestGetAudits(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	audits, resp := th.SystemAdminClient.GetAudits(0, 100, "")
	CheckNoError(t, resp)

	if len(audits) == 0 {
		t.Fatal("should not be empty")
	}

	audits, resp = th.SystemAdminClient.GetAudits(0, 1, "")
	CheckNoError(t, resp)

	if len(audits) != 1 {
		t.Fatal("should only be 1")
	}

	audits, resp = th.SystemAdminClient.GetAudits(1, 1, "")
	CheckNoError(t, resp)

	if len(audits) != 1 {
		t.Fatal("should only be 1")
	}

	_, resp = th.SystemAdminClient.GetAudits(-1, -1, "")
	CheckNoError(t, resp)

	_, resp = Client.GetAudits(0, 100, "")
	CheckForbiddenStatus(t, resp)

	Client.Logout()
	_, resp = Client.GetAudits(0, 100, "")
	CheckUnauthorizedStatus(t, resp)
}

func TestEmailTest(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	config := model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: model.NewString(""),
		},
		EmailSettings: model.EmailSettings{
			SMTPServer:             model.NewString(""),
			SMTPPort:               model.NewString(""),
			SMTPPassword:           model.NewString(""),
			FeedbackName:           model.NewString(""),
			FeedbackEmail:          model.NewString(""),
			ReplyToAddress:         model.NewString(""),
			SendEmailNotifications: model.NewBool(false),
		},
	}

	t.Run("as system user", func(t *testing.T) {
		_, resp := Client.TestEmail(&config)
		CheckForbiddenStatus(t, resp)
	})

	t.Run("as system admin", func(t *testing.T) {
		_, resp := th.SystemAdminClient.TestEmail(&config)
		CheckErrorMessage(t, resp, "api.admin.test_email.missing_server")
		CheckBadRequestStatus(t, resp)

		inbucket_host := os.Getenv("CI_INBUCKET_HOST")
		if inbucket_host == "" {
			inbucket_host = "dockerhost"
		}

		inbucket_port := os.Getenv("CI_INBUCKET_PORT")
		if inbucket_port == "" {
			inbucket_port = "9000"
		}

		*config.EmailSettings.SMTPServer = inbucket_host
		*config.EmailSettings.SMTPPort = inbucket_port
		_, resp = th.SystemAdminClient.TestEmail(&config)
		CheckOKStatus(t, resp)
	})

	t.Run("as restricted system admin", func(t *testing.T) {
		th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ExperimentalSettings.RestrictSystemAdmin = true })

		_, resp := th.SystemAdminClient.TestEmail(&config)
		CheckForbiddenStatus(t, resp)
	})
}

func TestDatabaseRecycle(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	t.Run("as system user", func(t *testing.T) {
		_, resp := Client.DatabaseRecycle()
		CheckForbiddenStatus(t, resp)
	})

	t.Run("as system admin", func(t *testing.T) {
		_, resp := th.SystemAdminClient.DatabaseRecycle()
		CheckNoError(t, resp)
	})

	t.Run("as restricted system admin", func(t *testing.T) {
		th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ExperimentalSettings.RestrictSystemAdmin = true })

		_, resp := th.SystemAdminClient.DatabaseRecycle()
		CheckForbiddenStatus(t, resp)
	})
}

func TestInvalidateCaches(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	t.Run("as system user", func(t *testing.T) {
		ok, resp := Client.InvalidateCaches()
		CheckForbiddenStatus(t, resp)
		if ok {
			t.Fatal("should not clean the cache due no permission.")
		}
	})

	t.Run("as system admin", func(t *testing.T) {
		ok, resp := th.SystemAdminClient.InvalidateCaches()
		CheckNoError(t, resp)
		if !ok {
			t.Fatal("should clean the cache")
		}
	})

	t.Run("as restricted system admin", func(t *testing.T) {
		th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ExperimentalSettings.RestrictSystemAdmin = true })

		ok, resp := th.SystemAdminClient.InvalidateCaches()
		CheckForbiddenStatus(t, resp)
		if ok {
			t.Fatal("should not clean the cache due no permission.")
		}
	})
}

func TestGetLogs(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	for i := 0; i < 20; i++ {
		mlog.Info(fmt.Sprint(i))
	}

	logs, resp := th.SystemAdminClient.GetLogs(0, 10)
	CheckNoError(t, resp)

	if len(logs) != 10 {
		t.Log(len(logs))
		t.Fatal("wrong length")
	}
	for i := 10; i < 20; i++ {
		assert.Containsf(t, logs[i-10], fmt.Sprintf(`"msg":"%d"`, i), "Log line doesn't contain correct message")
	}

	logs, resp = th.SystemAdminClient.GetLogs(1, 10)
	CheckNoError(t, resp)

	if len(logs) != 10 {
		t.Log(len(logs))
		t.Fatal("wrong length")
	}

	logs, resp = th.SystemAdminClient.GetLogs(-1, -1)
	CheckNoError(t, resp)

	if len(logs) == 0 {
		t.Fatal("should not be empty")
	}

	_, resp = Client.GetLogs(0, 10)
	CheckForbiddenStatus(t, resp)

	Client.Logout()
	_, resp = Client.GetLogs(0, 10)
	CheckUnauthorizedStatus(t, resp)
}

func TestPostLog(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	enableDev := *th.App.Config().ServiceSettings.EnableDeveloper
	defer func() {
		*th.App.Config().ServiceSettings.EnableDeveloper = enableDev
	}()
	*th.App.Config().ServiceSettings.EnableDeveloper = true

	message := make(map[string]string)
	message["level"] = "ERROR"
	message["message"] = "this is a test"

	_, resp := Client.PostLog(message)
	CheckNoError(t, resp)

	Client.Logout()

	_, resp = Client.PostLog(message)
	CheckNoError(t, resp)

	*th.App.Config().ServiceSettings.EnableDeveloper = false

	_, resp = Client.PostLog(message)
	CheckForbiddenStatus(t, resp)

	logMessage, resp := th.SystemAdminClient.PostLog(message)
	CheckNoError(t, resp)
	if len(logMessage) == 0 {
		t.Fatal("should return the log message")
	}
}

func TestGetAnalyticsOld(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	rows, resp := Client.GetAnalyticsOld("", "")
	CheckForbiddenStatus(t, resp)
	if rows != nil {
		t.Fatal("should be nil")
	}

	rows, resp = th.SystemAdminClient.GetAnalyticsOld("", "")
	CheckNoError(t, resp)

	found := false
	found2 := false
	for _, row := range rows {
		if row.Name == "unique_user_count" {
			found = true
		} else if row.Name == "inactive_user_count" {
			found2 = true
			assert.True(t, row.Value >= 0)
		}
	}

	assert.True(t, found, "should return unique user count")
	assert.True(t, found2, "should return inactive user count")

	_, resp = th.SystemAdminClient.GetAnalyticsOld("post_counts_day", "")
	CheckNoError(t, resp)

	_, resp = th.SystemAdminClient.GetAnalyticsOld("user_counts_with_posts_day", "")
	CheckNoError(t, resp)

	_, resp = th.SystemAdminClient.GetAnalyticsOld("extra_counts", "")
	CheckNoError(t, resp)

	rows, resp = th.SystemAdminClient.GetAnalyticsOld("", th.BasicTeam.Id)
	CheckNoError(t, resp)

	for _, row := range rows {
		if row.Name == "inactive_user_count" {
			assert.Equal(t, float64(-1), row.Value, "inactive user count should be -1 when team specified")
		}
	}

	rows2, resp2 := th.SystemAdminClient.GetAnalyticsOld("standard", "")
	CheckNoError(t, resp2)
	assert.Equal(t, "total_websocket_connections", rows2[5].Name)
	assert.Equal(t, float64(0), rows2[5].Value)

	WebSocketClient, err := th.CreateWebSocketClient()
	if err != nil {
		t.Fatal(err)
	}

	rows2, resp2 = th.SystemAdminClient.GetAnalyticsOld("standard", "")
	CheckNoError(t, resp2)
	assert.Equal(t, "total_websocket_connections", rows2[5].Name)
	assert.Equal(t, float64(1), rows2[5].Value)

	WebSocketClient.Close()

	rows2, resp2 = th.SystemAdminClient.GetAnalyticsOld("standard", "")
	CheckNoError(t, resp2)
	assert.Equal(t, "total_websocket_connections", rows2[5].Name)
	assert.Equal(t, float64(0), rows2[5].Value)

	Client.Logout()
	_, resp = Client.GetAnalyticsOld("", th.BasicTeam.Id)
	CheckUnauthorizedStatus(t, resp)
}

func TestS3TestConnection(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	s3Host := os.Getenv("CI_MINIO_HOST")
	if s3Host == "" {
		s3Host = "dockerhost"
	}

	s3Port := os.Getenv("CI_MINIO_PORT")
	if s3Port == "" {
		s3Port = "9001"
	}

	s3Endpoint := fmt.Sprintf("%s:%s", s3Host, s3Port)
	config := model.Config{
		FileSettings: model.FileSettings{
			DriverName:              model.NewString(model.IMAGE_DRIVER_S3),
			AmazonS3AccessKeyId:     model.NewString(model.MINIO_ACCESS_KEY),
			AmazonS3SecretAccessKey: model.NewString(model.MINIO_SECRET_KEY),
			AmazonS3Bucket:          model.NewString(""),
			AmazonS3Endpoint:        model.NewString(s3Endpoint),
			AmazonS3Region:          model.NewString(""),
			AmazonS3SSL:             model.NewBool(false),
		},
	}

	t.Run("as system user", func(t *testing.T) {
		_, resp := Client.TestS3Connection(&config)
		CheckForbiddenStatus(t, resp)
	})

	t.Run("as system admin", func(t *testing.T) {
		_, resp := th.SystemAdminClient.TestS3Connection(&config)
		CheckBadRequestStatus(t, resp)
		if resp.Error.Message != "S3 Bucket is required" {
			t.Fatal("should return error - missing s3 bucket")
		}

		// If this fails, check the test configuration to ensure minio is setup with the
		// `mattermost-test` bucket defined by model.MINIO_BUCKET.
		*config.FileSettings.AmazonS3Bucket = model.MINIO_BUCKET
		*config.FileSettings.AmazonS3Region = "us-east-1"
		_, resp = th.SystemAdminClient.TestS3Connection(&config)
		CheckOKStatus(t, resp)

		config.FileSettings.AmazonS3Region = model.NewString("")
		_, resp = th.SystemAdminClient.TestS3Connection(&config)
		CheckOKStatus(t, resp)

		config.FileSettings.AmazonS3Bucket = model.NewString("Wrong_bucket")
		_, resp = th.SystemAdminClient.TestS3Connection(&config)
		CheckInternalErrorStatus(t, resp)
		assert.Equal(t, "Unable to create bucket.", resp.Error.Message)

		*config.FileSettings.AmazonS3Bucket = "shouldcreatenewbucket"
		_, resp = th.SystemAdminClient.TestS3Connection(&config)
		CheckOKStatus(t, resp)
	})

	t.Run("as restricted system admin", func(t *testing.T) {
		th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ExperimentalSettings.RestrictSystemAdmin = true })

		_, resp := th.SystemAdminClient.TestS3Connection(&config)
		CheckForbiddenStatus(t, resp)
	})

}

func TestSupportedTimezones(t *testing.T) {
	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client

	supportedTimezonesFromConfig := th.App.Timezones.GetSupported()
	supportedTimezones, resp := Client.GetSupportedTimezone()

	CheckNoError(t, resp)
	assert.Equal(t, supportedTimezonesFromConfig, supportedTimezones)
}

func TestRedirectLocation(t *testing.T) {
	expected := "https://mattermost.com/wp-content/themes/mattermostv2/img/logo-light.svg"

	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Location", expected)
		res.WriteHeader(http.StatusFound)
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	mockBitlyLink := testServer.URL

	th := Setup().InitBasic()
	defer th.TearDown()
	Client := th.Client
	enableLinkPreviews := *th.App.Config().ServiceSettings.EnableLinkPreviews
	defer func() {
		th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.EnableLinkPreviews = enableLinkPreviews })
	}()

	*th.App.Config().ServiceSettings.EnableLinkPreviews = true
	*th.App.Config().ServiceSettings.AllowedUntrustedInternalConnections = "127.0.0.1"

	_, resp := th.SystemAdminClient.GetRedirectLocation("https://mattermost.com/", "")
	CheckNoError(t, resp)

	_, resp = th.SystemAdminClient.GetRedirectLocation("", "")
	CheckBadRequestStatus(t, resp)

	actual, resp := th.SystemAdminClient.GetRedirectLocation(mockBitlyLink, "")
	CheckNoError(t, resp)
	assert.Equal(t, expected, actual)

	// Check cached value
	actual, resp = th.SystemAdminClient.GetRedirectLocation(mockBitlyLink, "")
	CheckNoError(t, resp)
	assert.Equal(t, expected, actual)

	*th.App.Config().ServiceSettings.EnableLinkPreviews = false
	actual, resp = th.SystemAdminClient.GetRedirectLocation("https://mattermost.com/", "")
	CheckNoError(t, resp)
	assert.Equal(t, actual, "")

	actual, resp = th.SystemAdminClient.GetRedirectLocation("", "")
	CheckNoError(t, resp)
	assert.Equal(t, actual, "")

	actual, resp = th.SystemAdminClient.GetRedirectLocation(mockBitlyLink, "")
	CheckNoError(t, resp)
	assert.Equal(t, actual, "")

	Client.Logout()
	_, resp = Client.GetRedirectLocation("", "")
	CheckUnauthorizedStatus(t, resp)
}
