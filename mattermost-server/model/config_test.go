// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDefaults(t *testing.T) {
	t.Parallel()

	t.Run("somewhere nil when uninitialized", func(t *testing.T) {
		c := Config{}
		require.False(t, checkNowhereNil(t, "config", c))
	})

	t.Run("nowhere nil when initialized", func(t *testing.T) {
		c := Config{}
		c.SetDefaults()
		require.True(t, checkNowhereNil(t, "config", c))
	})

	t.Run("nowhere nil when partially initialized", func(t *testing.T) {
		var recursivelyUninitialize func(*Config, string, reflect.Value)
		recursivelyUninitialize = func(config *Config, name string, v reflect.Value) {
			if v.Type().Kind() == reflect.Ptr {
				// Set every pointer we find in the tree to nil
				v.Set(reflect.Zero(v.Type()))
				require.True(t, v.IsNil())

				// SetDefaults on the root config should make it non-nil, otherwise
				// it means that SetDefaults isn't being called recursively in
				// all cases.
				config.SetDefaults()
				if assert.False(t, v.IsNil(), "%s should be non-nil after SetDefaults()", name) {
					recursivelyUninitialize(config, fmt.Sprintf("(*%s)", name), v.Elem())
				}

			} else if v.Type().Kind() == reflect.Struct {
				for i := 0; i < v.NumField(); i++ {
					recursivelyUninitialize(config, fmt.Sprintf("%s.%s", name, v.Type().Field(i).Name), v.Field(i))
				}
			}
		}

		c := Config{}
		c.SetDefaults()
		recursivelyUninitialize(&c, "config", reflect.ValueOf(&c).Elem())
	})
}

func TestConfigEmptySiteName(t *testing.T) {
	c1 := Config{
		TeamSettings: TeamSettings{
			SiteName: NewString(""),
		},
	}
	c1.SetDefaults()

	if *c1.TeamSettings.SiteName != TEAM_SETTINGS_DEFAULT_SITE_NAME {
		t.Fatal("TeamSettings.SiteName should default to " + TEAM_SETTINGS_DEFAULT_SITE_NAME)
	}
}

func TestConfigDefaultFileSettingsDirectory(t *testing.T) {
	c1 := Config{}
	c1.SetDefaults()

	if *c1.FileSettings.Directory != "./data/" {
		t.Fatal("FileSettings.Directory should default to './data/'")
	}
}

func TestConfigDefaultEmailNotificationContentsType(t *testing.T) {
	c1 := Config{}
	c1.SetDefaults()

	if *c1.EmailSettings.EmailNotificationContentsType != EMAIL_NOTIFICATION_CONTENTS_FULL {
		t.Fatal("EmailSettings.EmailNotificationContentsType should default to 'full'")
	}
}

func TestConfigDefaultFileSettingsS3SSE(t *testing.T) {
	c1 := Config{}
	c1.SetDefaults()

	if *c1.FileSettings.AmazonS3SSE {
		t.Fatal("FileSettings.AmazonS3SSE should default to false")
	}
}

func TestConfigDefaultServiceSettingsExperimentalGroupUnreadChannels(t *testing.T) {
	c1 := Config{}
	c1.SetDefaults()

	if *c1.ServiceSettings.ExperimentalGroupUnreadChannels != GROUP_UNREAD_CHANNELS_DISABLED {
		t.Fatal("ServiceSettings.ExperimentalGroupUnreadChannels should default to 'disabled'")
	}

	// This setting was briefly a boolean, so ensure that those values still work as expected
	c1 = Config{
		ServiceSettings: ServiceSettings{
			ExperimentalGroupUnreadChannels: NewString("1"),
		},
	}
	c1.SetDefaults()

	if *c1.ServiceSettings.ExperimentalGroupUnreadChannels != GROUP_UNREAD_CHANNELS_DEFAULT_ON {
		t.Fatal("ServiceSettings.ExperimentalGroupUnreadChannels should set true to 'default on'")
	}

	c1 = Config{
		ServiceSettings: ServiceSettings{
			ExperimentalGroupUnreadChannels: NewString("0"),
		},
	}
	c1.SetDefaults()

	if *c1.ServiceSettings.ExperimentalGroupUnreadChannels != GROUP_UNREAD_CHANNELS_DISABLED {
		t.Fatal("ServiceSettings.ExperimentalGroupUnreadChannels should set false to 'disabled'")
	}
}

