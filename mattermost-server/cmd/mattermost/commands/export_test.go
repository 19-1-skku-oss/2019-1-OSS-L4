// Copyright (c) 2016-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package commands

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-server/model"
)

// There are no tests that actually run the Message Export job, because it can take a long time to complete depending
// on the size of the database that the config is pointing to. As such, these tests just ensure that the CLI command
// fails fast if invalid flags are supplied

func TestMessageExportNotEnabled(t *testing.T) {
	th := Setup()
	defer th.TearDown()

	config := th.Config()
	config.MessageExportSettings.EnableExport = model.NewBool(false)
	th.SetConfig(config)

	// should fail fast because the feature isn't enabled
	require.Error(t, th.RunCommand(t, "export", "schedule"))
}

func TestMessageExportInvalidFormat(t *testing.T) {
	th := Setup()
	defer th.TearDown()

	config := th.Config()
	config.MessageExportSettings.EnableExport = model.NewBool(true)
	th.SetConfig(config)

	// should fail fast because format isn't supported
	require.Error(t, th.RunCommand(t, "--format", "not_actiance", "export", "schedule"))
}

func TestMessageExportNegativeExportFrom(t *testing.T) {
	th := Setup()
	defer th.TearDown()

	config := th.Config()
	config.MessageExportSettings.EnableExport = model.NewBool(true)
	th.SetConfig(config)

	// should fail fast because export from must be a valid timestamp
	require.Error(t, th.RunCommand(t, "--format", "actiance", "--exportFrom", "-1", "export", "schedule"))
}

func TestMessageExportNegativeTimeoutSeconds(t *testing.T) {
	th := Setup()
	defer th.TearDown()

	config := th.Config()
	config.MessageExportSettings.EnableExport = model.NewBool(true)
	th.SetConfig(config)

	// should fail fast because timeout seconds must be a positive int
	require.Error(t, th.RunCommand(t, "--format", "actiance", "--exportFrom", "0", "--timeoutSeconds", "-1", "export", "schedule"))
}