func TestConfigDefaultNPSPluginState(t *testing.T) {
	c1 := Config{}
	c1.SetDefaults()

	if c1.PluginSettings.PluginStates["com.mattermost.nps"].Enable != true {
		t.Fatal("PluginSettings.PluginStates[\"com.mattermost.nps\"].Enable should default to true")
	}

	c1.PluginSettings.PluginStates["com.mattermost.nps"].Enable = false
	c1.SetDefaults()
	if c1.PluginSettings.PluginStates["com.mattermost.nps"].Enable != false {
		t.Fatal("PluginSettings.PluginStates[\"com.mattermost.nps\"].Enable should remain false")
	}
}

func TestTeamSettingsIsValidSiteNameEmpty(t *testing.T) {
	c1 := Config{}
	c1.SetDefaults()
	c1.TeamSettings.SiteName = NewString("")

	// should fail fast because ts.SiteName is not set
	err := c1.TeamSettings.isValid()
	if err == nil {
		t.Fatal("TeamSettings validation should fail with an empty SiteName")
	}
}

func TestMessageExportSettingsIsValidEnableExportNotSet(t *testing.T) {
	fs := &FileSettings{}
	mes := &MessageExportSettings{}

	// should fail fast because mes.EnableExport is not set
	require.Error(t, mes.isValid(*fs))
}

func TestMessageExportSettingsIsValidEnableExportFalse(t *testing.T) {
	fs := &FileSettings{}
	mes := &MessageExportSettings{
		EnableExport: NewBool(false),
	}

	// should fail fast because message export isn't enabled
	require.Nil(t, mes.isValid(*fs))
}

func TestMessageExportSettingsIsValidExportFromTimestampInvalid(t *testing.T) {
	fs := &FileSettings{}
	mes := &MessageExportSettings{
		EnableExport: NewBool(true),
	}

	// should fail fast because export from timestamp isn't set
	require.Error(t, mes.isValid(*fs))

	mes.ExportFromTimestamp = NewInt64(-1)

	// should fail fast because export from timestamp isn't valid
	require.Error(t, mes.isValid(*fs))

	mes.ExportFromTimestamp = NewInt64(GetMillis() + 10000)

	// should fail fast because export from timestamp is greater than current time
	require.Error(t, mes.isValid(*fs))
}

func TestMessageExportSettingsIsValidDailyRunTimeInvalid(t *testing.T) {
	fs := &FileSettings{}
	mes := &MessageExportSettings{
		EnableExport:        NewBool(true),
		ExportFromTimestamp: NewInt64(0),
	}

	// should fail fast because daily runtime isn't set
	require.Error(t, mes.isValid(*fs))

	mes.DailyRunTime = NewString("33:33:33")

	// should fail fast because daily runtime is invalid format
	require.Error(t, mes.isValid(*fs))
}

func TestMessageExportSettingsIsValidBatchSizeInvalid(t *testing.T) {
	fs := &FileSettings{
		DriverName: NewString("foo"), // bypass file location check
	}
	mes := &MessageExportSettings{
		EnableExport:        NewBool(true),
		ExportFromTimestamp: NewInt64(0),
		DailyRunTime:        NewString("15:04"),
	}

	// should fail fast because batch size isn't set
	require.Error(t, mes.isValid(*fs))
}

func TestMessageExportSettingsIsValidExportFormatInvalid(t *testing.T) {
	fs := &FileSettings{
		DriverName: NewString("foo"), // bypass file location check
	}
	mes := &MessageExportSettings{
		EnableExport:        NewBool(true),
		ExportFromTimestamp: NewInt64(0),
		DailyRunTime:        NewString("15:04"),
		BatchSize:           NewInt(100),
	}

	// should fail fast because export format isn't set
	require.Error(t, mes.isValid(*fs))
}

func TestMessageExportSettingsIsValidGlobalRelayEmailAddressInvalid(t *testing.T) {
	fs := &FileSettings{
		DriverName: NewString("foo"), // bypass file location check
	}
	mes := &MessageExportSettings{
		EnableExport:        NewBool(true),
		ExportFormat:        NewString(COMPLIANCE_EXPORT_TYPE_GLOBALRELAY),
		ExportFromTimestamp: NewInt64(0),
		DailyRunTime:        NewString("15:04"),
		BatchSize:           NewInt(100),
	}

	// should fail fast because global relay email address isn't set
	require.Error(t, mes.isValid(*fs))
}

func TestMessageExportSettingsIsValidActiance(t *testing.T) {
	fs := &FileSettings{
		DriverName: NewString("foo"), // bypass file location check
	}
	mes := &MessageExportSettings{
		EnableExport:        NewBool(true),
		ExportFormat:        NewString(COMPLIANCE_EXPORT_TYPE_ACTIANCE),
		ExportFromTimestamp: NewInt64(0),
		DailyRunTime:        NewString("15:04"),
		BatchSize:           NewInt(100),
	}

	// should pass because everything is valid
	require.Nil(t, mes.isValid(*fs))
}

func TestMessageExportSettingsIsValidGlobalRelaySettingsMissing(t *testing.T) {
	fs := &FileSettings{
		DriverName: NewString("foo"), // bypass file location check
	}
	mes := &MessageExportSettings{
		EnableExport:        NewBool(true),
		ExportFormat:        NewString(COMPLIANCE_EXPORT_TYPE_GLOBALRELAY),
		ExportFromTimestamp: NewInt64(0),
		DailyRunTime:        NewString("15:04"),
		BatchSize:           NewInt(100),
	}

	// should fail because globalrelay settings are missing
	require.Error(t, mes.isValid(*fs))
}

func TestMessageExportSettingsIsValidGlobalRelaySettingsInvalidCustomerType(t *testing.T) {
	fs := &FileSettings{
		DriverName: NewString("foo"), // bypass file location check
	}
	mes := &MessageExportSettings{
		EnableExport:        NewBool(true),
		ExportFormat:        NewString(COMPLIANCE_EXPORT_TYPE_GLOBALRELAY),
		ExportFromTimestamp: NewInt64(0),
		DailyRunTime:        NewString("15:04"),
		BatchSize:           NewInt(100),
		GlobalRelaySettings: &GlobalRelayMessageExportSettings{
			CustomerType: NewString("Invalid"),
			EmailAddress: NewString("valid@mattermost.com"),
			SmtpUsername: NewString("SomeUsername"),
			SmtpPassword: NewString("SomePassword"),
		},
	}

	// should fail because customer type is invalid
	require.Error(t, mes.isValid(*fs))
}

// func TestMessageExportSettingsIsValidGlobalRelaySettingsInvalidEmailAddress(t *testing.T) {
func TestMessageExportSettingsGlobalRelaySettings(t *testing.T) {
	fs := &FileSettings{
		DriverName: NewString("foo"), // bypass file location check
	}
	tests := []struct {
		name    string
		value   *GlobalRelayMessageExportSettings
		success bool
	}{
		{
			"Invalid email address",
			&GlobalRelayMessageExportSettings{
				CustomerType: NewString(GLOBALRELAY_CUSTOMER_TYPE_A9),
				EmailAddress: NewString("invalidEmailAddress"),
				SmtpUsername: NewString("SomeUsername"),
				SmtpPassword: NewString("SomePassword"),
			},
			false,
		},
		{
			"Missing smtp username",
			&GlobalRelayMessageExportSettings{
				CustomerType: NewString(GLOBALRELAY_CUSTOMER_TYPE_A10),
				EmailAddress: NewString("valid@mattermost.com"),
				SmtpPassword: NewString("SomePassword"),
			},
			false,
		},
		{
			"Invalid smtp username",
			&GlobalRelayMessageExportSettings{
				CustomerType: NewString(GLOBALRELAY_CUSTOMER_TYPE_A10),
				EmailAddress: NewString("valid@mattermost.com"),
				SmtpUsername: NewString(""),
				SmtpPassword: NewString("SomePassword"),
			},
			false,
		},
		{
			"Invalid smtp password",
			&GlobalRelayMessageExportSettings{
				CustomerType: NewString(GLOBALRELAY_CUSTOMER_TYPE_A10),
				EmailAddress: NewString("valid@mattermost.com"),
				SmtpUsername: NewString("SomeUsername"),
				SmtpPassword: NewString(""),
			},
			false,
		},
		{
			"Valid data",
			&GlobalRelayMessageExportSettings{
				CustomerType: NewString(GLOBALRELAY_CUSTOMER_TYPE_A9),
				EmailAddress: NewString("valid@mattermost.com"),
				SmtpUsername: NewString("SomeUsername"),
				SmtpPassword: NewString("SomePassword"),
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mes := &MessageExportSettings{
				EnableExport:        NewBool(true),
				ExportFormat:        NewString(COMPLIANCE_EXPORT_TYPE_GLOBALRELAY),
				ExportFromTimestamp: NewInt64(0),
				DailyRunTime:        NewString("15:04"),
				BatchSize:           NewInt(100),
				GlobalRelaySettings: tt.value,
			}

			if tt.success {
				require.Nil(t, mes.isValid(*fs))
			} else {
				require.Error(t, mes.isValid(*fs))
			}
		})
	}
}

func TestMessageExportSetDefaults(t *testing.T) {
	mes := &MessageExportSettings{}
	mes.SetDefaults()

	require.False(t, *mes.EnableExport)
	require.Equal(t, "01:00", *mes.DailyRunTime)
	require.Equal(t, int64(0), *mes.ExportFromTimestamp)
	require.Equal(t, 10000, *mes.BatchSize)
	require.Equal(t, COMPLIANCE_EXPORT_TYPE_ACTIANCE, *mes.ExportFormat)
}

func TestMessageExportSetDefaultsExportEnabledExportFromTimestampNil(t *testing.T) {
	// Test retained as protection against regression of MM-13185
	mes := &MessageExportSettings{
		EnableExport: NewBool(true),
	}
	mes.SetDefaults()

	require.True(t, *mes.EnableExport)
	require.Equal(t, "01:00", *mes.DailyRunTime)
	require.Equal(t, int64(0), *mes.ExportFromTimestamp)
	require.True(t, *mes.ExportFromTimestamp <= GetMillis())
	require.Equal(t, 10000, *mes.BatchSize)
}

func TestMessageExportSetDefaultsExportEnabledExportFromTimestampZero(t *testing.T) {
	// Test retained as protection against regression of MM-13185
	mes := &MessageExportSettings{
		EnableExport:        NewBool(true),
		ExportFromTimestamp: NewInt64(0),
	}
	mes.SetDefaults()

	require.True(t, *mes.EnableExport)
	require.Equal(t, "01:00", *mes.DailyRunTime)
	require.Equal(t, int64(0), *mes.ExportFromTimestamp)
	require.True(t, *mes.ExportFromTimestamp <= GetMillis())
	require.Equal(t, 10000, *mes.BatchSize)
}

func TestMessageExportSetDefaultsExportEnabledExportFromTimestampNonZero(t *testing.T) {
	mes := &MessageExportSettings{
		EnableExport:        NewBool(true),
		ExportFromTimestamp: NewInt64(12345),
	}
	mes.SetDefaults()

	require.True(t, *mes.EnableExport)
	require.Equal(t, "01:00", *mes.DailyRunTime)
	require.Equal(t, int64(12345), *mes.ExportFromTimestamp)
	require.Equal(t, 10000, *mes.BatchSize)
}

func TestMessageExportSetDefaultsExportDisabledExportFromTimestampNil(t *testing.T) {
	mes := &MessageExportSettings{
		EnableExport: NewBool(false),
	}
	mes.SetDefaults()

	require.False(t, *mes.EnableExport)
	require.Equal(t, "01:00", *mes.DailyRunTime)
	require.Equal(t, int64(0), *mes.ExportFromTimestamp)
	require.Equal(t, 10000, *mes.BatchSize)
}

func TestMessageExportSetDefaultsExportDisabledExportFromTimestampZero(t *testing.T) {
	mes := &MessageExportSettings{
		EnableExport:        NewBool(false),
		ExportFromTimestamp: NewInt64(0),
	}
	mes.SetDefaults()

	require.False(t, *mes.EnableExport)
	require.Equal(t, "01:00", *mes.DailyRunTime)
	require.Equal(t, int64(0), *mes.ExportFromTimestamp)
	require.Equal(t, 10000, *mes.BatchSize)
}

func TestMessageExportSetDefaultsExportDisabledExportFromTimestampNonZero(t *testing.T) {
	// Test retained as protection against regression of MM-13185
	mes := &MessageExportSettings{
		EnableExport:        NewBool(false),
		ExportFromTimestamp: NewInt64(12345),
	}
	mes.SetDefaults()

	require.False(t, *mes.EnableExport)
	require.Equal(t, "01:00", *mes.DailyRunTime)
	require.Equal(t, int64(12345), *mes.ExportFromTimestamp)
	require.Equal(t, 10000, *mes.BatchSize)
}

func TestDisplaySettingsIsValidCustomUrlSchemes(t *testing.T) {
	tests := []struct {
		name  string
		value []string
		valid bool
	}{
		{
			name:  "empty",
			value: []string{},
			valid: true,
		},
		{
			name:  "custom protocol",
			value: []string{"steam"},
			valid: true,
		},
		{
			name:  "multiple custom protocols",
			value: []string{"bitcoin", "rss", "redis"},
			valid: true,
		},
		{
			name:  "containing numbers",
			value: []string{"ut2004", "ts3server", "h323"},
			valid: true,
		},
		{
			name:  "containing period",
			value: []string{"iris.beep"},
			valid: false, // should technically be true, but client doesn't support it
		},
		{
			name:  "containing hyphen",
			value: []string{"ms-excel"},
			valid: true,
		},
		{
			name:  "containing plus",
			value: []string{"coap+tcp", "coap+ws"},
			valid: false, // should technically be true, but client doesn't support it
		},
		{
			name:  "starting with number",
			value: []string{"4four"},
			valid: false,
		},
		{
			name:  "starting with period",
			value: []string{"data", ".dot"},
			valid: false,
		},
		{
			name:  "starting with hyphen",
			value: []string{"-hyphen", "dns"},
			valid: false,
		},
		{
			name:  "invalid symbols",
			value: []string{"!!fun!!"},
			valid: false,
		},
		{
			name:  "invalid letters",
			value: []string{"école"},
			valid: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ds := &DisplaySettings{}
			ds.SetDefaults()

			ds.CustomUrlSchemes = test.value

			if err := ds.isValid(); err != nil && test.valid {
				t.Error("Expected CustomUrlSchemes to be valid but got error:", err)
			} else if err == nil && !test.valid {
				t.Error("Expected CustomUrlSchemes to be invalid but got no error")
			}
		})
	}
}

func TestListenAddressIsValidated(t *testing.T) {

	testValues := map[string]bool{
		":8065":                true,
		":9917":                true,
		"0.0.0.0:9917":         true,
		"[2001:db8::68]:9918":  true,
		"[::1]:8065":           true,
		"localhost:8065":       true,
		"test.com:8065":        true,
		":0":                   true,
		":33147":               true,
		"123:8065":             false,
		"[::1]:99999":          false,
		"[::1]:-1":             false,
		"[::1]:8065a":          false,
		"0.0.0:9917":           false,
		"0.0.0.0:9917/":        false,
		"0..0.0:9917/":         false,
		"0.0.0222.0:9917/":     false,
		"http://0.0.0.0:9917/": false,
		"http://0.0.0.0:9917":  false,
		"8065":                 false,
		"[2001:db8::68]":       false,
	}

	for key, expected := range testValues {
		ss := &ServiceSettings{
			ListenAddress: NewString(key),
		}
		ss.SetDefaults()
		if expected {
			require.Nil(t, ss.isValid(), fmt.Sprintf("Got an error from '%v'.", key))
		} else {
			err := ss.isValid()
			require.NotNil(t, err, fmt.Sprintf("Expected '%v' to throw an error.", key))
			require.Equal(t, "model.config.is_valid.listen_address.app_error", err.Message)
		}
	}

}

func TestImageProxySettingsSetDefaults(t *testing.T) {
	ss := ServiceSettings{
		DEPRECATED_DO_NOT_USE_ImageProxyType:    NewString(IMAGE_PROXY_TYPE_ATMOS_CAMO),
		DEPRECATED_DO_NOT_USE_ImageProxyURL:     NewString("http://images.example.com"),
		DEPRECATED_DO_NOT_USE_ImageProxyOptions: NewString("1234abcd"),
	}

	t.Run("default, no old settings", func(t *testing.T) {
		ips := ImageProxySettings{}
		ips.SetDefaults(ServiceSettings{})

		assert.Equal(t, false, *ips.Enable)
		assert.Equal(t, IMAGE_PROXY_TYPE_LOCAL, *ips.ImageProxyType)
		assert.Equal(t, "", *ips.RemoteImageProxyURL)
		assert.Equal(t, "", *ips.RemoteImageProxyOptions)
	})

	t.Run("default, old settings", func(t *testing.T) {
		ips := ImageProxySettings{}
		ips.SetDefaults(ss)

		assert.Equal(t, true, *ips.Enable)
		assert.Equal(t, *ss.DEPRECATED_DO_NOT_USE_ImageProxyType, *ips.ImageProxyType)
		assert.Equal(t, *ss.DEPRECATED_DO_NOT_USE_ImageProxyURL, *ips.RemoteImageProxyURL)
		assert.Equal(t, *ss.DEPRECATED_DO_NOT_USE_ImageProxyOptions, *ips.RemoteImageProxyOptions)
	})

	t.Run("not default, old settings", func(t *testing.T) {
		url := "http://images.mattermost.com"
		options := "aaaaaaaa"

		ips := ImageProxySettings{
			Enable:                  NewBool(false),
			ImageProxyType:          NewString(IMAGE_PROXY_TYPE_LOCAL),
			RemoteImageProxyURL:     &url,
			RemoteImageProxyOptions: &options,
		}
		ips.SetDefaults(ss)

		assert.Equal(t, false, *ips.Enable)
		assert.Equal(t, IMAGE_PROXY_TYPE_LOCAL, *ips.ImageProxyType)
		assert.Equal(t, url, *ips.RemoteImageProxyURL)
		assert.Equal(t, options, *ips.RemoteImageProxyOptions)
	})
}

func TestImageProxySettingsIsValid(t *testing.T) {
	for _, test := range []struct {
		Name                    string
		Enable                  bool
		ImageProxyType          string
		RemoteImageProxyURL     string
		RemoteImageProxyOptions string
		ExpectError             bool
	}{
		{
			Name:        "disabled",
			Enable:      false,
			ExpectError: false,
		},
		{
			Name:                    "disabled with bad values",
			Enable:                  false,
			ImageProxyType:          "garbage",
			RemoteImageProxyURL:     "garbage",
			RemoteImageProxyOptions: "garbage",
			ExpectError:             false,
		},
		{
			Name:           "missing type",
			Enable:         true,
			ImageProxyType: "",
			ExpectError:    true,
		},
		{
			Name:                    "local",
			Enable:                  true,
			ImageProxyType:          "local",
			RemoteImageProxyURL:     "garbage",
			RemoteImageProxyOptions: "garbage",
			ExpectError:             false,
		},
		{
			Name:                    "atmos/camo",
			Enable:                  true,
			ImageProxyType:          IMAGE_PROXY_TYPE_ATMOS_CAMO,
			RemoteImageProxyURL:     "someurl",
			RemoteImageProxyOptions: "someoptions",
			ExpectError:             false,
		},
		{
			Name:                    "atmos/camo, missing url",
			Enable:                  true,
			ImageProxyType:          IMAGE_PROXY_TYPE_ATMOS_CAMO,
			RemoteImageProxyURL:     "",
			RemoteImageProxyOptions: "garbage",
			ExpectError:             true,
		},
		{
			Name:                    "atmos/camo, missing options",
			Enable:                  true,
			ImageProxyType:          IMAGE_PROXY_TYPE_ATMOS_CAMO,
			RemoteImageProxyURL:     "someurl",
			RemoteImageProxyOptions: "",
			ExpectError:             true,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			ips := &ImageProxySettings{
				Enable:                  &test.Enable,
				ImageProxyType:          &test.ImageProxyType,
				RemoteImageProxyURL:     &test.RemoteImageProxyURL,
				RemoteImageProxyOptions: &test.RemoteImageProxyOptions,
			}

			err := ips.isValid()
			if test.ExpectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestLdapSettingsIsValid(t *testing.T) {
	for _, test := range []struct {
		Name         string
		LdapSettings LdapSettings
		ExpectError  bool
	}{
		{
			Name: "disabled",
			LdapSettings: LdapSettings{
				Enable: NewBool(false),
			},
			ExpectError: false,
		},
		{
			Name: "missing server",
			LdapSettings: LdapSettings{
				Enable:            NewBool(true),
				LdapServer:        NewString(""),
				BaseDN:            NewString("basedn"),
				EmailAttribute:    NewString("email"),
				UsernameAttribute: NewString("username"),
				IdAttribute:       NewString("id"),
				LoginIdAttribute:  NewString("loginid"),
				UserFilter:        NewString(""),
			},
			ExpectError: true,
		},
		{
			Name: "empty user filter",
			LdapSettings: LdapSettings{
				Enable:            NewBool(true),
				LdapServer:        NewString("server"),
				BaseDN:            NewString("basedn"),
				EmailAttribute:    NewString("email"),
				UsernameAttribute: NewString("username"),
				IdAttribute:       NewString("id"),
				LoginIdAttribute:  NewString("loginid"),
				UserFilter:        NewString(""),
			},
			ExpectError: false,
		},
		{
			Name: "valid user filter #1",
			LdapSettings: LdapSettings{
				Enable:            NewBool(true),
				LdapServer:        NewString("server"),
				BaseDN:            NewString("basedn"),
				EmailAttribute:    NewString("email"),
				UsernameAttribute: NewString("username"),
				IdAttribute:       NewString("id"),
				LoginIdAttribute:  NewString("loginid"),
				UserFilter:        NewString("(property=value)"),
			},
			ExpectError: false,
		},
		{
			Name: "invalid user filter #1",
			LdapSettings: LdapSettings{
				Enable:            NewBool(true),
				LdapServer:        NewString("server"),
				BaseDN:            NewString("basedn"),
				EmailAttribute:    NewString("email"),
				UsernameAttribute: NewString("username"),
				IdAttribute:       NewString("id"),
				LoginIdAttribute:  NewString("loginid"),
				UserFilter:        NewString("("),
			},
			ExpectError: true,
		},
		{
			Name: "invalid user filter #2",
			LdapSettings: LdapSettings{
				Enable:            NewBool(true),
				LdapServer:        NewString("server"),
				BaseDN:            NewString("basedn"),
				EmailAttribute:    NewString("email"),
				UsernameAttribute: NewString("username"),
				IdAttribute:       NewString("id"),
				LoginIdAttribute:  NewString("loginid"),
				UserFilter:        NewString("()"),
			},
			ExpectError: true,
		},
		{
			Name: "valid user filter #2",
			LdapSettings: LdapSettings{
				Enable:            NewBool(true),
				LdapServer:        NewString("server"),
				BaseDN:            NewString("basedn"),
				EmailAttribute:    NewString("email"),
				UsernameAttribute: NewString("username"),
				IdAttribute:       NewString("id"),
				LoginIdAttribute:  NewString("loginid"),
				UserFilter:        NewString("(&(property=value)(otherthing=othervalue))"),
			},
			ExpectError: false,
		},
		{
			Name: "valid user filter #3",
			LdapSettings: LdapSettings{
				Enable:            NewBool(true),
				LdapServer:        NewString("server"),
				BaseDN:            NewString("basedn"),
				EmailAttribute:    NewString("email"),
				UsernameAttribute: NewString("username"),
				IdAttribute:       NewString("id"),
				LoginIdAttribute:  NewString("loginid"),
				UserFilter:        NewString("(&(property=value)(|(otherthing=othervalue)(other=thing)))"),
			},
			ExpectError: false,
		},
		{
			Name: "invalid user filter #3",
			LdapSettings: LdapSettings{
				Enable:            NewBool(true),
				LdapServer:        NewString("server"),
				BaseDN:            NewString("basedn"),
				EmailAttribute:    NewString("email"),
				UsernameAttribute: NewString("username"),
				IdAttribute:       NewString("id"),
				LoginIdAttribute:  NewString("loginid"),
				UserFilter:        NewString("(&(property=value)(|(otherthing=othervalue)(other=thing))"),
			},
			ExpectError: true,
		},
		{
			Name: "invalid user filter #4",
			LdapSettings: LdapSettings{
				Enable:            NewBool(true),
				LdapServer:        NewString("server"),
				BaseDN:            NewString("basedn"),
				EmailAttribute:    NewString("email"),
				UsernameAttribute: NewString("username"),
				IdAttribute:       NewString("id"),
				LoginIdAttribute:  NewString("loginid"),
				UserFilter:        NewString("(&(property=value)((otherthing=othervalue)(other=thing)))"),
			},
			ExpectError: true,
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			test.LdapSettings.SetDefaults()

			err := test.LdapSettings.isValid()
			if test.ExpectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
